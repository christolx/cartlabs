//go:build integration

package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresIndexIsIdempotentAndRejectsOlderUpdates(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_SEARCH_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_SEARCH_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repository := NewPostgresRepository(pool)
	productID, eventID := uuid.NewString(), uuid.NewString()
	existingRows, err := pool.Query(ctx, "SELECT product_id::text FROM search_documents")
	if err != nil {
		t.Fatal(err)
	}
	existingIDs := make([]string, 0)
	for existingRows.Next() {
		var id string
		if err := existingRows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		existingIDs = append(existingIDs, id)
	}
	if err := existingRows.Err(); err != nil {
		t.Fatal(err)
	}
	existingRows.Close()
	now := time.Now().UTC()
	document := Document{ProductID: productID, Name: "Learning Boundary Basket", Description: "Searchable woven basket", CategorySlug: "home", UpdatedAt: now}
	applied, err := repository.Upsert(ctx, eventID, document)
	if err != nil || !applied {
		t.Fatalf("first upsert applied=%v err=%v", applied, err)
	}
	applied, err = repository.Upsert(ctx, eventID, document)
	if err != nil || applied {
		t.Fatalf("duplicate upsert applied=%v err=%v", applied, err)
	}
	older := document
	older.Name, older.UpdatedAt = "Stale Name", now.Add(-time.Hour)
	if _, err := repository.Upsert(ctx, uuid.NewString(), older); err != nil {
		t.Fatal(err)
	}
	ids, err := repository.Search(ctx, "Learning Boundary", 10)
	if err != nil || len(ids) != 1 || ids[0] != productID {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	if ids, err := repository.Search(ctx, "Stale Name", 10); err != nil || len(ids) != 0 {
		t.Fatalf("stale ids=%v err=%v", ids, err)
	}
	staleID, concurrentID := uuid.NewString(), uuid.NewString()
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM search_documents WHERE product_id=ANY($1::uuid[])", []string{productID, staleID, concurrentID})
	}()
	if _, err := repository.Upsert(ctx, uuid.NewString(), Document{ProductID: staleID, Name: "Prune Stale", CategorySlug: "home", UpdatedAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Upsert(ctx, uuid.NewString(), Document{ProductID: concurrentID, Name: "Concurrent Write", CategorySlug: "home", UpdatedAt: now.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	keepIDs := append(existingIDs, productID)
	deleted, err := repository.Prune(ctx, keepIDs, now)
	if err != nil || deleted < 1 {
		t.Fatalf("prune deleted=%d err=%v", deleted, err)
	}
	if ids, err := repository.Search(ctx, "Prune Stale", 10); err != nil || len(ids) != 0 {
		t.Fatalf("pruned ids=%v err=%v", ids, err)
	}
	if ids, err := repository.Search(ctx, "Concurrent Write", 10); err != nil || !containsID(ids, concurrentID) {
		t.Fatalf("concurrent ids=%v err=%v", ids, err)
	}
}

func containsID(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
