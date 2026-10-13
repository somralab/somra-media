# Sprint 12 — Streaming-App Experience (Netflix-like UI)

> **Sprint goal:** Replace the admin-panel look with a **consumer streaming experience**: a
> full-bleed, artwork-first UI with hero, rows, rich detail pages and a full-screen player.
> Netflix-like **layout and interaction patterns**, with Somra's **own** visual identity — no
> third-party logos, colours or trade dress ([`../../docs/design/brand-guide.md`](../../docs/design/brand-guide.md)).
>
> **Related:** [`../current-state.md`](../current-state.md) ·
> [`../sprint-11-playback-foundation/01-tasks.md`](../sprint-11-playback-foundation/01-tasks.md) ·
> [`../i18n-localization.md`](../i18n-localization.md)
>
> **Estimate:** ~3 weeks · **Milestone:** M5.5 (Core experience & performance)

## Responsible Role(s)
- Frontend (primary), UI/UX Designer, Backend (metadata + images)

## Dependencies
- Sprint 11 (working playback, TV model, runtime, tracks, next-episode, signed media URLs).

## Epics and Tasks

### Epic A: Design foundation
- [ ] A1 — Extend tokens: spacing, type scale, elevation, gradients/scrims, motion durations,
  `prefers-reduced-motion` | Acceptance: documented in `docs/design/`; all four themes pass
  WCAG AA contrast (fixes white-on-white primary button in Noir).
- [ ] A2 — Self-hosted UI font (OFL licence) + icon set (ISC/MIT licence, e.g. Lucide)
  | Acceptance: no text-glyph icons ("←") left.
- [ ] A3 — Primitives on Radix: IconButton, Badge, Select, Tabs, Menu, Tooltip, Avatar, Sheet
  | Acceptance: no ad-hoc native `<select>` styling in pages.

### Epic B: App shell split (consumer vs admin)
- [ ] B1 — Consumer shell: transparent → solid top bar on scroll, logo, Home / Movies / Shows /
  My List, expanding search, profile avatar menu; mobile bottom nav | Acceptance: ≤ 5 primary
  items; works at 360 px width.
- [ ] B2 — Admin area `/admin/*` with sidebar: libraries & paths, scans, users, status, settings,
  subtitles management, (on-hold) requests/automation | Acceptance: no admin/ops controls on
  consumer screens.

### Epic C: Home
- [ ] C1 — Hero billboard: backdrop with scrim, title logo (fallback: title), overview, Play /
  More info; rotates featured items | Acceptance: LCP image ≤ 2.5 s on a 4c/8GB server over LAN.
- [ ] C2 — Rows: Continue watching (landscape cards + progress), Recently added, My List,
  per-genre rows; arrow buttons, keyboard, scroll-snap | Acceptance: rows scroll with mouse,
  touch and keyboard.
- [ ] C3 — Hover/focus card preview (desktop) with quick Play / My List | Acceptance: no layout
  shift; disabled under reduced motion.

### Epic D: Detail page
- [ ] D1 — Full-bleed backdrop, metadata line (year · runtime · rating · genres), Play/Resume,
  My List, cast with photos, "More like this" | Acceptance: matches design review.
- [ ] D2 — Shows: season selector + episode list with stills, overview, runtime, progress
  | Acceptance: episode Play starts that episode.

### Epic E: Player
- [ ] E1 — Full-screen player route without app chrome; custom overlay controls: play/pause,
  ±10 s, seek bar with buffered range, volume, fullscreen, auto-hide | Acceptance: keyboard
  shortcuts and touch work; native controls removed.
- [ ] E2 — Audio and subtitle track picker (WebVTT), quality picker | Acceptance: switching
  keeps the position.
- [ ] E3 — Next-episode card with countdown, title overlay | Acceptance: binge flow S01E01 →
  S01E02 without leaving the player.

### Epic F: Backend support (metadata + images)
- [ ] F1 — Metadata enrichment from TMDB: runtime, rating, tagline, logos, episode stills /
  overview, trailers (video keys), genre list endpoint, rows-by-genre endpoint | Acceptance:
  OpenAPI updated first; FE types generated (no hand-written browse types).
- [ ] F2 — Local image pipeline: download + cache artwork, serve size variants (e.g. 342 / 780
  / 1280 px), ETag + immutable caching, blurhash or dominant colour; pure-Go, CGO-free
  | Acceptance: no TMDB hotlinks; `original`-size backdrops never sent to cards.
- [ ] F3 — Subtitle delivery as WebVTT: external `.srt`/`.ass` converted, embedded text
  subtitles extracted; image subtitles (PGS) burned in only when selected | Acceptance:
  subtitle picker in the player shows and switches all text tracks.

### Epic G: Browse pages
- [ ] G1 — Movies / Shows pages: responsive poster grid with correct virtualization
  (measured rows), cursor-based infinite scroll, genre / year / watched filters, sort
  | Acceptance: no overlapping rows at any breakpoint; 10k items scroll smoothly.
- [ ] G2 — Search page: instant results grid with keyboard navigation (arrows / Enter / Esc)
  | Acceptance: results appear within 150 ms on a 10k library.

## Acceptance Criteria (Sprint Output)
- A first-time user finds, starts and switches between titles without seeing any admin control;
  the UI is responsive (360 px → 4K) and fully localized (en-US + tr-TR).

## Risks
- Visual scope creep → fixed design review checkpoint at the end of Epic A/B before pages.

## Out of Scope
- Native mobile/TV apps, trailers autoplay in hero (later), profiles picker redesign (later).
