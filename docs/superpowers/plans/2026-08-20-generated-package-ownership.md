# Generated Package Ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every generated package-level declaration and every reference use one name collected once, frozen once, and retained through rendering.

**Architecture:** One generation run instantiates fresh core and plugin objects, permits root mutation only during preparation, and builds one typed `generator.Plan`. Service, HTTP, JSON-RPC, gRPC, OpenAPI, example, and goa-ai plans collect package-owned `NameDeclaration` records, freeze them, link those records once into complete immutable render models, and render without rebuilding analysis or allocating a name.

**Tech Stack:** Go 1.25, Goa evaluation and code generation, Protocol Buffers and `protoc-gen-go`, goa-ai plugins, `testify/require`

**Spec:** `codegen/ARCHITECTURE.md`

## Global Constraints

- Never edit generated output; regenerate it from the owning design.
- Preparation is the last phase allowed to mutate an expression root.
- One run owns one immutable root snapshot, one `codegen.Generation`, and one typed `generator.Plan`.
- Every package-level type, function, constant, and variable has a package-owned `NameDeclaration` before freeze.
- Exact symbols reject normalized collisions; preferred symbols receive deterministic suffixes from stable typed ordering.
- `NameDeclaration.Name()` panics before freeze and is stable after freeze.
- Linking resolves retained facts through frozen declarations exactly once; it cannot collect another declaration or import.
- Render accepts retained typed plans. It does not accept roots, a generated module path, or callbacks that reconstruct analysis.
- Complete import path is the only import identity. Different package identities that normalize to one output import path or directory are rejected.
- Recursion uses `UserType.Origin()` only for cycle detection. Emitted declarations use complete typed declaration identities.
- Keep `expr.Union.Hash()` unchanged.
- Protoc-generated Go names come from one explicit, versioned naming contract that covers complete declaration families.
- Plugins consume the exact core service plan. Do not add `PlanKey`, plan registries, generic plan bags, reconstruction, decorated string keys, or process-global run state.
- `SectionTemplate.Name` is diagnostic metadata, not declaration identity.
- Never add fallbacks or compatibility modes. Migrate all in-tree callers and delete replaced APIs.
- Every exported construct needs GoDoc; every non-trivial file needs a concrete purpose and invariant header.

---

## Completed foundation

The following tasks are complete against their reviewed contracts. Their tests,
typed identities, import-path ownership, and transport correctness remain
required. The retained-plan audit supersedes their transitional callback and
reconstruction APIs; “complete” here records delivered history, not approval to
preserve those APIs.

### Task 1: Executable failure contracts — complete

**Commits:** `0103f6ab`, `dac45fe7`

- [x] Added a real two-service generated-module test with relocated nested unions and HTTP/gRPC compilation.
- [x] Added exact relocated-name collision coverage.
- [x] Added the same-label file-section regression later resolved by Task 6.

### Task 2: Generation-owned type catalog — complete

**Commits:** `15f86ce7`, `7353f34b`, `8bda1ae4`

- [x] Added one package catalog per generated import path.
- [x] Reserved exact user declarations before deterministic union allocation.
- [x] Froze package scopes and rejected planning through render-only accessors.

### Task 3: Typed union identity and initial lifecycle — complete

**Commits:** `4957ddef`, `a120d8ea`

- [x] Added `UnionTypeID` without changing expression hash semantics.
- [x] Established prepare, plan, freeze, and render ordering.
- [x] Proved one generation reaches core and plugin callbacks.

The retained-plan audit replaces `Genfunc`, callback plugin instances, and
render-time reconstruction introduced during this transition.

### Task 4: Package-owned service declarations — complete

**Commits:** `81b8bc37`, `4ca1f70c`, `5469ad1e`, `14b1a3c4`, `1bf62f51`, `839791a6`

- [x] Added typed authored origin, derived method/view identities, complete union families, full-path import aliases, and cross-root package emission.
- [x] Made declaration records immutable and deterministic.
- [x] Bound service and views references to frozen package records.
- [x] Rejected unregistered roots and immutable-generation mutation.

The retained-plan audit expands this ownership from selected type families to
every service package-level symbol and replaces `Plan` plus
`NewServicesData` re-analysis with one retained `service.Plan`.

### Task 5: Transport declaration ownership — complete

**Commits:** `595f50ad`, `8dc8ea8e`, `c6a0e1e0`, `393e9102`, `ddfc0472`

- [x] Routed HTTP, gRPC, and JSON-RPC service references through exact frozen service declarations.
- [x] Added HTTP/JSON-RPC and protobuf wire catalogs, independent transform ownership, native gRPC metadata, generated import planning, and effective inherited-error validation.
- [x] Distinguished cycle identity from transport declaration identity and compiled the integrated generated modules.

The retained-plan audit keeps these semantics and converts the render-time
catalog construction into retained HTTP, JSON-RPC, protobuf, and gRPC plans.

---

## Remaining implementation

### Task 6: Common declarations and fresh run lifecycle

**Files:**
- Modify: `codegen/generation.go`
- Modify: `codegen/generated_types.go`
- Modify: `codegen/generated_types_test.go`
- Modify: `codegen/import_aliases.go`
- Modify: `codegen/normalize.go`
- Delete: `codegen/plugin.go`
- Delete: `codegen/plugin_test.go`
- Modify: `codegen/generator/generate.go`
- Replace: `codegen/generator/generators.go`
- Create: `codegen/generator/plan.go`
- Create: `codegen/generator/plugin.go`
- Create: `codegen/generator/plugin_test.go`
- Modify: `codegen/generator/generation_test.go`
- Modify: `codegen/generator/purity_test.go`
- Modify: `codegen/generator/generate_merge_test.go`
- Modify: `codegen/generator/service_union_package_scope_test.go`
- Modify: `codegen/generator/generate_http_union_shape_integration_test.go`
- Modify: `codegen/generator/generate_union_merge_integration_test.go`
- Modify: `codegen/walk.go`
- Modify: `codegen/import.go`
- Modify: `codegen/validation.go`
- Modify: `codegen/example/plan.go`
- Modify: `codegen/example/example_client.go`
- Modify: `codegen/example/example_server.go`

**Interfaces:**
- Produces: package-owned `NameDeclaration` for type/function/constant/variable symbols
- Produces: `generator.Plugin`, `PluginFactory`, fresh core factories, and private-field `generator.Plan`
- Preserves: `Generation`, import-path bindings, `TypeDeclaration`, `UnionDeclaration`, and typed declaration identities

- [x] **Step 1: Add declaration and lifecycle RED tests**

Add table-driven tests proving one package namespace catches cross-kind
collisions, exact names reject, preferred names suffix in stable typed order,
`Name()` panics before freeze, and every existing type/union record returns its
contained canonical name record. Add canonical output-path tests where two
different package identities normalize to one import path or directory.

Add repeated and concurrent generator tests. Register a factory whose plugin
keeps per-run counters, run generation twice and in parallel, and prove each run
starts at zero and receives only its own roots, plan, and files. Attempt root
mutation after preparation and require rejection or a purity failure at the
owning boundary.

Run:

```bash
go test ./codegen ./codegen/generator \
  -run 'TestNameDeclaration|TestGeneratedOutputPath|TestPluginFactory|TestConcurrentGeneration|TestPreparedRoots' \
  -count=1
```

Expected: FAIL because names are still type-family-specific, plugins are
registered as callback instances in `codegen`, and `Generators` is mutable
process-global run state.

- [x] **Step 2: Implement the common declaration owner**

Add private preferred/final state and a package-level symbol kind to
`NameDeclaration`. Make exact and preferred declaration APIs return the same
record on idempotent typed identity and reject one identity binding to two
records. Allocate exact records first and preferred records in stable typed
order during `Generation.Freeze`.

Embed or reference `NameDeclaration` from existing type, union, union branch,
imported toolchain, and later subsystem records. Remove duplicate name fields
as each owner migrates. Canonicalize output paths during collection and reject
different package owners that converge after normalization.

- [x] **Step 3: Move orchestration and plugin registration into generator**

Implement the approved public surface:

```go
type Plugin struct {
    Prepare  PrepareFunc
    Plan     func(*Plan) error
    Generate func(*Plan, []*codegen.File) ([]*codegen.File, error)
}

type PluginFactory func() Plugin

func RegisterPlugin(name, command string, factory PluginFactory)
func RegisterPluginFirst(name, command string, factory PluginFactory)
func RegisterPluginLast(name, command string, factory PluginFactory)

func (p *Plan) Generation() *codegen.Generation
```

Store immutable factory descriptors and instantiate fresh plugins and core
generators before each run. Make `Generation` construction the final
preparation operation: normalize raw method objects there, snapshot the design
immediately afterward, and reject every later mutation. Delete `Genfunc`, the public
replaceable `Generators` variable, `renderOnly`, and the callback registry in
`codegen/plugin.go`. Tests install an isolated registry or command factory
through a private test seam, not a mutable production global.

- [x] **Step 4: Finish mechanical identity and example cleanup**

Audit every cycle-only walk and key it by `UserType.Origin()`. Keep semantic
`ID()` only where it identifies a named user type or a public semantic
identifier. Give every generated example a kind-tagged identity derived from
its exact owning expression: user type, method payload/result/error, HTTP
request/success/error body, object member, array element, map key/value, or
union branch. Reject unanchored draws and remove delimiter-joined paths,
caller-supplied response ordinals, and shared sequential collection streams.
Give independently mapped HTTP and JSON-RPC body types distinct stable semantic
IDs derived from their exact typed body owners so the recursive example cache
cannot return one transport's body for the other. Preserve authored type IDs
and expression hash behavior.
Remove render-time example scopes that own package-level names; leave local
argument and field scopes local. Add focused cross-kind, delimiter, response
reordering, dual-transport order, repeated-analysis, and concurrent-run
counterexamples.

- [x] **Step 5: Verify and commit Task 6**

Run:

```bash
go fmt ./...
go test ./codegen ./codegen/generator -count=1
go test ./... -run '^$'
git diff --check
```

All commands pass.
Commit the common owner and fresh-run lifecycle together because retained plans
depend on both contracts.

### Task 7: Retained service plan and complete core symbols

**Files:**
- Replace: `codegen/service/generated_package.go`
- Replace: `codegen/service/service_data.go`
- Modify: `codegen/service/service.go`
- Modify: `codegen/service/client.go`
- Modify: `codegen/service/endpoint.go`
- Modify: `codegen/service/views.go`
- Modify: `codegen/service/convert.go`
- Modify: `codegen/validation.go`
- Modify: `codegen/service/example_svc.go`
- Modify: `codegen/service/declaration_resolver.go`
- Modify: service headers/import builders that emit package symbols
- Modify: `codegen/generator/plan.go`
- Modify: `codegen/generator/service.go`
- Test: `codegen/service/*_test.go`
- Test: `codegen/generator/service_union_package_scope_test.go`

**Interfaces:**
- Consumes: Task 6 `Generation`, `NameDeclaration`, and prepared root snapshot
- Produces: `service.NewPlans(generation, inputs...) ([]*service.Plan, error)`
- Produces: `service.NewPlan(root, generation, examples) (*service.Plan, error)`
- Produces: `generator.Plan.Service(root) *service.Plan`
- Produces: one post-freeze `service.Plan.Link()` operation before rendering
- Produces: service render functions that accept retained plans only

- [x] **Step 1: Inventory and test every service package-level symbol**

Build a table from templates and render data covering service and views types,
method wrappers, union families, endpoint constructors, clients,
errors, validators, conversions, interceptors, stream interfaces and helpers,
view constructors, and package variables. For each family, add a collision
fixture against a type and another generated function or constant. Assert the
declaration and every call site share the same `NameDeclaration` pointer, then
compile the generated service and views packages.

Run:

```bash
go test ./codegen/service ./codegen/generator \
  -run 'TestServicePlan|TestServicePackageDeclarations|TestRelocatedUnionPackageNamesCompile' \
  -count=1
```

Expected: FAIL where `NewServicesData` and private render scopes still allocate
package-level endpoint, constructor, validator, conversion, or stream names.

- [x] **Step 2: Build and retain the complete service-plan batch**

Replace the declaration-only `service.Plan` function and render-time
`NewServicesData` reconstruction with one run-wide constructor:

```go
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error)
```

The constructor requires every Goa root owned by the generation exactly once.
It collects every root-local design fact, package import, output owner, and
package-level declaration, then assigns shared conversion methods and relocated
files across the complete run without reading provisional names. Structurally
equivalent compiler copies may share one declaration; candidates with different
retained Go layouts or union branch facts are rejected before freeze. Exact
duplicate external conversions across roots are rejected rather than receiving
an artificial numeric suffix. `NewPlan` remains only as the strict single-root
form and rejects a multi-root generation. After Generation freezes, `Plan.Link`
resolves the frozen records into the immutable render model without another
declaration traversal. `generator.Plan` stores the exact result by root and
returns it through `Service(root)`; unknown roots fail fast.

- [x] **Step 3: Render only the retained plan**

Change service, views, client, endpoint, conversion, validation, interceptor,
and starter implementation renderers to accept `*service.Plan` or typed values
owned by that plan. Remove root, generation, generated module path, and mutable
scope parameters that permit re-analysis or redirected output. Keep lexical
scopes only for locals, parameters, fields, and methods.

Delete `NewServicesData`, the old `ServicesData` reconstruction constructor,
duplicate planning traversals, and any record that carries a second final name.

- [x] **Step 4: Prove aggregation, order independence, and purity**

Generate two roots contributing to one relocated package, reverse root and
service traversal, and assert byte-identical declarations. Mutate the source
service, method, type-location, field, validation, and conversion expressions
after planning, then assert core service output is byte-identical wherever the
core service plan owns those facts. Preserve distinct transport validation as
a valid counterexample: HTTP and gRPC own those validation programs when the
shared service declaration's Go layout is unchanged. Render twice without new
catalog entries and compile service, views, and example implementation packages.

- [x] **Step 5: Verify and commit Task 7**

Run:

```bash
go fmt ./...
go test ./expr ./dsl ./codegen ./codegen/service ./codegen/generator -count=1
go test ./... -run '^$'
git diff --check
```

All commands must pass.

### Task 8: Retained HTTP and JSON-RPC plans

**Files:**
- Replace: `http/codegen/plan.go`
- Replace: `http/codegen/service_data.go`
- Replace: `http/codegen/wire_catalog.go`
- Modify: `http/codegen/client.go`
- Modify: `http/codegen/server.go`
- Modify: `http/codegen/websocket.go`
- Modify: `http/codegen/sse.go`
- Modify: `http/codegen/sse_client.go`
- Modify: `http/codegen/types.go`
- Modify: `http/codegen/client_cli.go`
- Modify: `http/codegen/example_cli.go`
- Modify: `http/codegen/example_server.go`
- Replace: `jsonrpc/codegen/plan.go`
- Modify: `jsonrpc/codegen/client.go`
- Modify: `jsonrpc/codegen/server.go`
- Modify: `jsonrpc/codegen/websocket_client.go`
- Modify: `jsonrpc/codegen/websocket_server.go`
- Modify: `jsonrpc/codegen/example_server.go`
- Modify: `codegen/generator/plan.go`
- Modify: `codegen/generator/transport.go`
- Test: `http/codegen/plan_test.go`
- Test: `http/codegen/wire_catalog_test.go`
- Test: `http/codegen/service_data_purity_test.go`
- Test: `http/codegen/streaming_test.go`
- Test: `jsonrpc/codegen/plan_test.go`
- Test: `jsonrpc/codegen/kitchen_sink_test.go`
- Test: `jsonrpc/codegen/sse_integration_test.go`
- Test: `codegen/generator/service_union_package_scope_test.go`

**Interfaces:**
- Consumes: exact retained `*service.Plan`
- Produces: typed retained HTTP and JSON-RPC plans with complete package declarations
- Preserves: independent wire/service transform ownership and detached HTTP bodies

**Progress ledger (2026-08-21):**

- Task 7 now reserves service-view imports only in the HTTP, JSON-RPC, and gRPC
  files whose rendered sections reference them. The focused generated-module
  proof covers viewed and ordinary services, unary and streaming gRPC, HTTP
  SSE and WebSocket, and JSON-RPC unary code. The complete variable-view
  transport behavior remains Task 8 work rather than an import-planning
  exception.
- JSON-RPC unary responses currently discard the selected view: the server
  always renders the first retained response-body variant and sends no view,
  while the client tries to read `goa-view` from the HTTP response header.
  Task 8 must make the selected representation explicit and reconstruct the
  same view-specific body on both sides.
- JSON-RPC SSE and WebSocket clients currently decode `params` or `result`
  bytes directly into the service result. This is not a valid shortcut. For
  example, the generated Feed response body maps the wire property
  `event_id` to `EventID`, but the service result has no JSON tag; direct
  decoding silently leaves `EventID` unset. Task 8 must decode the retained
  transport body, run its generated constructor and validation, then return
  the canonical service result.

- [ ] **Step 1: Add complete HTTP/JSON-RPC declaration REDs**

Inventory request, response, WebSocket, SSE, error, union, constructor,
validator, codec, stream, client, server, CLI, and example package symbols.
Create collisions between wire types and validators/constructors, between
request and response policy for one origin, and between HTTP and JSON-RPC
sections sharing an output package. Require stable names under reversed
endpoint order and compile the full generated module.

Add two-view streaming-result contracts for HTTP SSE, JSON-RPC SSE, and
JSON-RPC WebSocket. Prove each method/request stream implements `SetView`,
retains its own view, projects through the canonical service constructor, and
selects the response-body declaration for that exact view. Use two concurrent
requests on one JSON-RPC WebSocket connection as the counterexample: a view
stored on the connection is invalid because one request may select `summary`
while another selects `detailed`.

Cover the direct JSON-RPC `StreamHandler` API separately. For a method whose
view is not fixed by the design, each `Send<Method>Notification` and
`Send<Method>Response` call must carry its own view; it must not inherit a
connection-global or latest value. Fixed-view methods remain specialized and
do not expose a redundant selector. Add client runtime proofs with nested and
transport-mapped fields so SSE and WebSocket receivers cannot pass by decoding
view-specific wire JSON directly into the service result. Include a required
snake-case field such as `event_id`: decoding it into a service field named
`EventID` must fail the test unless the generated transport-body constructor
performs the mapping.

- [ ] **Step 2: Build retained HTTP plans from exact service plans**

Make HTTP `NewPlan` consume the prepared root's HTTP expressions and exact
`*service.Plan`. Collect detached client and server wire models, union families,
validators, helpers, imports, and file membership once. Move every current
`NewServicesData` and `wire_catalog` allocation into this constructor. Store
canonical service and wire declaration records in transform data.

Retain one method/request-scoped view value for variable-view SSE and
WebSocket streams. The stream's `SetView` updates that value; send operations
use it to select both the service projection and the already-retained
view-specific response body. Never place mutable view selection on a shared
connection.

- [ ] **Step 3: Make JSON-RPC retain the HTTP plan it shares**

Build one typed JSON-RPC plan that points at the exact HTTP plan used for HTTP
codecs and body files, then collects JSON-RPC-only declarations. Do not invoke
HTTP planning or analysis again. Make JSON-RPC render functions accept this
plan and delete their root/service reconstruction paths.

Define one explicit viewed-stream wire contract shared by JSON-RPC SSE and
WebSocket. Every viewed streamed message carries the selected view together
with its view-specific body. Generated clients must select the matching
retained body decoder, reconstruct the projected value, validate the viewed
result, and return the canonical service result. Do not decode a projected
wire body directly into the service result, infer a default for a variable-view
method, or ask callers to construct generated views-package values.

Apply the same representation contract to unary JSON-RPC. A variable-view
success response must carry the selected view with its view-specific body;
the server cannot choose the first body variant and the client cannot recover
the view from an unset HTTP header. Fixed-view unary methods remain fully
specialized and need no runtime discriminator.

- [ ] **Step 4: Remove context-dependent helper naming**

Validators, constructors, conversions, stream helpers, and codecs must read
their `NameDeclaration`; call-site traversal selects a record but cannot name
it. Keep local field and variable scopes. Prove request/response and
WebSocket/SSE transforms enter service and wire owners independently.

- [ ] **Step 5: Verify and commit Task 8**

Run:

```bash
go fmt ./...
go test ./codegen/service ./http/codegen/... ./jsonrpc/codegen/... ./codegen/generator -count=1
go test ./... -run '^$'
git diff --check
```

All commands must pass.

### Task 9: Versioned protobuf descriptor plan and retained gRPC plan

**Files:**
- Replace: `grpc/codegen/plan.go`
- Replace: `grpc/codegen/service_data.go`
- Replace: `grpc/codegen/protobuf_catalog.go`
- Modify: `grpc/codegen/protobuf.go`
- Modify: `grpc/codegen/proto.go`
- Modify: `grpc/codegen/proto_hooks.go`
- Modify: `grpc/codegen/types.go`
- Modify: `grpc/codegen/client.go`
- Modify: `grpc/codegen/server.go`
- Modify: `grpc/codegen/client_cli.go`
- Modify: `grpc/codegen/example_cli.go`
- Modify: `grpc/codegen/example_server.go`
- Create: `grpc/codegen/protoc_names.go`
- Create: `grpc/codegen/protoc_names_test.go`
- Modify: `codegen/generator/plan.go`
- Modify: `codegen/generator/transport.go`
- Test: `grpc/codegen/plan_test.go`
- Test: `grpc/codegen/proto_test.go`
- Test: `grpc/codegen/protobuf_test.go`
- Test: `grpc/codegen/protobuf_transform_test.go`
- Test: `grpc/codegen/service_data_traversal_test.go`
- Test: `grpc/codegen/service_metadata_reference_test.go`
- Test: `grpc/codegen/streaming_test.go`
- Test: `codegen/generator/service_union_package_scope_test.go`

**Interfaces:**
- Consumes: exact retained `*service.Plan`
- Produces: retained protobuf descriptor plans and retained gRPC plans
- Produces: one explicit supported protoc/protoc-gen-go naming version and complete Go declaration families

- [ ] **Step 1: Capture the real protoc naming contract as RED tests**

Create descriptor fixtures for acronym and digit names, reserved words, nested
messages, enums, oneofs, services, streams, and explicit preferred names. Run
the supported real `protoc` and `protoc-gen-go` toolchain in a temporary module,
then compare every Goa-predicted package-level symbol with generated Go source.
Include message, enum/value, oneof interface/wrapper, client/server, and support
families. Add a test that rejects an unknown naming-version selector.

Run:

```bash
go test ./grpc/codegen -run 'TestProtocNameVersion|TestProtocDeclarationFamilies' -count=1
```

Expected: FAIL because protoc naming is currently approximated across helpers
and the catalog does not retain complete versioned families.

- [ ] **Step 2: Build one retained descriptor plan per protobuf package**

Represent `.proto` declarations and protoc-generated Go declarations as
separate typed records. Give each family canonical `NameDeclaration` records
for every Go symbol Goa references. Identity includes complete emitted schema,
ordered fields/oneofs, field numbers, validation, defaults, source provenance,
and role where these facts change output; explicit protobuf names remain
preferences.

Put the supported toolchain naming algorithm behind one explicit versioned
implementation. Delete scattered protoc CamelCase, oneof-wrapper, and service
name reconstruction after their callers consume family records.

- [ ] **Step 3: Build and render one retained gRPC plan**

Make gRPC `NewPlan` consume the exact `*service.Plan` and retained protobuf
descriptor plans. Collect messages, validators, conversions, native metadata,
streams, clients, servers, CLI, examples, imports, and output files before
freeze. Render `.proto` and Go files from the same records.

- [ ] **Step 4: Make validators and transforms context-independent**

Store the exact message, wrapper, validator, and conversion declarations in
render data. A transform context may select source and target records but may
not calculate their names. Add equal semantic ID/different origin cases,
same-origin/different role cases, reversed endpoint order, and one type reused
across unary and streaming roles. Compile and round-trip native metadata.

- [ ] **Step 5: Verify and commit Task 9**

Run:

```bash
go fmt ./...
go test ./codegen/service ./grpc/codegen/... ./codegen/generator -count=1
go test ./... -run '^$'
git diff --check
```

All commands must pass.

### Task 10: OpenAPI, examples, and selective lifecycle integration

**Files:**
- Modify: `codegen/generator/plan.go`
- Replace: `codegen/generator/openapi.go`
- Replace: `codegen/generator/example.go`
- Modify: `codegen/example/plan.go`
- Modify: `codegen/example/example_client.go`
- Modify: `codegen/example/example_server.go`
- Modify: `http/codegen/openapi.go`
- Modify: `http/codegen/openapi/v2/builder.go`
- Modify: `http/codegen/openapi/v2/files.go`
- Modify: `http/codegen/openapi/v2/openapi.go`
- Modify: `http/codegen/openapi/v3/builder.go`
- Modify: `http/codegen/openapi/v3/example.go`
- Modify: `http/codegen/openapi/v3/files.go`
- Modify: `http/codegen/openapi/v3/openapi.go`
- Modify: `codegen/generator/generation_test.go`
- Modify: `codegen/generator/purity_test.go`

**Interfaces:**
- Consumes: retained service and selected transport plans
- Produces: retained OpenAPI and example plans
- Produces: one core command plan with no render-time root or generation reconstruction

- [ ] **Step 1: Add selective-command and plan-identity REDs**

For `gen`, `example`, and focused test commands, assert each selected subsystem
is planned once, each renderer receives the exact retained pointer, unselected
subsystems allocate nothing, and the prepared root remains unchanged after the
plan boundary. Cover OpenAPI-only semantic example IDs separately from Go
declaration identity.

- [ ] **Step 2: Retain OpenAPI and example analysis**

Build typed OpenAPI plans from prepared expressions and typed example plans
from exact service/transport plans. The example plan owns its server
composition data; delete the process-global `codegen/example.Servers` map.
Retain one private JSON Schema registry per OpenAPI plan; delete the exported
mutable `Definitions` map and process-global definition-name state. The
returned specification owns its definition map and schema values, so a later
build cannot mutate it. Use the typed example identities established in Task
6. Collect every example and CLI package-level constructor, variable, and
helper through the owning package catalog.

- [ ] **Step 3: Make the core plan the only command execution model**

Have command factories construct one private-field `generator.Plan` containing
the exact selected subsystem plans. Core render dispatch reads those fields;
it does not call `NewPlan`, `NewServicesData`, `Generation.Roots`, or accept a
second generated module path. Remove all remaining generator adapters and
callback-shaped lifecycle tests.

- [ ] **Step 4: Prove purity, selection, repeated runs, and compilation**

Run each command twice and concurrently with different roots. Assert byte-
identical output per input, no cross-run state, no late declarations, and no
unselected files. Build disjoint example servers and OpenAPI specifications
sequentially and behind a start barrier; assert no server or schema from one
design appears in the other and the first returned result remains unchanged
after the second build. Run both concurrency tests with the race detector.
Compile full HTTP/gRPC/JSON-RPC examples and validate both OpenAPI versions.

- [ ] **Step 5: Verify and commit Task 10**

Run:

```bash
go fmt ./...
go test ./codegen/... ./http/codegen/... ./grpc/codegen/... ./jsonrpc/codegen/... \
  -skip '^TestMergeFilesPreservesSameLabelSections$' -count=1
go test ./... -run '^$'
git diff --check
```

The skipped merge regression remains Task 12 work.

### Task 11: Goa-ai retained plans and plugin migration

**Files:**
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/init.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/data.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/generate.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/generate_toolset_specs.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/generate_agent_files.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_build.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_helpers.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_materialize.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_misc.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_type_info.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_types.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_unions.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/ir/build.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/mcp/init.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/mcp/generate.go`
- Modify: `/Users/raphael/src/goa-ai/eval/codegen/codegen.go`
- Test: `/Users/raphael/src/goa-ai/codegen/agent/generate_test.go`
- Test: `/Users/raphael/src/goa-ai/codegen/agent/specs_builder_internal_test.go`
- Test: `/Users/raphael/src/goa-ai/codegen/agent/uniontest/union_names_test.go`
- Test: `/Users/raphael/src/goa-ai/codegen/agent/tests/golden_deep_nested_validations_test.go`
- Test: `/Users/raphael/src/goa-ai/codegen/mcp/contract_test.go`
- Test: `/Users/raphael/src/goa-ai/codegen/mcp/state_test.go`
- Test: `/Users/raphael/src/goa-ai/eval/codegen/codegen_test.go`

**Interfaces:**
- Consumes: generator plugin factories and exact `generator.Plan.Service(root)`
- Produces: retained agent specification, MCP, and eval plans
- Removes: temporary-root rendering, repeated service/spec reconstruction, shared spec scopes, and string union companion keys

- [ ] **Step 1: Add plugin and specification REDs**

Add a repeated/concurrent run test in one process, an MCP service whose package
overlaps a core service package, and an AURA-shaped large tool-spec package
with colliding validator/constructor preferences. Assert public tool specs and
HTTP transport specs own independent package plans and natural names. Assert
MCP render consumes the exact core service-plan pointer created after prepare.

- [ ] **Step 2: Attach all generated expressions during prepare**

Register fresh agent, MCP, and eval plugin factories. MCP prepare creates and
validates its service/types/JSON-RPC expressions and attaches them to a
canonical registered root before normalization and core planning. Do not create
a render-time temporary root or a second core service plan.

- [ ] **Step 3: Build one retained goa-ai plan per output package**

Build agent IR and tool specification data once during plugin planning. Split
public-spec and transport-spec package owners. Retain typed declaration records
for every generated type, validator, constructor, tool variable, and union
family. Delete repeated `NewServicesData`/IR/spec builders,
`UnionTypeHash`-based companion keys, shared `NameScope` use across packages,
and any emitted-name reconstruction.

- [ ] **Step 4: Render through exact core and plugin plans**

Agent, MCP, and eval render callbacks accept `*generator.Plan` and their
factory-owned retained plugin plan. MCP consumes `Plan.Service(root)` and emits
only plugin-owned adapters or modifications; core service/JSON-RPC plans emit
the attached service declarations once. No plugin accesses a plan registry or
looks up a “latest” analysis.

- [ ] **Step 5: Verify and commit Goa-ai**

Use a disposable module replacement or the repository's established local Goa
development workflow without committing an unrelated replacement. Run:

```bash
go fmt ./...
go test ./codegen/... ./eval/codegen/... -count=1
go test ./... -run '^$'
git diff --check
```

Generate and compile the AURA-shaped goa-ai fixture. Record the Goa commit it
requires, commit the goa-ai changes separately, and open or update the goa-ai
pull request.

### Task 12: Lossless merge, full regeneration, review, and publication

**Files:**
- Modify: `codegen/generator/generate.go`
- Modify: `codegen/generator/generate_merge_test.go`
- Regenerate: `/Users/raphael/src/aura/gen` only through AURA generation scripts
- Update: Goa and goa-ai pull request descriptions

**Interfaces:**
- Consumes: complete retained plans and package-owned declaration deduplication
- Produces: lossless same-path assembly and fully verified Goa, goa-ai, and AURA branches

- [x] **Step 1: Make same-path file assembly lossless**

Merge compatible headers and imports, then append every non-header section in
producer order. Never deduplicate by `SectionTemplate.Name`. Require all
same-path contributors to name the same canonical package identity; package
planning already owns declaration reuse and collision rejection.

Run:

```bash
go test ./codegen/generator -run TestMergeFilesPreservesSameLabelSections -count=1
```

This was completed with Task 6 because lossless file assembly is part of the
common run lifecycle. Both same-label bodies are preserved, all contributor
finalizers run in order, and incompatible headers or output paths fail before
rendering.

- [ ] **Step 2: Delete every superseded mechanism**

Require these production searches to return no hits:

```bash
rg -n 'Genfunc|renderOnly|NewServicesData|NewServicesDataForRoots|PlanKey|UnionTypeHash|unionRegistryKey|unionCompanionKey|userTypePkgs' \
  --glob '*.go'
rg -n 'var Generators|codegen\.RegisterPlugin|RunPluginsPlan|RunPluginsPrepare' \
  --glob '*.go'
```

Inspect every surviving `NewNameScope` in service, transport, example, and
goa-ai code. Retain it only when it owns lexical local names; no package-level
declaration or import may depend on it. Search every render function for root,
generated module path, and `Generation` inputs and remove remaining analysis
or output redirection.

- [ ] **Step 3: Verify Goa completely**

Run:

```bash
go fmt ./...
go test ./... -count=1
make lint
git diff --check
```

Expected: all pass with no skipped regression.

- [ ] **Step 4: Verify goa-ai completely**

Run in `/Users/raphael/src/goa-ai` against the final local Goa commit:

```bash
go fmt ./...
go test ./... -count=1
git diff --check
```

Expected: all pass.

- [ ] **Step 5: Regenerate and verify AURA from scratch**

Run in `/Users/raphael/src/aura`:

```bash
./scripts/gen goa
./scripts/gen
cd gen && go test ./... -count=1
```

Do not patch anything under `gen/`. Each generation command deletes and
recreates its owned output. Review the generated diff for unexpected public
name changes, then run the relevant AURA service/eval suites identified by
`docs/TROUBLESHOOT.md` and the original production-task reproduction.

- [ ] **Step 6: Run independent whole-branch reviews**

Review Goa and goa-ai against `codegen/ARCHITECTURE.md`. Require explicit
checks for every package-level symbol, exact/preferred collision policy,
output-path normalization, retained plan identity, protoc family/version
accuracy, repeated/concurrent runs, plugin root ownership, dead APIs, and
lossless merging. Fix every confirmed finding and repeat full verification.

- [ ] **Step 7: Publish clear pull requests**

Update the Goa PR in plain language: describe the invalid AURA validation
function reference, why separate analyses disagreed, the one-plan/one-name
rule, breaking plugin API, protoc proof, generated-source effects, and exact
verification commands. Update or create the goa-ai PR with its prepare-time MCP
attachment and retained spec-plan changes. Address every applicable GitHub
Copilot review comment before merge, push only verified commits, and keep both
PRs draft until dependent verification is green.

## Final completion proof

The work is complete only when all of these statements are true:

- one fresh factory instance owns each core generator and plugin in each run;
- roots never change after preparation;
- one typed core plan retains exact typed subsystem plans;
- every emitted package-level symbol has one frozen `NameDeclaration`;
- every declaration and reference consumes that same record;
- no renderer reconstructs service, wire, protobuf, OpenAPI, example, or plugin analysis;
- protobuf Go names match the explicit supported real toolchain family;
- repeated and concurrent runs are isolated;
- same-path file contributions are lossless;
- Goa, goa-ai, and freshly regenerated AURA all compile and pass their tests; and
- independent review finds no registry, fallback, compatibility path, duplicate owner, or dead transitional API.
