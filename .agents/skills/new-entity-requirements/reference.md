# New entity requirements — clients, docs, interview bank

Do not copy path tables from [`repo-structure.md`](../../../dev-docs/high-level/repo-structure.md) or [`generated-clients.md`](../../../dev-docs/high-level/generated-clients.md); those are the source of truth. This file is the interview bank plus how to use docs.

## API clients (pointers)

| Surface | Client | Configure field |
| --- | --- | --- |
| Hosted ESS / ECE | [`cloud-sdk-go`](https://github.com/elastic/cloud-sdk-go) (`pkg/api/…`) | `internal.ProviderClients.Stateful` |
| Serverless projects | Generated `ec/internal/gen/serverless/` (`ClientWithResponsesInterface`) | `internal.ProviderClients.Serverless` |

`internal.ConvertProviderData` is how every resource/data source `Configure` unwraps provider data. There is no per-resource connection block.

**Finding a hosted API**: search `github.com/elastic/cloud-sdk-go` usage in this repo (e.g. `trafficfilterapi.Create`) or the module itself for the resource name.

**Finding a serverless API**: search `ec/internal/gen/serverless/` (generated types and `ClientWithResponses` methods). Wrappers live in `ec/ecresource/projectresource/` and `ec/ecresource/serverlesstrafficfilterresource/`. New operations may require refreshing the vendored OpenAPI spec (`scripts/update-serverless-spec.sh`) then `make gen` — call that out in design/tasks; do not hand-edit generated files.

## Elastic API documentation

### Elastic docs MCP (preferred)

When an MCP server exposes Elastic-docs search/fetch, use it for ESS, ECE, and serverless project APIs. Prefer it over ad-hoc web fetch when it returns current content.

### URL patterns (fallback and citations)

- **Elastic Cloud API (ESS)**: `https://www.elastic.co/docs/api/doc/cloud/` (and topic pages under that tree).
- **Cloud / ECE guides**: `https://www.elastic.co/guide/en/cloud/current/` and `https://www.elastic.co/guide/en/cloud-enterprise/current/`.
- **Serverless**: Elastic Cloud Serverless project APIs (search by product; the committed OpenAPI is `ec/internal/gen/serverless/serverless-project-api-dereferenced.yml`).

Existing `docs/resources/*.md` / `docs/data-sources/*.md` often already cite the right page — reuse that style.

### What to pull from docs

Endpoint and method; required vs optional body fields; identifier(s) returned; 404 and validation errors; notes that create and update share an API or do not.

## Interview question bank

Prefer a structured question tool (`AskQuestion` / `AskUserQuestion`) with concrete options when available; otherwise ask in chat.

### Identity and import

- **Identifier**: “How should Terraform identify this in state? (A) API-generated id only, (B) name, (C) composite, (D) other.”
- **Import**: “Support import? If yes, which import id format?”

### Schema and lifecycle

- **Required vs optional**: “For [field X]: required, optional with default, or computed-only?”
- **Replacement**: “When [name/region/…] changes, replace or update in place?”
- **Sensitive**: “Which attributes are sensitive (tokens, keys)?”

### Client and product

- **Hosted vs serverless**: “Does this call cloud-sdk-go (hosted ESS/ECE) or the generated serverless client? (This decides Configure, package layout, and `make gen`.)”
- **Resource vs data source**: “Managed resource or read-only data source?”
- **Type name**: “Proposed `ec_…` type. Confirm or suggest another.”
- **Slice of `ec_deployment`**: “Is this a new type, or a slice of `ec_deployment`? If a slice, capability id is the slice (`deployment-…`), not `deployment`.”

### State and mapping

- **Empty vs null**: “When the API returns empty for [field], store null or empty? (Drift.)”
- **Read-only**: “Which response fields are computed-only?”
- **Nested vs JSON**: “Expose [object] as nested blocks/attributes or a JSON string?”
- **State upgrade**: “Does the first schema need `Version` / `UpgradeState`, or is v0 enough until a breaking schema change?” Default to no version bump. `ec_deployment` already has Version 2 **without** an upgrader — do not copy that gap.

### Deferred

Leave TBD and an Open point in the delta spec.

Record each answer in the delta spec (schema notes, requirements, or Open points).
