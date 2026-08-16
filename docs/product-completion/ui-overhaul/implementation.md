# Phase 6 Implementation

## Mandatory Page Workflow

Apply [`image-to-code-pipeline.md`](./image-to-code-pipeline.md) independently to every page listed below. Each page requires product-owner direction, approved generated reference, pre-code image analysis, faithful implementation, screenshot comparison, and final product-owner approval. Family-level direction or approval cannot substitute for page-level gates.

## Pass 1: System

- Replace current green tokens with precision-retail tokens.
- Establish typography, spacing, radius, elevation, and responsive rules.
- Refactor shared buttons, fields, status, notices, headings, and states.
- Rebuild header and footer for every session role.

## Pass 2: Public Commerce

- Redesign homepage, store, product, login, and demo pages.
- Run full image-to-code pipeline for each page before starting next page.
- Run public-commerce family checkpoint after all five pages pass.
- Create required studio-first image assets.
- Verify hero, catalog, filters, product media, and purchase actions at mobile and desktop widths.

## Pass 3: Buyer

- Redesign cart, checkout, purchases, purchase detail, and notifications.
- Run full image-to-code pipeline for each page before starting next page.
- Run buyer-family checkpoint after all five pages pass.
- Preserve multi-seller grouping and all mutation behavior.

## Pass 4: Seller

- Redesign seller home, store setup, products, product management, inventory, and orders.
- Run full image-to-code pipeline for each page before starting next page.
- Run seller-family checkpoint after all six pages pass.
- Verify no-store, rejected-store, draft, suspended, archived, and fulfillment states.

## Pass 5: Admin

- Redesign overview, store moderation, listing enforcement, users, and audit history.
- Run full image-to-code pipeline for each page before starting next page.
- Run admin-family checkpoint after all five pages pass.
- Verify confirmation context, danger hierarchy, tables, and metadata.

## Pass 6: Polish

- Audit responsive collapse per page family.
- Audit visible copy, contrast, focus, touch targets, and reduced motion.
- Run cross-system drift audit across previously approved pages.
- Remove obsolete CSS and unused assets only after replacements are verified.
- Run typecheck, lint, unit tests, production build, and browser smoke tests.

## Definition of Ready for Manual Review

- Every route uses new system.
- Every page has pipeline tracking record with approved reference image, analysis, implementation screenshots, and recorded product-owner approvals.
- Every page family passed consistency review.
- No legacy green token or component styling remains.
- All shared states use new system.
- Both color modes reviewed.
- Mobile and desktop screenshots captured for each page family.
