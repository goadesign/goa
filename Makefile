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
DIRS=$(shell go list -f {{.Dir}} goa.design/goa/design/...)
ALIASER_DESTS=\
	http

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	github.com/sergi/go-diff/diffmatchpatch \
	github.com/golang/lint/golint \
	golang.org/x/tools/cmd/goimports

all: lint aliases gen test

travis: depend all

depend:
	@mkdir -p $(GOPATH)/src/golang.org/x
	@git clone https://github.com/golang/lint.git $(GOPATH)/src/golang.org/x/lint
	@go get -v $(DEPEND)
	@go get -t -v ./...

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
	for d in $(ALIASER_DESTS) ; do \
		./aliaser -src goa.design/goa/dsl -dest goa.design/goa/$$d/dsl > /dev/null; \
	done

gen:
	@cd cmd/goa && \
	go install && \
	goa gen goa.design/goa/examples/cellar/design -o $(GOPATH)/src/goa.design/goa/examples/cellar && \
	goa gen goa.design/goa/examples/calc/design -o $(GOPATH)/src/goa.design/goa/examples/calc && \
	goa gen goa.design/goa/examples/error/design -o $(GOPATH)/src/goa.design/goa/examples/error && \
	goa gen goa.design/goa/examples/security/design -o $(GOPATH)/src/goa.design/goa/examples/security && \
	goa gen goa.design/goa/examples/streaming/design -o $(GOPATH)/src/goa.design/goa/examples/streaming

test:
	go test ./...

test-aliaser: aliases
	@for d in $(ALIASER_DESTS) ; do \
		if [ "`git diff $$d/*/aliases.go | tee /dev/stderr`" ]; then \
			echo "^ - Aliaser tool output not identical!" && echo && exit 1; \
		else \
			echo "Aliaser tool output identical"; \
		fi \
	done

test-plugins:
	@if [ -z $(GOA_BRANCH) ]; then\
		GOA_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	fi
	@go get -d -v goa.design/plugins/... && \
	cd $(GOPATH)/src/goa.design/plugins && \
	git checkout $(GOA_BRANCH) || echo "Using master branch" && \
	make -k || (echo "Tests in plugin repo (https://github.com/goadesign/plugins) failed" \
                  "due to changes in goa repo (branch: $(GOA_BRANCH))!" \
                  "Create a branch with name '$(GOA_BRANCH)' in the plugin repo and fix these errors." && exit 1)
