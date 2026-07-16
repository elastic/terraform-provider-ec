# Releasing a new version

This guide describes how to release a new version of `terraform-provider-ec`.

> **Most of this is automated by the `/release` skill**
> ([`.agents/skills/release/SKILL.md`](../.agents/skills/release/SKILL.md)): it works out the next
> version, consolidates the `.changelog/` fragments into `CHANGELOG.md`, bumps the version, and
> prepares the release PR. This document is the manual reference for that same process plus the
> post-merge tagging step.

- [Prerequisites](#prerequisites)
  - [Make sure `VERSION` is up to date](#make-sure-version-is-up-to-date)
  - [Ensure the `elastic/cloud-sdk-go` dependency is up to date](#ensure-the-elasticcloud-sdk-go-dependency-is-up-to-date)
  - [Generate the changelog for the new version](#generate-the-changelog-for-the-new-version)
  - [Ensure the `NOTICE` file is up to date](#ensure-the-notice-file-is-up-to-date)
- [Executing the release](#executing-the-release)
- [Post-release tasks](#post-release-tasks)

## Prerequisites

Releasing a new version implies there have been source changes worth publishing. Check the
following before releasing.

### Make sure `VERSION` is up to date

The version is defined in the [Makefile](../Makefile) as an exported variable, in
[SemVer](https://semver.org) `MAJOR.MINOR.PATCH` form with a `-dev` suffix on the working copy:

```Makefile
export VERSION := 0.13.0-dev
```

The `-dev` suffix marks an unreleased, in-development version; the release tooling strips it when
tagging and publishing (so `0.13.0-dev` is tagged and published as `v0.13.0`). **Keep the `-dev`
suffix in the committed `Makefile`.**

Bump the version with the helper targets, which use `versionbump` to edit the `Makefile` and then
run [`scripts/update-provider-version.sh`](../scripts/update-provider-version.sh) to update the
provider version referenced in `README.md` and under `examples/`:

- `make minor` — new features / bug fixes, no breaking changes (the usual case).
- `make major` — breaking changes.
- `make patch` — bug-fix-only release (see [Patch versions](#patch-versions)).

`ec/version.go` (which embeds the version into the binary) is regenerated from `VERSION` by
`make gen` / `make build`, so it needs no manual edit. The bump happens on the release prep
branch/PR — this is what the `/release` skill prepares. There is no automatic version-bump
workflow.

#### Patch versions

Patch releases are cut from the corresponding minor branch. For example, to release `v0.2.1`, check
out the `0.2` branch, make the changes there, and use `make patch` to bump the `PATCH` component.

### Ensure the `elastic/cloud-sdk-go` dependency is up to date

The provider depends heavily on
[`github.com/elastic/cloud-sdk-go`](https://github.com/elastic/cloud-sdk-go); make sure it is on
the intended version. Renovate keeps it in sync, but it is worth double-checking.

### Generate the changelog for the new version

The changelog lives at the top-level [`CHANGELOG.md`](../CHANGELOG.md). It is assembled from the
per-PR `.changelog/{PR}.txt` fragments (see
[`high-level/contributing.md`](./high-level/contributing.md) for the fragment format). Run
`make changelog` (which runs `scripts/generate-changelog.sh`) to regenerate it. The `/release`
skill does this and deletes the fragments it consolidates.

### Ensure the `NOTICE` file is up to date

Regenerate the third-party `NOTICE` with `make notice` (it runs `go-licence-detector` over the
module graph). On Renovate dependency PRs this is done automatically by the
[`renovate-notice.yml`](../.github/workflows/renovate-notice.yml) workflow, so `NOTICE` is usually
already current by release time — but verify.

## Executing the release

Once the prerequisites are met and the prep PR is merged, from the release commit run:

```sh
make tag
```

`make tag` creates the `vX.Y.Z` tag (the `VERSION` with `-dev` stripped) and pushes it to the
`elastic/terraform-provider-ec` remote (it aborts if the tag already exists or that remote is not
configured). Pushing the tag triggers the Buildkite
[release pipeline](https://buildkite.com/elastic/terraform-provider-ec-release), which runs
[`.buildkite/release.sh`](../.buildkite/release.sh) → `make release` (GoReleaser) to build, sign,
and publish the artifacts.

Confirm the new version is listed in the
[Terraform registry](https://registry.terraform.io/providers/elastic/ec/latest/docs).

## Post-release tasks

- Confirm the release published successfully and `CHANGELOG.md` shows the new version's section at
  the top.
- No immediate version bump is required: the working-copy `VERSION` keeps its `-dev` suffix until
  the next release's prep PR bumps it to the next target version (via `make minor` / `make patch`,
  or the `/release` skill).
