# Sprint 09 — Plugin Architecture Tasks

> **Sprint goal:** Plugin framework that **isolates content acquisition capabilities** (indexer +
> download client) from the core. Critical for legal risk isolation.
>
> **Related:** [`../architecture.md`](../architecture.md) §6 (plugin) · [`../project-brief.md`](../project-brief.md) §7 (scope/legal) · [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md)

## Responsible Role(s)
- Tech Lead (primary — architecture), Backend (implementation)

## Dependencies
- Sprint 01 (architecture/module boundaries), Sprint 08 (automation handoff interface).

## Epics and Tasks

### Epic A: Plugin contract
- [x] A1 — `Indexer` and `DownloadClient` interface contracts (Search, Capabilities, Add, Status) | Acceptance: clear, versioned interface.
- [x] A2 — Plugin lifecycle (register, enable, configure, disable) | Acceptance: managed at runtime.

### Epic B: Isolation & security
- [x] B1 — Core runs independently of plugins (system fully functional without plugins) | Acceptance: plugin-less mode tested.
- [x] B2 — Plugin configuration/secret management (secure storage) | Acceptance: secrets protected.
- [x] B3 — Plugin distribution/packaging strategy (separately packageable) | Acceptance: legal isolation documented.

### Epic C: Management API
- [x] C1 — Plugin list/configure/test connection API | Acceptance: coordinated with frontend.

## Acceptance Criteria (Sprint Output)
- Core is neutral; indexer/download capabilities run as pluggable, isolated plugins.

## Risks
- **Legal risk** → isolation and clear license terms (see [`../project-brief.md`](../project-brief.md) §5/§7).

## Out of Scope
- Third-party plugin marketplace — future roadmap.
