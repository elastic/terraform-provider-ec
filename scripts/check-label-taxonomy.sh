#!/usr/bin/env bash
#
# Validates .github/label-taxonomy.json and that the doc table matches it.
# No network. Requires jq.
#
# Usage:
#   scripts/check-label-taxonomy.sh           # validate only
#   scripts/check-label-taxonomy.sh --write   # regenerate the doc table from the manifest

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="${REPO_ROOT}/.github/label-taxonomy.json"
DOC="${REPO_ROOT}/dev-docs/high-level/label-taxonomy.md"
BEGIN_MARKER='<!-- BEGIN LABEL TAXONOMY TABLE -->'
END_MARKER='<!-- END LABEL TAXONOMY TABLE -->'
WRITE=0

if [[ "${1:-}" == "--write" ]]; then
	WRITE=1
elif [[ "${1:-}" != "" ]]; then
	echo "usage: $0 [--write]" >&2
	exit 2
fi

if ! command -v jq >/dev/null 2>&1; then
	echo "error: jq is required" >&2
	exit 1
fi

if [[ ! -f "${MANIFEST}" ]]; then
	echo "error: missing manifest: ${MANIFEST}" >&2
	exit 1
fi

if [[ ! -f "${DOC}" ]]; then
	echo "error: missing doc: ${DOC}" >&2
	exit 1
fi

fail=0
err() {
	echo "error: $*" >&2
	fail=1
}

if ! jq -e . "${MANIFEST}" >/dev/null 2>&1; then
	echo "error: ${MANIFEST} is not valid JSON" >&2
	exit 1
fi

# Fields interpolated into the Markdown table must not break cells.
# Test in jq (not via $(...)): command substitution strips trailing newlines and
# would let values like "routing\n" pass a shell-side check.
bad_md="$(jq -r '
  .labels[]
  | . as $l
  | ["name","category","applied_by","lifecycle","description"][]
  | . as $f
  | ($l[$f] // "")
  | select(type == "string" and test("[\n|]"))
  | "\($l.name // "(missing name)"): \($f) must not contain newlines or \"|\""
' "${MANIFEST}")"
if [[ -n "${bad_md}" ]]; then
	while IFS= read -r msg; do
		[[ -z "${msg}" ]] && continue
		err "label ${msg}"
	done <<<"${bad_md}"
fi

# Reject non-string fields before jq -r coercion (numbers/bools would otherwise pass).
bad_types="$(jq -r '
  .labels
  | to_entries[]
  | . as $e
  | ["name","color","description","category","applied_by","lifecycle"][]
  | . as $f
  | select(($e.value[$f] | type) != "string")
  | "labels[\($e.key)].\($f) must be a string (got \(($e.value[$f] | type)))"
' "${MANIFEST}")"
if [[ -n "${bad_types}" ]]; then
	while IFS= read -r msg; do
		[[ -z "${msg}" ]] && continue
		err "${msg}"
	done <<<"${bad_types}"
fi

# Required field presence and description length (GitHub limit is 100 characters).
while IFS= read -r row; do
	name="$(jq -r '.name // empty' <<<"${row}")"
	color="$(jq -r '.color // empty' <<<"${row}")"
	desc="$(jq -r '.description // empty' <<<"${row}")"
	category="$(jq -r '.category // empty' <<<"${row}")"
	applied_by="$(jq -r '.applied_by // empty' <<<"${row}")"
	lifecycle="$(jq -r '.lifecycle // empty' <<<"${row}")"

	if [[ -z "${name}" ]]; then
		err "label entry missing name"
		continue
	fi
	if [[ -z "${color}" || ! "${color}" =~ ^[0-9A-Fa-f]{6}$ ]]; then
		err "label ${name}: color must be 6 hex digits"
	fi
	if [[ -z "${category}" ]]; then
		err "label ${name}: missing category"
	fi
	if [[ -z "${applied_by}" ]]; then
		err "label ${name}: missing applied_by"
	fi
	if [[ -z "${lifecycle}" ]]; then
		err "label ${name}: missing lifecycle"
	fi
	if [[ -z "${desc}" ]]; then
		err "label ${name}: missing description"
	elif ((${#desc} > 100)); then
		err "label ${name}: description is ${#desc} characters (GitHub limit 100)"
	fi
done < <(jq -c '.labels[]' "${MANIFEST}")

# required_names must all appear in labels[].name; labels may be a superset.
missing="$(jq -r '
  (.required_names - [.labels[].name])[]
' "${MANIFEST}")"
if [[ -n "${missing}" ]]; then
	while IFS= read -r n; do
		[[ -z "${n}" ]] && continue
		err "required name missing from labels[]: ${n}"
	done <<<"${missing}"
fi

required_count="$(jq -r '.required_names | length' "${MANIFEST}")"

# Detect duplicate names in labels[].
dupes="$(jq -r '.labels | group_by(.name)[] | select(length > 1) | .[0].name' "${MANIFEST}")"
if [[ -n "${dupes}" ]]; then
	while IFS= read -r d; do
		[[ -z "${d}" ]] && continue
		err "duplicate label name: ${d}"
	done <<<"${dupes}"
fi

render_table() {
	echo '| Name | Category | Applied by | Lifecycle | Color | Description |'
	echo '| --- | --- | --- | --- | --- | --- |'
	jq -r '.labels[] | "| `\(.name)` | \(.category) | \(.applied_by) | \(.lifecycle) | `#\(.color)` | \(.description) |"' "${MANIFEST}"
}

# Require exactly one begin marker followed by exactly one end marker.
# Match whole lines (-x) so the count agrees with awk's `$0 == marker` rewrite;
# substring matches would accept indented markers that --write silently skips.
# A mere "each marker appears somewhere" check also lets reversed/duplicate
# markers through; --write would then inject multiple tables or truncate the doc.
begin_count="$(grep -cxF "${BEGIN_MARKER}" "${DOC}" || true)"
end_count="$(grep -cxF "${END_MARKER}" "${DOC}" || true)"
if [[ "${begin_count}" -ne 1 || "${end_count}" -ne 1 ]]; then
	err "doc ${DOC} must contain exactly one begin and one end marker (begin=${begin_count} end=${end_count})"
else
	begin_line="$(grep -nxF "${BEGIN_MARKER}" "${DOC}" | cut -d: -f1)"
	end_line="$(grep -nxF "${END_MARKER}" "${DOC}" | cut -d: -f1)"
	if [[ "${begin_line}" -ge "${end_line}" ]]; then
		err "doc ${DOC}: begin marker (line ${begin_line}) must precede end marker (line ${end_line})"
	fi
fi

expected="$(render_table)"
# Extract current table (exclusive of markers).
actual="$(awk -v b="${BEGIN_MARKER}" -v e="${END_MARKER}" '
	$0 == b {grab=1; next}
	$0 == e {grab=0}
	grab {print}
' "${DOC}")"

if [[ "${WRITE}" -eq 1 ]]; then
	if [[ "${fail}" -ne 0 ]]; then
		echo "error: refusing --write because validation failed" >&2
		exit 1
	fi
	tmp="$(mktemp)"
	table_file="$(mktemp)"
	printf '%s\n' "${expected}" >"${table_file}"
	awk -v b="${BEGIN_MARKER}" -v e="${END_MARKER}" -v tf="${table_file}" '
		$0 == b {print; while ((getline line < tf) > 0) print line; close(tf); skip=1; next}
		$0 == e {skip=0; print; next}
		!skip {print}
	' "${DOC}" >"${tmp}"
	# Preserve the doc's mode bits (mktemp files are often 0600).
	# Prefer GNU `stat -c` first so Linux CI/devs get a real octal mode; macOS
	# rejects -c and falls through to BSD `stat -f '%Lp'`.
	mode="$(stat -c '%a' "${DOC}" 2>/dev/null || stat -f '%Lp' "${DOC}" 2>/dev/null || true)"
	if [[ -n "${mode}" && "${mode}" =~ ^[0-7]{3,4}$ ]]; then
		chmod "${mode}" "${tmp}" || true
	fi
	mv "${tmp}" "${DOC}"
	rm -f "${table_file}"
	echo "-> Wrote label table into ${DOC}"
elif [[ "${actual}" != "${expected}" ]]; then
	err "doc table does not match ${MANIFEST}; run: scripts/check-label-taxonomy.sh --write"
	echo "---- expected ----" >&2
	printf '%s\n' "${expected}" >&2
	echo "---- actual ----" >&2
	printf '%s\n' "${actual}" >&2
fi

if [[ "${fail}" -ne 0 ]]; then
	exit 1
fi

echo "check-label-taxonomy: ok (${required_count} required names present)"
