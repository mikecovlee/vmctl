#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
version=${1:?Usage: sh scripts/release.sh vMAJOR.MINOR.PATCH[-rc.N]}
if ! printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$'; then
    printf 'Invalid release version\n' >&2
    exit 1
fi
: "${VMCTL_COMMIT:?Set VMCTL_COMMIT to the source commit identifier}"
if ! printf '%s\n' "$VMCTL_COMMIT" | grep -Eq '^[0-9a-f]{40}$'; then
    printf 'VMCTL_COMMIT must be a full lowercase Git commit hash\n' >&2
    exit 1
fi
date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
mkdir -p dist
out="dist/vmctl-$version-linux-amd64"
archive="$out.tar.gz"
if [ -e "$archive" ] || [ -e "$archive.sha256" ]; then
    printf 'Release artifact already exists: %s\n' "$archive" >&2
    exit 1
fi
mkdir "$out"
trap 'rm -rf "$out"' EXIT HUP INT TERM
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -trimpath \
    -ldflags "-X vmctl/internal/cli.version=$version -X vmctl/internal/cli.commit=$VMCTL_COMMIT -X vmctl/internal/cli.buildDate=$date" \
    -o "$out/vmctl" ./cmd/vmctl
cp LICENSE NOTICE "$out/"
cp docs/v1.0.md "$out/RELEASE-GOALS.md"
tar -C dist -czf "$archive" "$(basename "$out")"
(cd dist && sha256sum "$(basename "$archive")" > "$(basename "$archive").sha256")
printf '%s\n' "$archive" "$archive.sha256"
