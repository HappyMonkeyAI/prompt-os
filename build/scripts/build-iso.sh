#!/usr/bin/env bash
set -euo pipefail
# build-iso.sh — build a bootable Alpine-based image containing the PromptOS TUI
# Intended for local execution on a Linux build host.
# This is a build script target, not a live test runner.

# ----- config -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

WORK_DIR="${WORK_DIR:-${REPO_ROOT}/.build-work}"
IMAGE_PATH="${IMAGE_PATH:-${WORK_DIR}/promptos.img}"
VMDK_PATH="${VMDK_PATH:-${WORK_DIR}/promptos.vmdk}"
IMAGE_SIZE_GB="${IMAGE_SIZE_GB:-3}"
ALPISO_BASE="${ALPISO_BASE:-https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.3-x86_64.tar.gz}"
PROMPTOS_BIN_SRC="${PROMPTOS_BIN_SRC:-${REPO_ROOT}/cmd/promptos}"
# -------------------

LOOP_DEV=""
cleanup() {
  echo "==> cleaning up mounts and loop devices"
  umount "${WORK_DIR}/mnt/dev" 2>/dev/null || true
  umount "${WORK_DIR}/mnt/proc" 2>/dev/null || true
  umount "${WORK_DIR}/mnt/sys" 2>/dev/null || true
  umount "${WORK_DIR}/mnt" 2>/dev/null || true
  if [[ -n "${LOOP_DEV}" ]]; then
    losetup -d "${LOOP_DEV}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

echo "==> WORK_DIR=${WORK_DIR}"
mkdir -p "${WORK_DIR}"

# ----- CLEAN: remove stale image and VMDK to prevent dirty rebuild -----
echo "==> cleaning stale build artifacts"
if [[ -f "${IMAGE_PATH}" ]]; then
  echo "    removing stale image: ${IMAGE_PATH}"
  # Detach any loop devices still associated with the old image
  for dev in $(losetup -j "${IMAGE_PATH}" | cut -d: -f1); do
    echo "    detaching stale loop device: ${dev}"
    losetup -d "${dev}" 2>/dev/null || true
  done
  rm -f "${IMAGE_PATH}"
fi
if [[ -f "${VMDK_PATH}" ]]; then
  echo "    removing stale VMDK: ${VMDK_PATH}"
  rm -f "${VMDK_PATH}"
fi

echo "==> download Alpine minirootfs"
curl -fSL "${ALPISO_BASE}" -o "${WORK_DIR}/minirootfs.tar.gz"

echo "==> ensure clean mount state"
umount "${WORK_DIR}/mnt/dev" 2>/dev/null || true
umount "${WORK_DIR}/mnt/proc" 2>/dev/null || true
umount "${WORK_DIR}/mnt/sys" 2>/dev/null || true
umount "${WORK_DIR}/mnt" 2>/dev/null || true
for dev in $(losetup -j "${IMAGE_PATH}" | cut -d: -f1); do
  losetup -d "$dev" 2>/dev/null || true
done
mkdir -p "${WORK_DIR}/mnt"

echo "==> create image"
fallocate -l "${IMAGE_SIZE_GB}G" "${IMAGE_PATH}" || truncate -s "${IMAGE_SIZE_GB}G" "${IMAGE_PATH}"

echo "==> partitioning image (MBR, single bootable partition)"
echo '2048,,83,*' | sfdisk "${IMAGE_PATH}"

echo "==> mount and format partition"
LOOP_DEV="$(losetup -P --show -f "${IMAGE_PATH}")"
sleep 1
PART_DEV="${LOOP_DEV}p1"
if [[ ! -b "${PART_DEV}" && -b "${LOOP_DEV}1" ]]; then
  PART_DEV="${LOOP_DEV}1"
fi
if [[ ! -b "${PART_DEV}" ]]; then
  echo "Waiting for partition block device to appear..."
  for i in {1..5}; do
    if [[ -b "${LOOP_DEV}p1" ]]; then
      PART_DEV="${LOOP_DEV}p1"
      break
    elif [[ -b "${LOOP_DEV}1" ]]; then
      PART_DEV="${LOOP_DEV}1"
      break
    fi
    sleep 1
  done
fi
if [[ ! -b "${PART_DEV}" ]]; then
  echo "Error: partition device not found for ${LOOP_DEV}" >&2
  exit 1
fi

mkfs.ext4 -F "${PART_DEV}"

# Capture the filesystem UUID for hardware-agnostic GRUB root arg
FS_UUID="$(blkid -s UUID -o value "${PART_DEV}")"
echo "==> filesystem UUID: ${FS_UUID}"

mount "${PART_DEV}" "${WORK_DIR}/mnt"
tar -xzf "${WORK_DIR}/minirootfs.tar.gz" -C "${WORK_DIR}/mnt"

# Ensure we have DNS resolution inside the chroot for package installation
cp /etc/resolv.conf "${WORK_DIR}/mnt/etc/resolv.conf"

echo "==> install base packages via apk.static"
APKTOOLS_STATIC_APK_URL="https://dl-cdn.alpinelinux.org/alpine/v3.20/main/x86_64/apk-tools-static-2.14.4-r1.apk"
APKTOOLS_STATIC_APK="${WORK_DIR}/apk-tools-static.apk"
mkdir -p "${WORK_DIR}/mnt/usr/sbin" "${WORK_DIR}/mnt/sbin"
curl -fSL "${APKTOOLS_STATIC_APK_URL}" -o "${APKTOOLS_STATIC_APK}"
tar -xf "${APKTOOLS_STATIC_APK}" -C "${WORK_DIR}/mnt" sbin/apk.static
cp "${WORK_DIR}/mnt/sbin/apk.static" "${WORK_DIR}/mnt/usr/sbin/apk.static"
mkdir -p "${WORK_DIR}/mnt/etc/apk"
printf '%s\n' \
  "https://dl-cdn.alpinelinux.org/alpine/v3.20/main" \
  "https://dl-cdn.alpinelinux.org/alpine/v3.20/community" \
  > "${WORK_DIR}/mnt/etc/apk/repositories"
# open-vm-tools: better VMware hardware integration (SCSI/SATA detection, guest services)
chroot "${WORK_DIR}/mnt" /bin/sh -lc '/usr/sbin/apk.static add --no-cache bash util-linux linux-virt grub-bios openrc eudev open-vm-tools'

echo "==> retry base packages if transient fetch failed"
if ! chroot "${WORK_DIR}/mnt" /bin/bash -lc '[[ -x /usr/bin/bash && -x /bin/fdisk ]]'; then
  sleep 3
  chroot "${WORK_DIR}/mnt" /bin/sh -lc '/usr/sbin/apk.static add --no-cache bash util-linux linux-virt grub-bios openrc eudev open-vm-tools' || true
fi

echo "==> verify required tooling"
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'bash --version >/dev/null && fdisk --version >/dev/null' || {
  echo 'Required Alpine packages were not installed; check network or change packages' >&2
  exit 1
}

echo "==> copy PromptOS binary"
mkdir -p "${WORK_DIR}/mnt/usr/local/bin"

if [[ -f "${PROMPTOS_BIN_SRC}" ]]; then
  echo "Using pre-built binary: ${PROMPTOS_BIN_SRC}"
  cp "${PROMPTOS_BIN_SRC}" "${WORK_DIR}/mnt/usr/local/bin/promptos"
elif [[ -d "${PROMPTOS_BIN_SRC}" ]]; then
  GO_BIN="go"
  if ! command -v go >/dev/null 2>&1; then
    if [[ -x "/usr/local/go/bin/go" ]]; then
      GO_BIN="/usr/local/go/bin/go"
    elif [[ -n "${SUDO_USER:-}" && -x "/home/${SUDO_USER}/go/bin/go" ]]; then
      GO_BIN="/home/${SUDO_USER}/go/bin/go"
    elif [[ -f "${REPO_ROOT}/promptos" ]]; then
      echo "go not found, but found pre-built binary at repo root: ${REPO_ROOT}/promptos"
      cp "${REPO_ROOT}/promptos" "${WORK_DIR}/mnt/usr/local/bin/promptos"
      GO_BIN=""
    else
      echo "Error: 'go' command not found, and no pre-built binary found at ${REPO_ROOT}/promptos." >&2
      echo "Please build the binary first (go build -o promptos ./cmd/promptos) or make sure 'go' is in PATH." >&2
      exit 1
    fi
  fi

  if [[ -n "${GO_BIN}" ]]; then
    echo "Building PromptOS binary using ${GO_BIN} (static, musl-compatible)..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "${GO_BIN}" build \
      -ldflags="-extldflags=-static" \
      -o "${WORK_DIR}/mnt/usr/local/bin/promptos" \
      "${PROMPTOS_BIN_SRC}"
  fi
else
  if [[ -f "${REPO_ROOT}/promptos" ]]; then
    echo "Source not found or invalid, falling back to pre-built binary: ${REPO_ROOT}/promptos"
    cp "${REPO_ROOT}/promptos" "${WORK_DIR}/mnt/usr/local/bin/promptos"
  else
    echo "PromptOS binary source or pre-built binary missing: ${PROMPTOS_BIN_SRC}"
    exit 1
  fi
fi

echo "==> install boot config (UUID-based root for hardware-agnostic boot)"
mkdir -p "${WORK_DIR}/mnt/boot/grub"
# Substitute the actual filesystem UUID into grub.cfg at build time
sed "s|ROOT_UUID_PLACEHOLDER|${FS_UUID}|g" \
  "${REPO_ROOT}/build/alpine-live/grub.cfg" \
  > "${WORK_DIR}/mnt/boot/grub/grub.cfg"

echo "==> configuring auto-login and TUI launch"
# Set hostname
echo 'promptos' > "${WORK_DIR}/mnt/etc/hostname"

# Patch inittab: replace plain BusyBox getty on tty1 with agetty --autologin root
# inittab controls getty in Alpine minirootfs (not OpenRC agetty service files)
sed -i 's|tty1::respawn:/sbin/getty 38400 tty1|tty1::respawn:/sbin/agetty --autologin root --noclear 38400 tty1|' \
  "${WORK_DIR}/mnt/etc/inittab"

# Create root .profile: exec PromptOS TUI automatically on login
# If the TUI exits or crashes, the user gets a recovery shell
# Only run if we are on the physical console tty1 (prevents running during build chroot)
mkdir -p "${WORK_DIR}/mnt/root"
printf '%s\n' \
  '#!/bin/sh' \
  '# PromptOS auto-start - runs on every root login (via agetty auto-login)' \
  'if [ "$(tty)" = "/dev/tty1" ]; then' \
  '    exec /usr/local/bin/promptos' \
  'fi' \
  > "${WORK_DIR}/mnt/root/.profile"
chmod 700 "${WORK_DIR}/mnt/root/.profile"

echo "==> configuring services and grub bootloader"
# Enable services in Alpine
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'rc-update add udev sysinit' || true
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'rc-update add udev-trigger sysinit' || true
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'rc-update add agetty.tty1 default' || true
chroot "${WORK_DIR}/mnt" /bin/sh -lc 'rc-update add vmware-tools default' 2>/dev/null || true

# Mount API filesystems to run grub-install
mount --bind /dev "${WORK_DIR}/mnt/dev"
mount --bind /proc "${WORK_DIR}/mnt/proc"
mount --bind /sys "${WORK_DIR}/mnt/sys"

# Install GRUB in the loop device MBR (legacy BIOS)
chroot "${WORK_DIR}/mnt" grub-install --target=i386-pc --force "${LOOP_DEV}"

# Unmount API filesystems before cleanup
umount "${WORK_DIR}/mnt/dev" 2>/dev/null || true
umount "${WORK_DIR}/mnt/proc" 2>/dev/null || true
umount "${WORK_DIR}/mnt/sys" 2>/dev/null || true

echo "==> cleanup"
# Cleanup is handled automatically by the EXIT trap.

echo "==> image ready: ${IMAGE_PATH}"

echo "==> converting raw image to VMDK for VMware"
if command -v qemu-img >/dev/null 2>&1; then
  qemu-img convert -f raw -O vmdk "${IMAGE_PATH}" "${VMDK_PATH}"
  echo "==> VMDK ready: ${VMDK_PATH}"
else
  echo "WARNING: qemu-img not found; skipping VMDK conversion."
  echo "         Install with: sudo apt install qemu-utils"
  echo "         Then run: qemu-img convert -f raw -O vmdk ${IMAGE_PATH} ${VMDK_PATH}"
fi

echo ""
echo "==> BUILD COMPLETE"
echo "    Raw image : ${IMAGE_PATH}"
echo "    VMDK      : ${VMDK_PATH}"
echo ""
echo "VMware import instructions:"
echo "  1. Guest OS: Linux → Other Linux 5.x or later kernel 64-bit"
echo "  2. Disk controller: SATA (not NVMe/SCSI)"
echo "  3. Firmware: BIOS (not UEFI)"
echo "  4. Import ${VMDK_PATH} as existing virtual disk"
