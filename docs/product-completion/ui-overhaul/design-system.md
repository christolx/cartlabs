# Design System

## Foundation

- Next.js Server Components by default.
- Tailwind CSS v4 plus semantic CSS variables.
- `next/font` for typography.
- Native interaction first. Add Motion only when existing dependencies and scope justify it.
- One icon family if icons become necessary. Prefer Phosphor.

## Tokens

Define semantic tokens before components:

- Surfaces: canvas, subtle, raised, overlay.
- Text: primary, secondary, disabled, inverse.
- Borders: subtle, default, strong.
- Brand: cobalt, cobalt-hover, cobalt-contrast.
- Feedback: success, warning, danger, info.
- Focus: high-contrast cobalt ring.
- Shadow: cool-tinted, low-opacity elevation.

Support light and dark modes through one token set. Respect system preference.

## Typography

- Sans-only system.
- Display: tight tracking, controlled scale, two-line maximum for public heroes.
- Body: readable measure, 16px minimum default.
- Mono: identifiers, SKUs, timestamps, and audit metadata only.
- Avoid uppercase micro-label repetition.

## Shape

- Cards and panels: 12px radius.
- Inputs: 8px radius.
- Buttons: 8px radius.
- Status badges: full pill only when status grouping benefits.
- Product media: 12px radius.

## Components

Build shared primitives before page redesign:

- Button and text action
- Field, select, textarea, checkbox
- Status badge and notice
- Product card and media fallback
- Page header and section header
- Data row, record card, and metric
- Dialog or disclosure action
- Empty, loading, error, and forbidden states
- Responsive navigation and footer

Every interactive component needs hover, active, focus, disabled, loading, success, and error behavior where applicable.

## Motion

- Animate transform and opacity only.
- Use motion for hierarchy, feedback, or state transition.
- Public pages: short entry and image-hover motion.
- Workspaces: state transitions only.
- Honor `prefers-reduced-motion`.
- No scroll hijacking, perpetual decoration, or custom cursor.

## Accessibility

- WCAG AA minimum contrast; target AAA for body copy.
- Labels stay above fields. Placeholder never replaces label.
- Visible keyboard focus on every control.
- Minimum 44px touch targets.
- Preserve semantic headings, landmarks, table roles, and live regions.

