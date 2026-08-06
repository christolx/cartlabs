# Milestones

## Milestone 1 — Identity and Catalog

**Status:** Complete (2026-08-08)

### Work

- [x] Authentication, refresh rotation, RBAC, auth rate limiting, and demo login
- [x] Seller store management and admin moderation
- [x] Products, variants, categories, images, and atomic inventory ledger
- [x] Public catalog, search, filters, product pages, and role demo workflows

### Exit evidence

Seller creates a store and receives admin approval. Seller creates a product,
SKU, image, and stock, then publishes it. Product remains hidden until admin
approval. Buyer discovers approved product through search and opens its detail
page. Buyer cannot access admin APIs.

Verified through:

- Unit tests for Argon2id, JWT validation, refresh replay revocation, RBAC,
  validation, strict JSON decoding, cookies, and auth rate limiting
- Compose-backed Hurl workflow with 18 API requests
- Browser workflow across seller, admin, and buyer roles
- One shared, attributed stock product image used by all demo products
- Desktop/mobile and light/dark visual checks
- Axe checks with zero violations on catalog, product, and demo pages
- `make check`

## Milestone 2 — Purchase Vertical Slice

**Status:** Next

### Work

- [ ] Multi-seller cart
- [ ] Checkout totals and inventory reservation
- [ ] Parent purchase and seller orders
- [ ] Mock payment intent, checkout UI, and signed webhook
- [ ] Transactional outbox, RabbitMQ worker, and notifications

### Exit

Buyer completes a purchase spanning two sellers without overselling.
