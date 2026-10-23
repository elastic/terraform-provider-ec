# Agentic workflow supply-chain & egress policy (`.github/aw/`)

This directory holds the **trust + egress boundary** for the repository's
[GitHub Agentic Workflows](https://github.com/githubnext/gh-aw) (gh-aw). It is **infrastructure,
not a workflow**: it is committed ahead of the Phase 3/4 agentic factories/scanners (see epic
[`elastic/cp-hosted-team#4362`](https://github.com/elastic/cp-hosted-team/issues/4362)) so that every
agentic workflow, once added under `.github/workflows/*.md`, runs inside a pinned, least-egress
envelope.

Agentic workflows pull third-party Actions and container images and make outbound network calls. If
any of those float on a mutable tag or reach an un-vetted host, a supply-chain compromise or
data-exfiltration path opens across the fleet. This policy pins every dependency to an immutable
digest and constrains each workflow to a minimal allowlist.

> **gh-aw version is pinned to `v0.92.0`** in
> [`.github/workflows/workflows.yml`](../workflows/workflows.yml) (`gh extension install
> githubnext/gh-aw --pin v0.92.0`). Every pin below corresponds to that compiler version. On a gh-aw
> bump, re-generate (see *Regenerating* / *Updating on a gh-aw bump*) and review the diff.

## `actions-lock.json`

A `package-lock.json` for the agentic stack. Its two sections behave differently:

- `entries` (Action SHAs) are **managed by gh-aw** — `gh aw compile` resolves each Action referenced by
  the workflow `.md` sources to a commit SHA, records it, and prunes the rest (see *Do not
  hand-maintain the entry set*).
- `containers` (image digests) are a **committed, authoritative pin** that compile **consumes** but
  never auto-generates or auto-corrects (see *Container image digests*).

Regenerate with:

```sh
make workflow-generate          # → gh aw compile --purge   (build/Makefile.test)
```

### Do not hand-maintain the entry set

`make workflow-generate` runs `gh aw compile **--purge**` (not `--approve`). `--purge` is
authoritative: it rewrites `entries` to contain **exactly** the Actions referenced by the compiled
output and **deletes any entry nothing references**. Consequences:

- Most committed `entries` are gh-aw's **standard scaffolding** Actions, which appear in every
  compiled workflow and therefore survive `--purge` (`actions/checkout`, `actions/download-artifact`,
  `actions/github-script`, `actions/setup-go`, `actions/setup-node`, `actions/upload-artifact`,
  `github/gh-aw-actions/setup`). The authoritative set is whatever `make workflow-generate` writes —
  read the diff, do not maintain the list by hand.
- **Do not seed an Action no workflow uses.** `--purge` strips any entry nothing references, and the
  CI drift check would then fail on an unrelated PR. `hashicorp/setup-terraform@v4.0.1` **is** now in
  `entries`: the `code-factory-issue` workflow imports `shared/setup-dev.md`, which uses the Terraform
  CLI for `make validate-examples`, so gh-aw records it. It stays only while a workflow references it.
- gh-aw's scaffolding Actions are emitted into **every** compiled workflow regardless of frontmatter —
  e.g. `gh aw compile` still injects an `actions/checkout` step even when a workflow sets
  `checkout: false` — so those pins are retained unconditionally. `--purge` prunes only a
  **non-scaffolding** entry (like `hashicorp/setup-terraform`) once nothing references it. A clean
  drift check therefore confirms the committed set matches the compiled output, but does not by itself
  prove every retained entry is individually needed.

### Container image digests

`containers` records the agentic container images pinned to a digest — the firewall sandbox
(`gh-aw-firewall/{squid,agent,api-proxy,cli-proxy}`), the MCP gateway (`gh-aw-mcpg`), the Node base
(`gh-aw-node`), and `github-mcp-server` — all at the versions shipped by gh-aw **v0.92.0**.

These committed digests are **authoritative**: `gh aw compile` reads them and pins them **verbatim**
into every generated `*.lock.yml` — the exact image the firewall / node / MCP containers pull at
runtime, fleet-wide. Change a digest here and the generated `*.lock.yml` pulls that image — this is how
the images are pinned.

For an image **not yet listed here**, compile pins a digest **only if gh-aw's embedded pin catalog
knows it** (it does for the firewall / `gh-aw-mcpg` / `github-mcp-server` / `gh-aw-node` images, so a
clean compile still pins those, per-architecture, into the `*.lock.yml` without writing them back). An
image absent from **both** this file **and** the embedded catalog is emitted with its **original,
possibly mutable tag** — compile does **not** fetch a digest for it. So any new image must be added
here with an explicitly reviewed digest; never rely on compile to pin an image gh-aw does not already
know.

Compile **preserves** this block across `--purge` and never auto-rewrites or auto-corrects it. Two
consequences follow, both security-relevant:

- **This is a manual-review surface, not a CI-enforced one.** The `Workflows CI` drift check only
  re-runs compile, which keeps whatever is committed here — so it will **not** catch a stale or
  malicious digest in `containers`. Any change to a digest is security-sensitive and must be reviewed
  by a human, especially on a gh-aw version bump (refresh these together with the regenerated
  `*.lock.yml`).
- **Commit `amd64` / CI-resolved digests** (CI runs on `ubuntu-latest`). Because the committed values
  are used verbatim on every platform, pinning the `amd64` digest makes `arm64` and `amd64` compiles
  identical — which is why this lock reproduces the same everywhere. The per-architecture caveat
  applies only when **adding** an image not yet in this block (e.g. a new firewall train on a gh-aw
  bump): resolve its digest from an `amd64` / CI context, since a local `arm64` resolve would record
  the wrong value. Digests already locked here are used as-is and are never re-resolved, so a local
  regenerate will not corrupt them.

## Network allowlist policy

Each workflow declares its egress in frontmatter as `network.allowed`. The firewall sandbox
(squid/api-proxy/cli-proxy) enforces it; named sets (`defaults`, `node`, `go`, …) expand to concrete
hosts. Inspect a workflow's effective egress by reading the compiled lock's `GH_AW_ALLOWED_DOMAINS`
(`grep -o 'GH_AW_ALLOWED_DOMAINS: "[^"]*"' <lock> | head -1 | tr ',' '\n'`). Note `gh aw domains
<workflow> --json` reports an **empty** list at the pinned v0.92.0 and must not be relied on.

The policy menu for this repo is:

```yaml
network:
  allowed: [defaults, node, go, registry.terraform.io, releases.hashicorp.com, www.elastic.co, openrouter.ai]
```

This is a **menu, not a block to paste verbatim** — each workflow includes only the entries it needs
(a triage/classifier agent that calls no docs MCP omits `www.elastic.co`, a workflow that never runs
Terraform omits `registry.terraform.io`, etc.). Per the extract-on-second-use `imports:` convention
(see [`dev-docs/high-level/coding-standards.md`](../../dev-docs/high-level/coding-standards.md)), once
a second workflow needs the same block, factor it into a shared fragment.

| Entry | On | Why |
| --- | --- | --- |
| `defaults` | ✅ | gh-aw baseline: GitHub API, TLS CA/OCSP/CRL endpoints, Ubuntu apt mirrors. Required by any runner. This set is itself broad (e.g. `storage.googleapis.com`, package registries), so "least-egress" below is **relative to this baseline**. |
| `node` | ✅ | npm registry + Node toolchain — the `.github/scripts/workflows` JS helpers and their tests. |
| `go` | ✅ | Go module proxy / checksum DB — building and testing the provider and the `scripts/changelogcheck` helper. |
| `registry.terraform.io` | ✅ | Terraform registry metadata for provider/schema lookups during `terraform init`. Pinned as a **single host** rather than gh-aw's `terraform` domain set for least-egress. gh-aw emits a non-fatal advisory suggesting the `terraform` set; we keep explicit hosts deliberately. At v0.92.0 the `terraform` set expands to `apt.releases.hashicorp.com, registry.terraform.io, releases.hashicorp.com, yum.releases.hashicorp.com` — i.e. it additionally opens the OS **apt/yum package mirrors** (not checkpoint endpoints), which this repo's workflows never use. |
| `releases.hashicorp.com` | ✅ (conditional) | Terraform **provider binaries** for `terraform init`. Required by a workflow that runs `make validate-examples`: `examples/deployment_ec2_instance` needs `hashicorp/aws`, which `scripts/validate_examples.sh` installs via `direct` (only `elastic/ec` is mirrored locally). Omit from any workflow that does not validate examples. |
| `www.elastic.co` | ✅ (conditional) | Egress host for the **Elastic docs MCP** (`https://www.elastic.co/docs/_mcp/`). Required **only** by a workflow that declares the `elastic-docs` MCP server (the research/change/code/reproducer factories consult Elastic Cloud / provider docs while drafting). Omit it from any workflow that does not wire up that MCP. |
| `openrouter.ai` | ✅ | The inference endpoint. Agentic workflows call **OpenRouter** (`ANTHROPIC_BASE_URL=https://openrouter.ai/api`, model `anthropic/claude-sonnet-5`); the only inference secret is the repository secret `OPENROUTER_API_KEY`. |

> **`make validate-examples` needs both Terraform hosts.** `registry.terraform.io` alone does **not**
> let a sandboxed `make validate-examples` finish. `scripts/validate_examples.sh` configures a
> Terraform `filesystem_mirror` for `elastic/ec` only and installs every other provider via `direct`;
> `examples/deployment_ec2_instance` requires `hashicorp/aws`, whose provider binary is fetched from
> `releases.hashicorp.com`. A workflow that validates examples therefore lists **both**
> `registry.terraform.io` (metadata) and `releases.hashicorp.com` (binaries) in its `network.allowed`
> — as `code-factory-issue` does. The local `elastic/ec` mirror does not remove released-provider
> downloads for other providers.

### Excluded on purpose (least-egress, cloud-provider-specific)

| Not allowed | Why it is excluded |
| --- | --- |
| `api.elastic.cloud` (and the live Elastic Cloud API generally) | Agents **never** call the live, paid Cloud API. Acceptance tests create/destroy real deployments against `EC_API_KEY` and run **out-of-band on Buildkite** (`buildkite/terraform-provider-ec-acceptance`), never inside an agentic workflow — no live-cloud credentials are exposed to agents. Allowing this host would open a spend / data-exfiltration path. (Verified: it is absent from the compiled firewall egress for the menu above.) |
| Local-stack ports (Elasticsearch `:9200`, Kibana `:5601`, Docker localstack) | **N/A** — unlike the Elastic Stack provider, the cloud provider has **no** local Docker stack; there is nothing to reach. Recorded here so these are not copied in out of habit when porting a stack-provider workflow. |

## Verifying / regenerating

- **Reproducibility (AC):** `make workflow-generate` followed by
  `git diff --exit-code -- .github/workflows .github/aw` must be clean. CI enforces this in the
  `Workflows CI` *test* job.
- **Updating on a gh-aw bump:** change the pin in `workflows.yml`, run `make workflow-generate`,
  refresh the `containers` digests here, and review the diff (new/changed Action SHAs, image digests,
  redirects) as a security review. Renovate must not edit `*.lock.yml` (compiled output) —
  `renovate.json` lists those under `ignorePaths`. Bump `github/gh-aw-actions` together with the
  compiler pin via `make workflow-generate`, not by patching lock files.
- **Inspect egress:** read `GH_AW_ALLOWED_DOMAINS` from the compiled `*.lock.yml` (see the Network
  allowlist policy section). `gh aw domains` returns an empty list at the pinned v0.92.0.
