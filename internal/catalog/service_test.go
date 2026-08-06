package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

type fakeRepository struct {
	product Product
	variant Variant
	filters Filters
}

func (f *fakeRepository) Categories(context.Context) ([]Category, error) {
	return []Category{{ID: "01989f00-0000-7000-8000-000000000101", Name: "Home", Slug: "home"}}, nil
}
func (f *fakeRepository) ListForSeller(context.Context, string) ([]Product, error) {
	return []Product{f.product}, nil
}
func (f *fakeRepository) FindForSeller(_ context.Context, sellerID, id string) (Product, error) {
	if sellerID != "seller" || f.product.ID != id {
		return Product{}, domain.ErrNotFound
	}
	return f.product, nil
}
func (f *fakeRepository) Create(_ context.Context, _ string, p Product, _ string) (Product, error) {
	f.product = p
	return p, nil
}
func (f *fakeRepository) Update(_ context.Context, p Product, _, _ string) (Product, error) {
	f.product = p
	return p, nil
}
func (f *fakeRepository) AddVariant(_ context.Context, _ string, v Variant, _ string) (Variant, error) {
	f.variant = v
	return v, nil
}
func (f *fakeRepository) AddImage(_ context.Context, _ string, i ProductImage, _ string) (ProductImage, error) {
	return i, nil
}
func (f *fakeRepository) Publish(_ context.Context, _ string, _ string, _ time.Time) (Product, error) {
	f.product.Status = "published"
	return f.product, nil
}
func (f *fakeRepository) AdjustInventory(_ context.Context, _ string, _ string, delta int, _ string, _ string, _ time.Time) (Variant, error) {
	f.variant.Stock += delta
	return f.variant, nil
}
func (f *fakeRepository) ListForAdmin(context.Context) ([]Product, error) {
	return []Product{f.product}, nil
}
func (f *fakeRepository) Moderate(_ context.Context, _ string, status, note, _ string, _ time.Time) (Product, error) {
	f.product.ModerationStatus = status
	f.product.ModerationNote = note
	return f.product, nil
}
func (f *fakeRepository) ListPublic(_ context.Context, filters Filters) (Page, error) {
	f.filters = filters
	return Page{Items: []Summary{}, Page: filters.Page, PageSize: filters.PageSize}, nil
}
func (f *fakeRepository) FindPublic(context.Context, string) (Product, error) { return f.product, nil }

func TestCatalogSellerOwnershipAndValidation(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	seller := identity.Principal{UserID: "seller", Role: identity.RoleSeller}
	categoryID := "01989f00-0000-7000-8000-000000000101"
	product, err := service.Create(context.Background(), seller, ProductInput{CategoryID: categoryID, Name: " Market Basket ", Slug: "market-basket", Description: "Woven"})
	if err != nil {
		t.Fatal(err)
	}
	if product.Status != "draft" || product.Name != "Market Basket" {
		t.Fatalf("product = %#v", product)
	}
	variant, err := service.AddVariant(context.Background(), seller, product.ID, VariantInput{SKU: "BASKET-1", Name: "Natural", Attributes: map[string]string{"color": "Natural"}, PriceMinor: 249000, Currency: "idr", Stock: 3})
	if err != nil {
		t.Fatal(err)
	}
	if variant.Currency != "IDR" || variant.Stock != 3 {
		t.Fatalf("variant = %#v", variant)
	}
	image, err := service.AddImage(context.Background(), seller, product.ID, ImageInput{URL: "/images/basket.webp", AltText: "Basket", Position: 0})
	if err != nil || image.URL != "/images/basket.webp" {
		t.Fatalf("image = %#v, %v", image, err)
	}
	buyer := identity.Principal{UserID: "buyer", Role: identity.RoleBuyer}
	if _, err := service.Create(context.Background(), buyer, ProductInput{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("buyer error = %v", err)
	}
}

func TestCatalogFiltersRejectInvalidRange(t *testing.T) {
	service := NewService(&fakeRepository{})
	min, max := int64(200), int64(100)
	if _, err := service.ListPublic(context.Background(), Filters{Page: 1, PageSize: 20, MinPrice: &min, MaxPrice: &max}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
}
