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
MINOR=6
BUILD=2

GOOS=$(shell go env GOOS)
GO_FILES=$(shell find . -type f -name '*.go')
GOPATH=$(shell go env GOPATH)

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1 \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1 \
	honnef.co/go/tools/cmd/staticcheck@v0.2.2

all: lint test

travis: depend all #test-examples test-plugins

# Install protoc
PROTOC_VERSION=3.19.4
UNZIP=unzip
ifeq ($(GOOS),linux)
	PROTOC=protoc-$(PROTOC_VERSION)-linux-x86_64
	PROTOC_EXEC=$(PROTOC)/bin/protoc
endif
ifeq ($(GOOS),darwin)
	PROTOC=protoc-$(PROTOC_VERSION)-osx-x86_64
	PROTOC_EXEC=$(PROTOC)/bin/protoc
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
	@go mod tidy -compat=1.17
	@echo INSTALLING PROTOC...
	@mkdir $(PROTOC)
	@cd $(PROTOC); \
	curl -O -L https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC).zip; \
	$(UNZIP) $(PROTOC).zip
	@cp $(PROTOC_EXEC) $(GOPATH)/bin && \
		rm -r $(PROTOC) && \
		echo "`protoc --version`"

lint:
ifneq ($(GOOS),windows)
	@if [ "`staticcheck -checks all ./... | grep -v ".pb.go" | tee /dev/stderr`" ]; then \
		echo "^ - staticcheck errors!" && echo && exit 1; \
	fi
endif

test:
	go test ./...

release: release-goa release-examples release-plugins

release-goa:
	# First make sure all is clean
	git diff-index --quiet HEAD
	cd $(GOPATH)/src/goa.design/examples && \
		git checkout master && \
		git pull origin master && \
		git diff-index --quiet HEAD
	cd $(GOPATH)/src/goa.design/plugins && \
		git checkout v$(MAJOR) && \
		git pull origin v$(MAJOR) && \
		git diff-index --quiet HEAD
	go mod tidy -compat=1.17
	# Bump version number, commit and push
	sed 's/Major = .*/Major = $(MAJOR)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/Minor = .*/Minor = $(MINOR)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/Build = .*/Build = $(BUILD)/' pkg/version.go > _tmp && mv _tmp pkg/version.go
	sed 's/Current Release: `v3\..*/Current Release: `v$(MAJOR).$(MINOR).$(BUILD)`/' README.md > _tmp && mv _tmp README.md
	sed 's/goa\/v3@v.*tab=doc/goa\/v3@v$(MAJOR).$(MINOR).$(BUILD)\/dsl?tab=doc/' README.md > _tmp && mv _tmp README.md
	git add .
	git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)"
	git tag v$(MAJOR).$(MINOR).$(BUILD)
	cd cmd/goa && go install
	git push origin v$(MAJOR)
	git push origin v$(MAJOR).$(MINOR).$(BUILD)

release-examples:
	cd $(GOPATH)/src/goa.design/examples && \
		sed 's/goa.design\/goa\/v.*/goa.design\/goa\/v$(MAJOR) v$(MAJOR).$(MINOR).$(BUILD)/' go.mod > _tmp && mv _tmp go.mod && \
		make && \
		git add . && \
		git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)" && \
		git tag v$(MAJOR).$(MINOR).$(BUILD) && \
		git push origin master && \
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
	echo DONE RELEASING v$(MAJOR).$(MINOR).$(BUILD)!

