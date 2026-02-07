#! /usr/bin/make
#
# Makefile for Goa v3
#
# Targets:
# - "depend" retrieves the Go packages needed to run the linter and tests
# - "lint" runs the linter
# - "test" runs the tests
# - "release" creates a new release commit, tags the commit and pushes the tag to GitHub.
#   "release" also updates the examples and plugins repo and pushes the updates to GitHub.
#
# Meta targets:
# - "all" is the default target, it runs "lint" and "test"
#
MAJOR=3
MINOR=24
BUILD=3

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GO_FILES=$(shell find . -type f -name '*.go')
GOPATH=$(shell go env GOPATH)

.PHONY: all all-tests ci depend lint test test-release integration-test build-goa release release-preflight release-goa release-examples release-plugins
.NOTPARALLEL: release release-goa release-examples release-plugins

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	google.golang.org/protobuf/cmd/protoc-gen-go@latest \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest 

all: lint test integration-test

all-tests: lint test integration-test

ci: depend all

# Install protoc
PROTOC_VERSION=25.0
UNZIP=unzip
ifeq ($(GOOS),linux)
	PROTOC=protoc-$(PROTOC_VERSION)-linux-x86_64
	PROTOC_EXEC=$(PROTOC)/bin/protoc
endif
ifeq ($(GOOS),darwin)
	ifeq ($(GOARCH),arm64)
		PROTOC=protoc-$(PROTOC_VERSION)-osx-aarch_64
		PROTOC_EXEC=$(PROTOC)/bin/protoc
	else
		PROTOC=protoc-$(PROTOC_VERSION)-osx-x86_64
		PROTOC_EXEC=$(PROTOC)/bin/protoc
	endif
endif
ifeq ($(GOOS),windows)
	PROTOC=protoc-$(PROTOC_VERSION)-win32
	PROTOC_EXEC="$(PROTOC)\bin\protoc.exe"
	GOPATH:=$(subst \,/,$(GOPATH))
endif

depend:
	@echo INSTALLING DEPENDENCIES...
	@go mod download
	@for package in $(DEPEND); do go install $$package; done
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin latest 
	@go mod tidy -compat=1.17
	@echo INSTALLING PROTOC...
	@mkdir $(PROTOC)
	@cd $(PROTOC); \
	curl -O -L https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC).zip; \
	$(UNZIP) $(PROTOC).zip
	@cp $(PROTOC_EXEC) $(GOPATH)/bin && \
		rm -rf $(PROTOC) && \
		echo "`protoc --version`"

lint:
ifneq ($(GOOS),windows)
	@golangci-lint run ./... || (echo "^ - lint errors!" && echo && exit 1)
endif

test:
	go test ./... --coverprofile=cover.out

test-release:
	go test -count=1 ./...

integration-test: build-goa
ifneq ($(GOOS),windows)
	cd jsonrpc/integration_tests && go test -count=1 -timeout 10m ./...
endif

# Needed for CI to run integration tests
build-goa:
	cd cmd/goa && go install .

release-preflight: lint test-release integration-test

release: release-goa release-examples release-plugins
	@echo "Release v$(MAJOR).$(MINOR).$(BUILD) complete"

release-goa:
	# First make sure all is clean
	@status="$$(git status --porcelain)"; \
	if [ -n "$$status" ]; then \
		echo "error: goa repo has uncommitted changes:"; \
		echo "$$status"; \
		exit 1; \
	fi
	cd $(GOPATH)/src/goa.design/examples && \
		git checkout main && \
		git pull origin main && \
		status="$$(git status --porcelain)" && \
		if [ -n "$$status" ]; then \
			echo "error: examples repo has uncommitted changes:"; \
			echo "$$status"; \
			exit 1; \
		fi
	cd $(GOPATH)/src/goa.design/plugins && \
		git checkout v$(MAJOR) && \
		git pull origin v$(MAJOR) && \
		status="$$(git status --porcelain)" && \
		if [ -n "$$status" ]; then \
			echo "error: plugins repo has uncommitted changes:"; \
			echo "$$status"; \
			exit 1; \
		fi
	go mod tidy 
	# Bump version number, commit and push
	sed 's/Major = .*/Major = $(MAJOR)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/Minor = .*/Minor = $(MINOR)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/Build = .*/Build = $(BUILD)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/goa\/v3@v.*tab=doc/goa\/v3@v$(MAJOR).$(MINOR).$(BUILD)\/dsl?tab=doc/' README.md > _tmp && mv _tmp README.md
	$(MAKE) release-preflight
	git add .
	git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)"
	git tag v$(MAJOR).$(MINOR).$(BUILD)
	cd cmd/goa && go install .
	git push origin v$(MAJOR)
	git push origin v$(MAJOR).$(MINOR).$(BUILD)
	# Wait for Go proxy to update
	sleep 10

release-examples:
	cd $(GOPATH)/src/goa.design/examples && \
		make GOA_VERSION=v$(MAJOR).$(MINOR).$(BUILD)&& \
		git add . && \
		git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)" && \
		git tag v$(MAJOR).$(MINOR).$(BUILD) && \
		git push origin main && \
		git push origin v$(MAJOR).$(MINOR).$(BUILD)

release-plugins:
	cd $(GOPATH)/src/goa.design/plugins && \
		sed 's/goa.design\/goa\/v.*/goa.design\/goa\/v$(MAJOR) v$(MAJOR).$(MINOR).$(BUILD)/' go.mod > _tmp && mv _tmp go.mod && \
		make && \
		git add . && \
		git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)" && \
		git tag v$(MAJOR).$(MINOR).$(BUILD) && \
		git push origin v$(MAJOR) && \
		git push origin v$(MAJOR).$(MINOR).$(BUILD)

