# Page Families

Families define shared constraints, not shared approval. Run mandatory [`image-to-code pipeline`](./image-to-code-pipeline.md) for each listed page. Dynamic routes need one page pipeline plus supplemental references for materially different layouts or states.

## Public Commerce

Routes: `/`, `/stores/[slug]`, `/products/[slug]`.

- Image-led composition and strongest visual variance.
- Catalog remains primary homepage job.
- Filters stay efficient and collapse explicitly on mobile.
- Product detail keeps media dominant and purchase controls obvious.
- Store page establishes seller identity without exposing private data.

## Access and Demo

Routes: `/login`, `/demo`.

- Login prioritizes account access.
- Demo utilities remain explicit but visually secondary.
- Quick-login preserves current role behavior.
- Demo page explains journey without making entire product look synthetic.

## Buyer Workspace

Routes: `/cart`, `/checkout`, `/purchases`, `/purchases/[purchaseId]`, `/notifications`.

- Maintain seller grouping throughout cart and fulfillment.
- Sticky summaries remain only where viewport permits.
- Payment, purchase, and seller-order statuses stay visually distinct.
- Destructive and irreversible actions require clear context.

## Seller Workspace

Routes: `/seller`, `/seller/store`, `/seller/products`, `/seller/products/new`, `/seller/products/[productId]`, `/seller/orders`.

- Readiness and next action lead seller home.
- Forms optimize scanning and error recovery.
- Product management separates identity, variants, media, inventory, and lifecycle.
- Orders needing action receive strongest hierarchy.

## Admin Workspace

Routes: `/admin`, `/admin/stores`, `/admin/products`, `/admin/users`, `/admin/audit`.

- Dense, trust-first layout with minimal motion.
- Moderation state and consequences remain visible before action.
- User management preserves table semantics and mobile restructuring.
- Audit history favors chronology, metadata legibility, and immutable context.

## Shared States

Apply same visual grammar to loading, empty, error, not-found, forbidden, and success states. Skeletons match final layout. Errors explain recovery. Empty states name next valid action.
