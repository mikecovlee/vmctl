#!/usr/bin/env bash
set -euo pipefail
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
export VMCTL_WINDOWS_PROFILE=server2025
exec bash "$root/../windows11/prepare.sh" "$@"
