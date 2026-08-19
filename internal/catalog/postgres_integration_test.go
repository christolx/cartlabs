//go:build integration

package catalog

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

func TestProductLifecycleEnforcesOwnershipCompletenessAndAdminSuspension(t *testing.T) {
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

	sellerID, adminID := uuid.NewString(), uuid.NewString()
	storeID, categoryID, productID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	variantID, imageID := uuid.NewString(), uuid.NewString()
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM outbox_events WHERE aggregate_id=$1`, productID)
		_, _ = pool.Exec(ctx, `DELETE FROM audit_log WHERE resource_id=ANY($1::uuid[])`, []string{productID, variantID, storeID})
		_, _ = pool.Exec(ctx, `DELETE FROM inventory_ledger WHERE variant_id=$1`, variantID)
		_, _ = pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, productID)
		_, _ = pool.Exec(ctx, `DELETE FROM stores WHERE id=$1`, storeID)
		_, _ = pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, categoryID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=ANY($1::uuid[])`, []string{sellerID, adminID})
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES ($1,$2,'hash','Lifecycle Seller','seller','active'),($3,$4,'hash','Lifecycle Admin','admin','active')`, sellerID, sellerID+"@example.com", adminID, adminID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO stores (id,seller_id,name,slug,description,status) VALUES ($1,$2,'Lifecycle Store',$3,'Description','pending')`, storeID, sellerID, "lifecycle-"+storeID[:8]); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO categories (id,name,slug) VALUES ($1,$2,$3)`, categoryID, "Lifecycle "+categoryID[:8], "lifecycle-"+categoryID[:8]); err != nil {
		t.Fatal(err)
	}

	repository := NewPostgresRepository(pool)
	now := time.Now().UTC()
	input := Product{ID: productID, Category: Category{ID: categoryID}, Name: "Lifecycle Product", Slug: "lifecycle-" + productID[:8], Description: "Description", CreatedAt: now, UpdatedAt: now}
	if _, err := repository.Create(ctx, sellerID, input, sellerID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unverified store create error = %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE stores SET status='approved' WHERE id=$1`, storeID); err != nil {
		t.Fatal(err)
	}
	product, err := repository.Create(ctx, sellerID, input, sellerID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Publish(ctx, product.ID, sellerID, now.Add(time.Second)); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("incomplete publish error = %v", err)
	}
	if _, err := repository.AddVariant(ctx, product.ID, Variant{ID: variantID, SKU: "LIFE-" + variantID[:8], Name: "Default", Attributes: map[string]string{}, PriceMinor: 1000, Currency: "IDR", Stock: 2, Active: true}, sellerID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.AddImage(ctx, sellerID, product.ID, ProductImage{ID: imageID, URL: "/images/shared-product.webp", AltText: "Product", Position: 0}, sellerID); err != nil {
		t.Fatal(err)
	}
	product, err = repository.Publish(ctx, product.ID, sellerID, now.Add(2*time.Second))
	if err != nil || product.Status != "published" {
		t.Fatalf("published=%#v err=%v", product, err)
	}
	if err := repository.DeleteImage(ctx, sellerID, product.ID, imageID, sellerID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("delete last published image error = %v", err)
	}
	secondImageID := uuid.NewString()
	if _, err := repository.AddImage(ctx, sellerID, product.ID, ProductImage{ID: secondImageID, URL: "/images/shared-product.webp", AltText: "Second", Position: 1}, sellerID); err != nil {
		t.Fatal(err)
	}
	replaced, err := repository.ReplaceImage(ctx, sellerID, product.ID, secondImageID, ProductImage{URL: "/images/catalog-v2/studio-tray.webp", AltText: "Replacement"}, sellerID)
	if err != nil || replaced.Position != 1 || replaced.AltText != "Replacement" {
		t.Fatalf("replaced=%#v err=%v", replaced, err)
	}
	if err := repository.DeleteImage(ctx, sellerID, product.ID, imageID, sellerID); err != nil {
		t.Fatal(err)
	}
	var compactedPosition int
	if err := pool.QueryRow(ctx, `SELECT position FROM product_images WHERE id=$1`, secondImageID).Scan(&compactedPosition); err != nil || compactedPosition != 0 {
		t.Fatalf("compacted position=%d err=%v", compactedPosition, err)
	}
	for position := 1; position < 8; position++ {
		if _, err := repository.AddImage(ctx, sellerID, product.ID, ProductImage{ID: uuid.NewString(), URL: "/images/shared-product.webp", AltText: "Extra", Position: position}, sellerID); err != nil {
			t.Fatalf("add image %d: %v", position, err)
		}
	}
	if _, err := repository.AddImage(ctx, sellerID, product.ID, ProductImage{ID: uuid.NewString(), URL: "/images/shared-product.webp", AltText: "Ninth", Position: 8}, sellerID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ninth image error = %v", err)
	}
	if _, err := repository.AddImage(ctx, uuid.NewString(), product.ID, ProductImage{ID: uuid.NewString(), URL: "/images/shared-product.webp", AltText: "Other seller", Position: 0}, sellerID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other seller add error = %v", err)
	}
	product, err = repository.UpdateStatus(ctx, product.ID, "suspended", "policy violation", adminID, now.Add(3*time.Second))
	if err != nil || product.Status != "suspended" || product.EnforcementReason != "policy violation" || product.EnforcedBy == nil || *product.EnforcedBy != "Lifecycle Admin" {
		t.Fatalf("suspended=%#v err=%v", product, err)
	}
	adminProducts, err := repository.ListForAdmin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	foundAdminProduct := false
	for _, item := range adminProducts {
		if item.ID == product.ID {
			foundAdminProduct = item.Status == "suspended" && item.EnforcementReason == "policy violation" && item.EnforcedAt != nil
		}
	}
	if !foundAdminProduct {
		t.Fatalf("admin products missing suspension context: %#v", adminProducts)
	}
	if _, err := repository.Publish(ctx, product.ID, sellerID, now.Add(4*time.Second)); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("seller suspension bypass error = %v", err)
	}
	product, err = repository.UpdateStatus(ctx, product.ID, "published", "corrected", adminID, now.Add(5*time.Second))
	if err != nil || product.Status != "published" || product.EnforcementReason != "policy violation" {
		t.Fatalf("reinstated=%#v err=%v", product, err)
	}
	product, err = repository.Archive(ctx, product.ID, sellerID, now.Add(6*time.Second))
	if err != nil || product.Status != "archived" {
		t.Fatalf("archived=%#v err=%v", product, err)
	}
	product, err = repository.Publish(ctx, product.ID, sellerID, now.Add(7*time.Second))
	if err != nil || product.Status != "published" {
		t.Fatalf("republished=%#v err=%v", product, err)
	}
	var metadata map[string]string
	var raw []byte
	if err := pool.QueryRow(ctx, `SELECT metadata FROM audit_log WHERE resource_type='product' AND resource_id=$1 AND action='product.status.suspended' ORDER BY created_at DESC LIMIT 1`, product.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &metadata); err != nil || metadata["from"] != "published" || metadata["to"] != "suspended" || metadata["reason"] != "policy violation" {
		t.Fatalf("metadata=%v err=%v", metadata, err)
	}
	var searchEvents int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id=$1 AND event_type='catalog.search.upsert.v1'`, product.ID).Scan(&searchEvents); err != nil || searchEvents != 6 {
		t.Fatalf("search events=%d err=%v", searchEvents, err)
	}
}

func TestPublicCatalogStoreFilterPreservesListingEligibility(t *testing.T) {
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
	if _, err := pool.Exec(ctx, `INSERT INTO products (id,store_id,category_id,name,slug,description,status) VALUES ($1,$2,$3,'Completion Basket',$4,'Searchable completion item','published')`, productID, storeID, categoryID, productSlug); err != nil {
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
	if _, err := pool.Exec(ctx, `UPDATE products SET status='suspended' WHERE id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	page, err = repository.ListPublicCandidates(ctx, filters, []string{productID})
	if err != nil || page.Total != 0 {
		t.Fatalf("suspended product page=%#v err=%v", page, err)
	}
}
