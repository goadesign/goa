# goa [![Build Status](https://travis-ci.org/raphael/goa.svg)](https://travis-ci.org/raphael/goa)

goa is a framework for building RESTful APIs in Go.

## Why goa?

There are a number of good Go packages for writing modular web applications out there so why build
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

The goa DSL allows writing self-explanatory code that describes the resources exposed by the API
and for each resource the properties and actions. goa comes with the `goagen` tool which runs the
DSL and generates various types of artifacts from the resulting metadata.

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
go get github.com/raphael/goa
go get github.com/raphael/goa/goagen
```
The code generation functionality relies on [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports):
```
go get golang.org/x/tools/cmd/goimports
```

## Getting Started

### Writing your first goa application

To make the above more concrete, let's write a simplistic API with goa. Let's take an example of a
"cellar" API that makes it possible to retrieve wine bottle information. To keep things simple the
only action supported by the API is to retrieve a bottle by ID.

The first thing to do when writing a goa application is to describe the API using the goa DSL.
Create a new directory under `$GOPATH/src` for the new goa application, say `$GOPATH/src/cellar`.
In that directory create a `design` sub directory and the file `design/design.go` with the
following content:
```go
package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = API("cellar", func() {
	Title("The virtual wine cellar")
	Description("A basic example of an API implemented with goa")
})

var _ = Resource("bottle", func() {
	BasePath("/bottles")
	DefaultMedia(BottleMedia)
	Action("show", func() {
		Description("Retrieve bottle with given id")
		Routing(GET("/:bottleID"))
		Params(func() {
			Param("bottleID", Integer, "Bottle ID")
		})
		Response(OK)
		Response(NotFound)
	})
})

var BottleMedia = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")
	Attributes(func() {
		Attribute("id", Integer, "Unique bottle ID")
		Attribute("href", String, "API href for making requests on the bottle")
		Attribute("name", String, "Name of wine")
	})
	View("default", func() {
		Attribute("id")
		Attribute("href")
		Attribute("name")
	})
})
```
Let's break this down:
* We define a `design` package and use an anonymous variable to declare the API, we could also have
  used a package `init` function. The actual name of the package could be anything, `design` is just
  a convention.
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
  responses. `goa` defines default response templates for all standard HTTP status code. Custom
  response templates may be defined to specify additional properties such as required headers and
  application specific media types.
* Finally we define the resource media type as a global variable so we can refer to it when
  declaring the `OK` response. A media type has an identifier as defined by [RFC 6838](https://tools.ietf.org/html/rfc6838)
  and describes the attributes of the response body (the JSON object fields in goa).

> The [DSL reference](DSL.md) lists all the goa DSL keywords together with a description and example usage.

Now that we have a design for our API we can use the `goagen` tool to generate all the boilerplate
code. The tool takes the path to the Go package as argument (the same path you'd use if you were to
import the design package in a Go source file). So for example if you created the design package
under `$GOPATH/src/cellar`, the command line would be:
```
goagen -d cellar/design
```
The tool outputs the names of the files it generates - by default it generates the files in the
current working directory. The list should look something like this:
```
app/contexts.go
app/controllers.go
app/hrefs.go
app/media_types.go
app/user_types.go
schema/schema.go
schema/schema.json
main.go
bottle.go
```
Note how `goagen` generated a main for our app as well as a skeleton controller (`bottle.go`). These
two files are meant to help bootstrap a new development, they won't be re-generated (by default) if
already present (re-run the tool again and note how it only generates the files under the "app" and
"schema" directories this time). This behavior and many other aspects are configurable via command
line arguments, see the [goagen docs](goagen.md) for details.

Back to the list of generated files:

* The `app` directory contains the underpinning algorithms that glue the low level HTTP router with
  your code.
* The `schema` directory contains a JSON Hyper-schema representation of the API together with the
  implementation of a controller that serves the file when requests are sent to `/schema`.
* As discussed above the `main.go` and `bottle.go` files provide a starting point for implementing
  the application entry point and the `bottle` controller respectively.

Looking at the content of the `app` package:

* `controllers.go` contains the controller interface type definitions. There is one such interface
  per resource defined in the DSL. The file also contains the code that "mount" implementations of
  these controller interfaces onto the application. The exact meaning of "mounting" a controller
  is discussed further below.
* `contexts.go` contains the context data structure definitions. Contexts play a similar role
  to Martini's `martini.Context`, goji's `web.C` or echo's `echo.Context` to take a few arbitrary
  examples: they are given as first argument to all controller actions and provide helper methods to
  retrieve the action parameters or write the response.
* `hrefs.go` provide global functions for building resource hrefs. Resource hrefs make it possible
  for responses to link to related resources. goa knows how to build these hrefs by looking at the
  request path for the resource "canonical" action (by default the "show" action). See the Action
  DSL function for additional information.
* `media_types.go` contains the media type data structures used by resource actions to build the
  responses. These data structures also expose methods that can be called to instantiate them from
  raw data or dump them back to raw data performing all the validations defined in the DSL in both
  cases.
* `user_types.go` contains the data structures defined via the "Type" DSL function. Such types may
  be used to define a request payload or response media types.

Now that `goagen` did its work the only thing left for us to do is to provide an implementation of
the `bottle` controller. The type definition generated by `goagen` is:
```go
type BottleController interface {
 	Show(*ShowBottleContext) error
}
```
Simple enough... Let's take a look at the definition of `ShowBottleContext` in `app/contexts.go`:
```go
// ShowBottleContext provides the bottle show action context.
type ShowBottleContext struct {
	*goa.Context
	BottleID int
}
```
The same file also defines two methods on the context data structure:
```go
// NotFound sends a HTTP response with status code 404.
func (c *ShowBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (c *ShowBottleContext) OK(resp *BottleMedia) error {
	r, err := resp.Dump()
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	return c.JSON(200, r)
}
```
`goagen` also provided an empty implementation of the controller in `bottle.go` so all we have left
to do is provide an actual implementation. Open the file `bottle.go` and replace the existing
`ShowBottle` function with:
```go
// ShowBottle implements the "show" action of the "bottles" controller.
func ShowBottle(ctx *app.ShowBottleContext) error {
	if ctx.ID == 0 {
		// Emulate a missing record with ID 0
		return ctx.NotFound()
	}
	// Build the resource using the generated data structure
	bottle := app.Bottle{ID: ctx.ID, Name: fmt.Sprintf("Bottle #%d", ctx.ID)}

	// Let the generated code produce the HTTP response using the
	// media type described in the design (BottleMedia).
	return ctx.OK(&bottle)
}
```
Before we build and run the app, let's take a look at `main.go`. The file contains a default
implementation of `main` that instantiates a new goa application, initializes default middleware,
mounts the `bottle` and `schema` controllers and finally run the HTTP server:
```go
func main() {
	// Create application
	api := goa.New("cellar")

	// Setup middleware
	api.Use(goa.Recover())
	api.Use(goa.RequestID())
	api.Use(goa.LogRequest())

	// Mount "bottles" controller
	c := NewBottleController()
	app.MountBottleController(api, c)

	// Mount JSON schema provider controller
	schema.MountController(api)

	// Run application, listen on port 8080
	api.Run(":8080")
}
```
Now compile and run the application:
```
go build -o cellar
./cellar
```
This should produce something like:
```
INFO[07-26|21:33:03] mouting                                  app=cellar ctl=Bottle
INFO[07-26|21:33:03] handler                                  app=cellar ctl=Bottle action=Show GET=/bottles/:bottleID
INFO[07-26|21:33:03] mounted                                  app=cellar ctl=Bottle
INFO[07-26|21:33:03] mouting                                  app=cellar ctl=Schema
INFO[07-26|21:33:03] handler                                  app=cellar ctl=Schema action=GetSchema GET=/schema
INFO[07-26|21:33:03] mounted                                  app=cellar ctl=Schema
INFO[07-26|21:33:03] listen                                   app=cellar addr=:8080
```
You can now test the app using `curl` for example:
```
$ curl -i localhost:8080/bottles/1
HTTP/1.1 200 OK
Date: Mon, 27 Jul 2015 04:35:22 GMT
Content-Length: 40
Content-Type: text/plain; charset=utf-8

{"ID":1,"Href":"/bottles/1","Name":"Bottle #1"}

$ curl -i localhost:8080/bottles/0
HTTP/1.1 404 NotFound
Date: Mon, 27 Jul 2015 04:35:22 GMT
```
You can exercise the validation code that `goagen` generated for you by passing an invalid (non-
integer) id:
```
$ curl -i localhost:8080/bottles/a
HTTP/1.1 400 Bad Request
Date: Mon, 27 Jul 2015 04:37:09 GMT
Content-Length: 17
Content-Type: text/plain; charset=utf-8

invalid value 'a' for parameter id, must be a int
```
Finally request the API JSON Hyper-schema with:
```
$ curl -i localhost:8080/schema
```
That's it! congratulations on writing you first goa application!

This example only covers a fraction of what goa can do but is hopefully enough to illustrate the
benefits of design-based API development.

Consult [goa, the Language](DSL.md) for more information.
