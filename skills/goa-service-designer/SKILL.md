---
name: goa-service-designer
description: Design and evolve Goa application services. Use for Goa DSL changes, HTTP/gRPC/JSON-RPC contracts, validation, errors, interceptors, code generation, and service implementation.
---

# Goa Service Designer

Use this skill when working in an application repository that uses Goa. If the task is about changing
Goa's own compiler, runtime, templates, or generators, follow that repository's contributor
instructions instead.

Let Goa generate the mechanical code. Use the relevant design and generated service interface as
the working context; inspect transport output when the task changes wire behavior. Implement the
application's decisions outside `gen/`.

## Default Workflow

1. Find the local design, generated service contract, implementation, generation wrapper, validation
   commands, and affected consumers. Check the Goa version and any replacement in `go.mod`.
2. Reuse established local patterns for routine edits. Consult goa.design when choosing an
   unfamiliar feature, resolving uncertain DSL or generated behavior, or troubleshooting.
   Read only the relevant section, using these topic routes:
   - New project: [Quickstart](https://goa.design/docs/1-goa/quickstart/) and [Code Generation](https://goa.design/docs/1-goa/code-generation/).
   - Types, validation, views, or unions: [DSL Reference](https://goa.design/docs/1-goa/dsl-reference/).
   - HTTP mapping, WebSocket, or SSE: [HTTP Guide](https://goa.design/docs/1-goa/http-guide/).
   - gRPC, protobuf tags, metadata, or streaming: [gRPC Guide](https://goa.design/docs/1-goa/grpc-guide/).
   - JSON-RPC: [transport reference](references/transport-errors-interceptors.md#json-rpc-mapping) and the selected Goa version's `dsl.JSONRPC` documentation.
   - Errors: [Error Handling](https://goa.design/docs/1-goa/error-handling/).
   - Interceptors or middleware: [Interceptors](https://goa.design/docs/1-goa/interceptors/).
   - Security or production wiring: [Production](https://goa.design/docs/1-goa/production/).
   Online docs may describe a different release. For release-specific behavior and upgrades, use the
   selected Goa checkout's `UPGRADING.md` and generated output. Use a page's Markdown version
   when available; do not preload every guide or reference.
3. State the intended contract change in one or two sentences.
4. Edit the Goa design first.
5. Regenerate with the project's wrapper. Without one, use
   `go run goa.design/goa/v3/cmd/goa gen <design-package-import-path>` so the command uses the
   application's selected module version. Never use `goa gen ./design`.
6. Implement the generated interface outside `gen/`.
7. Update only affected consumers: mocks, clients, docs, examples, API snapshots, OpenAPI/protobuf
   consumers, command wiring, and project-specific SDKs generated from Goa output.
8. Run the generation, lint, type-check, and test commands allowed by the repository instructions.
   If a command is prohibited or too broad, state exactly what was not run.
9. Review generated and hand-written diffs together. Every generated change should follow from
   the design or an intentional tooling upgrade.

Use `goa example <design-package-import-path>` when starter files are needed. It creates missing
application files and does not update existing files. Review new files and update existing wiring
yourself; do not expect it to migrate an implementation.

## Hard Rules

- Treat the Goa design as the canonical contract. If payload shape, result shape, validation,
  security, errors, transport mapping, documentation, OpenAPI, protobuf, generated clients, or CLIs
  need to change, edit the design first and regenerate.
- Never patch generated files, OpenAPI output, protobuf output, or generated clients to hide a stale
  design.
- Keep protobuf field numbers unique within the final message, including inherited fields.
  Never renumber or reuse released field numbers to make a sequence contiguous.
- Put boundary validation in Goa. Use `Required`, `Enum`, `Format`, `Pattern`, `Minimum`,
  `Maximum`, `MinLength`, `MaxLength`, defaults, security fields, and explicit transport mappings in
  the design.
- Model error applicability at the narrowest correct Goa scope. API-level errors define reusable
  potential errors and let transports map them once; they do not mean every endpoint returns every
  API-level error. Service-level errors apply to all methods in that service. Method-level errors
  apply only to that method. Do not list an error at service scope just to make a generated
  constructor convenient if only some methods can return it.
- In service code, trust payloads already validated by the generated transport and enforce business
  invariants. Direct service or endpoint calls bypass transport decoding; their callers must supply
  valid values or validate at their own input boundary. Do not add nil guards,
  fallback behavior, silent recovery, blanket string trimming, or compatibility shims for values the
  design guarantees.
- Do not use nil versus empty slices or maps as domain meaning. `Required` and `MinLength` express
  different constraints, and JSON and protobuf have different collection presence rules. See
  [presence and collections](references/modeling-and-validation.md#presence-and-collections).
- Preserve compatibility for shipped public APIs, persisted data, external SDKs, and documented
  behavior. For internal-only contracts or unshipped branch work, update callers cleanly instead of
  adding shims.

## Design Conventions

Follow the application's existing conventions and compatibility requirements. For new designs:

- Prefix explicit generated-package aliases with `gen`, such as `genfront` or `genfrontgrpc`.
- Prefer lower camel case for public HTTP fields and parameters (`accountId`) and snake case for
  method names (`get_account`). Map established wire names explicitly, for example
  `Header("tusResumable:Tus-Resumable")`.
- Use literal integer `Field` tags so wire identities are easy to review. Start a new standalone
  definition at `1`. For an `Extend` definition, `100` can start a new range only if inherited fields
  leave that range free. Inspect the complete inheritance chain and assign unused numbers; nested
  `Extend` calls must not restart at an occupied `100`. Existing tags always keep their numbers.

These are authoring conventions, not requirements imposed by Goa.

## What To Read Next

Load these references only when the task needs them:

- `references/modeling-and-validation.md`: type modeling, primitive aliases, required fields,
  pointer/default semantics, views, presence, and compatibility.
- `references/transport-errors-interceptors.md`: HTTP, gRPC, JSON-RPC, streaming, errors, security,
  interceptors, middleware, and observability labels.
- `references/generated-code-and-implementation.md`: generated diff triage, dirty worktrees,
  implementation updates, downstream artifacts, and validation commands.
- `references/troubleshooting.md`: generation failures, decode bugs, gRPC output surprises,
  interceptor ordering, and common mini-patterns.

## Final Review

Before finishing, verify:

- The design remains the source of truth.
- Generated code was regenerated, not edited.
- Payloads, results, validation, security, errors, and transport mappings are explicit.
- Service code trusts established validation boundaries and handles real dependency failures.
- Compatibility choices are represented in the design and tests, not hidden in service fallbacks.
- Affected downstream consumers are updated; unaffected artifacts are left alone.
- Test coverage matches the risk of the contract change.
