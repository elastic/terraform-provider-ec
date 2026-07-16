---
name: release
description: Prepare and publish a new release of terraform-provider-ec. Bumps version, consolidates changelog, creates prep PR, and optionally tags.
disable-model-invocation: true
user-invocable: true
argument-hint: "[patch|minor|X.Y.Z]"
---

# Release terraform-provider-ec

Prepare a new release for the Terraform Provider for Elastic Cloud.

## Determine version

If the user provided an explicit version like `X.Y.Z`, use that. Otherwise determine the bump type:

- `$ARGUMENTS` can be `patch`, `minor`, or an explicit version like `0.13.0`
- If no argument is given, analyze the changes to decide:
  - **patch**: only bug fixes since last release
  - **minor**: any new features, enhancements, or breaking changes
  - No major releases — the project doesn't do major bumps yet
- Tell the user what version you're targeting and why before proceeding.

## Step 1: Identify changes since last release

1. Find the last release tag: `git describe --tags $(git rev-list --tags --max-count=1)`
2. Get its date: `git log -1 --format=%ai <tag>`
3. Fetch all PRs merged since that date:
   ```
   gh pr list --repo elastic/terraform-provider-ec --state merged --search "merged:>YYYY-MM-DD" --json number,title,mergedAt --limit 100
   ```
4. Filter out dependency updates (`fix(deps):`, `chore(deps):`), CI changes (`chore:` actions/checkout, actions/setup-go, etc.), and the previous release prep PR.
5. For each remaining PR, check if it has a `.changelog/` entry. Flag PRs that look user-facing but have no changelog entry.

## Step 2: Build changelog

1. Check existing `.changelog/*.txt` files in the repo.
2. **IMPORTANT**: Verify each `.changelog/*.txt` file's content matches upstream, not just the local copy. Use `git show upstream/master:.changelog/NNN.txt` or check GitHub.
3. Draft the full changelog section using the format from `scripts/changelog.tmpl`:
   - BREAKING CHANGES (release-note:breaking-change)
   - NOTES (release-note:note)
   - FEATURES (release-note:feature, release-note:new-resource, release-note:new-data-source)
   - ENHANCEMENTS (release-note:enhancement)
   - BUG FIXES (release-note:bug)
4. Include entries for user-facing PRs that are missing `.changelog/` files.
5. **Present the changelog draft to the user for review before writing any files.**

## Step 3: Apply changes

After the user approves the changelog:

1. Create a new branch: `prep-X.Y.Z`
2. Bump version in `Makefile` and `ec/version.go` to `X.Y.Z-dev` (keep the `-dev` suffix)
3. Update provider version (without `-dev`) in:
   - `README.md`
   - All files under `examples/` that declare the `ec` provider
   You can use `./scripts/update-provider-version.sh X.Y.Z` for this.
4. Write the new changelog section to `CHANGELOG.md` (prepend before the previous release)
5. Delete the `.changelog/*.txt` files that were consolidated
6. Commit with message: `Prepare X.Y.Z release`

## Step 4: Create PR

Ask the user before creating the PR. When approved:
- Push the branch
- Create PR with title `Prepare X.Y.Z release` targeting `master`

## Step 5: Tag (after merge)

Only when the user explicitly asks to tag (typically after the PR is merged):
1. Ensure you're on `master` with the merge commit
2. Run `make tag` — it creates the `vX.Y.Z` tag (the `VERSION` minus its `-dev` suffix) and pushes it to the remote pointing at `elastic/terraform-provider-ec`, which triggers the Buildkite release pipeline.
3. Equivalently, tag and push by hand: `git tag vX.Y.Z` then `git push <remote> vX.Y.Z`, where `<remote>` is the one pointing at `elastic/terraform-provider-ec` (often `upstream`).

## Important rules

- **Don't post anything to GitHub until the user explicitly asks.**
- **Always present the changelog for review before writing files.**
- **Keep `-dev` suffix in `Makefile` and `ec/version.go`.**
- **Verify `.changelog/` file contents against upstream.**
- **Check all merged PRs, not just those with `.changelog/` entries.**
