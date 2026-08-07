//go:build integration

package purchase

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConcurrentCheckoutPreventsOversell(t *testing.T) {
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
	repository := NewPostgresRepository(pool)
	variantID := "01989f00-0000-7000-8000-000000000402"
	buyers := []string{"01989f00-0000-7000-8000-000000000001", "01989f00-0000-7000-8000-000000000005"}
	if _, err := pool.Exec(ctx, `UPDATE product_variants SET stock=1 WHERE id=$1`, variantID); err != nil {
		t.Fatal(err)
	}
	for _, buyerID := range buyers {
		if _, err := repository.SetCartItem(ctx, buyerID, variantID, 1, time.Now().UTC()); err != nil {
			t.Fatal(err)
		}
	}

	type result struct {
		buyerID string
		key     string
		value   Purchase
		err     error
	}
	results := make(chan result, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	start := make(chan struct{})
	for index, buyerID := range buyers {
		go func(index int, buyerID string) {
			ready.Done()
			<-start
			key := "concurrent-checkout-" + string(rune('a'+index))
			value, _, err := repository.Checkout(ctx, buyerID, key, time.Now().UTC(), time.Now().UTC().Add(15*time.Minute))
			results <- result{buyerID: buyerID, key: key, value: value, err: err}
		}(index, buyerID)
	}
	ready.Wait()
	close(start)
	first, second := <-results, <-results
	close(results)
	all := []result{first, second}
	var winner result
	var successes, conflicts int
	for _, current := range all {
		switch {
		case current.err == nil:
			successes++
			winner = current
		case errors.Is(current.err, domain.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected checkout error: %v", current.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d results=%#v", successes, conflicts, all)
	}
	var stock, activeReservations, purchaseCount int
	if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id=$1`, variantID).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM inventory_reservations WHERE variant_id=$1 AND status='active'`, variantID).Scan(&activeReservations); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM purchases`).Scan(&purchaseCount); err != nil {
		t.Fatal(err)
	}
	if stock != 0 || activeReservations != 1 || purchaseCount != 1 {
		t.Fatalf("stock=%d activeReservations=%d purchases=%d", stock, activeReservations, purchaseCount)
	}
	replayed, created, err := repository.Checkout(ctx, winner.buyerID, winner.key, time.Now().UTC(), time.Now().UTC().Add(15*time.Minute))
	if err != nil || created || replayed.ID != winner.value.ID {
		t.Fatalf("replay=%#v created=%v err=%v", replayed, created, err)
	}
	paid, err := repository.AttachPaymentIntent(ctx, winner.value.ID, "integration-intent", time.Now().UTC())
	if err != nil || paid.PaymentIntentID != "integration-intent" {
		t.Fatalf("attach payment intent: purchase=%#v err=%v", paid, err)
	}
	event := PaymentEvent{ID: "integration-event", Type: "payment.succeeded", CreatedAt: time.Now().UTC(),
		Data:    PaymentEventData{IntentID: "integration-intent", Reference: winner.value.Reference, AmountMinor: winner.value.TotalMinor, Currency: "IDR"},
		Payload: []byte(`{"integration":true}`)}
	for range 2 {
		paid, err = repository.HandlePaymentEvent(ctx, event, time.Now().UTC())
		if err != nil || paid.Status != "paid" {
			t.Fatalf("handle payment event: purchase=%#v err=%v", paid, err)
		}
	}
	var paymentEvents, paidOutbox int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events WHERE event_id=$1`, event.ID).Scan(&paymentEvents); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id=$1 AND event_type='purchase.paid'`, winner.value.ID).Scan(&paidOutbox); err != nil {
		t.Fatal(err)
	}
	if paymentEvents != 1 || paidOutbox != 1 {
		t.Fatalf("paymentEvents=%d paidOutbox=%d", paymentEvents, paidOutbox)
	}

	expiryBuyer := buyers[0]
	expiryVariant := "01989f00-0000-7000-8000-000000000401"
	if _, err := pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM carts WHERE buyer_id=$1)`, expiryBuyer); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.SetCartItem(ctx, expiryBuyer, expiryVariant, 1, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	expiring, created, err := repository.Checkout(ctx, expiryBuyer, "expiry-checkout-key", time.Now().UTC(), time.Now().UTC().Add(-time.Minute))
	if err != nil || !created {
		t.Fatalf("expiring checkout=%#v created=%v err=%v", expiring, created, err)
	}
	expired, err := repository.ExpireReservations(ctx, time.Now().UTC())
	if err != nil || expired != 1 {
		t.Fatalf("expired=%d err=%v", expired, err)
	}
	expiring, err = repository.FindPurchase(ctx, expiryBuyer, expiring.ID)
	if err != nil || expiring.Status != "expired" || expiring.PaymentStatus != "expired" {
		t.Fatalf("expired purchase=%#v err=%v", expiring, err)
	}
	if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id=$1`, expiryVariant).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if stock != 18 {
		t.Fatalf("released stock=%d", stock)
	}
}

func TestFulfillmentCancellationAndVerifiedReviews(t *testing.T) {
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
	repository := NewPostgresRepository(pool)
	now := time.Now().UTC()
	buyerID := "01989f00-0000-7000-8000-000000000007"
	firstSellerID := "01989f00-0000-7000-8000-000000000004"
	secondSellerID := "01989f00-0000-7000-8000-000000000006"
	firstVariantID := "01989f00-0000-7000-8000-000000000401"
	secondVariantID := "01989f00-0000-7000-8000-000000000402"

	if _, err := pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM carts WHERE buyer_id=$1)`, buyerID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE product_variants SET stock=10 WHERE id=ANY($1::uuid[])`, []string{firstVariantID, secondVariantID}); err != nil {
		t.Fatal(err)
	}
	for _, variantID := range []string{firstVariantID, secondVariantID} {
		if _, err := repository.SetCartItem(ctx, buyerID, variantID, 1, now); err != nil {
			t.Fatal(err)
		}
	}
	purchase, created, err := repository.Checkout(ctx, buyerID, "fulfillment-integration-checkout", now, now.Add(15*time.Minute))
	if err != nil || !created || len(purchase.SellerOrders) != 2 {
		t.Fatalf("checkout=%#v created=%v err=%v", purchase, created, err)
	}
	purchase, err = repository.AttachPaymentIntent(ctx, purchase.ID, "fulfillment-integration-intent", now)
	if err != nil {
		t.Fatal(err)
	}
	event := PaymentEvent{ID: "fulfillment-integration-event", Type: "payment.succeeded", CreatedAt: now,
		Data:    PaymentEventData{IntentID: purchase.PaymentIntentID, Reference: purchase.Reference, AmountMinor: purchase.TotalMinor, Currency: purchase.Currency},
		Payload: []byte(`{"fulfillment":true}`)}
	purchase, err = repository.HandlePaymentEvent(ctx, event, now)
	if err != nil || purchase.Status != "paid" {
		t.Fatalf("paid purchase=%#v err=%v", purchase, err)
	}

	ordersByStore := map[string]SellerOrder{}
	for _, order := range purchase.SellerOrders {
		ordersByStore[order.StoreID] = order
	}
	firstOrder := ordersByStore["01989f00-0000-7000-8000-000000000201"]
	secondOrder := ordersByStore["01989f00-0000-7000-8000-000000000202"]
	if firstOrder.ID == "" || secondOrder.ID == "" {
		t.Fatalf("seller split missing: %#v", purchase.SellerOrders)
	}
	if _, err := repository.UpdateSellerOrder(ctx, secondSellerID, firstOrder.ID, "processing", "", now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("wrong seller error=%v", err)
	}
	for _, status := range []string{"processing", "shipped", "delivered"} {
		firstOrder, err = repository.UpdateSellerOrder(ctx, firstSellerID, firstOrder.ID, status, "", now)
		if err != nil || firstOrder.Status != status {
			t.Fatalf("transition %s order=%#v err=%v", status, firstOrder, err)
		}
	}
	secondOrder, err = repository.UpdateSellerOrder(ctx, secondSellerID, secondOrder.ID, "cancelled", "seller cannot fulfill", now)
	if err != nil || secondOrder.Status != "cancelled" {
		t.Fatalf("cancelled order=%#v err=%v", secondOrder, err)
	}
	var firstStock, secondStock int
	if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id=$1`, firstVariantID).Scan(&firstStock); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id=$1`, secondVariantID).Scan(&secondStock); err != nil {
		t.Fatal(err)
	}
	if firstStock != 9 || secondStock != 10 {
		t.Fatalf("delivered stock=%d cancelled stock=%d", firstStock, secondStock)
	}

	reviewInput := ReviewInput{PurchaseItemID: firstOrder.Items[0].ID, Rating: 5, Title: "Delivered safely", Body: "Verified purchase review."}
	review, err := repository.CreateReview(ctx, buyerID, reviewInput, now)
	if err != nil || review.Rating != 5 {
		t.Fatalf("review=%#v err=%v", review, err)
	}
	if _, err := repository.CreateReview(ctx, buyerID, reviewInput, now); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate review error=%v", err)
	}
	if _, err := repository.CreateReview(ctx, buyerID, ReviewInput{PurchaseItemID: secondOrder.Items[0].ID, Rating: 4, Title: "No delivery", Body: "Must be rejected."}, now); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("cancelled order review error=%v", err)
	}
	summary, err := repository.ReviewsByProductSlug(ctx, "handwoven-market-basket")
	if err != nil || summary.Count != 1 || summary.Average != 5 {
		t.Fatalf("review summary=%#v err=%v", summary, err)
	}

	if _, err := repository.SetCartItem(ctx, buyerID, firstVariantID, 1, now); err != nil {
		t.Fatal(err)
	}
	pending, created, err := repository.Checkout(ctx, buyerID, "cancellation-integration-checkout", now, now.Add(15*time.Minute))
	if err != nil || !created {
		t.Fatalf("pending checkout=%#v created=%v err=%v", pending, created, err)
	}
	cancelled, err := repository.CancelPurchase(ctx, buyerID, pending.ID, "buyer changed mind", now)
	if err != nil || cancelled.Status != "cancelled" || cancelled.PaymentStatus != "cancelled" {
		t.Fatalf("cancelled purchase=%#v err=%v", cancelled, err)
	}
	if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id=$1`, firstVariantID).Scan(&firstStock); err != nil {
		t.Fatal(err)
	}
	if firstStock != 9 {
		t.Fatalf("buyer cancellation stock=%d", firstStock)
	}
	overview, err := repository.AdminOverview(ctx)
	if err != nil || overview.DeliveredOrders < 1 || overview.Purchases < 2 {
		t.Fatalf("overview=%#v err=%v", overview, err)
	}
	audit, err := repository.AuditEvents(ctx)
	if err != nil || len(audit) < 6 {
		t.Fatalf("audit count=%d err=%v", len(audit), err)
	}
}
