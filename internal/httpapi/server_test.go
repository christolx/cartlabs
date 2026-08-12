package httpapi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/christolx/cartlabs/internal/observability"
)

type fakeChecker struct {
	ready bool
}

type fakeIdentity struct{}

const testUserID = "01989f00-0000-7000-8000-000000000001"

type fakeLimiter struct {
	allowed bool
	err     error
}

func (f fakeLimiter) Allow(context.Context, string, int, time.Duration) (bool, error) {
	return f.allowed, f.err
}

func (fakeIdentity) Login(_ context.Context, email, password string) (identity.Session, string, error) {
	if email != "buyer@example.com" || password != "valid-password" {
		return identity.Session{}, "", domain.ErrUnauthorized
	}
	user := identity.User{ID: testUserID, Email: email, DisplayName: "Buyer", Role: identity.RoleBuyer}
	return identity.Session{AccessToken: "access", TokenType: "Bearer", ExpiresIn: 900, User: user}, "refresh", nil
}
func (fakeIdentity) DemoLogin(context.Context, identity.Role) (identity.Session, string, error) {
	return identity.Session{}, "", domain.ErrForbidden
}
func (fakeIdentity) Refresh(context.Context, string) (identity.Session, string, error) {
	return identity.Session{}, "", domain.ErrUnauthorized
}
func (fakeIdentity) Logout(context.Context, string) error { return nil }
func (fakeIdentity) Authenticate(_ context.Context, raw string) (identity.Principal, error) {
	if raw != "access" {
		return identity.Principal{}, domain.ErrUnauthorized
	}
	return identity.Principal{UserID: testUserID, Role: identity.RoleBuyer}, nil
}
func (fakeIdentity) User(_ context.Context, principal identity.Principal) (identity.User, error) {
	return identity.User{ID: principal.UserID, Email: "buyer@example.com", DisplayName: "Buyer", Role: principal.Role}, nil
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
	if response.Header().Get("X-Request-ID") == "" || response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("operability headers = %#v", response.Header())
	}
}

func TestMutationRateLimit(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithServices(fakeIdentity{}, nil, nil), WithAuthRateLimiter(fakeLimiter{allowed: false}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(`{"purchaseItemId":"01989f00-0000-7000-8000-000000000101","rating":5,"title":"Good","body":"Works"}`))
	request.Header.Set("Authorization", "Bearer access")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestMetricsUseBoundedRouteLabels(t *testing.T) {
	metrics := observability.NewHTTPMetrics("api")
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithMetrics(metrics))
	server.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `route="/api/v1/health/live"`) {
		t.Fatalf("metrics status=%d body=%s", response.Code, response.Body.String())
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

func TestLoginSetsHttpOnlyRefreshCookie(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithServices(fakeIdentity{}, nil, nil), WithRefreshCookie(true, 24*time.Hour))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"buyer@example.com","password":"valid-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "__Host-cartlabs_refresh" || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie = %#v", cookies)
	}
}

func TestLoginRateLimit(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithServices(fakeIdentity{}, nil, nil), WithAuthRateLimiter(fakeLimiter{allowed: false}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"buyer@example.com","password":"valid-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Retry-After") != "60" {
		t.Fatalf("Retry-After = %q", response.Header().Get("Retry-After"))
	}
}

func TestLoginRejectsUnknownJSONField(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithServices(fakeIdentity{}, nil, nil))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"buyer@example.com","password":"valid-password","admin":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestProtectedRouteRequiresAndAcceptsBearerToken(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithServices(fakeIdentity{}, nil, nil))
	missing := httptest.NewRecorder()
	server.Handler().ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing status = %d", missing.Code)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Authorization", "Bearer access")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
