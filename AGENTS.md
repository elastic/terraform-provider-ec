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

## Testing note — acceptance tests hit the real, paid Elastic Cloud API

- Acceptance tests (`make testacc`, and anything gated by `TF_ACC=1`) create and destroy **real
  deployments** against the live Elastic Cloud API (`EC_API_KEY`) and cost real money. **Never run
  acceptance tests from an agentic workflow** — no live-cloud credentials are exposed to agents. The
  full suite runs on Buildkite per PR and is a **required** status check on `master`
  (`buildkite/terraform-provider-ec-acceptance`); a human working on a change
  should run the targeted `TestAcc…` case(s) locally first. See [`testing.md`](./dev-docs/high-level/testing.md).
- There is **no local Docker stack** for this provider (unlike the Elastic Stack provider). Unit
  tests (`make unit`) need no credentials and are always safe to run.

## After making changes

- Build: `make build`
- Lint: `make lint`
- If you changed `openspec/`: `make check-openspec`
- Unit tests (no cloud, always safe): `make unit`
- If you changed resource/data-source schemas or examples, regenerate docs with `make docs-generate`
  and verify with `make tfproviderdocs`. See [`documentation.md`](./dev-docs/high-level/documentation.md).
- If you changed the serverless client inputs, regenerate with `make gen`. See
  [`generated-clients.md`](./dev-docs/high-level/generated-clients.md).
- Add a `.changelog/{PR}.txt` entry for user-facing changes (see [`contributing.md`](./dev-docs/high-level/contributing.md)).

> The OpenSpec CLI is installed via `make setup-openspec` (Node.js 24, `npm ci`) and validated by
> `make check-openspec` / `.github/workflows/openspec.yml`. Author specs per
> [`openspec-requirements.md`](./dev-docs/high-level/openspec-requirements.md). Lifecycle skills and
> GitHub Agentic Workflows land in later Phase 1–4 issues of the LLM-driven SDLC epic.
