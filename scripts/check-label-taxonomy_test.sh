#!/usr/bin/env bash
# Self-check for check-label-taxonomy.sh. No network. No GitHub writes.

set -euo pipefail

dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
check="${dir}/check-label-taxonomy.sh"
fail=0

assert_ok() {
	local name="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo "ok: ${name}"
	else
		echo "FAIL: ${name} (expected success)" >&2
		fail=1
	fi
}

assert_fails() {
	local name="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo "FAIL: ${name} (expected non-zero exit)" >&2
		fail=1
	else
		echo "ok: ${name}"
	fi
}

write_fixture_manifest() {
	cat >"${tmpdir}/.github/label-taxonomy.json" <<'EOF'
{
  "version": 1,
  "required_names": ["needs-research", "triaged"],
  "labels": [
    {
      "name": "needs-research",
      "color": "0d9488",
      "description": "Bot routing: feature needs research.",
      "category": "routing",
      "applied_by": "issue-classifier (bot)",
      "lifecycle": "Exactly one needs-* with triaged."
    },
    {
      "name": "triaged",
      "color": "134e4a",
      "description": "Sticky marker: already routed.",
      "category": "routing",
      "applied_by": "issue-classifier (bot)",
      "lifecycle": "Sticky."
    }
  ]
}
EOF
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

mkdir -p "${tmpdir}/.github" "${tmpdir}/dev-docs/high-level" "${tmpdir}/scripts"
cp "${check}" "${tmpdir}/scripts/check-label-taxonomy.sh"
chmod +x "${tmpdir}/scripts/check-label-taxonomy.sh"

write_fixture_manifest
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF

assert_ok "write regenerates table" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

assert_ok "validate after write" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"

# Break the table body while keeping markers.
{
	echo '# fixture'
	echo
	echo '<!-- BEGIN LABEL TAXONOMY TABLE -->'
	echo '| bogused |'
	echo '<!-- END LABEL TAXONOMY TABLE -->'
} >"${tmpdir}/dev-docs/high-level/label-taxonomy.md"

assert_fails "mismatched table fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"

# Over-long description (character count, not bytes).
write_fixture_manifest
jq --arg d "$(printf 'x%.0s' {1..101})" '.labels[0].description = $d' \
	"${tmpdir}/.github/label-taxonomy.json" >"${tmpdir}/long.json"
mv "${tmpdir}/long.json" "${tmpdir}/.github/label-taxonomy.json"
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "description over 100 chars fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"

# Missing required name.
cat >"${tmpdir}/.github/label-taxonomy.json" <<'EOF'
{
  "version": 1,
  "required_names": ["needs-research", "triaged", "phase-coding"],
  "labels": [
    {
      "name": "needs-research",
      "color": "0d9488",
      "description": "ok",
      "category": "routing",
      "applied_by": "bot",
      "lifecycle": "x"
    },
    {
      "name": "triaged",
      "color": "134e4a",
      "description": "ok",
      "category": "routing",
      "applied_by": "bot",
      "lifecycle": "x"
    }
  ]
}
EOF
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "missing required name fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"

# Live repo check (real files).
assert_ok "live repository check" "${check}"

if [[ "${fail}" -ne 0 ]]; then
	exit 1
fi
echo "check-label-taxonomy_test: all ok"
