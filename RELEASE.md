# Releasing Goa

This document is intended to help Goa maintainers release new versions of Goa.

## Using `make release`

1. Update `MAJOR`, `MINOR` and `BUILD` as needed in `Makefile`.
2. Run `make release`

## Manual release procedure

1. Update `MAJOR`, `MINOR` and `BUILD` as needed in `Makefile`.
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
