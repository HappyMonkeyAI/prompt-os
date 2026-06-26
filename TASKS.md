# PromptOS Task List

This file tracks active and upcoming work items.  
For historical decisions, see `docs/decisions/`.  
**Handoff:** see `progress.md` / `progress.Md`.

**Last Updated:** 2026-06-26

---

## Active / Next

- [x] Wire TUI: blueprint approval → execute pipeline (dry-run default)
- [x] Phase 4: Security & key baking hardening (`internal/security/`)
- [x] Task 4.2: Emergency triage shell (`internal/security/triage_agent.go` + stub shell)
- [x] Task 5.1: Minimal live ISO build artifacts (`build/`)
- [x] Task 5.2: End-to-end VM test
- [ ] Optional: PR `main` ← `dev` for Phase 3/4 milestone

## Completed

- [x] Phase 3: Execution Engine (3.1 disk, 3.2 bootstrap, 3.3 config drop)
- [x] Phase 1: TUI Foundation (all tasks + code review fixes)
- [x] Phase 2: LLM Orchestration (all tasks + code review fixes)
- [x] Project Foundations F1–F3 (MANIFEST, TASKS, ADR 0003, docs/decisions/)
- [x] Update main implementation plan with Project Foundations section

## Backlog / Future

- [ ] Custom "AI Linux Installer Builder" agent role (`internal/agent/`)
- Integration with Article Research MCP for ongoing research
- Advanced memory system (observational memory patterns)
- Cancellation signals and improved error handling in TUI
- Phase 5: Bootable Alpine live ISO

---

**Note:** This list is kept lightweight. For detailed planning, see `docs/plans/`.