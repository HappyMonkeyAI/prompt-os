# PromptOS Research Handoff Note

**Project:** PromptOS (AI-first dynamic Linux bootstrap installer)

**Working Thesis:** A lightweight TUI orchestrator running from a minimal live ISO (Alpine) that authenticates a user AI provider upfront, runs a conversational wizard, generates hardware-aware JSON blueprints via LLM, executes installs via standard tools (archinstall/debootstrap), bakes keys securely, and provides an AI-powered emergency triage shell. Minimal scope: installer framework only.

**Target Users:** DIY Linux users (minimal technical skills to advanced) wanting personalized, AI-native, self-healing setups without maintaining a full custom distro.

**Key Differentiation:** Live/runtime LLM orchestration vs. static pre-generated configs or heavy enterprise MLOps stacks. "Never the same twice" + baked credentials + triage agent.

**Competitors / Adjacent (from research):**
- archinstall (flexible library + interactive/declarative JSON-driven Arch installs) — strong foundation to build on or invoke.
- EndeavourOS/Calamares GUI installers.
- RHEL AI / Canonical Charmed (enterprise scale, not DIY conversational).
- AI package managers (Renovate etc.) and post-install LLM server agents — complementary but not installer-focused.
- No direct live conversational bootstrap matches found.

**Research Angles Explored (initial):**
- No exact equivalents; concept appears novel.
- TUI: bubbletea (Go, static binary ideal for minimal ISO) vs Textual (Python, rapid dev with AI SDKs) — tradeoff analysis pending deeper dive.
- Key storage: /etc/environment.d/ + systemd EnvironmentFile= recommended and matches transcript (OpenAI best practices align); alternatives: LoadCredential, encrypted keyrings, .env with perms. Risks (env var exposure) mitigable with proper setup.
- Rescue/triage: Standard systemd rescue.target / emergency.target easily extended with custom scripts that run diagnostics (journalctl) and invoke LLM via baked keys for interactive repair.
- ISO: Alpine live + custom scripts or archiso patterns viable for tiny RAM-disk host.

**Risks / Unknowns:**
- TUI perf/size on minimal live env.
- LLM reliability for complex hardware/driver mapping.
- Security of baked keys (user education + permissions).
- Exact blueprint schema and error handling in execution engine.

**Immediate Prep / Next Actions:**
- Complete 9-lens deep research or targeted follow-ups (TUI benchmarks, full ISO examples, keyring integration code patterns).
- Writing-plans for TUI prototype or blueprint schema.
- ADR for TUI stack and security model.
- Index in launcher registry once stable.

**Sources (selected):** archinstall GitHub/docs, systemd rescue.target docs (Red Hat/Ubuntu), OpenAI API key safety guide, various Reddit/dev.to discussions on AI Linux tools. All claims verified via tool output.

**Date:** 2026-06-25

This note enables quick resumption. Update with new findings.