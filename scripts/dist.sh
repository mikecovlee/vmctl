#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
mkdir -p dist
manifest=$(mktemp)
trap 'rm -f "$manifest"' EXIT HUP INT TERM
printf '%s\0' go.mod go.sum Makefile README.md CONTRIBUTING.md SECURITY.md LICENSE NOTICE > "$manifest"
# Select source file types explicitly; never recursively archive user media or instance data.
find cmd internal -type f -name '*.go' -print0 >> "$manifest"
find runtime -type f \( -name Dockerfile -o -name '*.sh' \) -print0 >> "$manifest"
find examples -type f \( -name vm.json -o -name '*-manual.json' -o -name '*.tmpl' -o -name README.md \) -print0 >> "$manifest"
find docs -type f -name '*.md' -print0 >> "$manifest"
find scripts -type f -name '*.sh' -print0 >> "$manifest"
find .github/workflows -type f \( -name '*.yml' -o -name '*.yaml' \) -print0 >> "$manifest"
tar --null -T "$manifest" -czf dist/vmctl-source.tar.gz
printf '%s\n' 'dist/vmctl-source.tar.gz'
