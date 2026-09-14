---
name: "openspec-implementation-loop"
description: "Orchestrates an end-to-end implementation loop for a single OpenSpec change: select a change, ask commit-only vs PR delivery, triage (inline, single-implementor, or per-task), implement, run review and validation (make lint/build/unit always; targeted acc only after an explicit yes), push, then either watch GitHub Actions (commit mode) or create a PR and monitor it (PR mode). Use when the user wants to implement an approved OpenSpec proposal/change with iterative review and CI feedback."
license: "MIT"
compatibility: "Requires openspec CLI, git, and GitHub CLI."
metadata:
  author: openspec
  version: "3.0"
---

Orchestrate an implementation loop around a single OpenSpec change.

This skill is **hand-maintained** (not emitted by `make gen-openspec-skills`). Keep its directory name in `scripts/gen-openspec-skills.mjs` so regeneration does not delete it.

**Input**: Optionally specify a change name. If omitted, you MUST ask the user which change to implement.

**High-level flow**

1. Select the change
2. **Ask delivery mode (commit vs PR) - do this immediately after selecting the change, not after implementation**
3. Load OpenSpec context
4. **Triage: determine execution strategy** - classify as inline, single-implementor, or per-task based on change size and complexity
5. Determine remaining top-level tasks
6. Implement tasks using the chosen strategy
7. Run validation and review at the cadence determined by the strategy
8. Aggregate findings and fix; repeat until clean
9. Push the branch to `origin`
10. **Commit mode**: Watch GitHub Actions that run on a branch push (typically `Go`). Do not wait for `OpenSpec CI` — that workflow does not run on arbitrary branches. Surface Buildkite acceptance as out-of-band; never run or auto-fix it.
    **PR mode**: Create a PR, then monitor GitHub Actions (including `OpenSpec CI` when present), reviews, and comments. Delegate to `pr-monitoring-loop` when that skill is installed; otherwise watch with `gh` as described in step 11.
11. Report final outcome

**Steps**

1. **Select exactly one change**

   If a name is provided, use it.

   Otherwise:
   - Run `openspec list --json`
   - Use **AskUserQuestion** (or an equivalent explicit user prompt) to let the user choose a single active change
   - Show the change name, schema if available, and status

   **IMPORTANT**:
   - Do NOT guess
   - Do NOT auto-select
   - Do NOT implement multiple changes in one run

   Always announce: `Using change: <name>`.

2. **Choose delivery mode (ask immediately - before loading context or starting the implementor)**

   Use **AskUserQuestion** (or an equivalent explicit user prompt) right after step 1. Do **not** defer this until after implementation or push.

   Offer two options:

   - **Commit-only**: Push your work to `origin` on the current branch. After each push, monitor GitHub Actions for the **branch / commits** you pushed.
   - **Pull request**: After the **initial** push of the implementation loop, **create a PR** (for example with `gh pr create`). Then monitor **PR** checks and reviews as described in step 11.

   Record the user's choice and refer to it from push onward (steps 9–11).

3. **Load OpenSpec status and context**

   Run:
   ```bash
   openspec status --change "<name>" --json
   openspec instructions apply --change "<name>" --json
   ```

   Parse the outputs to determine:
   - `schemaName`
   - current task progress
   - `state`
   - `contextFiles`
   - the ordered list of top-level tasks (for example `1`, `2`, `3`) and which of them are still incomplete

   Read every file listed in `contextFiles`.

   **Handle states**:
   - If `state: "blocked"`: stop and explain what artifact is missing; suggest continuing the change artifacts first
   - If `state: "all_done"`: skip triage and implementation (steps 4–6). Continue at step 7: run 7a, 7b, and the 7d battery **once** (including `openspec-verify-change`). Do not use 7c or 7e. Then continue at step 8.
   - Otherwise: proceed to triage and implementation

4. **Triage: determine execution strategy**

   Evaluate the change to choose an execution strategy that balances implementation quality with subagent overhead. The goal is to avoid spawning unnecessary subagents for small changes while preserving full rigor for large ones.

   **Signals to evaluate:**

   | Signal | Source |
   |--------|--------|
   | Top-level task count | Count `## N.` headings in tasks |
   | Total subtask count | Count `- [ ]` / `- [x]` items |
   | File scope | Infer from task descriptions - single package/area vs. cross-cutting |
   | Inter-task coupling | Do later tasks build directly on earlier ones? |
   | Complexity | Do tasks involve non-trivial logic (custom types, plan modifiers, complex CRUD) or are they straightforward (config, docs, Makefile, CI, specs)? |

   **Choose one of three strategies:**

   | Strategy | When to use | Implementation | Review |
   |----------|-------------|----------------|--------|
   | **Inline** | ≤2 top-level tasks AND ≤~10 subtasks AND single area AND straightforward changes | Orchestrator implements directly, no implementor subagent | Run validation and `openspec-verify-change` directly; spawn other review subagents only if the change touches non-trivial logic |
   | **Single-implementor** | ≤4 top-level tasks OR ≤~15 subtasks, with coherent scope | One implementor subagent handles all remaining tasks | One round of parallel review after all tasks complete |
   | **Per-task** | >4 top-level tasks, OR >15 subtasks, OR multi-area scope, OR tasks are largely independent across different packages | Fresh implementor per top-level task | Task-scoped review after each task; `openspec-verify-change` only after the last |

   These thresholds are guidelines, not rigid rules. Use judgment:
   - A 3-task change where each task is in a different package might warrant **per-task**
   - A 5-task change that is all in one file might be fine as **single-implementor**
   - A 2-task change with complex custom type logic might benefit from **single-implementor** over **inline** for the review coverage

   Announce the chosen strategy and reasoning. The user can override.

5. **Determine the remaining top-level tasks**

   Build an ordered queue of incomplete top-level tasks from the OpenSpec task list.

   Interpret a top-level task as the parent task number such as `1`, `2`, or `3`. Each top-level task includes all of its nested subtasks such as `1.1`, `1.2`, `1.3`.

   For each incomplete top-level task:
   - gather the subtasks that belong to it
   - understand the intended scope from the proposal/design/specs
   - process the top-level tasks sequentially unless the user explicitly asks for a different strategy

6. **Implement tasks using the chosen strategy**

   **Inline strategy:**

   The orchestrator implements all tasks directly without spawning an implementor subagent:
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived. This loop never archives. After this skill returns, the user may run `openspec-archive-change` (or apply the later `verify-openspec` label when that workflow exists)
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing (`make unit` scoped to touched packages is fine; never `TF_ACC`)
   - do not push; the orchestrator pushes in step 9

   **Single-implementor strategy:**

   Launch one write-capable subagent for all remaining tasks:
   - instruct it to implement all remaining top-level tasks in sequence, completing all nested subtasks within each before moving to the next
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived. This loop never archives. After this skill returns, the user may run `openspec-archive-change` (or apply the later `verify-openspec` label when that workflow exists)
   - read the OpenSpec context files before editing
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing (`make unit` scoped to touched packages is fine; never `TF_ACC`)
   - do not push; the orchestrator pushes in step 9

   Ask the implementor to report back with:
   - top-level tasks completed
   - nested subtasks completed
   - commits created
   - tests run
   - blockers or open questions

   **Per-task strategy:**

   Launch a dedicated fresh write-capable subagent for the current top-level task only. Do **not** reuse the prior task's implementor for later top-level tasks.

   The implementor prompt should instruct it to:
   - implement only the selected change
   - focus only on the current top-level task and complete all of its nested subtasks before stopping
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived. This loop never archives. After this skill returns, the user may run `openspec-archive-change` (or apply the later `verify-openspec` label when that workflow exists)
   - read the OpenSpec context files before editing
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing (`make unit` scoped to touched packages is fine; never `TF_ACC`)
   - do not push; the orchestrator pushes in step 9

   Ask the implementor to report back with:
   - the top-level task completed
   - nested subtasks completed
   - commits created
   - tests run
   - blockers or open questions

   **For all strategies**: if the implementor (or the orchestrator in inline mode) is blocked, stop and surface the blocker to the user.

7. **Run validation and review**

   The validation and review cadence depends on the execution strategy.

   If this invocation skipped triage because `state` was `all_done`, run 7a, 7b, and the 7d battery once, then go to step 8.

   **7a. Determine the review type**

   Inspect the change artifacts and implementation to decide whether this is a Terraform entity change.

   Treat it as a Terraform entity change when the change is centered on a Terraform resource or data source, for example:
   - specs reference `Resource implementation:` or `Data source implementation:`
   - changed code is primarily in a resource/data source package under `ec/ecresource/` or `ec/ecdatasource/`

   **7b. Validation requirements (all strategies)**

   Required baseline commands — **always** run these:
   - `make lint`
   - `make build`
   - `make unit`
   - `make check-openspec` when the work touched `openspec/` (it is **not** part of `make lint`). If `openspec/` did not change, skip it and say so.

   **Acceptance tests — never auto-run. Ask once, named cases only:**
   - There is **no local Docker stack**. Acc hits the paid Elastic Cloud API (`EC_API_KEY`), creates real deployments, and costs money. The full suite is Buildkite-only. See `dev-docs/high-level/testing.md`.
   - Implementors and the validation runner MUST NOT set `TF_ACC` or run `make testacc`.
   - If the change has no runtime behavior (docs, skills, Makefile, spec-only), say so and skip acc.
   - Otherwise the **orchestrator** (not a subagent), **once per loop** (after the first complete 7b, and for per-task only with the final 7d battery — not after every task):
     1. Name the one or two `TestAcc…` cases that cover the change.
     2. Ask explicitly (AskUserQuestion or equivalent). Call out that this creates real Elastic Cloud resources and costs money. Default option: **skip** (human/Buildkite will run them). Other option: run those named cases.
     3. On skip, or if `EC_API_KEY` is unset: do not run acc; surface the names as a human/Buildkite step.
     4. On yes: run only `make testacc TEST_NAME='<exact TestAcc name or -run regex>'`. Never `make testacc` without `TEST_NAME`. Never the full suite.
     5. Do **not** retry acc on failure. Do **not** re-ask on later step 7–8 reruns (lint/unit only). If the opted-in run failed, that is push-blocking; after a code fix, ask again before running acc a second time.
     6. After an opted-in run, remind the user that leaked `terraform_acc_` resources need `make sweep`. Do not run sweep yourself (it is interactive and destructive).

   **7c. Inline strategy: lightweight review**

   The orchestrator runs the validation commands from step 7b directly rather than spawning a validation subagent.

   Always run `openspec-verify-change` for the same change. The orchestrator may run it itself; do not skip it on "straightforward" changes.

   For straightforward changes (config, docs, Makefile, CI, spec-only), that plus a self-review is sufficient. Do not spawn other review subagents.

   For changes that touch non-trivial logic (custom types, plan modifiers, complex CRUD, error handling), also spawn:
   - **Critical code review**: review for coding standards, idiomatic patterns, logic issues, error handling gaps
   - Add coverage review only if the change involves a Terraform entity (see 7d.d)

   **7d. Single-implementor strategy: one review round**

   After all tasks are complete, launch the full parallel review battery **once**:

   a. **Validation runner** - launch a dedicated validation subagent to run the validation requirements from step 7b (`make lint`, `make build`, `make unit`, and `make check-openspec` when `openspec/` changed). Use a subagent so these checks do not consume the orchestrator's working context.

   b. **Critical code review** - review for coding standards, idiomatic Go/Terraform Plugin Framework patterns, obvious logic issues, error handling gaps, and risky regressions. Return prioritized findings only.

   c. **Proposal compliance review** - run the `openspec-verify-change` skill/process for the same change. Return CRITICAL mismatches and missing work. Include WARNINGs and SUGGESTIONs in the report; they are not push-blocking (see step 8).

   d. **Coverage review for Terraform entities** - if this is a Terraform entity change and `.agents/skills/schema-coverage/SKILL.md` exists, run that skill. Focus on untested or weakly tested high-risk attributes and behaviors. If that skill is not installed, fall back to `go test -cover` on the touched packages (same as 7d.e) and say so.

   e. **Coverage review for non-entity changes** - if this is not a Terraform entity change, run a thorough test analysis instead. Prefer explicit coverage tooling where possible, for example `go test -cover`. Identify high-risk code paths that lack direct test coverage.

   Run the validation runner in parallel with the other review subagents, not as a separate serial phase.

   **7e. Per-task strategy: review after each top-level task**

   After each top-level task's implementor reports completion, launch validation runner, critical code review, and the appropriate coverage review for **that task's** diff.

   Do **not** run `openspec-verify-change` until every top-level task is complete. That skill treats remaining `- [ ]` tasks as CRITICAL, which would block advancing to the next task. After the **last** top-level task passes its task-scoped review, run the full battery from 7d once, including proposal compliance (`openspec-verify-change`).

   Run the validation runner in parallel with the other review subagents for the same top-level task.

   **For all review subagents**, ask them to return:
   - severity
   - concise finding
   - evidence
   - recommended fix

8. **Aggregate findings and decide whether to loop**

   Combine the validation results and review outputs.

   **Push-blocking** (must fix before advancing):
   - Failed step 7b commands (`make lint` / `build` / `unit` / `check-openspec`)
   - A **user-approved** targeted acc run that failed (do not retry acc; fix code, then ask again)
   - `openspec-verify-change` CRITICAL issues
   - Critical code-review findings that are actual defects

   **Not push-blocking** (report in the final summary; do not loop):
   - `openspec-verify-change` WARNINGs, including acc-only scenario coverage when the user skipped or was not asked
   - SUGGESTIONs
   - Coverage-gap notes that do not claim a missing requirement
   - Buildkite acc status (never auto-fix, never re-trigger)

   If there are no push-blocking findings:
   - **Inline / single-implementor / `all_done`**: proceed to push
   - **Per-task**: mark the current top-level task as locally complete; start a fresh implementor for the next incomplete top-level task, or proceed to push if none remain

   If there are push-blocking findings:
   - **Inline / `all_done`**: the orchestrator fixes the issues directly with minimal diffs and additional small focused commits
   - **Single-implementor**: resume the implementor subagent, give it the aggregated **push-blocking** findings, and ask it to fix them with minimal diffs and additional small focused commits
   - **Per-task**: resume the current top-level task's implementor subagent, give it the aggregated **push-blocking** findings, and ask it to fix them with minimal diffs and additional small focused commits

   After fixes, rerun validation and relevant reviews before advancing:
   - **Inline / single-implementor / `all_done`**: rerun before proceeding to push
   - **Per-task**: rerun before moving to the next top-level task

   Repeat until:
   - all tasks pass review and the loop reaches push readiness, or
   - the implementor (or orchestrator) becomes blocked, or
   - the same issue repeats without progress

   If the loop stalls, pause and ask the user how to proceed.

9. **Push the branch**

   After every incomplete top-level task has been implemented and passed local review:
   - verify the branch state is ready to push
   - push the current branch to `origin`
   - use upstream tracking if needed

   Example:
   ```bash
   git push -u origin HEAD
   ```

   **Guardrails**:
   - Never force-push unless the user explicitly asks
   - Do not push before local review passes

10. **Commit-only mode: watch GitHub Actions (branch / commits)**

    If the user chose **commit-only** in step 2:

    After pushing:
    - inspect workflow runs for the current branch and the **commits** you pushed (for example with `gh run list` filtered by branch, or `gh` against the latest commit SHA)
    - watch **GitHub Actions** that actually run on a branch push until they complete: typically `Go` (unit/lint/docs/NOTICE)
    - do **not** wait for `OpenSpec CI` in commit-only mode. That workflow's `push` trigger is `master` only; pull requests get it via `pull_request`. Structural spec validation is the local `make check-openspec` from step 7
    - **Do not** treat Buildkite acceptance as a blocking auto-fix loop. If an acceptance check is visible, report pass/fail to the user. Never re-run acc, never auto-fix acc failures (they create paid Elastic Cloud deployments)

    If GitHub Actions succeeds:
    - finish with a concise summary, including any out-of-band acc step that a human should watch (or continue to step 12 if you already reported)

    If GitHub Actions fails:
    - collect the failing workflow, job, and relevant log details
    - launch a fresh write-capable implementor subagent scoped only to resolving those **Actions** failures
    - ask it to fix the issues and commit the changes
    - rerun step 7 validation (and relevant reviews from step 8) before pushing again
    - then push and continue watching GitHub Actions

    Repeat until:
    - GitHub Actions is green, or
    - a failure cannot be resolved without user input

11. **PR mode: create PR, watch PR checks, poll reviews, address feedback**

    If the user chose **pull request** in step 2:

    **Create the PR after the initial push** (step 9), if it does not already exist:
    - use `gh pr create` (or equivalent) with an appropriate title and body tied to the OpenSpec change
    - record the PR number or URL

    **If `.agents/skills/pr-monitoring-loop/SKILL.md` exists**, delegate PR monitoring to that skill for the rest of this step (watcher/delegate subagents, state file, `verify-openspec` opt-in). Do not restate those rules here.

    **Otherwise** (skill not installed yet), monitor with `gh` from this agent or a single watcher subagent:
    - poll GitHub Actions on the PR (`gh pr checks`, `gh run list`) until `Go` / `OpenSpec CI` complete
    - poll reviews, PR comments, and review comments at a coarse cadence
    - fix **simple** GitHub Actions failures (lint/unit/docs) the same way as commit mode: small commits, **rerun step 7–8**, then push and re-watch Actions
    - **surface** Buildkite acceptance as out-of-band: report status if visible; never auto-fix acc failures; never re-trigger acc
    - do **not** apply a `verify-openspec` label unless that workflow exists in this repo
    - stop and ask the user when review feedback needs judgment, the branch is in merge conflict, or the loop stalls

    State-file and `.git/`-directory warnings from `pr-monitoring-loop` apply only when that skill is in use. This repo uses git worktrees; do not put watcher state under `.git/`.

12. **Report final outcome**

    Summarize:
    - change name
    - schema
    - delivery mode (commit-only vs PR)
    - execution strategy used (inline, single-implementor, or per-task) and reasoning
    - implementation/review/CI loop status
    - top-level tasks completed in the loop
    - commits created during the loop
    - local validation run during the loop (`make lint`, `make build`, `make unit`, and `make check-openspec` when `openspec/` changed)
    - targeted acc: skipped (no runtime / user declined / no `EC_API_KEY`) or `TEST_NAME=…` result; never imply the full suite ran
    - tests or coverage checks used
    - final GitHub Actions state (and PR link if PR mode); Buildkite acc status if known
    - PR review handling summary if PR mode
    - any remaining blockers or risks

**Recommended subagent responsibilities**

Subagent usage scales with the chosen strategy:

- **Implementor** (single-implementor and per-task strategies): makes code changes, updates tasks, runs targeted unit validation, and creates small focused commits. Per-task uses one fresh implementor per top-level task; single-implementor uses one for the entire change.
- **Validation runner** (single-implementor and per-task strategies): a subagent that runs `make lint`, `make build`, `make unit`, and `make check-openspec` when `openspec/` changed, then reports a concise validation summary. Use a subagent so validation does not consume the orchestrator's context. It MUST NOT set `TF_ACC` or run `make testacc`. Targeted acc is orchestrator-only after an explicit yes.
- **Critical reviewer**: reviews code quality and logic
- **Spec reviewer**: checks the implementation against the approved OpenSpec change
- **Coverage reviewer**: checks test coverage quality using the appropriate strategy
- **PR watcher**: when `pr-monitoring-loop` is installed, a fresh subagent using that skill; otherwise a watcher that polls `gh` as in step 11

For the **inline** strategy, the orchestrator fills the implementor and validation runner roles directly. It still runs `openspec-verify-change`. Other review subagents are spawned only when the change touches non-trivial logic.

**Guardrails**

- Operate on one change only
- Ask **commit vs PR** at the **start** (step 2), not when implementation is finished
- Always triage the change and announce the execution strategy before implementation, **except** when `state` is `all_done` (then skip 4–6 and run 7b + 7d once)
- The user can override the chosen strategy at the triage step
- Always read the OpenSpec context before implementation
- **Per-task strategy**: create a fresh implementor subagent for each top-level task; do not reuse one implementor across top-level tasks. Do not advance to the next top-level task until the current task has passed local review. Run `openspec-verify-change` only after the last top-level task.
- **Single-implementor strategy**: use one implementor for all tasks; run one review round after all tasks complete
- **Inline strategy**: the orchestrator implements directly and still runs `openspec-verify-change`; spawn other review subagents only for non-trivial logic changes
- Never archive a change in this workflow; only sync delta specs when needed. Archiving happens after this skill returns.
- Ignore tasks that request archiving the change proposal
- **Never auto-run acceptance tests.** No `TF_ACC` from implementors or the validation runner. No full `make testacc`. Targeted `TEST_NAME=…` only after an explicit yes, once per loop, never retried on failure. `openspec-verify-change` never runs acc.
- Implementor subagents (and step 6 work, including inline) never push; the orchestrator pushes in step 9 after local review passes
- For single-implementor and per-task strategies, run local validation in a dedicated subagent so the orchestrator does not spend its own context on lint/build/test execution
- Run the validation subagent in parallel with the other review subagents, not as a separate serial phase
- All strategies must include `make lint`, `make build`, and `make unit`, plus `make check-openspec` when `openspec/` changed
- After a GitHub Actions failure, rerun steps 7–8 before the next push
- Run reviewers in parallel whenever possible
- Prefer actionable findings over style nitpicks
- `openspec-verify-change` WARNINGs and SUGGESTIONs do not block push; only CRITICALs and failed 7b commands do
- Feed local review and commit-mode **GitHub Actions** failures back into the loop instead of fixing them ad hoc outside the loop. Do not feed Buildkite acc failures into an auto-fix loop.
- Keep commit sizes small and purpose-specific
- Stop and ask the user if the process becomes ambiguous or stuck
