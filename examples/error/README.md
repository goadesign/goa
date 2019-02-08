# Divider Service

The `divider` service illustrates error handling in goa v2.

See [basic example](https://github.com/goadesign/goa/tree/v2/examples/calc) for
understanding the fundamentals of goa v2, generating, and building code. 

## Design

Errors are defined in goa v2 using the `Error` DSL. Errors can be defined at
the service-level to define common errors across all the methods in the service
or at the method-level to define errors that method return. By default, goa v2
creates errors of type `ErrorResult` unless specified explicitly.

The `divider` service defines two methods `integer_divide` and `divide`.
Both the methods return the following errors `div_by_zero` (dividing by zero)
and `timeout` (operation timed out). `integer_divide` method defines an
additional error `has_remainder` which is returned when dividing two integers
leaves a remainder.

```
var _ = Service("divider", func() {
  Error("div_by_zero", ErrorResult, "divizion by zero")
  Error("timeout", ErrorResult, "operation timed out, retry later.", func() {
    // Timeout indicates an error due to a timeout.
    Timeout()
    // Temporary indicates that the request may be retried.
    Temporary()
  })
...

  Method("integer_divide", func() {
    Error("has_remainder", ErrorResult, "integer division has remainder")
  ...
  })

  Method("divide", func() {
  ...
  })
})

```

Errors defined at the transport-independent level are mapped to the transport
responses by defining `Response` DSL that takes error name as the first
argument.

```
var _ = Service("divider", func() {
  HTTP(func() {
    // Use HTTP status code 400 Bad Request for "div_by_zero"
    // errors.
    Response("div_by_zero", StatusBadRequest)

    // Use HTTP status code 504 Gateway Timeout for "timeout"
    // errors.
    Response("timeout", StatusGatewayTimeout)
  })

  GRPC(func() {
    // Use gRPC status code "InvalidArgument" for "div_by_zero"
    // errors.
    Response("div_by_zero", CodeInvalidArgument)

    // Use gRPC status code "DeadlineExceeded" for "timeout"
    // errors.
    Response("timeout", CodeDeadlineExceeded)
  })
  

  Method("integer_divide", func() {
    ...
    HTTP(func() {
      ...
      Response("has_remainder", StatusExpectationFailed)
    })
    GRPC(func() {
      ...
      Response("has_remainder", CodeUnknown)
    })
  })

  Method("divide", func() {
    HTTP(func() {
      ...
    })

    GRPC(func() {
      ...
    })
  })
})
```

The generated code creates `Make<ErrorName>` function for every error of
`ErrorResult` type in the service package which must be invoked in the service
implementation to return appropriate errors.

```
func (s *dividerSvc) IntegerDivide(ctx context.Context, p *dividersvc.IntOperands) (int, error) {
  if p.B == 0 {
    return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
  }
  if p.A%p.B != 0 {
    return 0, dividersvc.MakeHasRemainder(fmt.Errorf("remainder is %d", p.A%p.B))
  }
  return p.A / p.B, nil
}
```

