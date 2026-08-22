# Code Generation Architecture

This document defines how Goa turns one evaluated design into generated files.
It is the contract for generator authors: one run prepares the design, builds
one retained plan, freezes every emitted name, and renders that exact plan.

## Why this contract exists

The original AURA failure produced a reference to one validation function name
and a declaration with another name. The function and its caller were emitted
into the same generated Go package, but separate analyses allocated their names
from separately initialized scopes. The first ownership work fixed concrete
service, HTTP, JSON-RPC, and gRPC cases by adding a generation-wide package
catalog and transport-local catalogs. It also exposed the remaining design
mistake: planning still records only selected type families, then rendering
rebuilds service and transport data and allocates other package-level names.

For example, a render-time service scope can still choose names for endpoint
constructors, error constructors, validators, and stream helpers after the
generation has supposedly frozen. A second `NewServicesData` call can rebuild
the same logical wire model with a different traversal context. Both behaviors
let a declaration and its reference disagree.

The terminal contract is stronger and smaller: every package-level type,
function, constant, and variable has one declaration record owned by the
package that emits it. Every subsystem retains the analysis that created those
records. Rendering reads that analysis; it never reconstructs it.

## Run lifecycle

The `goa` command compiles and runs a temporary generator for one evaluated
design. One run follows this order:

1. Resolve the command and instantiate fresh core generator and plugin objects
   from immutable registered factories.
2. Evaluate and validate the design roots.
3. Run preparation plugins, which may add or change expressions. Construct one
   `codegen.Generation` with exclusive access to those evaluated roots. Its
   final preparation step wraps raw method attributes and records each exact
   generated wrapper. Concurrent runs use distinct expression graphs; the
   generation does not copy the graph or coordinate two runs mutating the same
   unprepared objects.
4. Record the prepared roots for mutation auditing. No later phase may mutate
   an expression root. The audit compares retained semantic state after every
   later callback; it does not claim that the expression graph was physically
   copied or made immutable.
5. Build one typed `generator.Plan`. It creates and retains the core service
   plan for each root, then the selected HTTP, gRPC, JSON-RPC, OpenAPI, and
   example plans that consume those exact service plans. Plugin planning
   receives the same typed plan.
6. Each subsystem completes collection, sorts declarations by stable typed
   identity, and declares every package-level symbol in its actual output
   package. The generation then freezes package names and import qualifiers.
7. Link each retained subsystem plan once. Linking converts recorded design
   facts and frozen declaration references into immutable template data. It
   cannot discover a declaration, reserve a name or import, mutate an
   expression, or create another analysis graph.
8. Core generators render their retained subsystem plans. Plugins render using
   the same `generator.Plan` and exact core service plans.
9. Merge contributions with the same canonical output path and render files.

Collection must be complete before freeze. Stable ordering makes preferred-name
suffixes independent of map iteration, traversal order, plugin registration
order, and process history. Freeze turns every declaration record into a
read-only value. Linking resolves those records exactly once before rendering;
it does not repeat collection or allocate another name. Render performs no
expression mutation, graph analysis, declaration discovery, name allocation,
or import allocation.

## Fresh run objects

Registration stores immutable factories, not mutable generator or plugin
instances. A factory is called once for each generation run:

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
```

These APIs belong to `codegen/generator`, which owns command orchestration.
The `codegen` package owns generated declarations and files; it does not own a
process-global plugin lifecycle. Core generator factories follow the same
fresh-instance rule.

Plugin names are non-empty and unique within one command across the First,
normal, and Last groups. Registration rejects unknown commands and stops after
the first run snapshots the registry. This makes alphabetical order a complete
ordering rule instead of relying on mutable registration sequence.

A factory may close over immutable configuration. Per-run roots, plans, files,
caches, and errors belong to the returned object. Concurrent and repeated
generation runs must not observe one another. The registry itself is immutable
while runs execute; tests install isolated registries rather than replacing a
public global `Generators` function.

## The retained core plan

`generator.Plan` is the typed value shared by core generators and plugins. Its
fields are private. It exposes the active `Generation` and the exact service
plan built for a registered root:

```go
func (p *Plan) Generation() *codegen.Generation
func (p *Plan) Service(root *expr.RootExpr) *service.Plan
```

Selected core subsystems are stored in typed fields, not in a generic map.
There is no `PlanKey`, string key, extension registry, or `any`-typed plan bag.
A plugin that needs core service declarations consumes `Plan.Service(root)`;
it may not call service analysis again or rebuild an equivalent plan from the
root.

Each subsystem has one retained planning entry point. Service files can be
shared by several Goa roots in one generation, so service planning accepts the
complete root batch:

```go
func service.NewPlans(generation *codegen.Generation, inputs ...service.PlanInput) ([]*service.Plan, error)
func service.NewPlan(root *expr.RootExpr, generation *codegen.Generation, examples *expr.ExampleGenerator) (*service.Plan, error)
```

`NewPlans` requires every service root owned by the generation exactly once.
It assigns relocated declaration files and external conversion methods across
the complete run before names freeze. Exact compiler copies with the same
retained Go layout share one declaration; copies that bind one declaration to
different fields, tags, pointer policies, union branches, or file facts are
rejected. `NewPlan` is only the strict single-root convenience form and rejects
a generation that contains more than one service root.

HTTP, gRPC, JSON-RPC, OpenAPI, and example generation use equivalent typed
constructors. A transport plan receives the exact `*service.Plan` for its root.
JSON-RPC may retain and reuse its HTTP plan because it emits HTTP codecs and
wire files, but it does not rebuild HTTP analysis. Render functions accept the
retained subsystem plan, not a `Generation`, generated module path, expression
root, or reconstructed `ServicesData`.

The plan stores collected design facts, immutable linked render data, and
canonical declaration pointers. It does not store callbacks that repeat
analysis. `NewServicesData`, `Genfunc`,
the replaceable `Generators` variable, `renderOnly`, and the callback plugin
registry are transition mechanisms to delete.

## Generated package ownership

The generation owns a package catalog keyed by the actual generated Go import
path. Each package record owns:

- one declaration namespace for every package-level type, function, constant,
  and variable;
- one import qualifier for every complete import path referenced by files in
  that package; and
- the canonical output directory that corresponds to its import path.

The common declaration record is `NameDeclaration`. It keeps its preferred and
final spellings private. `Name()` panics before freeze and returns the same
final spelling for the remainder of the run after freeze. Existing type,
union, branch, HTTP wire, protobuf, validator, and helper records contain or
reference `NameDeclaration`; they do not carry another independently mutable
name.

Package-level declarations include less obvious symbols: union discriminator
constants and constructors, endpoint constructors, error and
result constructors, validation functions, conversion functions, stream
interfaces and helpers, HTTP body constructors, protobuf oneof wrappers,
client and server constructors, and package variables emitted by templates.
Local variables, parameters, struct fields, and method names remain owned by
their lexical render scope because they cannot collide with package-level
declarations.

### Exact and preferred symbols

An exact symbol is part of an authored or external contract. Two distinct
exact declarations that normalize to the same Go identifier in one package are
rejected before rendering. Examples include two relocated authored types named
`foo-bar` and `foo_bar`, or two explicit external names that both require
`FooBar`.

A preferred symbol is generated from a semantic role. It may receive a stable
numeric suffix when another declaration already owns the preferred spelling.
Examples include a generated `ValidatePayload`, `NewValueText`, or protobuf
request message. The declaration's typed identity—not discovery order—decides
which record receives each spelling.

Exact declarations reserve first. Preferred declarations are sorted by stable
typed identity and allocated second. A subsystem must reject two distinct
identities whose ordering facts are equal; pointer addresses, expression
hashes, map order, and rendered text are not tie-breakers.

A companion whose spelling includes another declaration, such as
`Validate<Result>`, is registered as a dependent declaration before freeze.
The package freezes the base declaration first, then derives and reserves the
companion from that exact final name. Callers never rebuild the companion by
concatenating a separately resolved type string.

### Imports and output paths

Complete import path is the only import identity. Static-template requirements
have priority over generated-package preferences, which have priority over
design metadata preferences. References and `ImportSpec` values consume the
same frozen binding, while each file imports only the paths it uses.

The output planner canonicalizes both generated import paths and filesystem
paths before collection. If two different package identities normalize to the
same import path or output directory, planning rejects them. It does not let
one package win, merge their declarations, or add a suffix to a directory.
Multiple file contributions may share a canonical path only when they declare
the same package identity. The file merger appends every body section and runs
every file finalizer in contributor order. It rejects conflicting package
headers, import bindings, or keep-existing-file settings instead of choosing a
contributor.

Planning claims packages with the exact raw path supplied by the owner. The
claim is validated as a legal Go import path and preserves enough information
to reject two distinct raw paths that normalize to one import or output
directory. After freeze, generators use only canonical-path lookup; every
claim, including a repeated claim, is rejected because collection is closed.

## Expression identity and declaration identity

Expression identity answers a design question. Declaration identity answers
whether two generated package-level symbols are the same emitted contract.
They are deliberately separate.

`UserType.Origin()` identifies one authored declaration across exact compiler
copies. Recursion walkers use Origin only to detect a cycle in the current
graph traversal. A cycle set answers “have I entered this declaration on this
path?” It never proves that two emitted wire declarations, validators, or
helpers are interchangeable.

An emitted declaration identity contains every fact that changes its generated
source: owning package and role, source provenance, wire shape, validation,
defaults, views, pointer policy, ordered union branches, and protocol-specific
metadata as applicable. Equal semantic `ID()` values do not merge distinct
origins. Conversely, the same authored origin may produce distinct request and
response records when their emitted contracts differ.

`expr.Union.Hash()` remains expression identity. Typed code-generation
identities such as `UnionTypeID` describe emitted union families. Do not change
expression hashes, decorate string keys, or add general expression provenance
to coordinate code generation.

## Example value identity

Example configuration is immutable. One generation run creates an unanchored
`ExampleGenerator` for each prepared root, and every value draw must first
select a typed `ExampleIdentity`. The public identity constructors accept the
owning evaluated expression: a user type, method payload or result, method
error, HTTP request body, successful HTTP response, or HTTP error. Callers do
not join service names, response positions, or role labels into seed strings.

Structural descent is also typed. Object members, array elements, map keys, map
values, and union branches use distinct kind-tagged, length-framed segments.
For example, object member `"0"` cannot share a stream with array element zero,
and a result field named `NotFound` cannot share a stream with the method error
named `NotFound`. Length-constrained arrays and maps derive one stream per
element, key, and value instead of consuming a shared stream in traversal
order.

Named user types own their examples globally. Anonymous request, response, and
streaming shapes retain the explicit method or transport owner passed by their
caller. A zero-value generator intentionally disables OpenAPI examples; every
configured but unanchored value draw panics because it violates the identity
contract.

## Service plan

`service.Plan` owns every service and views package declaration for one root.
Its constructor collects service declarations, normalized method wrappers,
relocated authored types, projected view types, unions and their complete
families, endpoints, clients, constructors, validators, conversions,
interceptors, errors, stream types, and package variables. It also collects the
imports and exact output files those declarations require.

The plan retains one package-backed attributor for each service and views
package. HTTP, gRPC, JSON-RPC, example, and plugins use those attributors and
canonical declaration records. No consumer recreates a service `NameScope`,
calls `NewServicesData`, or reconstructs a name from a DSL spelling and package
alias.

When multiple prepared roots contribute to one generated package, the core
plan collects all their declarations before the package freezes and emits the
package once. A root not present in the Generation snapshot is rejected.

## HTTP and JSON-RPC plans

Each actual HTTP client or server output package owns a retained wire plan. It
collects complete detached request, response, WebSocket, SSE, error, union,
constructor, validator, codec, and helper declarations before freeze. The plan
keeps request and response policy in declaration identity, so one authored
origin can reuse a record only when the complete emitted wire contract agrees.

HTTP transforms enter the service and wire owners independently. Detached HTTP
bodies do not carry service package metadata. JSON-RPC consumes the exact HTTP
plan for the files and codecs it shares, then adds its own package declarations
to typed JSON-RPC plans. It does not create a second HTTP catalog.

Every validator and helper reference stores its canonical declaration. A call
site's traversal context may select which declaration it needs, but it never
selects or changes that declaration's name.

Reusable API- and service-level HTTP or gRPC error mappings select response
policy by error name. The endpoint method's effective error declaration owns
the service value that encoders and decoders carry. Planning compares pure,
fully finalized copies of the mapping and method attributes, including emitted
type shape, validations, defaults, struct metadata, and ordered union branches.
It accepts equivalent declarations and binds the mapping to the method record;
it rejects incompatible shadowing before rendering without mutating the
evaluated design.

## Protobuf and gRPC plans

The protobuf plan owns a descriptor model for each emitted `.proto` package and
the corresponding Go package produced by the supported `protoc` and
`protoc-gen-go` toolchain. Protobuf source declarations and protoc-generated Go
declarations are different, explicit families.

A declaration family records every package-level Go symbol Goa refers to,
including messages, nested messages, enums and enum values, oneof interfaces
and wrapper structs, service interfaces, client and server types, and version-
dependent support symbols. Preferred protobuf names do not become identity.
Field numbers, ordered fields and oneof branches, validation, defaults, source
provenance, and endpoint role are identity facts where they change output.

The protoc Go naming algorithm is selected by an explicit supported toolchain
version. One versioned implementation derives Go names for a descriptor family;
templates and transforms do not carry scattered approximations of protoc
CamelCase or oneof naming. Changing the supported compiler or plugin version
requires a new versioned naming contract and generated-module proof against the
real toolchain.

gRPC validators and conversions consume frozen descriptor-family records.
Validator identity is independent of the call site that first discovers it,
and conversion contexts cannot allocate a message, wrapper, or validator name.
Explicit metadata remains a detached native primitive wire contract and uses
the canonical service transform after parsing or before serialization.

## Plugins

Preparation is the only plugin phase allowed to mutate expression roots. A
plugin that adds a service, method, type, or transport mapping attaches it to a
registered root during preparation. Core normalization then observes it before
planning.

Plugin planning receives `*generator.Plan`. It may declare plugin-owned output
through the same Generation, and it consumes core declarations through the
exact retained service plan. Plugin rendering receives the same plan after
freeze. It may add files and sections, but it cannot create another root,
re-run service or transport analysis, reserve a name, or change an expression.

MCP generation therefore attaches its generated service expressions during
prepare and later consumes `Plan.Service(root)`. Agent tool specifications use
one retained typed specification plan for each output package; public specs and
transport specs are distinct packages with distinct declaration owners. No
plugin coordinates through a process-global map, a latest result, a decorated
hash, `PlanKey`, or render order.

## File assembly

`SectionTemplate.Name` labels a section for diagnostics. It is not declaration
identity. Package plans remove identical declaration contributions before they
become sections. The file merger combines imports and appends every non-header
section in producer order, even when diagnostic labels match. Conflicting
declarations remain visible and fail during planning or Go compilation instead
of disappearing silently.

## Compatibility and operations

This architecture intentionally breaks external generators and plugins that
register callback instances, replace `Generators`, call `NewServicesData`, or
render from roots and generated module paths. They must register factories,
retain typed plans, and render those plans.

Generated Go names may change where prior suffix ownership depended on
traversal or reconstruction. Goa, goa-ai, and applications must regenerate
together. There is no runtime fallback, persisted-data migration, or staged
dual mode. A normalized output-path collision now fails during planning rather
than overwriting or combining unrelated packages.

Fresh factories make repeated and concurrent generation independent. The main
operational risks are an uncollected template symbol, an incomplete emitted
identity, or an inaccurate protoc family name. Focused catalog tests, reversed-
order tests, concurrent-run tests, real generated-module compilation, the
supported protoc toolchain, goa-ai generation, and full AURA regeneration are
the required proof.

## Review gate

Before changing generation lifecycle, names, roots, transports, plugins, or
file merging, trace one representative declaration through:

1. the prepared root;
2. the retained service plan;
3. the selected transport or plugin plan;
4. the owning generated package and `NameDeclaration`;
5. stable collection and freeze;
6. every declaration and reference rendered from that record;
7. plugin contributions; and
8. the final merged file and compiled generated module.

Also prove a valid counterexample at the next wider lifetime: two declarations
with one semantic ID but different origins, one origin with different request
and response wire contracts, two packages with the same basename, repeated
runs in one process, and concurrent runs. A service-only render test or a
source-text assertion is insufficient when the failure can appear in a
transport, protoc-generated family, plugin, or merged output.
