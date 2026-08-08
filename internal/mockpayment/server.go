package mockpayment

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/christolx/cartlabs/internal/purchase"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	APIKey        string
	WebhookSecret []byte
	Client        *http.Client
	Now           func() time.Time
}

type Server struct {
	apiKey        string
	webhookSecret []byte
	client        *http.Client
	now           func() time.Time
	mu            sync.Mutex
	intents       map[string]*intentState
}

type intentState struct {
	purchase.PaymentIntent
	WebhookURL string
	EventID    string
	Outcome    string
	Payload    []byte
	Delivered  bool
}

func New(cfg Config) (*Server, error) {
	if len(cfg.APIKey) < 24 {
		return nil, fmt.Errorf("mock payment API key must contain at least 24 bytes")
	}
	if len(cfg.WebhookSecret) < 32 {
		return nil, fmt.Errorf("payment webhook secret must contain at least 32 bytes")
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 10 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Server{apiKey: cfg.APIKey, webhookSecret: cfg.WebhookSecret, client: cfg.Client,
		now: cfg.Now, intents: map[string]*intentState{}}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /v1/intents", s.auth(s.createIntent))
	mux.HandleFunc("POST /v1/intents/{intentId}/complete", s.auth(s.completeIntent))
	return mux
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(s.apiKey) || subtle.ConstantTimeCompare([]byte(provided), []byte(s.apiKey)) != 1 {
			writeProblem(w, http.StatusUnauthorized, "valid service credential required")
			return
		}
		next(w, r)
	}
}

func (s *Server) createIntent(w http.ResponseWriter, r *http.Request) {
	var input purchase.PaymentIntentRequest
	if err := decodeJSON(w, r, &input); err != nil || input.Reference == "" || input.AmountMinor < 0 ||
		input.Currency != "IDR" || !strings.HasPrefix(input.WebhookURL, "http://") && !strings.HasPrefix(input.WebhookURL, "https://") {
		writeProblem(w, http.StatusBadRequest, "request is invalid")
		return
	}
	id, err := uuid.NewV7()
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "intent ID generation failed")
		return
	}
	state := &intentState{PaymentIntent: purchase.PaymentIntent{ID: id.String(), Reference: input.Reference,
		AmountMinor: input.AmountMinor, Currency: input.Currency, Status: "pending"}, WebhookURL: input.WebhookURL}
	s.mu.Lock()
	s.intents[state.ID] = state
	s.mu.Unlock()
	writeJSON(w, http.StatusCreated, state.PaymentIntent)
}

func (s *Server) completeIntent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Outcome string `json:"outcome"`
	}
	if err := decodeJSON(w, r, &input); err != nil || input.Outcome != "succeeded" && input.Outcome != "failed" {
		writeProblem(w, http.StatusBadRequest, "request is invalid")
		return
	}
	id := r.PathValue("intentId")
	s.mu.Lock()
	state, ok := s.intents[id]
	if !ok {
		s.mu.Unlock()
		writeProblem(w, http.StatusNotFound, "payment intent not found")
		return
	}
	if state.Delivered {
		if state.Outcome != input.Outcome {
			s.mu.Unlock()
			writeProblem(w, http.StatusConflict, "payment intent already completed")
			return
		}
		output := state.PaymentIntent
		s.mu.Unlock()
		writeJSON(w, http.StatusOK, output)
		return
	}
	if state.Outcome != "" && state.Outcome != input.Outcome {
		s.mu.Unlock()
		writeProblem(w, http.StatusConflict, "payment completion already started")
		return
	}
	if state.EventID == "" {
		eventID, err := uuid.NewV7()
		if err != nil {
			s.mu.Unlock()
			writeProblem(w, http.StatusInternalServerError, "event ID generation failed")
			return
		}
		state.EventID = eventID.String()
		state.Outcome = input.Outcome
		event := purchase.PaymentEvent{ID: state.EventID, Type: "payment." + input.Outcome, CreatedAt: s.now().UTC(),
			Data: purchase.PaymentEventData{IntentID: state.ID, Reference: state.Reference, AmountMinor: state.AmountMinor, Currency: state.Currency}}
		state.Payload, err = json.Marshal(event)
		if err != nil {
			s.mu.Unlock()
			writeProblem(w, http.StatusInternalServerError, "event encoding failed")
			return
		}
	}
	payload := append([]byte(nil), state.Payload...)
	webhookURL := state.WebhookURL
	eventTime := s.now().UTC()
	s.mu.Unlock()

	request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		writeProblem(w, http.StatusBadGateway, "webhook request failed")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Cartlabs-Signature", purchase.SignWebhook(s.webhookSecret, payload, eventTime))
	response, err := s.client.Do(request)
	if err != nil {
		writeProblem(w, http.StatusBadGateway, "webhook delivery failed")
		return
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		writeProblem(w, http.StatusBadGateway, "webhook was rejected")
		return
	}
	s.mu.Lock()
	state.Delivered = true
	state.Status = input.Outcome
	output := state.PaymentIntent
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, output)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeProblem(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]any{"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail})
}
