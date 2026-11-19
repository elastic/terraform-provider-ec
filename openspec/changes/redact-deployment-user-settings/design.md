## Context

See `proposal.md` for motivation and `specs/deployment-user-settings/spec.md` for the behavior contract. The deployment service schemas define YAML/JSON settings as optional string attributes in Terraform Plugin Framework schemas; the legacy and current schema versions both need consistent metadata. Hosted deployment operations use `cloud-sdk-go`; this change does not alter API requests.

## Goals / Non-Goals

**Goals:**
- Redact user settings and ECE override values from normal Terraform CLI plan output.
- Preserve existing attribute types, values sent to Elastic Cloud, and state behavior.

**Non-Goals:**
- Mark Cloud IDs or endpoint URLs sensitive.
- Remove settings from Terraform state or saved plan files.
- Sanitize unrelated provider or API error messages.

## Decisions

- Set `Sensitive: true` on each user-settings YAML/JSON leaf attribute in the v1 and v2 deployment service schemas. Include ECE override variants because they accept the same kind of arbitrary configuration payload. Do not mark the parent service/config block sensitive, which would also hide non-sensitive sibling attributes.
- Keep the change to schema sensitivity metadata. Terraform Plugin Framework uses this metadata to obscure values in CLI output; changing the API payload or error handling is unnecessary for the plan-output leak described in the issue.
- Add no-cloud regression coverage for schema sensitivity. Cover representative user and override attributes across service schemas, including both Elasticsearch v1 settings locations; verify plan output redaction if the existing test harness supports it.

## Risks / Trade-offs

- Sensitive metadata does not remove values from Terraform state or saved plans → document that state and plan files must still be protected.
- Sensitive values propagate to expressions and outputs → expected for settings that may contain secrets; do not extend sensitivity to connection identifiers.

## Migration Plan

No state migration or API rollout is required. Rollback consists of removing the schema sensitivity metadata; existing state values remain unchanged.
