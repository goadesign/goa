# Releasing Goa

This document is intended to help Goa maintainers release new versions of Goa.

## Publishing an opt-in preview

A preview publishes Goa alone from the reviewed feature branch. It does not
move the stable `v3` branch, update the stable README badge, or change the
examples and plugins repositories. Because a stable release exists, the Go
command does not select a preview for `@latest`; testers must request the full
version.

Preview versions use `v3.MINOR.BUILD-preview.NUMBER`. For example:

```text
v3.31.0-preview.1
```

The preview version does not promise the final stable version. Decide the
stable major version after community feedback and the final compatibility
review.

### Prepare the preview

1. Confirm the complete preview version with the maintainer.
2. Make sure the current Goa feature branch contains the intended generator
   changes, public upgrade guide, and release notes.
3. Make sure the working tree is clean and the preview tag does not exist
   locally or on `origin`.
4. Run the preparation target with the confirmed version parts:

   ```bash
   make prepare-preview MINOR=31 BUILD=0 PREVIEW_NUMBER=1
   ```

   The target validates the version, runs the complete Goa release preflight,
   updates `Makefile` and `pkg/version.go`, verifies `goa version`, and creates
   the commit `Prepare v3.31.0-preview.1`. It does not create or push a tag.

5. Push the feature branch through its normal review workflow and wait for all
   required checks. Do not merge it into the stable branch merely to publish
   the preview.
6. Confirm that the public upgrade guide still matches the exact tagged code.

### Publish the preview

Publishing creates a public tag. Inventory the local and remote branch and tag
one final time, then get explicit maintainer approval before running:

```bash
make release-preview
```

The target reruns the complete preflight, checks that the current commit is
already present on the branch with the same name on `origin`, creates the
preview tag, and pushes only that tag. It does not push a branch or touch the
examples and plugins repositories.

If publication fails after the local tag is created, inspect the local and
remote tag before doing anything else. Do not rerun blindly, delete a public
tag, or force-push.

Create the GitHub release only after the tag push succeeds, and mark it as a
pre-release:

```bash
gh release create v3.31.0-preview.1 \
  --prerelease \
  --title "Goa v3.31.0-preview.1" \
  --notes-file PREVIEW_NOTES.md
```

Release notes should link to `UPGRADING.md`, state that the final stable
version is undecided, identify the users most likely to be affected, and ask
for reports with a small design and generated diff. Remove the temporary notes
file after GitHub has accepted the release.

Testers install the preview explicitly:

```bash
go get goa.design/goa/v3@v3.31.0-preview.1
go install goa.design/goa/v3/cmd/goa@v3.31.0-preview.1
```

## Publishing a stable release

## Using `make release`

1. Update `MAJOR`, `MINOR` and `BUILD` as needed in `Makefile`. Leave
   `PREVIEW_NUMBER` empty. The stable release target refuses to run while a
   preview number is set and clears the preview suffix in `pkg/version.go`.
2. Make sure the `goa.design/examples` and `goa.design/plugins` repositories exist in `$(go env GOPATH)/src` and are clean:
   - `$(go env GOPATH)/src/goa.design/examples` on branch `main`
   - `$(go env GOPATH)/src/goa.design/plugins` on branch `v3`
3. Run `make release`

`make release` runs a preflight check (`release-preflight`) after bumping the version and updating
the README badge. The preflight runs `lint`, `test-release` (no coverage artifact) and
`integration-test` before tagging and pushing.

## Manual release procedure

1. Update `MAJOR`, `MINOR` and `BUILD` as needed in `Makefile`, and leave
   `PREVIEW_NUMBER` empty.
2. Update `pkg/version.go` and `README.md` to reflect the new version.
3. Commit and push to v3.
4. Create and push release git tag.
5. Update `go.mod` in the examples repo `master` branch.
6. Run `make` in the examples repo.
7. Push the examples repo `master` branch.
8. Create and push release git tag.
9. Update `go.mod` in the plugins repo `v3` branch.
10. Run `make` in the plugins repo.
11. Create and push release git tag.
