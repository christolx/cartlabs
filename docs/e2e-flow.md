# End-to-End Marketplace Flows

Use this document when running, explaining, or extending Cartlabs role
workflows. Exact request and response shapes live in
[`api/openapi/openapi.yaml`](../api/openapi/openapi.yaml).

## Actors

| Actor | Primary responsibility |
| --- | --- |
| Buyer | Discover products, purchase, track delivery, review delivered items. |
| Seller | Create approved catalog supply, manage stock, fulfill owned orders. |
| Admin | Moderate stores/products, inspect platform totals and audit history. |
| Payment provider | Confirm or fail payment through signed webhook. |
| Worker | Deliver durable purchase and order notifications. |

## Marketplace happy path

```text
Seller creates store
  -> Admin approves store
  -> Seller creates product, variant, image, stock
  -> Seller submits product
  -> Admin approves product
  -> Buyer discovers product and adds variant to cart
  -> Buyer checks out
  -> System reserves stock and creates payment intent
  -> Payment provider confirms payment
  -> Seller fulfills own order
  -> Buyer tracks delivery and reviews delivered item
```

## Seller flow

1. Sign in as seller.
2. Create store with name, slug, and description. Store begins `pending`.
3. Admin approves store. Only approved store can create product supply.
4. Create product draft, then add at least one variant/SKU with price and stock,
   plus product image.
5. Submit product for moderation. Product remains unavailable to public catalog
   until admin approval.
6. Adjust stock through seller inventory endpoint. Seller may only change owned
   variants.
7. After buyer payment, list seller orders. Each order belongs to one store.
8. Fulfill each owned order independently:

```text
paid -> processing -> shipped -> delivered
paid -> cancelled
processing -> cancelled
```

Seller cannot change another seller's order. Cancellation is unavailable after
shipment.

## Admin flow

1. Sign in as admin.
2. Inspect pending stores. Approve or reject each store with moderation note.
3. Inspect submitted products. Approve or reject each product with moderation
   note.
4. Inspect marketplace overview: users, approved stores, published products,
   purchases, active/delivered orders, and gross merchandise value.
5. Inspect audit events for moderation and fulfillment history.

Admin does not act as buyer or seller. Moderation gates public supply.

## Buyer flow

1. Browse/search approved public products. View live variants, price, stock,
   and public verified reviews.
2. Add variants to cart or replace quantity. Cart groups lines by store and
   recalculates against live price and inventory.
3. Checkout with `Idempotency-Key`.
4. Checkout transaction:
   - locks cart and variant inventory;
   - reserves stock for 15 minutes;
   - creates one parent purchase;
   - creates one child seller order for each participating store;
   - writes immutable product/price item snapshots;
   - clears cart;
   - records notification event for later delivery.
5. Payment provider completes mock payment. Signed webhook transitions purchase
   and seller orders:

```text
pending_payment --payment succeeded--> paid
pending_payment --payment failed-----> payment_failed
pending_payment --reservation timeout-> expired
```

6. Track each seller order separately. One seller may ship or cancel without
   changing another seller's fulfillment state.
7. Cancel full purchase only while every child order remains `pending_payment`
   or `paid`. System releases affected stock once.
8. After seller marks an order `delivered`, create one verified review per
   purchased item.

## Multi-seller purchase

Buyer cart can contain variants from multiple stores. Checkout preserves a
shared payment boundary while separating fulfillment:

```text
Purchase P-123
|- Seller order: Store A
`- Seller order: Store B
```

Payment succeeds once for parent purchase. Store A and Store B advance only
their own child order. Parent purchase becomes `cancelled` only after every
child order is cancelled.

## Notifications and replay safety

Payment and fulfillment transactions write outbox events in same database
transaction as state change. Worker publishes event after commit. Notification
storage deduplicates `(user_id, source_event_id)`, so provider/webhook retries
or worker republishes do not create duplicate user notifications.

## Local demo

1. Set `DEMO_MODE=true`.
2. Start services and seed data; see repository [README](../README.md).
3. Open `/demo`.
4. Run roles in order: seller supply -> admin approval -> buyer purchase ->
   seller fulfillment -> buyer review.

Demo accounts use password `demo-pass-123`:

| Role | Email |
| --- | --- |
| Buyer | `buyer@demo.cartlabs.local` |
| Seller | `seller@demo.cartlabs.local` |
| Admin | `admin@demo.cartlabs.local` |
