# PromptOS Task List

This file tracks active and upcoming work items.
For historical decisions, see `docs/decisions/`.
**Handoff:** see `progress.md` / `progress.Md`.

**Last Updated:** 2026-06-26

---

## Active / Next

- [x] Optional: PR `main` ← `dev` for Phase 3/4 milestone (`f5736caa`, PR #7)
- [x] Re-run full verification on `internal/execute/` after latest edits
- [x] Refresh `progress.md` with PR fix status, execute phase, and verification results
- [ ] Optional: LLM context timeout improvements from Amazon Q #2/#4

## Completed

- [x] Phase 4: Security & key baking hardening (`internal/security/`)
- [x] Task 4.2: Emergency triage shell (`internal/security/triage_agent.go` + stub shell)
- [x] Task 5.1: Minimal live ISO build artifacts (`build/`)
- [x] Task 5.2: End-to-end VM test
- [x] Feature: AI Linux Installer Builder agent role (`internal/agent/installer_builder.go`)
- [x] Phase 5: Bootable Alpine live ISO follow-on

## Backlog / Future

- [ ] Phase 3: Execution Engine follow-up
- [ ] Advanced memory system (observational memory patterns)
- [ ] Cancellation signals and improved error handling in TUI

---

**Note:** This list is kept lightweight. For detailed planning, see `docs/plans/`.
