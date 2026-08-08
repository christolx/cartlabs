# Fulfillment and Trust

Milestone 3 extends each paid seller order independently while preserving the
purchase and inventory transaction boundaries established in Milestone 2.

## Seller order state machine

Allowed transitions are intentionally narrow:

```text
paid ──> processing ──> shipped ──> delivered
  └──────────┐
             v
          cancelled <── processing
```

Cancellation is unavailable after shipment. Every transition locks the seller
order, verifies store ownership, timestamps the new state, appends an audit
fact, and writes an outbox event in the same database transaction.

## Cancellation invariants

- Buyer cancellation accepts only pending-payment or paid purchases whose
  seller orders have not started processing.
- Seller cancellation accepts only that seller's paid or processing order.
- Each affected reservation moves from active or converted to released once.
- Stock restoration and inventory-ledger insertion occur atomically and in
  stable variant order.
- Cancelling one seller order does not alter another seller's fulfillment state.
- Parent purchase becomes cancelled when every child order is cancelled.

## Verified reviews

A review references one immutable purchase item. Repository authorization joins
purchase ownership with a delivered seller order. Database uniqueness on
`purchase_item_id` permits one review per purchased line. Public product review
responses return newest reviews, count, and arithmetic mean rating.

## Admin visibility

Overview reports users, approved stores, published products, purchases, active
seller orders, delivered orders, and active gross merchandise value. Admin
audit view reads the existing append-only `audit_log`, joining actor identity and
role so catalog moderation and fulfillment activity remain one chronological
record.

## Event delivery

Fulfillment transactions emit `order.processing`, `order.shipped`,
`order.delivered`, `order.cancelled`, or `purchase.cancelled`. Worker queue binds
both `purchase.#` and `order.#`; existing publisher confirms, manual consumer
acknowledgements, and `(user_id, source_event_id)` uniqueness retain at-least-once
delivery without duplicate notifications.
