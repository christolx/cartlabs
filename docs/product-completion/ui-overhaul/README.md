# UI Overhaul Plan

Execution plan for roadmap phases 6 and 7.

## Order

1. Lock tokens and shared primitives in [`design-system.md`](./design-system.md).
2. Build shared shell and states using [`implementation.md`](./implementation.md).
3. Redesign every page through mandatory generated-image-to-implementation workflow in [`image-to-code-pipeline.md`](./image-to-code-pipeline.md), with product-owner intervention and approval at each gate.
4. Sequence page families using [`page-families.md`](./page-families.md).
5. Review each page immediately, then run its family checkpoint using [`visual-review.md`](./visual-review.md).
6. Run cross-system drift audit and final regression using [`regression.md`](./regression.md).

## Constraints

- Preserve routes, API contracts, domain behavior, form fields, and role boundaries.
- Present Cartlabs as real marketplace first. Keep demo framing within demo utilities.
- Replace current green visual language completely.
- Use one coherent system across buyer, seller, and admin surfaces.
- Treat public commerce as image-led. Treat seller and admin pages as operational UI.
- Do not implement a page redesign until its generated reference image receives explicit product-owner approval.
- Do not start next page until product owner approves current implementation.
