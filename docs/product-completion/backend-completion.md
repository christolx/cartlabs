# Backend Completion Plan

## Goal

Close remaining backend gaps against the declared marketplace MVP before new
frontend development begins. Keep existing architecture, infrastructure, and
working end-to-end flows stable.

This plan covers missing product capabilities and one API ergonomics gap. It
does not reopen deferred commerce scope.

## Sources

- [`../product.md`](../product.md) defines MVP product intent.
- [`../e2e-flow.md`](../e2e-flow.md) remains the guide for existing marketplace
  behavior and must continue to pass unchanged.
- [`../../api/openapi/openapi.yaml`](../../api/openapi/openapi.yaml) remains the
  source of truth for HTTP contracts.
- [`page-plan.md`](./page-plan.md) defines target frontend assumptions that final
  backend contracts must satisfy before frontend implementation starts.

## Current Baseline

Already complete:

- Login, refresh rotation, logout, current user, RBAC, and demo login.
- Seller store creation/editing and admin store moderation.
- Product, variant, image metadata, inventory, publishing, and moderation.
- Public catalog search, filters, pagination, product detail, and reviews.
- Buyer cart, multi-seller checkout, mock payment, cancellation, and purchases.
- Seller fulfillment and per-seller order isolation.
- Notifications, transactional outbox, audit history, and admin overview.
- Existing automated, integration, Compose, operability, and platform checks.

Remaining gaps:

| Priority | Capability | Why needed |
| --- | --- | --- |
| Required | Admin user status management | Declared in `product.md`; database status exists but no service or API |
| Required | Public store profile | Declared in catalog MVP; only private seller/admin store APIs exist |
| Required | Public products filtered by store | Needed to make a public store profile useful |
| Required | Seller product detail endpoint | Removes list-fetch workaround for product management page |
| Follow-up | Contract and documentation reconciliation | Generated clients and frontend scope must match completed backend |

## Scope Guardrails

- Preserve existing routes and response behavior unless OpenAPI explicitly
  records a backward-compatible addition.
- Add no new service, queue, database, deployment component, or external
  provider.
- Reuse current identity, store, catalog, audit, HTTP validation, and
  authorization boundaries.
- Never expose password hashes, refresh sessions, seller email, or private
  moderation data through public store contracts.
- Do not build deferred features listed under Non-goals.

## Phase 1 — Contract Design

Update OpenAPI before implementation.

### Admin users

Add:

- `GET /admin/users`
  - Admin only.
  - Returns all users in deterministic newest-first order.
  - Response: `{ "items": AdminUser[] }`.
- `PATCH /admin/users/{userId}/status`
  - Admin only.
  - Request: `{ "status": "active" | "suspended", "reason": string }`.
  - Returns updated `AdminUser`.
  - Rejects invalid transition, self-suspension, missing user, and attempts that
    would leave no active admin.

Add schemas:

- `UserStatus`: `active | suspended`.
- `AdminUser`: `id`, `email`, `displayName`, `role`, `status`, `createdAt`,
  `updatedAt`.
- `UserStatusInput`: `status`, required trimmed `reason` with a bounded length.

Expected errors:

- `400` malformed identifier, invalid status, or invalid reason.
- `403` non-admin principal.
- `404` user not found.
- `409` unsafe or redundant transition.

### Public store profile

Add:

- `GET /catalog/stores/{slug}`
  - Public.
  - Returns approved store only.
  - Rejected, pending, and missing stores all return `404` publicly.
- `store` query parameter on `GET /catalog/products`.
  - Accepts store slug.
  - Composes with search, category, price, stock, and pagination filters.

Add schema:

- `StoreProfile`: `id`, `name`, `slug`, `description`, `sellerDisplayName`,
  `createdAt`.

Extend public catalog product responses with `storeSlug` so product cards and
details can link to the store profile without exposing seller identity data.

Expected errors:

- `400` invalid store slug or filter input.
- `404` store is missing or not approved.

### Seller product detail

Add:

- `GET /seller/products/{productId}`
  - Seller only.
  - Returns complete owned `ProductDetail`, including variants and images.
  - Returns `404` for missing or non-owned product to avoid ownership leakage.

Expected errors:

- `400` malformed identifier.
- `403` non-seller principal.
- `404` missing or non-owned product.

### Contract completion checklist

- [x] Define routes, parameters, schemas, examples, and problem responses.
- [x] Keep new object schemas closed with `additionalProperties: false`.
- [x] Run `make generate`.
- [x] Commit generated Go and TypeScript contracts with OpenAPI change.
- [x] Extend route/OpenAPI drift and request-validation tests.

## Phase 2 — Admin User Management

### Domain and repository

- [x] Expose user status and timestamps through an admin-safe domain model.
- [x] Add deterministic user listing to identity repository.
- [x] Add transactional user-status update.
- [x] Lock target user during status mutation.
- [x] Reject self-suspension.
- [x] Prevent removal of last active admin.
- [x] Revoke every active refresh session for suspended user in same transaction.
- [x] Write immutable audit event containing status transition and reason.
- [x] Keep reactivation sessionless; reactivated user must log in again.

No schema migration is currently expected. `users.status`, `updated_at`,
`refresh_sessions`, and `audit_log` already exist. Add a migration only if
implementation proves a missing constraint or index.

### Service and transport

- [x] Add admin-only identity service methods for list and status update.
- [x] Validate UUID, status enum, trimmed reason, and transition rules.
- [x] Add HTTP handlers and service wiring.
- [x] Map domain models to generated contract types explicitly.
- [x] Preserve generic unauthorized responses for suspended accounts.

### Security behavior

- [x] Suspended user cannot log in.
- [x] Suspended user cannot refresh.
- [x] Existing access token fails on next authenticated request.
- [x] All refresh families for suspended user become revoked.
- [x] Non-admin cannot list users or change status.
- [x] Public responses never expose user status or email through store pages.

### Tests

- [x] Table-driven service tests for RBAC, validation, transitions, self-action,
  and last-admin protection.
- [x] PostgreSQL integration test for atomic status change, session revocation,
  and audit insertion.
- [x] HTTP tests for success and all documented error mappings.
- [x] OpenAPI validation tests for unknown fields and invalid enum values.
- [x] Compose/Hurl flow: suspend user, reject existing session, reactivate user,
  require fresh login, and inspect audit event.

## Phase 3 — Public Store Profiles

### Domain and repository

- [x] Add public store-profile model separate from private moderation model.
- [x] Query approved store by slug and join only seller display name.
- [x] Return not found for pending or rejected store.
- [x] Add store slug to catalog product domain summaries/details.
- [x] Add optional store-slug filter to catalog filters.
- [x] Apply store filter in both normal SQL and search-candidate catalog paths.
- [x] Preserve catalog-authoritative moderation and publication checks.

No schema migration is currently expected. Store slug already has a unique
constraint, and current catalog queries already join stores and users can be
joined by seller ID.

### Service and transport

- [x] Validate and normalize public store slug.
- [x] Add public store-profile service method and HTTP handler.
- [x] Parse `store` catalog query parameter.
- [x] Map `StoreProfile` and `storeSlug` without private fields.
- [x] Register new route in HTTP server and OpenAPI router.

### Tests

- [x] Service tests for approved, pending, rejected, malformed, and missing
  store lookup.
- [x] Repository integration tests proving only approved store and approved
  published products become public.
- [x] Catalog tests for store filter combined with search and existing filters.
- [x] Search outage test proving store filter matches service and SQL fallback.
- [x] HTTP/OpenAPI tests proving public access and response privacy.
- [x] Compose/Hurl flow from approved product to store profile and filtered
  store catalog.

## Phase 4 — Seller Product Detail

Existing repository method `FindForSeller` already loads variants and images.
Expose it through service and HTTP layers rather than adding a second query
path.

- [x] Add seller-only service method using current ownership lookup.
- [x] Add handler for `GET /seller/products/{productId}`.
- [x] Return `404` for missing and non-owned product.
- [x] Reuse existing `ProductDetail` contract mapping.
- [x] Add service ownership/RBAC tests.
- [x] Add HTTP success, malformed ID, forbidden, and not-found tests.
- [x] Add Compose/Hurl ownership check with two sellers.

No schema migration is expected.

## Phase 5 — Cross-cutting Verification

- [x] Add new privileged mutation to audit expectations and observability checks.
- [x] Confirm generic HTTP metrics cover new bounded route patterns.
- [x] Confirm security headers, request IDs, request size limits, rate limits,
  panic recovery, and OpenAPI request validation cover new handlers.
- [x] Update deterministic seeds only if tests require a second admin; never add
  nondeterministic fixture data.
- [x] Run `gofmt` on changed Go files.
- [x] Run focused Go unit and integration tests.
- [x] Run `make generate` and confirm no generated drift.
- [x] Run `make test` and `make lint`.
- [x] Run Compose-backed Hurl and search-outage workflows.
- [x] Run `make check`.

## Phase 6 — Documentation and Frontend Handoff

- [x] Update `product.md` only if final implementation intentionally differs
  from declared MVP.
- [x] Update API examples and permanent architecture/behavior docs where new
  behavior requires it.
- [x] Keep `e2e-flow.md` unchanged unless its existing guide becomes factually
  wrong; backend completion must not repurpose it as a temporary work plan.
- [x] Verify completed contracts satisfy `page-plan.md` backend assumptions.
- [x] Update `page-plan.md` only if final contracts intentionally differ, then
  reapprove affected frontend scope.
- [x] Mark backend completion gate in `roadmap.md` complete only after all checks
  pass.

## Non-goals

- Registration, password reset, account deletion, or profile editing.
- Real payment or shipping providers.
- Address book, shipping quote, tax, refund, or dispute systems.
- Binary media upload or media processing pipeline.
- Variant/image deletion or full catalog lifecycle expansion.
- Notification read state, preferences, or deletion.
- Promotions, wishlists, chat, recommendations, or personalization.
- New microservices, queues, databases, infrastructure, or deployment targets.

## Backend Done

Backend completion requires:

- Admin can list users and safely suspend/reactivate accounts with audit history
  and immediate session enforcement.
- Public visitor can open approved store profile and browse that store's
  approved published products.
- Seller can fetch one complete owned product without loading full product list.
- Existing seller, admin, buyer, payment, worker, and search behavior remains
  unchanged.
- OpenAPI and generated clients match runtime routes.
- Focused, integration, Compose, outage, and full repository checks pass.
- Frontend page-plan assumptions are verified against final contracts.
