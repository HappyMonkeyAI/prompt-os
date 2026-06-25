# ADR 0001: Initial Scope, Name, and Project Bootstrap

**Date:** 2026-06-25

**Status:** Accepted

**Deciders:** Stephen Phillips (user), Hermes Agent (operator)

## Context
Transcript with Google Gemini outlined "AI First Linux" / PromptOS: a minimal TUI-based dynamic installer/bootstrap framework using live LLM calls for personalized installs, API key baking, and emergency AI triage agent. User requested bootstrap in ~/projects/prompt-os using standard repo governance (README/CONTEXT/ADR/research) + deep research. Confirmed preference for minimal DIY scope vs. full distro.

## Decision
- **Name:** PromptOS (verified as available for an OS/distribution; only minor academic paper precedent).
- **Scope:** Minimal viable orchestrator framework (TUI wizard + LLM blueprint generation + execution layer on top of archinstall/debootstrap/pacstrap + key handling + triage hooks). Not a full distro, package repo, or custom DE/kernel maintenance. POC TUI-first; GUI later.
- **Bootstrap:** Follow repository-documentation-governance (README, CONTEXT.md, docs/adr/, research/). Use writing-plans + research workflows for subsequent work.
- **Architecture Pillars (from transcript):** Provider-first auth, conversational preference mapping to JSON blueprints, "never the same twice" via live models, secure key baking, self-healing emergency shell.

## Consequences
- Positive: High leverage, focused development, leverages upstream tools + LLMs. Matches user's "not amazingly massive" expectation.
- Research needed: TUI framework choice, exact ISO build, key storage implementation details, triage integration.
- Future ADRs: TUI stack selection, blueprint JSON schema, security model, etc.
- Project now has canonical docs and structure for agents/contributors.

## References
- Original transcript (Google Doc ID: 1tE_YMHs8Lx6NCmBzKOT2vjllmOZqHd2dGRBm4QV5vIY)
- Project README.md and CONTEXT.md
- Initial web research on archinstall, systemd rescue targets, env var best practices (see research/ notes)