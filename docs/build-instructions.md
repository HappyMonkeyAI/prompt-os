# Building and Booting PromptOS

This document provides instructions on how to build, burn, and boot the PromptOS image.

---

## 1. Prerequisites

To build the bootable disk image, you need a Linux build host with the following packages installed:

* **coreutils / curl** (for downloading the Alpine minirootfs and apk static binaries)
* **util-linux** (for `losetup` and `sfdisk`)
* **e2fsprogs** (for `mkfs.ext4`)
* **grub-pc-bin** or **grub-bios** (for the host's GRUB installation support)
* **qemu-utils** (optional, for converting `.img` to `.vmdk` for VMware)

---

## 2. Build the Disk Image

Run the build script from the repository root:

```bash
sudo -E ./build/scripts/build-iso.sh
```

### What this script does:
1. **Creates a 3GB empty image file** at `.build-work/promptos.img`.
2. **Partitions it** using `sfdisk` with a legacy DOS partition table, allocating all space to a single bootable (active) partition.
3. **Mounts the partition** via a loopback device with partition scanning (`losetup -P`).
4. **Extracts the Alpine Linux minirootfs** into the partition.
5. **Configures DNS** inside the chroot by copying `/etc/resolv.conf`.
6. **Installs required packages**: `bash`, `util-linux`, `linux-virt` (VM-optimized kernel), `grub-bios` (bootloader), `openrc` (init system), and `eudev` (device manager).
7. **Compiles or copies the PromptOS binary** into `/usr/local/bin/promptos`.
8. **Enables services** (like `udev` and the `local` startup script that starts the PromptOS TUI on boot).
9. **Installs the GRUB bootloader** to the loop device's Master Boot Record (MBR) and writes `/boot/grub/grub.cfg`.
10. **Cleans up** all bind mounts and detaches loop devices.

The final output is a bootable raw disk image located at:
`./build-work/promptos.img`

---

## 3. How to Burn the Image to a USB Drive

The generated `.img` file is a **fully partitioned, bootable raw disk image**. You do not need a `.iso` file to burn it to a USB drive—you can flash the `.img` file directly.

### Option A: Using graphical tools (Windows, macOS, Linux)
Use a tool like [BalenaEtcher](https://etcher.balena.io/) or [Rufus](https://rufus.ie/):
1. Insert your USB flash drive.
2. Select `.build-work/promptos.img` as the source image.
3. Select your USB drive as the target.
4. Click **Flash!** or **Write**.

### Option B: Using the command line (Linux / macOS)
Identify your USB drive device node (e.g. `/dev/sdX` or `/dev/diskX`—**double check this to avoid wiping your host OS disk!**) and run:

```bash
sudo dd if=.build-work/promptos.img of=/dev/sdX bs=4M status=progress conv=fsync
```

---

## 4. How to Boot the Image in Virtual Machines

### QEMU (Command Line)
You can boot the image immediately in QEMU using the included helper script:
```bash
./build/scripts/vm-test.sh
```

### VMware (Workstation / Player)
1. Convert the raw image to VMware's VMDK format:
   ```bash
   qemu-img convert -f raw -O vmdk .build-work/promptos.img .build-work/promptos.vmdk
   ```
2. Copy `promptos.vmdk` to your target machine.
3. Create a new VM:
   * **Guest OS**: Linux (Other Linux 6.x kernel 64-bit).
   * **Firmware Type**: BIOS (Legacy).
   * **Hard Disk**: Add a new hard disk, select **Use an existing virtual disk**, and point it to the `promptos.vmdk` file.
4. Power on the VM.

### VirtualBox
1. Convert the raw image to VirtualBox's native VDI format:
   ```bash
   qemu-img convert -f raw -O vdi .build-work/promptos.img .build-work/promptos.vdi
   ```
2. Create a new VM:
   * **OS Type**: Linux / Other Linux (64-bit).
   * **Firmware**: Ensure EFI is disabled (use Legacy BIOS).
   * **Hard Disk**: Attach the `promptos.vdi` file.
3. Power on the VM.

---

## 5. (Optional) Creating an Optical ISO File
If you explicitly require a `.iso` file (e.g., for mounting in IPMI virtual media, legacy hypervisors, or burning to a CD/DVD), you can generate one using `grub-mkrescue` (which requires `xorriso` and `mtools` on the host):

```bash
# Create temporary directory for ISO contents
mkdir -p .build-work/iso/boot/grub
cp -r .build-work/mnt/boot/* .build-work/iso/boot/
# Copy the rest of the root filesystem as a squashfs or initramfs loop
# ...
grub-mkrescue -o .build-work/promptos.iso .build-work/iso
```
*(Note: Full live-ISO remastering typically packages the root filesystem into a Squashfs overlay. Direct `.img` flashing is the recommended pathway for physical installs).*
