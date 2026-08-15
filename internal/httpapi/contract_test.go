package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/christolx/cartlabs/internal/purchase"
	marketstore "github.com/christolx/cartlabs/internal/store"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

const (
	contractID1 = "01989f00-0000-7000-8000-000000000001"
	contractID2 = "01989f00-0000-7000-8000-000000000002"
	contractID3 = "01989f00-0000-7000-8000-000000000003"
	contractID4 = "01989f00-0000-7000-8000-000000000004"
)

func validateContractComponent(t *testing.T, name string, value any) {
	t.Helper()
	_ = newOpenAPIRouter()
	spec, err := contract.GetSwagger()
	if err != nil {
		t.Fatal(err)
	}
	schema := spec.Components.Schemas[name]
	if schema == nil || schema.Value == nil {
		t.Fatalf("component schema %q missing", name)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if err := schema.Value.VisitJSON(document, openapi3.VisitAsResponse(), openapi3.EnableFormatValidation(), openapi3.MultiErrors()); err != nil {
		t.Fatalf("%s invalid: %v\n%s", name, err, payload)
	}
}

func TestResponseMappersMatchOpenAPIComponents(t *testing.T) {
	now := time.Date(2026, time.August, 12, 10, 0, 0, 0, time.UTC)
	category := catalog.Category{ID: contractID1, Name: "Lighting", Slug: "lighting"}
	variant := catalog.Variant{ID: contractID2, SKU: "LAMP-01", Name: "Brass", Attributes: map[string]string{"finish": "brass"}, PriceMinor: 249000, Currency: "IDR", Stock: 2, Active: true}
	image := catalog.ProductImage{ID: contractID3, URL: "/images/lamp.webp", AltText: "Lamp", Position: 0}
	product := catalog.Product{
		ID: contractID4, StoreID: contractID1, StoreName: "Studio", StoreSlug: "studio", Category: category,
		Name: "Lamp", Slug: "lamp", Description: "Desk lamp", Status: "published",
		ModerationStatus: "approved", ModerationNote: "", Variants: []catalog.Variant{variant},
		Images: []catalog.ProductImage{image}, CreatedAt: now, UpdatedAt: now,
	}
	purchaseItem := purchase.PurchaseItem{
		ID: contractID1, ProductID: contractID4, VariantID: contractID2, ProductName: "Lamp",
		VariantName: "Brass", SKU: "LAMP-01", ImageURL: "/images/lamp.webp", Quantity: 1,
		UnitPriceMinor: 249000, LineTotalMinor: 249000, Currency: "IDR",
	}
	order := purchase.SellerOrder{
		ID: contractID2, PurchaseID: contractID3, Reference: "ORDER-1", StoreID: contractID1,
		StoreName: "Studio", Status: "paid", Currency: "IDR", SubtotalMinor: 249000,
		Items: []purchase.PurchaseItem{purchaseItem}, CancellationReason: "", CreatedAt: now, UpdatedAt: now,
	}
	purchaseValue := purchase.Purchase{
		ID: contractID3, Reference: "PURCHASE-1", BuyerID: contractID4, Status: "paid",
		PaymentStatus: "succeeded", PaymentIntentID: contractID1, Currency: "IDR",
		SubtotalMinor: 249000, TotalMinor: 249000, ReservationExpiresAt: now.Add(time.Hour),
		SellerOrders: []purchase.SellerOrder{order}, CreatedAt: now, UpdatedAt: now,
	}
	review := purchase.Review{
		ID: contractID1, BuyerID: contractID4, BuyerName: "Buyer", PurchaseItemID: contractID1,
		ProductID: contractID4, Rating: 5, Title: "Good", Body: "Works", CreatedAt: now, UpdatedAt: now,
	}

	session, err := toContractSession(identity.Session{AccessToken: "access", TokenType: "Bearer", ExpiresIn: 900, User: identity.User{ID: contractID1, Email: "buyer@example.com", DisplayName: "Buyer", Role: identity.RoleBuyer}})
	if err != nil {
		t.Fatal(err)
	}
	store, err := toContractStore(marketstore.Store{ID: contractID1, SellerID: contractID2, Name: "Studio", Slug: "studio", Description: "Store", Status: "approved", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	storeProfile, err := toContractStoreProfile(marketstore.Profile{ID: contractID1, Name: "Studio", Slug: "studio", Description: "Store", SellerDisplayName: "Seller", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	adminUser, err := toContractAdminUser(identity.AdminUser{ID: contractID2, Email: "seller@example.com", DisplayName: "Seller", Role: identity.RoleSeller, Status: "active", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	productResponse, err := toContractProduct(product)
	if err != nil {
		t.Fatal(err)
	}
	page, err := toContractProductPage(catalog.Page{Items: []catalog.Summary{{ID: contractID4, Name: "Lamp", Slug: "lamp", StoreName: "Studio", StoreSlug: "studio", Category: category, MinPriceMinor: 249000, Currency: "IDR", InStock: true, ImageURL: "/images/lamp.webp"}}, Page: 1, PageSize: 20, Total: 1})
	if err != nil {
		t.Fatal(err)
	}
	cart, err := toContractCart(purchase.Cart{ID: contractID1, Currency: "IDR", Stores: []purchase.CartStore{{StoreID: contractID2, StoreName: "Studio", Items: []purchase.CartItem{{VariantID: contractID2, ProductID: contractID4, ProductName: "Lamp", ProductSlug: "lamp", VariantName: "Brass", SKU: "LAMP-01", ImageURL: "/images/lamp.webp", Quantity: 1, AvailableStock: 2, UnitPriceMinor: 249000, LineTotalMinor: 249000, Currency: "IDR"}}, SubtotalMinor: 249000}}, SubtotalMinor: 249000, TotalQuantity: 1, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	purchaseResponse, err := toContractPurchase(purchaseValue)
	if err != nil {
		t.Fatal(err)
	}
	reviewSummary, err := toContractReviewSummary(purchase.ReviewSummary{Average: 5, Count: 1, Items: []purchase.Review{review}})
	if err != nil {
		t.Fatal(err)
	}
	overview, err := toContractAdminOverview(purchase.AdminOverview{Users: 1, ApprovedStores: 1, PublishedProducts: 1, Purchases: 1, ActiveSellerOrders: 1, DeliveredOrders: 0, GrossMerchandiseMinor: 249000, Currency: "IDR"})
	if err != nil {
		t.Fatal(err)
	}
	audit, err := toContractAuditEvent(purchase.AuditEvent{ID: contractID1, ActorID: contractID2, ActorName: "Admin", ActorRole: "admin", Action: "approve", ResourceType: "store", ResourceID: contractID3, Data: map[string]any{"status": "approved"}, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	notification, err := toContractNotification(purchase.Notification{ID: contractID1, Kind: "purchase.paid", Title: "Paid", Body: "Payment received", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		value any
	}{
		{"Session", session}, {"Store", store}, {"StoreProfile", storeProfile}, {"AdminUser", adminUser}, {"ProductDetail", productResponse},
		{"ProductPage", page}, {"Cart", cart}, {"Purchase", purchaseResponse},
		{"ReviewSummary", reviewSummary}, {"AdminOverview", overview},
		{"AuditEvent", audit}, {"Notification", notification},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { validateContractComponent(t, test.name, test.value) })
	}
}

func TestRequestMappersPreserveDomainInputs(t *testing.T) {
	id := uuid.MustParse(contractID1)
	productInput := toDomainProductInput(contract.ProductInput{CategoryId: id, Name: "Lamp", Slug: "lamp", Description: "Desk lamp"})
	if productInput.CategoryID != contractID1 || productInput.Name != "Lamp" || productInput.Slug != "lamp" || productInput.Description != "Desk lamp" {
		t.Fatalf("product input = %#v", productInput)
	}
	reviewInput := toDomainReviewInput(contract.ReviewInput{PurchaseItemId: id, Rating: 5, Title: "Good", Body: "Works"})
	if reviewInput.PurchaseItemID != contractID1 || reviewInput.Rating != 5 || reviewInput.Title != "Good" || reviewInput.Body != "Works" {
		t.Fatalf("review input = %#v", reviewInput)
	}
}

func TestResponseMappersRejectInvalidDomainOutput(t *testing.T) {
	tests := []struct {
		name     string
		mapValue func() error
	}{
		{"user UUID", func() error {
			_, err := toContractUser(identity.User{ID: "not-a-uuid", Role: identity.RoleBuyer})
			return err
		}},
		{"user role", func() error {
			_, err := toContractUser(identity.User{ID: contractID1, Role: identity.Role("operator")})
			return err
		}},
		{"store status", func() error {
			_, err := toContractStore(marketstore.Store{ID: contractID1, SellerID: contractID2, Status: "disabled"})
			return err
		}},
		{"variant currency", func() error {
			_, err := toContractVariant(catalog.Variant{ID: contractID1, Currency: "USD"})
			return err
		}},
		{"purchase status", func() error {
			_, err := toContractPurchase(purchase.Purchase{ID: contractID1, BuyerID: contractID2, Status: "unknown"})
			return err
		}},
		{"audit actor UUID", func() error {
			_, err := toContractAuditEvent(purchase.AuditEvent{ID: contractID1, ResourceID: contractID2, ActorID: "bad", ActorRole: "system"})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.mapValue(); err == nil {
				t.Fatal("expected mapping error")
			}
		})
	}
}
