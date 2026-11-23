#!/usr/bin/env bash
set -euo pipefail
umask 077
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
iso=${1:?Usage: bash prepare.sh /path/to/windows11.iso}
[[ -f "$iso" ]] || { printf 'ISO not found: %s\n' "$iso" >&2; exit 1; }
for tool in 7z openssl genisoimage; do
    command -v "$tool" >/dev/null || { printf 'Missing tool: %s\n' "$tool" >&2; exit 1; }
done
template="$root/Autounattend.xml.tmpl"
[[ -f "$template" ]] || { printf 'Missing answer template\n' >&2; exit 1; }
# Atomic directory creation refuses to replace credentials or media in use.
mkdir -m 700 "$root/generated"
out="$root/generated"
mkdir "$out/tree"
7z x -y "$iso" "-o$out/tree" > "$out/extract.log"
[[ -f "$out/tree/efi/microsoft/boot/efisys_noprompt.bin" ]] || {
    printf 'This media lacks the expected no-prompt UEFI boot image\n' >&2
    exit 1
}
password="Vm9!$(openssl rand -hex 12)"
printf '%s\n' "$password" > "$out/password.txt"
# This Pro default setup key selects the edition; it does not activate Windows.
sed "s/__PASSWORD__/$password/g; s#<UserData>#<UserData><ProductKey><Key>VK7JG-NPHTM-C97JM-9MPGT-3V66T</Key><WillShowUI>Never</WillShowUI></ProductKey>#" \
    "$template" > "$out/tree/Autounattend.xml"
genisoimage -quiet -o "$out/windows-noprompt.iso" -iso-level 4 -udf \
    -allow-limited-size -J -joliet-long -relaxed-filenames -V VMCTL_WINDOWS \
    -e efi/microsoft/boot/efisys_noprompt.bin -no-emul-boot "$out/tree"
printf 'Prepared private installation media: %s\n' "$out/windows-noprompt.iso"
