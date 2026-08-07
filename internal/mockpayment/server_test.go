package mockpayment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/purchase"
)

func TestIntentCompletionDeliversSignedWebhookOnce(t *testing.T) {
	now := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	var deliveries atomic.Int32
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var raw bytes.Buffer
		_, _ = raw.ReadFrom(r.Body)
		if err := purchase.VerifyWebhookSignature(secret, raw.Bytes(), r.Header.Get("X-Cartlabs-Signature"), now, time.Minute); err != nil {
			t.Errorf("signature: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		deliveries.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer webhook.Close()
	server, err := New(Config{APIKey: "012345678901234567890123", WebhookSecret: secret, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	application := httptest.NewServer(server.Handler())
	defer application.Close()

	createBody := []byte(`{"reference":"CL-01989F000000","amountMinor":408000,"currency":"IDR","webhookUrl":"` + webhook.URL + `"}`)
	request, _ := http.NewRequest(http.MethodPost, application.URL+"/v1/intents", bytes.NewReader(createBody))
	request.Header.Set("Authorization", "Bearer 012345678901234567890123")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d", response.StatusCode)
	}
	var intent purchase.PaymentIntent
	if err := json.NewDecoder(response.Body).Decode(&intent); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		complete, _ := http.NewRequest(http.MethodPost, application.URL+"/v1/intents/"+intent.ID+"/complete", bytes.NewBufferString(`{"outcome":"succeeded"}`))
		complete.Header.Set("Authorization", "Bearer 012345678901234567890123")
		complete.Header.Set("Content-Type", "application/json")
		completed, err := http.DefaultClient.Do(complete)
		if err != nil {
			t.Fatal(err)
		}
		_ = completed.Body.Close()
		if completed.StatusCode != http.StatusOK {
			t.Fatalf("complete status=%d", completed.StatusCode)
		}
	}
	if deliveries.Load() != 1 {
		t.Fatalf("deliveries=%d", deliveries.Load())
	}
}

func TestIntentEndpointRequiresServiceCredential(t *testing.T) {
	server, err := New(Config{APIKey: "012345678901234567890123", WebhookSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/intents", bytes.NewBufferString(`{}`)))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestIntentCompletionRetriesSameWebhookEvent(t *testing.T) {
	now := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	var attempts atomic.Int32
	var mu sync.Mutex
	eventIDs := []string{}
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var event purchase.PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode webhook: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		eventIDs = append(eventIDs, event.ID)
		mu.Unlock()
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer webhook.Close()
	server, err := New(Config{APIKey: "012345678901234567890123", WebhookSecret: secret, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	application := httptest.NewServer(server.Handler())
	defer application.Close()

	createBody := []byte(`{"reference":"CL-01989F000000","amountMinor":408000,"currency":"IDR","webhookUrl":"` + webhook.URL + `"}`)
	request, _ := http.NewRequest(http.MethodPost, application.URL+"/v1/intents", bytes.NewReader(createBody))
	request.Header.Set("Authorization", "Bearer 012345678901234567890123")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var intent purchase.PaymentIntent
	if err := json.NewDecoder(response.Body).Decode(&intent); err != nil {
		t.Fatal(err)
	}

	wantStatuses := []int{http.StatusBadGateway, http.StatusOK, http.StatusOK}
	for _, wantStatus := range wantStatuses {
		complete, _ := http.NewRequest(http.MethodPost, application.URL+"/v1/intents/"+intent.ID+"/complete", bytes.NewBufferString(`{"outcome":"succeeded"}`))
		complete.Header.Set("Authorization", "Bearer 012345678901234567890123")
		complete.Header.Set("Content-Type", "application/json")
		completed, requestErr := http.DefaultClient.Do(complete)
		if requestErr != nil {
			t.Fatal(requestErr)
		}
		_ = completed.Body.Close()
		if completed.StatusCode != wantStatus {
			t.Fatalf("complete status=%d, want %d", completed.StatusCode, wantStatus)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if attempts.Load() != 2 || len(eventIDs) != 2 || eventIDs[0] == "" || eventIDs[0] != eventIDs[1] {
		t.Fatalf("attempts=%d eventIDs=%v", attempts.Load(), eventIDs)
	}
}
