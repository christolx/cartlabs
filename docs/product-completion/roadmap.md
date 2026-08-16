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
- Run manual visual review after each redesigned page, then check each page family and final cross-system consistency.

## Phase 0 - Backend Completion

- [x] Add admin user listing and safe account status management.
- [x] Add verified public store profiles and store-filtered catalog results.
- [x] Add owned seller product detail endpoint.
- [x] Synchronize OpenAPI and generated Go/TypeScript contracts.
- [x] Pass backend unit, integration, Compose, outage, and full checks.
- [x] Verify completed contracts match approved frontend page assumptions.

## Phase 1 - Foundation

- [x] Build the catalog-first homepage and login page.
- [x] Add durable sessions, role navigation, and shared UI states.
- [x] Pass foundation functional QA.

## Phase 2 - Buyer

- [x] Build catalog, store, product, cart, checkout, purchase, and review pages.
- [x] Complete buyer flow through normal navigation.
- [x] Pass buyer functional QA.

## Phase 3 - Seller

- [x] Build seller home, store, product, inventory, and order pages.
- [x] Complete seller supply and fulfillment flows through normal navigation.
- [x] Pass seller functional QA.

## Phase 4 - Admin

- [x] Build admin overview, store verification, listing enforcement,
  user-management, and audit pages.
- [x] Complete admin flow through normal navigation.
- [x] Pass admin functional QA.

## Phase 5 - Functional Integration

- [x] Replace `/demo` with role quick-login into real pages.
- [x] Remove `DemoConsole` and all legacy UI.
- [x] Add frontend and browser tests.
- [x] Exercise complete seller -> admin -> buyer journey in browser.
- [x] Check session reload, direct URLs, role boundaries, and failure states.
- [x] Check basic mobile/desktop usability and accessibility.
- [x] Pass integrated functional QA.

## Phase 6 - Visual Overhaul

- [ ] Choose art direction and create the frontend design system.
- [ ] Apply the approved visual system across every page and shared state.
- [ ] Complete responsive, interaction, accessibility, and content polish.
- [ ] Pass per-page **MANUAL VISUAL REVIEW**, family checkpoints, and final cross-system drift audit.

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
