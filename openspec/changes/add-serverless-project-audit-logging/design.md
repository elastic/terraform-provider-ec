## Context

See proposal.md — Why. The three serverless project resources are Terraform Plugin Framework resources. CRUD goes through the generated serverless client in `ec/internal/gen/serverless/`, not `cloud-sdk-go`. Create, patch, and read bodies are built by hand in `ec/ecresource/projectresource` (`elasticsearch.go`, `observability.go`, `security.go`) from the generated models.

The raw project API on `elastic/serverless-api-specification` `main` already defines `monitoring.logging.audit` (create, read, and patch, including JSON null to reset a category). Those schemas are marked `x-exclude-from-documentation: true`. The file this provider vendors, `specifications/generated/production/public-user-serverless-api-dereferenced.yml`, strips them. The committed spec (`serverless-project-api.source` ref `059c09ac`) has no `monitoring` object.

`client-config.yaml` uses default oapi-codegen. Nullable properties become `*T` with `json:",omitempty"`. A nil pointer is omitted, so it cannot encode `null`. `OptionalTrafficFilters` is cleared today by a non-nil empty slice (`expandTrafficFilterIdsForPatch`), which marshals as `[]`. Linked-project nulls work only because they are nil map values, which `omitempty` does not strip.

`modify_spec.sh` already rewrites generated project schemas. Optional attributes that appear on both create and read come out `computed_optional`. Attributes required on create come out `required`. Read-only attributes come out `computed`. `traffic_filters` is deleted and replaced by the string set `traffic_filter_ids`. `linked.statuses` is a computed attribute beside the practitioner-controlled `projects` map.

## Goals / Non-Goals

**Goals:**

- One optional `audit` nested attribute, mapped the same way on all three project resources.
- Generated models once the public bundle contains the fields. Hand-written code maps them. Generated files are not edited by hand.
- Unit tests assert marshalled JSON for `audit` and `ignore_filters` null clears. Acceptance stays a human/Buildkite step.

**Non-Goals:**

- Vendoring the raw user spec, or deleting `x-exclude-from-documentation` in a local copy of the public bundle.
- Turning on oapi-codegen `nullable-type` in `client-config.yaml`. That would change `OptionalLinkConfiguration`, `OptionalTrafficFilters`, `OptionalElasticsearchSearchLake`, and their callers.
- A shared Terraform resource type. The three resources stay separate.
- Modeling `logging` categories other than `audit`.

## Decisions

- **Stop until the public bundle contains the fields.** Bump `serverless-project-api.source` `ref` only to a commit whose `public-user-serverless-api-dereferenced.yml` includes `monitoring.logging.audit`, then run `scripts/update-serverless-spec.sh` and `env -u TF_ACC make gen`. Alternative: hand-write the Go types now and regenerate later — rejected because the next `make gen` would delete them.
- **Force the nested attributes the generator would mark computed_optional.** In `modify_spec.sh`, set `monitoring`, `logging`, and `audit` to optional. Leave `destination` required when `audit` is set, because the create schema requires `destination`, not because `project_id` is required inside it. Leave `enabled` computed_optional and give it `Default: true`. A default on a non-computed attribute fails provider schema validation. Leave `destination.project_type` computed_optional so read can store the API value without a perpetual diff when config omits it. Leave `destination.status` computed inside `destination`. `destination` is one nested object, so a computed child does not hide removal the way a computed field inside the linked-project map entry does. Create and patch copy `project_id` and `project_type` only, and only when `project_type` is known.
- **Replace `ignore_filters` with a set of ids.** The API field is an array of `{id}` objects with no order promise, so the generator emits a nested list. Delete that attribute and add an optional set of strings, the same kind of attribute as `traffic_filter_ids`. A list would plan a diff whenever the API returned ids in another order. Re-add `setvalidator.SizeAtMost(10)` because replacing the attribute drops the generated `maxItems` validator. Do not reject an empty set. Null and empty both mean no filters. The mapper writes each string as `{"id":"..."}`.
- **Clear with JSON null, and keep the configured absence.** The project API documents `ignore_filters: null` as the way to remove every filter. Removing a non-empty set, whether the new configuration is null or empty, sends that null. A generated nil pointer cannot encode it, because `omitempty` drops the key. The same raw patch body that sends `"audit":null` sends `"ignore_filters":null`. On read, an absent or empty API array keeps the null or empty set already in the plan or state, so `ignore_filter_ids = []` does not come back as null; a non-empty prior set becomes null, so removed filters show as drift. Create omits `ignore_filters` for both null and empty.
- **Clear the category with JSON null.** Removing `audit` must send `"audit":null` and must not send `"monitoring":null`. A nil `*OptionalLoggingCategory` would omit `audit` and leave the category in place. The clear path uses `Patch*ProjectWithBodyWithResponse` on `ClientWithResponsesInterface`. Tests compare the marshalled body, not the Go struct. Alternative: global `nullable-type` — rejected under Non-Goals.
- **Read keeps a configured empty shell.** No `monitoring.logging.audit` does not by itself null `monitoring`. If the API has no audit and the plan or state `audit` is already null, keep that `monitoring` value, including `monitoring = {}` or `monitoring = { logging = {} }`. Store `monitoring` as null when the API has no audit and the plan or state `monitoring` is null. When the API returns audit and the plan or state `audit` is null, shell or not, store the API object with `ignore_filter_ids` null if the API array is absent or empty, so import and refresh can see console configuration and a shell cannot hide it. When the plan or state has audit and the API has none, store `monitoring` as null. One `Read` serves create, update, and refresh with no call-site marker, so after an apply Terraform's inconsistent-result check reports the dropped write, the same way `linked` behaves today.

## Risks / Trade-offs

- [Public bundle still strips the schemas] → Implementation stops before `make gen`. No local spec fork. Dropping `x-exclude-from-documentation` is an upstream project-API change; this repo does not publish that bundle.
- [Hand-built patch JSON drifts from the generated request struct] → Use `Patch*ProjectWithBodyWithResponse` when the body must contain `audit` or `ignore_filters` null. Other patch fields stay on the generated struct, merged into that JSON. The unit test locks those bytes.
- [`destination.status` and an omitted `project_type` show as known after apply] → Both are computed and are not sent on write. An unrelated update may mark them `(known after apply)`. That is expected.
- [API adds another `logging` key later] → This change only reads and writes `audit`. A null `audit` leaves sibling keys alone.
- [Acceptance needs a real second project as the destination] → Unit tests cover the request body. A human runs any `TestAcc…` case; the agent does not set `TF_ACC`.

## Migration Plan

Projects created by this provider with no `monitoring` block stay unconfigured. A project whose audit logging was set outside Terraform, such as in the console, will show a plan to remove it after the first refresh. Add the block to configuration before applying, or the apply sends `audit: null`. Removing a block that this provider manages also sends `audit: null`. No state migration.

## Open Questions

None.
