# Generated Code And Implementation

Use this reference when regenerating Goa output, reviewing generated diffs, implementing service
interfaces, or updating downstream artifacts.

## Dirty Worktree Protocol

Before running generation in a dirty worktree, snapshot the relevant design and generated files:

```bash
git diff -- design services gen
```

Adjust paths to the project layout. Goa recreates generated output, so distinguish pre-existing
generated drift from changes caused by your design edit. Do not fix unrelated generated churn unless
the design change requires it.

## Regeneration

- Prefer the project's wrapper command, such as `make gen`, `scripts/gen`, or a task runner entry.
- If there is no wrapper, use `goa gen <design-package-import-path>`.
- Never use `goa gen ./design`; Goa expects a Go import path.
- `goa example <design-package-import-path>` is only for first-time scaffolding.

## Generated Diff Triage

After `goa gen`, inspect generated changes by artifact role instead of scanning every file uniformly:

- `gen/<service>/service.go`: canonical service interface, payload/result pointer semantics,
  generated error constructors, `ServiceName`, and `MethodNames`. Implementation compile failures
  should be resolved by matching this file exactly.
- `gen/<service>/endpoints.go`: endpoint construction, endpoint middleware application, and
  client/server endpoint signatures. Wiring changes should trace back here.
- `gen/<service>/interceptor_wrappers.go`, `service_interceptors.go`, and `client_interceptors.go`:
  interceptor ordering, accessor availability, and stream send/recv wrapping.
- `gen/http/<service>/{server,client}/encode_decode.go`, `paths.go`, and `types.go`: HTTP
  path/query/header/body mapping, element wire names, content negotiation, and request/response body
  shape.
- `gen/grpc/<service>/pb/*.proto`, `pb/*.pb.go`, and transport `types.go`: protobuf field numbers,
  message names, metadata, streaming RPC shape, and gRPC status mapping.
- OpenAPI output: public HTTP contract shape.

Goa may emit OpenAPI, protobuf, transport clients, and CLIs. It does not generate third-party SDKs.

Stop generated-diff review when every changed generated artifact has a direct design cause. If a
generated file changed without an obvious cause, inspect the relevant Goa docs and the design
expression before editing implementation code.

## Implementation Updates

- Implement the generated service interface exactly.
- Keep service methods small and direct.
- Check every dependency error.
- Wrap dependency errors with useful context using `%w`.
- Use `errors.Is` or `errors.As` when branching on error kind.
- Do not ignore errors or assign them to `_`.
- Do not revalidate values guaranteed by Goa validation.
- Prefer named helpers over large anonymous functions.
- Split complex methods around clear contracts, but add abstractions only when they simplify the
  domain or remove meaningful duplication.
- Follow the project's Go style: grouped imports, formatted code, short lowercase package names,
  exported identifier comments, and contract comments for non-trivial helpers.

## Downstream Artifact Discovery

After any design change, search for consumers of the generated service, payload, result, method,
route, or error name. Update only artifacts affected by the contract:

- Service implementations and tests.
- Generated mocks or hand-written fakes.
- Goa-generated CLI clients, HTTP clients, gRPC clients, and downstream code that imports them.
- Docs, examples, API snapshots, OpenAPI consumers, protobuf consumers, and release notes when
  user-facing behavior changes.
- Deployment or runtime config only when ports, services, auth, health checks, or production wiring
  changed.

If a project has its own SDK generator built from Goa output, treat that as a project-specific
downstream artifact, not a Goa step. Follow the project's instructions for when to run it. Do not
patch SDK drift with casts, duplicated types, or hand-written request code.

## Validation Commands

Run the relevant generation, lint, type-check, and test commands allowed by the repository
instructions. If a command is prohibited, unavailable, or too broad for the request, state exactly
what was not run and why.

Test coverage should scale with contract risk:

- Narrow design-only doc changes may need no tests.
- Service behavior changes need focused service tests.
- Transport mapping changes need transport-level checks when practical.
- Error contract changes need tests for generated error names and status/code mappings.
- Streaming or interceptor changes need cleanup, cancellation, and ordering coverage.
