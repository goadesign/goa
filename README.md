# <img src="http://goa.design/img/goa-logo.svg">

goa is a framework for building micro-services and APIs in Go using a unique
design-first approach.

---
[![Build Status](https://travis-ci.org/goadesign/goa.svg?branch=v2)](https://travis-ci.org/goadesign/goa)
[![Windows Build status](https://ci.appveyor.com/api/projects/status/vixp37loj5i6qmaf/branch/v2?svg=true)](https://ci.appveyor.com/project/RaphaelSimon/goa-oqtis/branch/master)
[![Sourcegraph](https://sourcegraph.com/github.com/goadesign/goa/-/badge.svg)](https://sourcegraph.com/github.com/goadesign/goa?badge)
[![Godoc](https://godoc.org/goa.design/goa?status.svg)](https://godoc.org/goa.design/goa)
[![Slack](https://img.shields.io/badge/slack-gophers-orange.svg?style=flat)](https://gophers.slack.com/messages/goa/)

## Why goa?

goa takes a different approach to building services by making it possible to
describe the *design* of the service API using a simple Go DSL. goa uses the
description to generate specialized service helper code, client code and
documentation. goa is extensible via plugins, for example the
[goakit](https://github.com/goadesign/plugins/tree/master/goakit) plugin
generates code that leverage the [go-kit](https://github.com/go-kit/kit)
library.

The service design describes the transport independent layer of the services in
the form of simple methods that accept a context and a payload and return a
result and an error. The design also describes how the payloads, results and
errors are serialized in the transport (HTTP or gRPC). For example a service
method payload may be built from an HTTP request by extracting values from the
request path, headers and body. This clean separation of layers makes it
possible to expose the same service using multiple transports. It also promotes
good design where the service business logic concerns are expressed and
implemented separately from the transport logic.

The goa DSL consists of Go functions so that it may be extended easily to avoid
repetion and promote standards. The design code itself can easily be shared
across multiple services by simply importing the corresponding Go package again
promoting reuse and standardization across service boundaries.

## Code Generation

The goa tool accepts the Go design package import path as input and produces the
interface as well as the glue that binds the service and client code with the
underlying transport. The code is specific to the API so that for example there
is no need to cast or "bind" any data structure prior to using the request
payload or response result. The design may define validations in which case the
generated code takes care of validating the incoming request payload prior to
invoking the service method on the server, and validating the response prior to
invoking the client code.

## Installation

Assuming you have a working [Go](https://golang.org) setup:
```
go get -u goa.design/goa/...
```

### Vendoring

Because goa generates and compiles code `dep` is not able to properly identify all the dependencies
when running `dep init` in a goa project. The symptoms manifest themselves when running `goa gen`,
the error is:

```
exit status 1
design must define at least one service
```

Simply add the `goa.design/goa/codegen/generator` as a required package to `Gopkg.toml` to fix the
issue:

```
required = ["goa.design/goa/codegen/generator"]
```

### Stable Versions

goa follows [Semantic Versioning](http://semver.org/) which is a fancy way of saying it publishes
releases with version numbers of the form `vX.Y.Z` and makes sure that your code can upgrade to new
versions with the same `X` component without having to make changes.

Releases are tagged with the corresponding version number. There is also a branch for each major
version (`v1` and `v2`). The recommended practice is to vendor the stable branch.

Current Release: `v2.0.0`
Stable Branch: `v2`

## Teaser

### 1. Design

Create the file `$GOPATH/src/calcsvc/design/design.go` with the following content:
```go
package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"

// API describes the global properties of the API server.
var _ = API("calc", func() {
	Title("Calculator Service")
	Description("HTTP service for adding numbers, a goa teaser")
})

// Service describes a service
var _ = Service("calc", func() {
	Description("The calc service performs operations on numbers")
	// Method describes a service method (endpoint)
	Method("add", func() {
		// Payload describes the method payload
		// Here the payload is an object that consists of two fields
		Payload(func() {
			// Attribute describes an object field
			Attribute("a", Int, "Left operand")
      Attribute("b", Int, "Right operand")
      // Both attributes must be provided when invoking "add"
      Required("a", "b")
		})
		// Result describes the method result
		// Here the result is a simple integer value
		Result(Int)
		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			GET("/add/{a}/{b}")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})
})
```
This file contains the design for a `calc` service which accepts HTTP GET
requests to `/add/{a}/{b}` where `{a}` and `{b}` are placeholders for integer
values. The API returns the sum of `a` and `b` in the HTTP response body.

### 2. Implement

Now that the design is done, let's run `goa` on the design package:
```
cd $GOPATH/src/calcsvc
goa gen calcsvc/design
```
This produces a `gen` directory with the following directory structure:
```
gen
├── calc
│   ├── client.go
│   ├── endpoints.go
│   └── service.go
└── http
    ├── calc
    │   ├── client
    │   │   ├── client.go
    │   │   ├── cli.go
    │   │   ├── encode_decode.go
    │   │   ├── paths.go
    │   │   └── types.go
    │   └── server
    │       ├── encode_decode.go
    │       ├── paths.go
    │       ├── server.go
    │       └── types.go
    ├── cli
    │   └── cli.go
    └── openapi.json

6 directories, 14 files
```
* `calc` contains the service endpoints and interface as well as a service
  client.
* `http` contains the HTTP transport layer. This layer maps the service
  endpoints to HTTP handlers server side and HTTP client methods client side.
  The `http` directory also contains a complete OpenAPI 2.0 spec of the service.

The `goa` tool also exposes a `example` command which generates an example
implementation and provides a good starting point. Let's run it:
```
goa example calcsvc/design
calc.go
cmd/calccli/main.go
cmd/calcsvc/main.go
```
The tool generated the `main` functions for two tools: one that runs the server
and one the client. The tool also generated a dummy service implementation that
prints a log message. Again note that the `example` command is intended to
generate just that: an *example*, in particular it is not intended to be re-run
each time the design changes.

Let's implement our service by providing a proper implementation for the `add`
method. goa generated a payload struct for the `add` method that contains both
fields. goa also generated the transport layer that takes care of decoding the
request so all we have to do is to perform the actual sum. Edit the file
`calc.go` and change the code of the `add` function as follows:

```go
// Add returns the sum of attributes a and b of p.
func (s *calcsvcSvc) Add(ctx context.Context, p *calcsvc.AddPayload) (int, error) {
	return p.A + p.B, nil
}
```

That's it! we have now a full-fledged HTTP service with a corresponding OpenAPI
specification and a client tool.

### 3. Run

Now let's compile and run the service:

```
cd $GOPATH/src/calcsvc/cmd/calcsvc
go build
./calcsvc
[calc] 04:27:45 [INFO] service "calc" method "Add" mounted on GET /add/{a}/{b}
[calc] 04:27:45 [INFO] listening on :8080
```

Open a new console and compile the generated CLI tool:

```
cd $GOPATH/src/calcsvc/cmd/calccli
go build
```

and run it:

```
./calccli -a 1 -b 2
3
```

The tool includes contextual help:
```
./calccli --help
./calccli is a command line client for the calc API.

Usage:
    ./calccli [-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -url URL:    specify service URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
    calc add
    
Additional help:
    ./calccli SERVICE [ENDPOINT] --help

Example:
    ./calccli calc add --a 5952269320165453119 --b 1828520165265779840
```

Help is also available on each command:

```
./calccli calc add --help
./calccli [flags] calc add -a INT -b INT

Add implements add.
    -a INT: Left operand
    -b INT: Right operand

Example:
    ./calccli calc add --a 5952269320165453119 --b 1828520165265779840
```

Now let's see how robust our code is and try to use non integer values:

```
./calccli calc add -a 1 -b foo
invalid value for b, must be INT
run './calccli --help' for detailed usage.
```

As you can see the generated tool validated the command line arguments against
the types defined in the design. The server also validates the types when
decoding incoming requests so that your code only has to deal with the business
logic.

### 4. Document

The `http` directory contains the OpenAPI 2.0 specification in both YAML and
JSON format.

The specification can easily be served from the service itself using a file
server. The [Files](http://godoc.org/goa.design/goa/http/dsl/http.go#Files) DSL
function makes it possible to server static file. Edit the file
`design/design.go` and add:

```go
var _ = Service("openapi", func() {
  // Serve the file with relative path ../../gen/http/openapi.json for requests
  // sent to /swagger.json.
  Files("/swagger.json", "../../gen/http/openapi.json")
})
```

Re-run `goa gen calcsvc/design` and note the new directory `gen/openapi`
containing the implementation for a HTTP handler that serves the `openapi.json`
file.

All we need to do is mount the handler on the service mux. Add the corresponding
import statement:

```go
import openapisvr "calcsvc/gen/http/openapi/server"
```

and mount the handler by adding the following line in `cmd/calcsvc/main.go`
after the mux creation (e.g. one the line after the `// Configure the mux.`
comment):

```go
openapisvr.Mount(mux)
```

That's it, we now have a self-documenting service! Stop the running service
with CTRL-C. Rebuild and re-run it then make requests to the newly added
`/swagger.json` endpoint:

```
^C[calc] 05:04:28 exiting (interrupt)
[calc] 05:04:28 exited
go build
./calcsvc
```

In a different console:

```
curl localhost:8080/swagger.json
{"swagger":"2.0","info":{"title":"Calculator Service","description":...
```

## Resources

Consult the following resources to learn more about goa.

### goa.design

[goa.design](https://goa.design) contains further information on goa including a getting
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
