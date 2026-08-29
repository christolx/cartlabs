# About and Docs Pages

## Goal

Present Cartlabs as realistic marketplace engineering work and explain how to
use each product role.

## About (`/about`)

- Lead with: one checkout, many independent sellers.
- Show checkout lifecycle with responsive semantic HTML/CSS.
- Explain implemented decisions: order splitting, inventory reservation,
  idempotent payment webhooks, async work, role isolation, and search extraction.
- Show platform architecture with simple HTML/CSS or inline SVG.
- Summarize reliability, security, stack, and deployment.
- State current boundaries: mock payments, IDR, and demo mode.
- Link deeper repository documentation.

## Docs (`/docs`)

Single page with anchored sections:

- Getting started
- Buyer workflow
- Seller workflow
- Admin workflow
- Order and payment statuses
- Demo behavior and limitations

## Navigation

- Guest header: `Browse`, `About`, `Docs`.
- Keep `Browse` linked to `/#catalog`; remove duplicate `Stores` link.
- Keep authenticated role navigation unchanged.
- Add About, Docs, and repository links to footer.

## Constraints

- Describe implemented behavior only.
- Use direct, realistic case-study voice.
- Match existing editorial design.
- Avoid diagram libraries, animation, and unnecessary interaction.
- Keep pages responsive, accessible, and server-rendered where possible.

## Verification

- Run web typecheck, lint, and tests.
- Check guest and authenticated navigation.
- Check keyboard use and mobile layouts.
