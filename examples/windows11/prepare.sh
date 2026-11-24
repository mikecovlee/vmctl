#!/usr/bin/env bash
set -euo pipefail
umask 077
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
shared_root=$root
profile=${VMCTL_WINDOWS_PROFILE:-windows11}
case "$profile" in
    windows11) ;;
    server2025) root="$root/../windows-server2025" ;;
    *) printf 'Unknown Windows profile\n' >&2; exit 1 ;;
esac
iso=${1:?Usage: bash prepare.sh /path/to/windows11.iso}
[[ -f "$iso" ]] || { printf 'ISO not found: %s\n' "$iso" >&2; exit 1; }
for tool in 7z openssl genisoimage ssh-keygen sha256sum; do
    command -v "$tool" >/dev/null || { printf 'Missing tool: %s\n' "$tool" >&2; exit 1; }
done
download_dir=
trap '[[ -z "$download_dir" ]] || rm -rf -- "$download_dir"' EXIT
if [[ -n "${OPENSSH_MSI:-}" ]]; then
    msi=$OPENSSH_MSI
    digest=${OPENSSH_SHA256:?Set OPENSSH_SHA256 to the local MSI trusted SHA-256}
else
    [[ -z "${OPENSSH_SHA256:-}" ]] || { printf 'OPENSSH_SHA256 requires OPENSSH_MSI\n' >&2; exit 1; }
    command -v curl >/dev/null || { printf 'Missing tool: curl (required for automatic download)\n' >&2; exit 1; }
    [[ ! -e "$root/generated" ]] || { printf 'generated already exists; refusing to replace it\n' >&2; exit 1; }
    digest=788d8643464a01eb63640a2fd746f9ed23d106b5a901a35f88dc8cd40e446ac9
    download_dir=$(mktemp -d)
    msi="$download_dir/OpenSSH-Win64.msi"
    printf 'Downloading pinned OpenSSH 10.4.0.0p1-Preview (not production-ready) on the host\n' >&2
    curl_options=(--fail --location --silent --show-error --proto '=https' --proto-redir '=https' --connect-timeout 20 --max-time 300 --retry 2)
    if ! curl "${curl_options[@]}" -o "$msi" \
        'https://github.com/PowerShell/Win32-OpenSSH/releases/download/v10.4.0.0p1-Preview/OpenSSH-Win64.msi'; then
        curl "${curl_options[@]}" -H 'Accept: application/octet-stream' \
            -o "$msi" 'https://api.github.com/repos/PowerShell/Win32-OpenSSH/releases/assets/645185820'
    fi
fi
[[ -f "$msi" && "$digest" =~ ^[[:xdigit:]]{64}$ ]] || { printf 'Invalid MSI or SHA-256\n' >&2; exit 1; }
actual=$(sha256sum "$msi"); actual=${actual%% *}
[[ "${actual,,}" == "${digest,,}" ]] || { printf 'OpenSSH SHA-256 mismatch\n' >&2; exit 1; }
[[ -z "${POST_INSTALL_SCRIPT:-}" || -f "$POST_INSTALL_SCRIPT" ]] || { printf 'Post-install script not found\n' >&2; exit 1; }
[[ -z "${SOFTWARE_DIR:-}" || -d "$SOFTWARE_DIR" ]] || { printf 'Software directory not found\n' >&2; exit 1; }
[[ -z "${SOFTWARE_DIR:-}" || ! -L "$SOFTWARE_DIR" ]] || { printf 'Software directory cannot be a symbolic link\n' >&2; exit 1; }
if [[ -n "${SOFTWARE_DIR:-}" ]] && [[ -n "$(find "$SOFTWARE_DIR" -mindepth 1 ! -type f ! -type d -print -quit)" ]]; then
    printf 'Software directory must contain only regular files and directories\n' >&2; exit 1
fi
template="$root/Autounattend.xml.tmpl"
[[ -f "$template" ]] || { printf 'Missing answer template\n' >&2; exit 1; }
# Atomic directory creation refuses to replace credentials or media in use.
mkdir -m 700 "$root/generated"
out="$root/generated"
ssh-keygen -q -t ed25519 -N '' -f "$out/client-key"
mkdir "$out/tree"
7z x -y "$iso" "-o$out/tree" > "$out/extract.log"
payload="$out/tree/sources/\$OEM\$/\$1/vmctl"
mkdir -p "$payload"
cp "$msi" "$payload/OpenSSH.msi"
printf '%s\n' "$digest" > "$payload/OpenSSH.sha256"
cp "$shared_root/bootstrap.ps1" "$out/client-key.pub" "$payload/"
[[ -z "${POST_INSTALL_SCRIPT:-}" ]] || cp "$POST_INSTALL_SCRIPT" "$payload/post-install.ps1"
[[ -z "${SOFTWARE_DIR:-}" ]] || cp -R "$SOFTWARE_DIR" "$payload/software"
if [[ "$profile" == windows11 ]]; then
[[ -f "$out/tree/efi/microsoft/boot/efisys_noprompt.bin" ]] || {
    printf 'This media lacks the expected no-prompt UEFI boot image\n' >&2
    exit 1
}
else
    [[ -f "$out/tree/boot/etfsboot.com" ]] || { printf 'Missing BIOS boot image\n' >&2; exit 1; }
fi
password="Vm9!$(openssl rand -hex 12)"
printf '%s\n' "$password" > "$out/password.txt"
# This Pro default setup key selects the edition; it does not activate Windows.
if [[ "$profile" == windows11 ]]; then
sed "s/__PASSWORD__/$password/g; s#<UserData>#<UserData><ProductKey><Key>VK7JG-NPHTM-C97JM-9MPGT-3V66T</Key><WillShowUI>Never</WillShowUI></ProductKey>#" \
    "$template" > "$out/tree/Autounattend.xml"
genisoimage -quiet -o "$out/windows-noprompt.iso" -iso-level 4 -udf \
    -allow-limited-size -J -joliet-long -relaxed-filenames -V VMCTL_WINDOWS \
    -e efi/microsoft/boot/efisys_noprompt.bin -no-emul-boot "$out/tree"
else
    index=${WINDOWS_IMAGE_INDEX:-2}
    [[ "$index" =~ ^[1-9][0-9]*$ ]] || { printf 'Invalid WINDOWS_IMAGE_INDEX\n' >&2; exit 1; }
    key=${WINDOWS_PRODUCT_KEY:-}
    [[ -z "$key" || "$key" =~ ^[A-Za-z0-9]{5}(-[A-Za-z0-9]{5}){4}$ ]] || { printf 'Invalid WINDOWS_PRODUCT_KEY format\n' >&2; exit 1; }
    product_key=
    if [[ -n "$key" ]]; then
        product_key="<ProductKey><Key>$key</Key><WillShowUI>Never</WillShowUI></ProductKey>"
    fi
    sed "s/__PASSWORD__/$password/g; s/__IMAGE_INDEX__/$index/g; s#__PRODUCT_KEY__#$product_key#g" "$template" > "$out/tree/Autounattend.xml"
    rm -f "$out/tree/boot/bootfix.bin"
    genisoimage -quiet -o "$out/windows-noprompt.iso" -iso-level 4 -udf \
        -allow-limited-size -J -joliet-long -relaxed-filenames -V VMCTL_SERVER \
        -b boot/etfsboot.com -no-emul-boot -boot-load-size 8 "$out/tree"
fi
printf 'Prepared private installation media: %s\n' "$out/windows-noprompt.iso"
