## Why

Serverless projects can send audit logs to another project (`monitoring.logging.audit` on the project API), but the Terraform provider has no attributes for that configuration. Customers who manage projects with Terraform cannot enable, pause, retarget, or filter audit log delivery without the console or raw API calls (GitHub issue elastic/terraform-provider-ec#1065).

## What Changes

- Add an optional `monitoring.logging.audit` configuration to `ec_elasticsearch_project`, `ec_observability_project`, and `ec_security_project`.
- Practitioners set `enabled`, a destination project id and type, and `ignore_filter_ids` (a set of ignore-filter ids). The API stores those ids as `ignore_filters[].id`. Null and an empty set both mean no filters.
- `destination.project_type` is required and limited to `observability` or `security`, the only types the API accepts as logging destinations. The provider reads back computed `destination.status`.
- Removing the audit configuration clears that category by sending JSON `null` for `audit` only. Other monitoring fields stay as they are.

This change adds no breaking change to existing arguments. Omitting `monitoring` on a new project leaves audit logging unconfigured. A project that already has audit logging in the API, for example from the console, will plan to remove it on the next apply unless the configuration includes the block.

## Capabilities

### New Capabilities

- `project-audit-logging`: Audit log delivery on the three serverless project resources — schema, create, read, update, and clear — against `monitoring.logging.audit`.

### Modified Capabilities

<!-- none — no canonical spec covers the serverless project resources -->

## Impact

- Hand-written create, patch, and read mapping in `ec/ecresource/projectresource` for the elasticsearch, observability, and security project resources. Clearing `audit` or `ignore_filters` sends JSON `null` on the wire, which the generated `omitempty` pointers cannot express. `modify_spec.sh` replaces the generated `ignore_filters` object list with a set, `ignore_filter_ids`, the same kind of attribute as `traffic_filter_ids`.
- Generated serverless client and Plugin Framework models under `ec/internal/gen/serverless/`, after the vendored public project API bundle actually contains `monitoring.logging.audit`. That bundle currently strips these schemas (`x-exclude-from-documentation: true` on the raw project API). Implementation cannot regenerate the client until the fields are present in the file this provider vendors.
- Registry docs for the three project resources, plus a `.changelog/{PR}.txt` entry on the implementation PR. The ref bump regenerates the whole client and the three resource schemas, so unrelated upstream drift since the pinned ref comes with it; the implementation PR's changelog covers that drift, not this proposal.
- Out of scope: hosted `ec_deployment`, a VectorDB project resource, reading the ignore-filter catalog (the monitoring API's list/get endpoints), and log categories other than `audit`.
