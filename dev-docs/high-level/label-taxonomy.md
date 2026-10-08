# SDLC label taxonomy

Authoritative names, colors, and GitHub descriptions live in
[`.github/label-taxonomy.json`](../../.github/label-taxonomy.json). This page
explains the **roles** of those labels so Phase 3 factories and Phase 4 scanners
do not invent synonyms.

Create or refresh the labels on GitHub with:

```bash
./scripts/sync-sdlc-labels.sh --dry-run
./scripts/sync-sdlc-labels.sh
```

Validate the manifest and this page (no network):

```bash
make check-label-taxonomy
```

## Lifecycle (short)

1. **Classifier** (bot) applies exactly one `needs-*` routing label **and**
   `triaged` on a new or untriaged issue.
2. **Human** promotes by applying a factory trigger (`research-factory`,
   `change-factory`, `code-factory`, or `reproducer-factory`). Triggers are
   **consumed** (removed) when the factory starts; re-apply to re-run.
3. **Factory** (bot) sets exactly one `phase-*` label, swapping any prior
   `phase-*` (see `phase-label.js` in the Phase 2 helpers issue).
4. **Scanners** (bot) file issues already stamped with a topic label plus
   `triaged`, so the classifier skips them. Topic labels also drive issue-slot
   backpressure.
5. **`verify-openspec`** is a human-applied PR label that starts the OpenSpec
   verify gate; a deterministic step removes it when the gate starts.

`triaged` is sticky. It is **not** a factory trigger. Scanners apply it when
they file issues; the classifier applies it next to one `needs-*` label.

## Labels

The table below is generated from the manifest. Do not edit it by hand — change
the JSON and run `scripts/check-label-taxonomy.sh --write`.

<!-- BEGIN LABEL TAXONOMY TABLE -->
| Name | Category | Applied by | Lifecycle | Color | Description |
| --- | --- | --- | --- | --- | --- |
| `needs-research` | routing | issue-classifier (bot) | Exactly one needs-* per issue, with triaged. Not consumed; factory trigger is separate. | `#0d9488` | Bot routing: feature needs research. Human promotes with research-factory. |
| `needs-reproduction` | routing | issue-classifier (bot) | Exactly one needs-* per issue, with triaged. Not consumed; factory trigger is separate. | `#0f766e` | Bot routing: bug needs reproduction. Human promotes with reproducer-factory. |
| `needs-spec` | routing | issue-classifier (bot) | Exactly one needs-* per issue, with triaged. Not consumed; factory trigger is separate. | `#14b8a6` | Bot routing: ready for OpenSpec. Human promotes with change-factory. |
| `needs-human` | routing | issue-classifier (bot) | Exactly one needs-* per issue, with triaged. Stays until a human re-routes. | `#5eead4` | Bot routing: needs human judgment; not auto-promoted to a factory. |
| `triaged` | routing | issue-classifier (bot) or scanners when filing issues | Sticky. Applied with one needs-* (classifier) or with a topic label (scanners). Never a factory trigger. | `#134e4a` | Sticky marker: classifier or scanner already routed; skip re-classify. |
| `research-factory` | factory | human (maintainer) | Consumed (removed) when research-factory qualifies the trigger. Re-apply to re-run. | `#fd3b56` | Human trigger: run research-factory. Consumed when the factory starts. |
| `change-factory` | factory | human (maintainer) | Consumed (removed) when change-factory qualifies the trigger. Re-apply to re-run. | `#fdb096` | Human trigger: run change-factory. Consumed when the factory starts. |
| `code-factory` | factory | human (maintainer) or scanner dispatch | Consumed (removed) when code-factory qualifies the trigger. Re-apply to re-run. | `#e8f446` | Human trigger: run code-factory. Consumed when the factory starts. |
| `reproducer-factory` | factory | human (maintainer) | Consumed (removed) when reproducer-factory qualifies the trigger. Re-apply to re-run. | `#b60205` | Human trigger: run reproducer-factory. Consumed when the factory starts. |
| `phase-research` | phase | research-factory (bot) | Exactly one phase-* at a time. phase-label helper removes other phase-* labels. | `#aaaaaa` | Bot phase: research-factory done. Swapped for the next phase-* label. |
| `phase-reproduction` | phase | reproducer-factory (bot) | Exactly one phase-* at a time. phase-label helper removes other phase-* labels. | `#aaaaaa` | Bot phase: reproducer-factory done. Swapped for the next phase-* label. |
| `phase-specification` | phase | change-factory (bot) | Exactly one phase-* at a time. phase-label helper removes other phase-* labels. | `#aaaaaa` | Bot phase: change-factory done. Swapped for the next phase-* label. |
| `phase-coding` | phase | code-factory (bot) | Exactly one phase-* at a time. phase-label helper removes other phase-* labels. | `#aaaaaa` | Bot phase: code-factory done. Swapped for the next phase-* label. |
| `verify-openspec` | gate | human (maintainer) | Applied to a PR. Deterministic step removes it when openspec-verify-label starts. | `#9bf329` | Human gate: run openspec-verify on this PR. Removed when the gate starts. |
| `duplicate-code` | topic | duplicate-code-detector scanner (bot) | Topic label on scanner-filed issues. Used for issue-slots backpressure. | `#ededed` | Scanner topic: duplicate-code-detector finding. Pre-stamped with triaged. |
| `semantic-refactor` | topic | semantic-function-refactor scanner (bot) | Topic label on scanner-filed issues. Used for issue-slots backpressure. | `#ededed` | Scanner topic: semantic-function-refactor finding. Pre-stamped with triaged. |
| `schema-coverage` | topic | schema-coverage-rotation scanner (bot) | Topic label on scanner-filed issues. Used for issue-slots backpressure. | `#ededed` | Scanner topic: schema-coverage-rotation finding. Pre-stamped with triaged. |
| `flaky-test` | topic | flaky-test-catcher scanner (bot) | Topic label on scanner-filed issues. Used for de-duplication and slots. | `#ededed` | Scanner topic: flaky-test-catcher finding. Pre-stamped with triaged. |
| `serverless-spec` | topic | none yet (reserved) | Reserved name from the Phase 2 taxonomy. No workflow consumes it until #4391 lands. | `#ededed` | Reserved for serverless-spec-impact scanner (cp-hosted-team#4391). No consumer yet. |
<!-- END LABEL TAXONOMY TABLE -->

## Out of scope here

- **Auxiliary labels** some Phase 4 scanners also apply in the stack provider
  (`code-quality`, `automated-analysis`, `refactoring`, `testing`,
  `acceptance-tests`): add rows to the manifest when those workflows are ported,
  so GitHub does not invent random colors.
- **Repo settings** beyond labels (Actions permissions, “Allow GitHub Actions to
  create and approve pull requests”, branch protection, gh-aw secrets): not part
  of this taxonomy. They are not owned by another Phase 2–4 issue yet; track
  them separately when factories need them.
- **Manifest ↔ workflow name drift**: this check cannot yet assert that every
  label string under `.github/workflows` and `.github/scripts` exists in the
  manifest, because those consumers are not in the repo. Add that check with
  the Phase 2 helpers (`phase-label` / `producer-dispatch`) or the first Phase 3
  factory.

## Rename procedure

`gh label create --force` cannot rename. To rename: update the manifest, create
the new label with `./scripts/sync-sdlc-labels.sh`, manually migrate issues/PRs,
then delete the old label in the GitHub UI (or `gh label delete`).
