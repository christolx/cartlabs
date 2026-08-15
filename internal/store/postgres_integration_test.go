//go:build integration

package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPublicStoreProfileOnlyReturnsApprovedStore(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	sellerID, storeID := uuid.NewString(), uuid.NewString()
	slug := "public-profile-" + storeID[:8]
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM stores WHERE id=$1`, storeID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, sellerID)
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES ($1,$2,'hash','Public Seller','seller','active')`, sellerID, sellerID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO stores (id,seller_id,name,slug,description,status) VALUES ($1,$2,'Public Store',$3,'Public description','pending')`, storeID, sellerID, slug); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgresRepository(pool)
	if _, err := repository.FindPublicBySlug(ctx, slug); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending error = %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE stores SET status='rejected' WHERE id=$1`, storeID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindPublicBySlug(ctx, slug); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("rejected error = %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE stores SET status='approved' WHERE id=$1`, storeID); err != nil {
		t.Fatal(err)
	}
	profile, err := repository.FindPublicBySlug(ctx, slug)
	if err != nil || profile.ID != storeID || profile.SellerDisplayName != "Public Seller" {
		t.Fatalf("profile=%#v err=%v", profile, err)
	}
}
