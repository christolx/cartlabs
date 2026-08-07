# Purchase Vertical Slice

Milestone 2 keeps cart, checkout, purchase, payment state, and inventory reservation
inside one PostgreSQL consistency boundary. RabbitMQ handles downstream notification
delivery only after transaction commit.

## Checkout invariants

- Buyer owns one cart. Cart reads live variant prices and stock, grouped by store.
- Checkout requires an `Idempotency-Key` scoped to buyer. Reusing it returns same
  purchase without reserving inventory twice or creating another payment intent.
- Checkout locks cart row, then variant rows ordered by variant ID. This stable lock
  order and conditional stock update prevent overselling under concurrent checkout.
- One database transaction creates parent purchase, one seller order per store,
  immutable item snapshots, inventory reservations, inventory ledger entries, and
  `purchase.created` outbox event before clearing cart.
- Money remains integer IDR minor units. Parent total equals seller-order and item
  snapshot totals.
- Active reservation removes stock immediately and expires after 15 minutes. Paid
  purchase converts reservation; failed or expired purchase releases stock once.

## Payment boundary

API creates intent through provider interface after checkout transaction commits.
If provider setup fails or returns mismatched intent data, purchase becomes
`payment_failed` and reservation releases.

Mock provider requires service bearer credential. Completion emits `payment.succeeded`
or `payment.failed` to configured callback. Signature format is:

```text
t=<unix-seconds>,v1=<hex-hmac-sha256>
```

Signed message is `<unix-seconds>.<raw-json-body>`. API rejects timestamps outside
five-minute tolerance, malformed signatures, wrong amount/currency/reference, and
unknown intent. `payment_events.event_id` makes callback replay idempotent.

## State transitions

```text
pending_payment --payment.succeeded--> paid
pending_payment --payment.failed-----> payment_failed
pending_payment --reservation timeout-> expired
```

Seller orders move from `pending_payment` to `paid` or `cancelled`. Terminal purchase
states do not transition again.

## Outbox and notifications

Purchase state transaction inserts outbox fact with recipient IDs. Worker claims one
unpublished row using `FOR UPDATE SKIP LOCKED`, publishes persistent message to durable
RabbitMQ topic exchange, waits for publisher confirmation, then marks row published.
Commit failure can republish; notification consumer remains safe because
`(user_id, source_event_id)` is unique. Consumer manually acknowledges success,
discards permanently invalid payloads, and requeues transient database failures.

## Local verification

```bash
make reset
docker compose -f deploy/compose/compose.yml --profile full up -d --build
make e2e-api
INTEGRATION_DATABASE_URL='postgres://cartlabs:cartlabs@localhost:5432/cartlabs?sslmode=disable' \
  go test -tags=integration ./internal/purchase -run TestConcurrentCheckoutPreventsOversell -count=1
make check
```

Demo web workflow lives at `http://localhost:3000/demo`. Choose buyer, add products
from two stores, reserve checkout, complete payment, then switch to participating
seller to inspect paid order and notification.
