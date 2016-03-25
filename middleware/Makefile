#! /usr/bin/make
#
# Makefile for middleware
#
# Targets:
# - "lint" runs the linter and checks the code format using goimports
# - "test" runs the tests
#
# Meta targets:
# - "all" is the default target, it runs all the targets in the order above.
#
DIRS=$(shell go list -f {{.Dir}} ./...)
DEPEND=\
 	github.com/dgrijalva/jwt-go \
	github.com/goadesign/goa.design/tools/mdc \
	github.com/goadesign/godoc2md \
	github.com/golang/lint/golint \
	github.com/onsi/ginkgo \
	github.com/onsi/ginkgo/ginkgo \
	github.com/onsi/gomega \
	github.com/spf13/hugo \
 	golang.org/x/tools/cmd/goimports

all: depend lint test

docs:
	@git clone https://github.com/goadesign/goa.design
	@rm -rf goa.design/content/reference goa.design/public
	@mdc github.com/goadesign/middleware goa.design/content/reference --exclude goa.design
	@cd goa.design && hugo
	@rm -rf public
	@mv goa.design/public public
	@rm -rf goa.design

depend:
	@go get $(DEPEND)

lint:
	@for d in $(DIRS) ; do \
		if [ "`goimports -l $$d/*.go | tee /dev/stderr`" ]; then \
			echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
		fi \
	done
	@if [ "`golint ./... | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi

test:
	@ginkgo -r --randomizeAllSpecs --failOnPending --randomizeSuites --race -skipPackage vendor

