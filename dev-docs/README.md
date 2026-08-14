# Developer docs

## High level

Starting-point docs for contributors and coding agents. They stay high-level and link to the
canonical sources elsewhere in the repo:

* [Repo structure](./high-level/repo-structure.md) — where code lives
* [Development workflow](./high-level/development-workflow.md) — day-to-day tasks and `make` targets
* [Testing expectations](./high-level/testing.md) — unit vs. acceptance, live-cloud creds, Buildkite
* [Coding standards](./high-level/coding-standards.md)
* [Generated clients](./high-level/generated-clients.md) — the serverless OpenAPI codegen chain
* [Documentation generation](./high-level/documentation.md) — `tfplugindocs`
* [Contributing](./high-level/contributing.md) — pointer to the top-level `CONTRIBUTING.md`

## Release

* [Release runbook](./RELEASE.md) — how to cut and publish a new provider version

> The OpenSpec requirements layer and the GitHub Agentic Workflows (factories / scanners /
> verify gate) are being introduced in later phases of the LLM-driven SDLC epic; their docs will
> be added here as those land.
