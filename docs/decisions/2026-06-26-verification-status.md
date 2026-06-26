# PromptOS Decision Record

**Date:** 2026-06-26

## Decision
Treat the remaining high-risk items as validated by ad-hoc verification only:
- local `go test ./...`
- local `go build ./...`

There is no project-wide CI config yet.

## Rationale
Avoid overstating verification coverage. Keep `progress.md` aligned with what was actually checked.
