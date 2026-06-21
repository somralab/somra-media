# Translating Somra

Binding i18n rules: [`.plan/i18n-localization.md`](../../.plan/i18n-localization.md).

## In-repo workflow (PR)

1. Add keys to `web/src/i18n/locales/en-US/*.json` first.
2. Mirror keys in `web/src/i18n/locales/tr-TR/*.json`.
3. Backend messages: `internal/platform/i18n/locales/active.en-US.toml` + `active.tr-TR.toml`.
4. Run `make i18n-check` (parity + overflow ratio check).

Key format: `domain.context.key` — never hardcode user-visible strings in code.

## Weblate (self-host MVP)

```bash
cd deploy/weblate
cp .env.example .env
docker compose up -d
```

1. Create a **Component** pointing at this Git repository.
2. Map file masks:
   - `web/src/i18n/locales/en-US/*.json` (template)
   - `web/src/i18n/locales/tr-TR/*.json`
   - `internal/platform/i18n/locales/active.*.toml`
3. Configure **Git export** (push) and **Git pull** on a schedule or webhook.

Full automation (pre-1.0): manual export/import PRs are acceptable. Post-1.0: add CI hook to verify Weblate commits pass `i18n-check`.

## Pseudo-locale / overflow

CI enforces translation length ratio (default 1.5× en-US) via `scripts/i18n-check.sh --overflow-ratio=1.5`.

## Adding a new language (post-1.0)

1. Add locale folder under `web/src/i18n/locales/<code>/`.
2. Add backend `active.<code>.toml`.
3. Register in i18n bundle negotiation.
4. Extend `i18n-check` target locales.

No code changes beyond registration when following existing patterns.
