# Modeling And Validation

Use this reference when changing Goa types, payloads, results, validation, views, shared types, or
compatibility-sensitive schemas.

## Type Design

- Prefer focused `Type` definitions with field descriptions.
- Use `Field(number, name, type, description)` for gRPC or multi-transport APIs.
- Use `Attribute` only for HTTP-only designs.
- Keep protobuf field numbers stable after release. Prefer numbers 1-15 for frequent fields.
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
- Do not duplicate Goa validation in service code.

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

Make presence semantics explicit. Do not use nil versus empty slices or maps to encode meaning:
empty arrays are indistinguishable from nil after marshaling/unmarshalling.

- If empty is valid, make that clear in the field description and do not infer special meaning from
  slice shape.
- If presence matters, add a separate explicit field such as `replace_items`, `items_present`, or a
  patch object that expresses the operation.
- Required arrays should not be empty unless the design explicitly permits that shape.
- Do not rely on nil versus empty maps for meaning either.

## Generated Type Semantics

- Required or defaulted payload/result fields are direct values.
- Optional fields are pointers.
- Objects are pointers.
- Arrays and maps are not pointers.
- Defaults apply to missing optional fields during unmarshalling, not to missing required fields.

## Views

If result views are involved:

- Goa emits a service-level views package.
- Nil viewed attributes are omitted.
- The selected view is carried through the `Goa-View` header.
- View-specific validation is generated automatically.

## Compatibility Decision

Before preserving old behavior, classify it:

- Shipped public API, persisted data, external SDK, or documented behavior: preserve compatibility or
  plan a migration/deprecation.
- Internal-only service contract: update callers in the same change instead of adding shims.
- Unshipped branch work or temporary generated output: replace it cleanly and delete obsolete paths.

When compatibility is required, express it in the design and tests. Do not hide it in service
fallbacks that the generated contract does not describe.
