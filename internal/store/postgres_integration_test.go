//go:build integration

package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

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
	sellerID, adminID, storeID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	slug := "public-profile-" + storeID[:8]
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_log WHERE resource_type='store' AND resource_id=$1`, storeID)
		_, _ = pool.Exec(ctx, `DELETE FROM stores WHERE id=$1`, storeID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=ANY($1::uuid[])`, []string{sellerID, adminID})
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES ($1,$2,'hash','Public Seller','seller','active'),($3,$4,'hash','Verification Admin','admin','active')`, sellerID, sellerID+"@example.com", adminID, adminID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO stores (id,seller_id,name,slug,description,status) VALUES ($1,$2,'Public Store',$3,'Public description','pending')`, storeID, sellerID, slug); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgresRepository(pool)
	if _, err := repository.FindPublicBySlug(ctx, slug); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending error = %v", err)
	}
	if _, err := repository.Moderate(ctx, storeID, "rejected", "profile mismatch", adminID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindPublicBySlug(ctx, slug); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("rejected error = %v", err)
	}
	var raw []byte
	if err := pool.QueryRow(ctx, `SELECT metadata FROM audit_log WHERE resource_type='store' AND resource_id=$1 AND action='store.verified.rejected' ORDER BY created_at DESC LIMIT 1`, storeID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var metadata map[string]string
	if err := json.Unmarshal(raw, &metadata); err != nil || metadata["from"] != "pending" || metadata["to"] != "rejected" || metadata["note"] != "profile mismatch" {
		t.Fatalf("metadata=%v err=%v", metadata, err)
	}
	if _, err := repository.Moderate(ctx, storeID, "approved", "", adminID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	profile, err := repository.FindPublicBySlug(ctx, slug)
	if err != nil || profile.ID != storeID || profile.SellerDisplayName != "Public Seller" {
		t.Fatalf("profile=%#v err=%v", profile, err)
	}
}
