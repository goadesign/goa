# goa

goa is a framework for building RESTful microservices in Go.

[![Build Status](https://travis-ci.org/raphael/goa.svg?branch=master)](https://travis-ci.org/raphael/goa)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/raphael/goa/blob/master/LICENSE)
[![Godoc](https://godoc.org/github.com/raphael/goa?status.svg)](http://godoc.org/github.com/raphael/goa)
[![Slack](https://img.shields.io/badge/slack-goa-ff69b4.svg?style=flat)](https://gophers.slack.com/messages/goa/)

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

![goagen diagram](https://cdn.rawgit.com/raphael/goa/master/images/goagenv3.svg "goagen")
## Runtime

While goa and `goagen` help you get going quickly, goa strives to make as little assumption as
possible and as such is **not** an opinionated framework. The only loose requirement is that the
API exposes some form of resources which only serve as a way to group API endpoints (*actions*).
The semantic backing these resource is irrelevant to goa and is left completely up to you.

### Request Contexts

Resources are implemented through *controllers* - each exposing the underlying resource actions.
The action methods all accept a context as first argument. The context is bound to the request
and provides:
* A concurrency-safe way of storing and retrieving state.
* The ability to set deadlines and send cancellation signals. See the [Timeout middleware](https://godoc.org/github.com/raphael/goa#Timeout)
  for an example. goa automatically sends a cancellation signal upon request completion.
* "Typed" access to the request state. This means custom data structures generated from the API
  design that are initialized and validated by goa prior to calling your code.
* Response helper methods also generated from the API design that take care of serializing any
  data structure that needs to be written to the response.

### Logging

The request contexts implement the Logger interface from inconshreveable's excellent [log15 package](https://godoc.org/gopkg.in/inconshreveable/log15.v2). This provides goa with structured
logging where the log context is inherited all the way from the service to the request context,
each log entry contains the name of the service, the controller and the action as well as a request
specific identifier. The logger can be configured to write to many different backends including
standard output or error but also syslog or even loggly.

### Middleware

goa supports both [classic middleware](http://www.alexedwards.net/blog/making-and-using-middleware) implemented
using the http package or middleware using goa request handlers which have access to the request context.
goa comes with several middleware that address common needs such as request logging, support for the
[X-Request-Id](https://devcenter.heroku.com/articles/http-request-id) header, Timeout enforcement
through the [context.Context](https://godoc.org/golang.org/x/net/context#Context) interface implemented
by the goa request contexts, etc.
Middleware can be mounted globally to the service via the [Service](https://godoc.org/github.com/raphael/goa#Service)
interface or on a specific controller via the [Controller](https://godoc.org/github.com/raphael/goa#Controller) interface.

### Graceful Shutdown

goa services may be instantiated through the [NewGraceful](https://godoc.org/github.com/raphael/goa#NewGraceful) method
which return an HTTP server backed by the [graceful](https://godoc.org/gopkg.in/tylerb/graceful.v1) package.
This means that sending a TERM signal to the goa process won't kill ongoing requests - instead a
graceful shutdown is initiated preventing new requests from being accepted and waiting until ongoing
requests return prior to exiting the process.

### Error Handling

goa also supports error handling via service-wide or controller specific error handlers: an action
that returns a non nil *error* triggers the controller error handler if one is defined - the service
wide error handler otherwise. The default error handler responds with status code 500 providing the
error details in the body. goa comes with the [Terse](https://godoc.org/github.com/raphael/goa#TerseErrorHandler)
error handler which does not write the error details to the response body in case of internal errors.

## Getting Started

Can't wait to give it a try? the easiest way is to follow the short [getting started](http://www.goa.design/getting-started.html) guide.

## Contributing

Did you fix a bug? write docs or additional tests? or implement some new awesome functionality?
You're a rock star!! Just make sure that `make` succeeds (or that TravisCI is green) and send a PR
over.

And if you're looking for inspiration the [wookie](https://github.com/raphael/goa/wiki) contains a
roadmap document with many good suggestions...
