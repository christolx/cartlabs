# Page Families

Overhaul by shared user journey and reusable UI, not route order.

## 1. Storefront

Routes: `/`, `/products/[slug]`, `/stores/[slug]`

Purpose: discovery, product evaluation, seller trust. Establish product media, cards, pricing, variants, stock, reviews, filters, and pagination.

Order: homepage complete → product detail → store page.

## 2. Commerce

Routes: `/cart`, `/checkout`, `/purchases`, `/purchases/[purchaseId]`

Purpose: cart review, checkout confirmation, payment state, fulfillment tracking, and reviews. Reuse line items, totals, status, actions, and order summaries.

Order: cart → checkout → purchase detail → purchase list.

## 3. Account

Routes: `/login`, `/notifications`

Purpose: account access and event history. Keep forms, messages, empty states, and feeds consistent with commerce pages.

Order: login → notifications.

## 4. Seller Workspace

Routes: `/seller`, `/seller/store`, `/seller/products`, `/seller/products/new`, `/seller/products/[productId]`, `/seller/orders`

Purpose: store readiness, catalog management, inventory, publishing, and fulfillment. Establish workspace navigation, metrics, forms, records, and operational states.

Order: dashboard → products → product editor → orders → store settings.

## 5. Admin Workspace

Routes: `/admin`, `/admin/products`, `/admin/stores`, `/admin/users`, `/admin/audit`

Purpose: marketplace oversight, enforcement, verification, access control, and audit history. Reuse seller workspace shell while increasing data density.

Order: overview → products and stores → users → audit.

## 6. System States

Routes: `error.tsx`, `not-found.tsx`, product/store not-found, product loading.

Purpose: loading, empty, unavailable, invalid, unauthorized, and missing states. Finish after core primitives stabilize; verify during every family.

## Delivery Sequence

Storefront → Commerce → Account → Seller → Admin → final system-state pass.

Each family ships responsive, keyboard-accessible, and complete across loading, empty, error, and success states.
