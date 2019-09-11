# Releasing Goa

This document is intended to help Goa maintainers release new versions of Goa.

## Release Procedure

1. Update `pkg/version.go` and `README.md` to reflect the new version.
2. Create git tags in the Goa repo for both the `v2` and `v3` branches.
3. Generate and push the Goa examples from the `goadesign/examples` repo `master` and `v2` branches.
4. Generate and push the plugin examples from the `goadesign/plugins` repo `v3` and `v2` branches.
5. Create git tags for plugin repo (both `v2` and `v3`).
5. Push the tags.
6. Write and publish blog to goa.design.
