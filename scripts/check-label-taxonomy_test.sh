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
# Use a recording gh stub so a missing real `gh` (or auth failure) cannot
# make these tests pass for the wrong reason.
sync="${dir}/sync-sdlc-labels.sh"
cp "${sync}" "${tmpdir}/scripts/sync-sdlc-labels.sh"
chmod +x "${tmpdir}/scripts/sync-sdlc-labels.sh"

stub_bin="${tmpdir}/stub-bin"
mkdir -p "${stub_bin}"
cat >"${stub_bin}/gh" <<'EOF'
#!/usr/bin/env bash
echo "gh-invoked:$*" >>"${GH_STUB_LOG:?GH_STUB_LOG unset}"
exit 99
EOF
chmod +x "${stub_bin}/gh"

run_sync_with_stub() {
	: >"${stub_bin}/gh.log"
	# GH_STUB_LOG must be in the env prefix so the stub grandchild can append.
	GH_STUB_LOG="${stub_bin}/gh.log" PATH="${stub_bin}:${PATH}" GITHUB_TOKEN=test-token \
		"${tmpdir}/scripts/sync-sdlc-labels.sh" --dry-run
}

assert_sync_fails_without_gh() {
	local name="$1"
	if run_sync_with_stub >/dev/null 2>&1; then
		echo "FAIL: ${name} (expected non-zero exit)" >&2
		fail=1
		return
	fi
	if [[ -s "${stub_bin}/gh.log" ]]; then
		echo "FAIL: ${name} (gh was invoked: $(cat "${stub_bin}/gh.log"))" >&2
		fail=1
		return
	fi
	echo "ok: ${name}"
}

# Control: a valid manifest must reach gh (proves the stub records when invoked).
write_fixture_manifest
run_sync_with_stub >/dev/null 2>&1 || true
if [[ ! -s "${stub_bin}/gh.log" ]]; then
	echo "FAIL: valid sync fixture did not invoke recording gh stub" >&2
	fail=1
else
	echo "ok: recording gh stub captures invocations"
fi

echo 'not-json' >"${tmpdir}/.github/label-taxonomy.json"
assert_sync_fails_without_gh "sync rejects malformed JSON before gh"

echo '{"labels":[]}' >"${tmpdir}/.github/label-taxonomy.json"
assert_sync_fails_without_gh "sync rejects empty labels array before gh"

echo '{}' >"${tmpdir}/.github/label-taxonomy.json"
assert_sync_fails_without_gh "sync rejects missing labels array before gh"

# Non-string fields must fail (jq -r would otherwise coerce them).
write_fixture_manifest
jq '.labels[0].name = 123' \
	"${tmpdir}/.github/label-taxonomy.json" >"${tmpdir}/num.json"
mv "${tmpdir}/num.json" "${tmpdir}/.github/label-taxonomy.json"
cat >"${tmpdir}/dev-docs/high-level/label-taxonomy.md" <<'EOF'
# fixture

<!-- BEGIN LABEL TAXONOMY TABLE -->
<!-- END LABEL TAXONOMY TABLE -->
EOF
assert_fails "non-string name fails check" \
	"${tmpdir}/scripts/check-label-taxonomy.sh"
assert_sync_fails_without_gh "sync rejects non-string name before gh"

# Live repo check (real files). Do not call `make check-label-taxonomy` here —
# that target also runs this self-test and would recurse.
assert_ok "live repository check" "${check}"

if [[ "${fail}" -ne 0 ]]; then
	exit 1
fi
echo "check-label-taxonomy_test: all ok"
