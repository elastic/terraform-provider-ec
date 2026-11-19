#!/usr/bin/env bash

#
# This script takes in a single parameter with the current provider version
# and updates all the files that contain a declaration of the 'ec' provider
# to use the current version.
#

set -euo pipefail

VERSION="${1:?usage: $0 <version>}"
if [[ ! "${VERSION}" =~ ^[0-9A-Za-z.+-]+$ ]]; then
    echo "error: invalid version '${VERSION}'" >&2
    exit 1
fi

echo "-> Updating the version field on references to the previous 'ec' provider declaration"

declare -a UPDATE_FILES=("README.md")
while IFS= read -r f; do
    UPDATE_FILES+=("$f")
done < <(grep -RlE '^[[:space:]]*ec[[:space:]]*=[[:space:]]*\{' examples)

for f in "${UPDATE_FILES[@]}"; do
    # Only rewrite the `version` attribute inside the `ec = { ... }` block, so
    # other providers declared in the same file (e.g. hashicorp/aws) keep their
    # own version constraints regardless of attribute order.
    # Instead of in-place editing, whose flags differ on macOS and Linux, write
    # to a temporary file and move it over the original.
    tmp="$(mktemp)"
    if ! awk -v version="${VERSION}" '
        /^[[:space:]]*ec[[:space:]]*=[[:space:]]*\{/ { in_ec = 1 }
        in_ec && /^[[:space:]]*version[[:space:]]*=/ {
            sub(/version[[:space:]]*=[[:space:]]*"[^"]*"/, "version = \"" version "\"")
            updated++
        }
        in_ec && /}/ { in_ec = 0 }
        { print }
        END { if (updated != 1) exit 1 }
    ' "$f" > "$tmp"; then
        rm -f "$tmp"
        echo "error: expected exactly one ec provider version in $f" >&2
        exit 1
    fi
    # Copy rather than mv so the original file keeps its permissions.
    cat "$tmp" > "$f"
    rm -f "$tmp"
done

echo "-> Updated all the ec provider declarations"
