package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	marketstore "github.com/christolx/cartlabs/internal/store"
)

const (
	testAdminID   = "01989f00-0000-7000-8000-000000000011"
	testSellerID  = "01989f00-0000-7000-8000-000000000012"
	testProductID = "01989f00-0000-7000-8000-000000000013"
)

type backendIdentity struct{ fakeIdentity }

func (backendIdentity) Authenticate(_ context.Context, raw string) (identity.Principal, error) {
	switch raw {
	case "admin":
		return identity.Principal{UserID: testAdminID, Role: identity.RoleAdmin}, nil
	case "seller":
		return identity.Principal{UserID: testSellerID, Role: identity.RoleSeller}, nil
	case "buyer":
		return identity.Principal{UserID: testUserID, Role: identity.RoleBuyer}, nil
	default:
		return identity.Principal{}, domain.ErrUnauthorized
	}
}

func (backendIdentity) ListUsers(_ context.Context, principal identity.Principal) ([]identity.AdminUser, error) {
	if principal.Role != identity.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	return []identity.AdminUser{{ID: testSellerID, Email: "seller@example.com", DisplayName: "Seller", Role: identity.RoleSeller, Status: "active", CreatedAt: now, UpdatedAt: now}}, nil
}

func (backendIdentity) UpdateUserStatus(_ context.Context, principal identity.Principal, id, status, _ string) (identity.AdminUser, error) {
	if principal.Role != identity.RoleAdmin {
		return identity.AdminUser{}, domain.ErrForbidden
	}
	if id == testAdminID {
		return identity.AdminUser{}, domain.ErrConflict
	}
	if id != testSellerID {
		return identity.AdminUser{}, domain.ErrNotFound
	}
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	return identity.AdminUser{ID: id, Email: "seller@example.com", DisplayName: "Seller", Role: identity.RoleSeller, Status: status, CreatedAt: now, UpdatedAt: now}, nil
}

type publicStoreService struct{ StoreService }

func (publicStoreService) FindPublic(_ context.Context, slug string) (marketstore.Profile, error) {
	if slug != "north-star" {
		return marketstore.Profile{}, domain.ErrNotFound
	}
	return marketstore.Profile{ID: testSellerID, Name: "North Star", Slug: slug, Description: "Local goods", SellerDisplayName: "Seller", CreatedAt: time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)}, nil
}

type sellerCatalogService struct{ CatalogService }

func (sellerCatalogService) FindOwn(_ context.Context, principal identity.Principal, id string) (catalog.Product, error) {
	if principal.Role != identity.RoleSeller {
		return catalog.Product{}, domain.ErrForbidden
	}
	if id != testProductID {
		return catalog.Product{}, domain.ErrNotFound
	}
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	return catalog.Product{ID: id, StoreID: testSellerID, StoreName: "North Star", StoreSlug: "north-star",
		Category: catalog.Category{ID: testAdminID, Name: "Home", Slug: "home"}, Name: "Lamp", Slug: "lamp",
		Description: "Desk lamp", Status: "draft", ModerationStatus: "pending", Variants: []catalog.Variant{}, Images: []catalog.ProductImage{}, CreatedAt: now, UpdatedAt: now}, nil
}

func TestBackendCompletionHandlers(t *testing.T) {
	server := New(fakeChecker{ready: true}, slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithServices(backendIdentity{}, publicStoreService{}, sellerCatalogService{}))
	tests := []struct {
		name, method, path, token, body string
		want                            int
	}{
		{"public store", http.MethodGet, "/api/v1/catalog/stores/north-star", "", "", http.StatusOK},
		{"seller detail", http.MethodGet, "/api/v1/seller/products/" + testProductID, "seller", "", http.StatusOK},
		{"seller non-owner", http.MethodGet, "/api/v1/seller/products/01989f00-0000-7000-8000-000000000099", "seller", "", http.StatusNotFound},
		{"buyer seller detail", http.MethodGet, "/api/v1/seller/products/" + testProductID, "buyer", "", http.StatusForbidden},
		{"admin users", http.MethodGet, "/api/v1/admin/users", "admin", "", http.StatusOK},
		{"buyer admin users", http.MethodGet, "/api/v1/admin/users", "buyer", "", http.StatusForbidden},
		{"update status", http.MethodPatch, "/api/v1/admin/users/" + testSellerID + "/status", "admin", `{"status":"suspended","reason":"policy"}`, http.StatusOK},
		{"buyer update status", http.MethodPatch, "/api/v1/admin/users/" + testSellerID + "/status", "buyer", `{"status":"suspended","reason":"policy"}`, http.StatusForbidden},
		{"update missing user", http.MethodPatch, "/api/v1/admin/users/01989f00-0000-7000-8000-000000000099/status", "admin", `{"status":"suspended","reason":"policy"}`, http.StatusNotFound},
		{"unsafe status transition", http.MethodPatch, "/api/v1/admin/users/" + testAdminID + "/status", "admin", `{"status":"suspended","reason":"policy"}`, http.StatusConflict},
		{"missing public store", http.MethodGet, "/api/v1/catalog/stores/missing", "", "", http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			validateRecordedResponse(t, request, response)
			if test.name == "public store" {
				var body map[string]any
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				for _, private := range []string{"email", "sellerId", "status", "moderationNote"} {
					if _, exists := body[private]; exists {
						t.Fatalf("public response exposes %q: %s", private, response.Body.String())
					}
				}
			}
		})
	}
}
