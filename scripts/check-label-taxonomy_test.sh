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

# Pipe in a table field must fail (would break the generated Markdown table).
write_fixture_manifest
jq '.labels[0].applied_by = "bot|classifier"' \
	"${tmpdir}/.github/label-taxonomy.json" >"${tmpdir}/pipe.json"
mv "${tmpdir}/pipe.json" "${tmpdir}/.github/label-taxonomy.json"
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "pipe in applied_by fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_fails "write refused when manifest invalid" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

# Newline in a table field must fail (internal).
write_fixture_manifest
jq '.labels[0].category = "a\nb"' \
	"${tmpdir}/.github/label-taxonomy.json" >"${tmpdir}/nl.json"
mv "${tmpdir}/nl.json" "${tmpdir}/.github/label-taxonomy.json"
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "newline in category fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"

# Trailing newline must also fail — $(jq -r) would strip it in a shell-side check.
write_fixture_manifest
jq '.labels[0].category = ("routing" + "\n")' \
	"${tmpdir}/.github/label-taxonomy.json" >"${tmpdir}/trail.json"
mv "${tmpdir}/trail.json" "${tmpdir}/.github/label-taxonomy.json"
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "trailing newline in category fails check" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_fails "trailing newline refuses --write" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

# Duplicate / reversed markers must fail (and refuse --write).
write_fixture_manifest
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "duplicate begin marker fails" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_fails "duplicate begin marker refuses --write" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

write_fixture_manifest
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- END LABEL TAXONOMY TABLE -->
<!-- BEGIN LABEL TAXONOMY TABLE -->
EOF
assert_fails "reversed markers fail" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_fails "reversed markers refuse --write" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

# Indented markers are not exact lines — must fail (awk rewrite would miss them).
write_fixture_manifest
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

  <!-- BEGIN LABEL TAXONOMY TABLE -->
  <!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "indented markers fail" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_fails "indented markers refuse --write" \
	"${tmpdir}/scripts/check-label-taxonomy.sh" --write

# sync-sdlc-labels: malformed / empty labels must fail before any API call.
sync="${dir}/sync-sdlc-labels.sh"
cp "${sync}" "${tmpdir}/scripts/sync-sdlc-labels.sh"
chmod +x "${tmpdir}/scripts/sync-sdlc-labels.sh"

echo 'not-json' >"${tmpdir}/.github/label-taxonomy.json"
assert_fails "sync rejects malformed JSON" \
	"${tmpdir}/scripts/sync-sdlc-labels.sh" --dry-run

echo '{"labels":[]}' >"${tmpdir}/.github/label-taxonomy.json"
assert_fails "sync rejects empty labels array" \
	"${tmpdir}/scripts/sync-sdlc-labels.sh" --dry-run

echo '{}' >"${tmpdir}/.github/label-taxonomy.json"
assert_fails "sync rejects missing labels array" \
	"${tmpdir}/scripts/sync-sdlc-labels.sh" --dry-run

# Live repo check (real files). Do not call `make check-label-taxonomy` here —
# that target also runs this self-test and would recurse.
assert_ok "live repository check" "${check}"

if [[ "${fail}" -ne 0 ]]; then
	exit 1
fi
echo "check-label-taxonomy_test: all ok"
