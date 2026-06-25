# PromptOS Context

**Source of Truth Hierarchy**
- README.md: High-level overview and goals.
- This CONTEXT.md: Operating manual, rules, stack, decisions.
- docs/adr/: Architecture and workflow decisions (durable).
- research/: External references only (not source of truth).
- HERMES.md: Agent-specific guidance (if needed).

**Project Assumptions**
- Minimal bootstrap DIY system, not a full distro or package repository.
- TUI-based proof-of-concept (web/GUI later).
- Relies on live LLM APIs (user keys baked securely) + standard bootstrap tools (archinstall, debootstrap, pacstrap).
- Host: Minimal live env (Alpine preferred for size/speed).
- Target: User-chosen base distro customized via AI blueprint.

**Non-Negotiable Rules**
- Never ship heavy local LLMs in the ISO; use external providers or local endpoints via API.
- Keys must be handled securely (baked for convenience but protected; emergency triage agent uses them).
- Prefer upstream tools and package managers; AI only orchestrates and generates configs.
- Scope control: Focus on installer/orchestrator + self-healing hooks. No custom DE, kernel, or package maintenance.
- Verification: All claims backed by research or live execution; no invented results.

**Stack & Runtime (Initial)**
- TUI: TBD (bubbletea/Go for static binary vs Textual/Python for rapid AI SDK integration). Research pending.
- Execution: chroot, standard debootstrap/pacstrap + AI JSON blueprints for packages/drivers/configs.
- Configs: systemd units, display managers (GDM/SDDM/LightDM), remote access (Open Cloud?), environment for keys.
- Security: /etc/environment.d/ai-keys.conf or keyring; rescue target reroute for triage.
- Research/Planning: Use local research workflows, writing-plans, repository governance.

**Workflow Protocols**
- Bootstrap: Standard repo docs (README/CONTEXT/HERMES/ADR/research).
- Development: writing-plans → subagent-driven-development or coding CLI (codex/gemini/agy) when specified.
- Research: deep-research or research-workflows for topics like TUIs, installers, key storage.
- Changes: ADR for any durable decision affecting behavior or workflow.
- Commits: Frequent, after tasks; TDD where code is involved.

**Resolved Decisions**
- Name: PromptOS (verified available as OS/distribution name).
- TUI over GUI for POC.
- Bake keys for "it just works" experience + emergency AI triage.
- Pivot from full custom distro ("Vibes Linux") to AI-first installer framework.

**What Not To Do**
- Do not over-scope into full distro maintenance or custom packages.
- Do not hardcode configs that stale; always dynamic via LLM.
- Do not assume local LLM in bootstrap ISO.

**Mutable State / Entry Points (Future)**
- TUI wizard screens.
- Hardware scan + preference mapper.
- Blueprint JSON schema.
- Execution engine (partition, chroot, install, config drop).
- Rescue/triage shell hook.

This file is the compact operating manual. Update only when architecture or rules change. Legacy or exploratory notes go to research/ or ADR.