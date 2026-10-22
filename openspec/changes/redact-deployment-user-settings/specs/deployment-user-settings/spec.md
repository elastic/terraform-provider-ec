## Purpose

Defines how the Elastic Cloud deployment resource identifies user-supplied service settings that may contain sensitive values, so Terraform can redact them from normal CLI output.

## ADDED Requirements

### Requirement: Deployment user settings are marked sensitive

The deployment resource SHALL mark user settings supplied as YAML or JSON, including ECE override variants, as sensitive.

#### Scenario: Redacting configured service settings

- **GIVEN** a deployment configuration contains YAML or JSON user settings for a supported service
- **WHEN** Terraform displays a plan containing those settings
- **THEN** Terraform SHALL redact the settings values from normal CLI output

#### Scenario: Redacting ECE override settings

- **GIVEN** a deployment configuration contains YAML or JSON ECE override settings
- **WHEN** Terraform displays a plan containing those settings
- **THEN** Terraform SHALL redact the override values from normal CLI output
