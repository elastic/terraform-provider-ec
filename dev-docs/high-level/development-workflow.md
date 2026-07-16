# Development workflow

A "what to do when" guide for the cloud provider (`terraform-provider-ec`), anchored on `make`
targets. **Run `make help` for the full, current list of targets with descriptions** — this page
covers only the non-obvious bits and the typical change loop, so it doesn't go stale as targets
change. For contributor setup and PR expectations see [`contributing.md`](./contributing.md); for
where code lives see [`repo-structure.md`](./repo-structure.md).

The root `Makefile` just `include`s split fragments under `build/` (`Makefile.build`, `.test`,
`.dev`, `.deps`, `.lint`, `.format`, `.release`, `.version`) plus `scripts/Makefile.help`; those
fragments are the source of truth for exact behavior.

> **Acceptance tests cost real money.** `make testacc` (and anything gated by `TF_ACC=1`) creates and
> destroys **real, paid Elastic Cloud deployments** via `EC_API_KEY` and can run up to ~2 hours. They
> run automatically for every PR on the **Buildkite acceptance pipeline** (a human reviews the
> result), so you rarely need to run them yourself. If you do run them locally, mind the cost and
> `make sweep` any leftovers; and **never run them from an agentic workflow** — agents rely on the
> Buildkite run rather than provisioning paid infrastructure. See [`testing.md`](./testing.md).
> `make unit` needs no credentials and is always safe. There is **no local Docker stack** for this
> provider.

## Worth knowing (beyond `make help`)

- **`make build` runs `make gen` first**, so a plain build regenerates code before compiling to
  `bin/terraform-provider-ec`; `make install` then copies the binary into your local Terraform
  plugin path.
- **`make lint` is the umbrella check** — Go and Terraform-provider linters, license-header check,
  generated-docs validation, and `.tf` formatting. `make format` applies the fixes. Run `make lint`
  before opening a PR.
- **`make gen`** (alias `make generate`) regenerates the serverless client *and* `ec/version.go`. To
  refresh the vendored serverless OpenAPI spec, use `scripts/update-serverless-spec.sh` — see
  [`generated-clients.md`](./generated-clients.md).
- **`make docs-generate`** whenever you change a resource/data-source schema or `examples/`; it's
  validated by `make tfproviderdocs` (part of `make lint`). See [`documentation.md`](./documentation.md).
- **`make unit`** (alias `make tests`) is the safe, credential-free test target; scope it with
  `TEST=./ec/...` and `TESTARGS=...`.
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
2. `make lint` — Go + provider linters, license headers, docs check, `.tf` formatting.
3. `make unit` — safe unit tests, no credentials.
4. `make docs-generate` — only if you changed resource/data-source schemas or `examples/`.
5. Add a changelog entry at `.changelog/{PR}.txt` for any user-facing change (one file per PR; see
   [`contributing.md`](./contributing.md)).

**Agents must not run `make testacc`.** Rely on the Buildkite acceptance run and human review instead.
