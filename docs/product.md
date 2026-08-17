# Product

## Goal

Demonstrate a credible multi-vendor marketplace through a deployable personal
project that friends and recruiters can explore.

## Actors

- **Buyer:** browse, search, manage cart, checkout, track orders, review products
- **Seller:** manage store, products, variants, inventory, and fulfillment
- **Admin:** verify stores, enforce listings, manage users, and inspect platform activity

Three deterministic demo accounts provide quick login for each role.

## Core Marketplace Flow

1. Seller creates store, products, variants, prices, and stock.
2. Buyer browses products and builds a multi-seller cart.
3. Checkout creates one parent purchase and one child order per seller.
4. Mock payment provider completes, fails, or cancels payment.
5. Signed, idempotent webhook updates payment and order state.
6. Sellers fulfill only their child orders.
7. Buyer tracks each seller shipment independently.

## MVP Scope

### Catalog

- Stores and seller profiles
- Simple products with size/color variants
- SKU-level price and inventory
- Categories, search, filters, and pagination
- Product images

### Commerce

- Persistent buyer cart
- Multi-seller checkout
- Parent purchase and per-seller orders
- Inventory reservation and release
- Mock payment with webhook flow
- Per-seller shipping status
- Basic order cancellation

### Administration

- Store verification and reactive product listing enforcement
- User status management
- Marketplace overview
- Audit trail for privileged actions

## Deferred

- Real payment and shipping providers
- Email delivery and email notification workflows
- Promotions, vouchers, flash sales, and loyalty
- Chat, disputes, and complex refund workflows
- Product bundles, digital goods, subscriptions, and auctions
- Recommendations and personalization
- Multiple currencies, languages, warehouses, or tax jurisdictions
- Native mobile applications

## Demo Policy

- Seed buyer, seller, and admin accounts with visible credentials.
- Quick-login buttons work only when explicit demo mode is enabled.
- Reset demo environment daily from migrations plus deterministic seed data.
- Never layer seed data over mutated records.
- Display reset schedule and warn that demo data is temporary.

## Success Criteria

- Fresh checkout reaches usable state through one documented command.
- Buyer can complete full multi-seller purchase flow.
- Seller and admin permissions remain isolated and test-covered.
- Compose and k3s deployments use equivalent application contracts.
- CI validates contracts, code, images, migrations, and deployment manifests.
