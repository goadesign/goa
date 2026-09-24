# Transport, Errors, And Interceptors

Use this reference when changing HTTP, gRPC, or JSON-RPC mappings, streaming, errors, security,
interceptors, middleware, or production wiring. The gRPC presence and JSON-RPC behavior below
describes Goa v3.31 and later. Check the application's selected module version and its
`UPGRADING.md` before changing older code.

## HTTP Mapping

- Use resource nouns, plural names, stable hierarchy, and HTTP methods for actions.
- Put stable prefixes in API-level or service-level `HTTP(Path(...))`.
- Use `Parent` and `CanonicalMethod` only when nested resources become clearer.
- Map path, query, header, and body fields deliberately with `Param`, `Header`, and `Body`.
- Use explicit wire-name mapping when attribute names differ from HTTP element names.
- Choose status codes intentionally. Use `StatusCreated` for creates when appropriate.
- Use default content negotiation unless a real media type requires custom encoders or decoders.
- Use the CORS plugin for browser-facing cross-origin policy.
- Use `Files` only for HTTP static content.

HTTP object payload checks:

- Every path token must map to a payload field.
- Every non-body attribute should land where intended.
- Use `Body("field")` only when the request body should be that field's raw value.
- Unmapped object attributes are encoded in the body.

```go
HTTP(func() {
    GET("/accounts/{account_id}/items")
    Param("accountId:account_id")
    Param("limit")
    Response(StatusOK)
    Response("not_found", StatusNotFound)
})
```

## Streaming

Goa's streaming DSL is transport-agnostic. The transport mapping determines HTTP behavior:

- Plain HTTP streaming defaults to WebSocket.
- Add `ServerSentEvents()` only for one-way server-to-client streams.
- WebSocket endpoints use `GET`.
- For gRPC streaming, check generated protobuf and server/client stream types.
- JSON-RPC supports unary calls or a single request followed by an
  explicit SSE stream. It rejects client or bidirectional streaming, WebSocket streams, and a
  method combining `Result` with `StreamingResult`. Ordinary HTTP WebSocket support is separate.
- Streaming implementations must handle flow control, `io.EOF`, send errors, context cancellation,
  timeouts, and cleanup.

## gRPC Mapping

- Set package and version metadata for public or versioned APIs.
- Keep domain data in messages.
- Use metadata, headers, and trailers only for protocol metadata.
- Use streaming for large or continuous datasets.
- Never renumber released `Field` values.
- Check `.proto` output after changing shared types, streaming methods, or custom protobuf metadata.
- Required singular scalars preserve protobuf presence. An omitted required
  boolean is invalid; an explicitly supplied `false` can be valid. Generated decoders validate
  before converting to service values. See the modeling reference for collections and defaults.

## JSON-RPC Mapping

- Use service-level `JSONRPC` to configure the shared HTTP POST endpoint, and method-level
  `JSONRPC` for method-specific behavior and error mappings.
- Let generated code own the JSON-RPC envelope, dispatch, request IDs, batches, and protocol errors.
  The payload describes `params`; do not handwrite transport envelopes in service code.
- Ordinary generated client calls have an ID even without an `ID` payload field.
  `Notification()` explicitly declares a one-way call. Having no result does not by itself make
  a method a notification. Notifications cannot declare a result, ID field, or stream.
- `ID` belongs directly in the payload, not the result. The transport returns the request ID;
  the service does not choose a response ID.
- Map designed errors with JSON-RPC `Response` codes. Verify the generated error `data` shape
  before changing a custom client's decoding.
- For SSE, pair `StreamingResult` with method-level `ServerSentEvents()`. Verify event field
  mappings and cancellation through the generated client and server.

## Errors

- Define reusable potential errors at API scope when you want one canonical error type/name and one
  transport mapping. API-level `Error(...)` declarations do not make those errors applicable to
  every endpoint. Select a reusable definition with `Error("name")` at service or method scope.
- Define an error at service scope only when every method on that service can return it.
- Define an error at method scope when only that method can return it.
- Do not promote a method-specific error to service scope just to avoid repeating a declaration or to
  make generated constructors easier to reach; the design would falsely advertise that all service
  methods can return it.
- Use descriptive names and descriptions.
- Prefer `ErrorResult` for most errors.
- Add `Temporary()`, `Timeout()`, or `Fault()` when clients should behave differently.
- Use custom error types only when clients need structured context.
- If multiple custom errors can return from the same method, include a field marked with
  `Meta("struct:error:name")`.
- Map every exposed error for each enabled transport with `HTTP(Response(...))` and
  `GRPC(Response(...))`, or the corresponding `JSONRPC(Response(...))`.
- In implementation code, return generated error constructors or generated custom error payloads.
- Wrap underlying causes for logs and tracing, but keep client-facing messages safe.
- Test errors through the service API. Verify generated error names, transport status mappings when
  practical, and custom error payloads.

## Security

- Declare security schemes, required scopes, and credential fields in the design.
- Prefer API-level defaults.
- Override at service or method scope when needed.
- Use `NoSecurity()` explicitly for public methods.
- Add matching security payload fields such as `TokenField` or `APIKeyField`.
- Map credentials through HTTP headers/query parameters or gRPC metadata.
- Implement the generated authentication callbacks and application authorization rules.
  A scheme declaration does not verify a token or establish resource ownership by itself.

## Goa Interceptors Vs Middleware

Use Goa interceptors for type-safe domain concerns:

- Business validation that is not boundary validation.
- Request enrichment.
- Response enrichment.
- Auditing.
- Domain-level transformations.

Use HTTP middleware or gRPC interceptors for protocol concerns:

- Logging.
- Tracing.
- CORS.
- Compression.
- Request IDs.
- Rate limiting.
- Panic recovery.
- Wire-level metadata.

Goa interceptors are generated endpoint wrappers, not transport middleware:

- Server interceptors run after transport decoding and before the service method.
- Client interceptors run around the typed client endpoint before transport encoding and after
  transport decoding.
- Use generated typed accessors.
- Call `next` exactly once when continuing.
- If order matters, inspect generated `Wrap<Method>Endpoint`; the last wrapper runs first on the
  request path.

## Observability

For production services, wire observability, health checks, timeouts, graceful shutdown, and
configuration at application boundaries. Goa docs recommend Clue/OpenTelemetry. Use generated
`ServiceName`, `APIVersion`, method names, endpoint wrappers, and interceptor metadata for stable
telemetry labels.
