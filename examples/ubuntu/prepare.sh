#!/usr/bin/env bash
set -euo pipefail
umask 077
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
image=${1:?Usage: bash prepare.sh /path/to/ubuntu-24.04-live-server-amd64.iso}
digest=${UBUNTU_SHA256:?Set UBUNTU_SHA256 from the official image checksum list}
[[ -f "$image" && "$digest" =~ ^[[:xdigit:]]{64}$ ]] || { printf 'Invalid image or SHA-256\n' >&2; exit 1; }
for tool in ssh-keygen sha256sum xorriso sed grep; do
    command -v "$tool" >/dev/null || { printf 'Missing tool: %s\n' "$tool" >&2; exit 1; }
done
actual=$(sha256sum "$image"); actual=${actual%% *}
[[ "${actual,,}" == "${digest,,}" ]] || { printf 'Ubuntu SHA-256 mismatch\n' >&2; exit 1; }
mkdir -m 700 "$root/generated"
out="$root/generated"
ssh-keygen -q -t ed25519 -N '' -f "$out/client-key"
pub=$(cat "$out/client-key.pub")
cat > "$out/autoinstall.yaml" <<EOF
autoinstall:
  version: 1
  locale: en_US.UTF-8
  keyboard: {layout: us}
  refresh-installer: {update: false}
  source: {id: ubuntu-server, search_drivers: false}
  apt:
    geoip: false
    fallback: offline-install
  storage:
    layout: {name: direct}
  identity:
    hostname: vmctl-ubuntu
    username: vmctl
    password: '!'
  ssh:
    install-server: true
    allow-pw: false
    authorized-keys:
      - $pub
  updates: security
  user-data:
    disable_root: true
    ssh_pwauth: false
  late-commands:
    - "printf 'vmctl ALL=(ALL) NOPASSWD:ALL\\n' > /target/etc/sudoers.d/90-vmctl"
    - chmod 440 /target/etc/sudoers.d/90-vmctl
    - touch /target/etc/vmctl-install-complete
  shutdown: poweroff
EOF
xorriso -osirrox on -indev "$image" -extract /boot/grub/grub.cfg "$out/grub.cfg" > "$out/extract.log" 2>&1
grep -Eq '^[[:space:]]*linux[[:space:]].*/casper/vmlinuz' "$out/grub.cfg" || {
    printf 'Expected Ubuntu Server GRUB kernel entry missing\n' >&2; exit 1;
}
sed -i -E '/^[[:space:]]*linux[[:space:]].*\/casper\/vmlinuz/s@/casper/vmlinuz@/casper/vmlinuz autoinstall@; s/^set timeout=.*/set timeout=0/' "$out/grub.cfg"
# Replay original BIOS/UEFI boot equipment instead of reconstructing boot images.
xorriso -indev "$image" -outdev "$out/ubuntu-autoinstall.iso" \
    -map "$out/autoinstall.yaml" /autoinstall.yaml \
    -map "$out/grub.cfg" /boot/grub/grub.cfg \
    -boot_image any replay > "$out/repack.log" 2>&1
printf 'Prepared private installation ISO: %s\n' "$out/ubuntu-autoinstall.iso"
