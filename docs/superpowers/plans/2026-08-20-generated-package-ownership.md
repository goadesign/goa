# Generated Package Ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every generated service-type declaration and reference use one name frozen by the generation that produces the output.

**Architecture:** After prepare plugins and root normalization, one `codegen.Generation` runs all selected core generators and plugins through plan, freeze, and render phases. Its generated-package catalog owns package scopes and canonical user-type and union records. Service, HTTP, gRPC, JSON-RPC, standalone callers, and goa-ai plugins resolve names through those records; package owners render each declaration once.

**Tech Stack:** Go 1.25, Goa evaluation and code generation, goa-ai plugins, `testify/require`

**Spec:** `codegen/ARCHITECTURE.md`

## Global Constraints

- Never edit generated output; regenerate it from the owning design.
- Keep `expr.Union.Hash()` unchanged and use a typed code-generation identity for emitted unions.
- Reject relocated declared user types whose names become the same Go identifier in one output package.
- Definitions and references must consume the same frozen package-owned declaration record.
- Rendering cannot add declarations; standalone generation runs the same plan, freeze, and render phases.
- Do not use decorated strings, synthetic map keys, process-global registries, fallbacks, or traversal-order heuristics to coordinate ownership.
- `SectionTemplate.Name` is diagnostic metadata, not declaration identity.
- Every exported construct needs GoDoc; non-trivial files need a concrete header comment.

---

### Task 1: Executable failure contracts

**Files:**
- Modify: `codegen/generator/service_union_package_scope_test.go`
- Modify: `codegen/generator/generate_merge_test.go`
- Create: `codegen/generated_types_test.go`

**Interfaces:**
- Consumes: current full generator, HTTP/gRPC DSL, and same-path file merging
- Produces: red tests for relocated nested-union references, relocated declared-name collisions, and same-label section preservation

- [ ] **Step 1: Add the real generated-module regression**

Use one evaluated design root with two services. Put distinct user types in the
same `struct:pkg:path` package and give each a different nested union whose
natural name is `Value`. Enable HTTP and gRPC, generate the module, and run `go
test ./...` inside it.

- [ ] **Step 2: Prove the nested-union regression fails for the intended reason**

Run:

```bash
go test ./codegen/generator -run TestRelocatedUnionPackageNamesCompile -count=1
```

Expected: FAIL with a generated Go reference to a union name different from the
declaration in the shared package.

- [ ] **Step 3: Add the collision and merge regressions**

The collision test plans `foo-bar` and `foo_bar` into package `types` and
expects an error naming both inputs and `FooBar`. The merge test contributes two
different sections with the same `SectionTemplate.Name` and expects both
rendered bodies to remain.

- [ ] **Step 4: Prove both contracts fail before implementation**

Run:

```bash
go test ./codegen ./codegen/generator -run 'TestGeneratedTypesRejectRelocatedNameCollision|TestMergeFilesPreservesSameLabelSections' -count=1
```

Expected: FAIL because the generation-owned type catalog does not exist and the
merger drops the second same-label section.

### Task 2: Typed emitted-union identity

**Files:**
- Modify: `codegen/union.go`
- Modify: `codegen/scope.go`
- Modify: `codegen/scope_test.go`

**Interfaces:**
- Consumes: `expr.Union`, `NameScope.HashedUnique(Hasher, string, ...string)`
- Produces: `type UnionTypeID string`, `func NewUnionTypeID(*expr.Union) UnionTypeID`, and generic `Hasher.Hash()` behavior

- [ ] **Step 1: Add a focused `HashedUnique` contract test**

Use a custom `Hasher` and assert that `HashedUnique` keys only on its exact
`Hash()` result. Keep the existing tests that distinguish emitted unions by
wire keys, branch order, branch Go shape, and relocated package.

- [ ] **Step 2: Run the focused test and record the hidden special-case failure**

Run:

```bash
go test ./codegen -run 'TestNameScope_HashedUnique|TestUnionTypeID' -count=1
```

- [ ] **Step 3: Introduce the typed identity and restore `HashedUnique`**

Use this public contract:

```go
type UnionTypeID string

func NewUnionTypeID(union *expr.Union) UnionTypeID
```

`NewUnionTypeID` contains the current emitted-definition hashing algorithm.
`NameScope.HashedUnique` calls `key.Hash()` directly. `GoFullTypeName` may look
up an already planned union through an explicit emitted-union key, but the
generic scope API must not reinterpret arbitrary hashers.

- [ ] **Step 4: Run all scope and union tests**

Run:

```bash
go test ./codegen -run 'TestNameScope|TestUnionType' -count=1
```

Expected: PASS.

### Task 3: Generation plan, freeze, and render contract

**Files:**
- Create: `codegen/generation.go`
- Create: `codegen/generated_types.go`
- Modify: `codegen/plugin.go`
- Modify: `codegen/plugin_test.go`
- Modify: `codegen/generator/generators.go`
- Modify: `codegen/generator/generate.go`
- Create: `codegen/generator/generation_test.go`

**Interfaces:**
- Consumes: normalized `[]eval.Root`, `codegen.NameScope`, `codegen.UnionTypeID`
- Produces: `Generation`, generated-package records, plan-aware core generator and plugin APIs

- [ ] **Step 1: Add phase and declaration tests**

Test user-type idempotency, union idempotency, different same-base unions,
exact relocated-name rejection, lookup before and after freeze, declaration
after freeze rejection, and isolation between two standalone generations.

- [ ] **Step 2: Implement generation-owned package records**

Use these public contracts:

```go
type Generation struct {
    GenPkg string
    Roots  []eval.Root
}

func NewGeneration(genpkg string, roots []eval.Root) *Generation
func (g *Generation) GeneratedPackage(path string) *GeneratedPackage
func (g *Generation) Freeze() error

type GeneratedPackage struct{}

func (p *GeneratedPackage) DeclareUserType(userType expr.UserType) (*TypeDeclaration, error)
func (p *GeneratedPackage) DeclareUnion(union *expr.Union) (*TypeDeclaration, error)
func (p *GeneratedPackage) UserType(userType expr.UserType) (*TypeDeclaration, error)
func (p *GeneratedPackage) Union(union *expr.Union) (*TypeDeclaration, error)
func (p *GeneratedPackage) Scope() *NameScope

type TypeDeclaration struct {
    Name        string
    PackagePath string
}
```

Declaration methods allocate only during planning. Lookup methods never
allocate. `Freeze` makes every package immutable. User types reserve the exact
`Goify(Name(), true)` name and report collisions; unions allocate through their
typed emitted identity.

- [ ] **Step 3: Change core and plugin lifecycle APIs**

Use these contracts:

```go
type PlanFunc func(*Generation) error
type GenerateFunc func(*Generation, []*File) ([]*File, error)

type Genfunc struct {
    Plan     codegen.PlanFunc
    Generate func(*codegen.Generation) ([]*codegen.File, error)
}
```

`RegisterPlugin`, `RegisterPluginFirst`, and `RegisterPluginLast` accept
prepare, plan, and generate functions. `Generate` runs prepare, normalization,
every core/plugin plan, `Freeze`, every core render, then every plugin render.
No render callback may declare a new type.

- [ ] **Step 4: Run lifecycle tests**

Run:

```bash
go test ./codegen ./codegen/generator -run 'TestGeneration|TestGeneratedPackage|TestRegisterPlugin|TestGeneratePhases' -count=1
```

Expected: PASS.

### Task 4: Package-owned service analysis and emission

**Files:**
- Create: `codegen/service/generated_package.go`
- Modify: `codegen/service/service_data.go`
- Modify: `codegen/service/service.go`
- Modify: `codegen/service/convert.go`
- Modify: `codegen/service/views.go`
- Modify: `codegen/service/service_test.go`
- Modify: `codegen/service/service_data_union_order_test.go`
- Modify: `codegen/generator/service.go`
- Modify: `codegen/generator/example.go`
- Modify: `codegen/generator/openapi.go`

**Interfaces:**
- Consumes: frozen or planning `*codegen.Generation`, `*codegen.TypeDeclaration`
- Produces: `NewServicesData(*expr.RootExpr, *codegen.Generation) (*ServicesData, error)` and root-level package-owned service files

- [ ] **Step 1: Add package analysis and emission tests**

Test that all services in one root bind to the same declaration records,
identical unions emit once, different same-base unions receive distinct frozen
names, relocated user types emit once at their metadata paths, and each owning
package emits one `unions.go`.

- [ ] **Step 2: Replace local package priming with frozen declarations**

Delete `NewServicesDataForRoots`, `packageScopes`, `serviceNameScopes`,
`unionCompanionKey`, and every decorated union key. During planning,
`NewServicesData` declares every relocated user type before unions. During
rendering, it looks up the same records. `UserTypeData` and `UnionTypeData`
retain their `*codegen.TypeDeclaration`; `buildUnionTypeData` allocates the kind
name once from the owning package scope and stores it in the union render data.

- [ ] **Step 3: Make the package owner render all service types**

Change the public renderer to:

```go
func Files(genpkg string, services *ServicesData) []*codegen.File
```

It renders service-local files, each relocated user type at its configured
file, and one sorted `unions.go` per package. Remove `userTypePkgs`, `~union:`,
and `unionRegistryKey`. `ConvertFiles` uses the owning package scope rather than
a fresh one.

- [ ] **Step 4: Migrate core service, example, and OpenAPI generators**

Their plan callback analyzes each design root with the active generation. Their
render callback repeats analysis against frozen records and propagates errors.
The Service renderer calls the root-level `service.Files` once.

- [ ] **Step 5: Run service and generator tests**

Run:

```bash
go test ./codegen/service ./codegen/generator -count=1
```

Expected: PASS.

### Task 5: Frozen service names in HTTP, gRPC, and JSON-RPC

**Files:**
- Modify: `codegen/transformer.go`
- Modify: `http/codegen/service_data.go`
- Modify: `http/codegen/websocket.go`
- Modify: `http/codegen/sse.go`
- Modify: `http/codegen/client.go`
- Modify: `http/codegen/server.go`
- Modify: `grpc/codegen/service_data.go`
- Modify: `grpc/codegen/types.go`
- Modify: `grpc/codegen/server.go`
- Modify: `codegen/generator/transport.go`
- Test: `codegen/generator/service_union_package_scope_test.go`

**Interfaces:**
- Consumes: service data bound to frozen `TypeDeclaration` records and package scopes
- Produces: package-aware `Attributor` contexts for every recursive service-type transform

- [ ] **Step 1: Add focused transport reference assertions**

For the two-service nested-union design, assert that HTTP and gRPC conversion
helpers refer to the exact union names declared in the relocated package. Keep
transport wire-type scopes independent.

- [ ] **Step 2: Make attribute contexts carry package ownership**

Add the generated package path or frozen declaration resolver required for an
`Attributor` to select the enclosing service package while recursion enters a
relocated user type. `AttributeContext.Dup` must preserve it, and helper
generation must update it when `struct:pkg:path` changes the enclosing package.

- [ ] **Step 3: Replace direct service-scope recomputation**

HTTP, WebSocket, SSE, client/server callbacks, gRPC conversions, gRPC
`fullTypeName`, and transport generator setup resolve service types through the
frozen service attributor or existing canonical method declaration. `sd.Scope`
continues to name only HTTP/protobuf wire declarations.

- [ ] **Step 4: Run the generated-module regression**

Run:

```bash
go test ./codegen/generator -run TestRelocatedUnionPackageNamesCompile -count=1
```

Expected: PASS with HTTP and gRPC enabled.

- [ ] **Step 5: Run all core codegen tests**

Run:

```bash
go test ./codegen/... ./http/codegen/... ./grpc/codegen/... ./jsonrpc/codegen/... -count=1
```

Expected: PASS.

### Task 6: Goa-ai plugin participation

**Files:**
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/init.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/data.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/ir/build.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/mcp/init.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/mcp/generate.go`
- Modify: `/Users/raphael/src/goa-ai/eval/codegen/codegen.go`
- Create: `/Users/raphael/src/goa-ai/codegen/mcp/generate_test.go`

**Interfaces:**
- Consumes: Goa's plan-aware plugin API, active `*codegen.Generation`, root-level `service.Files`
- Produces: agent, MCP, and eval plugins that plan generated service types before freeze and render from the same records

- [ ] **Step 1: Add MCP planning tests**

Create an MCP temporary service whose relocated type package overlaps a core
service package. Assert that identical declarations share a record and
incompatible declared names or shapes return an error during planning, before
any core or MCP file renders.

- [ ] **Step 2: Migrate plugin registration and builders**

Add plan callbacks to agent, MCP, and eval registration. Builders that need
service data accept the active generation and propagate `NewServicesData`
errors. MCP plans its temporary root into the active context and calls
root-level `service.Files` during render. Delete its `userTypePkgs` map.

- [ ] **Step 3: Run goa-ai tests against local Goa**

Run in `/Users/raphael/src/goa-ai` with its existing local Goa replacement:

```bash
go test ./codegen/... ./eval/codegen/... -count=1
```

Expected: PASS.

### Task 7: File assembly, full regeneration, and publication

**Files:**
- Modify: `codegen/generator/generate.go`
- Modify: `codegen/generator/generate_merge_test.go`

**Interfaces:**
- Consumes: package-owned emission and same-path file contributions
- Produces: lossless same-path merging, verified Goa/goa-ai/AURA, and updated pull requests

- [ ] **Step 1: Make same-path merging lossless**

Merge header imports and append every non-header section in generator order.
Do not deduplicate by `SectionTemplate.Name`. Package owners already remove
identical type declarations; conflicting output must remain visible as an
explicit generation or Go compilation failure.

- [ ] **Step 2: Remove obsolete mechanisms**

Confirm these searches return no production hits:

```bash
rg -n 'NewServicesDataForRoots|~union:|unionRegistryKey|unionCompanionKey|userTypePkgs|scopedTypeHash' --glob '*.go'
rg -n 'NewNameScope\(\)' codegen/service http/codegen grpc/codegen --glob '*.go'
```

Inspect every remaining fresh scope and retain it only for a package whose
declarations it exclusively owns.

- [ ] **Step 3: Verify Goa**

Run:

```bash
go fmt ./...
go test ./... -count=1
make lint
```

Expected: all commands pass.

- [ ] **Step 4: Regenerate and verify AURA from scratch**

Run in `/Users/raphael/src/aura`:

```bash
./scripts/gen goa
./scripts/gen
cd gen && go test ./... -count=1
```

Never patch files under `gen/`; each generation command owns deletion and
recreation.

- [ ] **Step 5: Review and publish**

Run independent whole-branch reviews in Goa and goa-ai, address every confirmed
finding, and update the Goa PR with the concrete failure, generation ownership
rule, breaking API, generated-source change, rejected ambiguous design, plugin
behavior, and exact verification commands. Push only after all proof passes.
