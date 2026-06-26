# PromptOS

**AI-First Dynamic Linux Bootstrap Framework**

PromptOS is a minimal, TUI-driven installer bootstrap that uses live LLM calls (via user-provided API keys from providers like OpenAI, Anthropic, Gemini, or local Ollama) to generate personalized, hardware-aware Linux installations on the fly.

Instead of a static ISO or pre-baked distro, it acts as an intelligent orchestrator: boot a tiny live environment (e.g., Alpine), connect your AI provider, answer high-level preference questions in a conversational wizard, and let the AI architect the exact package list, drivers, configs, and services for your chosen base (Arch, Ubuntu, Debian, etc.).

**Core Idea (from initial concept transcript):** A "never the same twice" installer that keeps the maintenance burden on upstream package managers and LLMs, delivering a self-healing, voice-first, AI-ready desktop with baked-in provider credentials and an emergency triage agent.

**Status:** Phase 3 (Execution Engine) complete on `dev`. Disk prep, bootstrap plans, and config/key drop implemented in `internal/execute/`. Next: wire TUI → execute pipeline, then Phase 4 (security & triage).

See `progress.md` (or `progress.Md`) for handoff details when resuming.

## Quick Start (Future)
- Build or download minimal live ISO with TUI.
- Run wizard, provide API key.
- AI generates blueprint → chroot install → reboot into personalized system.
- Emergency shell fallback with AI triage using baked keys.

## Name
PromptOS (chosen after availability check; academic paper only, no active OS/distributor conflict).

## Goals
- Minimal scope: TUI + orchestrator + secure key handling + self-repair hooks. Not a full distro.
- High impact for DIY users wanting AI-native setup without complexity of maintaining a custom OS.
- Extensible to GUI later.

## Tech Directions (Initial)
- TUI: Go + Bubble Tea (chosen for static binary on minimal live ISO)
- LLM: Pluggable client abstraction (OpenAI, Anthropic, Gemini, Ollama)
- Structured output: JSON Blueprint validated against schema
- Execution: archinstall / debootstrap / pacstrap

See `docs/plans/2026-06-25-prompt-os-implementation-plan.md` for the detailed roadmap.