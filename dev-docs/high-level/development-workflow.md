# Development workflow

A "what to do when" guide for the cloud provider (`terraform-provider-ec`), anchored on `make`
targets. **Run `make help` for the full, current list of targets with descriptions** — this page
covers only the non-obvious bits and the typical change loop, so it doesn't go stale as targets
change. For contributor setup and PR expectations see [`contributing.md`](./contributing.md); for
where code lives see [`repo-structure.md`](./repo-structure.md).

The root `Makefile` just `include`s split fragments under `build/` (`Makefile.build`, `.test`,
`.dev`, `.openspec`, `.deps`, `.lint`, `.format`, `.release`, `.version`) plus `scripts/Makefile.help`; those
fragments are the source of truth for exact behavior.

> **Acceptance tests hit the real, paid Elastic Cloud API.** `make testacc` (anything gated by
> `TF_ACC=1`) provisions and destroys real deployments/projects via `EC_API_KEY` and costs money.
> **Before opening a PR, run the _targeted_ test(s) covering your change locally** —
> `make testacc TEST_NAME='^TestAccMyThing$'` — for fast feedback, then `make sweep` any leftovers.
> Don't run the **full** suite locally for routine iteration (~2 hours); the Buildkite acceptance
> pipeline runs the full suite for every PR and is a required status check on `master`. **Agents
> never run acceptance tests.** The implementation loop never sets `TF_ACC`. In PR mode it
> recommends the human run named `TestAcc…` cases locally and confirms before `gh pr create`
> (skip = wait on Buildkite). Implementors, `openspec-verify-change`, and CI reuse
> never set `TF_ACC`. No live-cloud credentials are exposed to agentic workflows by default. See
> [`testing.md`](./testing.md). `env -u TF_ACC make unit` needs no credentials and is always safe
> (plain `make unit` with inherited `TF_ACC=1` runs `ec/acc`). There is **no
> local Docker stack** for this provider.

## Worth knowing (beyond `make help`)

- **`make build` runs `make gen` first**, so a plain build regenerates code before compiling to
  `bin/terraform-provider-ec`; `make install` then copies the binary into your local Terraform
  plugin path.
- **`make setup-openspec`** installs the OpenSpec CLI (`npm ci`). Node.js 24 is required only for
  OpenSpec, not for `make lint` or the provider build. Go modules remain `make vendor`.
- **`make lint` is the Go/Terraform umbrella** — golangci-lint, license-header check, provider
  linters, generated-docs validation, and `.tf` formatting. `make format` applies those fixes. Run
  `make lint` before opening a PR.
- **`make check-openspec`** structurally validates `openspec/` (`openspec validate --all`). It is
  not part of `make lint`; CI runs it in `.github/workflows/openspec.yml`. Run it locally when you
  change specs. It installs the CLI via `setup-openspec` if needed. Authoring conventions:
  [`openspec-requirements.md`](./openspec-requirements.md). Which skill to use for a change:
  [`openspec-workflows.md`](./openspec-workflows.md). After an OpenSpec CLI bump, regenerate the
  seven lifecycle skills with **`make gen-openspec-skills`** — not `openspec init` or
  `openspec update`. That target leaves the hand-written `openspec-implementation-loop` and
  `openspec-verify-change` skills in place. Until 1.9, call the pinned CLI as `npx openspec` or
  `./node_modules/.bin/openspec`.
- **`make gen`** (alias `make generate`) regenerates the serverless client *and* `ec/version.go`. To
  refresh the vendored serverless OpenAPI spec, use `scripts/update-serverless-spec.sh` — see
  [`generated-clients.md`](./generated-clients.md).
- **`make docs-generate`** whenever you change a resource/data-source schema, a template, or `examples/`; it's
  validated by `make tfproviderdocs` (part of `make lint`). See [`documentation.md`](./documentation.md).
- **`make unit`** (alias `make tests`) is the safe, credential-free test target when `TF_ACC` is
  unset; agents use `env -u TF_ACC make unit`. Scope it with `TEST=./ec/...` and `TESTARGS=...`.
- **`make sweep`** is a cleanup tool for *leaked* cloud resources (it prompts for confirmation), not
  part of the normal loop — see [`testing.md`](./testing.md).
- **`make vendor`** also regenerates `NOTICE` (via `make notice`) after tidying modules.

## Releasing

Releases are usually driven through the [`/release` skill](../../.agents/skills/release/SKILL.md)
(version bump, `.changelog/` consolidation, prep PR). Version bumps are `make major` / `make minor` /
`make patch`; `make tag` pushes the release tag and triggers the Buildkite release pipeline. For the
full manual runbook see [`../RELEASE.md`](../RELEASE.md).

## Recommended pre-PR local loop

1. `make build` — regenerates code and compiles.
2. `make docs-generate` — only if you changed resource/data-source schemas, templates, or `examples/` (before
   lint so new entity docs exist for `tfproviderdocs`; then commit the generated `docs/` so the tree is clean).
3. `make lint` — Go + provider linters, license headers, docs check, `.tf` formatting.
4. `env -u TF_ACC make unit` — unit tests; unset `TF_ACC` so `ec/acc` does not run.
5. `make check-openspec` — only if you changed `openspec/` (CI also runs this in `openspec.yml`).
6. Run the **targeted** acceptance test(s) covering your change —
   `make testacc TEST_NAME='^TestAcc…$'` — then `make sweep` any leftovers. Skip if the change has no
   runtime behavior (docs/config only).
7. Add a changelog entry at `.changelog/{PR}.txt` for any user-facing change (one file per PR; see
   [`contributing.md`](./contributing.md)).

The **full** acceptance suite runs on Buildkite for every PR and must pass before merge; run only
the targeted cases locally. Agents never run acceptance tests. The implementation loop never sets
`TF_ACC`; in PR mode it recommends named local cases and confirms before `gh pr create`.
