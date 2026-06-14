# Accessibility Checklist (M3 Alpha)

Use this checklist when validating browse/discover/detail screens across all four themes.

## Keyboard

- [ ] Tab order follows visual layout (header → main → actions).
- [ ] Poster cards and nav links are focusable with visible focus ring.
- [ ] Search dropdown items reachable via keyboard (Enter activates link).
- [ ] Modal/dialog controls (if any) trap focus appropriately.

## Screen readers

- [ ] Page landmarks: `header`, `main`, `nav` with `aria-label`.
- [ ] Shelves use `aria-labelledby` tied to heading ids.
- [ ] Search uses `role="search"` and labeled input.
- [ ] Error states use `role="alert"`.

## Contrast (each theme: Cinematic, Aurora, Noir, Minimal)

- [ ] Body text (`text-text`) on `bg-bg` ≥ 4.5:1.
- [ ] Muted text on surface ≥ 4.5:1 for essential labels.
- [ ] Primary buttons/links meet 4.5:1 against background.
- [ ] Progress bars and badges remain visible in dark and light themes.

## Motion & media

- [ ] `prefers-reduced-motion`: skeleton pulse acceptable; no required motion for tasks.
- [ ] Poster `alt=""` decorative; title exposed in link `aria-label`.

## i18n / overflow

- [ ] Long Turkish strings do not break grid/list layouts (truncate where needed).
- [ ] No hardcoded user-facing strings in browse/discover/search/detail flows.
