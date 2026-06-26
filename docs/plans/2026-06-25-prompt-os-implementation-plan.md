# PromptOS Implementation Plan

> **For Hermes:** Use subagent-driven-development skill (or user-specified CLI) to implement this plan task-by-task.

**Goal:** Build a minimal, TUI-driven AI-first Linux bootstrap installer that uses live LLM calls to generate personalized installation blueprints, executes them via standard tools, bakes provider keys securely, and provides an emergency AI triage agent.

**Architecture:** Thin TUI orchestrator (provider auth + conversational wizard) → LLM blueprint generator (structured JSON) → Execution engine (chroot + bootstrap tools) + security & self-healing layers. All on top of existing base installers (archinstall / debootstrap / pacstrap). No custom distro maintenance.

**Tech Stack (initial direction):** Go + Bubble Tea (preferred for static binary on minimal live ISO) or Python + Textual. Alpine Linux live environment as installer host. JSON blueprint schema as the contract between LLM and executor.

**Branching:** Work on `dev`. PR to `main` at phase boundaries.

---

## Phase 0: Research & Specification

### Task 0.1: Complete targeted research on TUI frameworks
**Objective:** Decide between Go+Bubble Tea vs Python+Textual for a minimal live-ISO TUI.

**Files:**
- Create: `research/tui-framework-comparison.md`

**Steps:**
1. Search for size, dependencies, and live-ISO examples of both frameworks.
2. Document binary size, RAM usage, ease of structured output, and packaging notes.
3. Record recommendation with rationale.

**Verification:** `research/tui-framework-comparison.md` exists and contains a clear winner + justification.

### Task 0.2: Research minimal live ISO creation patterns
**Objective:** Identify the lightest reliable way to ship the TUI + dependencies in a bootable environment.

**Files:**
- Create: `research/live-iso-build-notes.md`

**Steps:**
1. Investigate Alpine Linux live + custom init scripts vs archiso.
2. Note exact commands and file layouts for embedding a Go binary or Python app.
3. Capture size targets (< 300 MB ideal).

**Verification:** Notes file contains reproducible build commands.

### Task 0.3: Define the conversational flow and JSON blueprint schema
**Objective:** Create the exact user flow and the strict JSON contract the LLM must return.

**Files:**
- Create: `docs/blueprint-schema.md`
- Create: `docs/conversational-flow.md`

**Steps:**
1. Map the wizard questions from the original transcript into discrete screens.
2. Design a minimal JSON schema (packages, drivers, configs, services, environment).
3. Include validation rules and example valid/invalid outputs.

**Verification:** Both docs exist and the schema is unambiguous.

### Task 0.4: Research secure key baking patterns
**Objective:** Finalize how API keys will be stored in the target system.

**Files:**
- Update: `research/initial-web-research-notes.md` or new file `research/key-baking-implementation.md`

**Steps:**
1. Evaluate `/etc/environment.d/`, systemd `LoadCredential`, encrypted keyrings.
2. Choose primary method + fallback with security considerations.
3. Document exact file paths and permissions.

**Verification:** Chosen method is recorded with implementation commands.

---

## Phase 1: TUI Foundation

### Task 1.1: Project scaffolding for chosen TUI language
**Objective:** Set up the Go (or Python) module with TUI dependencies.

**Files:**
- Create: `cmd/promptos/main.go` (or equivalent Python structure)
- Modify: `go.mod` (or `pyproject.toml`)

**Step 1:** Initialize module and add Bubble Tea (or Textual) + lipgloss (or Rich).

**Verification:** `go build` (or equivalent) succeeds with no errors.

### Task 1.2: Implement provider selection screen
**Objective:** First screen that lets the user choose provider and enter API key (or local endpoint).

**Files:**
- Create: `internal/tui/provider.go`

**Step 1:** Write failing test for key input validation.
**Step 2:** Implement minimal screen.
**Step 3:** Run test → pass.

**Verification:** Manual run shows working input screen with validation.

### Task 1.3: Build preference wizard screens
**Objective:** Conversational questions mapped to blueprint fields.

**Files:**
- Create: `internal/tui/wizard.go`

**Steps:** One task per major question group (stability, display, remote, hardware).

**Verification:** All wizard screens render and collect answers correctly.

### Task 1.4: Hardware scan stub
**Objective:** Collect basic system info (CPU, GPU, RAM, disk).

**Files:**
- Create: `internal/hardware/scan.go`

**Verification:** Scan returns plausible data structure.

---

## Phase 2: LLM Orchestration

### Task 2.1: Provider client abstraction
**Objective:** Unified interface for multiple LLM providers.

**Files:**
- Create: `internal/llm/provider.go` + concrete implementations.

**Verification:** Can call at least two providers with the same interface.

### Task 2.2: System prompt + structured output
**Objective:** Reliable JSON blueprint generation.

**Files:**
- Create: `internal/llm/prompts.go`
- Create: `internal/llm/validator.go`

**Verification:** LLM responses are validated against the schema defined in Phase 0.3.

### Task 2.3: Blueprint generation integration in TUI
**Objective:** Wire wizard answers + hardware scan into LLM call and display result.

**Verification:** End-to-end flow from wizard → JSON blueprint preview.

---

## Phase 3: Execution Engine

### Task 3.1: Target disk preparation
**Objective:** Safe partitioning and formatting (or delegate).

**Verification:** Dry-run and real-run modes work in a VM.

### Task 3.2: Base system installation via chroot
**Objective:** Invoke archinstall / debootstrap / pacstrap using the blueprint.

**Verification:** Target system boots with chosen base.

### Task 3.3: Config & key drop
**Objective:** Write AI-generated configs and environment files for keys.

**Verification:** Keys are present and services start after reboot.

---

## Phase 4: Security & Self-Healing

### Task 4.1: Key baking implementation
**Objective:** Implement the chosen secure storage method from Phase 0.4.

**Verification:** Keys survive reboot and are not world-readable.

### Task 4.2: Emergency triage shell
**Objective:** Reroute failed boot to custom rescue target with AI agent.

**Verification:** Triggering a failure drops to the triage TUI which can call the LLM.

---

## Phase 5: Bootable Artifact

### Task 5.1: Minimal live ISO build
**Objective:** Produce a bootable ISO containing the TUI.

**Verification:** ISO boots in QEMU/VirtualBox and runs the wizard.

### Task 5.2: End-to-end VM test
**Objective:** Full install from live ISO to working desktop.

**Verification:** Installed system matches user preferences and has working AI keys + triage.

---

**Plan Notes**
- All tasks are intended to be 2–15 minutes of focused work.
- Use TDD for any new code.
- Commit after every task (or small logical group).
- Update ADRs when architecture decisions are locked.
- Revisit this plan after Phase 0 research completes.

**Next after this plan is written:** Proceed to Task 0.1 (Phase 0 research).
## Task 5.2 command
Runs the VM test helper script to verify end-to-end behavior.
