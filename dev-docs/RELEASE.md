# Releasing a new version

This guide aims to provide guidance on how to release new versions of `terraform-provider-ec`.

- [Releasing a new version](#releasing-a-new-version)
  - [Prerequisites](#prerequisites)
    - [Make sure `VERSION` is up to date](#make-sure-version-is-up-to-date)
    - [Ensure the `elastic/cloud-sdk-go` dependency is up to date](#ensure-the-elasticcloud-sdk-go-dependency-is-up-to-date)
    - [Generating a changelog for the new version](#generating-a-changelog-for-the-new-version)
      - [Patch version changelog](#patch-version-changelog)
  - [Executing the release](#executing-the-release)
  - [Post release tasks](#post-release-tasks)

## Prerequisites

Releasing a new version implies that there have been changes in the source code which are meant to be released for wider consumption. Before releasing a new version there's some prerequisites that have to be checked.

### Make sure `VERSION` is up to date

Since the source has changed, we need to update the current committed version to a higher version so that the release is published.

The version is currently defined in the [Makefile](./Makefile) as an exported environment variable called `VERSION` in the [SEMVER](https://semver.org) format: `MAJOR.MINOR.PATCH`

```Makefile
SHELL := /bin/bash
export VERSION ?= v1.0.0
```

Say we want to perform a minor version release (i.e. no breaking changes and only new features and bug fixes are being included); in which case we'll update the _MINOR_ part of the version, this can be done with the `make minor` target, but it should have been updated automatically via GitHub actions.

```Makefile
SHELL := /bin/bash
export VERSION ?= v1.1.0
```

If a patch version needs to be released, the release will be done from the minor branch. For example, if we want to release `v0.2.1`, we will check out the `0.2` branch and perform any changes in that branch. The VERSION variable in the Makefile should already be up to date, but in case it's not, it can be bumped with the `make patch` target.

### Ensure the `elastic/cloud-sdk-go` dependency is up to date

Since we heavily depend on `github.com/elastic/cloud-sdk-go`, make sure that dependency is updated to the latest version. The Renovate bot does a pretty good job a keeping that in sync, but it's worth double checking.

### Generating a changelog for the new version

The changelog can be found at the top level under `CHANGELOG.md`. It is generated from a set of `<pr>.txt` files that
are saved as a changelog.

To update the changelog, run `make changelog`.

#### Patch version changelog

When releasing patch versions, the changelog will be branched out in the minor branch.

### Ensure the NOTICE file is up-to-date

Run `make vendor` to update the `NOTICE`.

## Executing the release

After all the prerequisites have been ticked off, the only thing remaining is to run `make tag`. The Buildkite Release job will attempt to release a new version. Make sure the published version is listed in the [Terraform registry](https://registry.terraform.io/providers/elastic/ec/latest/docs), you can follow the progress on the [Buildkite job](https://buildkite.com/elastic/terraform-provider-ec-release).

## Post release tasks

- After the release has been completed, the next version header should be added in the changelog as `# <VERSION> (Unreleased)` at the top of the `CHANGELOG.md` file.
- The VERSION in the [Makefile](../Makefile) should be updated
- The version in all examples should be updated to the next version to be released. 
