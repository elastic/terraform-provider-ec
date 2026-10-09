#!/usr/bin/env bash
#
# Creates or updates GitHub labels from .github/label-taxonomy.json.
# Idempotent for end state: lists current labels once and only writes when
# color or description differs. Never deletes labels. Never touches labels
# outside the manifest. Cannot rename (document a manual rename + manifest edit).
#
# Usage:
#   scripts/sync-sdlc-labels.sh [--dry-run] [--repo owner/repo]
#
# Auth: GITHUB_TOKEN, else `gh auth token` (same pattern as update-serverless-spec.sh).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="${REPO_ROOT}/.github/label-taxonomy.json"
REPO="${GITHUB_REPOSITORY:-elastic/terraform-provider-ec}"
DRY_RUN=0

while [[ $# -gt 0 ]]; do
	case "$1" in
	--dry-run)
		DRY_RUN=1
		shift
		;;
	--repo)
		REPO="${2:?--repo requires owner/repo}"
		shift 2
		;;
	*)
		echo "usage: $0 [--dry-run] [--repo owner/repo]" >&2
		exit 2
		;;
	esac
done

if ! command -v jq >/dev/null 2>&1; then
	echo "error: jq is required" >&2
	exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
	echo "error: gh is required" >&2
	exit 1
fi

if [[ ! -f "${MANIFEST}" ]]; then
	echo "error: missing manifest: ${MANIFEST}" >&2
	exit 1
fi

# Materialize and validate rows synchronously before auth or any API call.
# Process substitution would hide jq failures from `set -e` and let a missing
# `labels` array report a successful no-op (created=0 updated=0 unchanged=0).
rows_file="$(mktemp)"
current_json="$(mktemp)"
trap 'rm -f "${rows_file}" "${current_json}"' EXIT

if ! jq -e 'type == "object" and (.labels | type == "array")' "${MANIFEST}" >/dev/null; then
	echo "error: ${MANIFEST} must be a JSON object with a labels array" >&2
	exit 1
fi

if ! jq -c '.labels[]' "${MANIFEST}" >"${rows_file}"; then
	echo "error: failed to read labels from ${MANIFEST}" >&2
	exit 1
fi

if [[ ! -s "${rows_file}" ]]; then
	echo "error: ${MANIFEST} labels array is empty" >&2
	exit 1
fi

while IFS= read -r row; do
	name="$(jq -r '.name // empty' <<<"${row}")"
	desc="$(jq -r '.description // empty' <<<"${row}")"
	color="$(jq -r '.color // empty' <<<"${row}")"
	if [[ -z "${name}" ]]; then
		echo "error: label entry missing name" >&2
		exit 1
	fi
	if [[ -z "${color}" || ! "${color}" =~ ^[0-9A-Fa-f]{6}$ ]]; then
		echo "error: label ${name}: color must be 6 hex digits" >&2
		exit 1
	fi
	if [[ -z "${desc}" ]]; then
		echo "error: label ${name}: missing description" >&2
		exit 1
	fi
	if ((${#desc} > 100)); then
		echo "error: label ${name}: description is ${#desc} characters (GitHub limit 100)" >&2
		exit 1
	fi
done <"${rows_file}"

# Prefer GITHUB_TOKEN; otherwise require an authenticated gh session.
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
	if ! gh auth status >/dev/null 2>&1; then
		echo "error: set GITHUB_TOKEN or run gh auth login" >&2
		exit 1
	fi
fi

echo "-> Syncing labels to ${REPO} (dry-run=${DRY_RUN})"

gh label list --repo "${REPO}" --limit 1000 --json name,color,description >"${current_json}"

created=0
updated=0
unchanged=0

while IFS= read -r row; do
	name="$(jq -r '.name' <<<"${row}")"
	color="$(jq -r '.color' <<<"${row}" | tr 'A-F' 'a-f')"
	desc="$(jq -r '.description' <<<"${row}")"

	have="$(jq -r --arg n "${name}" '
		map(select(.name == $n)) | first |
		if . == null then empty else
			(.color | ascii_downcase) + "\t" + (.description // "")
		end
	' "${current_json}")"
	want="${color}"$'\t'"${desc}"

	if [[ -z "${have}" ]]; then
		echo "   + ${name}"
		created=$((created + 1))
	elif [[ "${have}" == "${want}" ]]; then
		echo "   = ${name}"
		unchanged=$((unchanged + 1))
		continue
	else
		echo "   ~ ${name} (update color/description)"
		updated=$((updated + 1))
	fi

	if [[ "${DRY_RUN}" -eq 0 ]]; then
		# --force updates color and description when the label already exists.
		gh label create "${name}" \
			--repo "${REPO}" \
			--color "${color}" \
			--description "${desc}" \
			--force >/dev/null
	fi
done <"${rows_file}"

echo "-> Done. created=${created} updated=${updated} unchanged=${unchanged} dry-run=${DRY_RUN}"
if [[ "${DRY_RUN}" -eq 1 ]]; then
	echo "-> Re-run without --dry-run to apply."
fi
