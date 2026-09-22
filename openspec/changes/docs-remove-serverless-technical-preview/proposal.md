## Why

Registry pages for the serverless project resources still open with a **Technical preview** heading even though those resources are no longer preview. The heading and its two-sentence disclaimer should come off the published docs.

This is mechanical documentation copy only (GitHub issue elastic/terraform-provider-ec#944 as the subject). Terraform schema, CRUD, and resource behavior do not change.

## What Changes

- Delete the `## Technical preview` heading and the two sentences that follow it from:
  - `templates/resources/elasticsearch_project.md.tmpl`
  - `templates/resources/observability_project.md.tmpl`
  - `templates/resources/security_project.md.tmpl`
- Regenerate registry markdown with `env -u TF_ACC make docs-generate` so `docs/resources/{elasticsearch,observability,security}_project.md` no longer start that block.

No Go, schema, resource CRUD, examples, or `.changelog/` entry (docs-only skip).

## Capabilities

### New Capabilities

<!-- none — skip_specs: true; mechanical docs copy, no spec-level behavior -->

### Modified Capabilities

<!-- none — no Resource implementation: specs; repo rule skips a spec for mechanical docs -->

## Impact

- Source templates under `templates/resources/` for `ec_elasticsearch_project`, `ec_observability_project`, and `ec_security_project`.
- Generated pages under `docs/resources/` for those three types (do not hand-edit `docs/`).
- Out of scope: provider schema, generated serverless client, CRUD, acceptance tests, changelog, canonical `openspec/specs/`.
