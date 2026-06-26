#!/usr/bin/env bash
set -euo pipefail
# vm-test.sh — run end-to-end VM test for PromptOS.
# Requires: qemu-system-x86_64, KVM enabled or TCG fallback, built image from build-iso.sh.

IMAGE_PATH="${IMAGE_PATH:-$(pwd)/.build-work/promptos.img}"
QEMU_ARGS="${QEMU_ARGS:-}"
MEM_MB="${MEM_MB:-1024}"
DISK_GB="${DISK_GB:-3}"

if [ ! -f "${IMAGE_PATH}" ]; then
  echo "Missing image: ${IMAGE_PATH}"
  exit 1
fi

echo "==> booting PromptOS image in QEMU"
qemu-system-x86_64 -m "${MEM_MB}" \
  -drive file="${IMAGE_PATH}",format=raw,if=virtio ${QEMU_ARGS} \
  -nographic -serial mon:stdio
