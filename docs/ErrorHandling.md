# Handling Errors

goa makes it possible to describe the errors that a service method may return.
This allows goa to generate documentation and code that support the encoding of
the errors. Errors have a name, a type which may be a primitive type or a user
defined type and a description that is used to generate comments and
documentation.

This document describes how to define errors in goa designs and how to leverage
the generated code to return errors from service methods.

## Design

The goa DSL makes it possible to define error results on methods and on entire
services using the [Error](https://godoc.org/goa.design/goa/dsl#Error)
expression:

```go
var _ = Service("divider", func() {

        // The "div_by_zero" error is defined at the service level and
        // thus may be returned by both "divide" and "integer_divide".
        Error("div_by_zero")

        Method("integer_divide", func() {

                // The "has_remainder" error is defined at the method
                // level and is thus specific to "integer_divide".
                Error("has_remainder")
                // ...
        })

        Method("divide", func() {
                // ...
        })
})
```

In this example both the `div_by_zero` and `has_remainder` errors use the
default error type `ErrorResult`. This type defines the following fields:

* `Name` is the name of the error. The generated code takes care of initializing
  the field with the name defined in the design during response encoding.
* `ID` is a unique identifier for the specific instance of the error. The idea
  is that this ID may be instrumented making it possible to correlate a user
  error report with service logs, traces etc.
* `Message` is the error message.
* `Temporary` indicates whether the error is temporary.
* `Timeout` indicates whether the error is due to a timeout.

The DSL makes is possible to specify whether an error denotes a temporary
condition and/or a timeout, here are some examples:

```go
        Error("network_failure", func() {
                Temporary()
        })

        Error("timeout"), func() {
                Timeout()
        })

       Error("remote_timeout", func() {
                Temporary()
                Timeout()
        })
```

The generated code takes care of initializing the `ErrorResult` `Temporary` and
`Timeout` fields appropriately when encoding the error response.

### Designing HTTP Responses

The HTTP DSL `Response` expression makes it possible to define the HTTP status
code associated with a given error. Going back to our `divider` service example,
the HTTP transport could be designed as follows:

```go
var _ = Service("divider", func() {
        Error("div_by_zero")
        HTTP(func() {
                // Use HTTP status code 400 Bad Request for "div_by_zero"
                // errors.
                Response("div_by_zero", StatusBadRequest)
        })

        Method("integer_divide", func() {
                Error("has_remainder")
                HTTP(func() {
                        Response("has_remainder", StatusExpectationFailed)
                        // ...
                })
        })
        // ...
})
```

## Returning Errors

Given the divider service design above goa generates a `ErrorResult` data
structure in the `divider` service package. goa also generates two helper
functions that build the corresponding errors: `MakeDivByZero` and
`MakeHasRemainder`. These functions accept a Go error as argument making it
convenient to map a business logic error to a specific error result.

Here is an example of what an implementation of `integer_divide` could look
like:

```go
func (s *dividerSvc) IntegerDivide(ctx context.Context, p *dividersvc.IntOperands) (int, error) {
        if p.B == 0 {
                // Use generated function to create error result
                return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
        }
        if p.A%p.B != 0 {
                return 0, dividersvc.MakeHasRemainder(fmt.Errorf("remainder is %d", p.A%p.B))
        }
        return p.A / p.B, nil
}
```

And that's it! given this goa knows to initialize a `ErrorResult` using the
provided error to initiliaze the message field and initializing all the other
fields from the information provided in the design. The generated transport code
also writes the proper HTTP status code to use for each. 
Using the generated command line tool to verify:

```bash
./dividercli -v divider integer-divide -a 1 -b 2 
> GET http://localhost:8080/idiv/1/2
< 417 Expectation Failed
< Content-Length: 68
< Content-Type: application/json
< Date: Thu, 22 Mar 2018 01:34:33 GMT
{"name":"has_remainder","id":"dlqvenWL","message":"remainder is 1"}
```