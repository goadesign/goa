---
name: goa-release
description: Release Goa v3, including preflight tests, dependency updates, semver version bumps, release preparation commits, examples/plugins repo checks, and babysitting make release. Use when the user asks to release Goa, prepare a Goa release, bump the Goa version, or run make release.
---

# Goa Release

Use this skill only from the Goa repository. `make release` pushes branches and tags for Goa,
examples, and plugins, so do not run it unless the user explicitly wants to perform the release.

## Release Contract

- Goa major version is always `3`. Never increment `MAJOR`; if it is not `3`, stop and ask.
- Resolve the target version before making dependency-update commits in any repository. The agent
  should pick the version from the changes, explain the reasoning, and get confirmation before
  editing files or running dependency updates.
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
3. Inspect the commits and merged PRs since the previous release tag.
4. Pick the semver bump:
   - Minor when the release includes a user-visible feature, new DSL/runtime capability, or
     backward-compatible public API addition.
   - Patch when the release only includes fixes, dependency updates, docs, tests, or internal
     maintenance.
5. Tell the user the selected target version and the specific changes that justify it, then ask for
   confirmation. If the user supplied an explicit version, validate it against the same reasoning and
   call out any mismatch before continuing.
6. Use the exact confirmed target version for every preparation commit, tag check, and release
   command.

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

## GitHub Release Notes

After `make release` succeeds, create GitHub release notes for the Goa tag.

1. Gather the final commit/PR list from the previous Goa tag to the new tag. Include merged PR
   numbers, titles, authors, and any notable dependency or downstream examples/plugins release work.
2. Identify every human contributor in that range from commit authors and PR authors. Do not miss
   contributors whose commits were merged by someone else. Exclude bots from the thank-you list
   unless their work is directly meaningful to users.
3. Write notes that are clear and useful to someone who does not know Goa internals:
   - Start with a short plain-language summary of why the release matters.
   - Use sections only when the amount of change justifies them, such as `Highlights`, `Fixes`,
     `Dependencies`, `Upgrade Notes`, and `Contributors`.
   - Explain user impact before implementation detail. Avoid raw commit-log dumps.
   - Include upgrade notes only for actions users may need to take.
   - Thank every contributor by name or handle.
4. Create or update the GitHub release for `v3.x.y` with the final notes. Do not mention agent or AI
   tooling in public release notes.

## Announcements

After GitHub release notes are ready, draft announcements for Slack, Bluesky, and Substack. Keep the
tone excited, concrete, and not cheesy.

- Slack: 2-4 short paragraphs or bullets. Lead with the most useful user-facing change, link to the
  GitHub release, and thank contributors.
- Bluesky: Check the current Bluesky character limit before drafting. If it cannot be checked,
  assume 300 characters and stay under the limit. Include the version, one concrete highlight, and a
  release link.
- Substack: Write a short note for readers who may not know Goa deeply. Use a friendly title, a
  brief explanation of Goa, the release highlights, upgrade guidance if any, and contributor thanks.
  Keep it proportional to the release size.

