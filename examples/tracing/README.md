# Tracing Example

This example illustrates how to use the tracing and AWS X-Ray middleware in
goa v2.

## Server Tracing Setup

Tracing middleware can be mounted on the `net/http` `Handler` for HTTP
transport or when initializing a gRPC server for gRPC transport. The ordering
of mounting the tracing and X-Ray middleware is important as shown below.

```
  // HTTP

  var handler http.Handler = mux
  {
    xrayHndlr, err := xray.New("calc", daemon)
    if err != nil {
      logger.Printf("[WARN] cannot connect to xray daemon %s: %s", daemon, err)
    }
    // Wrap the Xray and the tracing handler. The order is very important.
    handler = xrayHndlr(handler)
    handler = httpmdlwr.Trace()(handler)
  }

  // gRPC

  xm, err := xray.NewUnaryServer("calc", daemon)
  if err != nil {
    logger.Printf("[WARN] cannot connect to xray daemon %s: %s", daemon, err)
  }
  // Initialize gRPC server with the middleware.
  srv := grpc.NewServer(
    grpcmiddleware.WithUnaryServerChain(
      // Mount the trace and X-Ray middleware. Order is very important.
      grpcmdlwr.UnaryServerTrace(),
      xm,
    ),
  )
```

## Client Tracing Setup

Tracing middleware can be wrapped around a HTTP client for HTTP transport or
when initializing a gRPC client connection for gRPC transport. The order of
mounting the tracing and X-Ray middleware is important as shown below.

```
  // HTTP

  var (
    doer goahttp.Doer
  )
  {
    doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
    // Wrap doer with X-Ray and trace client middleware. Order is very important.
    doer = xray.WrapDoer(doer)
    doer = middleware.WrapDoer(doer)
  }

  // gRPC

  conn, err := grpc.Dial(
    host,
    grpc.WithInsecure(),
    grpc.WithUnaryInterceptor(grpcmiddleware.ChainUnaryClient(
      // Mount the X-Ray and trace client middleware. Order is very important.
      xray.UnaryClient(host),
      middleware.UnaryClientTrace(),
    )),
  )
```
