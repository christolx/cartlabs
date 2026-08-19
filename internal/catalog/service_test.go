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
	product    Product
	variant    Variant
	filters    Filters
	candidates []string
}

type fakeCandidateSearcher struct {
	ids []string
	err error
}

func (f fakeCandidateSearcher) Search(context.Context, string, int) ([]string, error) {
	return f.ids, f.err
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
func (f *fakeRepository) AddImage(_ context.Context, _, _ string, i ProductImage, _ string) (ProductImage, error) {
	return i, nil
}
func (f *fakeRepository) ReplaceImage(_ context.Context, _, _, _ string, i ProductImage, _ string) (ProductImage, error) {
	return i, nil
}
func (f *fakeRepository) DeleteImage(context.Context, string, string, string, string) error {
	return nil
}
func (f *fakeRepository) Publish(_ context.Context, _ string, _ string, _ time.Time) (Product, error) {
	f.product.Status = "published"
	return f.product, nil
}
func (f *fakeRepository) Archive(_ context.Context, _ string, _ string, _ time.Time) (Product, error) {
	f.product.Status = "archived"
	return f.product, nil
}
func (f *fakeRepository) AdjustInventory(_ context.Context, _ string, _ string, delta int, _ string, _ string, _ time.Time) (Variant, error) {
	f.variant.Stock += delta
	return f.variant, nil
}
func (f *fakeRepository) ListForAdmin(context.Context) ([]Product, error) {
	return []Product{f.product}, nil
}
func (f *fakeRepository) UpdateStatus(_ context.Context, _ string, status, _, _ string, _ time.Time) (Product, error) {
	f.product.Status = status
	return f.product, nil
}
func (f *fakeRepository) ListPublic(_ context.Context, filters Filters) (Page, error) {
	f.filters = filters
	return Page{Items: []Summary{}, Page: filters.Page, PageSize: filters.PageSize}, nil
}
func (f *fakeRepository) ListPublicCandidates(_ context.Context, filters Filters, candidates []string) (Page, error) {
	f.filters = filters
	f.candidates = candidates
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
	position := 0
	image, err := service.AddImage(context.Background(), seller, product.ID, ImageInput{URL: "/images/basket.webp", AltText: "Basket", Position: &position})
	if err != nil || image.URL != "/images/basket.webp" {
		t.Fatalf("image = %#v, %v", image, err)
	}
	buyer := identity.Principal{UserID: "buyer", Role: identity.RoleBuyer}
	if _, err := service.Create(context.Background(), buyer, ProductInput{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("buyer error = %v", err)
	}
}

func TestImageURLValidation(t *testing.T) {
	productID := "01989f00-0000-7000-8000-000000000201"
	seller := identity.Principal{UserID: "seller", Role: identity.RoleSeller}
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "local seed", url: "/images/catalog-v2/basket.webp"},
		{name: "configured cloud", url: "https://res.cloudinary.com/demo-cloud/image/upload/v1/basket.webp"},
		{name: "wrong cloud", url: "https://res.cloudinary.com/other/image/upload/v1/basket.webp", wantErr: true},
		{name: "wrong host", url: "https://example.com/demo-cloud/image/upload/v1/basket.webp", wantErr: true},
		{name: "insecure", url: "http://res.cloudinary.com/demo-cloud/image/upload/v1/basket.webp", wantErr: true},
		{name: "other local path", url: "/uploads/basket.webp", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeRepository{product: Product{ID: productID}}, WithCloudinaryCloudName("demo-cloud"))
			_, err := service.AddImage(context.Background(), seller, productID, ImageInput{URL: test.url, AltText: "Basket"})
			if test.wantErr && !errors.Is(err, domain.ErrInvalid) {
				t.Fatalf("error = %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCatalogFiltersRejectInvalidRange(t *testing.T) {
	service := NewService(&fakeRepository{})
	min, max := int64(200), int64(100)
	if _, err := service.ListPublic(context.Background(), Filters{Page: 1, PageSize: 20, MinPrice: &min, MaxPrice: &max}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
}

func TestSellerProductDetailOwnershipAndValidation(t *testing.T) {
	id := "01989f00-0000-7000-8000-000000000201"
	service := NewService(&fakeRepository{product: Product{ID: id}})
	product, err := service.FindOwn(context.Background(), identity.Principal{UserID: "seller", Role: identity.RoleSeller}, id)
	if err != nil || product.ID != id {
		t.Fatalf("product = %#v, err = %v", product, err)
	}
	if _, err := service.FindOwn(context.Background(), identity.Principal{UserID: "other", Role: identity.RoleSeller}, id); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-owner error = %v", err)
	}
	if _, err := service.FindOwn(context.Background(), identity.Principal{Role: identity.RoleBuyer}, id); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("buyer error = %v", err)
	}
	if _, err := service.FindOwn(context.Background(), identity.Principal{UserID: "seller", Role: identity.RoleSeller}, "bad"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid ID error = %v", err)
	}
}

func TestCatalogStoreFilterValidation(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	if _, err := service.ListPublic(context.Background(), Filters{StoreSlug: " Bad Store ", Page: 1, PageSize: 20}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid slug error = %v", err)
	}
	if _, err := service.ListPublic(context.Background(), Filters{StoreSlug: " NUSANTARA-GOODS ", Page: 1, PageSize: 20}); err != nil {
		t.Fatal(err)
	}
	if repository.filters.StoreSlug != "nusantara-goods" {
		t.Fatalf("store slug = %q", repository.filters.StoreSlug)
	}
}

func TestCatalogSearchUsesCandidatesAndFallsBack(t *testing.T) {
	repository := &fakeRepository{}
	results := []string{}
	service := NewService(repository, WithCandidateSearcher(fakeCandidateSearcher{ids: []string{"01989f00-0000-7000-8000-000000000201"}}),
		WithSearchObserver(func(result string) { results = append(results, result) }))
	if _, err := service.ListPublic(context.Background(), Filters{Search: "basket", StoreSlug: "nusantara-goods", Page: 1, PageSize: 20}); err != nil {
		t.Fatal(err)
	}
	if len(repository.candidates) != 1 || results[0] != "service" || repository.filters.StoreSlug != "nusantara-goods" {
		t.Fatalf("candidates=%v results=%v", repository.candidates, results)
	}

	repository.candidates = nil
	service = NewService(repository, WithCandidateSearcher(fakeCandidateSearcher{err: errors.New("search unavailable")}),
		WithSearchObserver(func(result string) { results = append(results, result) }))
	if _, err := service.ListPublic(context.Background(), Filters{Search: "basket", StoreSlug: "nusantara-goods", Page: 1, PageSize: 20}); err != nil {
		t.Fatal(err)
	}
	if repository.filters.Search != "basket" || repository.filters.StoreSlug != "nusantara-goods" || results[len(results)-1] != "fallback" {
		t.Fatalf("filters=%#v results=%v", repository.filters, results)
	}
}

func TestProductEditsPreserveLifecycleStatus(t *testing.T) {
	id := "01989f00-0000-7000-8000-000000000201"
	categoryID := "01989f00-0000-7000-8000-000000000101"
	repository := &fakeRepository{product: Product{ID: id, Status: "suspended"}}
	service := NewService(repository)
	updated, err := service.Update(context.Background(), identity.Principal{UserID: "seller", Role: identity.RoleSeller}, id,
		ProductInput{CategoryID: categoryID, Name: "Updated Product", Slug: "updated-product", Description: "Description"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "suspended" {
		t.Fatalf("status = %q", updated.Status)
	}
}

func TestProductLifecycleAuthorizationAndValidation(t *testing.T) {
	id := "01989f00-0000-7000-8000-000000000201"
	seller := identity.Principal{UserID: "seller", Role: identity.RoleSeller}
	admin := identity.Principal{UserID: "admin", Role: identity.RoleAdmin}
	service := NewService(&fakeRepository{product: Product{ID: id, Status: "published"}})

	archived, err := service.Archive(context.Background(), seller, id)
	if err != nil || archived.Status != "archived" {
		t.Fatalf("archived=%#v err=%v", archived, err)
	}
	if _, err := service.Archive(context.Background(), admin, id); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("admin archive error = %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), seller, id, "suspended", "policy"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("seller enforcement error = %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), admin, id, "suspended", "   "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("blank reason error = %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), admin, id, "archived", "policy"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid status error = %v", err)
	}
}
