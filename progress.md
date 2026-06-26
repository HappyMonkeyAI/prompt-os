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
| **Phase 4 — Security & triage** | **Done** | Key baking hardening, triage shell in `internal/security/` |
| **Phase 5 — Bootable ISO** | **Partial** | Minimal live ISO artifacts created; full boot validation still open |

**Current status:** `dev` has a stable Phase 3/4 checkpoint. `main` was last advanced by PR #7 (`f5736caa`).

---

## Recent commits (newest first)

```
22e7d10 feat(agent): add AI Linux Installer Builder agent role
1a80077 Harden destructive execution safeguards
6cdb9a0 Merge pull request #6 from HappyMonkeyAI/pr-5
d23804c fix(security): resolve issues raised by Amazon Q Developer
44582fa docs: mark Task 5.2 complete and add VM test helper placeholder
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
| Agent role | `internal/agent/installer_builder.go` | AI Linux Installer Builder |
| Execute | `internal/execute/disk.go` | GPT layout, dry-run / `ConfirmWipe` |
| Execute | `internal/execute/bootstrap.go` | arch / debian / ubuntu install steps |
| Execute | `internal/execute/configdrop.go` | Write `configs` under chroot, `0600` for keys |
| Security | `internal/security/keybake.go`, `triage_agent.go`, `triage.sh` | Key hardening + emergency triage shell |

**Safety pattern (all execute steps):** `DryRun: true` plans only; real work needs explicit `Confirm` / `ConfirmWipe`.

---

## Verification (local, ad-hoc)

Fresh ad-hoc verification for the changed `internal/execute/` area:

````bash
export PATH=$HOME/go/bin:$PATH
cd /home/stephen/projects/prompt-os
go build ./...
go test ./internal/execute/... -count=1 -v
# 2026-06-26 result: all execute tests passed
````

No project-wide CI config yet — treat the above as the current smoke check.

---

## Documentation index

- `README.md` — overview & status
- `CONTEXT.md` — operating rules & stack
- `MANIFEST.md` — scope in/out
- `TASKS.md` — active/completed tasks
- `progress.md` / `progress.Md` — this handoff file
- `docs/plans/2026-06-25-prompt-os-implementation-plan.md` — full roadmap
- `docs/blueprint-schema.md` — JSON contract
- `docs/adr/` — ADR history
- `docs/decisions/` — dated decision records

---

## Pick up tomorrow — suggested order

1. **Refresh execution pipeline in TUI** — choose target disk → dry-run plan → confirm → bootstrap → config drop.
2. **Run full verification sweep** — `go build ./... && go test ./... -count=1`; capture output in a temp `/tmp/hermes-verify-*` script if needed.
3. **Refresh this handoff** — keep `progress.md` / `progress.Md` aligned with `TASKS.md`.
4. **Optional** — address LLM context timeout improvements from Amazon Q #2/#4.

---

## Environment reminders

- Project path: `/home/stephen/projects/prompt-os`
- `go` may need: `export PATH=$HOME/go/bin:$PATH`
- API keys: collected in TUI only; never commit real keys
- PR history: PR #3 was phase 2; PR #7 merged the Phase 3/4 checkpoint to `main`

---

**Handoff complete.** Start with `TASKS.md` and this file when resuming.
