# Agent guide (start here)

This repo is the **Terraform provider for Elastic Cloud** (the "cloud provider"), written in Go.
It manages Elastic Cloud hosted deployments, serverless projects, traffic filters, extensions,
and related resources through the Elastic Cloud API.

## Before making changes

- Follow the project's coding conventions in [`coding-standards.md`](./dev-docs/high-level/coding-standards.md).
- For contributor setup, PR flow, and the per-PR `.changelog/` convention, see [`contributing.md`](./dev-docs/high-level/contributing.md).

## High-level dev docs

- Repo orientation and where code lives: [`repo-structure.md`](./dev-docs/high-level/repo-structure.md)
- Common workflows and "what to do when": [`development-workflow.md`](./dev-docs/high-level/development-workflow.md)
- Testing (unit + acceptance) and required env: [`testing.md`](./dev-docs/high-level/testing.md)
- Generated clients (serverless OpenAPI) and regeneration: [`generated-clients.md`](./dev-docs/high-level/generated-clients.md)
- Documentation generation (`tfplugindocs`): [`documentation.md`](./dev-docs/high-level/documentation.md)
- OpenSpec authoring (Purpose / SHALL-MUST / Scenarios): [`openspec-requirements.md`](./dev-docs/high-level/openspec-requirements.md)
- OpenSpec change loop (explore / propose / apply / loop / verify / sync / archive): [`openspec-workflows.md`](./dev-docs/high-level/openspec-workflows.md)

## Testing note — acceptance tests hit the real, paid Elastic Cloud API

- Acceptance tests (`make testacc`, and anything gated by `TF_ACC=1`) create and destroy **real
  deployments** against the live Elastic Cloud API (`EC_API_KEY`) and cost real money. **Never
  auto-run** acceptance tests from an agentic workflow — no live-cloud credentials are exposed to
  agents by default. The `openspec-implementation-loop` **never** sets `TF_ACC`. In PR mode it
  recommends the human run named `TestAcc…` cases locally and confirms before `gh pr create`
  (skip = wait on Buildkite);
  implementors, `openspec-verify-change`, and CI reuse stay acc-free. The full suite runs on
  Buildkite per PR and is a **required** status check on
  `master` (`buildkite/terraform-provider-ec-acceptance`); a human working on a change should run
  the targeted `TestAcc…` case(s) locally first. See [`testing.md`](./dev-docs/high-level/testing.md).
- There is **no local Docker stack** for this provider (unlike the Elastic Stack provider). Unit
  tests (`env -u TF_ACC make unit`) need no credentials and are always safe to run.

## After making changes

- Build: `env -u TF_ACC make build`
- If you changed resource/data-source schemas, templates, or examples, regenerate docs with `env -u TF_ACC make docs-generate`
  and verify with `make tfproviderdocs`. See [`documentation.md`](./dev-docs/high-level/documentation.md).
- Lint: `env -u TF_ACC make lint`
- `env -u TF_ACC make notice` (commit if `NOTICE` is dirty; Go CI fails on a dirty NOTICE)
- If you changed `openspec/`: `env -u TF_ACC make check-openspec`
- Unit tests (no cloud, always safe): `env -u TF_ACC make unit`
- If this is a Terraform entity change, or examples/provider schemas changed: `env -u TF_ACC make install validate-examples`
- If you changed the serverless client inputs, regenerate with `env -u TF_ACC make gen`. See
  [`generated-clients.md`](./dev-docs/high-level/generated-clients.md).
- Add a `.changelog/{PR}.txt` entry for user-facing changes (see [`contributing.md`](./dev-docs/high-level/contributing.md)).

> The OpenSpec CLI is installed via `make setup-openspec` (Node.js 24, `npm ci`) and validated by
> `make check-openspec` / `.github/workflows/openspec.yml`. Author specs per
> [`openspec-requirements.md`](./dev-docs/high-level/openspec-requirements.md). Drive a change with
> the skills under [`.agents/skills/`](./.agents/skills/) — see
> [`openspec-workflows.md`](./dev-docs/high-level/openspec-workflows.md). GitHub Agentic Workflows
> land in later phases of the LLM-driven SDLC epic.
