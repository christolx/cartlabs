# Review and Demo Seed Fix Plan

## 1. Persist review eligibility

- Add review state to purchase items (`reviewed` or `reviewId`).
- Populate it in purchase queries and OpenAPI contracts.
- Render review form only for delivered, unreviewed items.
- Replace raw `409` text with a clear already-reviewed message.

## 2. Repair lifecycle seed data

- Add one inventory reservation per seeded purchase item.
- Use `active`, `converted`, or `released` status matching purchase lifecycle.
- Verify paid and processing seller-order cancellation restores stock.

## 3. Improve demo coverage

- Distribute purchases across products, stores, buyers, and lifecycle states.
- Add several delivered purchases and reviewable items for primary buyer.
- Seed reviews across multiple products and ratings.
- Include one multi-store purchase.

## 4. Align events and audits

- Implement review notification events or remove unreachable seeded notifications.
- Match seeded audit action and resource names to runtime values.

## 5. Verify

- Test existing review hidden on first load and reload.
- Test eligible review succeeds once; duplicate returns friendly UI state.
- Test seeded paid and processing cancellations.
- Run `make seed` twice, `make test`, `make web-e2e`, and `make check`.
