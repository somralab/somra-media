# Somra Brand Guide (M3 Alpha)

Somra is a self-hosted media server with an original visual identity — not a clone of
third-party streaming services.

## Logo & wordmark

- Wordmark: **Somra** in `--font-sans`, semibold, sentence case.
- No third-party logos or trade dress are used in themes or marketing assets.

## Color themes

Four user-selectable themes ship in M3:

| Theme | Character | Primary accent |
|-------|-----------|----------------|
| **Cinematic** (default) | Dark, warm violet/pink accents | `#7C3AED` / `#F472B6` |
| **Aurora** | Cool teal/sky on deep blue | `#38BDF8` / `#34D399` |
| **Noir** | High-contrast monochrome | `#F0F0F0` on near-black |
| **Minimal** | Light neutral UI | `#4F46E5` on white |

Tokens live in `web/src/styles/tokens.css` and map to Tailwind via CSS variables.

## Typography

- UI: system sans stack (`--font-sans`).
- Monospace: paths, diagnostics (`--font-mono`).

## Components

Browse UI building blocks under `web/src/components/browse/`:

- **PosterCard** — 2:3 poster, progress bar, link to detail.
- **MediaRow** — horizontal shelf with scroll.
- **MediaGrid** — virtualized grid/list for large libraries.
- **Skeleton / EmptyState / ErrorState** — consistent loading and edge states.

Search: **SearchBar**, **SearchResultsDropdown** in `web/src/components/search/`.

## Accessibility

- WCAG 2.1 AA contrast targets for text on `--color-bg` and `--color-surface`.
- Focus rings on interactive cards and nav items.
- See `docs/design/a11y-checklist.md` for the M3 audit checklist.
