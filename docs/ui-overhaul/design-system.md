# Cartlabs UI System

## Direction

Editorial brutalist commerce: strict black keylines, square geometry, condensed display type, mono data type, warm product photography, sparse signal red.

## Contract

This document is source of truth for visual implementation. Mockups resolve composition questions; tokens and component rules resolve implementation questions. Components consume semantic tokens, never raw palette values. Route-specific CSS may compose primitives but must not redefine their visual states.

## Tokens

```css
/* Raw palette: tokens.css only. */
--color-black: #050505;
--color-white: #fff;
--color-red: #e60000;
--color-warm: #e8e0d4;

/* Public semantic interface. */
--color-canvas: var(--color-white);
--color-ink: var(--color-black);
--color-border: var(--color-black);
--color-action: var(--color-red);
--color-image: var(--color-warm);
--keyline: 1px solid var(--color-border);
--control-height: 44px;
```

Full contract lives in `apps/web/src/app/styles/tokens.css`. Spacing uses a 4px grid: 4, 8, 12, 16, 24, 32, 48, 64. No radii, decorative shadows, glass, gradients, automatic dark theme, or extra accent colors. Functional success, warning, and danger colors are state signals, not decorative accents.

## Type

- Wordmark, hero, product titles: Anton, uppercase.
- Navigation, controls, metadata, prices: Roboto Mono.
- Hero: 84px desktop, `.88` line-height.
- Product title: 16px. UI: 12px. Metadata: 10–11px.
- Page title: responsive 42–72px. Section title: responsive 32–50px.
- Body copy uses UI mono at 14px; long editorial descriptions may use 16–18px.
- Prices, quantities, dates, identifiers, status, and tabular data use mono data role.

## Layout and Density

- Maximum content width: 1400px. Desktop gutter: 24px; mobile gutter: 16px.
- Public pages use editorial scale, generous image fields, and sparse copy.
- Commerce pages use medium density: readable line items plus persistent summaries where space permits.
- Seller and admin workspaces use compact controls, keylined records, and denser data without abandoning type or color rules.
- Breakpoints: desktop above 1023px, tablet 768–1023px, mobile below 768px.
- Sticky content becomes static when its column collapses.

## Primitives

- Buttons: 44px minimum target, square, mono uppercase label. Primary uses action red; secondary uses white with black keyline; destructive uses danger state and must say destructive verb.
- Fields: 44px minimum height, square black keyline, persistent visible label except familiar search controls with accessible names. Help and error text sits directly after control.
- Panels and cards: white or neutral surface, black keyline, no radius. Product images may use warm image surface.
- Tables and records: aligned numeric data, clear row keylines, header labels at metadata size. Collapse into labeled records on mobile when horizontal scrolling would hide actions.
- Status: text plus color; color never carries meaning alone. Success/stock, warning/pending, danger/failure, and neutral states use separate semantic tokens.
- Notices: left keyline plus heading and message. Error and warning notices receive matching state color.

## States and Accessibility

- Hover: black fill with white text, action-red fill change, or restrained image zoom. Never move layout.
- Focus: visible 3px action-colored outline with 3px offset on every interactive control, including dark and red surfaces.
- Disabled: preserve readable label, use disabled token, remove hover/active motion, and retain `disabled` or `aria-disabled` semantics.
- Loading: preserve final layout geometry. Reduced-motion disables pulsing and transitions.
- Empty and error states: name state, explain next step, and provide action when recovery exists.
- Success: announce async completion through existing live-region/form-message patterns; never rely on color alone.
- Text contrast targets WCAG AA. Interactive targets are at least 44×44px on touch layouts.

## Desktop Composition

- Header: 58px; black wordmark cell, white navigation, red cart cell.
- Hero: 307px; 66px vertical rail, white headline field, diagonal red collage field, 69px black index rail.
- Filters: one 67px band, six horizontal controls.
- Catalog: four equal columns, 14px gutters, two visible rows.
- Cards: 1px border, wide editorial image, black number badge, compact 78px data panel.
- Footer: five bordered cells plus legal strip; black brand cell and terminal red block.

## Responsive

- Tablet: two product columns; filters wrap; hero keeps split composition.
- Mobile: two product columns where viable; hero stacks headline over collage; navigation becomes second header row.

## Interaction

- Action red reserved for primary action and cart. Stock/success uses success token; destructive and error states use danger token.
- Hover: black fill/white text or restrained image zoom.
- Focus: visible 3px outline. Reduced-motion preference disables transitions.
