//go:build integration

package catalog

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestProductWriteAtomicallyCreatesSearchEvent(t *testing.T) {
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

	productID := uuid.NewString()
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM outbox_events WHERE aggregate_id=$1", productID)
		_, _ = pool.Exec(ctx, "DELETE FROM products WHERE id=$1", productID)
	}()
	now := time.Now().UTC()
	created, err := NewPostgresRepository(pool).Create(ctx, "01989f00-0000-7000-8000-000000000004", Product{
		ID: productID, Category: Category{ID: "01989f00-0000-7000-8000-000000000102"},
		Name: "Atomic Search Event", Slug: "atomic-search-event-" + productID[:8], Description: "Integration proof.", CreatedAt: now, UpdatedAt: now,
	}, "01989f00-0000-7000-8000-000000000004")
	if err != nil {
		t.Fatal(err)
	}
	var eventType string
	var payload []byte
	if err := pool.QueryRow(ctx, `SELECT event_type,payload FROM outbox_events WHERE aggregate_id=$1`, productID).Scan(&eventType, &payload); err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		t.Fatal(err)
	}
	if eventType != "catalog.search.upsert.v1" || data["productId"] != created.ID || data["categorySlug"] != "home-living" {
		t.Fatalf("eventType=%q payload=%s", eventType, payload)
	}
}
