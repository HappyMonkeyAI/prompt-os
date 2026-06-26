# PromptOS Build Artifacts

This folder contains build inputs for Phase 5: bootable installer artifacts.

## Task 5.1 — Minimal Live ISO Build
**Status:** Recipe and script target implemented (2026-06-26). Actual ISO build and boot test remain as execution tasks.

**Build host assumptions**
- Linux x86_64 host with `sudo`, `qemu-img`, `curl`, `git`, `bash`
- Network egress to download Alpine minirootfs + packages
- Enough disk for a ~2–4 GB image file

**Quick start**
```bash
cd ~/projects/prompt-os/build/scripts
bash build-iso.sh
```
This will:
- create a raw disk image
- install Alpine into it via `apk.static`
- install the PromptOS static binary
- add an init script to auto-start the TUI on first boot
- add GRUB/Syslinux boot config suitable for ISO conversion or direct disk boot

**Artifacts**
- `build/alpine-live/` — overlay files, GRUB config, init script
- `build/README.md` — this file
- `build/scripts/build-iso.sh` — main build script
