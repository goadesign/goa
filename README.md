# goa

goa is a framework for building RESTful microservices in Go.

[![Build Status](https://travis-ci.org/goadesign/goa.svg?branch=master)](https://travis-ci.org/goadesign/goa)
[![Windows Build status](https://ci.appveyor.com/api/projects/status/veujsyg4sxod40vw/branch/master?svg=true)](https://ci.appveyor.com/project/RaphaelSimon/goa/branch/master)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/goadesign/goa/blob/master/LICENSE)
[![Godoc](https://godoc.org/github.com/goadesign/goa?status.svg)](http://godoc.org/github.com/goadesign/goa)
[![Slack](https://img.shields.io/badge/slack-gophers-orange.svg?style=flat)](https://gophers.slack.com/messages/goa/)
[![Intro](https://img.shields.io/badge/post-gopheracademy-ff69b4.svg?style=flat)](https://blog.gopheracademy.com/advent-2015/goaUntanglingMicroservices/)

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

## Other Whys and Hows

If you are new to goa I can't recommend enough that you read the
[Gopher Academy blog post](https://blog.gopheracademy.com/advent-2015/goaUntanglingMicroservices/).
goa may look a little bit different at first, the post explains the thinking behind it so that you
can better take advantage of the framework.

## Installation

Assuming you have a working Go setup:
```
go get github.com/goadesign/goa/goagen
```
The code generation functionality relies on [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports):
```
go get golang.org/x/tools/cmd/goimports
```

## Teaser

### 1. Design

Create the file `$GOPATH/src/goa-adder/design/design.go` with the following content:
```go
package design

import (
        . "github.com/goadesign/goa/design"
        . "github.com/goadesign/goa/design/dsl"
)

var _ = API("adder", func() {
        Title("The adder API")
        Description("A teaser for goa")
        Host("localhost:8080")
        Scheme("http")
})

var _ = Resource("operands", func() {
        Action("add", func() {
                Routing(GET("add/:left/:right"))
                Description("add returns the sum of the left and right parameters in the response body")
                Params(func() {
                        Param("left", Integer, "Left operand")
                        Param("right", Integer, "Right operand")
                })
                Response(OK, "plain/text")
        })

})
```
This file contains the design for an `adder` API which accepts HTTP GET requests to `/add/:x/:y`
where `:x` and `:y` are placeholders for integer values. The API returns the sum of `x` and `y` in
its body.

### 2. Implement

Now that the design is done, let's run `goagen` on the design package:
```
$ cd $GOPATH/src/goa-adder
$ goagen bootstrap -d goa-adder/design
```
This produces the following outputs:

* `main.go` and `adder.go` contain scaffolding code to help bootstrap the implementation.
  running `goagen` again does no recreate them so that it's safe to edit their content.
* an `app` package which contains glue code that binds the low level HTTP server to your
  implementation.
* a `client` package with a `Client` struct that implements a `AddOperands` function which calls
  the API with the given arguments and returns the `http.Response`. The `client` directory also
  contains the complete source for a client CLI tool (see below).
* a `swagger` package with implements the `GET /swagger.json` API endpoint. The response contains
  the full Swagger specificiation of the API.

### 3. Run

First let's implement the API - edit the file `operands.go` and replace the content of the `Add`
function with:
```
// Add runs the add action.
func (c *OperandsController) Add(ctx *app.AddOperandsContext) error {
        sum := ctx.Left + ctx.Right
        return ctx.OK([]byte(strconv.Itoa(sum)))
}
```
Now let's compile and run the service:
```
$ cd $GOPATH/src/goa-adder
$ go build
$ ./goa-adder
INFO[01-04|08:24:06] mount                                    app=API ctrl=Operands action=Add route="GET /add/:left/:right"
INFO[01-04|08:24:06] mount                                    app=API file=swagger/swagger.json route="GET /swagger.json"
INFO[01-04|08:24:06] listen                                   app=API addr=:8080
```
Open a new console and compile the generated CLI tool:
```
cd $GOPATH/src/goa-adder/client/adder-cli
go build
```
The tool includes contextual help:
```
$ ./adder-cli --help
usage: adder-cli [<flags>] <command> [<args> ...]

CLI client for the adder service

Flags:
      --help           Show context-sensitive help (also try --help-long and --help-man).
  -s, --scheme="http"  Set the requests scheme
  -h, --host=HOST      API hostname
  -t, --timeout=20s    Set the request timeout, defaults to 20s
      --dump           Dump HTTP request and response.
      --pp             Pretty print response body

Commands:
  help [<command>...]
    Show help.

  add operands <path>
    add adds the left and right parameters and returns the result
```

```
$ ./adder-cli add operands --help
usage: adder-cli add operands <path>

Args:
  <path>  Request path, format is /add/:left/:right
```
Now let's run it:
```
$ ./adder-cli add operands /add/1/2
INFO[01-04|08:30:43] started                                  id=+LG8rvid GET=http://localhost:8080/add/1/2
INFO[01-04|08:30:43] completed                                id=+LG8rvid status=200 time=842.472µs
3
```
The console running the service shows the request that was just handled:
```
INFO[01-04|08:30:43] started                                  app=API ctrl=OperandsController action=Add id=k5QShkGsd5-1 GET=/add/1/2
DBUG[01-04|08:30:43] params                                   app=API ctrl=OperandsController action=Add id=k5QShkGsd5-1 right=2 left=1
INFO[01-04|08:30:43] completed                                app=API ctrl=OperandsController action=Add id=k5QShkGsd5-1 status=200 bytes=1 time=61.176µs
```
Now let's see how robust our service is and try to use non integer values:
```
./adder-cli add operands add/1/d
INFO[01-04|08:32:53] started                                  id=hq3zYwXp GET=http://localhost:8080/add/1/d
INFO[01-04|08:32:53] completed                                id=hq3zYwXp status=400 time=847.297µs
error: 400: [{"id":1,"title":"invalid parameter value","msg":"invalid value \"d\" for parameter \"right\", must be a integer"}]
```
As you can see the generated code validated the incoming request state against the types defined
in the design.

### 4. Document

The `swagger` directory contains the entire Swagger specification in the `swagger.json` file. The
specification can also be accessed through the service:
```
$ curl localhost:8080/swagger.json
```
For open source services hosted on github [swagger.goa.design](http://swagger.goa.design) provides
a free service that renders the Swagger representation dynamically from goa design packages.

## Resources

### GoDoc

* Package [goa](https://godoc.org/github.com/goadesign/goa) contains the data structures and algorithms
  used at runtime.
* Package [dsl](https://godoc.org/github.com/goadesign/goa/design/dsl) contains the implementation of
  the goa design language.
* Package [design](https://godoc.org/github.com/goadesign/goa/design) defines the output data
  structures of the design language.

### Website

[http://goa.design](http://goa.design) contains further information on goa.

### Getting Started

Can't wait to give it a try? the easiest way is to follow the short
[getting started](http://www.goa.design/getting-started.html) guide.


### Middleware

The [goa-middleware](https://github.com/goadesign/goa-middleware) repo provides a number of
middlewares covering most common needs. It also provides a good source of examples for writing new
middlewares.

### Examples

The [goa-cellar](https://github.com/goadesign/goa-cellar) repo contains the implementation for a
goa service which demonstrates many aspects of the design language. It is kept up-to-date and
provides a reference for testing functionality.

## Contributing

Did you fix a bug? write docs or additional tests? or implement some new awesome functionality?
You're a rock star!! Just make sure that `make` succeeds (or that TravisCI is g104.197.54.215reen) and send a PR
over.

And if you're looking for inspiration the [wookie](https://github.com/goadesign/goa/wiki) contains a
roadmap document with many good suggestions...
