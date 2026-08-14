# Documentation

The cloud provider's registry documentation (`docs/`) is **generated** with
[`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs). Do not hand-edit `docs/` —
edit the sources and regenerate.

## Where docs come from

`tfplugindocs` builds each page from three inputs:

- **Templates** in `templates/` — the page layout and prose.
  - `index.md.tmpl` — the provider landing page.
  - `resources/*.md.tmpl` and `data-sources/*.md.tmpl` — one template per resource/data source.
  - `guides/` — long-form guides (e.g. `configuring-sso-ec-deployment.md`); these are plain
    `.md` files that pass through unchanged.
- **Schema descriptions** compiled into the provider binary. Templates render them with
  `{{ .SchemaMarkdown | trimspace }}`, and pull metadata such as `{{ .Name }}`, `{{ .Type }}`,
  and `{{ .Description }}`. So attribute docs come from the `Description` fields on the schema in
  Go code, not from Markdown.
- **Example `.tf`/shell files** in `examples/`, injected by template functions:
  - `{{ tffile "examples/resources/ec_deployment_extension/with-file/resource.tf" }}` embeds a
    Terraform example as a fenced code block.
  - `{{ codefile "shell" .ImportFile }}` embeds the resource's `import.sh`.
  - Examples follow a fixed layout: `examples/resources/<name>/`, `examples/data-sources/<name>/`,
    plus standalone example directories at the top of `examples/` (see `examples/README.md`).

The rendered output lands in `docs/` (`docs/index.md`, `docs/resources/`, `docs/data-sources/`,
`docs/guides/`), mirroring the `templates/` tree.

> `docs/` is generated output. Never edit it by hand — your change will be overwritten on the next
> run and CI will fail (see below). Change the template, the schema `Description`, or the example
> file instead, then regenerate.

## Generate

```sh
make docs-generate
```

This runs `go tool tfplugindocs` (defined in `build/Makefile.build`) with no arguments, so it uses
the default layout: read `templates/` + `examples/` + the provider schema, write `docs/`.

## Validate

```sh
make tfproviderdocs
```

This runs `go tool tfproviderdocs check -provider-name terraform-provider-ec .`
(`build/Makefile.lint`), which checks the generated docs for structural problems (missing pages,
mislabeled files, etc.). It is also part of `make lint`. Note that `make lint` validates the docs
but does **not** regenerate them — run `make docs-generate` yourself after editing sources.

## CI keeps docs in sync

CI regenerates the docs and fails if the working tree changes:

```sh
make docs-generate
git diff --exit-code docs/   # errors if docs are stale
```

(see `.github/workflows/go.yml`). So whenever you change a resource/data-source **schema**, a
**template**, or an **example**, run `make docs-generate` and commit the resulting `docs/` updates
in the same PR.

## Related

- This page covers **documentation** generation. For **code** generation (the serverless OpenAPI
  clients), see [`generated-clients.md`](./generated-clients.md).
- For where `make docs-generate` and `make tfproviderdocs` sit in the day-to-day loop, see
  [`development-workflow.md`](./development-workflow.md).
