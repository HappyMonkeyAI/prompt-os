# ADR 0003: Project Foundations and Governance Practices

**Date:** 2026-06-26  
**Status:** Accepted

## Context
As PromptOS moves beyond the initial research and TUI/LLM foundation phases, we need stronger ongoing governance practices to keep the project maintainable, auditable, and aligned with its minimal scope philosophy.

## Decision
We adopt three lightweight foundation practices:

1. **Enhanced ADR + CONTEXT.md**
   - All significant architectural, scope, and technology decisions must be recorded as ADRs in `docs/adr/`.
   - `CONTEXT.md` serves as the living operating manual and source of truth for conventions.

2. **Manifest + Task List**
   - `MANIFEST.md` defines project scope, components, and non-goals.
   - `TASKS.md` provides a lightweight, visible list of active and upcoming work.

3. **File-based Decision & Memory Record**
   - `docs/decisions/` will contain dated, lightweight decision records.
   - This creates a canonical, file-based audit trail of build-time choices (inspired by Hermes memory patterns and LLM Codex Reference Vault ideas).

## Consequences
- Improved long-term maintainability and onboarding.
- Better traceability of why decisions were made.
- Slight increase in documentation overhead (mitigated by keeping formats lightweight).

## Related
- Supersedes informal documentation practices used in Phases 0–2.
- Prepares the project for Phase 3 (Execution Engine) and beyond.