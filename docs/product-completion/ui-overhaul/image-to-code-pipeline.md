# Per-Page Image-to-Code Pipeline

Every page redesign must move from a generated reference image to implementation. The product owner directs each page manually and owns every approval gate. Existing UI, written direction, or developer judgment cannot replace this pipeline.

## Required Sequence

1. **Scope page.** Inventory route, purpose, primary action, content, reachable states, role rules, mutations, and required mobile behavior. Preserve these constraints in the redesign brief.
2. **Take manual direction.** Product owner supplies or revises composition, hierarchy, mood, density, imagery, and interaction priorities. Record decisions and unresolved questions.
3. **Generate reference.** Produce a page-specific design image from approved direction. Generate supplemental images when mobile layout or a materially different state cannot be inferred faithfully from the primary image.
4. **Approve reference.** Product owner reviews, annotates, and either rejects or approves generated image. Revise image until explicit approval. Do not write page implementation before this gate passes.
5. **Analyze image.** Treat approved image as primary visual source. Extract visible text, layout, typography, spacing, color, imagery, controls, component logic, responsive intent, and ambiguity before coding.
6. **Implement page.** Translate approved image into responsive frontend while preserving route behavior, API contracts, accessibility, states, and shared design-system rules. Do not replace distinctive image decisions with generic UI patterns.
7. **Compare implementation.** Capture implementation at required widths and modes. Compare it with approved reference, list fidelity gaps, and fix material differences.
8. **Approve page.** Product owner manually reviews implementation and reachable states. Iterate from owner feedback until explicit page approval. Only then start next page.

Sequence is strict: **manual direction -> generated image -> manual image approval -> image analysis -> implementation -> manual implementation approval**.

## Manual Intervention Gates

Product owner must intervene at three points for every page:

- **Brief gate:** confirm page-specific direction before image generation.
- **Reference gate:** approve generated image before implementation.
- **Implementation gate:** approve coded result after visual and behavioral review.

Silence, prior family approval, automated checks, or developer self-review do not count as approval. Shared-component changes affecting an approved page reopen its implementation gate.

## Reference Requirements

- One approved primary reference image per page minimum.
- Reference must show enough detail to implement hierarchy, spacing, typography, color, imagery, and major controls.
- Supplemental reference required for materially different responsive compositions or states that cannot be inferred safely.
- All supplied and generated references must be analyzed. Conflicts resolve through product-owner direction.
- Approved reference remains visual source of truth. Functional requirements and accessibility override only when exact copying would break behavior or compliance; record deviation for review.

## Tracking Record

Record this evidence for each page:

- route and page name
- scoped states and functional constraints
- latest manual direction
- approved reference image path or artifact link
- reference approval and date
- extracted design notes and declared ambiguities
- implementation screenshot paths for required widths and modes
- fidelity gaps and resolutions
- implementation approval and date
- reopened approval caused by later shared changes

Missing evidence means page is not complete.
