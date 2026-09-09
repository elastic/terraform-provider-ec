#!/bin/bash
set -euo pipefail

# Docs-only PRs still need a green Buildkite status: master requires
# buildkite/terraform-provider-ec-acceptance, and skip_ci_on_only_changed would
# leave that context missing (blocking merge). Exit 0 here instead of running
# the paid suite.
docs_only_pr() {
  [[ "${BUILDKITE_PULL_REQUEST:-false}" == "false" ]] && return 1
  local base="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-master}"
  # Fail-safe: if we cannot see the base, run the suite rather than skip.
  git fetch --depth=50 origin "$base" || return 1
  local files
  files=$(git diff --name-only "origin/${base}...HEAD") || return 1
  [[ -n "$files" ]] || return 1
  ! echo "$files" | grep -qvE '^(docs/|dev-docs/)'
}

if docs_only_pr; then
  echo "--- Skip acceptance tests (docs/dev-docs only)"
  # pre-exit still runs; skip sweep so a leftover cleanup failure cannot fail the required check.
  touch .buildkite/.skip-acceptance-sweep
  exit 0
fi

echo "--- Download dependencies"
make vendor

echo "--- Run acceptance tests"
EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc
