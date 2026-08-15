package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

type fakeRepository struct{ value Store }

func (f *fakeRepository) FindBySeller(_ context.Context, sellerID string) (Store, error) {
	if f.value.SellerID != sellerID {
		return Store{}, domain.ErrNotFound
	}
	return f.value, nil
}
func (f *fakeRepository) Create(_ context.Context, value Store, _ string) (Store, error) {
	if f.value.ID != "" {
		return Store{}, domain.ErrConflict
	}
	f.value = value
	return value, nil
}
func (f *fakeRepository) Update(_ context.Context, value Store, _ string) (Store, error) {
	f.value = value
	return value, nil
}
func (f *fakeRepository) List(context.Context) ([]Store, error) { return []Store{f.value}, nil }
func (f *fakeRepository) Moderate(_ context.Context, id, status, note, _ string, _ time.Time) (Store, error) {
	if f.value.ID != id {
		return Store{}, domain.ErrNotFound
	}
	f.value.Status = status
	f.value.ModerationNote = note
	return f.value, nil
}
func (f *fakeRepository) FindPublicBySlug(_ context.Context, slug string) (Profile, error) {
	if f.value.Slug != slug || f.value.Status != "approved" {
		return Profile{}, domain.ErrNotFound
	}
	return Profile{ID: f.value.ID, Name: f.value.Name, Slug: f.value.Slug, Description: f.value.Description, CreatedAt: f.value.CreatedAt}, nil
}

func TestSellerStoreLifecycleAndRBAC(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	seller := identity.Principal{UserID: "seller-1", Role: identity.RoleSeller}
	created, err := service.Create(context.Background(), seller, Input{Name: " North Star ", Slug: "north-star", Description: " Demo store "})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != "pending" || created.Name != "North Star" {
		t.Fatalf("created = %#v", created)
	}
	created.Status = "approved"
	repository.value = created
	updated, err := service.Update(context.Background(), seller, Input{Name: "North Star Goods", Slug: "north-star-goods", Description: "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "pending" {
		t.Fatalf("status = %q", updated.Status)
	}
	if _, err := service.ListForAdmin(context.Background(), seller); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("seller admin access error = %v", err)
	}
	admin := identity.Principal{UserID: "admin-1", Role: identity.RoleAdmin}
	approved, err := service.Moderate(context.Background(), admin, updated.ID, "approved", "verified")
	if err != nil || approved.Status != "approved" {
		t.Fatalf("moderate = %#v, %v", approved, err)
	}
}

func TestStoreRejectsInvalidSlug(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Create(context.Background(), identity.Principal{UserID: "seller", Role: identity.RoleSeller}, Input{Name: "Valid", Slug: "Not Valid", Description: ""})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
}

func TestPublicStoreRequiresApprovedValidSlug(t *testing.T) {
	service := NewService(&fakeRepository{value: Store{ID: "store", Name: "North", Slug: "north-star", Status: "pending"}})
	if _, err := service.FindPublic(context.Background(), "north-star"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending error = %v", err)
	}
	service.repository.(*fakeRepository).value.Status = "approved"
	profile, err := service.FindPublic(context.Background(), " NORTH-STAR ")
	if err != nil || profile.Slug != "north-star" {
		t.Fatalf("profile = %#v, err = %v", profile, err)
	}
	if _, err := service.FindPublic(context.Background(), "bad slug"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid error = %v", err)
	}
}
