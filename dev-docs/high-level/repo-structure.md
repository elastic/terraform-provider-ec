# Repo structure

A map of where things live in the cloud provider (`terraform-provider-ec`), so a human or a
coding agent can orient quickly. This page stays high-level and links to the deeper docs; for
build/test commands see [`development-workflow.md`](./development-workflow.md) and
[`testing.md`](./testing.md).

The Go module is `github.com/elastic/terraform-provider-ec`. The provider is built entirely on
the [terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework)
(no legacy SDKv2). `main.go` is the entrypoint: it serves `ec.New(ec.Version)` via
`providerserver.Serve` under the address `registry.terraform.io/elastic/ec`.

## Top level

| Path | What it is |
|------|------------|
| `main.go` | Provider entrypoint (`providerserver.Serve`). |
| `ec/` | All provider code — resources, data sources, and internals (see below). |
| `gen/` | `gen.go`, a small `go:generate` program that writes `ec/version.go` from the `Makefile` `VERSION`. |
| `Makefile` + `build/` | The root `Makefile` `include`s the split fragments under `build/` (`Makefile.build`, `.test`, `.dev`, `.deps`, `.lint`, `.format`, `.release`, `.version`) plus `scripts/Makefile.help`. |
| `scripts/` | Helper scripts (changelog, version bump, `Makefile.help`, etc.). |
| `templates/` + `docs/` | Doc **sources** (`templates/`) and the **generated** registry docs (`docs/`). See [`documentation.md`](./documentation.md). |
| `examples/` | Example Terraform configs, also pulled into the generated docs. |
| `.changelog/` | Per-PR changelog fragments (`{PR}.txt`); consolidated into `CHANGELOG.md` at release. See [`contributing.md`](./contributing.md). |
| `.buildkite/` | Buildkite pipelines — notably the per-PR **acceptance** pipeline. See [`testing.md`](./testing.md). |
| `.github/` | GitHub Actions (unit/lint/docs via `go.yml`) and repo config. |
| `dev-docs/` | Developer docs (this set), including [`RELEASE.md`](../RELEASE.md) — the release runbook. |
| `docs-elastic/` | AsciiDoc source (`index.asciidoc`) for the Elastic docs site. |
| `tools/` | Tool dependencies (pinned via `go` tooling). |
| `CONTRIBUTING.md`, `README.md`, `NOTICE`, `LICENSE` | Standard top-level files. |

## Inside `ec/`

The provider itself is registered in `ec/provider.go` (`Resources()` / `DataSources()`), with
provider-level config in `ec/provider_config.go` and shared plumbing in `ec/internal/`.

### Resources and data sources

Resources live in `ec/ecresource/<name>resource/` and data sources in
`ec/ecdatasource/<name>datasource/` — one package each. The **authoritative, always-current list**
is the provider's own registry: `Resources()` and `DataSources()` in
[`ec/provider.go`](../../ec/provider.go). Consult that rather than a hand-maintained list here (the
surface is small — roughly a dozen resources and a handful of data sources).

A few landmarks worth knowing:

- `deploymentresource` (`ec_deployment`) is by far the largest package — schema versions plus
  sub-resources for the Elasticsearch topology, Kibana, APM, Integrations Server, etc.
- `projectresource` is a single package that registers **three** serverless project types
  (`ec_elasticsearch_project`, `ec_observability_project`, `ec_security_project`).
- `privatelinkdatasource` provides the three PrivateLink endpoint data sources (AWS / GCP / Azure).

Each resource and data-source package follows the same Schema / Model / CRUD split described in
[`coding-standards.md`](./coding-standards.md).

### Internals — `ec/internal/*`

- `converters/` — helpers to convert between API and framework types.
- `planmodifiers/` — custom plan modifiers.
- `validators/` — custom schema validators.
- `util/` — shared utilities.
- `serverlesshttp/` — an `http.RoundTripper` that retries serverless-API `429` responses with
  bounded backoff + jitter; wired into the generated serverless client in `ec/provider.go`.
- `gen/serverless/` — the **generated** serverless project API client (`client.gen.go`, mocks,
  and framework code). This is codegen output — do not hand-edit it. See
  [`generated-clients.md`](./generated-clients.md).

### Acceptance tests — `ec/acc/`

Live-cloud acceptance tests plus `ec/acc/testdata/`. These provision real deployments and are
**not** run locally or by agents — see [`testing.md`](./testing.md).

## Two API clients (don't confuse them)

- **Classic / hosted ESS** calls go through [`cloud-sdk-go`](https://github.com/elastic/cloud-sdk-go),
  an external, hand-maintained Go module (bumped by Renovate). It is a normal dependency, **not**
  generated in this repo.
- **Serverless project** calls go through the client generated under `ec/internal/gen/serverless/`
  from a committed OpenAPI spec. See [`generated-clients.md`](./generated-clients.md).
