#!/usr/bin/env bash
set -euo pipefail
# Auto-start PromptOS TUI after Alpine boot (local.d hook).
# Running inside the target system.

exec /usr/local/bin/promptos </dev/tty1 >/dev/tty1 2>&1
