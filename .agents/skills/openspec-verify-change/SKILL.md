---
name: openspec-verify-change
description: Verify implementation matches change artifacts. Use when the user wants to validate that implementation is complete, correct, and coherent before archiving. Also used standalone by the later verify-openspec CI gate.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
---

Verify that an implementation matches the change artifacts (specs, tasks, design).

This skill is **hand-maintained** (not emitted by `make gen-openspec-skills`). Keep its directory name in `scripts/gen-openspec-skills.mjs` so regeneration does not delete it. The report shape (Summary, Issues by priority, Final assessment) is the contract the later `verify-openspec` workflow consumes.

**Cloud-provider constraints**

- Do **not** run `make testacc` or set `TF_ACC` as part of verification. This skill is also the later CI `verify-openspec` gate; acc stays out-of-band. The implementation loop never sets `TF_ACC`; in PR mode it asks whether the human ran named `TestAcc…` cases. See `dev-docs/high-level/testing.md`.
- Treat unit tests (`env -u TF_ACC make unit`, package `*_test.go` via `env -u TF_ACC go test`) and code paths as sufficient implementation evidence. Never run those without unsetting `TF_ACC`.
- If a scenario is covered **only** by an acceptance test under `ec/acc/`, record a WARNING that names the `TestAcc…` case and that acc is out-of-band — not a CRITICAL "unimplemented" finding. Do **not** name a function that unconditionally `t.Skip`s (for example `TestAccDeploymentTrafficFilter_UpgradeFrom0_4_1`); that is not coverage.
- Requirements describe Terraform / API-contract behavior, not a line-by-line Go transcription.

**Store selection:** Only if the **user names** a store, run `openspec store list --json`, pick that id, and pass `--store <id>` on CLI commands that take it. Treat that id as sticky for the rest of this run. Do **not** auto-discover a store because one is registered on the machine. Default for this repo: unscoped CLI against the nearest local `openspec/`.

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

When invoked **non-interactively** (CI / the later `verify-openspec` gate), require an explicit change name. If none is provided, fail with a clear error rather than prompting.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use **AskUserQuestion** (or an equivalent explicit user prompt) to let the user select.

   Show changes that have implementation tasks (tasks artifact exists).
   Include the schema used for each change if available.
   Mark changes with incomplete tasks as "(In Progress)".

   **IMPORTANT**: Do NOT guess or auto-select a change. When a user is present, always let them choose. When non-interactive, require the name and exit with an error.

2. **Check status to understand the schema**
   ```bash
   openspec status --change "<name>" --json
   ```
   Parse the JSON to understand:
   - `schemaName`: The workflow being used (e.g., "spec-driven")
   - `planningHome`, `changeRoot`, `artifactPaths`, and `actionContext`: path and scope context
   - Which artifacts exist for this change

3. **Get planning context and load artifacts**

   ```bash
   openspec instructions apply --change "<name>" --json
   ```

   This returns the change directory, `contextFiles` (artifact ID -> array of concrete file paths), optional `context`, optional `operationGuidance`, and `state`.

   Read all available artifacts from `contextFiles`.

   Treat `context` as required prompt-level input (the same contract as `openspec-apply-change`). Read it and apply relevant project facts, conventions, and constraints while verifying — for example changelog rules, Plugin Framework vs SDKv2, and the no-auto-run-`TF_ACC` constraint. Treat `operationGuidance` as optional additive advice; follow entries that apply. If `context` conflicts with this skill, an explicit user choice, or a CLI-controlled value, report the conflict and preserve the controlling value.

   **Handle states:**
   - If `state: "blocked"`: stop. Report the missing required artifact. Do **not** fall through to graceful degradation or produce an archivable report.
   - If `state: "all_done"`: proceed with verification (tasks should already be complete).
   - Otherwise: proceed.

4. **Initialize verification report structure**

   Create a report structure with three dimensions:
   - **Completeness**: Track tasks and spec coverage
   - **Correctness**: Track requirement implementation and scenario coverage
   - **Coherence**: Track design adherence and pattern consistency

   Each dimension can have CRITICAL, WARNING, or SUGGESTION issues.

5. **Verify Completeness**

   **Task Completion**:
   - If `contextFiles.tasks` exists, read every file path in it
   - Parse checkboxes: `- [ ]` (incomplete) vs `- [x]` (complete)
   - Count complete vs total tasks
   - If incomplete tasks exist:
     - Do **not** add a CRITICAL issue for a task whose only remaining action is to archive the change (`openspec-archive-change`, `openspec archive`, or equivalent). Note it as skipped: archiving is a later step, not this skill.
     - Do **not** add a CRITICAL issue for a task whose only remaining action is to sync/merge delta specs into canonical `openspec/specs/` (`openspec-sync-specs` or equivalent). Note it as deferred: sync is a later Land-specs step, not this skill. The implementation loop also ignores those tasks.
     - Incomplete tasks that **only** write/add a `TestAcc…` test file stay CRITICAL until the file exists **and** the checkbox is `- [x]`. Combined "add and run on Buildkite" tasks: CRITICAL until the file exists; once it exists, the remaining execute-acc portion is a WARNING (this skill never runs acc). Do not keep CRITICAL solely because the **combined** checkbox is still `- [ ]` after the file is present.
     - Regardless of checkbox state: if a task (including a checked `- [x]`) says to write/add a `TestAcc…` file or function, CRITICAL until that file and `func TestAcc…` exist. Only the out-of-band execute-acc portion may be WARNING.
     - Do **not** add a CRITICAL issue for an incomplete task that is **execute-acc only** (Human/Buildkite *runs* a named case, `make testacc`, or `TF_ACC=1`, with no write/add). Record a WARNING that names the case; this skill never runs acc.
     - Add CRITICAL issue for each other incomplete task
     - Recommendation: "Complete task: <description>" or "Mark as done if already implemented"

   **Spec Coverage**:
   - If delta specs exist in `contextFiles.specs`:
     - Extract all requirements (marked with "### Requirement:")
     - For each requirement:
       - Search codebase for keywords related to the requirement
       - Assess if implementation likely exists
     - If requirements appear unimplemented:
       - Add CRITICAL issue: "Requirement not found: <requirement name>"
       - Recommendation: "Implement requirement X: <description>"

6. **Verify Correctness**

   **Requirement Implementation Mapping**:
   - For each requirement from delta specs:
     - Search codebase for implementation evidence
     - If found, note file paths and line ranges
     - Assess if implementation matches requirement intent
     - If divergence detected:
       - Add WARNING: "Implementation may diverge from spec: <details>"
       - Recommendation: "Review <file>:<lines> against requirement X"

   **Scenario Coverage**:
   - For each scenario in delta specs (marked with "#### Scenario:"):
     - Check if conditions are handled in code
     - Check if **unit** tests exist covering the scenario
     - If the only coverage is an acceptance test under `ec/acc/`:
       - Add WARNING: "Scenario covered only by out-of-band acc: <TestAcc name>"
       - Recommendation: "Human/Buildkite runs <test>; consider a unit test if the behavior can be asserted without the live API"
     - If the scenario appears uncovered in both code and unit tests:
       - Add WARNING: "Scenario not covered: <scenario name>"
       - Recommendation: "Add test or implementation for scenario: <description>"

7. **Verify Coherence**

   **Design Adherence**:
   - If `contextFiles.design` exists:
     - Extract key decisions (look for sections like "Decision:", "Approach:", "Architecture:")
     - Verify implementation follows those decisions
     - If contradiction detected:
       - Add WARNING: "Design decision not followed: <decision>"
       - Recommendation: "Update implementation or revise design.md to match reality"
   - If no design.md: Skip design adherence check, note "No design.md to verify against"

   **Code Pattern Consistency**:
   - Review new code for consistency with project patterns (Plugin Framework, no SDKv2)
   - Check file naming, directory structure, coding style
   - If significant deviations found:
     - Add SUGGESTION: "Code pattern deviation: <details>"
     - Recommendation: "Consider following project pattern: <example>"

8. **Generate Verification Report**

   **Summary Scorecard**:
   ```
   ## Verification Report: <change-name>

   ### Summary
   | Dimension    | Status           |
   |--------------|------------------|
   | Completeness | X/Y tasks, N reqs|
   | Correctness  | M/N reqs covered |
   | Coherence    | Followed/Issues  |
   ```

   **Issues by Priority**:

   1. **CRITICAL** (Must fix before archive):
      - Incomplete tasks other than archive-only, sync-to-canonical-specs, or execute-acc Human/Buildkite steps
      - Incomplete or checked-but-missing write/add `TestAcc…` file/function tasks (file and `func TestAcc…` must exist; remaining execute-acc on a combined write/run task is WARNING once the file exists)
      - Missing requirement implementations
      - Each with specific, actionable recommendation

   2. **WARNING** (Should fix):
      - Spec/design divergences
      - Missing scenario coverage
      - Acc-only scenario coverage (out-of-band)
      - Incomplete execute-acc Human/Buildkite tasks (out-of-band; not push-blocking), including the remaining run portion of a combined write/run `TestAcc…` task after the file exists
      - Each with specific recommendation

   3. **SUGGESTION** (Nice to fix):
      - Pattern inconsistencies
      - Minor improvements
      - Each with specific recommendation

   **Final Assessment**:
   - This skill never archives or syncs. If invoked from the implementation loop, do **not** tell the caller to run `openspec-archive-change` or `openspec-sync-specs`.
   - If CRITICAL issues: "X critical issue(s) found. Fix before archiving."
   - If only warnings: "No critical issues. Y warning(s) to consider. Ready for archive (with noted improvements)." Standalone only; the loop ignores the archive sentence.
   - If all clear: "All checks passed. Ready for archive." Standalone only; the loop ignores the archive sentence.

**Verification Heuristics**

- **Completeness**: Focus on objective checklist items (checkboxes, requirements list)
- **Correctness**: Use keyword search, file path analysis, reasonable inference - don't require perfect certainty
- **Coherence**: Look for glaring inconsistencies, don't nitpick style
- **False Positives**: When uncertain, prefer SUGGESTION over WARNING, WARNING over CRITICAL
- **Actionability**: Every issue must have a specific recommendation with file/line references where applicable

**Graceful Degradation**

- If only tasks.md exists: verify task completion only, skip spec/design checks
- If tasks + specs exist: verify completeness and correctness, skip design
- If full artifacts: verify all three dimensions
- Always note which checks were skipped and why

**Output Format**

Use clear markdown with:
- Table for summary scorecard
- Grouped lists for issues (CRITICAL/WARNING/SUGGESTION)
- Code references in format: `file.go:123`
- Specific, actionable recommendations
- No vague suggestions like "consider reviewing"
