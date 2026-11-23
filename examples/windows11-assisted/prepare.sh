#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
umask 077
command -v openssl >/dev/null
command -v genisoimage >/dev/null
mkdir generated
mkdir generated/answer
password="Vm9!$(openssl rand -hex 12)"
printf '%s\n' "$password" > generated/password.txt
sed "s/__PASSWORD__/$password/g" Autounattend.xml.tmpl > generated/answer/Autounattend.xml
genisoimage -quiet -J -r -V VMCTL_ANSWER -o generated/answer.iso generated/answer
printf 'Created generated/answer.iso; account vmctl, password in generated/password.txt\n'
