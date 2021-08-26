# Contributing to terraform-provider-ec

Contributions are very welcome, these can include documentation, bug reports, issues, feature requests, feature implementations or tutorials.

- [Contributing to terraform-provider-ec](#contributing-to-terraform-provider-ec)
  - [Reporting Issues](#reporting-issues)
  - [Code Contribution Guidelines](#code-contribution-guidelines)
    - [Workflow](#workflow)
    - [Commit Messages](#commit-messages)
  - [Setting up a dev environment](#setting-up-a-dev-environment)
    - [Environment prerequisites](#environment-prerequisites)
  - [Development](#development)
    - [Running tests](#running-tests)
      - [Unit](#unit)
      - [Acceptance](#acceptance)
        - [Sweepers](#sweepers)
    - [Build terraform-provider-ec locally with your changes](#build-terraform-provider-ec-locally-with-your-changes)

## Reporting Issues

If you have found an issue or defect in the `terraform-provider-ec` or the latest documentation, use the GitHub [issue tracker](https://github.com/elastic/terraform-provider-ec/issues) to report the problem. Make sure to follow the template provided for you to provide all the useful details possible.

## Code Contribution Guidelines

For the benefit of all and to maintain consistency, we have come up with some simple guidelines. These will help us collaborate in a more efficient manner.

- Unless the PR is very small (e.g. fixing a typo), please make sure there is an issue for your PR. If not, make an issue with the provided template first, and explain the context for the change you are proposing.

  Your PR will go smoother if the solution is agreed upon before you've spent a lot of time implementing it. We encourage PRs to allow for review and discussion of code changes.

- We encourage PRs to be kept to a single change. Please don't work on several tasks in a single PR if possible.

- When you're ready to create a pull request, please remember to:
  - Make sure you've signed the Elastic [contributor agreement](https://www.elastic.co/contributor-agreement).

  - Have test cases for the new code. If you have questions about how to do this, please ask in your pull request.
  
  - Run `make format`, `make lint` and `make fmt`.
  
  - Ensure that [unit](#unit) and [acceptance](#acceptance) tests succeed with `make unit testacc`.

  - After you've opened your Pull request and have a PR number, make sure to generate a changelog entry, [See example entries](https://github.com/elastic/terraform-provider-ec/tree/95e7f5c7fe6795163aff1118a7f7add44e23de50/.changelog).
  
  - Use the provided PR template, and assign any labels which may fit your PR.
  
  - There is no need to add reviewers, the code owners will be automatically added to your PR.

### Workflow

The codebase is maintained using the "contributor workflow" where everyone without exception contributes patch proposals using "pull requests". This facilitates social contribution, easy testing and peer review.

To contribute a patch, make sure you follow this workflow:

1. Fork repository.
2. Enable GitHub actions to run in your fork.
3. Create topic branch.
4. Commit patches.

### Commit Messages

In general commits should be atomic and diffs should be easy to read.

Commit messages should be verbose by default consisting of a short subject line. A blank line and detailed explanatory text as separate paragraph(s), unless the title alone is self-explanatory ("trivial: Fix comment typo in main.go"). Commit messages should be helpful to people reading your code in the future, so explain the reasoning for your decisions.

If a particular commit references another issue, please add the reference (e.g. refs #123 or closes #124).

Example:

```console
doc: Complete ec_deployment attributes and help

Completes the `ec_deployment` markdown document with all of the fields.

Closes #1234
```

## Setting up a dev environment

### Environment prerequisites

- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- [Go](https://golang.org/doc/install) 1.13

This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH). Running `make vendor` will download all the required dependencies.

## Development

### Running tests

#### Unit

There are three variables that can be passed to the `unit` make target:

- `TEST` controls which directories or files to test. Defaults to `./...` which means all packages.
- `TESTUNITARGS` controls the flags that are sent to `go test`. Defaults to `-timeout 10s -p 4 -race -cover`.
- `TESTARGS` controls any additional flags you may want to pass to `go test`.

#### Acceptance

Before running the acceptance tests make sure you have exported your API key to the `EC_API_KEY` environment variable. There are three variables that can be passed to the `testacc` make target:

- `TEST_NAME` controls which test names to test. Defaults to `./...` which means all.
- `TESTARGS` controls any additional flags you may want to pass to `go test`.
- `TEST_COUNT` controls how many times each test is run. Defaults to 1.

_Note: Acceptance tests may incur in charges for the deployments that are created. If you do not wish to run acceptance tests locally, you can rely on the acceptance tests which are run automatically on every pull request._

##### Sweepers

Additionally, there is a `make sweep` target which destroys any dangling infrastructure created by the acceptance tests. For more information on acceptance testing, check out the official Terraform [documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html).

Running the sweepers will remove all the `terraform_acc_` prefixed resources for the registered sweepers, each resource should add its own sweepers so that in cases dangling resources are left, they can be cleaned up with `make sweep`. There are three variables that can be passed to `sweep`:

- `SWEEP` controls the region (if any) to pass to the resources to be swept. Defaults to `us-east-1`.
- `SWEEP_DIR` controls the directory where the sweepers are found. Defaults to `github.com/elastic/terraform-provider-ec/ec/acc`.
- `SWEEPARGS` controls the additional command flags to pass to the `go test` command:
  - `-sweep-run` can be set to filter which sweepers are run (matching is done with `string.Contains`).
  - `-sweep-allow-failures` can be set to allow all sweepers to be run when one or more have failed.

### Build terraform-provider-ec locally with your changes

To build a temporary binary inside your project's root run:

```console
$ cd terraform-provider-ec
$ make build
```

You can also use the `make install` target if you wish. This target will install the binary and move it to your Terraform plugin location.
