# ADR 0002: Branching Model and Development Workflow

**Date:** 2026-06-25

**Status:** Accepted

**Context**
The project is in early bootstrap on GitHub (`main` + `dev` branches already created). We need a clear, lightweight branching model that supports the minimal scope and future contributors/agents while keeping `main` stable.

**Decision**
- `main`: Production-ready / released state. Only merged from `dev` after review.
- `dev`: Primary development branch. All new work happens here.
- Feature / task branches: Created from `dev` as `feature/xxx` or `task/xxx` when a plan phase or large task warrants isolation.
- PRs: Required for `dev → main`. Small task branches may merge directly to `dev` after local verification.
- Commits: Follow conventional style where practical (`feat:`, `fix:`, `docs:`, `chore:`).

**Workflow**
1. Start new work on `dev` (or a short-lived task branch off `dev`).
2. Use the implementation plan (`docs/plans/...`) as the source of truth.
3. Commit frequently after each task (or logical group).
4. When a phase or meaningful milestone is complete, open PR `dev → main`.
5. Update ADRs for any architecture or process changes.

**Consequences**
- Keeps `main` clean for cloning and early testers.
- Simple enough for a solo/minimal project while scaling if more contributors join.
- Aligns with the existing `writing-plans` + `subagent-driven-development` approach.

**References**
- Initial repo bootstrap commands (main created first, then dev).
- Implementation plan `docs/plans/2026-06-25-prompt-os-implementation-plan.md`