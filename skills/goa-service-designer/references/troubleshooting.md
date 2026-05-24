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

Fix the design and regenerate. Do not edit generated output.

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
- Package and version metadata.
- `GRPC(Message(...))` customizations.
- Metadata mappings.
- Generated protobuf output.
- Generated gRPC transport `types.go`.

Fix the design and regenerate. Never patch `.proto` or generated `.pb.go` output directly.

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
    Field(1, "AccountID", AccountID, "Account identifier.")
    Field(2, "Limit", Int, "Maximum number of items.", func() {
        Default(50)
        Minimum(1)
        Maximum(100)
    })
    Required("AccountID")
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
    Param("limit")
    Response(StatusOK)
    Response("not_found", StatusNotFound)
})
```
