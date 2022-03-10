![Goa logo](https://goa.design/img/goa-banner.png "Goa")

## Overview

[![Build Status](https://github.com/goadesign/goa/workflows/build/badge.svg?branch=v3&event=push)](https://github.com/goadesign/goa/actions?query=branch%3Av3+event%3Apush)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/goa.design/goa/v3@v3.6.2/dsl?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/goadesign/goa)](https://goreportcard.com/report/goadesign/goa)
[![Slack](https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=red)](https://gophers.slack.com/messages/goa)
[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/goadesign.svg?style=social&label=Follow%20%40goadesign)](https://twitter.com/goadesign)

Goa takes a different approach to building services by making it possible to
describe the *design* of the service API using a simple Go DSL. Goa uses the
description to generate specialized service helper code, client code and
documentation. Goa is extensible via plugins, for example the
[goakit](https://github.com/goadesign/plugins/tree/v3/goakit) plugin
generates code that leverage the Go kit library.

The service design describes the transport independent layer of the services in
the form of simple methods that accept a context and a payload and return a
result and an error. The design also describes how the payloads, results and
errors are serialized in the transport (HTTP or gRPC). For example a service
method payload may be built from an HTTP request by extracting values from the
request path, headers and body. This clean separation of layers makes it
possible to expose the same service using multiple transports. It also promotes
good design where the service business logic concerns are expressed and
implemented separately from the transport logic.

The Goa DSL consists of Go functions so that it may be extended easily to avoid
repetition and promote standards. The design code itself can easily be shared
across multiple services by simply importing the corresponding Go package again
promoting reuse and standardization across services.

## Code Generation

The Goa tool accepts the Go design package import path as input and produces the
interface as well as the glue that binds the service and client code with the
underlying transport. The code is specific to the API so that for example there
is no need to cast or "bind" any data structure prior to using the request
payload or response result. The design may define validations in which case the
generated code takes care of validating the incoming request payload prior to
invoking the service method on the server, and validating the response prior to
invoking the client code.

## Installation

```bash
go install goa.design/goa/v3/cmd/goa@v3
```

>Note: Goa requires the use of Go modules.

Current Release: `v3.6.2`

## Teaser

### 1. Design

Create a new Goa project:

```bash
mkdir -p calcsvc/design
cd calcsvc
go mod init calcsvc
```

Create the file `design.go` in the `design` directory with the following
content:

```go
package design

import . "goa.design/goa/v3/dsl"

// API describes the global properties of the API server.
var _ = API("calc", func() {
        Title("Calculator Service")
        Description("HTTP service for multiplying numbers, a goa teaser")
        Server("calc", func() {
                Host("localhost", func() { URI("http://localhost:8088") })
        })
})

// Service describes a service
var _ = Service("calc", func() {
        Description("The calc service performs operations on numbers")
        // Method describes a service method (endpoint)
        Method("multiply", func() {
                // Payload describes the method payload
                // Here the payload is an object that consists of two fields
                Payload(func() {
                        // Attribute describes an object field
                        Attribute("a", Int, "Left operand")
                        Attribute("b", Int, "Right operand")
                        // Both attributes must be provided when invoking "multiply"
                        Required("a", "b")
                })
                // Result describes the method result
                // Here the result is a simple integer value
                Result(Int)
                // HTTP describes the HTTP transport mapping
                HTTP(func() {
                        // Requests to the service consist of HTTP GET requests
                        // The payload fields are encoded as path parameters
                        GET("/multiply/{a}/{b}")
                        // Responses use a "200 OK" HTTP status
                        // The result is encoded in the response body
                        Response(StatusOK)
                })
        })
})
```

This file contains the design for a `calc` service which accepts HTTP GET
requests to `/multiply/{a}/{b}` where `{a}` and `{b}` are placeholders for integer
values. The API returns the product of `a` multiplied by `b` in the HTTP response body.

### 2. Implement

Now that the design is done, let's run `goa` on the design package.
In the `calcsvc` directory run:

``` bash
goa gen calcsvc/design
```

This produces a `gen` directory with the following directory structure:

``` text
gen
├── calc
│   ├── client.go
│   ├── endpoints.go
│   └── service.go
└── http
    ├── calc
    │   ├── client
    │   │   ├── cli.go
    │   │   ├── client.go
    │   │   ├── encode_decode.go
    │   │   ├── paths.go
    │   │   └── types.go
    │   └── server
    │       ├── encode_decode.go
    │       ├── paths.go
    │       ├── server.go
    │       └── types.go
    ├── cli
    │   └── calc
    │       └── cli.go
    ├── openapi.json
    └── openapi.yaml

7 directories, 15 files
```

* `calc` contains the service endpoints and interface as well as a service
  client.
* `http` contains the HTTP transport layer. This layer maps the service
  endpoints to HTTP handlers server side and HTTP client methods client side.
  The `http` directory also contains a complete
  [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md)
  spec for the service.

The `goa` tool can also generate example implementations for both the service
and client. These examples provide a good starting point:

``` text
goa example calcsvc/design

calc.go
cmd/calc-cli/http.go
cmd/calc-cli/main.go
cmd/calc/http.go
cmd/calc/main.go
```

The tool generated the `main` functions for two commands: one that runs the
server and one the client. The tool also generated a dummy service
implementation that prints a log message. Again note that the `example` command
is intended to generate just that: an *example*, in particular it is not
intended to be re-run each time the design changes (as opposed to the `gen`
command which should be re-run each time the design changes).

Let's implement our service by providing a proper implementation for the `multiply`
method. Goa generated a payload struct for the `multiply` method that contains both
fields. Goa also generated the transport layer that takes care of decoding the
request so all we have to do is to perform the actual multiplication. Edit the file
`calc.go` and change the code of the `multiply` function as follows:

```go
// Multiply returns the multiplied value of attributes a and b of p.
func (s *calcsrvc) Multiply(ctx context.Context, p *calc.MultiplyPayload) (res int, err error) {
        return p.A * p.B, nil
}
```

That's it! we have now a full-fledged HTTP service with a corresponding OpenAPI
specification and a client tool.

### 3. Run

Now let's compile and run the service:

```bash
cd cmd/calc
go build
./calc
[calcapi] 16:10:47 HTTP "Multiply" mounted on GET /multiply/{a}/{b}
[calcapi] 16:10:47 HTTP server listening on "localhost:8088"
```

Open a new console and compile the generated CLI tool:

```bash
cd calcsvc/cmd/calc-cli
go build
```

and run it:

```bash
./calc-cli calc multiply -a 1 -b 2
3
```

The tool includes contextual help:

``` bash
./calc-cli --help
```

Help is also available on each command:

``` bash
./calc-cli calc multiply --help
```

Now let's see how robust our code is and try to use non integer values:

``` bash
./calc-cli calc multiply -a 1 -b foo
invalid value for b, must be INT
run './calccli --help' for detailed usage.
```

The generated code validates the command line arguments against the types
defined in the design. The server also validates the types when decoding
incoming requests so that your code only has to deal with the business logic.

### 4. Document

The `http` directory contains OpenAPI 2.0 and 3.0 specifications in both YAML
and JSON format.

The specification can easily be served from the service itself using a file
server. The [Files](http://godoc.org/goa.design/goa/dsl/http.go#Files) DSL
function makes it possible to serve a static file. Edit the file
`design/design.go` and add:

```go
var _ = Service("openapi", func() {
	// Serve the file gen/http/openapi3.json for requests sent to
	// /openapi.json. The HTTP file system is created below.
	Files("/openapi.json", "openapi3.json")
})
```

Re-run `goa gen calcsvc/design` and note the new directory `gen/openapi` and
`gen/http/openapi` which contain the implementation for a HTTP handler that
serves the `openapi.json` file.

All we need to do is mount the handler on the service mux. Add the corresponding
import statement to `cmd/calc/http.go`:

```go
import openapisvr "calcsvc/gen/http/openapi/server"
```

and mount the handler by adding the following line in the same file and after
the mux creation (e.g. one the line after the `// Configure the mux.` comment):

```go
svr := openapisvr.New(nil, mux, dec, enc, nil, nil, http.Dir("../../gen/http"))
openapisvr.Mount(mux, svr)
```

That's it! we now have a self-documenting service. Stop the running service
with CTRL-C. Rebuild and re-run it then make requests to the newly added
`/openapi.json` endpoint:

``` bash
^C[calcapi] 16:17:37 exiting (interrupt)
[calcapi] 16:17:37 shutting down HTTP server at "localhost:8088"
[calcapi] 16:17:37 exited
go build
./calc
```

In a different console:

``` bash
curl localhost:8088/openapi.json
{"openapi":"3.0.3","info":{"title":"Calculator Service","description":...
```

## Resources

### Docs

The [goa.design](https://goa.design) website provides a high level overview of
Goa and the DSL.

In particular the page
[Implementing a Goa Service](https://goa.design/implement/implementing/)
explains how to leverage the generated code to implement an HTTP or gRPC
service.

The [![DSL GoDoc](https://img.shields.io/badge/godoc-DSL-blue)](https://pkg.go.dev/goa.design/goa/v3@v3.6.2/dsl?tab=doc)
contains a fully documented reference of all the DSL functions.

### Instrumentation and System Example

The [clue](https://github.com/goadesign/clue) project provides observability
packages that work in tandem with Goa. The packages cover
[logging](https://github.com/goadesign/clue/tree/main/log),
[tracing](https://github.com/goadesign/clue/tree/main/trace),
[metrics](https://github.com/goadesign/clue/tree/main/metrics),
[health checks](https://github.com/goadesign/clue/tree/main/health)
and service client
[mocking](https://github.com/goadesign/clue/tree/main/mock). clue also includes a fully featured
[example](https://github.com/goadesign/clue/tree/main/example/weather)
consisting of three instrumented Goa microservices that communicate with each other.

### Getting Started Guides

A couple of Getting Started guides produced by the community.

Joseph Ocol from Pelmorex Corp. goes through a complete example writing a server
and client service using both HTTP and gRPC transports.

[![GOA Design Tutorial](https://tech.pelmorex.com/wp-content/uploads/2020/07/GOA-Design-Tutorial-Screencap-800x470.png)](https://vimeo.com/437928805)

Gleidson Nascimento goes through how to create a complete service that using both
[CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) and
[JWT](https://jwt.io/) based authentication to secure access.

[![API Development in Go Using Goa](https://bs-uploads.toptal.io/blackfish-uploads/uploaded_file/file/275966/image-1592349920607-734c25f64461bf3c482bac1d73c26432.png)](https://www.toptal.com/go/goa-api-development)

### Examples

The [examples](https://github.com/goadesign/examples) directory
contains simple examples illustrating basic concepts.

### Troubleshooting

Q: I'm seeing an error that says:

> generated code expected `goa.design/goa/v3/codegen/generator` to be present in the vendor directory, see documentation for more details

How do I fix this?

A: If you are vendoring your dependencies Goa will not attempt to satisfy its
dependencies by retrieving them with `go get`. If you see the above error message, it
means that the `goa.design/goa/v3/codegen/generator` package is not included in your
vendor directory.

To fix, ensure that `goa.design/goa/v3/codegen/generator` is being imported somewhere in your project. This can be as a bare import (e.g. `import _ "goa.design/goa/v3/codegen/generator"`)
in any file or you can use a dedicated `tools.go` file (see [Manage Go tools via Go modules](https://marcofranssen.nl/manage-go-tools-via-go-modules) and [golang/go/issues/25922](https://github.com/golang/go/issues/25922) for more details.) Finally, run `go mod vendor` to ensure
the imported packages are properly vendored.

## Contributing

See [CONTRIBUTING](https://github.com/goadesign/goa/blob/v3/CONTRIBUTING.md).
