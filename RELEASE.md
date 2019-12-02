# Releasing Goa

This document is intended to help Goa maintainers release new versions of Goa.

## Release Procedure

1. Update `pkg/version.go` and `README.md` to reflect the new version.
2. Create git tags in the Goa repo for both the `v2` and `v3` branches.
3. Update `go.mod` in the examples repo `master` branch.
4. Generate the Goa examples from the examples repo `master` and `v2` branches.
5. Push the examples repo `master` and `v2` branches.
6. Update the plugins repo `v3` branch `go.mod` file.
6. Generate the plugin examples from the plugins repo `v3` and `v2` branches.
5. Create git tags for plugin repo (both `v2` and `v3`).
5. Push the tags.
6. Write and publish blog to goa.design.
