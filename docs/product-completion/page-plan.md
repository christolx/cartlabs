# Frontend Page and Feature Plan

## Purpose

Define the frontend surface and required behavior before visual design begins.
This document decides which pages the completed product needs and what each
page must accomplish. It does not decide layout, component composition,
typography, color, imagery, motion, or other art direction.

Status: approved target frontend scope. It assumes
[`backend-completion.md`](./backend-completion.md) has finished and its OpenAPI
contracts are available before frontend implementation begins.

Use these documents alongside this plan:

- [`../product.md`](../product.md) defines product intent and MVP boundaries.
- [`../e2e-flow.md`](../e2e-flow.md) remains the guide for marketplace behavior.
- [`../../api/openapi/openapi.yaml`](../../api/openapi/openapi.yaml) remains the
  source of truth for exact API contracts.
- [`roadmap.md`](./roadmap.md) defines implementation order.
- [`validation.md`](./validation.md) defines release checks.

## Scope Rules

- Redesign every existing UI while preserving working functionality.
- Build only behavior supported by completed OpenAPI contracts.
- Keep payment webhook, health/readiness, worker, and outbox operations outside
  browser UI.
- A page may use sections, dialogs, drawers, or dedicated subpages. Art
  direction will decide that composition later without removing required
  behavior.
- All protected pages enforce authentication and role authorization on reload
  and direct navigation, not only through hidden links.
- Demo shortcuts enter normal product pages. They must not recreate workflows
  in a separate console.

## Route Summary

| Area | Route | Access | Purpose |
| --- | --- | --- | --- |
| Public | `/` | Everyone | Landing and catalog discovery |
| Public | `/products/[slug]` | Everyone | Product detail and verified reviews |
| Public | `/stores/[slug]` | Everyone | Approved store profile and product catalog |
| Identity | `/login` | Anonymous | Password and demo-role login |
| Shared | `/notifications` | Signed-in users | Durable actor notifications |
| Buyer | `/cart` | Buyer | Review and update cart |
| Buyer | `/checkout` | Buyer | Confirm and submit checkout |
| Buyer | `/purchases` | Buyer | Purchase history |
| Buyer | `/purchases/[purchaseId]` | Buyer | Payment, tracking, cancellation, and review |
| Seller | `/seller` | Seller | Seller status and next actions |
| Seller | `/seller/store` | Seller | Store creation and editing |
| Seller | `/seller/products` | Seller | Owned product list |
| Seller | `/seller/products/new` | Seller | Draft product creation |
| Seller | `/seller/products/[productId]` | Seller | Product, variant, image, and inventory management |
| Seller | `/seller/orders` | Seller | Owned order fulfillment |
| Admin | `/admin` | Admin | Marketplace overview |
| Admin | `/admin/stores` | Admin | Store moderation |
| Admin | `/admin/products` | Admin | Product moderation |
| Admin | `/admin/users` | Admin | User status management |
| Admin | `/admin/audit` | Admin | Privileged-action history |
| Demo | `/demo` | Everyone | Demo explanation and role entry |

## Global Product Shell

Required across applicable pages:

- Navigation reflects anonymous, buyer, seller, and admin state.
- Signed-in navigation identifies current user and exposes logout.
- Buyer navigation exposes cart, purchases, and notifications.
- Seller navigation exposes store, products, orders, and notifications.
- Admin navigation exposes overview, moderation queues, user management, audit
  history, and notifications.
- Unknown routes use a useful not-found state.
- Protected routes distinguish unauthenticated, forbidden, and unavailable
  states.
- Actions expose pending, success, validation, conflict, and dependency-failure
  feedback where relevant.
- Responsive navigation and primary actions remain usable at mobile and desktop
  widths.

## Identity and Session Contract

Identity behavior is product-wide rather than owned by one page:

- Login accepts email and password through `POST /auth/login`.
- Demo quick login uses `POST /auth/demo-login` only when demo mode is enabled.
- Current identity is restored through refresh plus `GET /me` after reload.
- Access tokens remain in runtime memory; refresh cookie remains HTTP-only.
- Expired access triggers one refresh attempt and safe request retry.
- Failed refresh clears local identity and sends protected navigation to login.
- Logout calls `POST /auth/logout`, clears local identity, and returns to a
  public page.
- Login preserves a safe intended destination when authentication was required
  by a user action or protected route.
- Wrong-role direct navigation renders or redirects from a forbidden state
  without leaking protected data.

## Public Pages

### `/` — Landing and Catalog

Purpose: explain Cartlabs quickly and let visitors discover approved products.

Required functionality:

- Present product purpose and clear catalog entry.
- Search published products by query.
- Filter by category, minimum price, maximum price, and stock availability.
- Paginate results and keep discovery state represented in URL parameters.
- Link every result to its product detail page.
- Link store identity to its public store profile.
- Provide clear zero-result, invalid-filter, loading, and API-failure states.
- Preserve demo entry without making demo tooling the primary product UI.

API dependencies: `GET /categories`, `GET /catalog/products`.

Acceptance:

- Reload and share preserve active search, filters, and page.
- Only published, approved products returned by API are displayed.
- Removing filters restores broader results without losing navigation.

### `/products/[slug]` — Product Detail

Purpose: support product evaluation and variant selection.

Required functionality:

- Show product name, description, images, linked store identity, and category.
- Show active variants with attributes, price, and current stock.
- Require one valid in-stock variant and quantity before adding to cart.
- Add or replace buyer cart quantity.
- Send anonymous purchase intent through login and return to this product.
- Explain seller/admin role restriction instead of offering buyer actions.
- Show verified-review summary and newest reviews.
- Handle missing product, no image, no review, and out-of-stock states.

API dependencies: `GET /catalog/products/{slug}`,
`GET /catalog/products/{slug}/reviews`, `PUT /cart/items/{variantId}`.

Acceptance:

- Selected variant controls displayed price, stock, and submitted cart item.
- Stock conflicts return actionable feedback without losing selection.
- Product remains fully readable without authentication.

### `/stores/[slug]` — Store Profile

Purpose: let visitors understand an approved seller and browse its public
catalog.

Required functionality:

- Show store name, description, seller display name, and creation context.
- List only approved, published products belonging to this store.
- Support search, category, price, stock, and pagination while keeping store
  scope fixed.
- Link every result to product detail.
- Handle unknown, pending, or rejected stores as not found.
- Provide clear no-product, no-filter-result, loading, and API-failure states.

API dependencies: `GET /catalog/stores/{slug}`, `GET /categories`,
`GET /catalog/products?store={slug}`.

Acceptance:

- Store scope cannot be removed accidentally while changing catalog filters.
- Reload and share preserve store catalog filters and page.
- Private seller email, user status, and moderation data never render.
- Public page exposes no pending or rejected store.

## Identity and Shared Pages

### `/login` — Sign In

Purpose: establish a normal product session.

Required functionality:

- Accept email and password.
- Show seeded demo credentials and quick-role actions only when demo mode is
  enabled.
- Show invalid credentials, rate limit, validation, and service-unavailable
  states.
- Redirect successful login to a safe intended destination or actor home.
- Redirect already-authenticated users to their actor home.

API dependencies: `POST /auth/login`, `POST /auth/demo-login`.

Acceptance:

- Password login and enabled demo login create the same frontend session shape.
- Refreshing the destination page preserves the session.
- Repeated submit cannot create competing login requests.

### `/notifications` — Notifications

Purpose: expose durable events relevant to the current actor.

Required functionality:

- List notifications newest first.
- Communicate event type, message, and creation time.
- Provide role-appropriate navigation when notification data contains a usable
  reference.
- Handle empty and unavailable states.
- Do not imply read/unread or delete behavior; current API provides neither.

API dependency: `GET /notifications`.

Acceptance:

- Buyer, seller, and admin see only notifications returned for their identity.
- Page does not fabricate unsupported notification mutations.

## Buyer Pages

### `/cart` — Cart

Purpose: let buyer review live cart state before checkout.

Required functionality:

- Group items by store.
- Show product and variant identity, unit price, quantity, available stock, and
  line/store/cart totals.
- Replace quantity within supported range and remove items.
- Link items back to public product detail.
- Prevent checkout from an empty or currently invalid cart.
- Explain price, stock, missing-item, and conflict responses without silently
  discarding the cart.

API dependencies: `GET /cart`, `PUT /cart/items/{variantId}`,
`DELETE /cart/items/{variantId}`.

Acceptance:

- Successful quantity and removal mutations reconcile UI from returned cart.
- Cart remains grouped by seller and totals match API response.
- Empty cart provides a route back to catalog discovery.

### `/checkout` — Checkout Confirmation

Purpose: confirm current multi-seller cart and create one purchase safely.

Required functionality:

- Re-fetch and present current cart grouped by store.
- Explain one parent purchase with independent seller fulfillment.
- Show final quantities and totals before submission.
- Generate and retain one idempotency key for a submission attempt.
- Prevent duplicate submission while request is pending.
- Handle empty cart, inventory conflict, dependency failure, and retry.
- Redirect successful checkout to created purchase detail.

API dependencies: `GET /cart`, `POST /checkout`.

Acceptance:

- Retrying the same interrupted submission does not create another purchase.
- Successful checkout routes to returned purchase identifier.
- Checkout never claims address, shipping quote, tax, or real payment behavior
  absent from current contracts.

### `/purchases` — Purchase History

Purpose: help buyer find and understand past purchases.

Required functionality:

- List purchases newest first.
- Show reference, creation time, purchase/payment status, total, and seller-order
  summary.
- Link each purchase to detail.
- Handle empty and unavailable states.

API dependency: `GET /purchases`.

Acceptance:

- Status labels distinguish purchase, payment, and seller fulfillment state.
- Empty history offers catalog navigation.

### `/purchases/[purchaseId]` — Purchase Detail

Purpose: complete demo payment, track seller orders, cancel when allowed, and
review delivered items.

Required functionality:

- Show immutable purchase reference, item snapshots, totals, and timestamps.
- Show payment state and reservation expiry.
- In demo mode, complete mock payment with succeeded or failed outcome while
  purchase is eligible.
- Show each seller order independently with its status and timeline.
- Allow full-purchase cancellation only while server rules permit and require a
  reason.
- Offer one review per eligible delivered purchase item with rating, title, and
  body.
- Explain expired, payment-failed, cancelled, and state-conflict outcomes.
- Handle inaccessible or missing purchases without leaking data.

API dependencies: `GET /purchases/{purchaseId}`,
`POST /purchases/{purchaseId}/pay`, `POST /purchases/{purchaseId}/cancel`,
`POST /reviews`.

Acceptance:

- Seller orders never collapse into one misleading fulfillment status.
- Successful actions reconcile from returned server state.
- UI hides or disables impossible transitions but still handles server conflict.

## Seller Pages

### `/seller` — Seller Home

Purpose: summarize seller readiness and route to next required work.

Required functionality:

- Show store existence and moderation state.
- Summarize owned products by lifecycle state.
- Summarize owned orders requiring action.
- Route seller to store creation, rejected moderation fixes, product work, or
  fulfillment as appropriate.
- Handle seller with no store as a first-class onboarding state.

API dependencies: `GET /seller/store`, `GET /seller/products`,
`GET /seller/orders`.

Acceptance:

- Home never assumes seller already owns an approved store.
- Every blocking state exposes a supported next action.

### `/seller/store` — Store Setup and Editing

Purpose: create and maintain seller store identity.

Required functionality:

- Create store with name, slug, and description when none exists.
- Edit existing store fields.
- Show pending, approved, or rejected moderation status and note.
- Warn that material edits return store to pending moderation.
- Explain that product supply requires approved store status.

API dependencies: `GET /seller/store`, `POST /seller/store`,
`PATCH /seller/store`.

Acceptance:

- Missing store renders creation, not generic failure.
- Saved response becomes displayed source of truth.
- Rejected store exposes moderation note beside corrective action.

### `/seller/products` — Product List

Purpose: manage seller catalog supply.

Required functionality:

- List owned products with image, name, category, lifecycle status, moderation
  status, variant count, and stock summary.
- Distinguish draft, submitted/pending, published/approved, rejected, and
  archived states represented by API data.
- Link to product management and draft creation.
- Explain store approval requirement when creation is forbidden.
- Handle no-product and unavailable states.

API dependency: `GET /seller/products`.

Acceptance:

- Moderation note remains discoverable for rejected products.
- Product status is not inferred from styling alone.

### `/seller/products/new` — Create Product

Purpose: create product draft before supply details are added.

Required functionality:

- Collect category, name, slug, and description.
- Validate against current contract constraints.
- Explain approved-store prerequisite.
- Redirect successful creation to product management.

API dependencies: `GET /categories`, `POST /seller/products`.

Acceptance:

- Successful creation routes to returned product identifier.
- Validation and forbidden responses preserve entered values.

### `/seller/products/[productId]` — Product Management

Purpose: prepare, submit, and maintain one owned product.

Required functionality:

- Edit category, name, slug, and description while server state allows.
- Display existing variants and add variant/SKU, attributes, price, currency,
  and initial stock.
- Adjust inventory by non-zero delta with required reason.
- Display existing image metadata and add image URL, alt text, and position.
- Clearly describe URL registration; current API does not upload binary files.
- Show lifecycle, moderation status, and moderation note.
- Submit eligible draft for moderation.
- Explain missing prerequisites and publish conflicts.
- Avoid unsupported variant/image edit or delete controls.

API dependencies: `GET /seller/products/{productId}`,
`PATCH /seller/products/{productId}`,
`POST /seller/products/{productId}/variants`,
`POST /seller/products/{productId}/images`,
`POST /seller/products/{productId}/publish`,
`PATCH /seller/variants/{variantId}/inventory`.

Acceptance:

- Direct navigation loads one complete owned product.
- Missing and non-owned products render the same not-found behavior.
- Returned mutation state replaces optimistic assumptions.
- Unsupported mutations are not represented as functional controls.

### `/seller/orders` — Fulfillment

Purpose: let seller fulfill only orders owned by their store.

Required functionality:

- List order reference, purchase reference, items, totals, timestamps, and
  current status.
- Make orders needing seller action easy to identify.
- Support valid transitions from paid to processing, shipped, and delivered.
- Support cancellation only before shipment and collect a reason.
- Explain pending-payment, cancelled, and immutable states.
- Handle transition conflict after stale data by refreshing server state.

API dependencies: `GET /seller/orders`,
`PATCH /seller/orders/{orderId}/status`.

Acceptance:

- UI offers only transitions valid from known state.
- Server remains authority when state changed concurrently.
- No control suggests access to another seller's orders.

## Admin Pages

### `/admin` — Marketplace Overview

Purpose: summarize marketplace operating state.

Required functionality:

- Show users, approved stores, published products, purchases, active orders,
  delivered orders, and gross merchandise value returned by API.
- Link relevant summaries to supported moderation pages.
- Show data timestamp context and unavailable state without inventing trends.

API dependency: `GET /admin/overview`.

Acceptance:

- Values and currency formatting reflect API response.
- No chart implies historical data because API exposes current totals only.

### `/admin/stores` — Store Moderation

Purpose: approve or reject marketplace stores.

Required functionality:

- List stores with identity, description, status, moderation note, and dates.
- Make pending items distinguishable and reviewable.
- Approve or reject with moderation note.
- Confirm consequential moderation action and reconcile returned state.
- Support empty and unavailable states.

API dependencies: `GET /admin/stores`,
`PATCH /admin/stores/{storeId}/moderation`.

Acceptance:

- Pending work can be found without hiding previously moderated records.
- Repeated or stale moderation handles server state safely.

### `/admin/products` — Product Moderation

Purpose: approve or reject submitted products.

Required functionality:

- List product identity, store, category, description, variants, images,
  lifecycle state, moderation state, and note.
- Make pending products distinguishable and fully inspectable.
- Approve or reject with moderation note.
- Confirm consequential moderation action and reconcile returned state.
- Support empty and unavailable states.

API dependencies: `GET /admin/products`,
`PATCH /admin/products/{productId}/moderation`.

Acceptance:

- Admin can inspect supply details required for a moderation decision.
- Product does not appear approved in UI before server confirmation.

### `/admin/users` — User Status Management

Purpose: inspect marketplace accounts and suspend or reactivate access safely.

Required functionality:

- List user email, display name, role, status, and account dates.
- Support client-side search and filtering over returned users.
- Distinguish active and suspended status without relying on color alone.
- Suspend or reactivate with a required reason.
- Confirm status mutation and reconcile from returned server state.
- Explain that suspension invalidates current access and refresh sessions.
- Prevent self-suspension and surface last-active-admin protection.
- Handle redundant transition, stale state, missing user, and service failure.

API dependencies: `GET /admin/users`,
`PATCH /admin/users/{userId}/status`.

Acceptance:

- Non-admin actors cannot access user data or mutation controls.
- Admin cannot suspend own account or remove last active admin.
- Suspended account loses authenticated access immediately.
- Reactivated account must establish a new login session.
- Status change becomes visible in audit history with its reason.

### `/admin/audit` — Audit History

Purpose: inspect newest trust and fulfillment actions.

Required functionality:

- List actor, action, target, metadata, and timestamp available from API.
- Distinguish moderation and fulfillment events.
- Handle empty and unavailable states.
- Do not imply unsupported search, filtering, export, or pagination.

API dependency: `GET /admin/audit-events`.

Acceptance:

- Events remain readable when optional metadata is absent or unfamiliar.
- UI does not reinterpret audit history as mutable data.

## Demo Page

### `/demo` — Demo Guide and Role Entry

Purpose: help visitors exercise real product workflows quickly.

Required functionality:

- Explain seller to admin to buyer journey order.
- Show seeded account identities and visible development credentials.
- Show reset schedule and warn that demo data is temporary.
- Offer quick login for buyer, seller, and admin only when demo mode is enabled.
- Redirect each role into its normal product home.
- Replace and remove `DemoConsole`; no business mutation happens on this page.

API dependency: `POST /auth/demo-login`.

Acceptance:

- Each shortcut establishes a normal durable session.
- Disabled demo mode exposes no working quick-login action.
- All demonstrated business actions occur on normal actor pages.

## Explicitly Unsupported Browser Features

Do not design functional controls for these without accepted API changes:

- User registration, password reset, profile editing, or account deletion.
- Public store directory independent from product/store links.
- Binary product image upload, image edit, or image deletion.
- Variant edit, activation toggle, or deletion.
- Product deletion.
- Notification read/unread state or deletion.
- Shipping address, shipping quote/provider, tax calculation, or real payment.
- Audit search, filtering, export, or pagination.
- Wishlist, promotion, chat, refund, dispute, or recommendation behavior.

## Backend Assumptions

- Admin user listing and safe status management are implemented.
- Approved public store lookup, public `storeSlug` product fields, and catalog
  filtering by store are implemented.
- Owned seller product detail lookup is implemented.
- **Demo payment controls:** show them only when demo mode is enabled; otherwise
  render payment state without a completion action.

## Page Approval Checklist

Before frontend implementation begins:

- [x] Route inventory approved.
- [x] Required features approved for every route.
- [x] Unsupported features accepted as out of scope.
- [x] Scope decisions resolved or explicitly deferred.
- [x] Page boundaries may change visually without losing required behavior.
- [x] Backend assumptions verified against implemented OpenAPI contracts.
