# goa [![Build Status](https://travis-ci.org/raphael/goa.svg)](https://travis-ci.org/raphael/goa)

goa is a framework for building RESTful APIs in Go.

## Why goa?

There are a number of good Go packages for writing modular web
applications out there so why build another one? Glad you asked :)
The existing packages tend to focus on providing small and highly
modular frameworks that are purposefully narrowly focused. The
intent is to keep things simple and to avoid mixing concerns.

This is great when writing simple APIs that tend to change rarely
however there are a number of problems that any non trivial API
implementation must address. Things like documentation, request
validation, response media type definitions are hard to do
in a way that stays consistent and flexible as the API surface
grows.

goa takes a different approach to building web applications: instead of
focusing solely on helping with implementation, goa makes it possible
to describe the *design* of an API in an holistic way. goa then uses that
description to provide specialized helper code to the implementation,
and in the future generate documentation, API clients, tests and others
(see TODO if you would like to contribute).

The goa DSL allows writing self-explanatory code that describes the
API, the resources it exposes and for each resource its properties
and actions. The DSL gets translated into metadata that describes the
API. goa comes with the `codegen` tool which generates both code and
(in the future) documentation from that metadata.

The resulting code is specific to your API so that for example
there is no need to cast or bind any handler argument prior to
using them. Each generated handler has a signature that is specific
to the corresponding resource action. It's not just the parameters
though, each handler also has access to specific helper methods to
generate the possible responses for that action. The metadata can
also include validation rules so that the generated code also takes
care of validating the incoming request parameters and payload prior
to invoking your code.

The end result is controller code that is terse and clean, the
boilerplate is all gone. Another big benefit is the clean separation
of concern between design and implementation: on bigger projects it's
often the case that API design changes require careful review, being
able to generate a new version of the documentation without having to
write a single line of implementation is a big boon.

This idea of separating design and implementation is not new, the
[Praxis](http://praxis-framework.io) framework from RightScale
follows the same pattern and was a big inspiration for goa.

## Installation

Assuming you have a working Go setup:
```
go get github.com/raphael/goa
```

## Getting Started

### Writing your first goa application

The first thing to do when writing a goa application is to describe
the API via the goa DSL, create a file `design/design.go` with the
following content:
```go
package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = API("cellar", func() {
	Title("The virtual wine cellar")
	Description("A basic example of a CRUD API implemented with goa")
})

var _ = Resource("bottle", func() {
	MediaType(BottleMediaType)
	Action("show", func() {
		Description("Retrieve bottle with given id")
		Routing(
			GET("/:id"),
		)
		Params(func() {
			Param("id", Integer, "Account ID")
		})
		Response(OK, BottleMediaType)
		Response(NotFound)
	})
})

var BottleMediaType = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")
	Attributes(func() {
		Attribute("id", Integer, "Unique bottle ID")
		Attribute("href", String, "API href for making requests on the bottle")
		Attribute("name", String, "Name of wine")
	})
})
```
Let's break this down:
* We define a `design` package and use an anonymous variable to declare the API, we could also have
  used a package `init` function.
* The `API` function takes two arguments: the name of the API and an anonymous function that
  defines additional properties, here a title and a description.
* We then declare a resource "bottle" using the `Resource` function which also takes a name and an
  anonymous function. Properties defined in the anonymous function includes the actions supported by
  the resource.
* Each resource action is declared using the `Action` function which follows the same pattern of
  name and anonymous function. Actions are defined in resources, they can be CRUD
  (Create/Read/Update/Delete) actions or so-called "custom" actions. Here we define a single Read
  (`show`) action.
* The `Action` function defines the action endpoint, parameters, payload (not used here) and
  responses.
* Finally we define the resource media type as a global variable so we can refer to it when
  declaring the `OK` response. A media type has a name as defined by [RFC 6838](https://tools.ietf.org/html/rfc6838)
  and describes the attributes of the response body (the JSON object fields in goa).

The DSL reference contains more details for each of the functions used in the example above and
others that help define payload types and custom responses as well.

Now that we have a design for the API we can use the `codegen` tool to generate all the boilerplate
for our app. The tool takes the path to the Go package as argument (the same path you'd use if you
were to import the design package in a Go source file). So for example if you created the design
package under `$GOPATH/src/app`, the command line would be:
```
codegen -d app/design
```
This creates a `autogen` folder containing three files:
* `resources.go` contains the bottle resource data structure definition.
* `contexts.go` contains the context data structure definitions. Contexts play a similar role
  to Martini's `martini.Context`, goji's `web.C` or echo's `echo.Context` to take a few arbitrary
  examples: they are given as argument to controller actions and provide helper methods to
  retrieve the action parameters or write the response.
* `handlers.go` provide the glue between the underlying go http server handler and the controller
  actions. They create the action specific context and call your code.

`codegen` also created the files `main.go` and `bottle.go` which provide a starting point for your
app. The next and final step consists of implementing the `bottle` resource `show` action, open the
 file `bottle.go` and replace the existing `ShowBottle` function with:
```go
// ShowBottle implements the "show" action of the "bottles" controller.
func ShowBottle(c *autogen.ShowBottleContext) error {
	if c.ID == 0 {
		// Emulate a missing record with ID 0
		return c.NotFound()
	}
	// Build the resource using the generated data structure
	bottle := autogen.Bottle{ID: c.ID, Name: fmt.Sprintf("Bottle #%d", c.ID)}

	// Let the generated code produce the HTTP response using the
	// media type described in the design (BottleMediaType).
	return c.OK(&bottle)
}
```
Looking at the content of the file main.go:
* The `main` function first creates a new controller for the `bottles` resource.
* It then associates each resource action with the function that implements it using the
  `SetHandlers` controller method. Here the only action of the bottles resource - `show` - is
  implemented by the `ShowBottle` function.
* Finally it creates a goa application, mounts the `bottles` controller on it and runs it. Mounting
  a controller on an application causes `goa` to validate the actions (i.e. make sure they
  all have handlers, that handlers have the proper signature etc.). An error at this point stops
  the application to avoid running a mis-configured application.
* The `ShowBottle` handler contains the application specific logic. It leverages the generated
  `ShowBottleContext` data structure which exposes the action `ID` parameter as an `int` so that
  no cast is required. The context also exposes a `OK` and `NotFound` functions that take care of
  creating the corresponding HTTP responses including any specific header, the appropriate status
  code and serialized body in the case of "OK".

Now compile and run the application:
```
go build -o app
./app
```
This should produce something like:
```
INFO[07-26|21:33:03] mouting                                  app=cellar ctl=bottles
INFO[07-26|21:33:03] handler                                  app=cellar ctl=bottles action=show GET=/:id
INFO[07-26|21:33:03] mounted                                  app=cellar ctl=bottles
INFO[07-26|21:33:03] listen                                   app=cellar addr=:8080
```
You can now test the app using `curl` for example:
```
$ curl -i localhost:8080/1
HTTP/1.1 200 OK
Date: Mon, 27 Jul 2015 04:35:22 GMT
Content-Length: 40
Content-Type: text/plain; charset=utf-8

{"ID":1,"Href":"/1","Name":"Bottle #1"}

$ curl -i localhost:8080/0
HTTP/1.1 404 NotFound
Date: Mon, 27 Jul 2015 04:35:22 GMT
```
You can exercise the validation code that `codegen` generated for you by passing an invalid (non-
integer) id:
```
$ curl -i localhost:8080/a
HTTP/1.1 400 Bad Request
Date: Mon, 27 Jul 2015 04:37:09 GMT
Content-Length: 17
Content-Type: text/plain; charset=utf-8

invalid value 'a' for parameter id, must be a int
```
That's it! congratulations on writing you first goa application!

This example only covers a fraction of what goa can do, additional documentation is in the works.
As next steps try `codegen --help` to get a sense of what can be tweaked.
