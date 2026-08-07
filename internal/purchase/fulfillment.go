package purchase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) UpdateSellerOrder(ctx context.Context, sellerID, orderID, target, reason string, now time.Time) (SellerOrder, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return SellerOrder{}, fmt.Errorf("begin fulfillment update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var purchaseID, buyerID, reference, current, ownerID string
	err = tx.QueryRow(ctx, `
		SELECT so.purchase_id::text,p.buyer_id::text,p.reference,so.status::text,s.seller_id::text
		FROM seller_orders so JOIN purchases p ON p.id=so.purchase_id JOIN stores s ON s.id=so.store_id
		WHERE so.id=$1 FOR UPDATE OF so`, orderID).Scan(&purchaseID, &buyerID, &reference, &current, &ownerID)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && ownerID != sellerID {
		return SellerOrder{}, domain.ErrNotFound
	}
	if err != nil {
		return SellerOrder{}, fmt.Errorf("lock seller order: %w", err)
	}
	allowed := map[string]map[string]bool{
		"paid":       {"processing": true, "cancelled": true},
		"processing": {"shipped": true, "cancelled": true},
		"shipped":    {"delivered": true},
	}
	if !allowed[current][target] {
		return SellerOrder{}, domain.ErrConflict
	}
	if target == "cancelled" {
		if err := releaseOrderInventory(ctx, tx, orderID, purchaseID, sellerID, reason, now); err != nil {
			return SellerOrder{}, err
		}
	}
	_, err = tx.Exec(ctx, `
		UPDATE seller_orders SET status=$2::seller_order_status,updated_at=$3,
			processing_at=CASE WHEN $2::text='processing' THEN $3 ELSE processing_at END,
			shipped_at=CASE WHEN $2::text='shipped' THEN $3 ELSE shipped_at END,
			delivered_at=CASE WHEN $2::text='delivered' THEN $3 ELSE delivered_at END,
			cancelled_at=CASE WHEN $2::text='cancelled' THEN $3 ELSE cancelled_at END,
			cancellation_reason=CASE WHEN $2::text='cancelled' THEN $4 ELSE cancellation_reason END
		WHERE id=$1`, orderID, target, now, reason)
	if err != nil {
		return SellerOrder{}, fmt.Errorf("update seller order: %w", err)
	}
	if target == "cancelled" {
		var remaining int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM seller_orders WHERE purchase_id=$1 AND status <> 'cancelled'`, purchaseID).Scan(&remaining); err != nil {
			return SellerOrder{}, fmt.Errorf("count active seller orders: %w", err)
		}
		if remaining == 0 {
			if _, err := tx.Exec(ctx, `UPDATE purchases SET status='cancelled',updated_at=$2 WHERE id=$1`, purchaseID, now); err != nil {
				return SellerOrder{}, fmt.Errorf("cancel parent purchase: %w", err)
			}
		}
	}
	if err := insertAudit(ctx, tx, sellerID, "seller_order."+target, "seller_order", orderID,
		map[string]any{"from": current, "to": target, "reason": reason}, now); err != nil {
		return SellerOrder{}, err
	}
	if err := insertOutbox(ctx, tx, purchaseID, "order."+target, map[string]any{
		"purchaseId": purchaseID, "orderId": orderID, "reference": reference, "status": target,
		"recipientIds": []string{buyerID, sellerID},
	}, now); err != nil {
		return SellerOrder{}, err
	}
	orders, err := loadSellerOrders(ctx, tx, purchaseID, "")
	if err != nil {
		return SellerOrder{}, err
	}
	var result SellerOrder
	for _, order := range orders {
		if order.ID == orderID {
			result = order
			break
		}
	}
	if result.ID == "" {
		return SellerOrder{}, domain.ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return SellerOrder{}, fmt.Errorf("commit fulfillment update: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) CancelPurchase(ctx context.Context, buyerID, purchaseID, reason string, now time.Time) (Purchase, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Purchase{}, fmt.Errorf("begin purchase cancellation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var reference, status, paymentStatus string
	err = tx.QueryRow(ctx, `SELECT reference,status::text,payment_status::text FROM purchases WHERE id=$1 AND buyer_id=$2 FOR UPDATE`,
		purchaseID, buyerID).Scan(&reference, &status, &paymentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Purchase{}, domain.ErrNotFound
	}
	if err != nil {
		return Purchase{}, fmt.Errorf("lock purchase cancellation: %w", err)
	}
	if status != "pending_payment" && status != "paid" {
		return Purchase{}, domain.ErrConflict
	}
	rows, err := tx.Query(ctx, `SELECT id::text,status::text FROM seller_orders WHERE purchase_id=$1 ORDER BY id FOR UPDATE`, purchaseID)
	if err != nil {
		return Purchase{}, fmt.Errorf("lock cancellation orders: %w", err)
	}
	type orderState struct{ id, status string }
	orders := []orderState{}
	for rows.Next() {
		var order orderState
		if err := rows.Scan(&order.id, &order.status); err != nil {
			rows.Close()
			return Purchase{}, fmt.Errorf("scan cancellation order: %w", err)
		}
		if order.status != "pending_payment" && order.status != "paid" {
			rows.Close()
			return Purchase{}, domain.ErrConflict
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Purchase{}, fmt.Errorf("iterate cancellation orders: %w", err)
	}
	rows.Close()
	for _, order := range orders {
		if err := releaseOrderInventory(ctx, tx, order.id, purchaseID, buyerID, reason, now); err != nil {
			return Purchase{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE seller_orders SET status='cancelled',cancelled_at=$2,cancellation_reason=$3,updated_at=$2 WHERE purchase_id=$1`,
		purchaseID, now, reason); err != nil {
		return Purchase{}, fmt.Errorf("cancel purchase seller orders: %w", err)
	}
	if status == "pending_payment" {
		paymentStatus = "cancelled"
	}
	if _, err := tx.Exec(ctx, `UPDATE purchases SET status='cancelled',payment_status=$2,updated_at=$3 WHERE id=$1`, purchaseID, paymentStatus, now); err != nil {
		return Purchase{}, fmt.Errorf("cancel purchase: %w", err)
	}
	recipients, err := purchaseRecipients(ctx, tx, purchaseID)
	if err != nil {
		return Purchase{}, err
	}
	if err := insertAudit(ctx, tx, buyerID, "purchase.cancelled", "purchase", purchaseID,
		map[string]any{"from": status, "reason": reason}, now); err != nil {
		return Purchase{}, err
	}
	if err := insertOutbox(ctx, tx, purchaseID, "purchase.cancelled", map[string]any{
		"purchaseId": purchaseID, "reference": reference, "status": "cancelled", "recipientIds": recipients,
	}, now); err != nil {
		return Purchase{}, err
	}
	result, err := loadPurchase(ctx, tx, purchaseID)
	if err != nil {
		return Purchase{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Purchase{}, fmt.Errorf("commit purchase cancellation: %w", err)
	}
	return result, nil
}

func releaseOrderInventory(ctx context.Context, tx pgx.Tx, orderID, purchaseID, actorID, reason string, now time.Time) error {
	rows, err := tx.Query(ctx, `
		SELECT pi.variant_id::text,pi.quantity,ir.id::text
		FROM purchase_items pi JOIN inventory_reservations ir ON ir.purchase_id=$2 AND ir.variant_id=pi.variant_id
		WHERE pi.seller_order_id=$1 AND ir.status IN ('active','converted') ORDER BY pi.variant_id FOR UPDATE OF ir`, orderID, purchaseID)
	if err != nil {
		return fmt.Errorf("lock order inventory: %w", err)
	}
	type item struct {
		variantID     string
		quantity      int
		reservationID string
	}
	items := []item{}
	for rows.Next() {
		var current item
		if err := rows.Scan(&current.variantID, &current.quantity, &current.reservationID); err != nil {
			rows.Close()
			return fmt.Errorf("scan order inventory: %w", err)
		}
		items = append(items, current)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate order inventory: %w", err)
	}
	rows.Close()
	if len(items) == 0 {
		return domain.ErrConflict
	}
	for _, current := range items {
		var stock int
		if err := tx.QueryRow(ctx, `UPDATE product_variants SET stock=stock+$2,updated_at=$3 WHERE id=$1 RETURNING stock`,
			current.variantID, current.quantity, now).Scan(&stock); err != nil {
			return fmt.Errorf("restore cancelled inventory: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',resolved_at=$2 WHERE id=$1`, current.reservationID, now); err != nil {
			return fmt.Errorf("release cancelled reservation: %w", err)
		}
		ledgerID, err := newID()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_ledger (id,variant_id,actor_id,delta,resulting_stock,reason,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, ledgerID, current.variantID, actorID, current.quantity, stock,
			"order cancelled: "+reason, now); err != nil {
			return fmt.Errorf("record cancelled inventory: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepository) CreateReview(ctx context.Context, buyerID string, input ReviewInput, now time.Time) (Review, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Review{}, fmt.Errorf("begin review: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var productID string
	err = tx.QueryRow(ctx, `
		SELECT pi.product_id::text FROM purchase_items pi
		JOIN seller_orders so ON so.id=pi.seller_order_id JOIN purchases p ON p.id=so.purchase_id
		WHERE pi.id=$1 AND p.buyer_id=$2 AND so.status='delivered'`, input.PurchaseItemID, buyerID).Scan(&productID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, domain.ErrForbidden
	}
	if err != nil {
		return Review{}, fmt.Errorf("authorize review: %w", err)
	}
	id, err := newID()
	if err != nil {
		return Review{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO reviews (id,buyer_id,purchase_item_id,product_id,rating,title,body,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, id, buyerID, input.PurchaseItemID, productID, input.Rating, input.Title, input.Body, now)
	if err != nil {
		if isUniqueViolation(err) {
			return Review{}, domain.ErrConflict
		}
		return Review{}, fmt.Errorf("create review: %w", err)
	}
	if err := insertAudit(ctx, tx, buyerID, "review.created", "review", id,
		map[string]any{"productId": productID, "purchaseItemId": input.PurchaseItemID, "rating": input.Rating}, now); err != nil {
		return Review{}, err
	}
	result, err := loadReview(ctx, tx, id)
	if err != nil {
		return Review{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Review{}, fmt.Errorf("commit review: %w", err)
	}
	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}

func loadReview(ctx context.Context, db queryer, id string) (Review, error) {
	var result Review
	err := db.QueryRow(ctx, `SELECT r.id::text,r.buyer_id::text,u.display_name,r.purchase_item_id::text,r.product_id::text,
		r.rating,r.title,r.body,r.created_at,r.updated_at FROM reviews r JOIN users u ON u.id=r.buyer_id WHERE r.id=$1`, id).
		Scan(&result.ID, &result.BuyerID, &result.BuyerName, &result.PurchaseItemID, &result.ProductID,
			&result.Rating, &result.Title, &result.Body, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, domain.ErrNotFound
	}
	if err != nil {
		return Review{}, fmt.Errorf("load review: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ReviewsByProductSlug(ctx context.Context, slug string) (ReviewSummary, error) {
	var productID string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM products WHERE slug=$1 AND status='published' AND moderation_status='approved'`, slug).Scan(&productID); errors.Is(err, pgx.ErrNoRows) {
		return ReviewSummary{}, domain.ErrNotFound
	} else if err != nil {
		return ReviewSummary{}, fmt.Errorf("find reviewed product: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT r.id::text,r.buyer_id::text,u.display_name,r.purchase_item_id::text,r.product_id::text,
		r.rating,r.title,r.body,r.created_at,r.updated_at FROM reviews r JOIN users u ON u.id=r.buyer_id
		WHERE r.product_id=$1 ORDER BY r.created_at DESC`, productID)
	if err != nil {
		return ReviewSummary{}, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	result := ReviewSummary{Items: []Review{}}
	var ratingTotal int
	for rows.Next() {
		var item Review
		if err := rows.Scan(&item.ID, &item.BuyerID, &item.BuyerName, &item.PurchaseItemID, &item.ProductID,
			&item.Rating, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return ReviewSummary{}, fmt.Errorf("scan review: %w", err)
		}
		result.Items = append(result.Items, item)
		ratingTotal += item.Rating
	}
	if err := rows.Err(); err != nil {
		return ReviewSummary{}, fmt.Errorf("iterate reviews: %w", err)
	}
	result.Count = len(result.Items)
	if result.Count > 0 {
		result.Average = float64(ratingTotal) / float64(result.Count)
	}
	return result, nil
}

func (r *PostgresRepository) AdminOverview(ctx context.Context) (AdminOverview, error) {
	var result AdminOverview
	err := r.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM users),(SELECT count(*) FROM stores WHERE status='approved'),
		(SELECT count(*) FROM products WHERE status='published' AND moderation_status='approved'),
		(SELECT count(*) FROM purchases),(SELECT count(*) FROM seller_orders WHERE status IN ('paid','processing','shipped')),
		(SELECT count(*) FROM seller_orders WHERE status='delivered'),
		COALESCE((SELECT sum(subtotal_minor) FROM seller_orders WHERE status IN ('paid','processing','shipped','delivered')),0)`).
		Scan(&result.Users, &result.ApprovedStores, &result.PublishedProducts, &result.Purchases,
			&result.ActiveSellerOrders, &result.DeliveredOrders, &result.GrossMerchandiseMinor)
	if err != nil {
		return AdminOverview{}, fmt.Errorf("load admin overview: %w", err)
	}
	result.Currency = "IDR"
	return result, nil
}

func (r *PostgresRepository) AuditEvents(ctx context.Context) ([]AuditEvent, error) {
	rows, err := r.pool.Query(ctx, `SELECT ae.id::text,ae.actor_id::text,u.display_name,u.role::text,
		ae.action,ae.resource_type,ae.resource_id::text,ae.metadata,ae.created_at
		FROM audit_log ae JOIN users u ON u.id=ae.actor_id ORDER BY ae.created_at DESC,ae.id DESC LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()
	result := []AuditEvent{}
	for rows.Next() {
		var event AuditEvent
		var raw []byte
		if err := rows.Scan(&event.ID, &event.ActorID, &event.ActorName, &event.ActorRole, &event.Action,
			&event.ResourceType, &event.ResourceID, &raw, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		if err := json.Unmarshal(raw, &event.Data); err != nil {
			return nil, fmt.Errorf("decode audit event: %w", err)
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func insertAudit(ctx context.Context, tx pgx.Tx, actorID, action, resourceType, resourceID string, data any, now time.Time) error {
	id, err := newID()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode audit event: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_log (id,actor_id,action,resource_type,resource_id,metadata,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, actorID, action, resourceType, resourceID, raw, now); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
