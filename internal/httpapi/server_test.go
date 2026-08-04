package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeChecker struct {
	ready bool
}

func (f fakeChecker) Check(context.Context) (map[string]string, bool) {
	state := "ok"
	if !f.ready {
		state = "unavailable"
	}
	return map[string]string{"postgres": state}, f.ready
}

func TestLiveness(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadinessFailure(t *testing.T) {
	server := New(fakeChecker{ready: false}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
