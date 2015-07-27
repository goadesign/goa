# goa

goa is a framework for building RESTful APIs in go.

## Why goa?

There are a number of good good go packages for writing modular web
applications out there so why build another one? Glad you asked...
goa takes a slightly different approach to the problem: instead of
focusing solely on helping with implementation, goa makes it possible
to describe the *design* of an API explicitely.

The goa DSL allows writing self-explanatory code that describes the
API, the resources it exposes and for each resource its properties
and actions. The DSL gets compiled into metadata that describes your
API. goa comes with the `goa` tool which can generate both code and
documentation from that metadata.

The resulting code is specific to your API so that for example
there is no need to cast or bind any handler argument prior to
using them. Each generated handler has a signature that is specific
to the corresponding resource action. It's not just the parameters
though, each handler also has access to specific helper methods to
generate the possible responses for that action. The metadata can
also include validation rules so that the generated code can take
care of validating the incoming request parameters and payload.

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

Assuming you have a working go setup, get the stable version with:
```
go get gopkg.in/raphael/goa.v1
```
or the latest development version with:
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

import . "github.com/raphael/goa/design/dsl"

func init() {

	API("cellar", func() {
		Title("The virtual wine cellar")
		Description("A basic example of a CRUD API implemented with goa")
	})

	Resource("bottle", func() {
		MediaType(BottleMediaType)

		Action("show", func() {
			Description("Retrieve bottle with given id")
			Routing(
				GET("/:id"),
			)
			Params(
				Param("id", Integer, "Account ID"),
			)
			Responses(
				goa.OK(BottleMediaType),
			)
		})
	})
}

var BottleMediaType = MediaType("application/vnd.goa.example.bottle", func() {
	Description("A bottle of wine")

	Attributes(func() {
		Attribute("id", Integer, "ID of bottle")
		Attribute("href", String, "API href of bottle")
		Attribute("name", String, "Name of wine")
	})
})
```
Let's break this down:
* We define a `design` package and use a `init` function to declare the API, we could also have
  used an anonymous variable: `var _ = API(...`.
* The `API` function takes two arguments: the name of the API and an anonymous function that 
  defines additional properties, here a title and a description.
* The `Resource` function also takes a name and an anonymous function. Properties defines in the
  anonymous function includes the actions supported by the resource.
* The `Action` function follows the same pattern of name + anonymous function. Actions are defined
  in resources, they can be CRUD (Create/Read/Update/Delete) actions or so-called "custom" actions.
  Here we define a Read (`show`) action.
* The `Action` function defines the action endpoint, parameters, payload (not used here) and
  responses.
* Finally we define the resource media type as a global variable so we can refer to it when
  declaring the `OK` response. A media type has a name as defined by [RFC 6838](https://tools.ietf.org/html/rfc6838)
  and describes the attributes of the response body (the JSON object fields in goa).

The DSL reference contains more details for each of the functions use in the example above.

Now that we have a design for the API we can use the `goa` tool to generate all the boilerplate for
our app. The goa tool takes the path to the package as argument (the same path you'd use if you
were to import the design package). So for example if you created the design package under
`$GOPATH/src/app`, the command line would be:
```
goa app/design
```
This creates a `autogen` folder containing three files:
* `resources.go` contains the bottle resource data structure definition.
* `contexts.go` contains the context data structure definitions. Contexts play a similar role
  to Martini's `martini.Context`, goji's `web.C` or echo's `echo.Context` to take a few arbitrary
  examples: they are given as argument to controller actions and provide helper methods to
  retrieve the action parameters or write the response.
* `handlers.go` provide the glue between the underlying go http server handler and your controller
  actions. They create the action specific context and call your code.

The next and final step consists of implementing the `bottle` resource `show` action, create a file
`main.go` with the following content:
```go
package main

import "./autogen"
import "github.com/raphael/goa"

func main() {
	c := goa.NewController("bottles")
	c.SetHandlers(goa.Handlers{ "show": ShowBottle })
	app := goa.New("cellar")
	app.Mount(c)
	app.Run(":8080")
} 

func ShowBottle(c *autogen.ShowBottleContext) error {
	bottle := Bottle{ID: c.ID, Name: fmt.Sprintf("Bottle #%d", c.ID)}
	bottle.Href := bottle.ComputeHref()
	return c.OK(&bottle)
}
```
Step by step:
* We first create a new controller for the `bottles` resource.
* We then associate each resource action with the function that implements it using the
  `SetHandlers` controller method. Here we implement the only action of the bottles resource 
  `show` in the `ShowBottle` function.
* Finally we create a goa application, mount the `bottles` controller on it and run it. Mounting
  a controller on an application will cause `goa` to validate the actions (i.e. make sure they
  all have handlers, that handlers have the proper signature etc.). An error at this point stops
  the application to avoid running a mis-configured application.
* The `ShowBottle` handler contains the application specific logic. It leverages the generated
  `ShowBottleContext` data structure which exposes the action `ID` parameter as an `int` so that
  no cast is required. The context also exposes a `OK` function that takes care of creating the 
  HTTP response including any specific header, defining the appropriate status code and properly
  serializing the given bottle resource.

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
curl -i localhost:8080/1
HTTP/1.1 200 OK
Date: Mon, 27 Jul 2015 04:35:22 GMT
Content-Length: 40
Content-Type: text/plain; charset=utf-8

{"ID":1,"Href":"/1","Name":"Bottle #1"}
```
Note how if you pass in an invalid id then `goa` takes care of generating the proper response:
```
curl -i localhost:8080/a
HTTP/1.1 400 Bad Request
Date: Mon, 27 Jul 2015 04:37:09 GMT
Content-Length: 17
Content-Type: text/plain; charset=utf-8

invalid value 'a' for parameter id, must be a int"
```
Congratulations on writing you first goa application!
This was just scratching the surface though - proceed to read the documentation and learn about all
the other things `goa` can do.
