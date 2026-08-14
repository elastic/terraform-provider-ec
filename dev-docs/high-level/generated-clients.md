# Generated clients

The cloud provider includes **one** generated API client: the **serverless project API
client**. Everything under `ec/internal/gen/serverless/` — the HTTP client, its request/response
models, mocks, and the Terraform Plugin Framework schema code for the serverless resources — is
generated from a single committed OpenAPI spec. Do **not** hand-edit any of these files (see
[Regenerating](#regenerating)).

> The classic hosted (ESS) Elastic Cloud API is **not** generated here. It is served by the
> external [`cloud-sdk-go`](#not-generated-cloud-sdk-go) module — see the note at the bottom.

## Source of truth

The input is a **vendored** (committed), fully dereferenced OpenAPI file:

- `ec/internal/gen/serverless/serverless-project-api-dereferenced.yml` — the spec itself.
- `ec/internal/gen/serverless/serverless-project-api.source` — records where the spec came from
  (the upstream spec repo, the path within it, and the exact gitref); see that file for the
  specifics.

There is **no network fetch during generation**: `make gen` runs entirely against the committed
`…-dereferenced.yml`, so regeneration is fully deterministic (an unchanged spec produces no diff).
To pick up an upstream API change, refresh the vendored spec with
[`scripts/update-serverless-spec.sh`](../../scripts/update-serverless-spec.sh) — it downloads the
file at the ref recorded in `serverless-project-api.source` (authenticating via `GITHUB_TOKEN` or
the `gh` CLI) — then run `make gen` and commit the result.

## What is generated

All outputs live under `ec/internal/gen/serverless/`:

| Output | Produced by | Contents |
| --- | --- | --- |
| `client.gen.go` | `oapi-codegen` | Go HTTP client (`ClientInterface`, `ClientWithResponsesInterface`) and request/response models for the serverless project API |
| `mocks/client.gen.go` | `mockgen` | GoMock mocks of the two client interfaces above, used by unit tests |
| `resource_elasticsearch_project/`, `resource_observability_project/`, `resource_security_project/`, `resource_serverless_traffic_filter/` | `tfplugingen-framework` | Per-resource Terraform Plugin Framework schema + model code (`*_resource_gen.go`) |
| `provider_ec/ec_provider_gen.go` | `tfplugingen-framework` | Generated provider-level schema code |

Intermediate artifacts `spec.json` and `spec-mod.json` are also written under this directory (see
the chain below). They are inputs to later steps, not the client itself.

## The generation chain

The chain is defined as ordered `//go:generate` directives in
[`ec/internal/gen/serverless/serverless.go`](../../ec/internal/gen/serverless/serverless.go). They
run top-to-bottom:

1. **`tfplugingen-openapi generate --config oapi-config.yaml --output spec.json serverless-project-api-dereferenced.yml`**
   — reads the OpenAPI spec and the resource-to-endpoint mapping in `oapi-config.yaml`
   (which maps `elasticsearch_project`, `observability_project`, and `security_project` to their
   create/read/update/delete paths and methods) and emits a Plugin Framework provider-code spec,
   `spec.json`.
2. **`sh modify_spec.sh`** — applies `jq` transforms to `spec.json`, writing `spec-mod.json`
   (see [modify_spec.sh](#modify_specsh) below).
3. **`tfplugingen-framework generate all --input spec-mod.json --output .`** — generates the
   framework schema/model code (the `resource_*/` and `provider_ec/` directories) from the
   modified spec.
4. **`oapi-codegen --config=client-config.yaml serverless-project-api-dereferenced.yml`** —
   generates `client.gen.go` directly from the OpenAPI spec. `client-config.yaml` sets the Go
   package (`serverless`), the output file, and requests both `models` and `client` generation.
5. **`mockgen ... -destination=mocks/client.gen.go`** — generates mocks of
   `ClientWithResponsesInterface` and `ClientInterface` (from the freshly generated `client.gen.go`)
   into `mocks/client.gen.go`.

Note that steps 3 and 4 both start from the spec but feed different toolchains: step 3 produces the
Terraform schema code from the *modified* spec, while step 4 produces the HTTP client from the
*original* dereferenced spec.

### modify_spec.sh

[`modify_spec.sh`](../../ec/internal/gen/serverless/modify_spec.sh) exists because the framework
generator's raw output needs a few targeted fixups that are easier to express as `jq` patches than
to carry in the spec. It runs four transforms in sequence (`spec.json` → `spec-mod.json`):

1. **String plan modifiers** — for the attributes listed in `string_use_state_for_unknown.json`
   (`alias`, `cloud_id`, `id`, `type`, `optimized_for`), it injects a
   `stringplanmodifier.UseStateForUnknown()` plan modifier so Terraform keeps the prior state
   value instead of showing "known after apply".
2. **Custom product-types type** — on `security_project`'s `product_types` attribute it swaps in
   the custom list type/value defined in `product_type_custom_type.json`.
3. **Traffic filter attributes** — on every `*_project` resource it adds an optional
   `traffic_filter_ids` set attribute and removes the generated `traffic_filters` attribute
   (only `traffic_filter_ids` is wired into the hand-written code; `traffic_filters` is unused API
   cruft).
4. **Linked projects (cross-project search)** — on every `*_project` resource it makes the `linked`
   block `Optional`-only and splits the API-computed `status` out of the practitioner-controlled
   `linked.projects` map element into a sibling computed `linked.statuses` map (keyed by project ID).
   Keeping the computed value out of the configured map element lets the framework detect removal of
   a linked project (and thus unlink it) instead of preserving omitted map keys.

## Regenerating

Regenerate with:

```sh
make gen
```

The `gen` target (in [`build/Makefile.build`](../../build/Makefile.build)) runs `go generate ./...`
— which executes all five `//go:generate` directives above in order — and then re-applies the
Apache license headers via `make license-header`. So a single `make gen` covers the whole serverless
chain end-to-end. (`go generate ./...` also runs the repo's other generators, such as the
`ec/version.go` generator, so the command is not serverless-specific.) `make build` runs `gen` first,
so a plain build stays in sync too.

Because every file above is generated, **editing them by hand is wrong** — the change will be
overwritten on the next `make gen`. To change behavior, edit the real inputs instead:

- the vendored OpenAPI spec — refresh it with `scripts/update-serverless-spec.sh` rather than
  editing `serverless-project-api-dereferenced.yml` by hand (see [Source of truth](#source-of-truth)),
- the `oapi-config.yaml` / `client-config.yaml` generator configs, or
- the `modify_spec.sh` transforms (and the `*.json` snippets they pull in),

then rerun `make gen` and commit the regenerated output.

See [`development-workflow.md`](./development-workflow.md) for where this sits in the day-to-day
workflow and [`contributing.md`](./contributing.md) for PR conventions.

## Not generated: `cloud-sdk-go`

[`github.com/elastic/cloud-sdk-go`](https://github.com/elastic/cloud-sdk-go) is an **external Go
module** — a hand-maintained client for the classic hosted (ESS) Elastic Cloud API that backs most
of the provider's resources (deployments, extensions, traffic filters, etc.). It is a normal `go.mod`
dependency, pinned to a released version and bumped by Renovate. It is **not** part of the codegen
described here: `make gen` neither reads nor writes it, and it lives in the module cache like any
other dependency. The serverless project API client above is the only client generated in this repo.
