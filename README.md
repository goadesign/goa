# <img src="http://goa.design/img/goa-logo.svg">

goa is a framework for building microservices in Go using a unique design-first approach.

[![Build Status](https://travis-ci.org/goadesign/goa.svg?branch=master)](https://travis-ci.org/goadesign/goa)
[![Windows Build status](https://ci.appveyor.com/api/projects/status/vixp37loj5i6qmaf/branch/master?svg=true)](https://ci.appveyor.com/project/RaphaelSimon/goa-oqtis/branch/master)
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

goa takes a different approach to building these applications: instead of focusing solely on helping
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
go get github.com/goadesign/goa
go get github.com/goadesign/goa/goagen
```

## Teaser

### 1. Design

Create the file `$GOPATH/src/goa-adder/design/design.go` with the following content:
```go
package design

import (
        . "github.com/goadesign/goa/design"
        . "github.com/goadesign/goa/design/apidsl"
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
                Response(OK, "text/plain")
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

* `main.go` and `operands.go` contain scaffolding code to help bootstrap the implementation.
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
2016/04/05 20:39:10 [INFO] mount ctrl=Operands action=Add route=GET /add/:left/:right
2016/04/05 20:39:10 [INFO] mount file name=swagger/swagger.json route=GET /swagger.json
2016/04/05 20:39:10 [INFO] listen transport=http addr=:8080
```
Open a new console and compile the generated CLI tool:
```
cd $GOPATH/src/goa-adder/client/adder-cli
go build
```
The tool includes contextual help:
```
$ ./adder-cli --help
CLI client for the adder service

Usage:
  adder-cli [command]

Available Commands:
  add         add returns the sum of the left and right parameters in the response body

Flags:
      --dump[=false]: Dump HTTP request and response.
  -H, --host="localhost:8080": API hostname
      --pp[=false]: Pretty print response body
  -s, --scheme="http": Set the requests scheme
  -t, --timeout=20s: Set the request timeout, defaults to 20s

Use "adder-cli [command] --help" for more information about a command.
```
To get information on how to call a specific API use:
```
$ ./adder-cli add operands --help
Usage:
  adder-cli add operands [/add/LEFT/RIGHT] or [flags]

Flags:
      --left int    Left operand
      --right int   Right operand

Global Flags:
      --dump               Dump HTTP request and response.
  -H, --host string        API hostname (default "localhost:8080")
      --pp                 Pretty print response body
  -s, --scheme string      Set the requests scheme
  -t, --timeout duration   Set the request timeout (default 20s)
```
Now let's run it:
```
$ ./adder-cli add operands /add/1/2
2016/04/05 20:43:18 [INFO] started id=HffVaGiH GET=http://localhost:8080/add/1/2
2016/04/05 20:43:18 [INFO] completed id=HffVaGiH status=200 time=1.028827ms
3⏎
```
This also works:
```
$ ./adder-cli add operands --left=1 --right=2
2016/04/25 00:08:59 [INFO] started id=ouKmwdWp GET=http://localhost:8080/add/1/2
2016/04/25 00:08:59 [INFO] completed id=ouKmwdWp status=200 time=1.097749ms
3⏎     
```
The console running the service shows the request that was just handled:
```
2016/04/05 20:43:18 [INFO] started action=Add id=cASjgqGiCP-1 GET=/add/1/2
2016/04/05 20:43:18 [INFO] params action=Add id=cASjgqGiCP-1 right=2 left=1
2016/04/05 20:43:18 [INFO] completed action=Add id=cASjgqGiCP-1 status=0 bytes=0 time=36.615µs
```
Now let's see how robust our service is and try to use non integer values:
```
./adder-cli add operands add/1/d
2016/04/05 20:44:56 [INFO] started id=5254tL8j GET=http://localhost:8080/add/1/d
2016/04/05 20:44:56 [INFO] completed id=5254tL8j status=500 time=840.12µs
error: 500: "Internal error: 400 invalid_request: invalid value \"d\" for parameter \"right\", must be a integer"
```
As you can see the generated code validated the incoming request against the types defined
in the design.

### 4. Document

The `swagger` directory contains the API Swagger specification in both YAML and JSON format.

For open source projects hosted on github [swagger.goa.design](http://swagger.goa.design) provides a
free service that renders the Swagger representation dynamically from goa design packages. Simply
set the `url` query string with the import path to the design package. For example displaying the
docs for `github.com/goadesign/goa-cellar/design` is done by browsing to:

http://swagger.goa.design/?url=goadesign%2Fgoa-cellar%2Fdesign

Note that the above generates the swagger spec dynamically and does not require it to be present in
the Github repo.

The Swagger JSON can also easily be served from the documented service itself using a simple
[Files](http://goa.design/reference/goa/design/apidsl/#func-files-a-name-apidsl-files-a)
definition in the design, for example:

```go
var _ = Resource("swagger", func() {
        Origin("*", func() {
               Methods("GET") // Allow all origins to retrieve the Swagger JSON (CORS)
        })
        Files("/swagger.json", "swagger/swagger.json")
})
```

The generated controller is then mounted as follows in the `main` function for example:

```go
app.MountSwaggerController(service, service.NewController("swagger"))
```

Requests made to `/swagger.json` now return the Swagger specification. The generated controller also
takes care of adding the proper CORS headers so that the JSON may be retrieved from anywhere e.g.
via Swagger UI.

## Resources

### goa.design

[http://goa.design](http://goa.design) contains further information on goa including a getting
started guide, detailed DSL documentation as well as information on how to implement a goa service.

### Examples

The [examples](https://github.com/goadesign/examples) repo contains simple examples illustrating
basic concepts.

The [goa-cellar](https://github.com/goadesign/goa-cellar) repo contains the implementation for a
goa service which demonstrates many aspects of the design language. It is kept up-to-date and
provides a reference for testing functionality.

## Contributing

Did you fix a bug? write docs or additional tests? or implement some new awesome functionality?
You're a rock star!! Just make sure that `make` succeeds (or that TravisCI is green) and send a PR
over.

The [issues](https://github.com/goadesign/goa/issues) contain entries tagged with
[help wanted:
beginners](https://github.com/goadesign/goa/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
which provide a great way to get started!
