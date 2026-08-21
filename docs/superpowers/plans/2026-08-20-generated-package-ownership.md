# Generated Package Ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every relocated Goa declaration and reference use one name assigned by its generated Go package.

**Architecture:** `service.ServicesData` owns a generation-lifetime catalog keyed by generated import path. Each package record owns its `codegen.NameScope`, relocated user types, and canonical union rendering records; package-aware attribute scopes route service and transport transforms to those records. `service.Files` renders the complete analyzed root, including one `unions.go` per shared package.

**Tech Stack:** Go 1.25, Goa expression evaluation, Goa service/HTTP/gRPC code generation, `testify/require`

**Spec:** `codegen/ARCHITECTURE.md`

## Global Constraints

- Never edit generated output; regenerate it from the owning design.
- Keep `expr.Union.Hash()` unchanged and use a typed code-generation identity for emitted unions.
- Reject relocated declared user types whose names become the same Go identifier in one output package.
- Definitions and references must consume the same package-owned declaration record.
- Do not use decorated strings, synthetic map keys, global registries, fallbacks, or traversal-order heuristics to coordinate ownership.
- Every exported construct needs GoDoc; non-trivial files need a concrete header comment.

---

### Task 1: Regression contracts

**Files:**
- Modify: `codegen/generator/service_union_package_scope_test.go`
- Create: `codegen/service/generated_package_test.go`

**Interfaces:**
- Consumes: existing `generator.Generate`, HTTP/gRPC DSL, and `codegen.Goify`
- Produces: a compile regression for nested relocated unions and a validation regression for colliding relocated declared names

- [ ] **Step 1: Write the generated-module regression**

Create one real design root with two services. Put distinct user types in the
same `struct:pkg:path` package and give each a different nested union whose
natural name is `Value`. Enable HTTP and gRPC and run the complete generated
module's tests.

- [ ] **Step 2: Run the generated-module regression and record the failure**

Run:

```bash
go test ./codegen/generator -run TestRelocatedUnionPackageNamesCompile -count=1
```

Expected: FAIL because a transport reference names a different union than the
shared package declares.

- [ ] **Step 3: Write the collision regression**

Construct relocated declared types named `foo-bar` and `foo_bar` in the same
package and call `NewServicesData`:

```go
_, err := NewServicesData(root)
require.ErrorContains(t, err, `"foo-bar" and "foo_bar" both generate Go type "FooBar" in package "types"`)
```

- [ ] **Step 4: Run the collision regression and record the failure**

Run:

```bash
go test ./codegen/service -run TestGeneratedPackagesRejectUserTypeNameCollision -count=1
```

Expected: FAIL because `NewServicesData` does not yet return an error.

### Task 2: Explicit code-generation identity

**Files:**
- Modify: `codegen/union.go`
- Modify: `codegen/scope.go`
- Modify: `codegen/scope_test.go`

**Interfaces:**
- Consumes: `expr.Union`, `NameScope.HashedUnique(Hasher, string, ...string)`
- Produces: `type UnionTypeID string`, `func NewUnionTypeID(*expr.Union) UnionTypeID`, and an explicit private `Hasher` used at union-generation sites

- [ ] **Step 1: Preserve the public scope contract in a failing test**

Add a test whose custom `Hasher.Hash()` value must be the exact key used by
`HashedUnique`, including when the concrete value wraps a union.

- [ ] **Step 2: Run the scope test and confirm the hidden union special case fails**

Run:

```bash
go test ./codegen -run 'TestNameScope_HashedUnique|TestUnionTypeID' -count=1
```

- [ ] **Step 3: Introduce the typed identity and remove the hidden special case**

Use these contracts:

```go
type UnionTypeID string

func NewUnionTypeID(union *expr.Union) UnionTypeID

type unionNameKey struct {
    id   UnionTypeID
    role unionNameRole
}

func (k unionNameKey) Hash() string
```

`NameScope.HashedUnique` calls `key.Hash()` directly. Union declaration sites
pass an explicit union name key. `expr.Union.Hash()` remains unchanged.

- [ ] **Step 4: Run the focused identity tests**

Run:

```bash
go test ./codegen -run 'TestNameScope|TestUnionType' -count=1
```

Expected: PASS.

### Task 3: Package-owned service analysis

**Files:**
- Create: `codegen/service/generated_package.go`
- Modify: `codegen/service/service_data.go`
- Modify: `codegen/service/service_data_union_order_test.go`
- Modify: `codegen/service/views.go`

**Interfaces:**
- Consumes: `codegen.NameScope`, `codegen.UnionTypeID`, `UserTypeData`, `UnionTypeData`
- Produces: private `generatedPackages`, `generatedPackage`, `typeDeclaration`, and `packageAttributeScope`; `NewServicesData(*expr.RootExpr) (*ServicesData, error)`

- [ ] **Step 1: Add focused package-catalog tests**

Test that one package returns the same declaration record for the same union
identity, distinct records for different shapes, deterministic names after
declared user names are reserved, and an error for duplicate declared Go names.

- [ ] **Step 2: Run the focused tests and confirm the catalog is absent**

Run:

```bash
go test ./codegen/service -run 'TestGeneratedPackage|TestUnionOrder' -count=1
```

- [ ] **Step 3: Implement the package catalog and package-aware attributor**

Use these private contracts:

```go
type generatedPackages struct {
    packages map[string]*generatedPackage
}

type generatedPackage struct {
    path      string
    scope     *codegen.NameScope
    userTypes map[string]*typeDeclaration
    unions    map[codegen.UnionTypeID]*typeDeclaration
}

type typeDeclaration struct {
    name     string
    userType *UserTypeData
    union    *UnionTypeData
}
```

`packageAttributeScope` implements `codegen.Attributor`. It selects a relocated
user type's package from `struct:pkg:path`, a union's package from the enclosing
declaration path, and the service scope otherwise. It delegates name, reference,
and field formatting to `codegen.NewAttributeScope` using the selected
`NameScope`.

- [ ] **Step 4: Make analysis eager and single-root**

Delete `NewServicesDataForRoots`, `serviceNameScopes`, package priming, and the
union companion sentinel. Reserve every relocated declared user type before
registering unions, analyze services in declaration order, and make `Get` a
lookup over the completed map.

- [ ] **Step 5: Run service analysis tests**

Run:

```bash
go test ./codegen/service -count=1
```

Expected: PASS.

### Task 4: One owner renders shared packages

**Files:**
- Modify: `codegen/service/service.go`
- Modify: `codegen/service/service_test.go`
- Modify: `codegen/generator/service.go`
- Modify: `codegen/generator/example.go`
- Modify: `codegen/generator/transport.go`

**Interfaces:**
- Consumes: completed `service.ServicesData` package catalog
- Produces: `func Files(genpkg string, services *ServicesData) []*codegen.File`, including one `unions.go` per relocated package

- [ ] **Step 1: Update rendering tests to assert package ownership**

Assert that relocated user-type files contain only their declared types, the
package has exactly one `unions.go`, and repeated service references do not
duplicate declarations.

- [ ] **Step 2: Run focused rendering tests and record the failure**

Run:

```bash
go test ./codegen/service -run 'TestFiles.*Union|TestFiles.*Package' -count=1
```

- [ ] **Step 3: Replace per-service rendering and external deduplication**

Change `Files` to render the whole analyzed root. Delete `userTypePkgs`,
`~union:`, `unionRegistryKey`, and `unionCompanionKey`. Render each relocated
package from its catalog, sorting package paths, user-type file paths, and union
identities for stable output.

- [ ] **Step 4: Restore generator callers to one real root**

`Service`, `Example`, and `Transport` each call `NewServicesData(root)` and
propagate its error. The service generator calls `service.Files(genpkg,
services)` once per root. Remove all cross-root service-data maps.

- [ ] **Step 5: Run service and generator tests**

Run:

```bash
go test ./codegen/service ./codegen/generator -count=1
```

Expected: PASS.

### Task 5: Transport and conversion references

**Files:**
- Modify: `codegen/service/convert.go`
- Modify: `http/codegen/service_data.go`
- Modify: `http/codegen/websocket.go`
- Modify: `http/codegen/sse.go`
- Modify: `grpc/codegen/service_data.go`
- Test: `codegen/generator/service_union_package_scope_test.go`

**Interfaces:**
- Consumes: package-aware `codegen.Attributor` values from `ServicesData`
- Produces: every recursive service-type transform uses the package that owns the enclosing generated declaration

- [ ] **Step 1: Route conversion generation through the package catalog**

Replace fresh `NameScope` instances used for relocated conversion files with
package-aware contexts from `ServicesData`. The current package path is the
relocated user's `RelImportPath`.

- [ ] **Step 2: Route HTTP service contexts through the package catalog**

At each HTTP conversion, validation, SSE, and WebSocket site, derive the
enclosing service type's location and request the service attributor from
`ServicesData`. Keep `sd.Scope` for HTTP wire declarations only.

- [ ] **Step 3: Route gRPC service contexts through the package catalog**

Use the same service attributor for payload, result, error, and streaming
transforms. Keep protobuf scopes responsible only for protobuf declarations.

- [ ] **Step 4: Run the real generated-module regression**

Run:

```bash
go test ./codegen/generator -run TestRelocatedUnionPackageNamesCompile -count=1
```

Expected: PASS with HTTP and gRPC enabled.

- [ ] **Step 5: Run core generator tests**

Run:

```bash
go test ./codegen/... ./http/codegen/... ./grpc/codegen/... ./jsonrpc/codegen/... -count=1
```

Expected: PASS.

### Task 6: Plugins, full proof, and publication

**Files:**
- Modify: `/Users/raphael/src/goa-ai/codegen/agent/data.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/ir/build.go`
- Modify: `/Users/raphael/src/goa-ai/codegen/mcp/generate.go`
- Test: `/Users/raphael/src/goa-ai/codegen/mcp/generate_test.go`
- Modify: `codegen/ARCHITECTURE.md` if implementation evidence changes its contract
- Modify: `AGENTS.md` if review finds a missing durable gate

**Interfaces:**
- Consumes: generated-package ownership contract
- Produces: verified plugin isolation or shared-context participation, regenerated AURA, and an updated pull request

- [ ] **Step 1: Audit plugin callers**

Update the goa-ai agent and intermediate-representation builders to propagate
the `NewServicesData` error. Update the MCP generator to call the whole-root
`service.Files` contract for its temporary root. Prove in
`codegen/mcp/generate_test.go` that the temporary root emits only beneath its
MCP service package and cannot contribute declarations to a core service's
relocated package.

- [ ] **Step 2: Run Goa verification**

Run:

```bash
go fmt ./...
go test ./... -count=1
make lint
```

Expected: all commands pass.

- [ ] **Step 3: Regenerate and test AURA from scratch**

Run in `/Users/raphael/src/aura`:

```bash
./scripts/gen goa
cd gen && go test ./... -count=1
```

If AURA's full `./scripts/gen` is required by the repository workflow, run it
before the generated-module test. Never patch files under `gen/`.

- [ ] **Step 4: Remove obsolete mechanisms**

Search for and remove `NewServicesDataForRoots`, `~union:`,
`unionRegistryKey`, `unionCompanionKey`, `userTypePkgs`, union-specific behavior
inside `NameScope.HashedUnique`, and fresh scopes used for relocated service
types. Confirm `SectionTemplate.Name` is not used as declaration identity.

- [ ] **Step 5: Review and publish the pull request**

Run an independent code review, address each confirmed finding, and update the
PR with a plain-language explanation of the failure, package ownership rule,
generated-source change, newly rejected ambiguous design, and exact proof.
Push only after every verification command passes.
