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

func TestPublicCatalogStoreFilterPreservesModerationRules(t *testing.T) {
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
	sellerID, storeID, categoryID, productID, variantID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	storeSlug := "completion-store-" + storeID[:8]
	productSlug := "completion-product-" + productID[:8]
	categorySlug := "completion-category-" + categoryID[:8]
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM product_variants WHERE id=$1`, variantID)
		_, _ = pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, productID)
		_, _ = pool.Exec(ctx, `DELETE FROM stores WHERE id=$1`, storeID)
		_, _ = pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, categoryID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, sellerID)
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES ($1,$2,'hash','Catalog Seller','seller','active')`, sellerID, sellerID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO stores (id,seller_id,name,slug,description,status) VALUES ($1,$2,'Completion Store',$3,'Description','approved')`, storeID, sellerID, storeSlug); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO categories (id,name,slug) VALUES ($1,$2,$3)`, categoryID, "Completion "+categoryID[:8], categorySlug); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO products (id,store_id,category_id,name,slug,description,status,moderation_status) VALUES ($1,$2,$3,'Completion Basket',$4,'Searchable completion item','published','approved')`, productID, storeID, categoryID, productSlug); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO product_variants (id,product_id,sku,name,price_minor,currency,stock,active) VALUES ($1,$2,$3,'Default',1000,'IDR',2,true)`, variantID, productID, "CMP-"+variantID[:8]); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgresRepository(pool)
	filters := Filters{Search: "completion", StoreSlug: storeSlug, CategorySlug: categorySlug, InStock: true, Page: 1, PageSize: 20}
	assertVisible := func(t *testing.T, page Page, err error) {
		t.Helper()
		if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != productID || page.Items[0].StoreSlug != storeSlug {
			t.Fatalf("page=%#v err=%v", page, err)
		}
	}
	page, err := repository.ListPublic(ctx, filters)
	assertVisible(t, page, err)
	page, err = repository.ListPublicCandidates(ctx, filters, []string{productID})
	assertVisible(t, page, err)

	if _, err := pool.Exec(ctx, `UPDATE stores SET status='pending' WHERE id=$1`, storeID); err != nil {
		t.Fatal(err)
	}
	page, err = repository.ListPublic(ctx, filters)
	if err != nil || page.Total != 0 {
		t.Fatalf("pending store page=%#v err=%v", page, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE stores SET status='approved' WHERE id=$1`, storeID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE products SET moderation_status='pending' WHERE id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	page, err = repository.ListPublicCandidates(ctx, filters, []string{productID})
	if err != nil || page.Total != 0 {
		t.Fatalf("pending product page=%#v err=%v", page, err)
	}
}
