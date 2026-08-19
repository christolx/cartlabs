package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

type rowScanner interface{ Scan(...any) error }

const productSelect = `
	p.id::text,p.store_id::text,s.name,s.slug,c.id::text,c.name,c.slug,p.name,p.slug,p.description,
	p.status::text,p.created_at,p.updated_at`

func scanProduct(row rowScanner) (Product, error) {
	var value Product
	err := row.Scan(&value.ID, &value.StoreID, &value.StoreName, &value.StoreSlug, &value.Category.ID, &value.Category.Name,
		&value.Category.Slug, &value.Name, &value.Slug, &value.Description, &value.Status,
		&value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, domain.ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("scan product: %w", err)
	}
	value.Variants = []Variant{}
	value.Images = []ProductImage{}
	return value, nil
}

func (r *PostgresRepository) Categories(ctx context.Context) ([]Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,name,slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	items := make([]Category, 0)
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListForSeller(ctx context.Context, sellerID string) ([]Product, error) {
	return r.listProducts(ctx, `SELECT `+productSelect+` FROM products p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id WHERE s.seller_id=$1 ORDER BY p.updated_at DESC`, sellerID)
}

func (r *PostgresRepository) FindForSeller(ctx context.Context, sellerID, productID string) (Product, error) {
	value, err := scanProduct(r.pool.QueryRow(ctx, `SELECT `+productSelect+` FROM products p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id WHERE s.seller_id=$1 AND p.id=$2`, sellerID, productID))
	if err != nil {
		return Product{}, err
	}
	return r.loadDetails(ctx, value)
}

func (r *PostgresRepository) Create(ctx context.Context, sellerID string, value Product, actorID string) (Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin create product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	created, err := scanProduct(tx.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO products (id,store_id,category_id,name,slug,description,status,created_at,updated_at)
			SELECT $1,s.id,$3,$4,$5,$6,'draft',$7,$7 FROM stores s WHERE s.seller_id=$2 AND s.status='approved'
			RETURNING *
		)
		SELECT `+productSelect+` FROM inserted p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id`,
		value.ID, sellerID, value.Category.ID, value.Name, value.Slug, value.Description, value.CreatedAt))
	if errors.Is(err, domain.ErrNotFound) {
		return Product{}, domain.ErrForbidden
	}
	if err != nil {
		return Product{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "product.created", "product", created.ID, created.CreatedAt); err != nil {
		return Product{}, err
	}
	if err := insertSearchOutbox(ctx, tx, created); err != nil {
		return Product{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit create product: %w", err)
	}
	return created, nil
}

func (r *PostgresRepository) Update(ctx context.Context, value Product, sellerID, actorID string) (Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin update product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanProduct(tx.QueryRow(ctx, `
		WITH changed AS (
			UPDATE products p SET category_id=$3,name=$4,slug=$5,description=$6,updated_at=$7
			FROM stores own WHERE p.id=$1 AND p.store_id=own.id AND own.seller_id=$2 RETURNING p.*
		)
		SELECT `+productSelect+` FROM changed p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id`,
		value.ID, sellerID, value.Category.ID, value.Name, value.Slug, value.Description, value.UpdatedAt))
	if err != nil {
		return Product{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "product.updated", "product", updated.ID, updated.UpdatedAt); err != nil {
		return Product{}, err
	}
	if err := insertSearchOutbox(ctx, tx, updated); err != nil {
		return Product{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit update product: %w", err)
	}
	return r.loadDetails(ctx, updated)
}

func (r *PostgresRepository) AddVariant(ctx context.Context, productID string, value Variant, actorID string) (Variant, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Variant{}, fmt.Errorf("begin create variant: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	attributes, err := json.Marshal(value.Attributes)
	if err != nil {
		return Variant{}, domain.ErrInvalid
	}
	created, err := scanVariant(tx.QueryRow(ctx, `
		INSERT INTO product_variants (id,product_id,sku,name,attributes,price_minor,currency,stock,active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true)
		RETURNING id::text,sku,name,attributes,price_minor,currency,stock,active`,
		value.ID, productID, value.SKU, value.Name, attributes, value.PriceMinor, value.Currency, value.Stock))
	if err != nil {
		return Variant{}, mapWriteError(err)
	}
	now := time.Now().UTC()
	if value.Stock > 0 {
		ledgerID, err := uuid.NewV7()
		if err != nil {
			return Variant{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO inventory_ledger (id,variant_id,actor_id,delta,resulting_stock,reason,created_at) VALUES ($1,$2,$3,$4,$4,'initial stock',$5)`,
			ledgerID.String(), value.ID, actorID, value.Stock, now); err != nil {
			return Variant{}, fmt.Errorf("record initial inventory: %w", err)
		}
	}
	if err := insertAudit(ctx, tx, actorID, "variant.created", "product", productID, now); err != nil {
		return Variant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Variant{}, fmt.Errorf("commit create variant: %w", err)
	}
	return created, nil
}

func scanVariant(row rowScanner) (Variant, error) {
	var value Variant
	var attributes []byte
	err := row.Scan(&value.ID, &value.SKU, &value.Name, &attributes, &value.PriceMinor, &value.Currency, &value.Stock, &value.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return Variant{}, domain.ErrNotFound
	}
	if err != nil {
		return Variant{}, fmt.Errorf("scan variant: %w", err)
	}
	if err := json.Unmarshal(attributes, &value.Attributes); err != nil {
		return Variant{}, fmt.Errorf("decode variant attributes: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) AddImage(ctx context.Context, sellerID, productID string, value ProductImage, actorID string) (ProductImage, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ProductImage{}, fmt.Errorf("begin create product image: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := tx.QueryRow(ctx, `SELECT p.id FROM products p JOIN stores s ON s.id=p.store_id WHERE p.id=$1 AND s.seller_id=$2 FOR UPDATE OF p`, productID, sellerID).Scan(&productID); errors.Is(err, pgx.ErrNoRows) {
		return ProductImage{}, domain.ErrNotFound
	} else if err != nil {
		return ProductImage{}, fmt.Errorf("lock product for image add: %w", err)
	}
	var imageCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM product_images WHERE product_id=$1`, productID).Scan(&imageCount); err != nil {
		return ProductImage{}, fmt.Errorf("count product images: %w", err)
	}
	if imageCount >= 8 {
		return ProductImage{}, domain.ErrConflict
	}
	if value.Position < 0 {
		value.Position = imageCount
	} else if value.Position > imageCount {
		return ProductImage{}, domain.ErrInvalid
	}
	var created ProductImage
	err = tx.QueryRow(ctx, `INSERT INTO product_images (id,product_id,url,alt_text,position) VALUES ($1,$2,$3,$4,$5) RETURNING id::text,url,alt_text,position`,
		value.ID, productID, value.URL, value.AltText, value.Position).Scan(&created.ID, &created.URL, &created.AltText, &created.Position)
	if err != nil {
		return ProductImage{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "product.image.created", "product", productID, time.Now().UTC()); err != nil {
		return ProductImage{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProductImage{}, fmt.Errorf("commit create product image: %w", err)
	}
	return created, nil
}

func (r *PostgresRepository) ReplaceImage(ctx context.Context, sellerID, productID, imageID string, value ProductImage, actorID string) (ProductImage, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ProductImage{}, fmt.Errorf("begin replace product image: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var replaced ProductImage
	err = tx.QueryRow(ctx, `
		UPDATE product_images i SET url=$4,alt_text=$5
		FROM products p JOIN stores s ON s.id=p.store_id
		WHERE i.id=$3 AND i.product_id=p.id AND p.id=$2 AND s.seller_id=$1
		RETURNING i.id::text,i.url,i.alt_text,i.position`, sellerID, productID, imageID, value.URL, value.AltText).
		Scan(&replaced.ID, &replaced.URL, &replaced.AltText, &replaced.Position)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductImage{}, domain.ErrNotFound
	}
	if err != nil {
		return ProductImage{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "product.image.replaced", "product", productID, time.Now().UTC()); err != nil {
		return ProductImage{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProductImage{}, fmt.Errorf("commit replace product image: %w", err)
	}
	return replaced, nil
}

func (r *PostgresRepository) DeleteImage(ctx context.Context, sellerID, productID, imageID, actorID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete product image: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status string
	var imageCount int
	if err := tx.QueryRow(ctx, `SELECT p.status::text FROM products p JOIN stores s ON s.id=p.store_id WHERE p.id=$1 AND s.seller_id=$2 FOR UPDATE OF p`, productID, sellerID).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("lock product for image delete: %w", err)
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM product_images WHERE product_id=$1`, productID).Scan(&imageCount); err != nil {
		return fmt.Errorf("count product images: %w", err)
	}
	var deletedPosition int
	if err := tx.QueryRow(ctx, `SELECT position FROM product_images WHERE id=$1 AND product_id=$2`, imageID, productID).Scan(&deletedPosition); errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("find product image for delete: %w", err)
	}
	if status == "published" && imageCount <= 1 {
		return domain.ErrConflict
	}
	if _, err := tx.Exec(ctx, `DELETE FROM product_images WHERE id=$1 AND product_id=$2`, imageID, productID); err != nil {
		return mapWriteError(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE product_images SET position=position+8 WHERE product_id=$1 AND position>$2`, productID, deletedPosition); err != nil {
		return fmt.Errorf("compact product image positions: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE product_images SET position=position-9 WHERE product_id=$1 AND position>$2+8`, productID, deletedPosition); err != nil {
		return fmt.Errorf("finalize product image positions: %w", err)
	}
	if err := insertAudit(ctx, tx, actorID, "product.image.deleted", "product", productID, time.Now().UTC()); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete product image: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Publish(ctx context.Context, productID, actorID string, now time.Time) (Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin publish product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current string
	if err := tx.QueryRow(ctx, `
		SELECT p.status::text
		FROM products p JOIN stores s ON s.id=p.store_id
		WHERE p.id=$1 AND s.seller_id=$2
		FOR UPDATE OF p,s`, productID, actorID).Scan(&current); errors.Is(err, pgx.ErrNoRows) {
		return Product{}, domain.ErrNotFound
	} else if err != nil {
		return Product{}, fmt.Errorf("lock product for publish: %w", err)
	}
	if current != "draft" && current != "archived" {
		return Product{}, domain.ErrConflict
	}
	updated, err := scanProduct(tx.QueryRow(ctx, `
		WITH changed AS (
			UPDATE products p SET status='published',updated_at=$2
			FROM stores own WHERE p.id=$1 AND p.store_id=own.id AND own.seller_id=$3 AND own.status='approved'
			AND p.status IN ('draft','archived')
			AND EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id=p.id AND v.active AND v.stock>0)
			AND EXISTS (SELECT 1 FROM product_images i WHERE i.product_id=p.id)
			RETURNING p.*
		)
		SELECT `+productSelect+` FROM changed p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id`, productID, now, actorID))
	if errors.Is(err, domain.ErrNotFound) {
		return Product{}, domain.ErrConflict
	}
	if err != nil {
		return Product{}, err
	}
	if err := insertAuditMetadata(ctx, tx, actorID, "product.status.published", "product", productID,
		map[string]any{"from": current, "to": "published"}, now); err != nil {
		return Product{}, err
	}
	if err := insertSearchOutbox(ctx, tx, updated); err != nil {
		return Product{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit publish product: %w", err)
	}
	return r.loadDetails(ctx, updated)
}

func (r *PostgresRepository) Archive(ctx context.Context, productID, actorID string, now time.Time) (Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin archive product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanProduct(tx.QueryRow(ctx, `
		WITH changed AS (
			UPDATE products p SET status='archived',updated_at=$3
			FROM stores own
			WHERE p.id=$1 AND p.store_id=own.id AND own.seller_id=$2 AND p.status='published'
			RETURNING p.*
		)
		SELECT `+productSelect+` FROM changed p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id`,
		productID, actorID, now))
	if errors.Is(err, domain.ErrNotFound) {
		return Product{}, domain.ErrConflict
	}
	if err != nil {
		return Product{}, err
	}
	if err := insertAuditMetadata(ctx, tx, actorID, "product.status.archived", "product", productID,
		map[string]any{"from": "published", "to": "archived"}, now); err != nil {
		return Product{}, err
	}
	if err := insertSearchOutbox(ctx, tx, updated); err != nil {
		return Product{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit archive product: %w", err)
	}
	return r.loadDetails(ctx, updated)
}

func (r *PostgresRepository) AdjustInventory(ctx context.Context, variantID, sellerID string, delta int, reason, actorID string, now time.Time) (Variant, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Variant{}, fmt.Errorf("begin adjust inventory: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanVariant(tx.QueryRow(ctx, `
		UPDATE product_variants v SET stock=v.stock+$3,updated_at=$4
		FROM products p,stores s WHERE v.id=$1 AND v.product_id=p.id AND p.store_id=s.id AND s.seller_id=$2 AND v.stock+$3>=0
		RETURNING v.id::text,v.sku,v.name,v.attributes,v.price_minor,v.currency,v.stock,v.active`, variantID, sellerID, delta, now))
	if errors.Is(err, domain.ErrNotFound) {
		return Variant{}, domain.ErrConflict
	}
	if err != nil {
		return Variant{}, err
	}
	ledgerID, err := uuid.NewV7()
	if err != nil {
		return Variant{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO inventory_ledger (id,variant_id,actor_id,delta,resulting_stock,reason,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		ledgerID.String(), variantID, actorID, delta, updated.Stock, reason, now); err != nil {
		return Variant{}, fmt.Errorf("record inventory adjustment: %w", err)
	}
	if err := insertAudit(ctx, tx, actorID, "inventory.adjusted", "product_variant", variantID, now); err != nil {
		return Variant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Variant{}, fmt.Errorf("commit inventory adjustment: %w", err)
	}
	return updated, nil
}

func (r *PostgresRepository) ListForAdmin(ctx context.Context) ([]Product, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+productSelect+`,COALESCE(enforcement.reason,''),enforcement.created_at,enforcement.actor_name
		FROM products p
		JOIN stores s ON s.id=p.store_id
		JOIN categories c ON c.id=p.category_id
		LEFT JOIN LATERAL (
			SELECT ae.metadata->>'reason' AS reason,ae.created_at,u.display_name AS actor_name
			FROM audit_log ae
			JOIN users u ON u.id=ae.actor_id
			WHERE ae.resource_type='product' AND ae.resource_id=p.id AND ae.action='product.status.suspended'
			ORDER BY ae.created_at DESC,ae.id DESC LIMIT 1
		) enforcement ON true
		WHERE p.status IN ('published','suspended')
		ORDER BY COALESCE(enforcement.created_at,p.updated_at) DESC,p.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list admin products: %w", err)
	}
	items := make([]Product, 0)
	for rows.Next() {
		item, scanErr := scanProductWithEnforcement(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range items {
		items[index], err = r.loadDetails(ctx, items[index])
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func scanProductWithEnforcement(row rowScanner) (Product, error) {
	var value Product
	err := row.Scan(&value.ID, &value.StoreID, &value.StoreName, &value.StoreSlug, &value.Category.ID, &value.Category.Name,
		&value.Category.Slug, &value.Name, &value.Slug, &value.Description, &value.Status, &value.CreatedAt, &value.UpdatedAt,
		&value.EnforcementReason, &value.EnforcedAt, &value.EnforcedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, domain.ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("scan admin product: %w", err)
	}
	value.Variants = []Variant{}
	value.Images = []ProductImage{}
	return value, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, productID, status, reason, actorID string, now time.Time) (Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin update product status: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current string
	if err := tx.QueryRow(ctx, `
		SELECT p.status::text
		FROM products p JOIN stores s ON s.id=p.store_id
		WHERE p.id=$1
		FOR UPDATE OF p,s`, productID).Scan(&current); errors.Is(err, pgx.ErrNoRows) {
		return Product{}, domain.ErrNotFound
	} else if err != nil {
		return Product{}, fmt.Errorf("lock product status: %w", err)
	}
	if status == "suspended" && current != "published" || status == "published" && current != "suspended" {
		return Product{}, domain.ErrConflict
	}
	updated, err := scanProduct(tx.QueryRow(ctx, `
		WITH changed AS (
			UPDATE products p SET status=$2::product_status,updated_at=$3
			WHERE p.id=$1
			AND ($2::text='suspended' OR (
				$2::text='published'
				AND EXISTS (SELECT 1 FROM stores own WHERE own.id=p.store_id AND own.status='approved')
				AND EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id=p.id AND v.active AND v.stock>0)
				AND EXISTS (SELECT 1 FROM product_images i WHERE i.product_id=p.id)
			))
			RETURNING p.*
		)
		SELECT `+productSelect+` FROM changed p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id`, productID, status, now))
	if errors.Is(err, domain.ErrNotFound) {
		return Product{}, domain.ErrConflict
	}
	if err != nil {
		return Product{}, err
	}
	if err := insertAuditMetadata(ctx, tx, actorID, "product.status."+status, "product", productID,
		map[string]any{"from": current, "to": status, "reason": reason}, now); err != nil {
		return Product{}, err
	}
	if err := insertSearchOutbox(ctx, tx, updated); err != nil {
		return Product{}, err
	}
	if err := tx.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT ae.metadata->>'reason' FROM audit_log ae WHERE ae.resource_type='product' AND ae.resource_id=$1 AND ae.action='product.status.suspended' ORDER BY ae.created_at DESC,ae.id DESC LIMIT 1),''),
			(SELECT ae.created_at FROM audit_log ae WHERE ae.resource_type='product' AND ae.resource_id=$1 AND ae.action='product.status.suspended' ORDER BY ae.created_at DESC,ae.id DESC LIMIT 1),
			(SELECT u.display_name FROM audit_log ae JOIN users u ON u.id=ae.actor_id WHERE ae.resource_type='product' AND ae.resource_id=$1 AND ae.action='product.status.suspended' ORDER BY ae.created_at DESC,ae.id DESC LIMIT 1)`,
		productID).Scan(&updated.EnforcementReason, &updated.EnforcedAt, &updated.EnforcedBy); err != nil {
		return Product{}, fmt.Errorf("load enforcement context: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit product status: %w", err)
	}
	return r.loadDetails(ctx, updated)
}

func (r *PostgresRepository) ListPublic(ctx context.Context, filters Filters) (Page, error) {
	return r.listPublic(ctx, filters, nil, false)
}

func (r *PostgresRepository) ListPublicCandidates(ctx context.Context, filters Filters, candidates []string) (Page, error) {
	return r.listPublic(ctx, filters, candidates, true)
}

func (r *PostgresRepository) listPublic(ctx context.Context, filters Filters, candidates []string, candidateSearch bool) (Page, error) {
	search, category, storeSlug := nullableText(filters.Search), nullableText(filters.CategorySlug), nullableText(filters.StoreSlug)
	var total int
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM products p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id
		JOIN LATERAL (SELECT min(price_minor) min_price,bool_or(stock>0) in_stock FROM product_variants WHERE product_id=p.id AND active) prices ON prices.min_price IS NOT NULL
		WHERE p.status='published' AND s.status='approved'
		AND ((NOT $6::boolean AND ($1::text IS NULL OR p.name ILIKE '%'||$1||'%' OR p.description ILIKE '%'||$1||'%'))
			OR ($6::boolean AND p.id=ANY($7::uuid[])))
		AND ($2::text IS NULL OR c.slug=$2) AND ($3::bigint IS NULL OR prices.min_price >= $3)
		AND ($4::bigint IS NULL OR prices.min_price <= $4) AND (NOT $5 OR prices.in_stock)
		AND ($8::text IS NULL OR s.slug=$8)`,
		search, category, filters.MinPrice, filters.MaxPrice, filters.InStock, candidateSearch, candidates, storeSlug).Scan(&total)
	if err != nil {
		return Page{}, fmt.Errorf("count public products: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT p.id::text,p.name,p.slug,s.name,s.slug,c.id::text,c.name,c.slug,prices.min_price,prices.currency,prices.in_stock,
		COALESCE((SELECT url FROM product_images WHERE product_id=p.id ORDER BY position LIMIT 1),'')
		FROM products p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id
		JOIN LATERAL (SELECT min(price_minor) min_price,min(currency) currency,bool_or(stock>0) in_stock FROM product_variants WHERE product_id=p.id AND active) prices ON prices.min_price IS NOT NULL
		WHERE p.status='published' AND s.status='approved'
		AND ((NOT $6::boolean AND ($1::text IS NULL OR p.name ILIKE '%'||$1||'%' OR p.description ILIKE '%'||$1||'%'))
			OR ($6::boolean AND p.id=ANY($7::uuid[])))
		AND ($2::text IS NULL OR c.slug=$2) AND ($3::bigint IS NULL OR prices.min_price >= $3)
		AND ($4::bigint IS NULL OR prices.min_price <= $4) AND (NOT $5 OR prices.in_stock)
		AND ($8::text IS NULL OR s.slug=$8)
		ORDER BY p.updated_at DESC,p.id DESC LIMIT $9 OFFSET $10`,
		search, category, filters.MinPrice, filters.MaxPrice, filters.InStock, candidateSearch, candidates, storeSlug, filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return Page{}, fmt.Errorf("list public products: %w", err)
	}
	defer rows.Close()
	items := make([]Summary, 0)
	for rows.Next() {
		var item Summary
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.StoreName, &item.StoreSlug, &item.Category.ID, &item.Category.Name, &item.Category.Slug,
			&item.MinPriceMinor, &item.Currency, &item.InStock, &item.ImageURL); err != nil {
			return Page{}, fmt.Errorf("scan product summary: %w", err)
		}
		items = append(items, item)
	}
	return Page{Items: items, Page: filters.Page, PageSize: filters.PageSize, Total: total}, rows.Err()
}

func insertSearchOutbox(ctx context.Context, tx pgx.Tx, product Product) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate search event ID: %w", err)
	}
	payload, err := json.Marshal(map[string]any{
		"productId": product.ID, "name": product.Name, "description": product.Description,
		"categorySlug": product.Category.Slug, "updatedAt": product.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("encode search event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox_events (id,aggregate_type,aggregate_id,event_type,payload,occurred_at)
		VALUES ($1,'catalog',$2,'catalog.search.upsert.v1',$3,$4)`, id.String(), product.ID, payload, product.UpdatedAt); err != nil {
		return fmt.Errorf("insert search event: %w", err)
	}
	return nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (r *PostgresRepository) FindPublic(ctx context.Context, slug string) (Product, error) {
	value, err := scanProduct(r.pool.QueryRow(ctx, `SELECT `+productSelect+` FROM products p JOIN stores s ON s.id=p.store_id JOIN categories c ON c.id=p.category_id WHERE p.slug=$1 AND p.status='published' AND s.status='approved'`, slug))
	if err != nil {
		return Product{}, err
	}
	return r.loadDetails(ctx, value)
}

func (r *PostgresRepository) listProducts(ctx context.Context, query string, args ...any) ([]Product, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	items := make([]Product, 0)
	for rows.Next() {
		item, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index], err = r.loadDetails(ctx, items[index])
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (r *PostgresRepository) loadDetails(ctx context.Context, value Product) (Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,sku,name,attributes,price_minor,currency,stock,active FROM product_variants WHERE product_id=$1 ORDER BY created_at,id`, value.ID)
	if err != nil {
		return Product{}, fmt.Errorf("list variants: %w", err)
	}
	for rows.Next() {
		variant, err := scanVariant(rows)
		if err != nil {
			rows.Close()
			return Product{}, err
		}
		value.Variants = append(value.Variants, variant)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Product{}, err
	}
	rows.Close()
	imageRows, err := r.pool.Query(ctx, `SELECT id::text,url,alt_text,position FROM product_images WHERE product_id=$1 ORDER BY position,id`, value.ID)
	if err != nil {
		return Product{}, fmt.Errorf("list images: %w", err)
	}
	defer imageRows.Close()
	for imageRows.Next() {
		var image ProductImage
		if err := imageRows.Scan(&image.ID, &image.URL, &image.AltText, &image.Position); err != nil {
			return Product{}, err
		}
		value.Images = append(value.Images, image)
	}
	return value, imageRows.Err()
}

type auditExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertAudit(ctx context.Context, tx auditExecutor, actorID, action, resourceType, resourceID string, now time.Time) error {
	return insertAuditMetadata(ctx, tx, actorID, action, resourceType, resourceID, map[string]any{}, now)
}

func insertAuditMetadata(ctx context.Context, tx auditExecutor, actorID, action, resourceType, resourceID string, metadata map[string]any, now time.Time) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate audit ID: %w", err)
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_log (id,actor_id,action,resource_type,resource_id,metadata,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, id.String(), actorID, action, resourceType, resourceID, payload, now); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func mapWriteError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return domain.ErrConflict
		case "23503", "23514", "22P02":
			return domain.ErrInvalid
		}
	}
	return err
}
