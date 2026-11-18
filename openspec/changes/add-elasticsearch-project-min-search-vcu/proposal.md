## Why

Elastic Cloud serverless Elasticsearch projects can now pin a lower bound on Search tier autoscaling (`search_lake.min_search_vcu` on the project API, GA on 2026-11-16). The Cloud UI and the API expose it, but `ec_elasticsearch_project` does not, so customers who manage projects with Terraform cannot set, change, or remove the floor without the console or raw API calls.

## What Changes

- Add an optional integer attribute `min_search_vcu` inside the existing `search_lake` nested attribute of `ec_elasticsearch_project`.
- Practitioners set a whole number of VCUs (at least 2). The provider validates only the lower bound; the upper bound (4992 in the API specification, possibly lower for some organizations) is enforced by the project API and surfaced as an apply-time error diagnostic.
- Removing `min_search_vcu` from a configured `search_lake` block (or setting it to `null`) unsets the floor by sending JSON `null` for that key only. `search_power` and `boost_window` are not touched.
- Omitting the whole `search_lake` block leaves the block unmanaged, as today: the plan shows no change to `search_lake`, the provider sends nothing for it, and state stores whatever the API returns, including a floor set outside Terraform. The provider plans the omitted block from prior state itself, because once a non-computed child holds a value Terraform core would otherwise propose the block as null and show it as "known after apply" on every plan.
- Read, import, and refresh store the floor the API returns and `null` when the API returns none.
- Attribute docs point at the public Minimum Search VCUs documentation for the pricing and inactivity-metering consequences instead of restating them.

This change adds no breaking change to existing arguments. Existing configurations without `min_search_vcu` keep working; a floor configured outside Terraform on a project whose configuration has a `search_lake` block without `min_search_vcu` shows as a planned removal and is removed on the next apply, which is standard Terraform behaviour and can be opted out of with `lifecycle { ignore_changes = [search_lake.min_search_vcu] }`; a configuration with no `search_lake` block at all needs no opt-out.

## Capabilities

### New Capabilities

- `elasticsearch-project-min-search-vcu`: The Minimum Search VCU floor on `ec_elasticsearch_project` — schema, validation, create, read, update, and unset — against `search_lake.min_search_vcu` of the serverless project API.

### Modified Capabilities

<!-- none — no canonical spec covers ec_elasticsearch_project or its search_lake block -->

## Impact

- Hand-written create, patch, and read mapping for `search_lake` in `ec/ecresource/projectresource/elasticsearch.go`. Unsetting the floor sends JSON `null` on the wire, which the generated `omitempty` pointer cannot express, so the patch reuses the raw-body path introduced for audit logging (`PatchElasticsearchProjectWithBodyWithResponse`).
- Generated serverless client and Plugin Framework models under `ec/internal/gen/serverless/`, regenerated from the vendored public project API bundle once `serverless-project-api.source` points at a commit of `elastic/serverless-api-specification` whose production bundle contains `search_lake.min_search_vcu`. The field appeared in the public production bundle on 2026-11-16.
- `ec/internal/gen/serverless/modify_spec.sh` marks the attribute optional (not computed), replaces the generated `maximum` validator with a lower-bound-only validator, and extends its description with the documentation link.
- The ref bump also pulls in unrelated upstream drift since the pinned commit `07c1cd8d`: today that is `metadata.annotations` on create, read, and patch. This change does not expose annotations; it keeps them out of the generated Terraform schema so the three project resources do not grow an unmapped attribute. Exposing annotations is its own change.
- Registry docs for `ec_elasticsearch_project` (regenerated), a new example, a unit test suite in `ec/ecresource/projectresource`, one acceptance test in `ec/acc` for humans and Buildkite, and a `.changelog/{PR}.txt` entry on the implementation PR.
- Out of scope: `ec_observability_project` (no `search_lake` attribute) and `ec_security_project` (its `search_lake` has only `data_retention`), hosted `ec_deployment`, a project data source (none exists), client-side knowledge of organization-specific maximums, and any provider awareness of how the field is rolled out.
