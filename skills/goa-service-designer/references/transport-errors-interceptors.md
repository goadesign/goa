# Transport, Errors, And Interceptors

Use this reference when changing HTTP or gRPC mappings, streaming, errors, security, interceptors,
middleware, or production wiring.

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
- Streaming implementations must handle flow control, `io.EOF`, send errors, context cancellation,
  timeouts, and cleanup.

## gRPC Mapping

- Set package and version metadata for public or versioned APIs.
- Keep domain data in messages.
- Use metadata, headers, and trailers only for protocol metadata.
- Use streaming for large or continuous datasets.
- Never renumber released `Field` values.
- Check `.proto` output after changing shared types, streaming methods, or custom protobuf metadata.

## Errors

- Define common errors at API or service scope and operation-specific errors at method scope.
- Use descriptive names and descriptions.
- Prefer `ErrorResult` for most errors.
- Add `Temporary()`, `Timeout()`, or `Fault()` when clients should behave differently.
- Use custom error types only when clients need structured context.
- If multiple custom errors can return from the same method, include a field marked with
  `Meta("struct:error:name")`.
- Map every exposed error for each enabled transport with `HTTP(Response(...))` and
  `GRPC(Response(...))`.
- In implementation code, return generated error constructors or generated custom error payloads.
- Wrap underlying causes for logs and tracing, but keep client-facing messages safe.
- Test errors through the service API. Verify generated error names, transport status mappings when
  practical, and custom error payloads.

## Security

- Define authentication and authorization in the design.
- Prefer API-level defaults.
- Override at service or method scope when needed.
- Use `NoSecurity()` explicitly for public methods.
- Add matching security payload fields such as `TokenField` or `APIKeyField`.
- Map credentials through HTTP headers/query parameters or gRPC metadata.

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
