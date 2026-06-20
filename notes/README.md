# Somra — Notes (Obsidian vault)

> Working notes vault for briefings, retros, and draft decisions.  
> **Not binding.** On conflict with [`.plan/`](../.plan/), `.plan/` wins.

Open this folder as your Obsidian vault. Link to binding docs with relative paths, e.g.  
`[[../.plan/project-brief|Project Brief]]` or `[Roadmap](../.plan/roadmap.md)`.

---

## Folder layout

| Folder | Purpose | Git |
|---|---|---|
| [`briefings/`](./briefings/) | Sprint kickoff, weekly status, demo scripts | Committed |
| [`decisions/`](./decisions/) | Draft ADRs — promote to `.plan/` or `docs/` when final | Committed |
| [`personal/`](./personal/) | Private scratch notes | **Gitignored** |

---

## Quick links (binding docs)

- [Planning index](../.plan/00-index.md) — active sprint dashboard
- [Project brief](../.plan/project-brief.md)
- [Roadmap & milestones](../.plan/roadmap.md)
- [Definition of Done](../.plan/definition-of-done.md)
- [AGENTS.md](../AGENTS.md) — agent workflow

---

## Rules

1. **Do not duplicate spec** — link to `.plan/` task files instead of copying acceptance criteria.
2. **Finalize decisions** — when a decision in `decisions/` is agreed, update the relevant `.plan/` or `docs/` file and add a “Promoted” line in the note.
3. **Issues for execution** — task *status* lives in GitHub Issues, not in Obsidian checkboxes alone.
4. **English** for committed notes (same as source code and `.plan/`).

---

## Obsidian tips

- Enable **Tasks** or **Kanban** plugins for active-sprint boards that mirror GitHub Issues.
- Pin [Planning index](../.plan/00-index.md) or this README as a tab for daily context.
- Use `briefings/sprint-NN-kickoff.md` from the template in [`briefings/README.md`](./briefings/README.md).
