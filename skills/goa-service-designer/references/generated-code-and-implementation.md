# Generated Code And Implementation

Use this reference when regenerating Goa output, reviewing generated diffs, implementing service
interfaces, or updating downstream artifacts.

## Dirty Worktree Protocol

Before running generation in a dirty worktree, inspect tracked, staged, and untracked changes:

```bash
git status --short
git diff -- design services gen
git diff --cached -- design services gen
git ls-files --others --exclude-standard -- design services gen
```

Adjust paths to the project layout. These commands inspect changes; they do not back them up.
Goa deletes and recreates `gen/`. Preserve any work that could be lost with a verified copy or an
isolated worktree before generation. Do not stash, reset, discard, or commit someone else's changes
as an automatic preparation step. Distinguish pre-existing drift from changes caused by this task.

## Regeneration

- Prefer the project's wrapper command, such as `make gen`, `scripts/gen`, or a task runner entry.
- Check the project's selected Goa module, including `replace` directives, before choosing tooling.
  Check that release's Go toolchain requirements before upgrading the module.
- If there is no wrapper, use
  `go run goa.design/goa/v3/cmd/goa gen <design-package-import-path>`. A standalone `goa` command
  should match the project's module version.
- Never use `goa gen ./design`; Goa expects a Go import path.
- For gRPC, follow the selected release's protobuf compiler and plugin requirements. Goa v3.31
  and later check `protoc-gen-go` and `protoc-gen-go-grpc` versions; do not fix a mismatch by changing
  generated files or upgrading the application without authorization.
- `goa example` creates missing starter files without overwriting existing files. It can scaffold
  a newly added service, but existing application wiring must be updated explicitly.

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
- `gen/jsonrpc/<service>/{server,client}/`: request IDs, notifications, parameter shapes, error
  mappings, batches, and SSE encoding.
- OpenAPI output: public HTTP contract shape.

Goa may emit OpenAPI, protobuf, transport clients, and CLIs. It does not generate third-party SDKs.

Every generated change should follow from the design or an intentional tooling upgrade. If a
file changes without an obvious cause, check the generator version, plugins, design, and relevant
documentation before editing implementation code. Never combine output from separate generation
runs. For generator upgrades, regenerate twice and confirm the second run introduces no changes.

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
- Goa-generated CLI clients, HTTP, gRPC, and JSON-RPC clients, and downstream code that imports them.
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
