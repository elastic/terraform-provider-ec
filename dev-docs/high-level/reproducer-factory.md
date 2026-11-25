# The `reproducer-factory` workflow

`reproducer-factory` is a GitHub Agentic Workflow (gh-aw). Its source is
[`reproducer-factory-issue.md`](../../.github/workflows/reproducer-factory-issue.md), compiled to
`reproducer-factory-issue.lock.yml`. The supply-chain and network policy for all agentic workflows is
in [`.github/aw/README.md`](../../.github/aw/README.md); the labels that drive them are in
[`label-taxonomy.md`](./label-taxonomy.md).

This page is for the maintainers who use the workflow: what it produces, how to run the test it
writes, and what that costs.

Like every agentic workflow in this repository, it runs no acceptance tests, never sets `TF_ACC`, and
holds no cloud credential. Acceptance tests create real, paid deployments, and they run only on
Buildkite. See [`testing.md`](./testing.md).

## What it does

A maintainer applies the `reproducer-factory` label to a bug report. An agent investigates the bug by
reading the code and, where it helps, by writing and running an offline unit test. It then does two
things:

1. **Always:** creates or updates one sticky comment on the issue with one of three outcomes:
   `reproduced-by-inspection`, `cannot-reproduce`, or `appears-fixed`.
2. **Only on `reproduced-by-inspection`:** opens one **draft** pull request on the branch
   `reproducer-factory/issue-<N>`. The pull request contains a single file,
   `ec/acc/issue_<N>_acc_test.go`, with a single test, `TestAccReproduceIssue<N>`.

The agent writes that acceptance test and checks that it compiles. It does not run it.

## The test fails on purpose

The repro test asserts the behaviour the provider *should* have, so it fails while the bug exists and
passes once the bug is fixed. It is meant to become the fix's acceptance test. That is why the pull
request is a draft: it is a handover to a human, not something to merge as opened.

A red acceptance check on a repro pull request therefore has two possible meanings: the bug
reproduced, or the environment flaked. The pull request body and the sticky comment both quote the
exact failure text the test is expected to produce. Compare the Buildkite log against it before
concluding anything.

## Nothing is spent until a maintainer asks

Opening the pull request does not start the acceptance pipeline. Buildkite builds a pull request on
open only when its author has standing in
[`.buildkite/pull-requests.json`](../../.buildkite/pull-requests.json): the author is in
`allowed_list`, is a member of the organisation (`allow_org_users`), or has `admin` or `write` on the
repository (`allowed_repo_permissions`). Factory pull requests are opened by `github-actions[bot]`,
which is none of those.

That depends on two things staying as they are:

- **`allowed_list` does not gain a factory identity.** A test pins this.
- **The `GH_AW_GITHUB_TOKEN` secret is not set to a token of an organisation member or of a user with
  write access.** The workflow opens the pull request with that secret when it exists and with the
  default workflow token otherwise. No test can see a secret, so setting it is a spend decision:
  every repro pull request would then start a paid run the moment it opens.

To run the repro test, a maintainer **reads the test file first** and then comments on the pull
request:

```
buildkite test this
```

That comment is the decision to spend. It is also a decision to execute the file: the test was
written by an agent that read an untrusted issue, and the build runs it with the live Elastic Cloud
API key in its environment. Review it as you would review code from an outside
contributor before asking Buildkite to run it.

## A repro pull request runs one test, until it stops being one

For a repro pull request the acceptance build runs only `TestAccReproduceIssue<N>` instead of the
whole suite. The rule is in
[`.buildkite/acceptance-test-selection.sh`](../../.buildkite/acceptance-test-selection.sh), and it
applies only while all of these hold:

- the branch is exactly `reproducer-factory/issue-<N>` and the build is a pull-request build;
- the pull request's whole diff is the addition of `ec/acc/issue_<N>_acc_test.go`, with renames
  counted as a deletion plus an addition;
- every commit on the branch, merge commits aside, was authored by `github-actions[bot]`;
- the branch has a single merge base with the base branch;
- that file declares `TestAccReproduceIssue<N>`.

A narrowed build also fails, rather than going green, when the run exits 0 without a `--- PASS` line
for the test: a skipped test, a file excluded by a build constraint, or a declaration that exists
only in a comment would otherwise report success having run nothing.

To run the test against current `master`, bring the pull request up to date with **Update branch**
on the pull request page, then comment. The merge commit that button creates is exempt from the
author rule above, so the build still runs the one test. The exemption covers only a plain merge of
the base branch that leaves the repro test as it was. A merge of any other branch, a merge with more
than two parents, or a merge that edits the test runs the full suite.

On a build agent whose checkout is shallow, narrowing only sees the last 50 commits of the base
branch, and a pull request that forked earlier than that runs the full suite.

If any of the conditions fails, the full suite runs. In particular, **pushing a fix to the repro branch
switches the build back to the full suite**, so a fix cannot merge having run a single test. Every
doubt resolves to the full suite, because `buildkite/terraform-provider-ec-acceptance` is a required
check on `master`.

## Re-running and duplicates

The trigger label is consumed when the factory starts. Re-apply it to run again.

While a pull request is open on `reproducer-factory/issue-<N>`, a new run is blocked and the workflow
says so in a comment on the issue. A re-run would otherwise recreate the branch underneath that pull
request. Close the pull request, or move it to another branch, then re-apply the label. Do not remove
the `reproducer-factory` label from a repro pull request.

A re-run **force-recreates** `reproducer-factory/issue-<N>`. Commits a human pushed to that branch are
dropped from it and stay reachable only through the closed pull request. If the branch carries work
worth keeping, push it to a branch of another name before re-applying the label.

## What the workflow can and cannot do

These hold whatever the agent does, and are asserted in tests against the compiled lock:

| Property | Where it is enforced |
| --- | --- |
| The agent job has read-only permissions | lock, `reproducer-factory-lock.test.mjs` |
| The pull request may contain only `ec/acc/issue_<N>_acc_test.go` for the triggering issue | `allowed-files`, same test |
| The pull request is a draft, on the triggering issue's branch, never closes the issue, and never falls back to opening an issue | same test |
| The lock and the prompt contain no acceptance-test invocation and no cloud credential | `acc-free-guard.test.mjs` |
| Opening a pull request does not start a paid run, while `GH_AW_GITHUB_TOKEN` is unset or has no standing with Buildkite | `allowed_list` and the token expression, pinned in `reproducer-factory-lock.test.mjs`; the secret itself is not testable |
| A build is narrowed only for a pull request whose diff is the one added repro file | `buildkite-acceptance-test-selection.test.mjs` |
| A narrowed build that did not run and pass the test fails | same test |

One thing is a prompt instruction rather than a control: that a pull request is opened only on the
`reproduced-by-inspection` outcome. gh-aw v0.92.0 does not accept a condition on the pull-request
output. A pull request opened on another outcome would still be a draft containing only a test file,
and would still cost nothing until a maintainer comments.

## Before enabling it

- The workflow needs the `OPENROUTER_API_KEY` secret, like the other factories.
- The agent can read `elastic/cloud-sdk-go` in addition to this repository.
- Free GitHub Actions checks (unit tests, lint, CLA) do not start on a pull request opened with the
  default workflow token. They start once a human pushes to the branch.
- The narrowing rule recognises the factory by the commit author `github-actions[bot]`. If factory
  pull requests are later opened under a different identity, repro builds run the full suite until
  `REPRODUCER_FACTORY_COMMIT_AUTHOR` in `acceptance-test-selection.sh` is updated.

## Working on the workflow

- Regenerate the lock with `make workflow-generate` after editing the `.md`, and commit both.
- Run `env -u TF_ACC make workflow-test`. It needs no credentials.
- The prompt must name forbidden things as bare tokens and never as a runnable command;
  `acc-free-guard.test.mjs` explains why and fails otherwise.
