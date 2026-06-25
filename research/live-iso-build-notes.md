# Live ISO Build Notes for PromptOS (Phase 0.2)

**Date:** 2026-06-25

**Primary Approach: Alpine Linux Live + Custom Init**

**Why Alpine**
- Extremely small base (~50-80 MB compressed live image possible).
- Simple APK package manager.
- Easy to add a single static Go binary.
- Well documented for custom live environments.

**High-Level Build Steps**
1. Use `alpine-make-vm-image` or manual `mkimage` scripts from Alpine.
2. Create a custom `init` or `local.d` script that launches the PromptOS TUI binary on boot.
3. Include minimal packages only: `apk add` for any runtime needs (none if fully static binary).
4. Package into squashfs + initramfs + bootloader (syslinux or GRUB).

**Alternative: archiso (if we later want Arch target closer)**
- More complex but gives native Arch environment.
- Larger base image.

**Size Target**
- Goal: < 250 MB final ISO.
- Go binary itself expected ~10-15 MB.

**Open Items**
- Exact commands and repo examples to be tested in a follow-up task.
- Need to verify how to embed the binary and auto-start the TUI without a full desktop.

**References**
- Alpine Linux wiki on custom live images
- Various minimal Alpine installer projects on GitHub (2025-2026)