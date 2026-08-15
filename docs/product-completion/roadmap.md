# Product Completion Roadmap

## Goal

Close remaining backend scope, then finish the marketplace product with a fully
redesigned frontend.

## Rules

- Freeze platform expansion; keep current Compose and deployment setup.
- Complete [`backend-completion.md`](./backend-completion.md) before new
  frontend implementation.
- Lock page purpose, functionality, and state behavior before visual direction.
- Build the complete functional product with a minimal coherent UI first.
- Start final visual direction from zero after functional browser QA passes.
- Replace every existing page and UI, including `DemoConsole`.
- Preserve working functionality, API contracts, and domain behavior.
- Run functional QA on every actor slice before moving forward.
- Run manual visual review after the visual overhaul, not per functional slice.

## Phase 0 - Backend Completion

- [ ] Add admin user listing and safe account status management.
- [ ] Add approved public store profiles and store-filtered catalog results.
- [ ] Add owned seller product detail endpoint.
- [ ] Synchronize OpenAPI and generated Go/TypeScript contracts.
- [ ] Pass backend unit, integration, Compose, outage, and full checks.
- [ ] Verify completed contracts match approved frontend page assumptions.

## Phase 1 - Foundation

- [ ] Build functional landing and login pages.
- [ ] Add durable sessions, role navigation, and shared UI states.
- [ ] Pass foundation functional QA.

## Phase 2 - Buyer

- [ ] Build catalog, store, product, cart, checkout, purchase, and review pages.
- [ ] Complete buyer flow through normal navigation.
- [ ] Pass buyer functional QA.

## Phase 3 - Seller

- [ ] Build seller home, store, product, inventory, and order pages.
- [ ] Complete seller supply and fulfillment flows through normal navigation.
- [ ] Pass seller functional QA.

## Phase 4 - Admin

- [ ] Build admin overview, moderation, user-management, and audit pages.
- [ ] Complete admin flow through normal navigation.
- [ ] Pass admin functional QA.

## Phase 5 - Functional Integration

- [ ] Replace `/demo` with role quick-login into real pages.
- [ ] Remove `DemoConsole` and all legacy UI.
- [ ] Add frontend and browser tests.
- [ ] Exercise complete seller -> admin -> buyer journey in browser.
- [ ] Check session reload, direct URLs, role boundaries, and failure states.
- [ ] Check basic mobile/desktop usability and accessibility.
- [ ] Pass integrated functional QA.

## Phase 6 - Visual Overhaul

- [ ] Choose art direction and create the frontend design system.
- [ ] Apply the approved visual system across every page and shared state.
- [ ] Complete responsive, interaction, accessibility, and content polish.
- [ ] Pass **MANUAL VISUAL REVIEW**.

## Phase 7 - Final Regression

- [ ] Re-run every actor journey after the visual overhaul.
- [ ] Verify no route, action, form, state, or role boundary regressed.
- [ ] Re-run frontend/browser tests at mobile and desktop widths.
- [ ] Run `make check`.

## Done

Product works through normal navigation, survives reload, handles failures,
passes functional and visual QA, passes final regression, and receives
**MANUAL VISUAL REVIEW** approval. Use [`e2e-flow.md`](../e2e-flow.md) for
marketplace behavior and [validation.md](./validation.md) for release checks.
