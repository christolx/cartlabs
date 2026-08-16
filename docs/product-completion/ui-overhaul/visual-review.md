# Manual Visual Review

## Cadence

Review each page throughout its image-to-code pipeline. Do not wait for implementation or full overhaul.

1. Product owner directs and approves page brief.
2. Generate page reference image; product owner reviews and approves it before code begins.
3. Analyze approved image, then implement one page and its reachable states.
4. Compare implementation screenshots against approved reference at required widths and color modes.
5. Product owner reviews implementation. Fix findings and obtain explicit approval before starting next page.
6. After every page in a family passes, run family consistency review.
7. After all families pass, run final cross-system audit. This audit checks drift; it is not first review.

Record brief approval, reference artifact and approval, image analysis, screenshots, fidelity findings, and implementation approval in tracking defined by [`image-to-code-pipeline.md`](./image-to-code-pipeline.md).

## Review Widths

- Mobile: 390px
- Tablet: 768px
- Desktop: 1440px

Review light and dark modes. Use populated, empty, loading, error, and restricted states where available.

## Per-Page Gate

- Product owner approved page-specific brief.
- Generated reference image exists and has explicit product-owner approval.
- Image analysis records layout, typography, spacing, color, imagery, controls, responsive intent, and ambiguity.
- Implementation remains faithful to approved reference or records approved deviations.
- Page purpose and primary action remain obvious.
- New art direction appears without breaking behavior.
- Default and reachable alternate states use same visual system.
- Mobile, tablet, and desktop layouts pass.
- Light and dark modes pass.
- Keyboard, focus, contrast, touch targets, and reduced motion pass.
- Visible copy and image treatment pass.
- Product owner approves coded page before next page begins.

## Family Checkpoint

- Shared navigation and task patterns remain consistent.
- Similar actions use same labels, hierarchy, and feedback.
- Page density and spacing feel related without forcing identical layouts.
- No earlier approved page drifted during shared-component changes.

## System Checks

- Cobalt is sole brand accent.
- Cool-neutral palette remains consistent.
- Radius rules remain consistent.
- Typography hierarchy stays coherent across roles.
- Navigation fits one line at desktop and remains usable on mobile.
- Buttons never wrap and maintain contrast.
- Focus states remain visible.

## Public Checks

- Product imagery leads without obscuring commerce controls.
- Homepage value and catalog appear within initial viewport context.
- Filters scan quickly and collapse cleanly.
- Product purchase action remains obvious.
- Store identity feels credible, not decorative.

## Workspace Checks

- Current status and next action are obvious.
- Dense content remains scannable without excessive cards.
- Forms preserve labels, helper text, errors, and disabled states.
- Dangerous actions remain visually distinct.
- Tables and records reflow without horizontal loss.

## Anti-Slop Checks

- No purple glow, green legacy theme, three-equal feature-card marketing row, fake screenshots, decorative status dots, scroll cues, version labels, or generic stock scenes.
- No repeated eyebrow pattern.
- No duplicate CTA intent.
- No animation without hierarchy, feedback, or state purpose.
- No em dash or en dash used as visible punctuation.

Phase 6 approval requires complete pipeline evidence and product-owner approval for every page, every family checkpoint passed, and final cross-system drift audit passed.
