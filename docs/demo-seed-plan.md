# Demo Seed Plan

## Goal

Show one coherent marketplace story across buyer, seller, admin, and public catalog. Periodic database reset restores baseline.

## Seed Data

### Buyer

- Cart with two in-stock products from different stores.
- Purchases covering awaiting payment, active fulfillment, delivered, and cancelled or failed payment.
- Read and unread notifications.
- One reviewed delivered item and one item eligible for review.

### Seller

- Demo Seller products covering draft, published, archived, and suspended states.
- Variants with healthy, low, and zero stock.
- Orders covering paid, processing, shipped, delivered, and cancelled states.
- Fulfillment timestamps, cancellation reason, and notifications.

### Admin

- Approved, pending, and rejected stores; rejected store includes moderation note.
- Published and suspended products; suspended product includes reason and actor.
- Active and suspended non-demo users.
- Audit events for verification, moderation, user status, inventory, fulfillment, cancellation, and review.
- Purchase and order data sufficient for non-zero overview totals and GMV.

### Public Catalog

- Verified reviews with varied ratings.
- Existing image, variant, attribute, and stock diversity preserved.

## Consistency

- Buyer purchases reference Demo Seller products.
- Seller orders match buyer purchases.
- Admin audit events reference same users, stores, products, and orders.
- Use fixed IDs and idempotent upserts.
- Use realistic relative timestamps and valid state transitions.
- Keep quick-login accounts active; use secondary accounts for suspended examples.

## Verification

- Run `make reset` and `make check`.
- Open every buyer, seller, and admin workspace page.
- Confirm no unexplained empty primary view.
- Confirm seeded actions remain usable until next periodic reset.
