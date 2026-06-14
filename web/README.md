# Somra — Web (Sprint 01)

React + Vite + TypeScript (strict) SPA. Sprint 01 M1 ships the full
shell: typed API client (`src/api/`), TanStack Query hooks for
`/api/v1/health` and `/api/v1/version`, an SSE event panel and a
Playwright smoke spec (`e2e/status.spec.ts`). Paket 3 produced the
scaffold; Paket 6 wired it up; Paket 8 closed the milestone with the
e2e harness.

## Quick start

```bash
pnpm install
pnpm dev           # http://127.0.0.1:5173
```

## Scripts

| Command              | What it does                                                |
| -------------------- | ----------------------------------------------------------- |
| `pnpm dev`           | Vite dev server with HMR.                                   |
| `pnpm build`         | Type-check + production bundle into `dist/`.                |
| `pnpm preview`       | Serve the built bundle locally.                             |
| `pnpm lint`          | ESLint (flat config, strict TS + React + a11y + Prettier).  |
| `pnpm lint:fix`      | Auto-fix lint issues.                                       |
| `pnpm format`        | Prettier write.                                             |
| `pnpm format:check`  | Prettier check (CI).                                        |
| `pnpm typecheck`     | `tsc -b --noEmit` across project references.                |
| `pnpm test`          | Vitest (jsdom) — i18n + theme suites.                       |
| `pnpm test:watch`    | Vitest watch mode.                                          |
| `pnpm test:coverage` | Vitest with v8 coverage (HTML + JSON summary).              |
| `pnpm i18n:parity`   | Parity check between `en-US` and `tr-TR` locale resources.  |
| `pnpm e2e:install`   | Install Playwright Chromium (one-time, required for `e2e`). |
| `pnpm e2e`           | Playwright status-page smoke against the Go binary.         |

The repo-wide `bash scripts/i18n-check.sh` runs a Go-backed parity check
that covers both the frontend JSON namespaces here and the backend
`active.*.toml` bundles in `internal/platform/i18n/locales/`. CI uses the
shell wrapper; this one only covers `web/`.

## Architecture overview

### Routing & providers

`src/main.tsx` mounts `<App/>` inside `StrictMode`, wrapped by:

```
I18nextProvider → QueryClientProvider → ThemeProvider → ToastProvider → BrowserRouter
```

`<App/>` exposes two routes:

| Path        | Purpose                                                     |
| ----------- | ----------------------------------------------------------- |
| `/`         | Status page (placeholder copy; real data lands in Paket 6). |
| `/settings` | Language + theme switchers.                                 |

### Theming

- 4 themes: `cinematic` (default), `aurora`, `noir`, `minimal`.
- Each theme is **only** a block of CSS custom properties in
  [`src/styles/tokens.css`](./src/styles/tokens.css). Adding a new theme requires no
  code change — declare another `:root[data-theme="..."]` block.
- The active theme is mirrored on `<html data-theme="…">` and persisted under the
  `somra.theme` localStorage key (via Zustand `persist`).
- Tailwind tokens (`bg-bg`, `text-text`, `bg-primary`, etc.) resolve from those CSS
  variables, so the SPA themes itself instantly without reload.

### i18n

- Library: `i18next` + `react-i18next` + `i18next-browser-languagedetector`.
- Source/fallback locale: **`en-US`**; first translation: **`tr-TR`**.
- Resources are loaded statically from JSON via [`src/i18n/index.ts`](./src/i18n/index.ts).
- Namespaces: `common`, `status`. Key style: `domain.context.key`.
- Locale detection order: `localStorage('somra.locale')` → browser `navigator`.
- No user-facing string is hardcoded; every label is resolved with `t(...)`. The
  `<title>` is also updated on `languageChanged` to keep `document.title` localized.
- Parity is enforced by [`scripts/i18n-parity.mjs`](./scripts/i18n-parity.mjs), exposed
  as `pnpm i18n:parity`.

### Styling

- **Tailwind CSS v3** + PostCSS + Autoprefixer.
  > Decision note: Tailwind v4 (`@tailwindcss/vite`) was evaluated. Sprint 01 sticks
  > with the proven Tailwind v3 + PostCSS pipeline because it interacts cleanly with
  > Vite test mode (`vitest` + `jsdom`) and ESLint/Prettier tooling without churn.
  > Upgrading to v4 is a single-PR follow-up once its plugin ecosystem (notably
  > `prettier-plugin-tailwindcss`) is fully stable for this stack.
- Class merging via `clsx` + `tailwind-merge` (see [`src/lib/cn.ts`](./src/lib/cn.ts)).

### State

- **TanStack Query** for server state (client created in `src/lib/queryClient.ts`).
- **Zustand** for UI state and theme persistence (see `src/stores/ui.ts`,
  `src/theme/ThemeProvider.tsx`).

### Tests

- Vitest + Testing Library + jsdom.
- Suites:
  - [`src/test/i18n.test.tsx`](./src/test/i18n.test.tsx) — `en-US ↔ tr-TR` switch
    re-renders the resolved copy.
  - [`src/test/theme.test.tsx`](./src/test/theme.test.tsx) — theme switch mutates
    `<html data-theme>` and writes through to `localStorage`.

## Sprint 01 — what shipped

- Typed API client + TanStack Query hooks for health / version (Paket 6).
- OpenAPI-generated client types under `src/api/generated/`
  (`pnpm gen:api`, Paket 2).
- SSE event panel wired into the status dashboard (Paket 6).
- Playwright smoke spec — `e2e/status.spec.ts` (this packet).
