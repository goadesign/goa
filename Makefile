#! /usr/bin/make
#
# Makefile for goa v2
#
# Targets:
# - "depend" retrieves the Go packages needed to run the linter and tests
# - "lint" runs the linter and checks the code format using goimports
# - "aliases" builds the DSL aliases files
# - "test" runs the tests
#
# Meta targets:
# - "all" is the default target, it runs all the targets in the order above.
#
DIRS=$(shell go list -f {{.Dir}} goa.design/goa.v2/design/...)

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	github.com/golang/lint/golint \
	golang.org/x/tools/cmd/goimports

all: depend lint aliases test

depend:
	@go get -t -v ./...
	@go get -v $(DEPEND)

lint:
	@for d in $(DIRS) ; do \
		if [ "`goimports -l $$d/*.go | tee /dev/stderr`" ]; then \
			echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
		fi \
	done
	@if [ "`golint ./... | grep -vf .golint_exclude | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi

aliases:
	@cd cmd/aliaser && \
	go build && \
	./aliaser > /dev/null

test:
	go test ./...
