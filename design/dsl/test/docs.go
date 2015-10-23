// Package test contains a self-contained DSL test.
// This test must be in its own package to emulate the proper order of global
// variables and package initialization.
// This file is needed for `go get ./...` and thus the build to succeed.
package test
