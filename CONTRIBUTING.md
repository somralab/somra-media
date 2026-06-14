# Contributing to Somra

Thanks for your interest in Somra. Please read this document and
[`AGENTS.md`](./AGENTS.md) before opening a PR. The authoritative product
scope and module boundaries live under [`plan/`](./plan/) — read
[`plan/project-brief.md`](./plan/project-brief.md) first if you are not sure
whether your change is in scope.

> All source code, comments and commit messages are in **English**.
> User-facing strings are localized; never hardcode them
> (see [`plan/i18n-localization.md`](./plan/i18n-localization.md)).

## Branch naming

Format: `<type>/<short-kebab-description>` (optionally suffixed by an
issue/sprint reference).

```
feat/library-scan
fix/transcode-session-leak
chore/ci-coverage-gate
docs/i18n-guide
```

Allowed types: `feat`, `fix`, `chore`, `refactor`, `test`, `docs`, `perf`,
`build`, `ci`. `main` is protected — no direct pushes, no force-pushes.

## Commit messages — Conventional Commits

Format: `<type>(<scope>): <subject>` — English, imperative mood, lower
case, no trailing period.

```
feat(library): add incremental scan with fs watcher
fix(auth): revoke refresh token on logout
test(streaming): cover transcode decision engine
```

Keep commits small and focused. Subject ≤ ~72 characters; explain the
"why" in the body.

## DCO sign-off (mandatory — no CLA)

Every commit must be signed off under the
[Developer Certificate of Origin](https://developercertificate.org):

```bash
git commit -s -m "feat(metadata): add tmdb provider"
```

`-s` appends a `Signed-off-by:` trailer. Commits without sign-off are
rejected by CI.

## Pull requests

- One logical change per PR; split large work.
- Reference the related task and acceptance criteria under [`plan/`](./plan/).
- Follow the PR template (see the repository PR template / the
  [`pr-and-code-review`](./.cursor/rules/pr-and-code-review.mdc) cursor rule).
- All CI gates must be green:
  `lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build`.
- The Definition of Done lives at
  [`plan/definition-of-done.md`](./plan/definition-of-done.md) §1.

## Code style & tests

- See [`AGENTS.md`](./AGENTS.md) for code style and testing requirements
  (Go and TypeScript/React).
- Business logic requires unit tests; coverage thresholds are enforced
  by CI (core ≥ 80 %, critical modules ≥ 90 %, frontend components
  ≥ 70 %).
- Every feature ships its `en-US` **and** `tr-TR` translation keys —
  otherwise it is not "Done".

## License

By contributing you agree that your contribution is licensed under
[AGPL-3.0-or-later](./LICENSE).
