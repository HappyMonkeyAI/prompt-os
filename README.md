# PromptOS

**AI-First Dynamic Linux Bootstrap Framework**

PromptOS is a minimal, TUI-driven installer bootstrap that uses live LLM calls (via user-provided API keys from providers like OpenAI, Anthropic, Gemini, or local Ollama) to generate personalized, hardware-aware Linux installations on the fly.

Instead of a static ISO or pre-baked distro, it acts as an intelligent orchestrator: boot a tiny live environment (e.g., Alpine), connect your AI provider, answer high-level preference questions in a conversational wizard, and let the AI architect the exact package list, drivers, configs, and services for your chosen base (Arch, Ubuntu, Debian, etc.).

**Core Idea (from initial concept transcript):** A "never the same twice" installer that keeps the maintenance burden on upstream package managers and LLMs, delivering a self-healing, voice-first, AI-ready desktop with baked-in provider credentials and an emergency triage agent.

**Status:** Bootstrap / Research Phase. Minimal viable TUI orchestrator + execution engine on top of proven tools like archinstall, debootstrap, or pacstrap.

See `CONTEXT.md` for operating manual, `research/` for references, and `docs/adr/` for decisions.

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
- TUI: Go + Bubble Tea or Python + Textual (tradeoff analysis pending research).
- Base: Alpine live for installer host; target any major distro.
- Execution: chroot + standard bootstrap tools + AI-generated JSON blueprints.
- Security: Bake keys into /etc/environment.d or user keyring; emergency agent for diagnostics.

Project bootstrapped per standard repository governance (README, CONTEXT, HERMES, ADR, research). Deep research on related installers, TUIs, key storage, and novelty ongoing.