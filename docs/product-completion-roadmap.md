# Product Completion Roadmap

## Goal

Finish the marketplace product with a fully redesigned frontend.

## Rules

- Freeze platform expansion; keep current Compose and deployment setup.
- Start frontend visual direction from zero. Design direction remains pending.
- Replace every existing page and UI, including `DemoConsole`.
- Preserve working API contracts and domain behavior.
- Every frontend slice must pass **MANUAL REVIEW**.

## Phase 1 - Foundation

- [ ] Confirm MVP scope and API gaps.
- [ ] Choose visual direction and create frontend design foundation.
- [ ] Build redesigned landing and login pages.
- [ ] Add durable sessions, role navigation, and shared UI states.

## Phase 2 - Buyer

- [ ] Build redesigned buyer pages.
- [ ] Build catalog, product, cart, checkout, order tracking, and review flows.
- [ ] Replace current public pages and shared buyer UI.
- [ ] Pass **MANUAL REVIEW**.

## Phase 3 - Seller

- [ ] Build redesigned seller pages.
- [ ] Build store, product, inventory, and fulfillment flows.
- [ ] Pass **MANUAL REVIEW**.

## Phase 4 - Admin

- [ ] Build redesigned admin pages.
- [ ] Build overview, moderation, and audit flows.
- [ ] Resolve admin user-management scope.
- [ ] Pass **MANUAL REVIEW**.

## Phase 5 - Hardening

- [ ] Replace `/demo` with role quick-login into real pages.
- [ ] Remove `DemoConsole` and all legacy UI.
- [ ] Add frontend and browser tests.
- [ ] Complete responsive, accessibility, and failure-state review.
- [ ] Run `make check`.

## Done

Product works through normal navigation, survives reload, handles failures, passes tests, and receives **MANUAL REVIEW** approval.
