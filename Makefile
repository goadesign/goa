#! /usr/bin/make
#
# Makefile for Goa v3
#
# Targets:
# - "depend" retrieves the Go packages needed to run the linter and tests
# - "lint" runs the linter and checks the code format using goimports
# - "test" runs the tests
# - "release" creates a new release commit, tags the commit and pushes the tag to GitHub.
#   "release" also updates the examples and plugins repo and pushes the updates to GitHub.
#
# Meta targets:
# - "all" is the default target, it runs "lint" and "test"
#
MAJOR=3
MINOR=1
BUILD=3

GOOS=$(shell go env GOOS)
GO_FILES=$(shell find . -type f -name '*.go')
GOPATH=$(shell go env GOPATH)

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	golang.org/x/lint/golint \
	golang.org/x/tools/cmd/goimports \
	github.com/golang/protobuf/protoc-gen-go \
	github.com/golang/protobuf/proto \
	honnef.co/go/tools/cmd/staticcheck \
	github.com/hashicorp/go-getter/cmd/go-getter

all: lint test

travis: depend all #test-examples test-plugins

# Install protoc
PROTOC_VERSION=3.11.4
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
GOPATH:=$(subst \,/,$(GOPATH))
		endif
	endif
endif
depend:
	@echo donwloading dependencies
	@go mod download
	@go get -v $(DEPEND) # Additional development dependencies
	@echo installing protoc
	go-getter https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC).zip $(PROTOC)
	@cp $(PROTOC_EXEC) $(GOPATH)/bin && \
		rm -r $(PROTOC) && \
		echo "`protoc --version`"
	@echo done installing dependencies

lint:
ifneq ($(GOOS),windows)
	@if [ "`goimports -l $(GO_FILES) | tee /dev/stderr`" ]; then \
		echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
	fi
	@if [ "`golint ./... | grep -vf .golint_exclude | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi
	@if [ "`staticcheck -checks all ./... | grep -v ".pb.go" | grep -v "SA1019" | tee /dev/stderr`" ]; then \
		echo "^ - staticcheck errors!" && echo && exit 1; \
	fi
endif

test:
	env GO111MODULE=on go test ./...

release:
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
	go mod tidy
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
	# Update examples
	cd $(GOPATH)/src/goa.design/examples && \
		sed 's/goa.design\/goa\/v.*/goa.design\/goa\/v$(MAJOR) v$(MAJOR).$(MINOR).$(BUILD)/' go.mod > _tmp && mv _tmp go.mod && \
		make && \
		git add . && \
		git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)" && \
		git tag v$(MAJOR).$(MINOR).$(BUILD) && \
		git push origin master
		git push origin v$(MAJOR).$(MINOR).$(BUILD)
	# Update plugins
	cd $(GOPATH)/src/goa.design/plugins && \
		sed 's/goa.design\/goa\/v.*/goa.design\/goa\/v$(MAJOR) v$(MAJOR).$(MINOR).$(BUILD)/' go.mod > _tmp && mv _tmp go.mod && \
		make && \
		git add . && \
		git commit -m "Release v$(MAJOR).$(MINOR).$(BUILD)" && \
		git tag v$(MAJOR).$(MINOR).$(BUILD) && \
		git push origin v$(MAJOR) && \
		git push origin v$(MAJOR).$(MINOR).$(BUILD)
	echo DONE RELEASING v$(MAJOR).$(MINOR).$(BUILD)!

