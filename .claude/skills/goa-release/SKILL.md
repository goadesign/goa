---
name: goa-release
description: Release Goa v3, including preflight tests, dependency updates, semver version bumps, release preparation commits, examples/plugins repo checks, and babysitting make release. Use when the user asks to release Goa, prepare a Goa release, bump the Goa version, or run make release.
---

# Goa Release

Use this skill only from the Goa repository. `make release` pushes branches and tags for Goa,
examples, and plugins, so do not run it unless the user explicitly wants to perform the release.

## Release Contract

- Goa major version is always `3`. Never increment `MAJOR`; if it is not `3`, stop and ask.
- Resolve the target version before making dependency-update commits in any repository. If the user
  did not give an explicit `v3.x.y` or say whether this is a patch or minor release, ask.
- Apply semver within v3:
  - Patch release: increment `BUILD`.
  - Minor release: increment `MINOR` and reset `BUILD=0`.
- The preparation commit message must be exactly `Prepare v3.x.y`, where `x` is `MINOR` and `y`
  is `BUILD`.
- Keep the working trees clean between phases. Never discard local changes unless the user
  explicitly approves it.
- Do not push preparation commits manually. `make release-goa` runs `git pull origin <branch>` in
  examples and plugins; local-ahead preparation commits are fine as long as `origin` is not ahead.
  Later, `make release-examples` and `make release-plugins` push those preparation commits together
  with the release commits and tags.

## Version Selection

Determine the target version before changing any repository:

1. Read `MAJOR`, `MINOR`, and `BUILD` from the Goa `Makefile`.
2. Confirm `MAJOR=3`.
3. Compute the target version from the requested bump. If the user did not give an explicit
   `v3.x.y` or say whether this is a patch or minor release, ask.
4. Use the exact target version for every preparation commit, tag check, and release command.

## Required Repositories

Before release, verify these repositories exist, are on the expected branches, are clean, and are
in sync with their upstreams:

- Goa: current repository, expected branch `v3`.
- Examples: `$(go env GOPATH)/src/goa.design/examples`, expected branch `main`.
- Plugins: `$(go env GOPATH)/src/goa.design/plugins`, expected branch `v3`.

For each repository:

1. Run `git status --porcelain` and stop if there are uncommitted changes.
2. Verify the current branch is the expected branch. If not, ask before switching branches.
3. Run `git fetch origin`.
4. Verify `git rev-list --left-right --count @{u}...HEAD` reports `0 0`. If behind, run
   `git pull --ff-only`. If ahead or diverged, stop and ask.
5. Verify the target tag does not already exist locally or on `origin`.

## Preparation Workflow

1. In the Goa repository, run `make`. Fix failures before continuing.
2. In the Goa repository, run:

   ```bash
   go get -u -v ./...
   go mod tidy
   make
   ```

   `go get` should only update `go.mod` and `go.sum`. If other files changed in any repository,
   review why before committing.

3. In the examples repository, run:

   ```bash
   go get -u -v ./...
   go mod tidy
   make
   ```

   If files changed, commit them with `Prepare v3.x.y`. Do not push; `make release-examples`
   pushes the branch.

4. In the plugins repository, run:

   ```bash
   go get -u -v ./...
   go mod tidy
   make
   ```

   If files changed, commit them with `Prepare v3.x.y`. Do not push; `make release-plugins`
   pushes the branch.

5. In the Goa repository, edit only `MINOR` and `BUILD` in `Makefile` for the target version.
   Leave `MAJOR=3`.
6. If Goa files changed from dependency updates or the version bump, commit them with
   `Prepare v3.x.y`.
7. Re-check all three repositories are clean. They may be ahead of upstream by their preparation
   commits; that is expected because `make release` pushes them.

## Final Pre-Release Check

`make release` runs its own clean checks and release preflight, but those checks happen inside a
command that can create commits, tags, and pushes. Before running it:

1. In all three repositories, run `git fetch origin`.
2. Verify `git status --porcelain` is empty.
3. Verify upstream is not ahead: the first number from
   `git rev-list --left-right --count @{u}...HEAD` must be `0`. Local-ahead preparation commits are
   expected.
4. Do not push the local-ahead preparation commits. `make release` will push them together with the
   release commits.
5. Reconfirm the target version and that the target tag is absent locally and on `origin` in all
   three repositories.

## Run And Babysit Release

1. From the Goa repository, run `make release`.
2. Watch the command until it exits. Track which phase is running: `release-goa`,
   `release-examples`, or `release-plugins`.
3. If it fails, stop and inspect the failing repository before rerunning anything. Do not blindly
   rerun the whole release after a tag or push may have succeeded.
4. Before any rerun, inventory local and remote branches and tags for `v3.x.y` in all three
   repositories. Never recreate an existing tag, delete a public tag, or force-push unless the user
   explicitly approves the recovery plan.
5. Fix the root cause, verify the affected repository is in the expected state, then rerun the
   narrowest safe target or command.
6. When release completes, report the released version and the branch/tag pushes that succeeded.

