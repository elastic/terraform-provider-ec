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

## Testing note — acceptance tests hit the real, paid Elastic Cloud API

- Acceptance tests (`make testacc`, and anything gated by `TF_ACC=1`) create and destroy **real
  deployments** against the live Elastic Cloud API (`EC_API_KEY`) and cost real money. **Never run
  acceptance tests from an agentic workflow** — no live-cloud credentials are exposed to agents. The
  full suite runs on Buildkite per PR (a human reviews the result); a human working on a change
  should run the targeted `TestAcc…` case(s) locally first. See [`testing.md`](./dev-docs/high-level/testing.md).
- There is **no local Docker stack** for this provider (unlike the Elastic Stack provider). Unit
  tests (`make unit`) need no credentials and are always safe to run.

## After making changes

- Build: `make build`
- Lint: `make lint`
- Unit tests (no cloud, always safe): `make unit`
- If you changed resource/data-source schemas or examples, regenerate docs with `make docs-generate`
  and verify with `make tfproviderdocs`. See [`documentation.md`](./dev-docs/high-level/documentation.md).
- If you changed the serverless client inputs, regenerate with `make gen`. See
  [`generated-clients.md`](./dev-docs/high-level/generated-clients.md).
- Add a `.changelog/{PR}.txt` entry for user-facing changes (see [`contributing.md`](./dev-docs/high-level/contributing.md)).

> The OpenSpec requirements layer and the GitHub Agentic Workflows (issue intake, quality scans,
> and PR verification) are being introduced in later phases of the LLM-driven SDLC epic. Their
> docs and links will be added here as those land.
