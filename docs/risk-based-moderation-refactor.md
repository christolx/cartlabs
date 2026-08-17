# Risk-Based Marketplace Moderation Refactor

Status: Implemented on 2026-08-16.

## Summary

Move from pre-approving every product to trust-based publishing: admin verifies
stores, verified sellers publish complete products immediately, and admin removes
unsafe listings through explicit enforcement. Admin retains store verification,
user enforcement, marketplace overview, and immutable audit history.

This refactor does not introduce automated risk scoring. “Risk-based” means
moderation effort moves from every listing to store verification and reactive
listing enforcement.

## Goals

- Remove product-by-product approval from normal seller publishing.
- Preserve store verification as prerequisite for creating and publishing supply.
- Give admin reversible listing suspension with mandatory reason and audit trail.
- Keep public catalog, cart, and checkout eligibility consistent.
- Preserve clear ownership: seller controls drafts and archives; admin controls
  suspension and reinstatement.

## Non-Goals

- KYC, document review, payout verification, or automated policy/risk scoring.
- Product-version review or delayed publication after edits.
- New media upload, variant editing/deletion, or inventory workflow.
- Historical analytics, case management, appeals, or enforcement notifications.

## Store Verification

- Keep store status `pending | approved | rejected` and existing persistence
  fields. Rename admin/store UI copy from “moderation” to “verification.”
- New store starts `pending`. Seller cannot create or publish products until store
  is `approved`.
- Approved store profile edits stay `approved` and public.
- Pending store edits stay `pending`. Rejected store edits return to `pending`,
  clear previous verification note, and require a new admin decision.
- Admin verification endpoint remains
  `PATCH /admin/stores/{storeId}/moderation` for contract compatibility in this
  refactor. UI calls action “verification.” Request remains
  `{ status: "approved" | "rejected", note }`.
- Trim `note`; require non-empty note for rejection and allow empty note for
  approval. Maximum length remains 500 characters.
- Every decision appends audit metadata containing `from`, `to`, and `note`.
- Moving an approved store to `rejected` immediately hides its products and makes
  them ineligible for cart additions and checkout. Product states remain unchanged;
  re-approving store restores eligibility for products still `published`.

## Product Lifecycle

Product status becomes `draft | published | archived | suspended`.
Product moderation status and note are removed.

| Current | Actor | Action | Next | Requirements |
| --- | --- | --- | --- | --- |
| `draft` | Seller | Publish | `published` | Owned product, approved store, at least one active variant with stock greater than zero, and at least one image |
| `published` | Seller | Archive | `archived` | Owned product |
| `archived` | Seller | Publish | `published` | Same completeness and approved-store checks as draft publish |
| `published` | Admin | Suspend | `suspended` | Non-empty reason |
| `suspended` | Admin | Reinstate | `published` | Non-empty reason, approved store, and same completeness checks as seller publish |

All other transitions return `409 Conflict`. In particular:

- Seller cannot publish, archive, or otherwise change status of a suspended
  product.
- Seller publish endpoint accepts only `draft` or `archived` products.
- Admin status endpoint accepts suspension only from `published` and reinstatement
  only from `suspended`.
- Repeating current status is a conflict, not a successful no-op.

Product identity edits, new variants, new images, and inventory adjustments do
not change status. Published products remain published; draft products remain
draft; archived products remain archived; suspended products remain suspended.
Sellers may edit product content and inventory while suspended, but only admin can
restore publication.

Completeness is checked when entering `published`. Later stock depletion may leave
a published product out of stock without changing status. Existing catalog
`inStock` behavior remains.

## API Contract

Update OpenAPI first, regenerate Go/web contracts, then adapt domain, repository,
HTTP, UI, and docs.

### Seller endpoints

- Keep `POST /seller/products/{productId}/publish`.
  - Publish immediately; remove review/submission wording.
  - Accept source status `draft | archived` only.
  - Return `409` for incomplete product, unapproved store, suspended product, or
    invalid source state.
- Add `POST /seller/products/{productId}/archive`.
  - Accept source status `published` only.
  - Return updated product or `409` for invalid source state.
- Existing product, variant, image, and inventory mutations preserve current
  product status.

### Admin endpoints

- Remove `PATCH /admin/products/{productId}/moderation`.
- Add `PATCH /admin/products/{productId}/status` with body:

  ```json
  {
    "status": "suspended",
    "reason": "Repeated marketplace policy violations"
  }
  ```

- `status` accepts `suspended | published`. Trim `reason`; require 1–500
  characters. Reject unknown fields.
- Require admin role. Return `404` for missing product and `409` for invalid
  transition, unapproved store, or incomplete reinstatement.
- Keep `GET /admin/products`, but return only `published | suspended` listings,
  newest enforcement activity first. Each item includes:
  - existing product detail;
  - `enforcementReason`: latest suspension reason, empty when product has never
    been suspended;
  - `enforcedAt`: latest suspension timestamp, nullable;
  - `enforcedBy`: latest suspending admin display name, nullable.
- Reinstatement does not erase suspension context. Console can show previous
  suspension reason beside current status.

### Product schema

- Change product `status` enum to `draft | published | archived | suspended`.
- Remove `moderationStatus` and `moderationNote` from product responses.
- Add admin-only enforcement fields through a dedicated admin product schema;
  do not expose enforcement metadata in public or seller responses.

## Persistence and Audit

- Add `suspended` to `product_status`.
- Drop `products.moderation_status` and `products.moderation_note` only after data
  conversion and audit backfill.
- Store enforcement history in existing append-only `audit_log`; do not add a
  mutable “current reason” column to products.
- Product status events use actions:
  - `product.status.suspended`
  - `product.status.published`
- Each admin event metadata contains `from`, `to`, and `reason`.
- Seller publish/archive events contain `from` and `to`.
- Enforcement mutation and audit insertion occur in one PostgreSQL transaction.
- Admin listing query obtains latest suspension event per product from
  `audit_log`. Add supporting index only if query plan requires one beyond current
  `(resource_type, resource_id, created_at DESC)` index.
- Product status changes append `catalog.search.upsert.v1` in same transaction so
  search projection stays current. Catalog remains authoritative and reapplies
  product/store eligibility to every candidate result.

## Migration

Create next ordered up/down migration. Up migration performs conversion before
dropping columns.

| Legacy product status | Legacy moderation status | New status |
| --- | --- | --- |
| `draft` | Any | `draft` |
| `published` | `approved` | `published` |
| `published` | `pending` | `published` |
| `published` | `rejected` | `draft` |
| `archived` | Any | `archived` |

Before dropping moderation fields:

- Backfill one audit event for each product with non-empty moderation note. Store
  legacy status and note in metadata; use a deterministic migration action such as
  `product.moderation.migrated`.
- Assert no null or unmapped state remains.
- Update product status, then drop moderation columns.
- Keep store verification records unchanged.

Down migration restores moderation columns with safe defaults but cannot recreate
their exact former row values. Reconstruct:

- `published` products as moderation `approved`;
- `draft | archived` products as moderation `pending`;
- `suspended` products as status `published`, moderation `rejected`, with latest
  suspension reason copied into moderation note when available.

Document downgrade as lossy. Preserve exact legacy decisions in audit history.

## UI Changes

- Rename store-facing moderation copy to verification copy.
- Seller product list/detail shows one lifecycle status, no moderation status or
  note.
- Publish action says “Publish”; archived products offer “Republish”; published
  products offer “Archive.” Suspended products show enforcement state and no
  seller status action.
- Replace admin product review queue with listing-enforcement console for
  published/suspended products. Show current status, latest suspension reason,
  actor, timestamp, and allowed action.
- Remove product approval/rejection controls and review-submission language.
- Update public empty states and product copy to remove “approved product” wording.

## Implementation Order

1. Update OpenAPI schemas, operations, responses, and validation examples.
2. Regenerate Go and web contracts.
3. Add migration with state conversion, audit backfill, and down migration.
4. Update catalog domain and repository transitions, audit metadata, and search
   outbox writes.
5. Update public catalog, product detail, cart, checkout, overview, and review
   eligibility queries to use product status plus approved store status.
6. Replace HTTP moderation handler with admin enforcement handler; add seller
   archive handler.
7. Update seller/admin/public UI and copy.
8. Update architecture, search, milestone, completion, and E2E documentation.
9. Regenerate fixtures where needed and run full verification.

## Test Plan

- Store creation starts pending; seller cannot create or publish products before
  approval.
- Approved store becomes public. Approved-store edit remains approved/public.
  Pending edit stays pending; rejected edit returns to pending.
- Store rejection/approval immediately removes/restores eligibility for unchanged
  published products.
- Draft or archived publish requires approved store, active in-stock variant, and
  image; success becomes public immediately.
- Published product edit remains published. Draft, archived, and suspended edits
  preserve their respective states.
- Seller can archive published product and republish complete archived product.
- Seller cannot publish/archive suspended product or bypass suspension through any
  product mutation.
- Admin can suspend only published product and reinstate only suspended product.
  Reinstatement revalidates store approval and completeness.
- Suspended product is absent from catalog and product detail, cannot be added to
  cart, and fails checkout when already present. Reinstate restores eligibility.
- Cart display may retain unavailable line for removal, but checkout must reject
  it; behavior remains explicit and tested.
- Product enforcement requires admin role and trimmed 1–500 character reason.
  Unknown fields and invalid transitions fail.
- Enforcement transaction writes status, audit metadata, and search outbox event
  atomically.
- Admin product list returns published/suspended products plus latest suspension
  context without exposing metadata through seller/public schemas.
- Store verification and user status decisions preserve their role, validation,
  and audit behavior.
- Migration test covers every legacy status pair, note backfill, no unmapped rows,
  and documented lossy downgrade mapping.
- Run `make check`.

## Acceptance Criteria

- Verified seller publishes complete product without admin product approval.
- Seller cannot reverse admin suspension.
- Public catalog, detail, cart mutation, and checkout use same listing eligibility:
  product `published`, store `approved`, active variant, and sufficient stock where
  operation requires stock.
- Every admin product status change has immutable actor, timestamp, transition,
  and reason.
- Admin can inspect current listing status and latest suspension context.
- Generated contracts contain no product moderation fields or endpoint.
- Migration preserves legacy moderation context in audit history.
- `make check` passes with synchronized generated artifacts and docs.
