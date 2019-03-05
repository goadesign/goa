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
	github.com/golang/protobuf/protoc-gen-go \
	github.com/golang/protobuf/proto

all: lint gen test

travis: depend all build-examples clean

# Install protoc
GOOS=$(shell go env GOOS)
PROTOC_VERSION="3.6.1"
ifeq ($(GOOS),linux)
PROTOC="protoc-$(PROTOC_VERSION)-linux-x86_64"
PROTOC_EXEC="$(PROTOC)/bin/protoc"
GOBIN="$(GOPATH)/bin"
else
	ifeq ($(GOOS),darwin)
PROTOC="protoc-$(PROTOC_VERSION)-osx-x86_64"
PROTOC_EXEC="$(PROTOC)/bin/protoc"
GOBIN="$(GOPATH)/bin"
	else
		ifeq ($(GOOS),windows)
PROTOC="protoc-$(PROTOC_VERSION)-win32"
PROTOC_EXEC="$(PROTOC)\bin\protoc.exe"
GOBIN="$(GOPATH)\bin"
		endif
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
	rm -rf $(GOPATH)/src/goa.design/goa/examples/basic/cmd              && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/cellar/cmd/cellar-cli  && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/error/cmd              && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/multipart/cmd          && \
	rm -rf $(GOPATH)/src/goa.design/goa/examples/security/cmd           && \
	goa gen     goa.design/goa/examples/basic/design      -o $(GOPATH)/src/goa.design/goa/examples/basic     && \
	goa example goa.design/goa/examples/basic/design      -o $(GOPATH)/src/goa.design/goa/examples/basic     && \
	goa gen     goa.design/goa/examples/cellar/design    -o $(GOPATH)/src/goa.design/goa/examples/cellar   && \
	goa example goa.design/goa/examples/cellar/design    -o $(GOPATH)/src/goa.design/goa/examples/cellar   && \
	goa gen     goa.design/goa/examples/error/design     -o $(GOPATH)/src/goa.design/goa/examples/error    && \
	goa example goa.design/goa/examples/error/design     -o $(GOPATH)/src/goa.design/goa/examples/error    && \
	goa gen     goa.design/goa/examples/multipart/design -o $(GOPATH)/src/goa.design/goa/examples/multipart && \
	goa example goa.design/goa/examples/multipart/design -o $(GOPATH)/src/goa.design/goa/examples/multipart && \
	goa gen     goa.design/goa/examples/security/design  -o $(GOPATH)/src/goa.design/goa/examples/security && \
	goa example goa.design/goa/examples/security/design  -o $(GOPATH)/src/goa.design/goa/examples/security && \
	goa gen     goa.design/goa/examples/streaming/design -o $(GOPATH)/src/goa.design/goa/examples/streaming  && \
	goa example goa.design/goa/examples/streaming/design -o $(GOPATH)/src/goa.design/goa/examples/streaming

build-examples:
	@cd $(GOPATH)/src/goa.design/goa/examples/basic && \
		go build ./cmd/calc && go build ./cmd/calc-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/cellar && \
		go build ./cmd/cellar && go build ./cmd/cellar-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/error && \
		go build ./cmd/divider && go build ./cmd/divider-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/multipart && \
		go build ./cmd/resume && go build ./cmd/resume-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/security && \
		go build ./cmd/multi_auth && go build ./cmd/multi_auth-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/streaming && \
		go build ./cmd/chatter && go build ./cmd/chatter-cli

clean:
	@cd $(GOPATH)/src/goa.design/goa/examples/basic && \
		rm -f calc calc-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/cellar && \
		rm -f cellar cellar-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/error && \
		rm -f divider divider-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/multipart && \
		rm -f resume resume-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/security && \
		rm -f multi_auth multi_auth-cli
	@cd $(GOPATH)/src/goa.design/goa/examples/streaming && \
		rm -f chatter chatter-cli

test:
	go test ./...

ifeq ($(GOOS),windows)
PLUGINS_BRANCH="$(GOPATH)\src\goa.design\plugins"
else
PLUGINS_BRANCH="$(GOPATH)/src/goa.design/plugins"
endif
test-plugins:
	@if [ -z $(GOA_BRANCH) ]; then\
		GOA_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	fi
	@if [ ! -d "$(GOPATH)/src/goa.design/plugins" ]; then\
		git clone https://github.com/goadesign/plugins.git $(PLUGINS_BRANCH); \
	fi
	@cd $(PLUGINS_BRANCH) && git checkout $(GOA_BRANCH) || echo "Using master branch in plugins repo" && \
	make -k test-plugins || (echo "Tests in plugin repo (https://github.com/goadesign/plugins) failed" \
                  "due to changes in goa repo (branch: $(GOA_BRANCH))!" \
                  "Create a branch with name '$(GOA_BRANCH)' in the plugin repo and fix these errors." && exit 1)
