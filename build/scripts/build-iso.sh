set -euo pipefail
# build-iso.sh — build a bootable Alpine-based image containing the PromptOS TUI
# Intended for local execution on a Linux build host.
# This is a build script target, not a live test runner.

# ----- config -----
WORK_DIR="${WORK_DIR:-$(pwd)/.build-work}"
IMAGE_PATH="${IMAGE_PATH:-${WORK_DIR}/promptos.img}"
IMAGE_SIZE_GB="${IMAGE_SIZE_GB:-3}"
ALPISO_BASE="${ALPISO_BASE:-https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.3-x86_64.tar.gz}"
PROMPTOS_BIN_SRC="${PROMPTOS_BIN_SRC:-$(pwd)/../cmd/promptos}"
# -------------------

echo "==> WORK_DIR=${WORK_DIR}"
mkdir -p "${WORK_DIR}"

echo "==> download Alpine minirootfs"
curl -fSL "${ALPISO_BASE}" -o "${WORK_DIR}/minirootfs.tar.gz"

echo "==> create image"
fallocate -l "${IMAGE_SIZE_GB}G" "${IMAGE_PATH}" || truncate -s "${IMAGE_SIZE_GB}G" "${IMAGE_PATH}"

echo "==> format as ext4"
mkfs.ext4 -F "${IMAGE_PATH}"

echo "==> mount and install base"
mkdir -p "${WORK_DIR}/mnt"
mount "${IMAGE_PATH}" "${WORK_DIR}/mnt"
tar -xzf "${WORK_DIR}/minirootfs.tar.gz" -C "${WORK_DIR}/mnt"

echo "==> install busybox extras via apk.static (will runResolver outside chroot)"
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'apk.static add --no-cache bash util-linux'

echo "==> copy PromptOS binary"
mkdir -p "${WORK_DIR}/mnt/usr/local/bin"
if [ -d "${PROMPTOS_BIN_SRC}" ]; then
  go build -o "${WORK_DIR}/mnt/usr/local/bin/promptos" "${PROMPTOS_BIN_SRC}"
else
  echo "PromptOS binary source not built or path missing: ${PROMPTOS_BIN_SRC}"
  exit 1
fi

echo "==> install overlay"
mkdir -p "${WORK_DIR}/mnt/etc/local.d"
cp "$(pwd)/../build/alpine-live/init-promptos.sh" "${WORK_DIR}/mnt/etc/local.d/"
chmod +x "${WORK_DIR}/mnt/etc/local.d/init-promptos.sh"

echo "==> install boot config"
mkdir -p "${WORK_DIR}/mnt/boot/grub"
cp "$(pwd)/../build/alpine-live/grub.cfg" "${WORK_DIR}/mnt/boot/grub/grub.cfg"

echo "==> cleanup"
umount "${WORK_DIR}/mnt" || true

echo "==> image ready: ${IMAGE_PATH}"
echo "Next: install GRUB and create ISO with xorriso/genisoimage or boot the raw image in QEMU."
