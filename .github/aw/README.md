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

> **gh-aw version is pinned to `v0.89.21`** in
> [`.github/workflows/workflows.yml`](../workflows/workflows.yml) (`gh extension install
> githubnext/gh-aw --pin v0.89.21`). Every pin below corresponds to that compiler version. On a gh-aw
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

- The committed `entries` are gh-aw's **standard scaffolding** Actions, which appear in every
  compiled workflow and therefore survive `--purge`: `actions/checkout`, `actions/download-artifact`,
  `actions/github-script`, `actions/setup-go`, `actions/setup-node`, `actions/upload-artifact`, and
  `github/gh-aw-actions/setup`.
- **Do not seed an Action no current workflow uses.** It would be stripped by the next
  `make workflow-generate`, and the CI drift check (below) would then fail on an unrelated PR. For
  example, `hashicorp/setup-terraform` is intentionally **absent**: gh-aw will add it — in that PR —
  when a workflow that runs `make validate-examples` (or otherwise uses the Terraform CLI) lands.
- The set is **self-healing**: if a workflow stops referencing a scaffolding Action (e.g. sets
  `checkout: false`), the next regenerate drops it and the drift check surfaces it in that workflow's
  PR, not silently.

### Container image digests

`containers` records the agentic container images pinned to a digest — the firewall sandbox
(`gh-aw-firewall/{squid,agent,api-proxy,cli-proxy}`), the MCP gateway (`gh-aw-mcpg`), the Node base
(`gh-aw-node`), and `github-mcp-server` — all at the versions shipped by gh-aw **v0.89.21**.

These committed digests are **authoritative**: `gh aw compile` reads them and pins them **verbatim**
into every generated `*.lock.yml` — the exact image the firewall / node / MCP containers pull at
runtime, fleet-wide. (Change a digest here and the generated `*.lock.yml` pulls that image — this is
how the images are pinned. Only when an image is *not yet* listed here does compile resolve its digest
freshly, per-architecture, into the `*.lock.yml`, without writing it back.)

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
hosts. Inspect a workflow's effective egress with `gh aw domains <workflow> --json`.

The policy menu for this repo is:

```yaml
network:
  allowed: [defaults, node, go, registry.terraform.io, www.elastic.co, openrouter.ai]
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
| `registry.terraform.io` | ✅ | Terraform provider/schema metadata for `make validate-examples`. Pinned as a **single host** rather than gh-aw's `terraform` domain set (which also opens `releases.hashicorp.com` / checkpoint), because the examples use the locally-built provider via dev overrides and do not download a released provider. gh-aw emits a non-fatal advisory suggesting the `terraform` set; we keep the single host deliberately, for least-egress. |
| `www.elastic.co` | ✅ (conditional) | Egress host for the **Elastic docs MCP** (`https://www.elastic.co/docs/_mcp/`). Required **only** by a workflow that declares the `elastic-docs` MCP server (the research/change/code/reproducer factories consult Elastic Cloud / provider docs while drafting). Omit it from any workflow that does not wire up that MCP. |
| `openrouter.ai` | ✅ | The inference endpoint. Agentic workflows call **OpenRouter** (`ANTHROPIC_BASE_URL=https://openrouter.ai/api`, model `anthropic/claude-sonnet-5`); the only inference secret is the repository secret `OPENROUTER_API_KEY`. |

### Excluded on purpose (least-egress, cloud-provider-specific)

| Not allowed | Why it is excluded |
| --- | --- |
| `api.elastic.cloud` (and the live Elastic Cloud API generally) | Agents **never** call the live, paid Cloud API. Acceptance tests create/destroy real deployments against `EC_API_KEY` and run **out-of-band on Buildkite** (`buildkite/terraform-provider-ec-acceptance`), never inside an agentic workflow — no live-cloud credentials are exposed to agents. Allowing this host would open a spend / data-exfiltration path. (Verified: it is absent from the compiled firewall egress for the menu above.) |
| Local-stack ports (Elasticsearch `:9200`, Kibana `:5601`, Docker localstack) | **N/A** — unlike the Elastic Stack provider, the cloud provider has **no** local Docker stack; there is nothing to reach. Recorded here so these are not copied in out of habit when porting a stack-provider workflow. |

## Verifying / regenerating

- **Reproducibility (AC):** with at least one workflow `.md` present, `make workflow-generate`
  followed by `git diff --exit-code -- .github/workflows .github/aw` must be clean. With **no** `.md`
  sources (the current state), `gh aw compile` has nothing to compile, so `make workflow-generate`
  skips and the committed lock is dormant. CI enforces this in the `Workflows CI` *test* job.
- **Updating on a gh-aw bump:** change the pin in `workflows.yml`, run `make workflow-generate`,
  refresh the `containers` digests here, and review the diff (new/changed Action SHAs, image digests,
  redirects) as a security review.
- **Inspect egress:** `gh aw domains <workflow> --json`.
