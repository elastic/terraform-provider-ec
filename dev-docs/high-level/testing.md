# Testing

The cloud provider has two test tiers with very different cost and safety profiles:

| Tier | Command | Credentials | Cost / duration | Safe to run locally? |
|------|---------|-------------|-----------------|----------------------|
| Unit | `make unit` (alias `make tests`) | none | seconds, free | Yes, always |
| Acceptance | `make testacc` (`TF_ACC=1`) | `EC_API_KEY` | up to ~2h, **real money** | **No — see warning below** |

Unlike the stack provider (`terraform-provider-elasticstack`), the cloud provider has
**no local Docker stack**. Acceptance tests run against the **live Elastic Cloud API**, provisioning
and destroying real deployments and serverless projects. All test wiring lives in
[`build/Makefile.test`](../../build/Makefile.test).

## Unit tests

- Run: `make unit` (or its alias `make tests`).
- No credentials, no network to Elastic Cloud, always safe and fast.
- Recipe: `go test $(TEST) $(TESTARGS) $(TESTUNITARGS)`, where `TEST ?= ./...` and
  `TESTUNITARGS ?= -timeout 10m -race -cover -coverprofile=reports/c.out`.
- Tests are co-located with the code as `*_test.go` files throughout `ec/…`.

## Acceptance tests

> ⚠️ **COST WARNING.** `make testacc` (and anything gated by `TF_ACC=1`) creates and destroys
> **real Elastic Cloud deployments** against the live API. Runs can last up to ~2 hours
> (`-timeout 120m`) and **cost real money**.

> 🚫 **Agents must never run acceptance tests** — no `TF_ACC` in agentic workflows, and no
> live-cloud credentials are exposed to agents; agents rely on the CI run instead of provisioning
> paid infrastructure. The tests *are* run automatically for every PR, but by the dedicated
> **Buildkite acceptance pipeline** (the GitHub Actions `go.yml` CI runs unit/lint/docs only), and a
> human reviews the result. A maintainer may also run them locally when a change needs it — mind the
> cost and `make sweep` any leftovers.

Gating and recipe (from `build/Makefile.test`):

```make
testacc:
	TF_ACC=1 go test $(TEST_ACC) -v -count $(TEST_COUNT) -parallel $(TEST_ACC_PARALLEL) \
	  $(TESTARGS) -timeout 120m -run $(TEST_NAME)
```

Defaults: `TEST_ACC ?= github.com/elastic/terraform-provider-ec/ec/acc`, `TEST_NAME ?= TestAcc`,
`TEST_COUNT ?= 1`, `TEST_ACC_PARALLEL = 6`. Tests skip themselves unless `TF_ACC=1` is set
(`requiresAPIConn` in [`ec/acc/acc_prereq.go`](../../ec/acc/acc_prereq.go)).

### Required environment

- `EC_API_KEY` — Elastic Cloud API key (the standard credential; see the "Generating an API Key"
  section of the top-level [`README.md`](../../README.md)). API-key vs. username/password is
  validated by `testAccPreCheck`.
- `EC_HOST` (optional) — override the API endpoint to target a non-prod / QA region. When unset it
  defaults to the production Elastic Cloud endpoint; setting a custom host also skips TLS
  verification. (`EC_ENDPOINT` is accepted as an alias.)

Acceptance fixtures are HCL `*.tf` files under
[`ec/acc/testdata/`](../../ec/acc/testdata/); the Go test files that drive them live in
[`ec/acc/`](../../ec/acc/). Created resources are named with the `terraform_acc_` prefix so the
sweepers can find them.

## Targeting a single test

Use `TEST_NAME` (matched by `go test -run`) to narrow a run to one `TestAcc…`, and `TESTARGS` for
any extra `go test` flags:

```sh
make testacc TEST_NAME='TestAccDeployment_basic'
```

## Buildkite (per-PR acceptance)

Acceptance runs are wired through Buildkite, not run inline by contributors:

- [`.buildkite/pull-requests.json`](../../.buildkite/pull-requests.json) gates the
  `terraform-provider-ec-acceptance` pipeline: org users with `admin`/`write` (plus `renovate[bot]`)
  trigger it on commit or on a `build this` / `test this` comment; changes touching only `^docs/`
  or `^dev-docs/` are skipped.
- [`.buildkite/acceptance_pipeline.yml`](../../.buildkite/acceptance_pipeline.yml) defines the
  single "Acceptance tests" step running [`.buildkite/acceptance.sh`](../../.buildkite/acceptance.sh)
  on a `golang` image.
- `acceptance.sh` runs `make vendor` then `EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc`.
- The [`pre-command`](../../.buildkite/hooks/pre-command) hook loads the API key from Vault and
  exports `BUILD_ID` (which makes `make sweep` skip its interactive confirmation).
- The [`pre-exit`](../../.buildkite/hooks/pre-exit) hook always sweeps afterward (see below).

A human reviews the Buildkite result as part of PR review.

## Sweepers

A passing acceptance test destroys the resources it created (Terraform's `CheckDestroy`). The
**sweepers** are the second line of defence: they reap `terraform_acc_`-prefixed resources that
**leaked** from a failed or interrupted run. `make sweep`:

```make
sweep:
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m
```

- Defaults: `SWEEP ?= us-east-1`, `SWEEP_DIR ?= $(TEST_ACC)`.
- Outside CI (`BUILD_ID` unset), the target prints a warning and prompts for confirmation before
  destroying anything (`ifndef BUILD_ID`); run it only against a development account. Buildkite's
  `pre-command` hook sets `BUILD_ID`, so CI skips the prompt.
- Registered sweepers (in `ec/acc/*_sweep_test.go`): `ec_deployments`, `ec_serverless_projects`,
  and `ec_deployment_traffic_filter`. The **deployment** and **serverless-project** sweepers only
  delete resources **older than 3h** (`deploymentStaleAfter` / `serverlessProjectStaleAfter`), so
  they never remove a still-running build's resources; the traffic-filter sweeper deletes all
  matching filters.
- **CI cleans up automatically:** Buildkite's `pre-exit` hook (on the `acceptance-tests` step) runs
  `SWEEPARGS='-sweep-run=ec_deployments,ec_serverless_projects' make sweep` after every acceptance
  build — so stale deployments **and** projects are reaped on exit regardless of pass/fail. (The
  `acceptance.sh` step itself only runs `make vendor` + `make testacc`.)
- **When to run manually:** after a *local* acceptance failure that may have left dangling
  infrastructure, or to reclaim serverless quota (see below).

## Known flake — serverless project quota (HTTP 403)

Serverless-project acceptance tests can intermittently fail with an HTTP `403` response reporting
`project limit [100]` reached. This is an **environmental serverless-quota** condition (too many
leftover projects in the shared org), **not a defect in the change under test**. Do not "fix" the
code to work around it — instead **run the sweepers** (`make sweep` filtered to
`ec_serverless_projects`) to reclaim quota and **retry** the acceptance run.

## See also

- Contributor setup and PR flow: [`contributing.md`](./contributing.md)
- Test-coverage expectations: [`coding-standards.md`](./coding-standards.md)
- Where code lives: [`repo-structure.md`](./repo-structure.md)
- Common workflows: [`development-workflow.md`](./development-workflow.md)
