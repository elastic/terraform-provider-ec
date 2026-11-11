#!/bin/bash
set -euo pipefail

# Text-only PRs still need a green Buildkite status: master requires
# buildkite/terraform-provider-ec-acceptance, and skip_ci_on_only_changed would
# leave that context missing (blocking merge). Exit 0 here instead of running
# the paid suite.
#
# openspec/ and .changelog/ are in the skip set alongside the docs trees: neither
# can change provider behaviour, so neither can break an acceptance test. This also
# keeps the change-factory workflow honest — it opens spec-only proposal PRs, and
# without the skip every proposal would spend real money against the live Elastic
# Cloud API. Note the predicate is "EVERY changed file matches", so an
# implementation PR that also touches openspec/ still runs the suite.
docs_only_pr() {
  [[ "${BUILDKITE_PULL_REQUEST:-false}" == "false" ]] && return 1
  local base="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-master}"
  # Fail-safe: if we cannot see the base, run the suite rather than skip.
  git fetch --depth=50 origin "$base" || return 1
  local files
  files=$(git diff --name-only "origin/${base}...HEAD") || return 1
  [[ -n "$files" ]] || return 1
  # Here-string: grep -q would SIGPIPE echo under pipefail and invert to a false skip.
  ! grep -qvE '^(docs/|dev-docs/|openspec/|\.changelog/)' <<< "$files"
}

if docs_only_pr; then
  echo "--- Skip acceptance tests (docs/dev-docs/openspec/changelog only)"
  # pre-exit still runs; skip sweep so a leftover cleanup failure cannot fail the required check.
  touch .buildkite/.skip-acceptance-sweep
  exit 0
fi

echo "--- Download dependencies"
make vendor

# terraform-plugin-testing's resource.Test looks Terraform up on PATH and, finding none on the
# golang image, downloads the latest CLI from checkpoint-api.hashicorp.com into a fresh temp dir
# in EVERY test (~60 times per run). Any hiccup there -- chmod on a half-written file, "text file
# busy", an HTTP 5xx -- calls os.Exit(1) and kills the whole package with a bare FAIL and no
# "--- FAIL" line. Install one pinned binary up front instead: with TF_ACC_TERRAFORM_PATH set,
# hc-install resolves that exact path and never downloads. Both pins are bumped by Renovate
# (see customManagers in renovate.json).
# renovate: datasource=github-releases depName=hashicorp/terraform
TERRAFORM_VERSION="1.17.1"
# renovate: datasource=go depName=github.com/hashicorp/hc-install
HC_INSTALL_VERSION="v0.9.4"

echo "--- Install Terraform ${TERRAFORM_VERSION}"
terraform_dir="$(mktemp -d)"
scripts/retry.sh 3 go run "github.com/hashicorp/hc-install/cmd/hc-install@${HC_INSTALL_VERSION}" \
  install -version "${TERRAFORM_VERSION}" -path "${terraform_dir}" terraform
export TF_ACC_TERRAFORM_PATH="${terraform_dir}/terraform"
"${TF_ACC_TERRAFORM_PATH}" version

echo "--- Run acceptance tests"
EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc
