# Phase 7 Regression

## Actor Journeys

- Anonymous: catalog, filters, store, product, login, demo entry.
- Buyer: login, cart changes, checkout, mock payment, purchases, cancellation, review, notifications.
- Seller: login, store setup, product creation, variants, images, inventory, publication, fulfillment, notifications.
- Admin: login, overview, store moderation, listing enforcement, user status, audit history, notifications.

## Boundary Checks

- Reload authenticated routes.
- Open direct URLs for every role.
- Verify anonymous and wrong-role rejection.
- Verify suspended-user session behavior.
- Verify unavailable API, invalid input, empty data, and failed mutations.
- Verify browser back, forward, and repeated submissions.

## Viewports and Modes

- Run browser tests at mobile and desktop widths.
- Review light and dark modes.
- Verify keyboard-only navigation and reduced motion.
- Check no content clipping, accidental overflow, or layout shift.

## Commands

```bash
pnpm --dir apps/web typecheck
pnpm --dir apps/web lint
pnpm --dir apps/web test
pnpm --dir apps/web test:e2e
make check
```

## Completion

- All actor journeys pass.
- No route, action, form, state, or role boundary regresses.
- Every page retains complete image-to-code pipeline evidence and product-owner approval.
- Manual visual review approved; shared changes did not introduce unreviewed drift from approved references.
- `make check` passes.
