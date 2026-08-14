#!/usr/bin/env bash
#
# Downloads the vendored serverless API spec from the upstream repository.
#
# The upstream gitref is read from:
#   ec/internal/gen/serverless/serverless-project-api.source
#
# After running this script, regenerate dependent files with:
#   make gen
#
# Authentication:
#   - If GITHUB_TOKEN is set, it is used as a Bearer token.
#   - Otherwise, if the gh CLI is installed and authenticated, its token is used.
#   - Otherwise the script attempts an unauthenticated download (this will fail
#     for private repositories).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_DIR="${REPO_ROOT}/ec/internal/gen/serverless"
LOCK_FILE="${SPEC_DIR}/serverless-project-api.source"
SPEC_FILE="${SPEC_DIR}/serverless-project-api-dereferenced.yml"

if [[ ! -f "${LOCK_FILE}" ]]; then
	echo "error: lock file not found: ${LOCK_FILE}" >&2
	exit 1
fi

SOURCE_REPO="$(awk '/^source:/{print $2; exit}' "${LOCK_FILE}")"
SOURCE_PATH="$(awk '/^path:/{print $2; exit}' "${LOCK_FILE}")"
REF="$(awk '/^ref:/{print $2; exit}' "${LOCK_FILE}")"

if [[ -z "${SOURCE_REPO}" || -z "${SOURCE_PATH}" || -z "${REF}" ]]; then
	echo "error: could not parse source, path, or ref from ${LOCK_FILE}" >&2
	exit 1
fi

# Convert https://github.com/owner/repo to owner/repo.
REPO_SLUG="${SOURCE_REPO#https://github.com/}"
RAW_URL="https://raw.githubusercontent.com/${REPO_SLUG}/${REF}/${SOURCE_PATH}"

echo "-> Downloading serverless API spec from ${REPO_SLUG}@${REF}..."

AUTH_HEADER=""
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
	AUTH_HEADER="Authorization: Bearer ${GITHUB_TOKEN}"
elif command -v gh &>/dev/null && gh auth status &>/dev/null; then
	AUTH_HEADER="Authorization: Bearer $(gh auth token)"
fi

if [[ -n "${AUTH_HEADER}" ]]; then
	curl -fsSL -H "${AUTH_HEADER}" -o "${SPEC_FILE}" "${RAW_URL}"
else
	echo "warning: no GITHUB_TOKEN or gh auth found; attempting unauthenticated download" >&2
	curl -fsSL -o "${SPEC_FILE}" "${RAW_URL}"
fi

echo "-> Done."
echo "-> Run 'make gen' to regenerate dependent files."
