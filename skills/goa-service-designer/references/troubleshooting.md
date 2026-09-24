# Troubleshooting

Use this reference when generation fails, generated code is surprising, implementation no longer
compiles, or transport behavior does not match the design.

## Generation Fails

Read the exact `goa gen` error first, then inspect the design around the named service, method,
field, or error. Common causes:

- Duplicate protobuf field numbers.
- Missing required field definitions.
- Invalid HTTP path parameter mappings.
- Path tokens that do not map to payload fields.
- Mismatched security fields.
- Invalid custom error metadata.
- Use of filesystem paths such as `goa gen ./design` instead of a Go import path.
- A Goa CLI or protobuf plugin version that differs from the selected module's requirements.

Fix the owning design or tooling mismatch and regenerate. Do not edit generated output.

## Implementation No Longer Compiles

Open `gen/<service>/service.go` and implement the new generated interface exactly. Do not change
generated signatures.

Then check:

- Payload/result pointer semantics.
- Generated error constructor names.
- Method names and result types.
- Endpoint or interceptor wrapper signature changes.

## HTTP Decoding Is Wrong

Compare payload fields with the HTTP mapping:

- `GET`, `POST`, `PUT`, `PATCH`, or `DELETE` path tokens.
- `Param` mappings.
- `Header` mappings.
- `Body` mappings.
- Element wire-name mappings such as `Header("version:X-Api-Version")`.

Every path token must map to a payload field. For object payloads, unmapped attributes are encoded in
the body unless `Body(...)` says otherwise.

## gRPC Output Is Wrong

Inspect:

- `Field` numbers.
- Inherited field numbers from every `Extend` ancestor, including collisions at `100`.
- Package and version metadata.
- `GRPC(Message(...))` customizations.
- Metadata mappings.
- Generated protobuf output.
- Generated gRPC transport `types.go`.
- Required scalar presence versus protobuf collection presence.
- Whether a default is applied while decoding absent input or incorrectly expected to replace an
  explicit service zero value.

Fix the design and regenerate. Never patch `.proto` or generated `.pb.go` output directly.

## JSON-RPC Behavior Is Wrong

Check the selected Goa version before applying examples from another release:

- Service-level POST path and method-level `JSONRPC` configuration.
- Ordinary request versus explicit `Notification()`.
- Request ID location and generated parameter representation.
- Designed error code and generated error `data` shape.
- Explicit `ServerSentEvents()` for a streaming result.
- Combinations rejected by Goa v3.31 and later, such as JSON-RPC client streams or `Result` plus
  `StreamingResult`.

Use the generated client/server contract and the selected release's upgrade notes. Do not repair
protocol drift with handwritten envelopes in service code.

## Interceptors Behave In The Wrong Order

Inspect generated `Wrap<Method>Endpoint` or stream wrappers. The wrapper chain is the source of
truth:

- On the request path, the last wrapper runs first.
- On the response path, the first wrapper runs first.
- Streaming methods may invoke interceptor code for unary setup, send, and receive paths.

## Mini Patterns

Good design-level validation:

```go
Payload(func() {
    Field(1, "accountId", AccountID, "Account identifier.")
    Field(2, "limit", Int, "Maximum number of items.", func() {
        Default(50)
        Minimum(1)
        Maximum(100)
    })
    Required("accountId")
})
```

Avoid service-level revalidation of the same contract:

```go
// Bad: duplicates Goa validation and hides design drift.
if payload.AccountID == "" {
    return nil, service.MakeInvalidInput(errors.New("missing account id"))
}
```

Good transport mapping:

```go
HTTP(func() {
    GET("/accounts/{account_id}/items")
    Param("accountId:account_id")
    Param("limit")
    Response(StatusOK)
    Response("not_found", StatusNotFound)
})
```
