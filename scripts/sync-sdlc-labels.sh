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

# Prefer GITHUB_TOKEN; otherwise require an authenticated gh session.
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
	if ! gh auth status >/dev/null 2>&1; then
		echo "error: set GITHUB_TOKEN or run gh auth login" >&2
		exit 1
	fi
fi

# Fail fast on over-long descriptions before any API call.
while IFS= read -r row; do
	name="$(jq -r '.name' <<<"${row}")"
	desc="$(jq -r '.description' <<<"${row}")"
	if ((${#desc} > 100)); then
		echo "error: label ${name}: description is ${#desc} characters (GitHub limit 100)" >&2
		exit 1
	fi
done < <(jq -c '.labels[]' "${MANIFEST}")

echo "-> Syncing labels to ${REPO} (dry-run=${DRY_RUN})"

current_json="$(mktemp)"
trap 'rm -f "${current_json}"' EXIT
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
done < <(jq -c '.labels[]' "${MANIFEST}")

echo "-> Done. created=${created} updated=${updated} unchanged=${unchanged} dry-run=${DRY_RUN}"
if [[ "${DRY_RUN}" -eq 1 ]]; then
	echo "-> Re-run without --dry-run to apply."
fi
