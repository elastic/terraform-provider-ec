#!/bin/bash
# Sourced by .buildkite/acceptance.sh. Decides whether a build runs the whole acceptance suite, one
# named test, or (docs_only_pr) nothing. Kept apart from acceptance.sh so the decision can be tested without running
# anything paid: see .github/scripts/workflows/lib/buildkite-acceptance-test-selection.test.mjs.
#
# Contract: acceptance_test_name prints a `go test -run` pattern, or prints nothing. Nothing means
# "run the full suite". Every doubt resolves to nothing, because the Buildkite context is a required
# check on master and a narrowed run that should have been a full one would let an untested change
# merge green. Each arm must therefore fail open towards the full suite, never towards a narrow one.
#
# One arm per factory branch prefix. To add one, add a case arm and a function; do not widen an
# existing arm.

# Fetches the base branch so the functions below can diff against it. A checkout that is already
# shallow stays shallow (--depth=50). A full clone is fetched without --depth: with it, git would cut
# the clone's history at 50 commits, and a pull request that forked earlier than that would lose its
# connection to the base, so every rule below would fall back to the full suite for it.
fetch_base() {
  local base="$1"
  if [[ "$(git rev-parse --is-shallow-repository 2>/dev/null)" == "true" ]]; then
    git fetch --depth=50 origin "$base" >&2
  else
    git fetch origin "$base" >&2
  fi
}

# True when HEAD has exactly one merge base with the base branch. With more than one (possible only
# when the base branch itself has merge commits) a three-dot diff is taken against whichever git
# lists first, and a merge on the pull request branch can then hide a change: the diff shows one
# harmless file while merging the pull request would also revert another. Both rules below read a
# three-dot diff, so both require this.
single_merge_base() {
  local base="$1" merge_bases
  merge_bases=$(git merge-base --all "origin/${base}" HEAD) || return 1
  [[ -n "$merge_bases" && "$merge_bases" != *$'\n'* ]]
}

# True when a pull request touches nothing but text that cannot change provider behaviour, in which
# case acceptance.sh reports the required check green without running the paid suite. Every doubt
# returns non-zero, which means "run the suite".
#
# Rename detection is off. With it on, moving a provider file into docs/ is reported as the single
# destination path, the deletion from the provider tree goes unseen, and the suite would be skipped
# for a change that removes provider code. With --no-renames the move is listed as both paths.
docs_only_pr() {
  [[ "${BUILDKITE_PULL_REQUEST:-false}" == "false" ]] && return 1
  local base="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-master}"
  # Fail-safe: if we cannot see the base, run the suite rather than skip.
  fetch_base "$base" || return 1
  single_merge_base "$base" || return 1
  local files
  files=$(git diff --no-renames --name-only "origin/${base}...HEAD") || return 1
  [[ -n "$files" ]] || return 1
  # Here-string: grep -q would SIGPIPE echo under pipefail and invert to a false skip.
  ! grep -qvE '^(docs/|dev-docs/|openspec/|\.changelog/)' <<< "$files"
}

# The reproducer-factory workflow opens a draft PR on reproducer-factory/issue-<N> carrying a single
# acceptance test that fails while the bug is present. Running the other ~60 tests against the live
# Elastic Cloud API for that PR buys nothing, so the build runs just that test.
#
# The prefix must stay equal to ISSUE_BRANCH_PREFIX in
# .github/scripts/workflows/lib/intake/reproducer-factory-constants.js, and the file and function
# names to the ones the prompt in .github/workflows/reproducer-factory-issue.md prescribes.
#
# Narrowing applies only while the PR is still exactly what the factory opened:
#   - the diff against the base is the addition of that one test file and nothing else. The moment
#     a fix (or any other change) lands on the branch, the full suite runs, so a fix cannot merge on
#     a one-test check. Rename detection is off: with it on, moving another file onto the test's
#     path shows up as that one path, and the deletion of the other file would go unseen;
#   - every commit on the branch, merge commits aside, was authored by the workflow's git identity.
#     This is a tripwire for a human continuing the work on the branch, not an authentication check:
#     an author name is whatever the committer says it is. The diff check above is what keeps the
#     result sound. Merge commits are left out so that a maintainer can bring the pull request up to
#     date with GitHub's "Update branch" button, whose merge commit carries their name, and still
#     get the one-test run. In exchange each merge commit must be a plain merge of the base that
#     leaves the test file alone, and the branch must have a single merge base with the base
#     branch; both are checked in the function;
#   - the file declares the test. This is only a cheap early check on the text: a build constraint,
#     a comment or a t.Skip can still leave nothing to run, and `go test -run` with a pattern that
#     matches nothing exits 0. narrowed_run_passed_test below is what catches that, on the output of
#     the run itself.
REPRODUCER_FACTORY_COMMIT_AUTHOR='github-actions[bot]'

reproducer_factory_test_name() {
  local branch="$1"
  [[ "$branch" =~ ^reproducer-factory/issue-([1-9][0-9]*)$ ]] || return 0
  local issue="${BASH_REMATCH[1]}"
  local test_file="ec/acc/issue_${issue}_acc_test.go"
  local test_name="TestAccReproduceIssue${issue}"

  [[ "${BUILDKITE_PULL_REQUEST:-false}" != "false" ]] || return 0
  local base="${BUILDKITE_PULL_REQUEST_BASE_BRANCH:-master}"
  fetch_base "$base" || return 0

  single_merge_base "$base" || return 0

  local changes
  changes=$(git diff --no-renames --name-status "origin/${base}...HEAD") || return 0
  [[ "$changes" == "A"$'\t'"$test_file" ]] || return 0

  local authors
  authors=$(git log --no-merges --format=%an "origin/${base}..HEAD") || return 0
  [[ -n "$authors" ]] || return 0
  # Here-string, not a pipe: grep -q would SIGPIPE the writer under pipefail.
  if grep -qvxF "$REPRODUCER_FACTORY_COMMIT_AUTHOR" <<< "$authors"; then
    return 0
  fi

  # Merge commits are exempt from the author check, so each must be nothing more than the base
  # branch merged in: two parents, the second already part of the base, and the repro test left as
  # the first parent had it. Anything else (a side branch, an octopus, an edit to the test made
  # inside the merge) is a human continuing the work, and the full suite runs.
  local merges merge
  merges=$(git rev-list --merges "origin/${base}..HEAD") || return 0
  for merge in $merges; do
    [[ "$(git rev-list --no-walk --parents "$merge" | wc -w)" -eq 3 ]] || return 0
    git merge-base --is-ancestor "${merge}^2" "origin/${base}" || return 0
    git diff --quiet "${merge}^1" "$merge" -- "$test_file" || return 0
  done

  # A regular file, not a symlink: the test Go compiles must be the text checked here.
  [[ -f "$test_file" && ! -L "$test_file" ]] || return 0
  grep -qE "^func ${test_name}\(t \*testing\.T\)" "$test_file" || return 0

  printf '^%s$' "$test_name"
}

# After a narrowed run that exited 0: did the selected test really run and pass? Takes the pattern
# acceptance_test_name printed and the captured `go test -v` output. Returns non-zero when there is
# no `--- PASS` line for the test or there is a `--- SKIP` line for it, which is what a skipped
# test, a file excluded by a build constraint, or a declaration that only exists inside a comment or
# a string looks like. Without this, such a run would paint the required check green having tested
# nothing.
narrowed_run_passed_test() {
  local pattern="$1" log="$2"
  local test_name="${pattern#^}"
  test_name="${test_name%\$}"
  [[ "$test_name" =~ ^[A-Za-z0-9_]+$ ]] || return 1
  if grep -qE "^--- SKIP: ${test_name} \(" "$log"; then
    return 1
  fi
  grep -qE "^--- PASS: ${test_name} \(" "$log"
}

acceptance_test_name() {
  local branch="${BUILDKITE_BRANCH:-}"
  case "$branch" in
    reproducer-factory/issue-*) reproducer_factory_test_name "$branch" ;;
    *) ;;
  esac
  return 0
}
