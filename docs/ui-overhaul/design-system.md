# Cartlabs UI System

## Direction

Editorial brutalist commerce: strict black keylines, square geometry, condensed display type, mono data type, warm product photography, sparse signal red.

## Tokens

```css
--black: #050505;
--white: #fff;
--red: #e60000;
--warm-image: #e8e0d4;
--border: 1px solid var(--black);
--font-display: "Anton", Impact, sans-serif;
--font-ui: "Roboto Mono", monospace;
--space: 4px 8px 12px 16px 24px 32px 48px 64px;
```

No radii, shadows, glass, decorative gradients, or extra accent colors.

## Type

- Wordmark, hero, product titles: Anton, uppercase.
- Navigation, controls, metadata, prices: Roboto Mono.
- Hero: 84px desktop, `.88` line-height.
- Product title: 16px. UI: 12px. Metadata: 10–11px.

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

- Red reserved for primary action, cart, and stock state.
- Hover: black fill/white text or restrained image zoom.
- Focus: visible 3px outline. Reduced-motion preference disables transitions.
