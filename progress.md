# PromptOS Progress Handoff

> **Resume file** — update this when pausing work. (Same content as `progress.Md` if you use that name locally.)

**Last updated:** 2026-06-26  
**Branch:** `dev`  
**Repo:** https://github.com/HappyMonkeyAI/prompt-os

---

## Where we are

| Phase | Status | Notes |
|-------|--------|--------|
| Phase 0 — Research & spec | Done | Blueprint schema, TUI research, ADRs |
| Project foundations F1–F3 | Done | `MANIFEST.md`, `TASKS.md`, ADR 0003, `docs/decisions/` |
| Phase 1 — TUI foundation | Done | Wizard, hardware scan, provider keys; review fixes on `dev` |
| Phase 2 — LLM orchestration | Done | `internal/llm/`, blueprint TUI; PR #3 fixes merged on `dev` |
| **Phase 3 — Execution engine** | **Done** | `internal/execute/` disk + bootstrap + config drop |
| Phase 4 — Security & triage | Not started | Key baking hardening, emergency shell |
| Phase 5 — Bootable ISO | Not started | Alpine live image |

**Current status:** Phase 3 complete on `dev`. Execution logic exists but is **not wired into the main TUI flow** yet.

---

## Recent commits (newest first)

```
68b105d feat(execute): Task 3.3 config and key drop into chroot
d37517f feat(execute): Task 3.2 bootstrap plan for arch/debian/ubuntu
086a71f feat(execute): Task 3.1 target disk preparation
815a181 docs: implement Project Foundations (F1-F3)
f97c29f docs: update main implementation plan with Project Foundations section
```

---

## Code map (what exists)

| Area | Path | Purpose |
|------|------|---------|
| Entry | `cmd/promptos/main.go` | TUI entry |
| Wizard / keys | `internal/tui/wizard.go`, `provider.go` | Conversational flow + API keys |
| Blueprint UI | `internal/tui/blueprint.go` | LLM → validated blueprint |
| LLM | `internal/llm/` | Providers, prompts, validator, blueprint struct |
| Hardware | `internal/hardware/scan.go` | CPU/RAM/GPU/disk summary for prompts |
| **Execute** | `internal/execute/disk.go` | GPT layout, dry-run / `ConfirmWipe` |
| | `internal/execute/bootstrap.go` | arch / debian / ubuntu install steps |
| | `internal/execute/configdrop.go` | Write `configs` under chroot, `0600` for keys |

**Safety pattern (all execute steps):** `DryRun: true` plans only; real work needs explicit `Confirm` / `ConfirmWipe`.

---

## Verification (local, ad-hoc)

From repo root (Go on `PATH`):

```bash
export PATH=$HOME/go/bin:$PATH
cd ~/projects/prompt-os
go test ./internal/execute/... -count=1
go build ./...
```

No project-wide CI config yet — treat the above as the current smoke check.

---

## Documentation index

- `README.md` — overview & status
- `CONTEXT.md` — operating rules & stack
- `MANIFEST.md` — scope in/out
- `TASKS.md` — active/completed tasks
- `docs/plans/2026-06-25-prompt-os-implementation-plan.md` — full roadmap (incl. foundations)
- `docs/blueprint-schema.md` — JSON contract
- `docs/adr/` — ADR 0001–0003
- `docs/decisions/` — dated decision records

---

## Pick up tomorrow — suggested order

1. **Wire execution pipeline in TUI** — after blueprint approval: choose target disk → dry-run plan → user confirm → bootstrap → config drop (still safe defaults).
2. **Phase 4** — `internal/security/` key baking review, emergency triage shell stub.
3. **Open/update PR** — `main` ← `dev` when you want a merge checkpoint (Phase 3 milestone).
4. **Optional** — refresh `CONTEXT.md` stack line (TUI is Bubble Tea, not TBD).

---

## Environment reminders

- Project path: `~/projects/prompt-os`
- `go` may need: `export PATH=$HOME/go/bin:$PATH`
- API keys: collected in TUI only; never commit real keys
- PR history: Phase 2 was **PR #3**; Phase 3 work is on `dev` after that line

---

**Handoff complete.** Start with `TASKS.md` and this file when resuming.