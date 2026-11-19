---
name: new-entity-requirements
description: Gather requirements for a brand-new Terraform resource or data source from Elastic Cloud APIs and user decisions, then write an OpenSpec change (proposal, design, tasks, delta specs)—not a canonical spec under openspec/specs/. Use when adding a new entity that does not exist in the provider yet. Do not use to document an existing type (existing-entity-requirements) or to backfill the surface.
---

# New entity requirements

Gather **initial requirements** for an entirely new Terraform resource or data source, then **materialize them as an OpenSpec change**: `openspec/changes/<name>/` with `proposal.md`, `design.md`, `tasks.md`, and delta capability specs. Sources: API clients in this repo, Elastic API docs (Elastic docs MCP when available, otherwise web), and the user for decisions code and docs cannot answer.

Do **not** write unimplemented behavior into `openspec/specs/`.

For the **CLI sequence** (create change, artifact order, `openspec instructions`), follow [openspec-propose](../openspec-propose/SKILL.md). This skill adds **what to research** and **what to put in each artifact**.

## Input

- **Entity concept**: the user names the target (e.g. “serverless snapshot policy”, “hosted extension variant”). Optionally: resource vs data source, proposed type name (`ec_…`).
- **API scope**: hosted Elastic Cloud (`cloud-sdk-go`) vs serverless (`ec/internal/gen/serverless/`), and an API name or doc URL if known.
- **Change name** (optional): kebab-case id for `openspec new change`. If missing, derive one and confirm when ambiguous.

## Workflow

### 1. Resolve API surface

Client map: [`repo-structure.md`](../../../dev-docs/high-level/repo-structure.md) (two API clients) and [`generated-clients.md`](../../../dev-docs/high-level/generated-clients.md) (serverless OpenAPI). Details in [reference.md](reference.md).

- **Hosted ESS / ECE**: `github.com/elastic/cloud-sdk-go` — typically `pkg/api/…`. Configure uses `internal.ProviderClients.Stateful`.
- **Serverless projects**: generated client `ec/internal/gen/serverless/` (`ClientWithResponsesInterface`). Configure uses `ProviderClients.Serverless`. Do not hand-edit generated files; a new serverless endpoint may need a spec refresh + `make gen` (human/follow-up).

Identify create/update (PUT/POST), read (GET), delete, request/response shapes, identifiers. Note feature flags if any.

If the request is a **slice of `ec_deployment`**, the capability id names the slice (`deployment-integrations-server`), not `deployment`. Confirm that with the user.

### 2. Examine Elastic API docs

- **Preferred**: Elastic docs MCP when configured.
- **Fallback**: web fetch/search. URL patterns in [reference.md](reference.md).
- Extract: endpoints, required vs optional fields, validation, errors (404), whether create and update share an API.
- If the user gave a doc URL, use it.

### 3. Create the OpenSpec change

1. `openspec new change "<name>"` (pinned CLI: `npx openspec` / `./node_modules/.bin/openspec` after `make setup-openspec`).
2. Follow [openspec-propose](../openspec-propose/SKILL.md) until every `applyRequires` artifact is `done`.
3. Entity-specific content:
   - **proposal.md**: what and why; scope / non-goals; doc URLs.
   - **design.md**: API ↔ Terraform mapping; which client; identity; import; error handling; Plugin Framework (no SDKv2). Sources note (client paths, docs, MCP vs web).
   - **tasks.md**: schema, CRUD, **unit** tests, docs. Acc tests (`ec/acc/`, `TestAcc…`) are a **separate human/Buildkite** task — never instruct an agent to run `make testacc` or set `TF_ACC`. Changelog `.changelog/{PR}.txt` for user-facing implementation.
   - **Delta spec(s)**: per [`openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md). Include `Resource implementation:` / `Data source implementation:` as the intended package path. SHALL/MUST + scenarios. Mark unknowns TBD. Freeze heading names — later deltas match on that text.

### 4. Interview the user

Unanswered questions from the draft: use the bank in [reference.md](reference.md). Prefer a structured question tool (`AskQuestion` / `AskUserQuestion`) when available; otherwise ask in chat. Record answers in the delta spec (and proposal/design as needed). Deferred decisions stay TBD with an Open point.

### 5. Finalize

Every requirement comes from the API/client/docs or a user answer. No invented behavior. `openspec status --change "<name>"` apply-ready. `make check-openspec` (or `openspec validate --all`).

## Output

- **Deliverable**: a complete change under `openspec/changes/<name>/`, not a file only under `openspec/specs/`.
- **Traceability**: requirements tied to API docs or user decisions.

## Reference

- CLI: [openspec-propose](../openspec-propose/SKILL.md)
- Authoring: `dev-docs/high-level/openspec-requirements.md`
- Resource exemplar (style only, after drafting the delta): `openspec/specs/deployment-traffic-filter/spec.md`
- Data-source exemplar (style only, after drafting the delta): `openspec/specs/stack/spec.md`
- Clients, docs URLs, interview bank: [reference.md](reference.md)
