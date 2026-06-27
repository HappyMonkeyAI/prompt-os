#!/bin/sh
# Auto-start PromptOS TUI after Alpine boot (local.d hook).
# Running inside the target system.
# Note: do NOT use set -e here — a TUI startup failure should not kill the boot.

exec /usr/local/bin/promptos </dev/tty1 >/dev/tty1 2>&1
