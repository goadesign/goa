# goa

goa is a framework for building RESTful microservices in Go.

[![Build Status](https://travis-ci.org/raphael/goa.svg?branch=master)](https://travis-ci.org/raphael/goa)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/raphael/goa/blob/master/LICENSE)
[![Godoc](https://godoc.org/github.com/raphael/goa?status.svg)](http://godoc.org/github.com/raphael/goa)
[![Slack](https://img.shields.io/badge/slack-gophers-orange.svg?style=flat)](https://gophers.slack.com/messages/goa/)
[![Intro](https://img.shields.io/badge/post-goa-ff69b4.svg?style=flat)](https://blog.gopheracademy.com/advent-2015/goaUntanglingMicroservices/)

## Why goa?

There are a number of good Go packages for writing modular web services out there so why build
another one? Glad you asked! The existing packages tend to focus on providing small and highly
modular frameworks that are purposefully narrowly focused. The intent is to keep things simple and
to avoid mixing concerns.

This is great when writing simple APIs that tend to change rarely. However there are a number of
problems that any non trivial API implementation must address. Things like request validation,
response media type definitions or documentation are hard to do in a way that stays consistent and
flexible as the API surface evolves.

goa takes a different approach to building web applications: instead of focusing solely on helping
with implementation, goa makes it possible to describe the *design* of an API in an holistic way.
goa then uses that description to provide specialized helper code to the implementation and to
generate documentation, API clients, tests, even custom artifacts.

The goa design language allows writing self-explanatory code that describes the resources exposed
by the API and for each resource the properties and actions. goa comes with the `goagen` tool which
runs the design language and generates various types of artifacts from the resulting metadata.

One of the `goagen` output is glue code that binds your code with the underlying HTTP server. This
code is specific to your API so that for example there is no need to cast or "bind" any handler
argument prior to using them. Each generated handler has a signature that is specific to the
corresponding resource action. It's not just the parameters though, each handler also has access to
specific helper methods that generate the possible responses for that action. The metadata can also
include validation rules so that the generated code also takes care of validating the incoming
request parameters and payload prior to invoking your code.

The end result is controller code that is terse and clean, the boilerplate is all gone. Another big
benefit is the clean separation of concern between design and implementation: on bigger projects
it's often the case that API design changes require careful review, being able to generate a new
version of the documentation without having to write a single line of implementation is a big boon.

This idea of separating design and implementation is not new, the excellent [Praxis](http://praxis-framework.io)
framework from RightScale follows the same pattern and was an inspiration to goa.

## The Whys and Hows

If you are new to goa I can't recommend enough that you read the
[Gopher Academy blog post](https://blog.gopheracademy.com/advent-2015/goaUntanglingMicroservices/).
goa may look a little bit different at first, the post explains the thinking behind it so that you
can better take advantage of the framework.

## Installation

Assuming you have a working Go setup:
```
go get github.com/raphael/goa/goagen
```
The code generation functionality relies on [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports):
```
go get golang.org/x/tools/cmd/goimports
```

## Development Workflow

1. Create API design using the [goa design language](https://godoc.org/github.com/raphael/goa/design/dsl).
2. [Optional] If API design package is a public github repo use [swagger.goa.design](http://swagger.goa.design) to verify the design.
3. Run [`goagen`](http://www.goa.design/goagen.html): `goagen bootstrap -d <design package path>`
4. Fill-in implementation of generated controller actions.

## Getting Started

Can't wait to give it a try? the easiest way is to follow the short
[getting started](http://www.goa.design/getting-started.html) guide.

## Contributing

Did you fix a bug? write docs or additional tests? or implement some new awesome functionality?
You're a rock star!! Just make sure that `make` succeeds (or that TravisCI is green) and send a PR
over.

And if you're looking for inspiration the [wookie](https://github.com/raphael/goa/wiki) contains a
roadmap document with many good suggestions...
