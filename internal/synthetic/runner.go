package synthetic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/christolx/cartlabs/synthetic"

type Config struct {
	BaseURL          string
	Email            string
	Password         string
	VariantID        string
	ReadInterval     time.Duration
	JourneyInterval  time.Duration
	RequestTimeout   time.Duration
	NotificationWait time.Duration
}

type Metrics struct {
	Registry        *prometheus.Registry
	Requests        *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	Journeys        *prometheus.CounterVec
	JourneyDuration prometheus.Histogram
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		Registry: registry,
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cartlabs_synthetic_requests_total", Help: "Synthetic requests by bounded step and result.",
		}, []string{"step", "result"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "cartlabs_synthetic_request_duration_seconds", Help: "Synthetic request duration by bounded step.",
		}, []string{"step"}),
		Journeys: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cartlabs_synthetic_journeys_total", Help: "Synthetic business journeys by result.",
		}, []string{"result"}),
		JourneyDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "cartlabs_synthetic_journey_duration_seconds", Help: "Synthetic business journey duration.",
		}),
	}
	registry.MustRegister(metrics.Requests, metrics.RequestDuration, metrics.Journeys, metrics.JourneyDuration)
	return metrics
}

type Runner struct {
	config  Config
	client  *http.Client
	metrics *Metrics
	tracer  trace.Tracer
}

func New(config Config, client *http.Client, metrics *Metrics) (*Runner, error) {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if config.BaseURL == "" || config.Email == "" || config.Password == "" || config.VariantID == "" {
		return nil, fmt.Errorf("synthetic base URL, credentials, and variant ID are required")
	}
	if config.ReadInterval <= 0 || config.JourneyInterval <= 0 || config.RequestTimeout <= 0 || config.NotificationWait <= 0 {
		return nil, fmt.Errorf("synthetic intervals and timeouts must be positive")
	}
	if client == nil || metrics == nil {
		return nil, fmt.Errorf("synthetic HTTP client and metrics are required")
	}
	return &Runner{config: config, client: client, metrics: metrics, tracer: otel.Tracer(instrumentationName)}, nil
}

func (r *Runner) Run(ctx context.Context, observe func(string, error)) {
	r.runObserved(ctx, "reads", r.RunReads, observe)
	reads := time.NewTicker(r.config.ReadInterval)
	journeys := time.NewTicker(r.config.JourneyInterval)
	defer reads.Stop()
	defer journeys.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-reads.C:
			r.runObserved(ctx, "reads", r.RunReads, observe)
		case <-journeys.C:
			r.runObserved(ctx, "journey", r.RunJourney, observe)
		}
	}
}

func (r *Runner) runObserved(ctx context.Context, name string, run func(context.Context) error, observe func(string, error)) {
	err := run(ctx)
	if observe != nil {
		observe(name, err)
	}
}

func (r *Runner) RunReads(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "synthetic.reads")
	defer span.End()
	reads := []struct{ step, path string }{
		{"categories", "/categories"},
		{"catalog", "/catalog/products?page=1&pageSize=20"},
		{"search", "/catalog/products?q=basket&page=1&pageSize=20"},
		{"product", "/catalog/products/handwoven-market-basket"},
		{"reviews", "/catalog/products/handwoven-market-basket/reviews"},
	}
	for _, read := range reads {
		if err := r.request(ctx, read.step, http.MethodGet, read.path, "", nil, nil); err != nil {
			span.RecordError(err)
			return err
		}
	}
	return nil
}

func (r *Runner) RunJourney(ctx context.Context) error {
	started := time.Now()
	ctx, span := r.tracer.Start(ctx, "synthetic.failed-payment-journey")
	defer span.End()
	result := "success"
	defer func() {
		r.metrics.Journeys.WithLabelValues(result).Inc()
		r.metrics.JourneyDuration.Observe(time.Since(started).Seconds())
	}()
	fail := func(err error) error {
		result = "failed"
		span.RecordError(err)
		return err
	}

	var session struct {
		AccessToken string `json:"accessToken"`
	}
	if err := r.request(ctx, "login", http.MethodPost, "/auth/login", "", map[string]string{
		"email": r.config.Email, "password": r.config.Password,
	}, &session); err != nil {
		return fail(err)
	}
	if session.AccessToken == "" {
		return fail(fmt.Errorf("login returned empty access token"))
	}
	knownNotifications, err := r.notificationIDs(ctx, session.AccessToken)
	if err != nil {
		return fail(err)
	}
	if err := r.request(ctx, "cart", http.MethodPut, "/cart/items/"+r.config.VariantID, session.AccessToken,
		map[string]int{"quantity": 1}, nil); err != nil {
		return fail(err)
	}
	var purchase struct {
		ID string `json:"id"`
	}
	key := "synthetic-" + uuid.NewString()
	if err := r.requestWithHeaders(ctx, "checkout", http.MethodPost, "/checkout", session.AccessToken, nil, &purchase,
		map[string]string{"Idempotency-Key": key}); err != nil {
		return fail(err)
	}
	if purchase.ID == "" {
		return fail(fmt.Errorf("checkout returned empty purchase ID"))
	}
	if err := r.request(ctx, "payment", http.MethodPost, "/purchases/"+purchase.ID+"/pay", session.AccessToken,
		map[string]string{"outcome": "failed"}, nil); err != nil {
		return fail(err)
	}
	if err := r.waitForNotification(ctx, session.AccessToken, knownNotifications); err != nil {
		return fail(err)
	}
	return nil
}

type notificationsResponse struct {
	Items []struct {
		ID   string `json:"id"`
		Kind string `json:"kind"`
	} `json:"items"`
}

func (r *Runner) notificationIDs(ctx context.Context, token string) (map[string]struct{}, error) {
	var response notificationsResponse
	if err := r.request(ctx, "notifications", http.MethodGet, "/notifications", token, nil, &response); err != nil {
		return nil, err
	}
	ids := make(map[string]struct{}, len(response.Items))
	for _, notification := range response.Items {
		ids[notification.ID] = struct{}{}
	}
	return ids, nil
}

func (r *Runner) waitForNotification(ctx context.Context, token string, known map[string]struct{}) error {
	deadline := time.Now().Add(r.config.NotificationWait)
	for {
		var response notificationsResponse
		if err := r.request(ctx, "notifications", http.MethodGet, "/notifications", token, nil, &response); err != nil {
			return err
		}
		for _, notification := range response.Items {
			if _, existed := known[notification.ID]; !existed && notification.Kind == "purchase.payment_failed" {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("payment failure notification not observed within %s", r.config.NotificationWait)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (r *Runner) request(ctx context.Context, step, method, path, token string, input, output any) error {
	return r.requestWithHeaders(ctx, step, method, path, token, input, output, nil)
}

func (r *Runner) requestWithHeaders(ctx context.Context, step, method, path, token string, input, output any, headers map[string]string) error {
	started := time.Now()
	result := "success"
	defer func() {
		r.metrics.Requests.WithLabelValues(step, result).Inc()
		r.metrics.RequestDuration.WithLabelValues(step).Observe(time.Since(started).Seconds())
	}()
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			result = "failed"
			return fmt.Errorf("encode %s request: %w", step, err)
		}
		body = bytes.NewReader(encoded)
	}
	requestCtx, cancel := context.WithTimeout(ctx, r.config.RequestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, method, r.config.BaseURL+path, body)
	if err != nil {
		result = "failed"
		return fmt.Errorf("create %s request: %w", step, err)
	}
	request.Header.Set("User-Agent", "cartlabs-synthetic/1.0")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := r.client.Do(request)
	if err != nil {
		result = "failed"
		return fmt.Errorf("perform %s request: %w", step, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		result = "failed"
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return fmt.Errorf("%s request returned status %d", step, response.StatusCode)
	}
	if output != nil {
		if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
			result = "failed"
			return fmt.Errorf("decode %s response: %w", step, err)
		}
		return nil
	}
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	if err != nil {
		result = "failed"
	}
	return err
}
