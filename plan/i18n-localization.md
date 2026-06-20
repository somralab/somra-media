# Somra — Internationalization & Localization (i18n / l10n)

> Multi-language support is a **cross-cutting and binding** requirement for the project.
> Development must be i18n-compliant **from day one**: no user-facing text may be hardcoded in
> source. This rule is enforced via [`definition-of-done.md`](./definition-of-done.md).

Related: [`project-brief.md`](./project-brief.md) · [`architecture.md`](./architecture.md) · [`tech-stack.md`](./tech-stack.md) · [`definition-of-done.md`](./definition-of-done.md) · [`roadmap.md`](./roadmap.md)

---

## 1. Target Languages

| Language | Code | Role |
|---|---|---|
| English (US) | **en-US** | **Source locale + fallback.** Reference for all keys. |
| Turkish (Turkey) | **tr-TR** | First additional translation; product priority. |

- **Source locale is en-US:** All text keys are defined in en-US first; missing translations fall back to en-US.
- New languages must be addable by **translation files only** (no code changes).

## 2. Scope

i18n/l10n covers all of the following layers:

1. **Frontend UI text** (required) — all labels, buttons, messages, empty/error states.
2. **Backend messages** — API error/validation messages (via keys + locale negotiation).
3. **Notification templates** — email / Discord / webhook (TR + EN templates). See Sprint 08.
4. **Media metadata language** — movie/series title/description fetched from providers per user locale. See Sprint 02.
5. **Date / number / format localization (l10n)** — display per locale conventions.
6. **Documentation** — user and developer docs in TR + EN. See Sprint 10.

## 3. Language Selection (Negotiation)

Priority order (high to low):

1. **User profile preference** (logged-in user) — persistent.
2. **System default** (set by admin during setup/settings).
3. **Browser auto-detection** (`Accept-Language`) — anonymous/first visit.
4. **Fallback:** en-US.

- User can always **override** system default (per-user).
- Backend carries locale context through all message generation.

## 4. Technical Approach

### 4.1 Translation files (in-repo)
- Translations live **in the repo** (JSON), contributions via PR.
- Namespace/key structure organized by module (`common`, `library`, `player`, `auth`, `errors`).
- Key naming standard: `domain.context.key`.
- **Weblate** (self-host, git-integrated) for community translation. Setup in Sprint 10.

### 4.2 Frontend
- i18n library: **`i18next` + `react-i18next`** (see [`tech-stack.md`](./tech-stack.md) §2).
- Pluralization, interpolation, date/number formatting (`Intl`) support.
- Dev-time missing key warnings + CI checks.

### 4.3 Backend
- Message catalog: **`nicksnyder/go-i18n/v2` + `golang.org/x/text`** (locale matching).
- API errors return both machine-readable code and localized message.

## 5. Binding Development Rules (Anti-Drift)

> These rules are mandatory for all sprint/task files and enforced in [`definition-of-done.md`](./definition-of-done.md).

1. **No hardcoded text:** No user-facing text directly in code; always resolve via keys.
2. **Source locale integrity:** Every new feature ships with en-US keys; missing keys fail CI.
3. **tr-TR in sync:** tr-TR translation added when feature completes (missing = not Done).
4. **Format safety:** Date/number/currency display via locale APIs; no manual string concatenation.
5. **Direction/length:** UI resilient to text length changes (TR↔EN).

## 6. Test & Quality

- **CI check:** Missing/extra keys, unused key detection.
- **QA:** New strings verified in both languages each sprint; pseudo-locale for overflow/length tests.
- **Release gate (Sprint 10):** **100% key completeness** for en-US and tr-TR is acceptance criteria.

## 7. Sprint Distribution (Cross-Cutting)

| Sprint | i18n contribution |
|---|---|
| 01 | Frontend + backend i18n infrastructure, key standard, CI check, locale negotiation skeleton |
| 02 | Metadata language fetched per user locale |
| 03 | User profile language preference + browser auto-detection |
| 06 | Language selection in onboarding + system default language setting |
| 08 | TR/EN localization of notification templates |
| 10 | TR+EN docs, translation completeness QA, (optional) translation platform setup |

## 8. Decisions (Closed)

| Decision | Outcome |
|---|---|
| Frontend i18n library | `i18next` + `react-i18next` |
| Backend i18n library | `nicksnyder/go-i18n/v2` + `golang.org/x/text` |
| Translation file format | JSON, in-repo |
| External translation platform | **Weblate** (self-host, OSS, git-integrated) — setup Sprint 10 |
