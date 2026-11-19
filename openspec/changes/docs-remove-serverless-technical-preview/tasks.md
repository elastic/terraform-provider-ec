## 1. Remove Technical preview copy from serverless project templates

- [x] 1.1 Delete the `## Technical preview` heading and the two sentences after it from `templates/resources/elasticsearch_project.md.tmpl`, `templates/resources/observability_project.md.tmpl`, and `templates/resources/security_project.md.tmpl`, and verify those three files go from H1 to `{{ .Description }}` with no preview block.

## 2. Regenerate registry docs

- [x] 2.1 Run `env -u TF_ACC make docs-generate` and verify `docs/resources/elasticsearch_project.md`, `docs/resources/observability_project.md`, and `docs/resources/security_project.md` no longer contain the Technical preview heading or the two disclaimer sentences.
