#! /usr/bin/make
#
# Makefile for goa v2
#
# Targets:
# - "depend" retrieves the Go packages needed to run the linter and tests
# - "lint" runs the linter and checks the code format using goimports
# - "test" runs the tests
#
# Meta targets:
# - "all" is the default target, it runs all the targets in the order above.
#
DIRS=$(shell go list -f {{.Dir}} goa.design/goa/expr/...)

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	github.com/sergi/go-diff/diffmatchpatch \
	golang.org/x/lint/golint \
	golang.org/x/tools/cmd/goimports \
	github.com/hashicorp/go-getter \
	github.com/cheggaaa/pb \
	github.com/golang/protobuf/protoc-gen-go

all: lint gen test

travis: depend all

# Install protoc
GOOS=$(shell go env GOOS)
PROTOC_VERSION="3.6.1"
ifeq ($(GOOS),linux)
PROTOC="protoc-$(PROTOC_VERSION)-linux-x86_64"
PROTOC_EXEC="$(PROTOC)/bin/protoc"
GOBIN="$(GOPATH)/bin"
else
	ifeq ($(GOOS),windows)
PROTOC="protoc-$(PROTOC_VERSION)-win32"
PROTOC_EXEC="$(PROTOC)\bin\protoc.exe"
GOBIN="$(GOPATH)\bin"
	endif
endif
depend:
	@go get -v $(DEPEND)
	@go install github.com/hashicorp/go-getter/cmd/go-getter && \
		go-getter https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC).zip $(PROTOC) && \
		cp $(PROTOC_EXEC) $(GOBIN) && \
		rm -r $(PROTOC)
	@go install github.com/golang/protobuf/protoc-gen-go
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

gen:
	@cd cmd/goa && \
	go install && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/calc/cmd              && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/cellar/cmd/cellar-cli && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/chatter/cmd/chatter   && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/error/cmd             && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/security/cmd          && \
	goa gen     goa.design/goa/examples/calc/design      -o $(GOPATH)/src/goa.design/goa/examples/calc     && \
	goa example goa.design/goa/examples/calc/design      -o $(GOPATH)/src/goa.design/goa/examples/calc     && \
	goa gen     goa.design/goa/examples/cellar/design    -o $(GOPATH)/src/goa.design/goa/examples/cellar   && \
	goa example goa.design/goa/examples/cellar/design    -o $(GOPATH)/src/goa.design/goa/examples/cellar   && \
	goa gen     goa.design/goa/examples/chatter/design   -o $(GOPATH)/src/goa.design/goa/examples/chatter  && \
	goa example goa.design/goa/examples/chatter/design   -o $(GOPATH)/src/goa.design/goa/examples/chatter  && \
	goa gen     goa.design/goa/examples/error/design     -o $(GOPATH)/src/goa.design/goa/examples/error    && \
	goa example goa.design/goa/examples/error/design     -o $(GOPATH)/src/goa.design/goa/examples/error    && \
	goa gen     goa.design/goa/examples/security/design  -o $(GOPATH)/src/goa.design/goa/examples/security && \
	goa example goa.design/goa/examples/security/design  -o $(GOPATH)/src/goa.design/goa/examples/security

test:
	go test ./...

test-plugins:
	@if [ -z $(GOA_BRANCH) ]; then\
		GOA_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	fi
	@if [ ! -d "$(GOPATH)/src/goa.design/plugins" ]; then\
		git clone https://github.com/goadesign/plugins.git $(GOPATH)/src/goa.design/plugins; \
	fi
	@cd $(GOPATH)/src/goa.design/plugins && git checkout $(GOA_BRANCH) || echo "Using master branch in plugins repo" && \
	make -k || (echo "Tests in plugin repo (https://github.com/goadesign/plugins) failed" \
                  "due to changes in goa repo (branch: $(GOA_BRANCH))!" \
                  "Create a branch with name '$(GOA_BRANCH)' in the plugin repo and fix these errors." && exit 1)
