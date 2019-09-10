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
GOOS=$(shell go env GOOS)
GO_FILES=$(shell find . -type f -name '*.go')

ifeq ($(GOOS),windows)
EXAMPLES_DIR="$(GOPATH)\src\goa.design\examples"
PLUGINS_DIR="$(GOPATH)\src\goa.design\plugins"
GOBIN="$(GOPATH)\bin"
else
EXAMPLES_DIR=$(GOPATH)/src/goa.design/examples
PLUGINS_DIR=$(GOPATH)/src/goa.design/plugins
GOBIN=$(GOPATH)/bin
endif

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	golang.org/x/lint/golint \
	golang.org/x/tools/cmd/goimports \
	github.com/golang/protobuf/protoc-gen-go \
	github.com/golang/protobuf/proto \
	honnef.co/go/tools/cmd/staticcheck

all: lint test

travis: depend all test-examples test-plugins

# Install protoc
PROTOC_VERSION=3.6.1
ifeq ($(GOOS),linux)
PROTOC=protoc-$(PROTOC_VERSION)-linux-x86_64
PROTOC_EXEC=$(PROTOC)/bin/protoc
else
	ifeq ($(GOOS),darwin)
PROTOC=protoc-$(PROTOC_VERSION)-osx-x86_64
PROTOC_EXEC=$(PROTOC)/bin/protoc
	else
		ifeq ($(GOOS),windows)
PROTOC=protoc-$(PROTOC_VERSION)-win32
PROTOC_EXEC="$(PROTOC)\bin\protoc.exe"
		endif
	endif
endif
depend:
	@go get -v $(DEPEND)
	@env GO111MODULE=off go get github.com/hashicorp/go-getter/cmd/go-getter && \
		go-getter https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC).zip $(PROTOC) && \
		cp $(PROTOC_EXEC) $(GOBIN) && \
		rm -r $(PROTOC) && \
		echo "`protoc --version`"
	@go get -t -v ./...

lint:
	@if [ "`goimports -l $(GO_FILES) | tee /dev/stderr`" ]; then \
		echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
	fi
	@if [ "`golint ./... | grep -vf .golint_exclude | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi
	@if [ "`staticcheck -checks all ./... | grep -v ".pb.go" | tee /dev/stderr`" ]; then \
		echo "^ - staticcheck errors!" && echo && exit 1; \
	fi

test:
	env GO111MODULE=on go test ./...

test-examples:
	@if [ -z $(GOA_BRANCH) ]; then\
		GOA_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	fi
	@if [ ! -d $(EXAMPLES_DIR) ]; then\
		git clone https://github.com/goadesign/examples.git $(EXAMPLES_DIR); \
	fi
	@cd $(EXAMPLES_DIR) && git checkout $(GOA_BRANCH) || echo "Using master branch in examples repo" && \
	make -k travis || (echo "Tests in examples repo (https://github.com/goadesign/examples) failed" \
                  "due to changes in Goa repo (branch: $(GOA_BRANCH))!" \
                  "Create a branch with name '$(GOA_BRANCH)' in the examples repo and fix these errors." && exit 1)

test-plugins:
	@if [ -z $(GOA_BRANCH) ]; then\
		GOA_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	fi
	@if [ ! -d $(PLUGINS_DIR) ]; then\
		git clone https://github.com/goadesign/plugins.git $(PLUGINS_DIR); \
	fi
	@cd $(PLUGINS_DIR) && git checkout $(GOA_BRANCH) || echo "Using master branch in plugins repo" && \
	make -k test-plugins || (echo "Tests in plugin repo (https://github.com/goadesign/plugins) failed" \
                  "due to changes in goa repo (branch: $(GOA_BRANCH))!" \
                  "Create a branch with name '$(GOA_BRANCH)' in the plugin repo and fix these errors." && exit 1)
