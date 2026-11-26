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
# The rule itself is docs_only_pr in acceptance-test-selection.sh, next to the other decisions about
# what a build runs, so it can be tested without running anything paid.
# shellcheck source=.buildkite/acceptance-test-selection.sh
source "$(dirname "${BASH_SOURCE[0]}")/acceptance-test-selection.sh"

# First, before anything from the branch is compiled: a reproducer-factory branch runs nothing unless
# a person asked for this build. The sweep in the pre-exit hook compiles the same package, so it is
# skipped too. The step fails rather than passes: a green required check must not appear on a pull
# request whose test nobody has run.
if reproducer_factory_build_refused; then
  echo "+++ Not running acceptance tests: nobody has asked for this build"
  echo "This is a reproducer-factory branch. Its acceptance test was written by an automated agent"
  echo "and runs with a live Elastic Cloud credential, so nothing runs until a person asks."
  echo "Read the test file first, then comment 'buildkite test this' on the pull request."
  echo "A push to the branch by a person with standing, including 'Update branch', can also start"
  echo "the run, so read the test before doing either."
  touch .buildkite/.skip-acceptance-sweep
  exit 1
fi

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
TERRAFORM_VERSION="1.17.2"
# renovate: datasource=go depName=github.com/hashicorp/hc-install
HC_INSTALL_VERSION="v0.9.4"

echo "--- Install Terraform ${TERRAFORM_VERSION}"
terraform_dir="$(mktemp -d)"
scripts/retry.sh 3 go run "github.com/hashicorp/hc-install/cmd/hc-install@${HC_INSTALL_VERSION}" \
  install -version "${TERRAFORM_VERSION}" -path "${terraform_dir}" terraform
export TF_ACC_TERRAFORM_PATH="${terraform_dir}/terraform"
"${TF_ACC_TERRAFORM_PATH}" version

# Most builds run the whole suite. A factory branch may run one named test instead; the rules, and
# why every doubt falls back to the whole suite, are in acceptance-test-selection.sh.
selected_test_name="$(acceptance_test_name)"
if [[ -n "$selected_test_name" ]]; then
  echo "--- Run acceptance tests (only ${selected_test_name})"
  # build/Makefile.test reads TEST_NAME from the environment (TEST_NAME ?= TestAcc).
  export TEST_NAME="$selected_test_name"
  test_log="$(mktemp)"
  # pipefail: a failing test run fails the pipeline, and errexit stops here.
  EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc 2>&1 | tee "$test_log"
  if ! narrowed_run_passed_test "$selected_test_name" "$test_log"; then
    echo "^^^ +++"
    echo "The run exited 0 but ${selected_test_name} did not run and pass (skipped, excluded by a build constraint, or not declared). Failing rather than reporting a green check for a test that did not run."
    exit 1
  fi
else
  echo "--- Run acceptance tests"
  EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc
fi
