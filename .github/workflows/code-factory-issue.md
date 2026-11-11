---
imports:
  - shared/setup-dev.md
name: Code Factory Issue Intake
timeout-minutes: 65
description: >-
  Reacts to trusted qualifying `code-factory` issue events or internal workflow dispatch requests,
  suppresses duplicate linked pull requests, and delegates implementation to an agent that creates
  exactly one linked PR per issue. Self-verifies with the cloud provider's `make` gate and never
  runs acceptance tests (`TF_ACC`).
on:
  issues:
    types: [opened, labeled]
  bots:
    - github-actions[bot]
  workflow_dispatch:
    inputs:
      issue_number:
        description: 'Issue number to implement'
        required: true
        type: number
      source_workflow:
        description: 'Source workflow that triggered this dispatch'
        required: false
        type: string
  status-comment: true
  permissions:
    contents: read
    issues: write
    pull-requests: read
  steps:
    - name: Checkout repository
      uses: actions/checkout@v7.0.1
      with:
        persist-credentials: false
        fetch-depth: 1
    - name: Determine intake mode
      id: determine_intake_mode
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const mode = context.eventName === 'workflow_dispatch' ? 'dispatch' : 'issue-event';
          core.setOutput('intake_mode', mode);
          core.info(`Intake mode: ${mode}`);
    - name: Qualify trigger event
      id: qualify_trigger
      # Gate on gh-aw's membership check: a denied/failed lookup yields is_team_member != 'true',
      # so an untrusted actor who can apply the trigger label cannot reach the downstream
      # label-mutating steps (remove_trigger_label / set_phase_label), which cascade from this.
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
        steps.check_membership.outputs.is_team_member == 'true'
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: code-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/qualify-trigger.js');
          await fn({ github, context, core });
    - name: Capture issue context
      id: capture_issue_context
      if: steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const { boundEnvPayload } = require('${{ github.workspace }}/.github/scripts/workflows/lib/env-payload.js');
          const { flattenTitle } = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/fetch-live-issue.js');
          // Bound the body to a UTF-8 byte budget: it is passed to the normalize/sanitize steps
          // via environment variables, and a value above ~128 KB (MAX_ARG_STRLEN) would make those
          // steps fail to launch with E2BIG. Shared helper (also drops an orphaned HTML-comment
          // opener so the downstream sanitiser cannot erase the tail).
          const issueBody = boundEnvPayload(context.payload.issue?.body ?? '');
          core.setOutput('issue_number', context.payload.issue?.number ?? '');
          // Flatten newlines: normalize_context exports the title with a bare `echo` into
          // GITHUB_OUTPUT, where a newline would start a second key=value line (e.g. an injected
          // issue_number= could redirect the run). The dispatch path's fetch-live-issue flattens
          // the same way.
          core.setOutput('issue_title', flattenTitle(context.payload.issue?.title));
          core.setOutput('issue_body', issueBody);
    - name: Validate dispatch inputs
      id: validate_dispatch_inputs
      # Same membership gate as the issue-event path. For dispatch the trusted actor is the
      # scanner bot, which is_team_member == 'true' only via on.bots: [github-actions[bot]].
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
        steps.check_membership.outputs.is_team_member == 'true'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/validate-dispatch-inputs.js');
          await fn({ github, context, core });
    - name: Fetch live issue
      id: fetch_live_issue
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
        steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
      env:
        INPUT_ISSUE_NUMBER: ${{ steps.validate_dispatch_inputs.outputs.issue_number }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/fetch-live-issue.js');
          await fn({ github, context, core });
    - name: Fetch issue comments
      id: fetch_issue_comments
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      env:
        INPUT_ISSUE_NUMBER: >-
          ${{ steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
            && steps.capture_issue_context.outputs.issue_number
            || steps.validate_dispatch_inputs.outputs.issue_number }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/fetch-issue-comments.js');
          await fn({ github, context, core });
    - name: Check duplicate PR
      id: check_duplicate_pr
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: code-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/check-duplicate-pr.js');
          await fn({ github, context, core });
    - name: Remove trigger label
      id: remove_trigger_label
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
        steps.qualify_trigger.outputs.event_eligible == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: code-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/remove-trigger-label.js');
          await fn({ github, context, core });
    - name: Set phase label
      id: set_phase_label
      if: >-
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true' &&
        (
          (
            steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
            steps.qualify_trigger.outputs.event_eligible == 'true'
          ) || (
            steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
            steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
          )
        )
      env:
        INPUT_ISSUE_NUMBER: >-
          ${{ steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
            && steps.capture_issue_context.outputs.issue_number
            || steps.validate_dispatch_inputs.outputs.issue_number }}
        PHASE_LABEL_NAME: phase-coding
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/set-phase-label.js');
          await fn({ github, context, core });
    - name: Normalize context
      id: normalize_context
      if: always()
      env:
        INTAKE_MODE: ${{ steps.determine_intake_mode.outputs.intake_mode }}
        ISSUE_NUMBER_EVENT: ${{ steps.capture_issue_context.outputs.issue_number }}
        ISSUE_TITLE_EVENT: ${{ steps.capture_issue_context.outputs.issue_title }}
        ISSUE_BODY_EVENT: ${{ steps.capture_issue_context.outputs.issue_body }}
        EVENT_ELIGIBLE_EVENT: ${{ steps.qualify_trigger.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON_EVENT: ${{ steps.qualify_trigger.outputs.event_eligible_reason }}
        TRIGGER_LABEL_REMOVED_EVENT: ${{ steps.remove_trigger_label.outputs.trigger_label_removed }}
        TRIGGER_LABEL_REMOVED_REASON_EVENT: ${{ steps.remove_trigger_label.outputs.trigger_label_removed_reason }}
        ISSUE_NUMBER_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_number }}
        ISSUE_TITLE_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_title }}
        ISSUE_BODY_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_body }}
        EVENT_ELIGIBLE_DISPATCH: ${{ steps.validate_dispatch_inputs.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON_DISPATCH: ${{ steps.validate_dispatch_inputs.outputs.event_eligible_reason }}
        SOURCE_WORKFLOW: ${{ github.event.inputs.source_workflow }}
        DUPLICATE_PR_FOUND: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
        DUPLICATE_PR_URL: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
      run: |
        echo "intake_mode=${INTAKE_MODE}" >> "$GITHUB_OUTPUT"
        EOF_DELIM="EOF_$(cat /proc/sys/kernel/random/uuid 2>/dev/null || date +%s%N)"

        if [ "${INTAKE_MODE}" = "issue-event" ]; then
          echo "issue_number=${ISSUE_NUMBER_EVENT}" >> "$GITHUB_OUTPUT"
          echo "issue_title=${ISSUE_TITLE_EVENT}" >> "$GITHUB_OUTPUT"
          {
            echo "issue_body<<${EOF_DELIM}"
            printf '%s\n' "${ISSUE_BODY_EVENT}"
            echo "${EOF_DELIM}"
          } >> "$GITHUB_OUTPUT"
          echo "event_eligible=${EVENT_ELIGIBLE_EVENT}" >> "$GITHUB_OUTPUT"
          echo "event_eligible_reason=${EVENT_ELIGIBLE_REASON_EVENT}" >> "$GITHUB_OUTPUT"
          echo "actor_trusted=true" >> "$GITHUB_OUTPUT"
          echo "actor_trusted_reason=Role-based gate guarantees trust for issue events." >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed=${TRIGGER_LABEL_REMOVED_EVENT}" >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed_reason=${TRIGGER_LABEL_REMOVED_REASON_EVENT}" >> "$GITHUB_OUTPUT"
        else
          echo "issue_number=${ISSUE_NUMBER_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "issue_title=${ISSUE_TITLE_DISPATCH}" >> "$GITHUB_OUTPUT"
          {
            echo "issue_body<<${EOF_DELIM}"
            printf '%s\n' "${ISSUE_BODY_DISPATCH}"
            echo "${EOF_DELIM}"
          } >> "$GITHUB_OUTPUT"
          echo "event_eligible=${EVENT_ELIGIBLE_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "event_eligible_reason=${EVENT_ELIGIBLE_REASON_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "actor_trusted=true" >> "$GITHUB_OUTPUT"
          echo "actor_trusted_reason=Dispatch intake bypasses actor trust check." >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed=false" >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed_reason=Trigger label removal is not required for dispatch intake." >> "$GITHUB_OUTPUT"
          echo "source_workflow=${SOURCE_WORKFLOW}" >> "$GITHUB_OUTPUT"
        fi

        echo "duplicate_pr_found=${DUPLICATE_PR_FOUND}" >> "$GITHUB_OUTPUT"
        echo "duplicate_pr_url=${DUPLICATE_PR_URL}" >> "$GITHUB_OUTPUT"
      shell: bash
    - name: Write sanitized context files
      id: sanitize_context
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      env:
        FACTORY_NAME: code-factory
        # write-context-files.js sanitizes ISSUE_BODY / ISSUE_COMMENTS (by value) and writes
        # issue_body.md / issue_comments.md to the factory context dir. Body and comments are
        # byte-bounded upstream (boundEnvPayload in capture / fetch-live-issue; the comment library
        # budget), so they stay under the step env-var size limit. Mirrors research-factory.
        ISSUE_BODY: ${{ steps.normalize_context.outputs.issue_body }}
        ISSUE_COMMENTS: ${{ steps.fetch_issue_comments.outputs.issue_comments }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/write-context-files.js');
          await fn({ github, context, core });
    - name: Upload issue context artifact
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      uses: actions/upload-artifact@v7.0.2
      with:
        name: code-factory-issue-context
        path: /tmp/code-factory-context/
        if-no-files-found: error
    - name: Finalize gate reason
      id: finalize_gate
      if: always()
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: code-factory
        EVENT_ELIGIBLE: ${{ steps.normalize_context.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON: ${{ steps.normalize_context.outputs.event_eligible_reason }}
        ACTOR_TRUSTED: ${{ steps.normalize_context.outputs.actor_trusted }}
        ACTOR_TRUSTED_REASON: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
        DUPLICATE_PR_FOUND: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
        DUPLICATE_PR_URL: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
        DUPLICATE_GATE_REASON: ${{ steps.check_duplicate_pr.outputs.gate_reason }}
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/finalize-gate.js');
          await fn({ github, context, core });
concurrency:
  group: gh-aw-code-factory-${{ github.event.issue.number || github.event.inputs.issue_number }}
  job-discriminator: ${{ github.event.issue.number || github.event.inputs.issue_number }}
  queue: max
if: >-
  needs.pre_activation.outputs.event_eligible == 'true' &&
  needs.pre_activation.outputs.actor_trusted == 'true' &&
  needs.pre_activation.outputs.duplicate_pr_found != 'true' &&
  needs.pre_activation.outputs.issue_number != ''
steps:
  - name: Download issue context artifact
    uses: actions/download-artifact@v8.0.2
    with:
      name: code-factory-issue-context
      # Extract under /tmp/gh-aw, the one host path the AWF agent container bind-mounts
      # read-write (--mount /tmp/gh-aw:/tmp/gh-aw:rw). A sibling /tmp path like
      # /tmp/code-factory-context is NOT mounted, so the agent could not read the sanitized
      # issue body / comments the prompt points it to.
      path: /tmp/gh-aw/code-factory-context/
model: "anthropic/claude-sonnet-5"
engine:
  id: claude
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://openrouter.ai/api"
    ANTHROPIC_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
# Disable the per-run AI Credits budget guard. The OpenRouter model slug
# "anthropic/claude-sonnet-5" may be absent from the AWF api-proxy's built-in
# pricing table; gh-aw's models.providers override does not propagate to
# apiProxy.defaultAiCreditsPricing (github/gh-aw#47365), so with the guard
# active the proxy could reject every request with HTTP 400. Setting -1 omits
# maxAiCredits from the generated AWF config. Mirrors issue-classifier.md.
max-ai-credits: -1
# The daily guardrail is also disabled: it fails closed ("Daily AIC accounting
# record could not be resolved") when prior runs lack resolvable AIC usage.
max-daily-ai-credits: -1
permissions:
  contents: read
  issues: read
  pull-requests: read
jobs:
  pre-activation:
    outputs:
      intake_mode: ${{ steps.normalize_context.outputs.intake_mode }}
      issue_number: ${{ steps.normalize_context.outputs.issue_number }}
      issue_title: ${{ steps.normalize_context.outputs.issue_title }}
      issue_body: ${{ steps.normalize_context.outputs.issue_body }}
      event_eligible: ${{ steps.normalize_context.outputs.event_eligible }}
      event_eligible_reason: ${{ steps.normalize_context.outputs.event_eligible_reason }}
      actor_trusted: ${{ steps.normalize_context.outputs.actor_trusted }}
      actor_trusted_reason: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
      duplicate_pr_found: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
      duplicate_pr_url: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
      trigger_label_removed: ${{ steps.normalize_context.outputs.trigger_label_removed }}
      trigger_label_removed_reason: ${{ steps.normalize_context.outputs.trigger_label_removed_reason }}
      source_workflow: ${{ steps.normalize_context.outputs.source_workflow }}
      gate_reason: ${{ steps.finalize_gate.outputs.gate_reason }}
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [issues, pull_requests, repos]
network:
  allowed: [defaults, node, go, registry.terraform.io, releases.hashicorp.com, www.elastic.co, openrouter.ai]
env:
  CHECKPOINT_DISABLE: "1"
mcp-servers:
  elastic-docs:
    url: "https://www.elastic.co/docs/_mcp/"
    allowed: ["*"]
checkout:
  fetch-depth: 0
safe-outputs:
  # Match the repo's agentic-workflow convention (issue-classifier.md): this human-triggered,
  # human-reviewed factory does not auto-file framework failure/incomplete issues. Disabling
  # report-failed-jobs also removes the generated conclusion step that passes `toJSON(needs)`
  # (including gh-aw's activation body/text) through an environment variable, so a pathological
  # issue body cannot exceed the env-var size limit there.
  report-failure-as-issue: false
  report-incomplete: false
  report-failed-jobs: false
  create-pull-request:
    labels: [code-factory]
    max: 1
    patch-format: am
    preserve-branch-name: true
    allowed-branches: ["code-factory/issue-*"]
    recreate-ref: true
    fallback-as-issue: false
    protected-files:
      exclude: [".changelog/"]
  noop:
    max: 1
    report-as-issue: false
  jobs:
    add-changelog:
      name: add-changelog
      description: >-
        Commit the .changelog/{PR}.txt fragment for this PR. Call exactly once, after creating the
        pull request. Pass a release-note type (and a one-line body unless the type is 'none').
      runs-on: ubuntu-latest
      needs: safe_outputs
      if: needs.safe_outputs.result == 'success' && needs.safe_outputs.outputs.created_pr_number != ''
      permissions:
        contents: write
        pull-requests: read
      inputs:
        release_note_type:
          description: "release-note category. Use 'none' for non-user-facing changes."
          required: true
          type: choice
          options: [none, bug, enhancement, feature, new-resource, new-data-source, new-guide, note, breaking-change]
        release_note_body:
          description: "One-line user-facing note. Leave empty when release_note_type is 'none'."
          required: false
          type: string
      steps:
        - name: Checkout helper scripts
          uses: actions/checkout@v7.0.1
          with:
            persist-credentials: false
            sparse-checkout: .github/scripts/workflows/lib
            sparse-checkout-cone-mode: true
            fetch-depth: 1
        - name: Render and validate the changelog fragment
          id: render
          run: |
            node .github/scripts/workflows/lib/code-factory-changelog.js --render-only \
              > "$RUNNER_TEMP/fragment.txt"
        - name: Commit the fragment to the PR branch
          env:
            GH_TOKEN: ${{ github.token }}
            PR_NUMBER: ${{ needs.safe_outputs.outputs.created_pr_number }}
          run: |
            set -euo pipefail
            # Derive the target branch from the TRUSTED created-PR number (not from any
            # agent-supplied value): commit the changelog onto the PR we just created. Assert it is
            # a factory branch so we never write elsewhere. (A safe-job cannot read the triggering
            # issue number from pre-activation, so the branch cannot be bound to the trigger here;
            # the create-PR allowed-branches glob + the duplicate-PR gate + the agent prompt are the
            # controls for that.)
            BRANCH=$(gh api "repos/${GITHUB_REPOSITORY}/pulls/${PR_NUMBER}" --jq .head.ref)
            # Match the whole string (bash =~ anchors to the full value), so a crafted multi-line
            # ref cannot pass on its first line.
            if [[ ! "${BRANCH}" =~ ^code-factory/issue-[0-9]+$ ]]; then
              echo "::error::PR #${PR_NUMBER} head ref '${BRANCH}' is not a code-factory/issue-<n> branch; refusing to write"
              exit 1
            fi
            # Upsert .changelog/{PR}.txt via the Contents API (verified commit), so the committed
            # file is ALWAYS exactly our validated render. If the path already exists on the branch
            # (a re-run, or the agent/a human wrote a different or invalid fragment — .changelog/ is
            # excluded from protected-files, so the agent *can* write there), fetch its blob sha and
            # PUT with it, overwriting the stale content. A byte-identical content+sha PUT is a
            # no-op on GitHub, so this stays idempotent without a spurious commit.
            # DEFERRED (plan D1): this commit is made with GITHUB_TOKEN, so by GitHub's
            # anti-recursion rule it does not re-trigger the PR's required checks; a human nudge (or
            # the later Vault/App-token fix) is needed to green the new HEAD.
            EXISTING_SHA=$(gh api \
              "repos/${GITHUB_REPOSITORY}/contents/.changelog/${PR_NUMBER}.txt?ref=${BRANCH}" \
              --jq .sha 2>/dev/null || true)
            # Base64-encode to a PERSISTENT temp file and pass it with `-F content=@<file>`. Not a
            # plain variable (`-f content="$B64"`): a pathologically large release note would make
            # the base64 string exceed MAX_ARG_STRLEN (~128 KB) and `gh api` could not launch
            # (E2BIG). Not a process substitution (`@<(base64 ...)`): a /dev/fd descriptor stored in
            # the array is not guaranteed open when gh api reads it later. A real file avoids both.
            base64 -w0 "$RUNNER_TEMP/fragment.txt" > "$RUNNER_TEMP/fragment.b64"
            PUT_ARGS=(--method PUT "repos/${GITHUB_REPOSITORY}/contents/.changelog/${PR_NUMBER}.txt"
              -f message="Add changelog entry for PR #${PR_NUMBER}"
              -f branch="${BRANCH}"
              -F content=@"$RUNNER_TEMP/fragment.b64")
            if [ -n "${EXISTING_SHA}" ]; then
              echo "existing .changelog/${PR_NUMBER}.txt found (sha ${EXISTING_SHA}); overwriting with the validated fragment"
              PUT_ARGS+=(-f sha="${EXISTING_SHA}")
            fi
            if ! gh api "${PUT_ARGS[@]}"; then
              echo "::error::failed to commit changelog fragment to ${BRANCH}"
              exit 1
            fi
            echo "Committed .changelog/${PR_NUMBER}.txt to ${BRANCH}"
---

# Code Factory issue intake worker

You implement exactly one GitHub issue labeled `code-factory`. The triggering issue is the sole
source of truth for scope and requested behavior. Any acceptance criteria defined by the issue must
be met, and your changes must be properly covered by automated testing. Do not broaden the scope
beyond what the issue describes unless the repository already requires it to make the implementation
viable.

## Pre-activation context

Deterministic pre-activation has already decided that this intake is eligible, the actor is trusted
(or dispatch bypasses trust), and there is no open linked `code-factory` pull request for the issue.

- **Gate reason**: ${{ needs.pre_activation.outputs.gate_reason }}
- **Intake mode**: `${{ needs.pre_activation.outputs.intake_mode }}`
- **Issue number**: `${{ needs.pre_activation.outputs.issue_number }}`
- **Issue title**: `${{ needs.pre_activation.outputs.issue_title }}`
- **Issue body** (sanitised): see `/tmp/gh-aw/code-factory-context/issue_body.md`
- **Comment history** (sanitised, human-authored): see `/tmp/gh-aw/code-factory-context/issue_comments.md`
- **Repository**: `${{ github.repository }}`
- **Triggered by**: `@${{ github.actor }}`
- **Required branch**: `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`

## This is the cloud provider — no acceptance tests, ever

This repository manages **real, paid Elastic Cloud deployments**. Acceptance tests
(`TF_ACC=1`, `make testacc`, `make sweep`) create and destroy live infrastructure and cost money.
They run **out-of-band on Buildkite** for every pull request and are reviewed by a human. You must
**never** set `TF_ACC`, run `make testacc` / `make sweep`, or touch `EC_API_KEY`. There is no local
Elastic stack. Your self-verification is the deterministic `make` gate below — nothing more.

## Elastic documentation

An `elastic-docs` MCP server is available (`search_docs`, `find_related_docs`,
`get_document_by_url`). Before writing implementation code, use `search_docs` to ground the Elastic
Cloud / provider API behavior relevant to the issue. If the tools are unavailable or unhelpful,
proceed from the issue content alone — do not block on documentation.

## Task

Implement the triggering issue on branch
`code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}` and create exactly one linked
pull request labeled `code-factory`.

1. Read the issue title and body carefully and treat them as authoritative.
2. Create or update the implementation on branch
   `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`.
3. Verify your changes according to the **Verification ladder** below. Run the steps **in order**,
   each to completion.
4. Open exactly one pull request for that branch using the `create-pull-request` safe output.
5. Include `Closes #${{ needs.pre_activation.outputs.issue_number }}` in the PR body — this exact
   string is the stable identifier that prevents duplicate PRs on future runs.
6. After the PR is created, call the `add-changelog` safe-job **exactly once** with a release-note
   type (and a one-line body unless the type is `none`) — see **Changelog**.
7. Keep the pull request labeled `code-factory`.

## Verification ladder (mirrors the `go.yml` gate; never sets `TF_ACC`)

Run **in order**, each to completion. Stop and fix on any failure.

0. **Preflight** — confirm the toolchain is present (a missing tool must fail, not pass silently):
   ```bash
   for t in terraform go node jq git base64; do
     command -v "$t" >/dev/null || { echo "::error::missing required tool: $t"; exit 1; }
   done
   command -v cc >/dev/null || command -v gcc >/dev/null \
     || { echo "::error::no C toolchain for go test -race"; exit 1; }
   ```
   **Reuse the primed Go cache:** the dev-setup step warmed the module and build caches under
   `/tmp/gh-aw/gocache` (a path this sandbox shares read-write with the host). Export them so your
   `make` commands do not re-download or rebuild from scratch:
   ```bash
   export GOMODCACHE=/tmp/gh-aw/gocache/mod GOCACHE=/tmp/gh-aw/gocache/build
   ```
   If `make install` / `make validate-examples` cannot write to `~/.terraform.d`, run them with a
   writable home **outside the repo checkout** (so it does not leave an untracked dir that fails the
   clean-tree check at step 10): `mkdir -p /tmp/gh-aw/home && HOME=/tmp/gh-aw/home make install validate-examples`.
1. `make check-label-taxonomy`
2. `make lint` (already runs `fmt`, `check-license`, `tfproviderlint`, `tfproviderdocs`)
3. `make build`
4. `make docs-generate` then `git diff --exit-code docs/`
5. `make notice` then `git diff --exit-code NOTICE`
6. `make gen` then `git add --intent-to-add -- ec/internal/gen/serverless && git diff --exit-code -- ec/internal/gen/serverless`
7. `make unit`
8. `make install validate-examples`
9. Only if you changed `.github/scripts/workflows/**` or `scripts/changelogcheck/**`: `make workflow-test`
10. `git status --porcelain` must be empty before you open the PR.

Do **not** run `make workflow-generate` (it needs the `gh aw` extension from a host outside the agent
firewall). Do **not** edit `.github/workflows/*.md`, `.github/aw/**`, or
`.agents/skills/pr-monitoring-loop/scripts/**` — those paths are verified by steps you cannot run in
this sandbox.

## Pull request contract

The linked pull request must:

- use branch `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`;
- be the only open `code-factory` pull request for this issue;
- include `Closes #${{ needs.pre_activation.outputs.issue_number }}` in the body;
- follow the repository PR template (Description / Related Issues / Motivation and Context / How Has
  This Been Tested? / Types of Changes / Readiness Checklist). Under **How Has This Been Tested?**
  list the verification ladder you ran and state plainly that **acceptance tests were not run and run
  out-of-band on Buildkite**;
- stay focused on implementing the triggering issue only.

## Changelog

Every PR needs a `.changelog/{PR}.txt` entry. You do not write the file yourself — after the PR
exists, call the **`add-changelog`** safe-job exactly once:

- `release_note_type`: one of `none`, `bug`, `enhancement`, `feature`, `new-resource`,
  `new-data-source`, `new-guide`, `note`, `breaking-change`.
- `release_note_body`: a **single-line** user-facing sentence with no code fences — **omit/empty only
  when the type is `none`**.

Pick the type from the nature of the change: refactoring / deduplication / test-coverage / CI /
docs → `none`; a user-visible bug fix → `bug`; a new attribute or capability → `enhancement` (or
`new-resource` / `new-data-source`). When in doubt prefer `none` over inventing customer impact the
issue does not describe.

## Guardrails

- Do not re-check trigger eligibility, actor trust, or duplicate PR state — pre-activation did that.
- **Never** set `TF_ACC`, run `make testacc` / `make sweep`, or touch `EC_API_KEY`.
- Run `make lint` and the rest of the ladder before opening the PR; an unformatted or failing tree
  blocks CI.
- Open exactly one pull request for this issue; do not open a second, and do not open issues.
- Call `add-changelog` exactly once, after the PR exists.
- Do not change the branch naming convention.
- Do not use another issue, pull request, or external request as the source of truth over the
  triggering issue.
- If you cannot make progress safely, use `noop` with a concise explanation instead of opening a
  half-finished pull request.
