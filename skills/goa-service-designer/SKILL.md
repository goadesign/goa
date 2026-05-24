---
name: goa-service-designer
description: Design and evolve Goa application services. Use for Goa DSL changes, service methods, payload/result modeling, validation, HTTP/gRPC mappings, errors, interceptors, generation, and implementation.
---

# Goa Service Designer

Use this skill when working in an application repository that uses Goa. If the task is about changing
Goa's own compiler, runtime, templates, or generators, follow that repository's contributor
instructions instead.

## Default Workflow

1. Classify the Goa change before editing:
   - New project or first service: read [Quickstart](https://goa.design/docs/1-goa/quickstart/) and [Code Generation](https://goa.design/docs/1-goa/code-generation/).
   - New method, payload, result, validation, views, streaming shape, or shared type: read [DSL Reference](https://goa.design/docs/1-goa/dsl-reference/).
   - HTTP path, query, headers, body, CORS, content negotiation, static files, WebSocket, or SSE: read [HTTP Guide](https://goa.design/docs/1-goa/http-guide/).
   - gRPC, protobuf field numbers, metadata, trailers, or streaming: read [gRPC Guide](https://goa.design/docs/1-goa/grpc-guide/).
   - Error names, error payloads, status-code mapping, or generated error constructors: read [Error Handling](https://goa.design/docs/1-goa/error-handling/).
   - Goa interceptors, HTTP middleware, gRPC interceptors, or ordering: read [Interceptors](https://goa.design/docs/1-goa/interceptors/).
   - Authentication, production wiring, observability, health checks, shutdown, or timeouts: read [Production](https://goa.design/docs/1-goa/production/).
2. Find the local design source, generated service contract, implementation, generation command,
   validation commands, and downstream consumers before choosing a pattern.
3. State the intended contract change in one or two sentences.
4. Edit the Goa design first.
5. Regenerate with the project's wrapper or `goa gen <design-package-import-path>`. Never use
   `goa gen ./design`.
6. Implement the generated interface outside `gen/`.
7. Update only affected consumers: mocks, clients, docs, examples, API snapshots, OpenAPI/protobuf
   consumers, command wiring, and project-specific SDKs generated from Goa output.
8. Run the generation, lint, type-check, and test commands allowed by the repository instructions.
   If a command is prohibited or too broad, state exactly what was not run.
9. Review generated and hand-written diffs together. Every generated change should have a direct
   design cause.

Use `goa example <design-package-import-path>` only for first-time scaffolding. It creates owned
implementation files and does not overwrite existing custom implementation later.

## Hard Rules

- Treat the Goa design as the canonical contract. If payload shape, result shape, validation,
  security, errors, transport mapping, documentation, OpenAPI, protobuf, generated clients, or CLIs
  need to change, edit the design first and regenerate.
- Never patch generated files, OpenAPI output, protobuf output, or generated clients to hide a stale
  design.
- Import generated packages with explicit aliases prefixed by `gen`, such as `genfront`,
  `gendefinitions`, or `genruns`, so generated Goa contracts are recognizable at call sites.
- In Goa design, type field names use CamelCase (`Field(1, "AccountID", ...)`,
  `Required("AccountID")`). Method names use snake_case (`Method("get_account", ...)`).
- Put boundary validation in Goa. Use `Required`, `Enum`, `Format`, `Pattern`, `Minimum`,
  `Maximum`, `MinLength`, `MaxLength`, defaults, security fields, and explicit transport mappings in
  the design.
- Model error applicability at the narrowest correct Goa scope. API-level errors define reusable
  potential errors and let transports map them once; they do not mean every endpoint returns every
  API-level error. Service-level errors apply to all methods in that service. Method-level errors
  apply only to that method. Do not list an error at service scope just to make a generated
  constructor convenient if only some methods can return it.
- In service code, trust decoded payloads and enforce business invariants. Do not add nil guards,
  fallback behavior, silent recovery, blanket string trimming, or compatibility shims for values the
  design guarantees.
- Do not use nil versus empty slices or maps as API meaning. Empty arrays are indistinguishable from
  nil after marshaling/unmarshalling.
- Preserve compatibility for shipped public APIs, persisted data, external SDKs, and documented
  behavior. For internal-only contracts or unshipped branch work, update callers cleanly instead of
  adding shims.

## What To Read Next

Load these references only when the task needs them:

- `references/modeling-and-validation.md`: type modeling, primitive aliases, required fields,
  pointer/default semantics, views, presence, and compatibility.
- `references/transport-errors-interceptors.md`: HTTP mapping, gRPC mapping, streaming, errors,
  security, interceptors, middleware, and observability labels.
- `references/generated-code-and-implementation.md`: generated diff triage, dirty worktrees,
  implementation updates, downstream artifacts, and validation commands.
- `references/troubleshooting.md`: generation failures, decode bugs, gRPC output surprises,
  interceptor ordering, and common mini-patterns.

## Final Review

Before finishing, verify:

- The design remains the source of truth.
- Generated code was regenerated, not edited.
- Payloads, results, validation, security, errors, and transport mappings are explicit.
- Service code trusts Goa boundary validation and handles real dependency failures.
- Compatibility choices are represented in the design and tests, not hidden in service fallbacks.
- Affected downstream consumers are updated; unaffected artifacts are left alone.
- Test coverage matches the risk of the contract change.
