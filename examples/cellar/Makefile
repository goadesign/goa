#! /usr/bin/make
#
# Makefile for goa cellar example
#
# Targets:
# - clean     delete all generated files
# - generate  (re)generate all goagen-generated files.
# - build     compile executable
# - ae-build  build appengine 
# - ae-dev    deploy to local (dev) appengine
# - ae-deploy deploy to appengine
#
# Meta targets:
# - all is the default target, it runs all the targets in the order above.
#
CURRENT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: clean generate build

clean:
	@rm -rf app
	@rm -rf client
	@rm -rf swagger
	@rm -f cellar

generate:
	@goagen app     -d github.com/raphael/goa/examples/cellar/design
	@goagen client  -d github.com/raphael/goa/examples/cellar/design
	@goagen swagger -d github.com/raphael/goa/examples/cellar/design

build:
	@go build -o cellar

ae-build:
	@if [ ! -d $(HOME)/cellar ]; then \
		mkdir $(HOME)/cellar; \
		ln -s $(CURRENT_DIR)/appengine.go $(HOME)/cellar/appengine.go; \
		ln -s $(CURRENT_DIR)/app.yaml     $(HOME)/cellar/app.yaml; \
	fi

ae-dev: ae-build
	@goapp serve $(HOME)/cellar

ae-deploy: ae-build
	@goapp deploy -application goa-cellar $(HOME)/cellar
