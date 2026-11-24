#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
find scripts runtime examples -type f -name '*.sh' -exec bash -c '
    for script do
        bash -n "$script" || exit 1
    done
' bash {} +
go run ./scripts/repository-check.go
sh scripts/dist.sh
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
tar -tzf dist/vmctl-source.tar.gz > "$scratch/manifest"
if grep -Ei '(^/|(^|/)\.\.(/|$)|/generated/|\.(iso|img|qcow2|pem|key)$|client-key|credentials\.json)' "$scratch/manifest"; then
    printf 'Unexpected private or generated archive entry\n' >&2
    exit 1
fi
tar -xzf dist/vmctl-source.tar.gz -C "$scratch"
(cd "$scratch" && go run ./scripts/repository-check.go && go build -buildvcs=false ./cmd/vmctl)
