package synthetic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunReads(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.RequestURI())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	runner := testRunner(t, server.URL, server.Client(), 100*time.Millisecond)
	if err := runner.RunReads(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/categories", "/catalog/products?page=1&pageSize=20", "/catalog/products?q=basket&page=1&pageSize=20",
		"/catalog/products/handwoven-market-basket", "/catalog/products/handwoven-market-basket/reviews",
	}
	if len(paths) != len(want) {
		t.Fatalf("paths = %v", paths)
	}
	for index := range want {
		if paths[index] != want[index] {
			t.Fatalf("path[%d] = %q, want %q", index, paths[index], want[index])
		}
	}
}

func TestRunJourney(t *testing.T) {
	var idempotencyKey string
	var steps []string
	var notificationRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/auth/login":
			var input map[string]string
			_ = json.NewDecoder(r.Body).Decode(&input)
			if input["email"] != "buyer3@demo.cartlabs.local" || input["password"] != "demo-pass-123" {
				t.Error("unexpected login credentials")
			}
			_, _ = w.Write([]byte(`{"accessToken":"token"}`))
		case strings.HasPrefix(r.URL.Path, "/cart/items/"):
			assertBearer(t, r)
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/checkout":
			assertBearer(t, r)
			idempotencyKey = r.Header.Get("Idempotency-Key")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"01989f00-0000-7000-8000-000000000701"}`))
		case strings.HasSuffix(r.URL.Path, "/pay"):
			assertBearer(t, r)
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/notifications":
			assertBearer(t, r)
			notificationRequests++
			if notificationRequests == 1 {
				_, _ = w.Write([]byte(`{"items":[{"id":"old","kind":"purchase.payment_failed"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"id":"new","kind":"purchase.payment_failed"},{"id":"old","kind":"purchase.payment_failed"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	runner := testRunner(t, server.URL, server.Client(), 100*time.Millisecond)
	if err := runner.RunJourney(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(idempotencyKey, "synthetic-") {
		t.Fatalf("idempotency key = %q", idempotencyKey)
	}
	want := []string{"POST /auth/login", "GET /notifications", "PUT /cart/items/variant", "POST /checkout", "POST /purchases/01989f00-0000-7000-8000-000000000701/pay", "GET /notifications"}
	if strings.Join(steps, "|") != strings.Join(want, "|") {
		t.Fatalf("steps = %v", steps)
	}
}

func TestRunJourneyRejectsInvalidLoginResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()
	runner := testRunner(t, server.URL, server.Client(), 100*time.Millisecond)
	err := runner.RunJourney(context.Background())
	if err == nil || !strings.Contains(err.Error(), "decode login response") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunReadsHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	runner := testRunner(t, server.URL, server.Client(), 5*time.Millisecond)
	err := runner.RunReads(context.Background())
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("error = %v", err)
	}
}

func testRunner(t *testing.T, baseURL string, client *http.Client, requestTimeout time.Duration) *Runner {
	t.Helper()
	runner, err := New(Config{
		BaseURL: baseURL, Email: "buyer3@demo.cartlabs.local", Password: "demo-pass-123", VariantID: "variant",
		ReadInterval: time.Second, JourneyInterval: time.Minute, RequestTimeout: requestTimeout, NotificationWait: 10 * time.Millisecond,
	}, client, NewMetrics())
	if err != nil {
		t.Fatal(err)
	}
	return runner
}

func assertBearer(t *testing.T, request *http.Request) {
	t.Helper()
	if request.Header.Get("Authorization") != "Bearer token" {
		t.Errorf("authorization = %q", request.Header.Get("Authorization"))
	}
}
