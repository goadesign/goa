---
name: goa-service-designer
description: Design and evolve Goa services using design-first workflows. Use when creating or modifying Goa DSL, methods, payloads, results, validations, HTTP/gRPC mappings, errors, security, interceptors, generated code, or service implementations in applications that use Goa.
---

# Goa Service Designer

## First, Classify The Task

Before editing, decide what kind of Goa work this is and read only the docs that match the task.
Prefer the project's checked-in Goa docs when the current workspace contains them; otherwise use
the official docs:

- New project or first service: read [Quickstart](https://goa.design/docs/1-goa/quickstart/) and [Code Generation](https://goa.design/docs/1-goa/code-generation/).
- New method, payload, result, validation, views, streaming shape, or shared type: read [DSL Reference](https://goa.design/docs/1-goa/dsl-reference/).
- HTTP path, query, headers, body, CORS, content negotiation, static files, WebSocket, or SSE: read [HTTP Guide](https://goa.design/docs/1-goa/http-guide/).
- gRPC, protobuf field numbers, metadata, trailers, or streaming: read [gRPC Guide](https://goa.design/docs/1-goa/grpc-guide/).
- Error names, error payloads, status-code mapping, or generated error constructors: read [Error Handling](https://goa.design/docs/1-goa/error-handling/).
- Goa interceptors, HTTP middleware, gRPC interceptors, or ordering: read [Interceptors](https://goa.design/docs/1-goa/interceptors/).
- Authentication, production wiring, observability, health checks, shutdown, or timeouts: read [Production](https://goa.design/docs/1-goa/production/).

If `goa gen` fails or generated code is surprising, read
[Code Generation](https://goa.design/docs/1-goa/code-generation/) and inspect the generated
`gen/<service>/service.go`, `endpoints.go`, transport server/client files, and any generated
interceptor wrappers.

## Non-Negotiable Goa Rules

Treat the Goa design as the canonical contract. If payload shape, result shape, validation,
security, errors, transport mapping, documentation, OpenAPI, protobuf, generated clients, or CLIs
need to change, edit the design first and regenerate. Never patch generated files, OpenAPI output,
protobuf output, or generated clients to hide a stale design.

Read the local code before choosing a pattern. Find the existing design package, generated service
interface, service implementation, command wiring, tests, generation command, lint command, and any
user-facing docs for the feature area. Match the project's layout and naming unless the request is
explicitly to redesign it.

Put boundary validation in Goa. Use `Required`, `Enum`, `Format`, `Pattern`, `Minimum`, `Maximum`,
`MinLength`, `MaxLength`, defaults, security fields, and explicit transport mappings in the design.
In service code, trust decoded payloads and enforce business invariants. Do not add nil guards,
fallback behavior, silent recovery, blanket string trimming, or compatibility shims for values the
design guarantees.

Keep the contract small. Add fields, types, errors, interceptors, and abstractions only when they
express a real public contract or remove real duplication. Delete obsolete in-progress design paths
instead of leaving "maybe later" branches. Preserve compatibility for shipped public APIs and
persisted data; for unshipped branch work, replace the old shape cleanly.

## Context To Gather

For every non-trivial change, gather these before editing:

- Design source: `design/*.go`, `services/*/design/*.go`, or the project's equivalent.
- Generated contract: `gen/<service>/service.go`, plus `gen/<service>/endpoints.go` when wiring or interceptors are involved.
- Transport mapping: `gen/http/...` or `gen/grpc/...` only to understand output, not to edit it.
- Implementation: the concrete service type implementing the generated interface.
- Generation commands: scripts, Makefile targets, task runner entries, or README instructions.
- Downstream consumers: generated clients, CLI clients, docs, examples, tests, mocks, and public API snapshots.

If you cannot identify the generation command, search project docs and scripts. If there is still
no wrapper, use `goa gen <design-package-import-path>` with a Go import path, never
`goa gen ./design`.

Before running generation in a dirty worktree, snapshot the relevant design and generated files
with `git diff -- <paths>`. Goa recreates `gen/` output, so distinguish pre-existing generated drift
from changes caused by your design edit. Do not "fix" unrelated generated churn unless the design
change requires it.

## Generated Diff Triage

After `goa gen`, inspect generated changes by artifact role instead of scanning every file uniformly:

- `gen/<service>/service.go`: canonical service interface, payload/result pointer semantics, generated error constructors, `ServiceName`, and `MethodNames`. Implementation compile failures should be resolved by matching this file exactly.
- `gen/<service>/endpoints.go`: endpoint construction, endpoint middleware application, and client/server endpoint signatures. Wiring changes should trace back here.
- `gen/<service>/interceptor_wrappers.go`, `service_interceptors.go`, and `client_interceptors.go`: interceptor ordering, accessor availability, and stream send/recv wrapping. If order matters, read the generated `Wrap<Method>Endpoint`; the last wrapper runs first on the request path.
- `gen/http/<service>/{server,client}/encode_decode.go`, `paths.go`, and `types.go`: HTTP path/query/header/body mapping, element wire names, content negotiation, and request/response body shape.
- `gen/grpc/<service>/pb/*.proto`, `pb/*.pb.go`, and transport `types.go`: protobuf field numbers, message names, metadata, streaming RPC shape, and gRPC status mapping.
- OpenAPI output: public HTTP contract shape. Goa may emit OpenAPI, protobuf, transport clients, and CLIs, but it does not generate third-party SDKs.

Stop generated-diff review when every changed generated artifact has a direct design cause. If a
generated file changed without an obvious cause, inspect the relevant Goa doc and the design
expression before editing implementation code.

## Goa-Specific Stop Criteria

Use these checks to avoid broad, generic API work:

- HTTP object payloads: confirm every path token maps to a payload field, every non-body attribute lands where intended, and `Body("field")` is used only when the request body should be that field's raw value. Unmapped object attributes are encoded in the body.
- HTTP streaming: plain HTTP streaming defaults to WebSocket; add `ServerSentEvents()` only for one-way server-to-client streams. WebSocket endpoints use `GET`.
- Views: if result views are involved, remember Goa emits a service-level views package, omits nil viewed attributes, and carries the selected view through the `Goa-View` header.
- Defaults and pointers: required or defaulted payload/result fields are direct values; optional fields are pointers; objects are pointers; arrays and maps are not pointers. Defaults apply to missing optional fields during unmarshalling, not to missing required fields. Empty arrays are indistinguishable from nil after marshaling/unmarshalling, so never use nil versus empty slice shape as API meaning.
- gRPC: never renumber released `Field` values. Check `.proto` output after changing shared types, streaming methods, or custom protobuf metadata.
- Errors: every exposed error must have per-transport `Response` mappings for each enabled transport. Prefer `ErrorResult`; use a custom error type only when clients need structured fields, and include `Meta("struct:error:name")` when multiple custom errors can return from one method.
- Interceptors: Goa interceptors are generated endpoint wrappers, not transport middleware. Server interceptors run after transport decoding and before the service method; client interceptors run around the typed client endpoint before transport encoding and after transport decoding.
- Observability wiring: prefer generated `ServiceName`, `APIVersion`, and method metadata for stable telemetry labels.

## Standard Change Loop

1. Explain the intended contract change in one or two sentences.
2. Edit the Goa design first. Organize changes as services, then methods, then shared types.
3. Regenerate using the project wrapper or `goa gen <design-package-import-path>`.
4. Implement the generated interface outside `gen/`.
5. Update mocks, clients, docs, examples, and command wiring that consume the changed contract.
6. Add deterministic tests for successful behavior and meaningful error paths.
7. Run the relevant generation, lint, type-check, and test commands allowed by the repository instructions. If tests are prohibited or too broad for the request, state exactly what was not run.
8. Review generated diffs and implementation diffs together. Generated changes should directly correspond to the design change.

Use `goa example <design-package-import-path>` only for first-time scaffolding. It creates owned
implementation files and does not overwrite existing custom implementation later.

## Designing Methods And Types

When adding a method, specify a clear method name, `Description`, `Payload`, `Result`, and the exact
errors the method can return. Do not leak transport details into domain type names; express HTTP and
gRPC concerns in transport mappings.

When modeling data, prefer focused `Type` definitions with field descriptions. Mark every required
field with `Required`. Use built-in formats such as `FormatUUID`, `FormatEmail`, `FormatDateTime`,
and `FormatURI` instead of ad hoc validators when they match the domain. Avoid `Any` unless
arbitrary JSON is truly the API contract.

For reusable constraints on primitive values, define a named primitive alias with
`Type("Name", Primitive, func() { ... })`, then use that type in payloads and results. This keeps
validation and examples at the domain type instead of repeating `Format`, `Pattern`, `Enum`, or
length rules on every field:

```go
var AccountID = Type("AccountID", String, func() {
    Description("Stable account identifier.")
    Format(FormatUUID)
    Example("2551dfde-513e-4840-b1be-9bb78d5930e9")
})

Payload(func() {
    Field(1, "account_id", AccountID, "Account to query.")
    Required("account_id")
})
```

Use this for IDs, constrained codes, enum-like strings, timestamps, and other primitive domains with
a single canonical validation rule. If the type must be shared across services or generated into a
shared package, use the project's established shared-type helper or package metadata rather than
duplicating aliases in each service design.

Make presence semantics explicit. Do not use nil versus empty slices or maps to encode meaning;
empty arrays are indistinguishable from nil after marshaling/unmarshalling. If empty is valid, make
that clear in the design; if presence matters, add a separate explicit field. Required arrays should
not be empty unless the design explicitly permits that shape.

For gRPC or multi-transport APIs, use `Field(number, name, type, description)` and keep protobuf
field numbers stable after release. Prefer numbers 1-15 for frequent fields. Use `Attribute` only for
HTTP-only designs.

Use `Reference`, `Extend`, views, shared generated packages, and `Meta(...)` customizations only when
they make the public contract clearer. Avoid broad shared types, deep inheritance, OpenAPI overrides,
or protobuf overrides that only hide duplication.

Remember generated type semantics when implementing: required or defaulted payload/result fields are
direct values, optional fields are pointers, objects are pointers, and arrays/maps are not pointers.

## Mapping HTTP And gRPC

For HTTP APIs, use resource nouns, plural names, stable hierarchy, and HTTP methods for actions. Put
stable prefixes in API-level or service-level `HTTP(Path(...))`. Use `Parent` and
`CanonicalMethod` only when nested resources become clearer. Map path, query, header, and body fields
deliberately with `Param`, `Header`, and `Body`, including explicit wire-name mapping when attribute
names differ.

Choose status codes intentionally. Use `StatusCreated` for creates when appropriate. Map every
exposed error to an HTTP status. Use default content negotiation unless a real media type requires
custom encoders or decoders. Use the CORS plugin for browser-facing cross-origin policy. Use `Files`
only for HTTP static content. Use `ServerSentEvents()` only for one-way server streams; otherwise
HTTP streaming defaults to WebSocket.

For gRPC APIs, set package and version metadata for public or versioned APIs. Keep domain data in
messages; use metadata, headers, and trailers only for protocol metadata. Use streaming for large or
continuous datasets. In streaming implementations, handle flow control, `io.EOF`, send errors,
context cancellation, timeouts, and cleanup.

## Handling Errors

Define common errors at API or service scope and operation-specific errors at method scope. Use
descriptive names and descriptions. Prefer `ErrorResult` for most errors, and add `Temporary()`,
`Timeout()`, or `Fault()` when clients should behave differently.

Use custom error types only when clients need structured context. If multiple custom errors can
return from the same method, include a field marked with `Meta("struct:error:name")`.

Map every exposed error for each enabled transport with `HTTP(Response(...))` and
`GRPC(Response(...))`. In implementation code, return generated error constructors or generated
custom error payloads. Wrap underlying causes for logs and tracing, but keep client-facing messages
safe.

Test errors through the service API. Verify generated error names, transport status mappings when
practical, and custom error payloads.

## Security, Interceptors, Middleware

Define authentication and authorization in the design. Prefer API-level defaults, override at service
or method scope when needed, and use `NoSecurity()` explicitly for public methods. Add matching
security payload fields such as `TokenField` or `APIKeyField`, then map them through HTTP
headers/query parameters or gRPC metadata.

Use Goa interceptors for type-safe domain concerns: business validation, request enrichment, response
enrichment, auditing, and domain-level transformations. Keep each interceptor focused on one
responsibility. Use generated typed accessors. Call `next` exactly once when continuing. If order
matters, inspect generated `Wrap<Method>Endpoint`; the last wrapper runs first on the request path.

Use HTTP middleware or gRPC interceptors for protocol concerns: logging, tracing, CORS, compression,
request IDs, rate limiting, panic recovery, and wire-level metadata. Test behavior-affecting
middleware and interceptors in isolation.

## Implementation Guidance

Keep service methods small and direct. Check every dependency error, wrap errors with useful context,
and use `errors.Is` or `errors.As` when branching on error kind. Do not ignore errors or assign them
to `_`.

Prefer named helpers over large anonymous functions. Split complex methods around clear contracts,
but add abstractions only when they simplify the domain or remove meaningful duplication.

Use the project's Go style: grouped imports, formatted code, short lowercase package names, comments
for exported identifiers, and contract comments for non-trivial helpers.

For production services, wire observability, health checks, timeouts, graceful shutdown, and
configuration at application boundaries. Goa docs recommend Clue/OpenTelemetry. Use generated service
names, method names, endpoint wrappers, and interceptor metadata for stable telemetry labels.

## Downstream Artifact Discovery

After any design change, search for consumers of the generated service, payload, result, method,
route, or error name. Update only artifacts affected by the contract:

- Service implementations and tests.
- Generated mocks or hand-written fakes.
- Goa-generated CLI clients, HTTP clients, gRPC clients, and downstream code that imports them.
- Docs, examples, API snapshots, OpenAPI consumers, protobuf consumers, and release notes when user-facing behavior changes.
- Deployment or runtime config only when ports, services, auth, health checks, or production wiring changed.

If a project has its own SDK generator built from Goa output, treat that as a project-specific
downstream artifact, not a Goa step. Follow the project's instructions for when to run it; do not
patch SDK drift with casts, duplicated types, or hand-written request code.

## Compatibility Decision

Before preserving old behavior, classify it:

- Shipped public API, persisted data, external SDK, or documented behavior: preserve compatibility or plan a migration/deprecation.
- Internal-only service contract: update callers in the same change instead of adding shims.
- Unshipped branch work or temporary generated output: replace it cleanly and delete obsolete paths.

When compatibility is required, express it in the design and tests. Do not hide it in service
fallbacks that the generated contract does not describe.

## Troubleshooting

If `goa gen` fails, read the exact error, then inspect the design around the named service, method,
field, or error. Common causes are duplicate field numbers, missing required field definitions,
invalid HTTP path parameter mappings, mismatched security fields, or invalid custom error metadata.

If implementation no longer compiles after generation, open `gen/<service>/service.go` and implement
the new generated interface exactly. Do not change generated signatures.

If HTTP decoding is wrong, compare payload fields with `GET`/`POST` path tokens, `Param`, `Header`,
and `Body` mappings. Every path token must map to a payload field.

If gRPC output is wrong, inspect `Field` numbers, package metadata, `GRPC(Message(...))`, metadata
mappings, and generated protobuf output. Fix the design and regenerate.

If interceptors behave in the wrong order, inspect generated `Wrap<Method>Endpoint` or stream
wrappers. The wrapper chain is the source of truth.

## Mini Patterns

Good design-level validation:

```go
Payload(func() {
    Field(1, "account_id", String, "Account identifier", func() {
        Format(FormatUUID)
    })
    Field(2, "limit", Int, "Maximum number of items", func() {
        Default(50)
        Minimum(1)
        Maximum(100)
    })
    Required("account_id")
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

## Final Review

Before finishing, verify:

- The design is still the canonical contract.
- Generated code was regenerated, not edited.
- Payloads, results, validation, security, errors, and transport mappings are explicit.
- Implementation code trusts Goa boundary validation and handles real dependency failures.
- Downstream clients, mocks, docs, examples, and tests are updated when affected.
- Tests cover successful behavior, meaningful error paths, and streaming or interceptor cleanup when relevant.
