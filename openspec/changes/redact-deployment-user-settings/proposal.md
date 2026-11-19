## Why

Deployment user settings can contain credentials or other private configuration, but Terraform currently displays their contents in plan output. Marking these attributes sensitive will redact them from normal CLI output.

## What Changes

- Mark user settings YAML and JSON attributes sensitive across deployment services and schema versions, including ECE override variants.
- Keep Cloud IDs and endpoint URLs unchanged; they are connection identifiers, not authentication secrets.
- Sensitive marking redacts CLI output but does not remove the values from Terraform state or saved plans.

## Capabilities

### New Capabilities
- `deployment-user-settings`: Define the handling of sensitive deployment user settings.

### Modified Capabilities

## Impact

- Affects the `ec_deployment` resource schemas for Elasticsearch, Kibana, APM, Enterprise Search, and Integrations Server.
- Adds regression coverage for sensitive schema metadata and Terraform plan redaction.
- No API payload, attribute type, or state-storage behavior changes.
