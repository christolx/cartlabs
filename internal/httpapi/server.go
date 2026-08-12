package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/christolx/cartlabs/internal/observability"
	"github.com/christolx/cartlabs/internal/purchase"
	marketstore "github.com/christolx/cartlabs/internal/store"
	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ReadinessChecker interface {
	Check(context.Context) (map[string]string, bool)
}

type IdentityService interface {
	Login(context.Context, string, string) (identity.Session, string, error)
	DemoLogin(context.Context, identity.Role) (identity.Session, string, error)
	Refresh(context.Context, string) (identity.Session, string, error)
	Logout(context.Context, string) error
	Authenticate(context.Context, string) (identity.Principal, error)
	User(context.Context, identity.Principal) (identity.User, error)
}

type RateLimiter interface {
	Allow(context.Context, string, int, time.Duration) (bool, error)
}

type StoreService interface {
	GetOwn(context.Context, identity.Principal) (marketstore.Store, error)
	Create(context.Context, identity.Principal, marketstore.Input) (marketstore.Store, error)
	Update(context.Context, identity.Principal, marketstore.Input) (marketstore.Store, error)
	ListForAdmin(context.Context, identity.Principal) ([]marketstore.Store, error)
	Moderate(context.Context, identity.Principal, string, string, string) (marketstore.Store, error)
}

type CatalogService interface {
	Categories(context.Context) ([]catalog.Category, error)
	ListOwn(context.Context, identity.Principal) ([]catalog.Product, error)
	Create(context.Context, identity.Principal, catalog.ProductInput) (catalog.Product, error)
	Update(context.Context, identity.Principal, string, catalog.ProductInput) (catalog.Product, error)
	AddVariant(context.Context, identity.Principal, string, catalog.VariantInput) (catalog.Variant, error)
	AddImage(context.Context, identity.Principal, string, catalog.ImageInput) (catalog.ProductImage, error)
	Publish(context.Context, identity.Principal, string) (catalog.Product, error)
	AdjustInventory(context.Context, identity.Principal, string, int, string) (catalog.Variant, error)
	ListForAdmin(context.Context, identity.Principal) ([]catalog.Product, error)
	Moderate(context.Context, identity.Principal, string, string, string) (catalog.Product, error)
	ListPublic(context.Context, catalog.Filters) (catalog.Page, error)
	FindPublic(context.Context, string) (catalog.Product, error)
}

type PurchaseService interface {
	Cart(context.Context, identity.Principal) (purchase.Cart, error)
	SetCartItem(context.Context, identity.Principal, string, int) (purchase.Cart, error)
	RemoveCartItem(context.Context, identity.Principal, string) (purchase.Cart, error)
	Checkout(context.Context, identity.Principal, string) (purchase.Purchase, error)
	ListPurchases(context.Context, identity.Principal) ([]purchase.Purchase, error)
	Purchase(context.Context, identity.Principal, string) (purchase.Purchase, error)
	SellerOrders(context.Context, identity.Principal) ([]purchase.SellerOrder, error)
	ConfirmPayment(context.Context, identity.Principal, string, string) (purchase.Purchase, error)
	HandleWebhook(context.Context, []byte, string) error
	Notifications(context.Context, identity.Principal) ([]purchase.Notification, error)
	UpdateSellerOrder(context.Context, identity.Principal, string, string, string) (purchase.SellerOrder, error)
	CancelPurchase(context.Context, identity.Principal, string, string) (purchase.Purchase, error)
	CreateReview(context.Context, identity.Principal, purchase.ReviewInput) (purchase.Review, error)
	Reviews(context.Context, string) (purchase.ReviewSummary, error)
	AdminOverview(context.Context, identity.Principal) (purchase.AdminOverview, error)
	AuditEvents(context.Context, identity.Principal) ([]purchase.AuditEvent, error)
}

type Server struct {
	handler http.Handler
}

type serverConfig struct {
	identity      IdentityService
	stores        StoreService
	catalog       CatalogService
	purchases     PurchaseService
	cookieSecure  bool
	refreshMaxAge int
	rateLimiter   RateLimiter
	cookieName    string
	cookiePath    string
	metrics       *observability.HTTPMetrics
}

func WithAuthRateLimiter(limiter RateLimiter) Option {
	return func(config *serverConfig) { config.rateLimiter = limiter }
}

type Option func(*serverConfig)

func WithServices(identityService IdentityService, storeService StoreService, catalogService CatalogService) Option {
	return func(config *serverConfig) {
		config.identity = identityService
		config.stores = storeService
		config.catalog = catalogService
	}
}

func WithPurchaseService(service PurchaseService) Option {
	return func(config *serverConfig) { config.purchases = service }
}

func WithMetrics(metrics *observability.HTTPMetrics) Option {
	return func(config *serverConfig) { config.metrics = metrics }
}

func WithRefreshCookie(secure bool, maxAge time.Duration) Option {
	return func(config *serverConfig) {
		config.cookieSecure = secure
		config.refreshMaxAge = int(maxAge.Seconds())
		if secure {
			config.cookieName = "__Host-cartlabs_refresh"
			config.cookiePath = "/"
		}
	}
}

type api struct {
	checker ReadinessChecker
	config  serverConfig
}

func New(checker ReadinessChecker, logger *slog.Logger, options ...Option) *Server {
	config := serverConfig{refreshMaxAge: int((7 * 24 * time.Hour).Seconds()), cookieName: "cartlabs_refresh", cookiePath: "/api/v1/auth"}
	for _, option := range options {
		option(&config)
	}
	application := &api{checker: checker, config: config}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health/live", application.liveness)
	mux.HandleFunc("GET /api/v1/health/ready", application.readiness)
	mux.HandleFunc("POST /api/v1/auth/login", application.login)
	mux.HandleFunc("POST /api/v1/auth/demo-login", application.demoLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", application.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", application.logout)
	mux.HandleFunc("GET /api/v1/me", application.auth(application.currentUser))
	mux.HandleFunc("GET /api/v1/categories", application.categories)
	mux.HandleFunc("GET /api/v1/catalog/products", application.catalogProducts)
	mux.HandleFunc("GET /api/v1/catalog/products/{slug}", application.catalogProduct)
	mux.HandleFunc("GET /api/v1/seller/store", application.auth(application.getStore))
	mux.HandleFunc("POST /api/v1/seller/store", application.auth(application.createStore))
	mux.HandleFunc("PATCH /api/v1/seller/store", application.auth(application.updateStore))
	mux.HandleFunc("GET /api/v1/seller/products", application.auth(application.sellerProducts))
	mux.HandleFunc("POST /api/v1/seller/products", application.auth(application.createProduct))
	mux.HandleFunc("PATCH /api/v1/seller/products/{productId}", application.auth(application.updateProduct))
	mux.HandleFunc("POST /api/v1/seller/products/{productId}/variants", application.auth(application.createVariant))
	mux.HandleFunc("POST /api/v1/seller/products/{productId}/images", application.auth(application.createImage))
	mux.HandleFunc("POST /api/v1/seller/products/{productId}/publish", application.auth(application.publishProduct))
	mux.HandleFunc("PATCH /api/v1/seller/variants/{variantId}/inventory", application.auth(application.adjustInventory))
	mux.HandleFunc("GET /api/v1/admin/stores", application.auth(application.adminStores))
	mux.HandleFunc("PATCH /api/v1/admin/stores/{storeId}/moderation", application.auth(application.moderateStore))
	mux.HandleFunc("GET /api/v1/admin/products", application.auth(application.adminProducts))
	mux.HandleFunc("PATCH /api/v1/admin/products/{productId}/moderation", application.auth(application.moderateProduct))
	mux.HandleFunc("GET /api/v1/cart", application.auth(application.getCart))
	mux.HandleFunc("PUT /api/v1/cart/items/{variantId}", application.auth(application.setCartItem))
	mux.HandleFunc("DELETE /api/v1/cart/items/{variantId}", application.auth(application.removeCartItem))
	mux.HandleFunc("POST /api/v1/checkout", application.auth(application.checkout))
	mux.HandleFunc("GET /api/v1/purchases", application.auth(application.listPurchases))
	mux.HandleFunc("GET /api/v1/purchases/{purchaseId}", application.auth(application.getPurchase))
	mux.HandleFunc("POST /api/v1/purchases/{purchaseId}/pay", application.auth(application.confirmPayment))
	mux.HandleFunc("GET /api/v1/seller/orders", application.auth(application.sellerOrders))
	mux.HandleFunc("PATCH /api/v1/seller/orders/{orderId}/status", application.auth(application.updateSellerOrder))
	mux.HandleFunc("POST /api/v1/purchases/{purchaseId}/cancel", application.auth(application.cancelPurchase))
	mux.HandleFunc("POST /api/v1/reviews", application.auth(application.createReview))
	mux.HandleFunc("GET /api/v1/catalog/products/{slug}/reviews", application.productReviews)
	mux.HandleFunc("GET /api/v1/admin/overview", application.auth(application.adminOverview))
	mux.HandleFunc("GET /api/v1/admin/audit-events", application.auth(application.auditEvents))
	mux.HandleFunc("GET /api/v1/notifications", application.auth(application.notifications))
	mux.HandleFunc("POST /api/v1/payments/webhook", application.paymentWebhook)
	if config.metrics != nil {
		mux.Handle("GET /metrics", config.metrics.Handler())
	}

	validatedMux := validateOpenAPIRequests(newOpenAPIRouter(), mux)
	var handler http.Handler = recoverPanic(logger, routeSpan(mux, validatedMux))
	handler = securityHeaders(handler)
	handler = requestLogger(logger, handler)
	if config.metrics != nil {
		handler = config.metrics.Middleware("api", handler)
	}
	handler = requestID(handler)
	handler = otelhttp.NewHandler(handler, "http.server", otelhttp.WithSpanNameFormatter(func(_ string, request *http.Request) string { return request.Method }))
	return &Server{handler: handler}
}

func (s *Server) Handler() http.Handler { return s.handler }

func routeSpan(mux *http.ServeMux, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		route := routeTemplate(pattern)
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + route)
		span.SetAttributes(attribute.String("http.route", route))
		next.ServeHTTP(w, r)
	})
}

func (a *api) liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, contractHealth(contract.Ok, nil))
}

func (a *api) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	dependencies, ready := a.checker.Check(ctx)
	if !ready {
		writeJSON(w, http.StatusServiceUnavailable, contractHealth(contract.Degraded, dependencies))
		return
	}
	writeJSON(w, http.StatusOK, contractHealth(contract.Ok, dependencies))
}

func (a *api) auth(next func(http.ResponseWriter, *http.Request, identity.Principal)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.config.identity == nil {
			writeProblem(w, http.StatusServiceUnavailable, "application services unavailable")
			return
		}
		header := r.Header.Get("Authorization")
		kind, raw, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(kind, "Bearer") || strings.TrimSpace(raw) == "" {
			writeProblem(w, http.StatusUnauthorized, "bearer token required")
			return
		}
		principal, err := a.config.identity.Authenticate(r.Context(), strings.TrimSpace(raw))
		if err != nil {
			writeError(w, err)
			return
		}
		if err := a.checkMutationLimit(r, principal); err != nil {
			writeError(w, err)
			return
		}
		trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("enduser.id", principal.UserID), attribute.String("enduser.role", string(principal.Role)))
		next(w, r, principal)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domain.ErrInvalid
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return domain.ErrInvalid
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(contractProblem(status, detail))
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		writeProblem(w, http.StatusBadRequest, "request is invalid")
	case errors.Is(err, domain.ErrUnauthorized):
		writeProblem(w, http.StatusUnauthorized, "credentials or session are invalid")
	case errors.Is(err, domain.ErrForbidden):
		writeProblem(w, http.StatusForbidden, "permission denied")
	case errors.Is(err, domain.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, domain.ErrConflict):
		writeProblem(w, http.StatusConflict, "resource state conflicts with request")
	case errors.Is(err, domain.ErrRateLimited):
		w.Header().Set("Retry-After", "60")
		writeProblem(w, http.StatusTooManyRequests, "too many requests")
	case errors.Is(err, domain.ErrUnavailable):
		writeProblem(w, http.StatusServiceUnavailable, "service temporarily unavailable")
	default:
		writeProblem(w, http.StatusInternalServerError, "internal server error")
	}
}

func (a *api) checkAuthLimit(r *http.Request, subject string) error {
	if a.config.rateLimiter == nil {
		return nil
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	allowed, err := a.config.rateLimiter.Allow(r.Context(), identity.RateLimitKey(host, subject), 10, time.Minute)
	if err != nil {
		return domain.ErrUnavailable
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}

func (a *api) checkMutationLimit(r *http.Request, principal identity.Principal) error {
	if a.config.rateLimiter == nil || r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return nil
	}
	allowed, err := a.config.rateLimiter.Allow(r.Context(), "mutation:"+principal.UserID, 120, time.Minute)
	if err != nil {
		return domain.ErrUnavailable
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}

func (a *api) setRefreshCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{Name: a.config.cookieName, Value: value, Path: a.config.cookiePath, MaxAge: a.config.refreshMaxAge,
		HttpOnly: true, Secure: a.config.cookieSecure, SameSite: http.SameSiteStrictMode})
}

func (a *api) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: a.config.cookieName, Path: a.config.cookiePath, MaxAge: -1,
		HttpOnly: true, Secure: a.config.cookieSecure, SameSite: http.SameSiteStrictMode})
}

func (a *api) refreshCookie(r *http.Request) string {
	cookie, err := r.Cookie(a.config.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := httpsnoop.CaptureMetrics(next, w, r)
		spanContext := trace.SpanContextFromContext(r.Context())
		route := routeTemplate(r.Pattern)
		logger.InfoContext(r.Context(), "request handled", "method", r.Method, "route", route, "status", captured.Code,
			"duration_ms", captured.Duration.Milliseconds(), "response_bytes", captured.Written,
			"request_id", RequestID(r.Context()), "trace_id", spanContext.TraceID().String())
	})
}

func routeTemplate(pattern string) string {
	if pattern == "" {
		return "unmatched"
	}
	if _, path, found := strings.Cut(pattern, " "); found {
		return path
	}
	return pattern
}

type requestIDKey struct{}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if len(value) < 8 || len(value) > 128 || strings.ContainsAny(value, "\r\n\t ") {
			value = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", value)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, value)))
	})
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func recoverPanic(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				spanContext := trace.SpanContextFromContext(r.Context())
				logger.ErrorContext(r.Context(), "request panic", "error", fmt.Sprint(recovered), "request_id", RequestID(r.Context()), "trace_id", spanContext.TraceID().String())
				writeProblem(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
