package purchase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func newID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	return id.String(), nil
}

func purchaseReference(id string) string {
	return "CL-" + strings.ToUpper(strings.ReplaceAll(id, "-", "")[:12])
}

func (r *PostgresRepository) Cart(ctx context.Context, buyerID string) (Cart, error) {
	if _, err := r.ensureCart(ctx, r.pool, buyerID, time.Now().UTC()); err != nil {
		return Cart{}, err
	}
	return loadCart(ctx, r.pool, buyerID)
}

func (r *PostgresRepository) ensureCart(ctx context.Context, db queryer, buyerID string, now time.Time) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}
	var cartID string
	err = db.QueryRow(ctx, `
		INSERT INTO carts (id,buyer_id,currency,created_at,updated_at)
		VALUES ($1,$2,'IDR',$3,$3)
		ON CONFLICT (buyer_id) DO UPDATE SET buyer_id=EXCLUDED.buyer_id
		RETURNING id::text`, id, buyerID, now).Scan(&cartID)
	if err != nil {
		return "", fmt.Errorf("ensure cart: %w", err)
	}
	return cartID, nil
}

func (r *PostgresRepository) SetCartItem(ctx context.Context, buyerID, variantID string, quantity int, now time.Time) (Cart, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Cart{}, fmt.Errorf("begin cart update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var stock int
	var active bool
	var productStatus, storeStatus string
	err = tx.QueryRow(ctx, `
		SELECT pv.stock,pv.active,p.status::text,s.status::text
		FROM product_variants pv
		JOIN products p ON p.id=pv.product_id
		JOIN stores s ON s.id=p.store_id
		WHERE pv.id=$1
		FOR SHARE OF p,s`, variantID).Scan(&stock, &active, &productStatus, &storeStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, domain.ErrNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("inspect cart variant: %w", err)
	}
	if !active || productStatus != "published" || storeStatus != "approved" || quantity > stock {
		return Cart{}, domain.ErrConflict
	}
	cartID, err := r.ensureCart(ctx, tx, buyerID, now)
	if err != nil {
		return Cart{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO cart_items (cart_id,variant_id,quantity,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (cart_id,variant_id) DO UPDATE SET quantity=EXCLUDED.quantity,updated_at=EXCLUDED.updated_at`,
		cartID, variantID, quantity, now)
	if err != nil {
		return Cart{}, fmt.Errorf("upsert cart item: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at=$2 WHERE id=$1`, cartID, now); err != nil {
		return Cart{}, fmt.Errorf("touch cart: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit cart update: %w", err)
	}
	return loadCart(ctx, r.pool, buyerID)
}

func (r *PostgresRepository) RemoveCartItem(ctx context.Context, buyerID, variantID string, now time.Time) (Cart, error) {
	tag, err := r.pool.Exec(ctx, `
		WITH target AS (SELECT id FROM carts WHERE buyer_id=$1)
		DELETE FROM cart_items ci USING target WHERE ci.cart_id=target.id AND ci.variant_id=$2`, buyerID, variantID)
	if err != nil {
		return Cart{}, fmt.Errorf("remove cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Cart{}, domain.ErrNotFound
	}
	if _, err := r.pool.Exec(ctx, `UPDATE carts SET updated_at=$2 WHERE buyer_id=$1`, buyerID, now); err != nil {
		return Cart{}, fmt.Errorf("touch cart: %w", err)
	}
	return loadCart(ctx, r.pool, buyerID)
}

func loadCart(ctx context.Context, db queryer, buyerID string) (Cart, error) {
	var result Cart
	err := db.QueryRow(ctx, `SELECT id::text,currency,updated_at FROM carts WHERE buyer_id=$1`, buyerID).
		Scan(&result.ID, &result.Currency, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, domain.ErrNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("scan cart: %w", err)
	}
	rows, err := db.Query(ctx, `
		SELECT s.id::text,s.name,pv.id::text,p.id::text,p.name,p.slug,pv.name,pv.sku,
			COALESCE((SELECT pi.url FROM product_images pi WHERE pi.product_id=p.id ORDER BY pi.position LIMIT 1),''),
			ci.quantity,pv.stock,pv.price_minor,(pv.price_minor*ci.quantity),pv.currency
		FROM cart_items ci
		JOIN product_variants pv ON pv.id=ci.variant_id
		JOIN products p ON p.id=pv.product_id
		JOIN stores s ON s.id=p.store_id
		WHERE ci.cart_id=$1
		ORDER BY s.name,p.name,pv.name`, result.ID)
	if err != nil {
		return Cart{}, fmt.Errorf("list cart items: %w", err)
	}
	defer rows.Close()
	result.Stores = []CartStore{}
	storeIndexes := map[string]int{}
	for rows.Next() {
		var storeID, storeName string
		var item CartItem
		if err := rows.Scan(&storeID, &storeName, &item.VariantID, &item.ProductID, &item.ProductName, &item.ProductSlug,
			&item.VariantName, &item.SKU, &item.ImageURL, &item.Quantity, &item.AvailableStock,
			&item.UnitPriceMinor, &item.LineTotalMinor, &item.Currency); err != nil {
			return Cart{}, fmt.Errorf("scan cart item: %w", err)
		}
		index, exists := storeIndexes[storeID]
		if !exists {
			index = len(result.Stores)
			storeIndexes[storeID] = index
			result.Stores = append(result.Stores, CartStore{StoreID: storeID, StoreName: storeName, Items: []CartItem{}})
		}
		result.Stores[index].Items = append(result.Stores[index].Items, item)
		result.Stores[index].SubtotalMinor += item.LineTotalMinor
		result.SubtotalMinor += item.LineTotalMinor
		result.TotalQuantity += item.Quantity
	}
	if err := rows.Err(); err != nil {
		return Cart{}, fmt.Errorf("iterate cart items: %w", err)
	}
	return result, nil
}

type checkoutItem struct {
	CartItem
	StoreID   string
	StoreName string
}

func (r *PostgresRepository) Checkout(ctx context.Context, buyerID, idempotencyKey string, now, expiresAt time.Time) (Purchase, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Purchase{}, false, fmt.Errorf("begin checkout: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if existing, findErr := findPurchaseByKey(ctx, tx, buyerID, idempotencyKey); findErr == nil {
		return existing, false, nil
	} else if !errors.Is(findErr, domain.ErrNotFound) {
		return Purchase{}, false, findErr
	}

	var cartID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM carts WHERE buyer_id=$1 FOR UPDATE`, buyerID).Scan(&cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Purchase{}, false, domain.ErrConflict
	}
	if err != nil {
		return Purchase{}, false, fmt.Errorf("lock cart: %w", err)
	}
	rows, err := tx.Query(ctx, `
		SELECT pv.id::text,p.id::text,p.name,p.slug,pv.name,pv.sku,
			COALESCE((SELECT pi.url FROM product_images pi WHERE pi.product_id=p.id ORDER BY pi.position LIMIT 1),''),
			ci.quantity,pv.stock,pv.price_minor,(pv.price_minor*ci.quantity),pv.currency,
			s.id::text,s.name,pv.active,p.status::text,s.status::text
		FROM cart_items ci
		JOIN product_variants pv ON pv.id=ci.variant_id
		JOIN products p ON p.id=pv.product_id
		JOIN stores s ON s.id=p.store_id
		WHERE ci.cart_id=$1
		ORDER BY pv.id
		FOR UPDATE OF pv
		FOR SHARE OF p,s`, cartID)
	if err != nil {
		return Purchase{}, false, fmt.Errorf("lock checkout items: %w", err)
	}
	items := []checkoutItem{}
	for rows.Next() {
		var item checkoutItem
		var active bool
		var productStatus, storeStatus string
		if err := rows.Scan(&item.VariantID, &item.ProductID, &item.ProductName, &item.ProductSlug, &item.VariantName,
			&item.SKU, &item.ImageURL, &item.Quantity, &item.AvailableStock, &item.UnitPriceMinor,
			&item.LineTotalMinor, &item.Currency, &item.StoreID, &item.StoreName, &active,
			&productStatus, &storeStatus); err != nil {
			rows.Close()
			return Purchase{}, false, fmt.Errorf("scan checkout item: %w", err)
		}
		if !active || productStatus != "published" || storeStatus != "approved" ||
			item.Currency != "IDR" || item.Quantity > item.AvailableStock {
			rows.Close()
			return Purchase{}, false, domain.ErrConflict
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Purchase{}, false, fmt.Errorf("iterate checkout items: %w", err)
	}
	rows.Close()
	if len(items) == 0 {
		return Purchase{}, false, domain.ErrConflict
	}

	purchaseID, err := newID()
	if err != nil {
		return Purchase{}, false, err
	}
	reference := purchaseReference(purchaseID)
	var total int64
	for _, item := range items {
		total += item.LineTotalMinor
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO purchases (id,reference,buyer_id,idempotency_key,status,payment_status,currency,subtotal_minor,total_minor,reservation_expires_at,created_at,updated_at)
		VALUES ($1,$2,$3,$4,'pending_payment','pending','IDR',$5,$5,$6,$7,$7)`,
		purchaseID, reference, buyerID, idempotencyKey, total, expiresAt, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "purchases_buyer_idempotency_unique" {
			_ = tx.Rollback(ctx)
			existing, findErr := findPurchaseByKey(ctx, r.pool, buyerID, idempotencyKey)
			return existing, false, findErr
		}
		return Purchase{}, false, fmt.Errorf("create purchase: %w", err)
	}

	orderIDs := map[string]string{}
	orderSubtotals := map[string]int64{}
	for _, item := range items {
		orderSubtotals[item.StoreID] += item.LineTotalMinor
	}
	for storeID, subtotal := range orderSubtotals {
		orderID, idErr := newID()
		if idErr != nil {
			return Purchase{}, false, idErr
		}
		orderIDs[storeID] = orderID
		if _, err := tx.Exec(ctx, `
			INSERT INTO seller_orders (id,purchase_id,store_id,status,currency,subtotal_minor,created_at,updated_at)
			VALUES ($1,$2,$3,'pending_payment','IDR',$4,$5,$5)`, orderID, purchaseID, storeID, subtotal, now); err != nil {
			return Purchase{}, false, fmt.Errorf("create seller order: %w", err)
		}
	}

	for _, item := range items {
		var resultingStock int
		err := tx.QueryRow(ctx, `
			UPDATE product_variants SET stock=stock-$2,updated_at=$3
			WHERE id=$1 AND stock >= $2 RETURNING stock`, item.VariantID, item.Quantity, now).Scan(&resultingStock)
		if errors.Is(err, pgx.ErrNoRows) {
			return Purchase{}, false, domain.ErrConflict
		}
		if err != nil {
			return Purchase{}, false, fmt.Errorf("reserve inventory: %w", err)
		}
		itemID, idErr := newID()
		if idErr != nil {
			return Purchase{}, false, idErr
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO purchase_items (id,seller_order_id,product_id,variant_id,product_name,variant_name,sku,image_url,quantity,unit_price_minor,line_total_minor,currency,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'IDR',$12)`, itemID, orderIDs[item.StoreID], item.ProductID,
			item.VariantID, item.ProductName, item.VariantName, item.SKU, item.ImageURL, item.Quantity,
			item.UnitPriceMinor, item.LineTotalMinor, now); err != nil {
			return Purchase{}, false, fmt.Errorf("snapshot purchase item: %w", err)
		}
		reservationID, idErr := newID()
		if idErr != nil {
			return Purchase{}, false, idErr
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_reservations (id,purchase_id,variant_id,quantity,status,expires_at,created_at)
			VALUES ($1,$2,$3,$4,'active',$5,$6)`, reservationID, purchaseID, item.VariantID, item.Quantity, expiresAt, now); err != nil {
			return Purchase{}, false, fmt.Errorf("record reservation: %w", err)
		}
		ledgerID, idErr := newID()
		if idErr != nil {
			return Purchase{}, false, idErr
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_ledger (id,variant_id,actor_id,delta,resulting_stock,reason,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, ledgerID, item.VariantID, buyerID, -item.Quantity, resultingStock,
			"reserved for "+reference, now); err != nil {
			return Purchase{}, false, fmt.Errorf("record reservation ledger: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1`, cartID); err != nil {
		return Purchase{}, false, fmt.Errorf("clear checked out cart: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at=$2 WHERE id=$1`, cartID, now); err != nil {
		return Purchase{}, false, fmt.Errorf("touch checked out cart: %w", err)
	}
	if err := insertOutbox(ctx, tx, purchaseID, "purchase.created", map[string]any{
		"purchaseId": purchaseID, "reference": reference, "status": "pending_payment", "recipientIds": []string{buyerID},
	}, now); err != nil {
		return Purchase{}, false, err
	}
	result, err := loadPurchase(ctx, tx, purchaseID)
	if err != nil {
		return Purchase{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Purchase{}, false, fmt.Errorf("commit checkout: %w", err)
	}
	return result, true, nil
}

func (r *PostgresRepository) AttachPaymentIntent(ctx context.Context, purchaseID, intentID string, now time.Time) (Purchase, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE purchases SET payment_intent_id=$2,updated_at=$3
		WHERE id=$1 AND status='pending_payment' AND (payment_intent_id IS NULL OR payment_intent_id=$2)`, purchaseID, intentID, now)
	if err != nil {
		return Purchase{}, fmt.Errorf("attach payment intent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Purchase{}, domain.ErrConflict
	}
	return loadPurchase(ctx, r.pool, purchaseID)
}

func (r *PostgresRepository) FailPaymentSetup(ctx context.Context, purchaseID string, now time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin payment setup failure: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var buyerID, reference, status string
	err = tx.QueryRow(ctx, `SELECT buyer_id::text,reference,status::text FROM purchases WHERE id=$1 FOR UPDATE`, purchaseID).
		Scan(&buyerID, &reference, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock failed purchase: %w", err)
	}
	if status != "pending_payment" {
		return nil
	}
	if err := releaseReservationRows(ctx, tx, purchaseID, buyerID, "released", "payment setup failed", now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE purchases SET status='payment_failed',payment_status='failed',updated_at=$2 WHERE id=$1`, purchaseID, now); err != nil {
		return fmt.Errorf("fail purchase payment setup: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE seller_orders SET status='cancelled',updated_at=$2 WHERE purchase_id=$1`, purchaseID, now); err != nil {
		return fmt.Errorf("cancel seller orders: %w", err)
	}
	if err := insertOutbox(ctx, tx, purchaseID, "purchase.payment_failed", map[string]any{
		"purchaseId": purchaseID, "reference": reference, "status": "payment_failed", "recipientIds": []string{buyerID},
	}, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) ListPurchases(ctx context.Context, buyerID string) ([]Purchase, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text FROM purchases WHERE buyer_id=$1 ORDER BY created_at DESC`, buyerID)
	if err != nil {
		return nil, fmt.Errorf("list purchase IDs: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan purchase ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate purchase IDs: %w", err)
	}
	result := make([]Purchase, 0, len(ids))
	for _, id := range ids {
		item, err := loadPurchase(ctx, r.pool, id)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *PostgresRepository) FindPurchase(ctx context.Context, buyerID, purchaseID string) (Purchase, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM purchases WHERE id=$1 AND buyer_id=$2)`, purchaseID, buyerID).Scan(&exists); err != nil {
		return Purchase{}, fmt.Errorf("authorize purchase: %w", err)
	}
	if !exists {
		return Purchase{}, domain.ErrNotFound
	}
	return loadPurchase(ctx, r.pool, purchaseID)
}

func findPurchaseByKey(ctx context.Context, db queryer, buyerID, key string) (Purchase, error) {
	var id string
	err := db.QueryRow(ctx, `SELECT id::text FROM purchases WHERE buyer_id=$1 AND idempotency_key=$2`, buyerID, key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Purchase{}, domain.ErrNotFound
	}
	if err != nil {
		return Purchase{}, fmt.Errorf("find idempotent purchase: %w", err)
	}
	return loadPurchase(ctx, db, id)
}

func loadPurchase(ctx context.Context, db queryer, purchaseID string) (Purchase, error) {
	var result Purchase
	var intentID *string
	err := db.QueryRow(ctx, `
		SELECT id::text,reference,buyer_id::text,status::text,payment_status::text,payment_intent_id,currency,
			subtotal_minor,total_minor,reservation_expires_at,created_at,updated_at
		FROM purchases WHERE id=$1`, purchaseID).Scan(&result.ID, &result.Reference, &result.BuyerID, &result.Status,
		&result.PaymentStatus, &intentID, &result.Currency, &result.SubtotalMinor, &result.TotalMinor,
		&result.ReservationExpiresAt, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Purchase{}, domain.ErrNotFound
	}
	if err != nil {
		return Purchase{}, fmt.Errorf("scan purchase: %w", err)
	}
	if intentID != nil {
		result.PaymentIntentID = *intentID
	}
	orders, err := loadSellerOrders(ctx, db, purchaseID, "")
	if err != nil {
		return Purchase{}, err
	}
	result.SellerOrders = orders
	return result, nil
}

func loadSellerOrders(ctx context.Context, db queryer, purchaseID, sellerID string) ([]SellerOrder, error) {
	query := `
		SELECT so.id::text,so.purchase_id::text,p.reference,so.store_id::text,s.name,so.status::text,so.currency,
			so.subtotal_minor,so.processing_at,so.shipped_at,so.delivered_at,so.cancelled_at,so.cancellation_reason,
			so.created_at,so.updated_at
		FROM seller_orders so JOIN purchases p ON p.id=so.purchase_id JOIN stores s ON s.id=so.store_id`
	args := []any{}
	if purchaseID != "" {
		query += ` WHERE so.purchase_id=$1 ORDER BY s.name`
		args = append(args, purchaseID)
	} else {
		query += ` WHERE s.seller_id=$1 ORDER BY so.created_at DESC`
		args = append(args, sellerID)
	}
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list seller orders: %w", err)
	}
	defer rows.Close()
	orders := []SellerOrder{}
	for rows.Next() {
		var order SellerOrder
		if err := rows.Scan(&order.ID, &order.PurchaseID, &order.Reference, &order.StoreID, &order.StoreName,
			&order.Status, &order.Currency, &order.SubtotalMinor, &order.ProcessingAt, &order.ShippedAt,
			&order.DeliveredAt, &order.CancelledAt, &order.CancellationReason, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan seller order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seller orders: %w", err)
	}
	for index := range orders {
		items, err := loadPurchaseItems(ctx, db, orders[index].ID)
		if err != nil {
			return nil, err
		}
		orders[index].Items = items
	}
	return orders, nil
}

func loadPurchaseItems(ctx context.Context, db queryer, orderID string) ([]PurchaseItem, error) {
	rows, err := db.Query(ctx, `
		SELECT id::text,product_id::text,variant_id::text,product_name,variant_name,sku,image_url,quantity,
			unit_price_minor,line_total_minor,currency,
			EXISTS (SELECT 1 FROM reviews WHERE reviews.purchase_item_id=purchase_items.id)
		FROM purchase_items WHERE seller_order_id=$1 ORDER BY product_name,variant_name`, orderID)
	if err != nil {
		return nil, fmt.Errorf("list purchase items: %w", err)
	}
	defer rows.Close()
	items := []PurchaseItem{}
	for rows.Next() {
		var item PurchaseItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.VariantID, &item.ProductName, &item.VariantName,
			&item.SKU, &item.ImageURL, &item.Quantity, &item.UnitPriceMinor, &item.LineTotalMinor, &item.Currency,
			&item.Reviewed); err != nil {
			return nil, fmt.Errorf("scan purchase item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListSellerOrders(ctx context.Context, sellerID string) ([]SellerOrder, error) {
	return loadSellerOrders(ctx, r.pool, "", sellerID)
}

func (r *PostgresRepository) HandlePaymentEvent(ctx context.Context, event PaymentEvent, now time.Time) (Purchase, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Purchase{}, fmt.Errorf("begin payment event: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var purchaseID, buyerID, reference, status, currency string
	var total int64
	err = tx.QueryRow(ctx, `
		SELECT id::text,buyer_id::text,reference,status::text,currency,total_minor
		FROM purchases WHERE payment_intent_id=$1 AND reference=$2 FOR UPDATE`, event.Data.IntentID, event.Data.Reference).
		Scan(&purchaseID, &buyerID, &reference, &status, &currency, &total)
	if errors.Is(err, pgx.ErrNoRows) {
		return Purchase{}, domain.ErrNotFound
	}
	if err != nil {
		return Purchase{}, fmt.Errorf("lock webhook purchase: %w", err)
	}
	if total != event.Data.AmountMinor || currency != event.Data.Currency {
		return Purchase{}, domain.ErrInvalid
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO payment_events (event_id,payment_intent_id,event_type,payload,received_at)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT (event_id) DO NOTHING`, event.ID, event.Data.IntentID, event.Type, event.Payload, now)
	if err != nil {
		return Purchase{}, fmt.Errorf("record payment event: %w", err)
	}
	if tag.RowsAffected() == 0 || status != "pending_payment" {
		result, loadErr := loadPurchase(ctx, tx, purchaseID)
		if loadErr != nil {
			return Purchase{}, loadErr
		}
		if err := tx.Commit(ctx); err != nil {
			return Purchase{}, fmt.Errorf("commit duplicate payment event: %w", err)
		}
		return result, nil
	}
	recipients, err := purchaseRecipients(ctx, tx, purchaseID)
	if err != nil {
		return Purchase{}, err
	}
	if event.Type == "payment.succeeded" {
		if _, err := tx.Exec(ctx, `UPDATE purchases SET status='paid',payment_status='succeeded',updated_at=$2 WHERE id=$1`, purchaseID, now); err != nil {
			return Purchase{}, fmt.Errorf("mark purchase paid: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE seller_orders SET status='paid',updated_at=$2 WHERE purchase_id=$1`, purchaseID, now); err != nil {
			return Purchase{}, fmt.Errorf("mark seller orders paid: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status='converted',resolved_at=$2 WHERE purchase_id=$1 AND status='active'`, purchaseID, now); err != nil {
			return Purchase{}, fmt.Errorf("convert reservations: %w", err)
		}
		if err := insertOutbox(ctx, tx, purchaseID, "purchase.paid", map[string]any{
			"purchaseId": purchaseID, "reference": reference, "status": "paid", "recipientIds": recipients,
		}, now); err != nil {
			return Purchase{}, err
		}
	} else {
		if err := releaseReservationRows(ctx, tx, purchaseID, buyerID, "released", "payment failed", now); err != nil {
			return Purchase{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE purchases SET status='payment_failed',payment_status='failed',updated_at=$2 WHERE id=$1`, purchaseID, now); err != nil {
			return Purchase{}, fmt.Errorf("mark purchase payment failed: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE seller_orders SET status='cancelled',updated_at=$2 WHERE purchase_id=$1`, purchaseID, now); err != nil {
			return Purchase{}, fmt.Errorf("cancel failed seller orders: %w", err)
		}
		if err := insertOutbox(ctx, tx, purchaseID, "purchase.payment_failed", map[string]any{
			"purchaseId": purchaseID, "reference": reference, "status": "payment_failed", "recipientIds": recipients,
		}, now); err != nil {
			return Purchase{}, err
		}
	}
	result, err := loadPurchase(ctx, tx, purchaseID)
	if err != nil {
		return Purchase{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Purchase{}, fmt.Errorf("commit payment event: %w", err)
	}
	return result, nil
}

func releaseReservationRows(ctx context.Context, tx pgx.Tx, purchaseID, actorID, resolution, reason string, now time.Time) error {
	rows, err := tx.Query(ctx, `
		SELECT variant_id::text,quantity FROM inventory_reservations
		WHERE purchase_id=$1 AND status='active' ORDER BY variant_id FOR UPDATE`, purchaseID)
	if err != nil {
		return fmt.Errorf("lock reservations: %w", err)
	}
	type reserved struct {
		variantID string
		quantity  int
	}
	reservations := []reserved{}
	for rows.Next() {
		var item reserved
		if err := rows.Scan(&item.variantID, &item.quantity); err != nil {
			rows.Close()
			return fmt.Errorf("scan reservation: %w", err)
		}
		reservations = append(reservations, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate reservations: %w", err)
	}
	rows.Close()
	for _, item := range reservations {
		var resultingStock int
		if err := tx.QueryRow(ctx, `UPDATE product_variants SET stock=stock+$2,updated_at=$3 WHERE id=$1 RETURNING stock`,
			item.variantID, item.quantity, now).Scan(&resultingStock); err != nil {
			return fmt.Errorf("release inventory: %w", err)
		}
		ledgerID, err := newID()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_ledger (id,variant_id,actor_id,delta,resulting_stock,reason,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, ledgerID, item.variantID, actorID, item.quantity, resultingStock, reason, now); err != nil {
			return fmt.Errorf("record inventory release: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status=$2,resolved_at=$3 WHERE purchase_id=$1 AND status='active'`,
		purchaseID, resolution, now); err != nil {
		return fmt.Errorf("resolve reservations: %w", err)
	}
	return nil
}

func purchaseRecipients(ctx context.Context, tx pgx.Tx, purchaseID string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT buyer_id::text FROM purchases WHERE id=$1
		UNION
		SELECT s.seller_id::text FROM seller_orders so JOIN stores s ON s.id=so.store_id WHERE so.purchase_id=$1`, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("list purchase recipients: %w", err)
	}
	defer rows.Close()
	recipients := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan purchase recipient: %w", err)
		}
		recipients = append(recipients, id)
	}
	return recipients, rows.Err()
}

func insertOutbox(ctx context.Context, tx pgx.Tx, aggregateID, eventType string, payload any, now time.Time) error {
	id, err := newID()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode outbox event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox_events (id,aggregate_type,aggregate_id,event_type,payload,occurred_at)
		VALUES ($1,'purchase',$2,$3,$4,$5)`, id, aggregateID, eventType, encoded, now); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ExpireReservations(ctx context.Context, now time.Time) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin reservation expiry: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `
		SELECT id::text,buyer_id::text,reference FROM purchases
		WHERE status='pending_payment' AND reservation_expires_at <= $1
		ORDER BY reservation_expires_at,id FOR UPDATE SKIP LOCKED LIMIT 50`, now)
	if err != nil {
		return 0, fmt.Errorf("list expired purchases: %w", err)
	}
	type expired struct{ id, buyerID, reference string }
	items := []expired{}
	for rows.Next() {
		var item expired
		if err := rows.Scan(&item.id, &item.buyerID, &item.reference); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan expired purchase: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate expired purchases: %w", err)
	}
	rows.Close()
	for _, item := range items {
		if err := releaseReservationRows(ctx, tx, item.id, item.buyerID, "expired", "reservation expired", now); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `UPDATE purchases SET status='expired',payment_status='expired',updated_at=$2 WHERE id=$1`, item.id, now); err != nil {
			return 0, fmt.Errorf("expire purchase: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE seller_orders SET status='cancelled',updated_at=$2 WHERE purchase_id=$1`, item.id, now); err != nil {
			return 0, fmt.Errorf("cancel expired seller orders: %w", err)
		}
		if err := insertOutbox(ctx, tx, item.id, "purchase.expired", map[string]any{
			"purchaseId": item.id, "reference": item.reference, "status": "expired", "recipientIds": []string{item.buyerID},
		}, now); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit reservation expiry: %w", err)
	}
	return len(items), nil
}

func (r *PostgresRepository) ListNotifications(ctx context.Context, userID string) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text,kind,title,body,href,read_at,created_at FROM notifications
		WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var item Notification
		if err := rows.Scan(&item.ID, &item.Kind, &item.Title, &item.Body, &item.Href, &item.ReadAt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
