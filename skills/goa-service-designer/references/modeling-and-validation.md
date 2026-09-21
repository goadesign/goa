# Modeling And Validation

Use this reference when changing Goa types, payloads, results, validation, views, shared types, or
compatibility-sensitive schemas.

The gRPC presence/default and JSON-RPC view behavior below describes Goa v3.31 and later.
Check the application's selected module version and its `UPGRADING.md` before changing older code.

## Type Design

- Prefer focused `Type` definitions with field descriptions.
- Use `Field(number, name, type, description)` for gRPC or multi-transport APIs.
- `Attribute` is sufficient for HTTP or JSON-RPC designs without gRPC. Use numbered `Field`
  declarations for attributes that become protobuf message fields.
- Keep protobuf field numbers stable after release. Prefer numbers 1-15 for frequent fields.
- Check inherited fields before assigning new numbers with `Extend`. Gaps are valid; do not
  renumber fields or restart an inherited range to satisfy an authoring convention.
- Avoid `Any` unless arbitrary JSON is truly the public API contract.
- Use `Reference`, `Extend`, views, shared generated packages, and `Meta(...)` only when they make
  the public contract clearer. Avoid broad shared types, deep inheritance, OpenAPI overrides, or
  protobuf overrides that only hide duplication.

## Primitive Aliases

For reusable constraints on primitive values, define a named primitive alias with
`Type("Name", Primitive, func() { ... })`, then use that type in payloads and results. This keeps
validation, examples, and domain meaning in one place instead of repeating rules on every field.

```go
var AccountID = Type("AccountID", String, func() {
    Description("Stable account identifier.")
    Format(FormatUUID)
    Example("2551dfde-513e-4840-b1be-9bb78d5930e9")
})

Payload(func() {
    Field(1, "accountId", AccountID, "Account to query.")
    Required("accountId")
})
```

Use primitive aliases for IDs, constrained codes, enum-like strings, timestamps, and other primitive
domains with one canonical validation rule. If the type must be shared across services or generated
into a shared package, use the project's established shared-type helper or package metadata rather
than duplicating aliases in each service design.

## Boundary Validation

- Mark required fields with `Required`.
- Use built-in formats such as `FormatUUID`, `FormatEmail`, `FormatDateTime`, and `FormatURI` when
  they match the domain.
- Use `Enum`, `Pattern`, `Minimum`, `Maximum`, `MinLength`, and `MaxLength` in the design.
- Do not duplicate checks already performed by Goa's transport decoder in service code.
  A direct Go call to a service or generated endpoint does not run transport validation. Validate
  external input where it enters that calling path, then pass valid typed values inward.

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

## Presence And Collections

Make presence semantics explicit. Do not use nil versus empty Go slices or maps to encode domain
operations such as "leave unchanged" versus "replace everything."

- For JSON, `Required("items")` requires a present, non-null property. `[]` remains valid unless
  `MinLength(1)` or another constraint forbids it. Required maps likewise accept `{}` without a
  minimum length. Do not remove `Required` merely because empty collections are allowed.
- Protobuf repeated and map fields cannot distinguish absent from empty after a wire round trip.
  Their generated validation checks length and contents, not collection presence. Singular
  message, scalar, and oneof presence checks remain independent.
- If at least one item is required, express that with `MinLength(1)`. Item constraints and
  `ArrayOfRequired` address the contents, not whether the collection itself is empty.
- If the application needs an explicit operation, model it with separate methods or a typed
  operation rather than depending on the in-memory shape of a slice or map.

## Generated Type Semantics

Read the generated type for the layer being changed. A service payload, incoming HTTP body, and
protobuf message can represent the same field differently.

- Service scalar fields are usually values when required or defaulted, and pointers when optional
  without a default. Objects are generally pointers. Bytes, arrays, maps, `Any`, and generated
  unions have their own representations; "every optional field is a pointer" is not a safe rule.
- Incoming HTTP body fields may be pointers even when required, so decoding can detect omission.
  Required protobuf scalars also preserve presence; most use pointers,
  while bytes remain slices and messages remain pointers.
- For gRPC input, defaults apply when an incoming value is absent. Explicit zero, `false`,
  or empty values remain explicit. Service-to-protobuf conversion preserves the supplied service
  values and does not replace zero values with defaults.
- Do not generalize a default or pointer rule across transports, directions, or releases.
  Inspect the generated constructor, validator, and decoder when presence affects behavior;
  use the selected version's upgrade notes for changes.

Use generated union constructors, setters, accessors, and `Kind` when available. Do not infer or
write private branch storage. A union selects one branch; generated validation enforces the
branch's contract.

## Views

If result views are involved:

- A result view selects which attributes belong to that representation. Inspect the generated
  views package and selected-view constructor and validator.
- Only fields included in the selected view participate in its contract. Do not treat a missing
  required field as valid merely because another view omits it.
- Dynamic HTTP views use the `Goa-View` response header; gRPC uses `goa-view` metadata.
  JSON-RPC uses a `{ "view": ..., "body": ... }` value inside `result` when the view is
  dynamic. Fixed views do not need that JSON-RPC envelope.
- Keep custom clients and servers aligned when upgrading view representations.

## Compatibility Decision

Before preserving old behavior, classify it:

- Shipped public API, persisted data, external SDK, or documented behavior: preserve compatibility or
  plan a migration/deprecation.
- Internal-only service contract: update callers in the same change instead of adding shims.
- Unshipped branch work or temporary generated output: replace it cleanly and delete obsolete paths.

When compatibility is required, express it in the design and tests. Do not hide it in service
fallbacks that the generated contract does not describe.
