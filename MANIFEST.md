# PromptOS Manifest

**Project:** PromptOS – AI-First Dynamic Linux Bootstrap Framework

**Version:** 0.3 (2026-06-26)

---

## Purpose
PromptOS is a minimal, TUI-driven installer that uses live LLM calls to generate personalized Linux installation blueprints. It acts as an intelligent orchestrator rather than a full custom distribution.

## Scope (In)

- TUI wizard (Go + Bubble Tea)
- LLM provider abstraction (OpenAI, Anthropic, Gemini, Ollama)
- Structured JSON blueprint generation + validation
- Hardware-aware installation planning
- Secure API key baking
- Emergency AI triage shell
- Minimal live ISO (Alpine base)

## Out of Scope (Non-Goals)

- Maintaining a full custom Linux distribution
- Heavy local LLM inference in the ISO
- GUI installer (future consideration only)
- Package repository or build system
- Kernel or bootloader development

## Core Components

| Component              | Location                        | Status     |
|------------------------|----------------------------------|------------|
| TUI Framework          | `cmd/promptos/`, `internal/tui/` | Complete   |
| LLM Abstraction        | `internal/llm/`                  | Complete   |
| Blueprint Schema       | `docs/blueprint-schema.md`       | Complete   |
| Hardware Scanner       | `internal/hardware/`             | Complete   |
| Execution Engine       | `internal/execute/`              | Complete   |
| Security & Key Baking  | `internal/security/`             | Phase 4    |
| Live ISO Build         | `build/`                         | Phase 5    |

## Technology Stack

- **Language:** Go 1.24+
- **TUI:** Bubble Tea + Lipgloss + Bubbles
- **LLM:** Pluggable clients (OpenAI-compatible + Ollama)
- **Live Environment:** Alpine Linux
- **Execution:** archinstall / debootstrap / pacstrap

## Documentation

- `README.md` – High-level overview
- `CONTEXT.md` – Operating manual
- `docs/adr/` – Architecture Decision Records
- `docs/plans/` – Implementation plans
- `MANIFEST.md` – This file (scope and components)
- `progress.md` – Current status and next actions

## Task Tracking

Active work items are tracked in:
- `TASKS.md` (or `.promptos/tasks/`)
- GitHub Issues / Project board (when enabled)

---

**Last Updated:** 2026-06-26
**Maintained By:** Stephen + Hermes Agent