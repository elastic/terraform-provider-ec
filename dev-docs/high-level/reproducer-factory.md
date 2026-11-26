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

A red acceptance check on a repro pull request therefore has three possible meanings: nobody has
asked for the run yet (see the next section), the bug reproduced, or the environment flaked. The pull request body and the sticky comment both quote the
exact failure text the test is expected to produce. Compare the Buildkite log against it before
concluding anything.

## Nothing runs until a maintainer asks

Opening the pull request starts a Buildkite build, but that build runs nothing.

The acceptance pipeline gets builds from two sources:

- **Buildkite's GitHub webhook** builds every pull request branch of this repository on open and on
  push, whoever the author is. A repro pull request is built this way within seconds of opening.
- **Elastic's PR bot** ([`.buildkite/pull-requests.json`](../../.buildkite/pull-requests.json))
  creates a build for an author with standing, for a push by a person with standing, and for
  anyone with standing who comments `buildkite test this`.

Because the webhook does not look at the author, the gate is in the pipeline's own script. On any
`reproducer-factory/` branch, [`.buildkite/acceptance.sh`](../../.buildkite/acceptance.sh) runs
nothing, neither the tests nor the clean-up sweep, unless the PR bot created the build on behalf of a
person. The step fails with a message saying so. A repro pull request therefore shows a **red**
acceptance check from the moment it opens; that red means "not run yet", not "the bug reproduced".

To run the repro test, a maintainer **reads the test file first** and then comments on the pull
request:

```
buildkite test this
```

That comment is the decision to spend. It is also a decision to execute the file: the test was
written by an agent that read an untrusted issue, and the build runs it with the live Elastic Cloud
API key in its environment. Review it as you would review code from an outside contributor before
asking Buildkite to run it.

Two things keep this true:

- **The refusal rule stays first in `acceptance.sh`.** Tests pin its position and its conditions.
- **The `GH_AW_GITHUB_TOKEN` secret is not set to a token of a user account.** The workflow opens
  the pull request with that secret when it exists and with the default workflow token otherwise.
  The refusal rule counts any requester whose name does not end in `[bot]` as a person, so a token
  of a member or of a machine user (anything but a GitHub App) would make the PR bot build the pull
  request at once, and the build would run. No test can see a secret, so setting it is a spend
  decision.

**A push counts as asking.** When a person with standing pushes to the repro branch, which includes
clicking **Update branch**, the PR bot can create a build for that push, and it runs. Read the test
before pushing to the branch, as you would before commenting.

A build started by a `buildkite test this` comment has been observed on a bot-authored pull request
from a branch of this repository (Buildkite build 3475): the PR bot creates it, with source `api`,
the branch named `elastic:<branch>`, and the commenter in `GITHUB_PR_TRIGGER_USER`. That is the shape
the refusal rule lets through.

## A repro pull request runs one test, until it stops being one

For a repro pull request the acceptance build runs only `TestAccReproduceIssue<N>` instead of the
whole suite. The rule is in
[`.buildkite/acceptance-test-selection.sh`](../../.buildkite/acceptance-test-selection.sh), and it
applies only while all of these hold:

- the branch is exactly `reproducer-factory/issue-<N>` in this repository, and the build is for a
  pull request into `master`;
- the pull request's whole diff is the addition of `ec/acc/issue_<N>_acc_test.go`, with renames
  counted as a deletion plus an addition;
- every commit on the branch, merge commits aside, was authored by `github-actions[bot]`;
- the branch has a single merge base with the base branch;
- that file declares `TestAccReproduceIssue<N>`.

A narrowed build also fails, rather than going green, when the run exits 0 without a `--- PASS` line
for the test: a skipped test, a file excluded by a build constraint, or a declaration that exists
only in a comment would otherwise report success having run nothing.

To run the test against current `master`, read the test file, then bring the pull request up to
date with **Update branch** on the pull request page. That click is a push by you, so it can start
the run by itself; if no build starts, comment `buildkite test this`. Do not do both at once: a
second build on the branch cancels the first. The merge commit that button creates is exempt from
the author rule above, so the build still runs the one test. The exemption covers only a plain merge of
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
| A build on a repro branch runs nothing unless the PR bot created it for a person | `reproducer_factory_build_refused`, first in `acceptance.sh`; `buildkite-acceptance-test-selection.test.mjs` |
| The factory identity is not allowlisted with the PR bot, and the pull request is opened with the expected token expression | pinned in `reproducer-factory-lock.test.mjs`; the secret itself is not testable |
| A build is narrowed only for a pull request whose diff is the one added repro file | `buildkite-acceptance-test-selection.test.mjs` |
| A narrowed build that did not run and pass the test fails | same test |

One thing is a prompt instruction rather than a control: that a pull request is opened only on the
`reproduced-by-inspection` outcome. gh-aw v0.92.0 does not accept a condition on the pull-request
output. A pull request opened on another outcome would still be a draft containing only a test file,
and would still run nothing until a maintainer comments.

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
