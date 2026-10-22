## 1. Schema

- [x] 1.1 Mark YAML/JSON user settings and ECE override attributes sensitive across the deployment service schemas and v1/v2 versions, including both Elasticsearch v1 locations; verify all affected schema attributes declare sensitivity.

## 2. Regression Tests

- [x] 2.1 Add no-cloud schema tests covering user settings and override sensitivity across service schemas; verify with `env -u TF_ACC go test ./ec/ecresource/deploymentresource/...`.

## 3. Validation

- [x] 3.1 Validate the change artifacts with `env -u TF_ACC make check-openspec`.
- [x] 3.2 Regenerate deployment documentation after the schema changes and verify it with `env -u TF_ACC make tfproviderdocs`.
