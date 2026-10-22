---
steps:
  - name: Setup Go
    uses: actions/setup-go@v7.0.0
    with:
      go-version-file: go.mod
      cache: true
  - name: Setup Terraform CLI
    uses: hashicorp/setup-terraform@v4.0.1
    with:
      terraform_wrapper: false
  - name: Export Go and Terraform paths for the AWF chroot
    run: |
      # The firewalled agent container sees $GITHUB_WORKSPACE but not the Go
      # install or the RUNNER_TEMP path hashicorp/setup-terraform uses, so export
      # the Go env roots and stage the terraform binary into the workspace.
      # GOCACHE is exported too: it is $HOME-derived, and make install /
      # validate-examples may override $HOME to reach a writable
      # ~/.terraform.d, which would otherwise discard the warm build cache.
      echo "GOROOT=$(go env GOROOT)" >> "$GITHUB_ENV"
      echo "GOPATH=$(go env GOPATH)" >> "$GITHUB_ENV"
      echo "GOMODCACHE=$(go env GOMODCACHE)" >> "$GITHUB_ENV"
      echo "GOCACHE=$(go env GOCACHE)" >> "$GITHUB_ENV"
      TERRAFORM_BIN=$(command -v terraform)
      mkdir -p "$GITHUB_WORKSPACE/bin"
      cp "$TERRAFORM_BIN" "$GITHUB_WORKSPACE/bin/terraform"
      echo "TERRAFORM_BIN=$GITHUB_WORKSPACE/bin/terraform" >> "$GITHUB_ENV"
      echo "PATH=$GITHUB_WORKSPACE/bin:$PATH" >> "$GITHUB_ENV"
  - name: Setup Node.js
    uses: actions/setup-node@v7.1.0
    with:
      node-version-file: package.json
  - name: Prime provider toolchain (host, pre-agent)
    run: |
      # make deps curls golangci-lint from raw.githubusercontent.com, which is NOT
      # in the firewalled agent's egress allowlist — so run it here, on the runner
      # host, before the sandbox starts. make vendor also tidies modules and runs
      # go-licenser + make notice.
      make vendor
      make deps
  - name: Assert a clean tree after setup
    run: |
      # make vendor can write go.mod/go.sum/source headers/NOTICE. Those must be
      # committed or reverted before the agent runs: go.mod and go.sum are in
      # gh-aw's protected-file set, so pre-agent dirt would flag every PR.
      if [ -n "$(git status --porcelain)" ]; then
        echo "::error::working tree is dirty after make vendor/deps; commit or revert before the agent runs"
        git status --porcelain
        exit 1
      fi
---

# Dev environment setup (shared)

Shared `imports:` fragment for cloud-provider factory workflows that need a build environment.
Installs Go (module + build cache on), the Terraform CLI (staged into `$GITHUB_WORKSPACE/bin` so the
AWF chroot can find it), and Node.js; exports the Go env roots including `GOCACHE`; runs
`make vendor` + `make deps` on the runner host (golangci-lint is fetched from a host outside the
agent firewall); and asserts a clean tree before the agent starts.

Declares **no** `network:` block: gh-aw merges an imported fragment's `network.allowed` into the
importer's, so the egress menu stays owned by each workflow's own frontmatter (see
[`.github/aw/README.md`](../../aw/README.md)).
