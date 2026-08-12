package httpapi

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
)

func TestOpenAPIRequestValidation(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithServices(fakeIdentity{}, nil, nil))
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
	}{
		{name: "missing required body field", method: http.MethodPost, path: "/api/v1/auth/login", body: `{"email":"buyer@example.com"}`, contentType: "application/json"},
		{name: "unknown body field", method: http.MethodPost, path: "/api/v1/auth/login", body: `{"email":"buyer@example.com","password":"valid-password","admin":true}`, contentType: "application/json"},
		{name: "malformed JSON", method: http.MethodPost, path: "/api/v1/auth/login", body: `{"email":`, contentType: "application/json"},
		{name: "missing content type", method: http.MethodPost, path: "/api/v1/auth/login", body: `{"email":"buyer@example.com","password":"valid-password"}`},
		{name: "invalid enum", method: http.MethodPost, path: "/api/v1/auth/demo-login", body: `{"role":"operator"}`, contentType: "application/json"},
		{name: "invalid body UUID", method: http.MethodPost, path: "/api/v1/seller/products", body: `{"categoryId":"not-a-uuid","name":"Lamp","slug":"lamp","description":"Desk lamp"}`, contentType: "application/json"},
		{name: "invalid path UUID", method: http.MethodPut, path: "/api/v1/cart/items/not-a-uuid", body: `{"quantity":1}`, contentType: "application/json"},
		{name: "body below minimum", method: http.MethodPut, path: "/api/v1/cart/items/01989f00-0000-7000-8000-000000000001", body: `{"quantity":0}`, contentType: "application/json"},
		{name: "invalid slug pattern", method: http.MethodPost, path: "/api/v1/seller/products", body: `{"categoryId":"01989f00-0000-7000-8000-000000000001","name":"Lamp","slug":"Bad Slug","description":"Desk lamp"}`, contentType: "application/json"},
		{name: "query below minimum", method: http.MethodGet, path: "/api/v1/catalog/products?page=0"},
		{name: "missing required header", method: http.MethodPost, path: "/api/v1/checkout"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			request.Header.Set("Authorization", "Bearer access")
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestOpenAPIValidationPassesUnknownRoutesToMux(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/not-defined", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestOpenAPIValidationPreservesWebhookBody(t *testing.T) {
	payload := []byte(`{"id":"01989f00-0000-7000-8000-000000000001","type":"payment.succeeded","createdAt":"2026-08-12T10:00:00Z","data":{"intentId":"01989f00-0000-7000-8000-000000000002","reference":"PAY-1","amountMinor":249000,"currency":"IDR"}}`)
	var received []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		received, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := validateOpenAPIRequests(newOpenAPIRouter(), next)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Cartlabs-Signature", "signature")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !bytes.Equal(received, payload) {
		t.Fatalf("body changed: got %q, want %q", received, payload)
	}
}

func validateRecordedResponse(t *testing.T, request *http.Request, response *httptest.ResponseRecorder) {
	t.Helper()
	route, pathParams, err := newOpenAPIRouter().FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	requestInput := &openapi3filter.RequestValidationInput{Request: request, PathParams: pathParams, Route: route}
	responseInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestInput,
		Status:                 response.Code,
		Header:                 response.Header(),
		Options:                &openapi3filter.Options{IncludeResponseStatus: true, MultiError: true},
	}
	responseInput.SetBodyBytes(response.Body.Bytes())
	if err := openapi3filter.ValidateResponse(request.Context(), responseInput); err != nil {
		t.Fatalf("response violates OpenAPI: %v\n%s", err, response.Body.String())
	}
}

func TestRepresentativeHandlerResponsesMatchOpenAPI(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithServices(fakeIdentity{}, nil, nil))
	tests := []struct {
		name    string
		request *http.Request
	}{
		{name: "liveness", request: httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)},
		{name: "login", request: httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"buyer@example.com","password":"valid-password"}`))},
		{name: "unauthorized problem", request: httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)},
	}
	tests[1].request.Header.Set("Content-Type", "application/json")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, test.request)
			validateRecordedResponse(t, test.request, response)
		})
	}
}
