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

1. Resolve the command and create fresh core generator and factory-plugin
   objects. Copy plugins registered through the released callback API into the
   same run.
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
   example plans that consume those exact service plans. Factory plugins
   receive the same plan when they declare their output.
6. Each subsystem completes collection, sorts declarations by stable typed
   identity, and declares every package-level symbol in its actual output
   package. The generation then freezes package names and import qualifiers.
7. Link each retained subsystem plan once. Linking converts recorded design
   facts and final declaration references into template data. All generated
   names are fixed, but later plugin callbacks may still make the permitted
   edits to ordinary section values. Linking cannot discover a declaration,
   reserve a name or import, mutate an expression, or create another analysis
   graph.
8. Core generators and factory plugins render from the same `generator.Plan`
   and exact core service plans. Released callbacks receive their original
   generated package, roots, and current file list instead.
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

These factory APIs belong to `codegen/generator`, which runs generation
commands. Goa also keeps the released four-argument registration functions in
`codegen`. Those functions store callback pairs in an internal registry; the
generator copies them into the same run when generation starts. Plugin authors
cannot inspect the registry or run callbacks themselves. Core generator
factories follow the same fresh-instance rule.

Factory plugin names are non-empty and unique within one command across the
First, normal, and Last groups. Released callback registrations may repeat a
name, as Goa v3 allowed; equal names keep registration order. Both APIs reject
unknown commands and stop accepting registrations when generation first
starts.

A factory may close over immutable configuration. Per-run roots, plans, files,
caches, and errors belong to the returned object. Concurrent and repeated
generation runs must not observe one another. The factory registry is immutable
while runs execute, and tests install isolated registries. The released
`Generators` variable remains replaceable for compatibility. Callers configure
it before starting concurrent runs; each run reads its function list once and
turns that list into fresh internal generators.

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

The plan stores collected design facts, linked render data, and final
declaration pointers. It does not store callbacks that repeat analysis.
`NewServicesData`, `renderOnly`, and the released functions that ran plugin
callbacks are old entry points that the retained plan replaces. The released
`Genfunc` type, replaceable `Generators` variable, and core generator functions
remain available. Goa adapts built-in functions to this plan and runs external
functions after every generated name is final.

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

Package-level declarations include less obvious symbols: constants that record
a union's selected branch, union constructors, endpoint constructors, error and
result constructors, validation functions, conversion functions, stream
interfaces and helpers, HTTP body constructors, protobuf oneof wrappers,
client and server constructors, and package variables emitted by templates.
Local variables, parameters, struct fields, and method names remain owned by
their lexical render scope because they cannot collide with package-level
declarations.

Service package paths are assigned once across every prepared design root in a
generation run. Equal authored service names share one generated package even
when they come from different APIs. Different names that reduce to the same Go
package name receive stable numeric suffixes. Natural package names are
reserved first, so an authored service whose normal package is `read_value2`
keeps that path while another collision advances to `read_value3`. The linked
service data exposes the final directory as `PathName`. HTTP, gRPC, JSON-RPC,
MCP, examples, and command-line generators read that retained path and its
saved import; they never rebuild a package path from the service name.

Generated command-line parsers also plan names at the scope that Go checks:
the complete `ParseEndpoint` function. The parser reserves imported package
names first, then parameters and fixed local variables, then each command's
flag variables and conversion variables. Templates receive those exact names
and write the selected conversion directly. They do not search generated text,
replace variable names, or decide a conversion from a type name at runtime.

### Exact and preferred symbols

An exact symbol is part of an authored or external contract. Two distinct
exact declarations that normalize to the same Go identifier in one package are
rejected before rendering. Examples include two relocated authored types named
`foo-bar` and `foo_bar`, or two explicit external names that both require
`FooBar`. A Goa `OneOf` also owns one exact public family: the union type, its
kind type, branch constants, constructors, and compiler-created branch types.
For a union named `Value` with a `text` branch, these include `Value`,
`ValueKind`, `ValueKindText`, `NewValueText`, and `ValueBranchText`.
HTTP transport copies add the exact body role to the union name:
`ValueRequestBody`, `ValueStreamingBody`, and `ValueResponseBody`. A response
view named `detailed` uses `ValueDetailedResponseBody`. Copies of one authored
union in the same role and view reuse that family. If those copies would emit
different definitions, generation fails and the design must use separate
`OneOf` declarations.

A preferred symbol is generated from a semantic role. It may receive a stable
numeric suffix when another declaration already owns the preferred spelling.
Examples include a generated `ValidatePayload` or protobuf request message.
The declaration's typed identity—not discovery order—decides which record
receives each spelling.

Exact declarations reserve first. Preferred declarations are sorted by stable
typed identity and allocated second. A subsystem must reject two distinct
identities whose ordering facts are equal; pointer addresses, expression
hashes, map order, and rendered text are not tie-breakers.

A companion whose spelling includes another declaration, such as
`Validate<Result>`, is registered as a dependent declaration before freeze.
The base and companion may live in different generated packages when both
belong to the same generation run. Goa first chooses every independent name in
every package, then derives and reserves each companion from its base's exact
final name. A declaration from another generation run is rejected. Callers
never rebuild the companion by concatenating a separately resolved type
string.

### Imports and output paths

Complete import path is the only import identity. Static-template requirements
have priority over generated-package preferences, which have priority over
design metadata preferences. References and `ImportSpec` values consume the
same frozen binding, while each file imports only the paths it uses.

Each generated source contribution owns a `GeneratedImportPlan`. Before
freeze, its generator records the runtime packages written directly by
templates, the generated packages visible in type expressions, and any
packages used while generated code reads nested fields. The generated package
chooses one qualifier for each complete path across all of its files. After
freeze, each contribution resolves only its saved paths to those chosen
qualifiers. Contributions that share an output file merge their completed
imports. A renderer must not walk the design again, scan generated source, or
add a broad list and remove unused imports later.

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

`UserType.Origin()` identifies the first declaration in one family of exact
copies. A generated transport type may start a new family, and renaming a type
deliberately starts another one. Goa uses this identity while binding types,
normalizing designs, planning service and transport declarations, creating
protobuf messages, transforming values, validating fields, building views,
and generating examples. It keeps unrelated same-named types separate, but it
does not by itself prove that two generated declarations have the same fields
or behavior.

Plugins that write a named type call `DeclareGeneratedType` to reserve its Go
name, then call `BindGeneratedType` for every user type expression that must
refer to that declaration. The binding follows `UserType.Origin()`, so copies
reuse one name while unrelated equal-shaped types may use different names. Do
not pass a user type to `DeclareName` to express this ownership: that older
lookup compares the type definition and can treat unrelated declarations as
the same type.

An emitted declaration identity contains every fact that changes its generated
source: owning package and role, source provenance, wire shape, validation,
defaults, views, pointer policy, ordered union branches, and protocol-specific
metadata as applicable. Equal semantic `ID()` values do not merge distinct
origins. Conversely, the same authored origin may produce distinct request and
response records when their emitted contracts differ.

`expr.Union.Hash()` remains expression identity. `UnionTypeID` describes the Go
and JSON definition emitted for one union occurrence. `UnionDeclarationID`
combines that definition with `AttributeExpr.AuthoredAttribute()`: copies of
one authored `OneOf` share a declaration only while their emitted definitions
also match, while separately authored unions remain separate even when every
branch is equal. Do not change expression hashes, decorate string keys, or add
general expression provenance to coordinate code generation.

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

## Value conversion plans

`TransformPlan` records one conversion from a source Go value to a target Go
value. Creating the plan copies both type graphs, finds every recursive helper
function the conversion will call, and records the choices made by hooks that
change the type shape. Later edits to the design types or hook value cannot
change the planned conversion. Two generated copies remain separate even when
they came from the same authored type; only an edge back to the exact copied
value closes a recursive cycle.

Planning hooks may return a different source or target attribute for one
conversion step, but they must not edit either graph they receive. The planner
checks this immediately after each hook returns, including changes to nested
defaults and metadata, and rejects the plan when a hook mutates its input.
This makes the returned choice explicit without letting one hook change what a
later hook or the renderer sees.

The caller declares every helper function in the package that will contain it,
binds those declarations and the final source and target type resolvers, and
then renders the conversion. Planning and rendering therefore use the same
function declarations and final package qualifiers. `GoTransformWithAttrs`
keeps its released interface, but now performs these same steps internally and
returns the released helper data after rendering once.

Each recursive call is a helper occurrence. `Helpers` returns those occurrences
with the requiredness that decides whether the generated caller checks for nil.
Several occurrences can call the same function. `HelperDefinitions` returns one
record for each distinct function body, comparing the complete retained source
and target attributes because hooks may inspect their metadata, defaults, and
validation. Caller requiredness is deliberately not part of a definition.
Each definition also carries a private authored location made from object
fields, array elements, map keys and values, and union branches. Equivalent
calls keep the earliest location, so reordering sibling fields does not change
which colliding declaration keeps its preferred Go name. The location is used
only for ordering and never appears in generated source.
Generators normally declare one name per definition and bind it with
`BindHelperDefinition`; `BindHelperDeclaration` remains available for callers
that manage individual occurrences. A transport may supply
`SameHelperDefinition` only when its complete hook set proves that one retained
difference cannot change the function body. gRPC uses this for protobuf field
numbers: a field number changes the serialized position, not the Go code that
copies the field value.

When several `TransformPlan` values write helpers to the same Go package, the
caller collects them in one `TransformHelperRegistry` before package names are
final. Each plan is supplied with the source and target `GoTypePlan` values that
describe the actual generated parameter and result types. The caller also
supplies the package name order for each exact helper location. After every plan
has been collected, the registry groups functions only when those Go types, the
conversion rules, and the ordered child function calls are equivalent. It
repeats the child comparison until recursive call graphs stop changing. Each
group returns the complete order and definition from one real helper occurrence;
it never combines the conversion identity from one occurrence with the location
from another. The registry never compares rendered Go source, hashes, or
preferred names. The caller declares one function for each returned group and
binds that declaration to the complete group before rendering.

Each helper occurrence retains its own authored location. A shared function
definition still keeps the earliest location for stable name ordering, but that
location cannot be used to find the generated type at another call. For
example, identical recursive values reached through a direct field and a map
value may share one function even though their generated layouts must first be
resolved at two different paths.

`Helpers` and `HelperDefinitions` return detached type descriptions with
plan-owned IDs. A caller
may inspect or change those descriptions while choosing a function declaration,
but cannot change the type graph used for rendering. Rendering also rejects a
hook that changes the retained graph and checks that calls sharing a declaration
produce the same parameter type, result type, and body. When one plan is
rendered again with the same source variable, target variable, and assignment
choice, it returns a copy of the first result instead of calling hooks again.
This lets one conversion plan serve repeated template requests without letting
hook state change previously generated code.

Most conversions can discover their helper calls from the source and target
types. A custom union renderer that calls helpers itself must also implement
`TransformHooks.PlanUnionHelpers`. That hook records the exact branch pairs for
which the renderer will request helpers. This keeps union rendering
specialized: generated code contains only the branch conversions selected
during generation and performs no runtime lookup by branch name or package.

Authored defaults use the same final type resolver and pointer rules as value
conversion. `RenderGoValue` writes the exact composite literal, local scalar
declarations, and planned `OneOf` constructor call while the complete design
and generated package names are available. Transform code and HTTP request
constructors both consume that result; they do not format defaults with Go
reflection in the generated program. HTTP command defaults are separate
because they are wire values: HTTP codegen first applies mapped JSON field
names and byte encoding, then the generic command generator serializes that
already selected value. This keeps service Go layout and HTTP JSON layout from
guessing about each other.

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

Every HTTP and JSON-RPC output file records the service, views, and authored
package paths it will use before package names become final. Linking resolves
only those saved paths to their final qualifiers; it does not inspect design
expressions again or rebuild a service path from its name. Public plan snapshots
return detached nested values, so changing a snapshot cannot change a later
read or rendered file.

A method with separate ordinary and streamed results plans the streamed SSE
body independently. When the retained service value and HTTP body have the
same Go layout, the client assigns the decoded value directly. When their Go
layouts differ, the client validates the decoded body and uses the exact saved
conversion. An empty streamed result emits its explicit zero-value return.
These choices are made while generating source; the generated client does not
inspect types or select a conversion at runtime.

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

After adding services and any new user types to that root, the preparation
plugin calls `(*expr.RootExpr).EvaluateAttachedServices`. This checks that the
new expressions belong to the same root, prepares and validates all of them,
and finishes none of them when any one is invalid. It is the public operation
for adding evaluated services; plugins must not run individual expression
steps or use the package-global root.

Factory plugin planning receives `*generator.Plan`. It may declare plugin-owned
output through the same Generation, and it consumes core declarations through
the exact retained service plan. Factory plugin rendering receives the same
plan after names are final. It may add files and sections, but it cannot create
another root, re-run service or transport analysis, reserve a name, or change
an expression. Released callbacks keep their original arguments and do not
gain a planning phase.

An HTTP plugin calls `Plan.HTTP(root)` with the exact prepared service root it
received. The method returns the ordinary HTTP plan for that root. It returns
false for a different root value and for a root that only has JSON-RPC methods.
During `Plugin.Plan`, the plugin may declare an exported server handler wrapper
for an exact HTTP service, an unexported handler wrapper for one exact HTTP
endpoint, or an extra server mount for an exact HTTP service in that plan. Goa
submits those function names to the service's generated server package, so
collisions are settled with Goa's own names before source is written.

A declared handler wrapper has the shape `func(http.Handler) http.Handler`.
Goa writes direct nested calls inside each exported endpoint and file mount
helper. Direct callers of those helpers therefore receive the same wrapping as
callers of the service's `Mount` function. Endpoint handlers, file handlers,
and redirects defined by the design are covered, and the first declared
wrapper is the outermost. The service's `Mount` function passes each handler to
its helper unchanged, so wrappers run exactly once.
The linked HTTP service data retains that exact declaration list and copies it
into each generated endpoint and file mount helper, so plugin output can match
the declaration by pointer even for a service that contains only files.
An endpoint handler wrapper also has the shape
`func(http.Handler) http.Handler`, but Goa writes it only in that endpoint's
exported mount helper. Service wrappers surround endpoint wrappers. File mount
helpers receive service wrappers only. This lets a plugin add behavior to one
designed endpoint without changing another endpoint, a file route, or an extra
plugin mount.
An extra server mount has the shape `func(goahttp.Muxer)` and supplies the
method label, HTTP verb, and path that Goa adds to the generated server's
`Mounts` list. Goa calls extra mounts after routes from the design, in
declaration order. Extra mounts are separate calls and are not passed through
the declared handler wrappers. Both declarations must happen before generation
freezes; later calls, JSON-RPC plans, foreign services, missing route fields,
and changes to the linked declarations are generation errors.

MCP generation therefore attaches its generated service expressions during
prepare and later consumes `Plan.Service(root)`. Agent tool specifications use
one retained typed specification plan for each output package; public specs and
transport specs are distinct packages with distinct declaration owners. No
plugin coordinates through a process-global map, a latest result, a decorated
hash, `PlanKey`, or render order.

When a plugin needs a generated payload field, it asks the retained service
plan for `MethodPayloadLayout`. The returned `GoTypePlan` contains the exact Go
field spelling and pointer choice already selected by service generation. For
example, a design field sent as JSON `cursor` may be generated as
`OriginalCursor` because of field metadata; the plugin keeps `cursor` on the
wire and emits `payload.OriginalCursor`. It must not call `Goify`, inspect the
JSON name, or repeat the service pointer rules.

## File assembly

`SectionTemplate.Name` labels a section for diagnostics. It is not declaration
identity. Package plans remove identical declaration contributions before they
become sections. The file merger combines imports and appends every non-header
section in producer order, even when diagnostic labels match. Conflicting
declarations remain visible and fail during planning or Go compilation instead
of disappearing silently.

## Compatibility and operations

Goa keeps the released callback registration API alongside the retained-plan
API. Both registration styles enter the same generation run, use the same
prepared designs and current file list, and run in the released first, normal,
and last order. The released API still accepts the same name more than once for
one command. Registrations with the same position and name run in registration
order, matching released Goa v3. Factory plugin names remain unique, and Goa rejects
a released registration whose command and name match a factory registration so
the same plugin cannot run twice through two APIs. Goa does not restore the
released functions that ran plugins; the generator remains the only code that
executes them.

This compatibility is deliberately limited. A released preparation callback
may change a design before Goa chooses names. A released generation callback
may edit ordinary values and remove or reorder files, sections, or list
entries. A nil file entry is ignored for every list size. Released Goa ignored
nil among several files but accidentally panicked when nil was the only file.
Goa's own templates read declarations chosen during planning rather than the
released string copies, so changing only a released name field does not rename
Goa's code. The public strings remain final name snapshots for existing plugin
templates. A plugin may deliberately replace declaration fields, sections,
templates, source, or file finalizers; as in released Goa, that plugin owns the
correctness of the resulting source. The released four-argument callbacks do
not receive the retained plan. Any declaration or reference they add after Goa
has fixed its planned names also remains plugin-owned: Goa preserves that output
for compatibility, but does not choose, reconcile, or check those added names.
Move a plugin to `generator.PluginFactory` and declare its names during
`Plugin.Plan` when Goa must prevent collisions with the rest of the generated
package.

An ordinary callback error is returned unchanged and stops the run. If the
same callback also changes a prepared design after planning, Goa reports that
forbidden change instead. The changed root remains visible after the failed
run, so hiding it behind the callback error could corrupt a later run.

A plugin that adds package declarations or chooses Go names must use a
`generator.PluginFactory` and declare that work through `Plugin.Plan` before
names become final. Rebuilding Goa's private service or transport records is
not supported; plugins should render their own typed values instead. Upgrade
those plugins with Goa, then regenerate the whole generated tree from an empty
output directory.

A gRPC plugin that renders viewed-result conversions must use
`ResponseData.ServerConverts` and `ClientConverts`, or
`StreamData.SendConverts` and `RecvConverts`, as appropriate. The singular
conversion fields remain available for existing templates, but they describe
the default or fixed view only. Reusing one singular conversion for every view
can read a field that the selected view deliberately omitted.

The remaining breaks are limited to generation and generated source except
where the runtime changes are listed below. There is no persisted-data
migration or staged generator mode. Already compiled programs do not start
using the new generator merely because the Goa module is updated.

### Generator library migration

The following table lists the preserved, removed, or signature-changed
exported APIs in this change. “No direct replacement” means callers must stop
performing that step; the run or the owning retained plan now performs it once.

| Package | Old API | New API or required change |
| --- | --- | --- |
| `codegen` and `codegen/generator` | `codegen.RegisterPlugin`, `RegisterPluginFirst`, and `RegisterPluginLast` accepted prepare and generate callbacks | These four-argument functions remain available. Goa runs their callbacks in the same run as plugins registered through `codegen/generator`. Registration now panics for an empty name, an unknown command, a nil generate callback, or a call made after generation has started. Use the factory API when a plugin must plan new package declarations or read service and transport plans from the current run. |
| `codegen` | `PrepareFunc`, `GenerateFunc`, `RunPluginsPrepare`, and `RunPlugins` | `PrepareFunc` and `GenerateFunc` remain available for registration. The two run functions have no replacement because `generator.Generate` owns the complete plugin lifecycle. |
| `codegen/generator` | `Genfunc`, the replaceable `Generators` variable, and the exported `Example`, `Service`, `Transport`, and `OpenAPI` functions | These released entry points remain available. Built-in functions selected through `Generators` join the command's shared planning pass. Other `Genfunc` values run after Goa has fixed every generated name. Calling a built-in function directly creates and runs its own complete plan without registered plugins. Use a plugin factory when added output must declare package names or inspect the command's service and transport plans. |
| `codegen` | `NormalizeRoot` | No direct replacement. `codegen.NewGeneration` performs normalization once and records the exact generated method types. |
| `codegen` | `AddServiceMetaTypeImports` | No direct replacement. The owning service or transport plan records the imports used by each output file. |
| `codegen` | `NewAttributeContextForConversion` | No direct replacement. Build and bind a `TransformPlan`; its source and target contexts retain the correct package owners. |
| `codegen` | Custom `Attributor` implementations needed only `Scoper`, `Name`, `Ref`, and `Field` | Implement `Package`, `Enter`, `IsSumType`, and `ValidatorCall` too. These methods make package ownership and exact validator calls explicit. |
| `codegen` | `AttributeContext.DefaultPkg` and `SamePackageConversion` | Removed. Package lookup belongs to the `Attributor`; enter the source or target attribute instead of setting a default package or a same-package mode. `ArrayElementPointer` is new and is only for a wire array that must distinguish JSON `null` from a primitive zero value. |
| `codegen` | `TransformHooks.HelperNameAttrs` | Removed. Bind each recursive helper through `TransformPlan.Helpers` and `BindHelperDeclaration`; do not derive a helper name from a second attribute walk. A custom `TransformUnion` that calls `TransformHelperName` must also implement `PlanUnionHelpers` so planning can declare those functions before rendering. |
| `codegen` | `WrapDirective.InitTypeName` | Set `WrapDirective.Target` to the wrapper attribute. Rendering resolves its already planned Go type name. |
| `codegen` | `TransformFunctionData` contained only `Name`, parameter and result references, and code | It now also identifies the planned helper with `ID` and `Declaration`. `Name` remains as a deprecated copy of `Declaration.Name()` for every rendered helper. Unkeyed struct literals must be updated. |
| `codegen` | `TransformPlan.Helpers` and `BindHelperDeclaration` | These methods remain available for individual helper calls. New generators should use `HelperDefinitions` and `BindHelperDefinition` so required and optional calls to the same conversion share one strict function. |
| `codegen` | `GoTransformWithAttrs` returned a separate nil-tolerant helper for some optional calls | The signature is unchanged, but returned helpers are strict: optional callers check for nil before calling them, and equivalent required and optional calls may return one shared helper instead of two. Plugins that call a returned helper directly must perform the same nil check at the call site. |
| `codegen/cli` | `BuildCommandData(data)`, `EndpointParserFile(..., data, parseSection)`, `UsageCommands(data)`, `UsageExamples(data)`, and `FlagsCode(data)` | These released signatures and section data remain available. Goa's HTTP and gRPC planners call private planned variants that carry final import, declaration, and local-variable names. |
| `codegen/cli` | `BuildFunctionData.Name`, `ActualParams`, and `FormalParams` | These fields remain and contain the final generated name and parameter lists. New planning code may also read `BuildFunctionData.Declaration`. |
| `codegen/cli` | `NewFlagData` accepted a type name; `FieldLoadCode` accepted type and validation source strings; `FlagArgData.TypeName` and `Validate`; positional `FlagData` literals | The released functions and fields remain available and keep their string-based behavior. Goa's transport planners use one opaque `FlagPlan`, created by `NewFlagPlan`, so validation is written against the exact parsed value without rewriting generated text. Supplying both `Validate` and `Plan` is rejected. Use named fields when constructing `FlagData` because it now contains private planning state. |
| `codegen/example` | Global `Servers`, `ServersData`, `ServersData.Get`, and `APIPkg` | No direct replacement. Create `example.Plan` with `example.NewPlan`, then get its copied `example.Root`. Package names come from the associated service plan. The pure `RootPath(genpkg)` helper remains available. |
| `codegen/example` | `CLIFiles(genpkg, root)` and `ServerFiles(genpkg, root, services)` | Call `CLIFiles(root)` and `ServerFiles(root, services)` with the `*example.Root` returned by the retained example plan. |
| `codegen/example` | `VariableData.VarName`; `HandlerArg.Endpoint` and `Service` contained generated local variable names | `VariableData.VarName` is removed. `HandlerArg.Service` contains the design service name, `Endpoint` reports whether the endpoint collection is needed, and `Variable` contains the final local variable name after the example plan is linked. `Data` also reports whether the server uses HTTP or JSON-RPC. |
| `codegen/service` | `NewServicesData` | Create one `codegen.Generation`, then call `service.NewPlans` for the complete root batch. Use `service.NewPlan` only when the generation has exactly one service root. After freeze and `Link`, use `Plan.Services`. |
| `codegen/service` | `ClientFile`, `EndpointFile`, `ConvertFiles`, `InterceptorsFiles`, and `ViewsFile`; `Files(genpkg, service, services, userTypePkgs)` | No direct per-file replacement. Call `service.Files(plans...)` after every supplied plan is linked and handle its `([]*codegen.File, error)` result. The plans decide the complete file set. |
| `codegen/service` | `SetUserTypeImports`, `AddServiceDataMetaTypeImports`, and `AddUserTypeImports` | No direct replacement. Imports are retained per file by the service plan. |
| `codegen/service` | `ExampleServiceFiles(genpkg, root, services)` and `ExampleInterceptorsFiles(genpkg, root, services)` | Pass the linked `*service.Plan` to `ExampleServiceFiles(plan)` or `ExampleInterceptorsFiles(plan)`. |
| `codegen/service` | Public render-data fields `Data.UserTypeImports`; `MethodData.IsJSONRPC`, `IsJSONRPCSSE`, and `IsJSONRPCWebSocket`; `EndpointMethodData.IsJSONRPC`, `IsJSONRPCSSE`, and `IsJSONRPCWebSocket`; and `StreamData.SendAndCloseName`, `SendAndCloseDesc`, `SendAndCloseWithContextName`, and `SendAndCloseWithContextDesc` | Removed. Factory plugins inspect the HTTP or JSON-RPC plan's endpoint data. Released service-template plugins must update templates that read these fields. There is no JSON-RPC WebSocket replacement because design validation rejects that transport combination. `EndpointsData.VarName`, `ClientVarName`, and `ServiceVarName` and `EndpointMethodData.ClientVarName` and `ServiceVarName` remain as deprecated copies of their final declarations. `Data.ViewsPkg` and `ProjectedTypeData.ViewsPkg` remain available when Goa generates a views package. `ErrorInitData.Name`, `InitData.Name`, and `ValidateData.Name` also remain as deprecated copies of real planned declarations. A custom error has no generated service constructor, so both its constructor declaration and compatibility name are empty. `ValidateData` contains function calls now, so it cannot be compared with `==` or used as a map key. |
| `codegen/service` interceptor section data | Interceptor wrapper sections exposed `map[string]any` values with keys such as `Method`, `Service`, and `Interceptors` | The sections now use private typed data built for the exact method and call kind. Plugins that replace only a section's source must update to the current section contract; plugins cannot type-assert the old map. This lets generated accessors use exact payload, result, and stream types without runtime method checks. |
| Generated service interceptors | An interceptor method accepted `info *NameInfo`, where `NameInfo` was an exported struct with private fields | The method accepts `info NameInfo`, where `NameInfo` is an interface with the same public accessor methods. Update handwritten interceptor signatures by removing `*`. Goa now writes a private implementation for each service method and call kind, so payload and stream accessors use the exact generated types without inspecting the method at runtime. |
| `expr` and `dsl` | `APIExpr.ExampleGenerator`; `dsl.Randomizer(expr.Randomizer)`; `NewRandom` | Store an immutable `APIExpr.RandomizerFactory`. Pass `NewFakerRandomizerFactory` or `NewDeterministicRandomizerFactory` to the DSL. For direct example generation, call `NewExampleGenerator(factory).At(identity)`. The standalone `NewFakerRandomizer` and `NewDeterministicRandomizer` constructors remain available. |
| `expr` | Exported concrete `FakerRandomizer` and `DeterministicRandomizer`; the embedded `ExampleGenerator.Randomizer`; `ExampleGenerator.Derived`, `Rebased`, `Field`, `PreviouslySeen`, and `HaveSeen` | The two standalone randomizer types and their released constructors remain available. Generation uses `RandomizerFactory` so every typed example identity receives a fresh value sequence. The embedded stream and recursion methods are removed; select a public typed `ExampleIdentity`, then descend with `Member`, `ArrayElement`, `MapKey`, `MapValue`, or `UnionMember`. |
| `expr` | Custom implementations of `UserType` | Add `Origin() UserType`. A copy returns the first declaration in its current family. An independently created or intentionally renamed type returns itself. Goa uses this identity to recognize copies without treating unrelated types with the same name as one type. |
| `expr` | Repeated calls to `(*AttributeExpr).Validate` in one process could skip errors reported by an earlier call | Every call now validates the supplied expression and reports its errors. Direct users must not rely on a previous call hiding a later validation failure. |
| `expr` | A non-pointer `ResultTypeExpr` value satisfied `UserType` through its embedded `*UserTypeExpr` | Use `*ResultTypeExpr`. Renaming a result must also clear the result's stored copy origin, so `Rename` now belongs to the pointer and a value no longer satisfies `UserType`. Ordinary result expressions were already created and passed as pointers. |
| `expr` | Preparation plugins added services and called expression steps themselves | After adding the services and any new user types to their owning root, call `(*RootExpr).EvaluateAttachedServices`. It prepares and validates the complete added set before finishing any of it. |
| `expr` | Positional struct literals for `ResultTypeExpr`, `SchemeExpr`, `ServiceExpr`, and `UserTypeExpr` | Use literals with named fields or the package constructors. These structs now contain private identity fields, so code outside `expr` cannot initialize every field by position. |
| `expr` | `UnionToObject` | No direct replacement. HTTP and gRPC plans retain their own wire representation for a union. |
| `codegen` | Comparing `TransformAttrs` values or using them as map keys | `TransformAttrs` now stores maps and function-planning state, so it is no longer comparable. Pass pointers or compare the specific public fields that matter to the caller. Positional literals also need to become named-field literals. |
| `grpc/codegen` | `NewServicesData(serviceData)` | Create `grpc/codegen.Plan` values with `NewPlans`. `ServicesData.GRPCServices` and `ServicesData.Get` remain available, but callers should read the instance built by the plan instead of rebuilding gRPC analysis. |
| `grpc/codegen` | `ClientFiles(genpkg, data)`, `ClientCLIFiles(genpkg, data)`, `ProtoFiles(genpkg, data)`, `ServerFiles(genpkg, data)`, `ServerTypeFiles(genpkg, data)`, and `ClientTypeFiles(genpkg, data)` | These released signatures remain available. They render the files already recorded by `data` and panic when `genpkg` differs from `data.GenPkg()`. New plugin code should prefer the corresponding linked `Plan` methods. |
| `grpc/codegen` | `ExampleCLIFiles` and `ExampleServerFiles` | Create an `ExamplePlan` with `NewExamplePlan`; call its `CLIFiles` and `ServerFiles` methods. |
| `grpc/codegen` | `EndpointData.ClientMethodName`, `MetadataData.Map`, `MetadataData.MapStringSlice`, and `ValidationData.Name` | All four remain as deprecated copies of final generated data. Valid designs now reject map-shaped gRPC metadata, so `Map` and `MapStringSlice` are false after successful generation. `InitArgData` and `MetadataData` now contain a validation function, so they cannot be compared with `==` or used as map keys. |
| gRPC generation tools | `protoc-gen-go` and `protoc-gen-go-grpc` were discovered by `protoc`; the Makefile installed their latest releases | Install `protoc-gen-go v1.36.12` and `protoc-gen-go-grpc 1.6.2`. Goa resolves those programs before planning and rejects another reported version or an attempt to replace them through `Meta("protoc:cmd")`. The exact pair is the tool contract currently covered by generated-module tests; version text alone is not a general proof that another binary would produce different or compatible declarations. |
| `http/codegen` | `NewServicesData` and `NewJSONRPCServicesData` | Create HTTP plans with `NewPlans` or JSON-RPC HTTP plans with `NewJSONRPCPlans`. Both require the exact `*service.Plan` for the root. |
| `http/codegen` | `ClientFiles`, `ClientEncodeDecodeFile`, `ClientCLIFiles`, `ServerFiles`, `ServerEncodeDecodeFile`, `ServerTypeFiles`, `ClientTypeFiles`, `PathFiles`, and `WebsocketClientFile` | These released signatures remain available. They return files already built by the retained plan and reject a generated package argument that differs from the supplied `ServicesData`. New plugin code should prefer the linked HTTP plan methods. |
| `http/codegen` | `ExampleCLI`, `ExampleCLIFiles`, `ExampleServer`, and `ExampleServerFiles` | Create an HTTP `ExamplePlan` and call its `CLIFiles`, `ServerFiles`, or `CombinedServerFiles` methods. |
| `http/codegen` | `OpenAPIFiles(root)` | Call `NewOpenAPIPlan(root, exampleGenerator)`, then `Files`. |
| `http/codegen` | `CreateHTTPServices` testing helper | No direct replacement. Tests must construct, freeze, and link service and HTTP plans like production. |
| `http/codegen` | `SSEData.DataFieldTypeRef`; `ServiceData.ServerTypeNames`, `ClientTypeNames`, and `UnionTypes` | `DataFieldTypeRef` remains as a deprecated copy of `SSEData.Data.TypeRef` for an explicitly mapped data field. The three service-wide type lists are removed; read the generated type declarations supplied by the linked HTTP plan. `AttributeData` now contains a validation function, so it cannot be compared with `==` or used as a map key. |
| `jsonrpc/codegen` | `ClientFiles`, `ServerFiles`, and `ExampleServerFiles` | Create and link a JSON-RPC plan; call its `ClientFiles` and `ServerFiles` methods. Use `NewExamplePlan` for example files. |
| `jsonrpc/codegen` | `CreateJSONRPCServices` testing helper | Use `CreateJSONRPCPlan` when a test needs the linked production plan. |
| `jsonrpc` | `IDToString(id any) string` | Call `IDToString(id any) (string, error)` and handle the error. The function accepts decoded string and `json.Number` IDs and returns an error for every other Go type instead of silently returning an empty string. |
| `jsonrpc` | Positional literals for `RawRequest` and `RawResponse` | Use named fields. `RawRequest` adds `Invalid` and `HasMethod`; `RawResponse` adds `Invalid`, `HasResult`, `HasError`, and `HasID`. Generated clients and servers use these fields to distinguish a missing JSON member from a member whose value is empty or null. Existing named-field literals continue to compile. |
| `jsonrpc` | WebSocket `StreamConfig`, `StreamConfigOption`, `StreamErrorType`, `StreamErrorHandler`, `StreamErrorConnection`, `StreamErrorProtocol`, `StreamErrorParsing`, `StreamErrorOrphaned`, `StreamErrorTimeout`, `StreamErrorNotification`, `NewStreamConfig`, `WithRequestTimeout`, `WithConnectionTimeout`, `WithCloseTimeout`, `WithResultChannelBuffer`, `WithWebSocketBuffers`, `WithRetryConfig`, `WithCompression`, `WithPingInterval`, `WithErrorHandler`, and `(*StreamConfig).Validate` | No replacement. JSON-RPC WebSocket generation was removed. JSON-RPC supports unary HTTP calls and server streams through explicit server-sent events. |
| `codegen/service` | Unkeyed `UnionTypeData` and `UnionFieldData` literals; public union name strings | Use named fields. `UnionTypeData` adds `TypeDeclaration` and `KindDeclaration`. Each `UnionFieldData` adds `KindDeclaration` and `ConstructorDeclaration` for its generated constant and constructor. `FieldName` is the public branch spelling used by accessors; `StorageName` is the private struct field that stores the selected value. The released `Name`, `KindName`, and `KindConst` strings remain final snapshots for existing plugin templates. `Constructor` is a new final-name snapshot alongside the new declaration fields. |
| `http/codegen/openapi` | Process-global `Definitions`; `APISchema`, `GenerateServiceDefinition`, `ResultTypeRef`, `ResultTypeRefWithPrefix`, `TypeRef`, `TypeRefWithPrefix`, `GenerateResultTypeDefinition`, `GenerateTypeDefinition`, `GenerateTypeDefinitionWithName`, `TypeSchema`, `TypeSchemaWithPrefix`, `AttributeTypeSchema`, and `AttributeTypeSchemaWithPrefix` | The global definition cache and its mutating helpers have no replacement. For OpenAPI 2 attribute schemas, use `v2.BuildAttributeSchema(api, attribute, exampleGenerator)`; otherwise build a complete v2 or v3 document so definitions remain local to that build. |
| `http/codegen/openapi/v2` | `NewV2(root, host)` and `Files(root, path)` | These released signatures remain available and use the evaluated design's randomizer factory. Call `NewV2WithValues` or `FilesWithValues` when supplying translated values or a specific example generator. |
| `http/codegen/openapi/v3` | `New(root, version)` and `Files(root, version, path)` | These released signatures remain available and use the evaluated design's randomizer factory. Call `NewWithValues` or `FilesWithValues` when supplying translated values or a specific example generator. |
| `codegen`, `codegen/cli`, `codegen/example`, `codegen/service`, `expr`, `grpc/codegen`, and `http/codegen` | Positional literals for changed render-data structs | Use named fields or, preferably, the owning plan constructor. The exact affected types are listed below. Some now contain private state and cannot be fully constructed outside their package. |

The following exported structs gained or replaced fields, so an unkeyed literal
that compiled with Goa v3 must be changed to named fields:

- `codegen`: `AttributeContext`, `TransformAttrs`, `TransformFunctionData`,
  `WrapDirective`.
- `codegen/cli`: `BuildFunctionData`, `CommandData`, `FlagArgData`, `FlagData`,
  `InterceptorData`, `SubcommandData`.
- `codegen/example`: `Data`, `HandlerArg`.
- `codegen/service`: `EndpointMethodData`, `EndpointsData`, `ErrorInitData`,
  `InitData`, `InterceptorData`, `MethodData`, `MethodInterceptorData`,
  `ProjectedTypeData`, `ServicesData`, `StreamInterceptorData`, `UnionFieldData`,
  `UnionTypeData`, `UserTypeData`, `ValidateData`, `ViewData`, and
  `ViewedResultTypeData`. `UnionTypeData` adds `TypeDeclaration` and
  `KindDeclaration`; `UnionFieldData` adds
  `KindDeclaration`, `ConstructorDeclaration`, and `StorageName`.
- `expr`: `APIExpr`, `HTTPResponseExpr`, `ResultTypeExpr`, `SchemeExpr`,
  `ServiceExpr`, and `UserTypeExpr`.
- `grpc/codegen`: `EndpointData`, `InitArgData`, `InitData`,
  `LegacyDecodeData`, `MetadataData`, `RequestData`, `ResponseData`,
  `ServiceData`, `ServicesData`, `StreamData`, and `ValidationData`.
- `http/codegen`: `AttributeData`, `CookieData`, `Element`, `EndpointData`,
  `FileServerData`, `HeaderData`, `InitArgData`, `InitData`,
  `MultipartData`, `ParamData`, `PayloadData`, `ResponseData`, `ServiceData`,
  `ServicesData`, `SSEData`, `TypeData`, and `WebSocketData`.
- `jsonrpc`: `RawRequest` and `RawResponse`.

`TransformAttrs`, `codegen/cli.FlagArgData`, `service.ValidateData`,
`grpc/codegen.InitArgData`, `grpc/codegen.MetadataData`,
`grpc/codegen.StreamData`, and `http/codegen.AttributeData` now contain maps,
slices, or functions. They can no longer be compared with `==` or used as map
keys.

Several exported template-data structures now carry `*codegen.NameDeclaration`
records so every use reads the exact name chosen during planning. Goa keeps
public name fields when they can be copied from one real declaration after
names are final. This includes HTTP names, service constructors and validators,
CLI payload builders, and gRPC client methods and validators. For HTTP data,
plugins may edit ordinary render values and remove or reorder entries.
Goa's templates read the declaration field for each generated name rather than
its released string copy. A planning plugin may create a simple template value
for its own declaration, such as an HTTP constructor or gRPC constructor or
validator. The returned file must be in the generated package that reserved
that declaration.

The preserved HTTP name fields are `ServiceData.ServerStruct`,
`MountPointStruct`, `ServerInit`, `MountServer`, and `ClientStruct`;
`EndpointData.MountHandler`, `HandlerInit`, `RequestDecoder`,
`ResponseEncoder`, `ErrorEncoder`, `ClientStruct`, `RequestEncoder`,
`ResponseDecoder`, and `BuildStreamPayload`; `MultipartData.FuncName` and
`InitName`; `SSEData.StructName`; `FileServerData.MountHandler`;
`WebSocketData.VarName`; `InitData.Name`; and `TypeData.VarName`,
`ValidatorName`, and `NestedValidatorName`. Existing templates that read these
fields continue to work. Copied JSON-RPC service data also keeps
`ServerStruct`, `ServerInit`, `MountServer`, and `ClientStruct`. Copied
JSON-RPC endpoint data keeps `HandlerInit`, `ClientStruct`, `RequestEncoder`,
`RequestDecoder`, and `ResponseDecoder`. Each string contains the final name
from its declaration. A plugin that creates a new package-level declaration
must use the factory API and declare it during `Plugin.Plan`.
Typed service, endpoint, and file values copied from Goa keep the wrappers and
extra mounts saved with them. Plugins may still remove the files or sections
that render those values.
Plugins should use their own template values for plugin-owned types rather than
constructing Goa transport data.

HTTP package definitions and uses read the same planned declarations. This
includes WebSocket stream types, request builders and conversion functions,
body types, and their public and nested validators. The released `VarName`,
`Name`, `ValidatorName`, and `NestedValidatorName` strings mirror those
declarations, but Goa's templates do not use those copies as declaration
identity. A primitive or inline composite Go type such as `string` or
`[]string` has no package declaration;
the plan records that complete type expression instead. JSON-RPC receives a
copy of the same body declaration when one exists. Request builders and body
conversion functions are declared before names freeze, including constructors
for inline request bodies, so another generator claiming the preferred name
changes both the generated definition and every call.

HTTP client command sections also keep the released `MultipartFuncName` and
`BuildStreamPayload` strings. They copy the matching declarations after names
are final, while Goa's templates read the declaration records directly.

gRPC preserves the same name snapshots. `ServiceData.ServerStruct`,
`ClientStruct`, `ServerInit`, and `ClientInit`; `EndpointData.ServerStruct`,
`ClientStruct`, `ClientBuild`, `ClientEncode`, `ClientDecode`, `ServerHandler`,
`ServerDecode`, and `ServerEncode`; `StreamData.VarName`; and
`LegacyDecodeData.FuncName` each contain a final planned name. Goa's own gRPC
templates read the declaration saved for each role. Plugin templates may
continue to read the public snapshots, and plugin-owned source remains the
plugin's responsibility.

Other affected data includes `service.EndpointsData`, `EndpointMethodData`,
`ErrorInitData`, view and interceptor data; gRPC service, method, request,
response, and transform data; and CLI parser and payload-builder data. Call
`Name()` only after the generation is frozen. Existing unkeyed literals for
changed exported data must become keyed literals or, preferably, be replaced
with the owning plan constructor. `NameScope.Unique` and a previously unseen
`HashedUnique` call now panic after that scope freezes. Use `NameScope.Fork`
only for private render-local helpers; package declarations must be collected
through their `GeneratedPackage`.

### Protobuf tools

Every gRPC generation run now requires these exact programs on `PATH`:

```text
protoc-gen-go v1.36.12
protoc-gen-go-grpc 1.6.2
```

Install them with:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
```

Goa resolves each program to an absolute path and verifies its `--version`
output before planning protobuf names. `Meta("protoc:cmd", ...)` may still
select the protobuf compiler or add compiler arguments, but it may not replace
either required Go plugin with `--plugin`. A missing program, another version,
or a plugin override now stops generation before files are written.

### Regenerated application source

Regenerate all Goa-owned files together. Do not copy a subset of a new `gen`
tree over an old one: generated declarations and their callers use the same
frozen name records and are not designed to compile across generations.
`goa example` deliberately keeps existing starter files, so separately update
handwritten starter code and any application code that imports generated
transport packages.

Generated command starters now run an endpoint and write its result instead of
returning `(goa.Endpoint, any, error)` to their caller. The private `doHTTP`,
`doGRPC`, and `doJSONRPC` functions accept a context and output writer, print
unary or streamed values, and return endpoint, stream receive, output, and
connection-close errors. Regenerate or update the whole command directory
together; an existing kept `main.go` cannot call a newly generated private
transport function with the old signature. HTTP and gRPC input-only or
bidirectional stream commands now return a clear unsupported-input error before
parsing an endpoint or opening a connection.

Generated gRPC command flags for complete protobuf messages now decode
protobuf JSON instead of decoding the generated Go struct with the standard Go
JSON package. Scripts must use field names and value spellings accepted by
protobuf JSON; the generated example shows one accepted spelling. This also
gives enums, wrapper values, and protobuf aliases their defined JSON behavior.
Primitive command flag names and accepted text do not change. The exported
`Build<Method>Payload` helper does change for a flag without an authored
default: its string argument becomes `*string`, where nil means the flag was
omitted and a pointer to an empty string means the caller supplied the empty or
zero value. Handwritten callers must pass pointers and handle required values
explicitly; Goa no longer inserts the text `REQUIRED` for an omitted required
flag.

An error declared at API level is now a reusable definition, not an error that
every service and method returns. A service or method selects that definition
with a name-only `Error("busy")`; the reference keeps its type, validation,
defaults, description, and `Temporary`, `Timeout`, and `Fault` settings. A call
that supplies any additional error argument defines a separate local error.
Add an explicit name-only reference to every service or method that previously
relied on finding an API error without selecting it.

An HTTP method that defines both an ordinary `Result` and a `StreamingResult`
uses the ordinary result for a normal HTTP response and the stream for its SSE
response. A regenerated example service now returns that ordinary result from
its method and uses the default result view when the service owns view
selection. Existing kept service starters for such a method must be updated;
regenerating `gen` alone does not rewrite them.

Most generated declarations that have one clear released name keep that name,
including HTTP request and response types, validators, body constructors,
handler constructors, and mount functions. Names still change when preserving
the old spelling would preserve a false conflict or require two declarations
for one generated operation. A service method named `FooEndpoint` becomes
`Foo` when `Foo` no longer exists in that package. When two gRPC methods perform
the same conversion, their two method-based constructors become one constructor
named from the source and target types. Real name collisions receive stable
numeric suffixes instead of suffixes determined by discovery order. These are
Go source breaks for handwritten code that names the affected declarations.
They do not change the wire format by themselves.

A relocated service type written to a package selected with `struct:pkg:path`
now uses the lowercase final path segment as its Go package clause. For
example, a path ending in `APIKeyService` now declares `package
apikeyservice`. The older mixed-case spelling could differ from the package
name used by the generated import. Handwritten imports that rely on the old
implicit qualifier must use the new lowercase name or add an explicit import
alias. Generated imports already carry the planned qualifier, and the file
path and exported type names do not change.

Union declarations requested in another generated package now live in that
package's `unions.go` file. Their public family uses the exact `OneOf` name; Goa
never adds a numeric suffix to resolve a collision. Separately authored unions
that ask for the same family name now fail generation, even when their branches
are equal. Give each declaration a distinct `TypeName`. Compiler-created branch
types now use `<Union>Branch<Branch>`, such as `ValueBranchText`, instead of the
old `<Union><Branch>` spelling. Authored branch type names remain unchanged.
Plugins and scripts that select generated files by filename must stop assuming
a union is written beside the service that first used it.

HTTP union names now state which wire body owns them. A `Value` union becomes
`ValueRequestBody` in an ordinary request, `ValueStreamingBody` in a streaming
request, `ValueResponseBody` in a response, and `ValueDetailedResponseBody` in
the `detailed` response view. Handwritten transport code and plugins that name
the old union declaration must use the new exact name after regeneration.

Generated union branch fields are now private so code cannot leave stale values
in branches that are no longer selected. Handwritten code must replace reads
such as `u.Text` with `u.AsText()` and writes such as `u.Text = value` with
`u.SetText(value)` or `NewValueText(value)`. Plugin templates use `FieldName`
for the public branch spelling and `StorageName` only when rendering the private
field inside the union implementation. JSON and protobuf data are unchanged.

Generated interceptor implementations must also change their method
signatures. An argument such as `*LoggingInfo` is now the read-only
`LoggingInfo` interface. Goa supplies a private implementation for the exact
service method and for an endpoint call, stream send, or stream receive.
Handwritten interceptors should accept the interface and continue calling its
accessor methods. These known values are no longer stored in a public struct at
runtime.

A handwritten multipart request decoder now fills the generated HTTP request
body instead of the service payload. For example,
`func(*multipart.Reader, **service.UploadPayload) error` becomes
`func(*multipart.Reader, *UploadRequestBody) error`. Goa validates that body
and then builds the service payload, just as it does for JSON requests. Array
and map bodies use pointers to the complete generated body value. Regenerate
first, then update each handwritten decoder to its new generated parameter
type. The multipart bytes on the network do not change.

Some exported helpers disappear when Goa can prove that they do no work. For
example, Goa no longer writes an empty `ValidateUserTypeView` function. Code
that called an empty generated validator must remove that call. JSON-RPC
server-stream methods no longer receive a unary `Decode<Method>Response`
function because their generated endpoint returns a client stream and never
calls that decoder. Generated gRPC conversion constructors may also be renamed
or combined when several methods perform the same conversion. Handwritten code
should normally enter through the generated client, server, or endpoint
constructors instead of calling transport conversion helpers directly.

Generated gRPC response encoders with mapped metadata now read the actual
result variable and use code specialized for the metadata type. Some older
combinations generated references to nonexistent `res` or `p` variables and
did not compile. Scalar text remains equivalent, but bytes now use their string
contents instead of Go's slice display: `[]byte{65, 66}` changes from
`"[65 66]"` to `"AB"`. Floating-point values use the exact width selected by
the design. A metadata consumer that compared the old text must update. For a
fixed-view result, generated gRPC encoders and decoders use the view selected in
the design instead of trusting a runtime `goa-view` metadata value. Valid peers
already send the designed view and keep the same result body.

For caller-selected gRPC result views, Goa now generates one protobuf
conversion for each view. The server writes only the fields in the selected
view, and the client uses the matching constructor instead of applying the
default-view constructor to every response. A dynamic gRPC server stream sends
the selected view in its initial `goa-view` metadata before its first message;
the generated client reads that value before decoding the first message. The
protobuf schema does not change. Regenerate both sides of a dynamic viewed
stream together because an older server does not send this metadata and an
older client always decodes the default view. Regenerate both sides of any
viewed method whose selected view omits default-view fields: older generated
conversions can read an omitted field and fail. Fixed-view methods otherwise
keep their existing view choice.

One generated command-line flag changes only for a design that used `domain`
for both a server variable and a URL variable. The URL variable is now
`-url-domain`; the old starter registered `-domain` twice and did not run.

Generated examples also change because each example now uses the exact design
declaration that owns it rather than a value consumed earlier in a shared
random stream. Design-authored example inputs are unchanged, but generated
OpenAPI documents, CLI examples, array lengths, and decoding-error examples may
change because Goa now uses the authored value for that declaration.

OpenAPI 3 documents may place generated examples under `examples.default`
instead of the single `example` field. Byte examples use base64 text in both
JSON and YAML documents. OpenAPI 2 security requirements without scopes use an
empty array instead of JSON `null`. OpenAPI 3.2 server-sent-event schemas mark
the event data field as required only when the selected stream field is
required. Snapshot consumers should regenerate and review the documents; the
service's accepted requests and returned results do not change from these
documentation-only differences.

OpenAPI server variables now include the descriptions written in the design.
Reusable viewed-result schemas use the designed result type name in their
description instead of the generated HTTP response-body type name. These are
documentation text changes only.

When examples are disabled, OpenAPI generation now omits examples written in
the design as well as examples computed by Goa. Security definitions used only
by excluded services or methods are also omitted. Each server variable now
uses only its own allowed values; an older document could accidentally copy
allowed values from the preceding variable. These changes affect generated
documents, not the service wire format.

When several methods use one error type, the shared OpenAPI schema now keeps
the reusable type's description. Each OpenAPI 3 operation keeps the description
written for that method's error response. This changes generated API
documentation only; it does not change response data or status codes.

A viewed-result constructor now returns a nonnil value carrying an unknown
requested view. Generated boundary validation can therefore report the precise
invalid view instead of receiving nil or panicking first. Valid view values and
their projected bodies are unchanged.

Generated gRPC starters now print the designed method names directly instead of
asking the running gRPC server which methods it registered. Goa-designed
methods keep correct startup logs. A plugin that registers additional gRPC
methods at runtime must print its own log lines; those methods are not part of
Goa's generated service plan.

A normalized output-path collision, conflicting same-file package or import,
conflicting keep-existing-file setting, or path that is absolute, escapes the
generation root, contains a backslash, or differs only by filesystem case now
fails during planning. Previously, one file could overwrite another or the
merge could produce invalid Go. Same-path contributions now keep every body
section even when two sections have the same diagnostic label, and every file
finalizer runs in contributor order. A plugin that relied on a duplicate label
to suppress output or on only the first finalizer running must remove that
assumption.

### Designs that now stop generation

These checks reject designs that Goa cannot generate faithfully. They run
before files are written, so they do not require a staged rollout or data
migration. Fix the design, then regenerate the complete generated tree.

| Design | What now fails | How to migrate |
| --- | --- | --- |
| Authored default values | Defaults on declared types, result types, and errors are now checked through the complete nested value. Generation fails when a default has the wrong Go type, omits a required nested field, selects an unknown `OneOf` branch, or violates a nested length, range, enum, pattern, or format rule. Arrays, maps, and their elements and keys are checked too. | Change the default so it is a valid value of the designed type. If the rejected value is intentional, change the type or its validation rule instead of relying on Goa to emit an invalid default. |
| Relocated authored types | A declared type moved with `Meta("struct:pkg:path", ...)` now fails when it refers to another declared type that has no explicit generated package. Without that information, Goa cannot safely import the dependency and may create an import cycle. Compiler-created nested types remain with their owner and do not need this metadata. | Give each referenced declared type its own `struct:pkg:path`, usually the same path when the types belong together, or stop relocating the outer type. |
| Ordinary HTTP and JSON-RPC route overlap | One server cannot mount an ordinary HTTP route and a JSON-RPC route with the same HTTP method and matching path pattern. Parameter names do not make routes different: `POST /tasks/{taskID}` and `POST /tasks/{id}` accept the same requests. Released generation could let one handler replace the other. | Change one path, or mount the services on different servers. Keep JSON-RPC routes on `POST`. |
| JSON-RPC route, method, and request ID rules | Every JSON-RPC route must use `POST`, and a JSON-RPC method name cannot start with the reserved `rpc.` prefix. A request ID must be at most one direct string field in the payload. It cannot also appear in `params`, an HTTP parameter, header, cookie, or `Last-Event-ID`, and results and errors cannot define another request ID. When an optional ID has no default, its validation rules must accept the UUID that the generated client creates. | Change the route to `POST` and rename a method that starts with `rpc.`. Keep one direct string ID field or let Goa create the ID. Remove duplicate transport mappings. If an optional ID has rules that reject a UUID, make it required, give it a valid default, or change those rules. |
| JSON-RPC notifications | `Notification()` is only valid for a one-way method. The method may have a payload, but it cannot define a request ID, result, streaming payload, or streaming result. | Remove the response and stream from the notification, or remove `Notification()` and use an ordinary JSON-RPC request with a non-null ID. |
| Multipart requests | `MultipartRequest()` now requires a non-empty request `Body`. A multipart decoder fills that generated body before Goa builds the service payload, so there is no valid decoder contract without one. | Add an explicit `Body` mapping for the multipart fields, or remove `MultipartRequest()` and use the ordinary request decoder. |
| Server-sent-event field mappings | `SSEEventData` cannot select one field from a viewed streaming result because that would omit the view name needed to decode the value. When `SSERequestID` is used, its field cannot also have a `Header` mapping, and `Last-Event-ID` cannot map a different field because the server-sent-event transport uses that header for the selected request ID. | Remove `SSEEventData` so the complete viewed result is sent. Use only `SSERequestID` for `Last-Event-ID`, and remove any competing header mapping. |
| Other designs Goa cannot represent safely | Generation rejects duplicate `ConvertTo` or `CreateFrom` mappings for the same Goa and external type, non-primitive gRPC metadata, streaming HTTP success fields mapped to headers or cookies, JSON-RPC success or error fields mapped to HTTP response headers or cookies, and inherited HTTP or gRPC error mappings whose actual method error has a different type, validation, default, or metadata. A service and its methods also cannot reuse one standard error name with different settings or value definitions. gRPC and JSON-RPC reject a method with both `Result` and `StreamingResult`. Ordinary HTTP allows both only with `ServerSentEvents()`. JSON-RPC rejects client and bidirectional streams and requires `ServerSentEvents()` for a server stream. | Remove the duplicate mapping or unsupported transport mapping. Give different errors distinct names. Split methods that have two result contracts, and select a streaming form supported by the transport. JSON-RPC response values must stay inside `result` or `error.data`; request headers and cookies remain valid. |

### Wire and runtime compatibility

| Change | Mixed old and new programs | Rollback effect |
| --- | --- | --- |
| Required Goa `OneOf` validation | Generated service validators and HTTP and gRPC boundaries now reject a union with no selected branch. They also reject a selected message, bytes, or `Any` wrapper whose branch value is nil. The protobuf encoding is unchanged. Valid branches, including a selected empty message, work across versions. | Rolling back re-allows invalid union values; it requires no data migration. |
| Required singular protobuf primitive presence | Goa now writes `optional` on singular protobuf booleans, numbers, strings, enums, bytes, and their aliases so a validator can distinguish omission from an explicit zero or empty value. Generated `pb.go` fields for booleans, numbers, strings, enums, and their aliases become pointers. Bytes and bytes aliases remain `[]byte`; `Any` and other message fields remain pointers. Goa service types do not change, and protobuf field numbers and binary tags stay compatible. An old client works with a new server when it sends an explicit nonzero or nonempty value, but old generated code cannot mark a required zero or empty value as present, so the new server rejects that omission. A new client can send an explicit zero or empty value to an old server. Regenerate handwritten protobuf struct literals and replace direct scalar field reads with the generated accessors where needed. | Rolling back accepts omitted required zero and empty values again. Protobuf data needs no migration. |
| Protobuf defaults | While converting protobuf input into a service value, Goa applies an authored default only when the protobuf field is absent. Explicit zero values, empty bytes, and protobuf null remain explicit values. Conversion in the other direction never adds a default; it writes exactly what the service returned, including zero, nil, empty arrays, and empty maps. Released conversions could replace those explicit values with defaults. | No schema or stored-data migration is required. Callers that depended on explicit zero, empty, null, or nil values being silently replaced must use omission on input or return the desired default from service code. Rolling back restores the accidental replacement. |
| `ArrayOfRequired` for primitive values and primitive aliases in JSON bodies | Valid JSON is unchanged. Incoming JSON representations use pointer elements: server request bodies and client response bodies use values such as `[]*string`, then convert them to service value slices such as `[]string`. Outgoing client request bodies and server response bodies remain value slices. Incoming `[null]` is now rejected. Handwritten code that constructs incoming transport body values must supply pointers. gRPC repeated scalar fields and service arrays remain value slices. | Rolling back accepts `null` again; it requires no data migration. |
| Exclusive maximum validation | Generated primitive validators now reject values equal to or above an exclusive maximum. Older validators accidentally repeated the exclusive minimum check and could accept those values. | Rolling back accepts values that violate the designed maximum; it requires no data migration. |
| Command payload builders | Primitive command flag names and their text format stay the same. In generated Go, a `Build<Method>Payload` string parameter for a flag without an authored default becomes `*string`. Nil means the caller omitted the flag; a pointer to `""` means the caller explicitly supplied the empty or zero value. Omitting a required value now returns an error instead of putting the text `REQUIRED` in the payload. | Update handwritten calls to pass a string pointer or nil. Scripts that invoke the generated command keep the same flag spelling. Rolling back restores the old string parameter and placeholder behavior. |
| Optional fields selected as request bodies | Generated HTTP clients omit the body when the selected service field is absent and keep an explicitly empty or zero value. Servers treat an empty HTTP body and JSON `null` as absence, while preserving path, query, header, and cookie fields decoded from the same request. When the selected field has a default, absence applies that default; explicit zero, empty array, and empty map values remain explicit. A defaulted scalar service field is a value and is always sent by a direct generated client. A nil defaulted collection is omitted and the server applies the default. Generated command clients use valid JSON text for collection defaults. JSON-RPC clients omit `params` for an absent direct object, map, array, or union, but keep present empty containers and selected empty-message union branches. Optional primitive JSON-RPC parameters remain positional: `[null]` means absent and an explicit zero remains present. Generated payload constructors for optional scalar bodies now take a pointer, and generated command builders use the same presence-aware constructor. | Regenerate clients, servers, and command packages together. Update handwritten calls to generated payload constructors for optional scalar bodies to pass a pointer or nil. Custom JSON-RPC peers must omit direct `params` for absence and must not send `params: null`; positional primitive `[null]` remains valid. Before this change, an absent selected body became the service zero value instead of an authored default, command clients formatted collection defaults as invalid JSON, and current development clients could emit invalid pointer operations for defaulted scalar aliases. Rolling back restores those behaviors. |
| Caller-selected JSON-RPC result views | A successful response is `{ "jsonrpc": "2.0", "id": ..., "result": { "view": "detailed", "body": ... } }`. The envelope is the method value inside JSON-RPC's standard `result` member. It is generated only when the caller chooses among views. The old result was the projected body alone; unary HTTP responses carried the view in the `goa-view` header, while stream messages had no reliable place for it. An old client and new server, or a new client and old server, are not compatible for these methods. Results without views and results whose view is fixed in the design keep their old body shape. The envelope is used consistently for unary and server-sent-event results. A configured response decoder now receives an HTTP response with status 200 instead of the previous zero status. JSON-RPC response fields must stay in `result` or `error.data`; design validation rejects response `Header` and `Cookie` mappings because one HTTP response may carry several JSON-RPC messages. | Deploy or roll back every client and server for a caller-selected view together. There is no dual decoder. A custom response decoder that inspected the synthetic status must accept 200. Move any JSON-RPC response header or cookie field into the result or error body before regenerating. Request headers and cookies are unchanged. This generic Goa envelope is valid JSON-RPC, but a method that belongs to another protocol layered on JSON-RPC must still use that protocol's required result schema. |
| Designed JSON-RPC errors | A designed service error uses its authored JSON-RPC code and keeps its generated message. API-level mappings now keep the code written in the JSON-RPC DSL; released generation could accidentally replace it with the API HTTP error status. The DSL accepts the five standard protocol codes and the server-error range from -32099 through -32000, but rejects other application codes in the reserved range from -32768 through -32000. The `error.data` value is now `{ "name": "busy", "body": ... }`, where `name` is the error name in the Goa design and `body` is the exact response body declared by `Body`. The name lets generated clients distinguish two errors that intentionally use the same code and prevents a protocol failure such as Internal Error from being mistaken for a designed error. A standard Parse Error or Invalid Request may use a null ID and remains a raw JSON-RPC error. If its code, name, and body instead identify a designed method error, a null ID is rejected as an invalid response because a designed endpoint error must repeat the request ID. Generated unary and server-sent-event clients also return the original JSON-RPC error when the data object is missing, malformed, names an error not allowed for that code, or uses an unknown code. Released servers put the raw service error in `data` and did not apply the planned response body conversion. | Regenerate generated clients and servers together. Custom clients must read the error name and decode the nested `body` instead of decoding `error.data` directly as the service error body. Custom servers must write the same object. Method- and service-level codes and all messages stay the same. An API-level mapping may now use its authored JSON-RPC code instead of the accidental HTTP status, so check custom peers that depended on that bug. No data migration is required. Rolling back requires rolling back both generated peers because released clients do not understand the new data object. |
| HTTP server-sent events | Event-write and flush failures are now returned instead of ignored, and clients decode retry values into the exact optional integer type. Clients ignore an empty retry field or one containing anything except ASCII digits; another valid retry field in the same event still applies. Clients remove an optional UTF-8 byte-order marker only from the first stream line. A valid event ID remains selected until another valid ID replaces it, an empty ID clears it, and an ID containing a NUL byte is ignored. Primitive and primitive-alias data use raw SSE text; an optional nil primitive omits the data line, while a present empty string writes an empty data line. The old optional-pointer server accidentally wrote JSON strings or `null`, so a new client reads the JSON quotes or `null` as part of the raw value; a new server remains readable by the old raw-text client. OpenAPI 3.2 now marks optional mapped data as optional and describes primitive data as raw text rather than JSON. For viewed results, the server writes one `goa-view` header, chooses the default when empty, rejects an unknown view before writing, and rejects changing the view after the first event. Generated servers reject IDs containing a NUL byte or line break and event names containing a line break because those values cannot be written safely. | Regenerate both sides for a stream with optional primitive data. Regenerate an OpenAPI 3.2 document if tooling relies on the event item schema. Other non-viewed event data keeps its shape, but code may now observe write errors and retry values. Do not mix versions for a variable-view stream. Invalid or partly written streams now fail instead of being accepted or followed by a second HTTP error response. |
| Viewed HTTP streams | HTTP SSE and WebSocket servers now reject an unknown requested view before encoding an event instead of writing a nil or `null` body. Valid fixed and selected views keep their existing body shape. | A new server can reject an invalid view that an old server accepted. No valid request needs a coordinated rollout. |
| Caller-selected views on HTTP client streams | Goa now rejects an HTTP WebSocket method that receives a `StreamingPayload` and also lets the caller choose the view of its `Result` or `StreamingResult`. This includes client-streaming and bidirectional methods. The request stream has no single request value from which the transport can select one response view. | Set a fixed `View` in the design before regenerating. Existing methods that already use a fixed view, and server-only streams where one request can select the view, are unchanged. Rolling back accepts the ambiguous design again. |
| Empty HTTP WebSocket streams | When a service returns successfully before sending or receiving a value, the generated server now upgrades the connection, sends a normal close frame, and closes it once. Released code returned success without completing the WebSocket handshake. Repeated `Close` calls on the generated server stream return the first result. The public stream methods and data frames do not change. | Clients observe a clean empty stream instead of a failed or hanging upgrade. No coordinated rollout or data migration is required. |
| JSON-RPC server-sent-event lifecycle | Each server `Send` accepts the streaming result directly and writes one JSON-RPC notification. Streaming methods are ordinary methods: the opening request requires a non-null ID, and the transport rejects a missing or null ID with Invalid Request before decoding parameters or calling the service. When the service method returns, the transport writes one terminal response: `result: null` for success or a JSON-RPC error for a returned error. Client `Recv(ctx)` becomes `Recv()` plus `RecvWithContext(ctx)`. Stream constructors return an interface that also implements the service client stream. After notifications, the client returns `io.EOF` or the terminal error. An optional primitive selected with `SSEEventData` writes `[null]` when nil and keeps explicit zero values. The client treats omitted params and `[null]` as absent and applies an authored default when present in the design. Empty retry fields and retry fields containing anything except ASCII digits are ignored, while a valid field in the same event still applies. The client now rejects an unknown server-sent-event name or a notification for another JSON-RPC method instead of silently skipping it. Body read and close failures are returned instead of discarded. Unary request reads and batch response delimiter writes now report failures through the server error handler. When an error code, designed-error name, or data object does not identify an allowed designed method error, generated clients return the original `*jsonrpc.RawErrorResponse` with its code, message, and data unchanged. Generated servers reject IDs containing a NUL byte or line break, event names containing a line break, and retry values that cannot be written as ASCII digits. | Generated service implementations must use the standard typed `Send`, `SendWithContext`, and `Close` methods. Client callers must use the new receive methods or service stream interface. The old JSON-RPC-only `StreamEvent`, `SendAndClose`, `SendError`, request ID, marker event, concrete client stream, and WebSocket service APIs are removed. Regenerate clients and servers together. Custom peers must send a non-null ID when opening a stream and must stop sending unknown event names or notifications for another method. Valid generated Goa clients already satisfy these rules. |
| JSON-RPC request and response rules | The method declaration now decides whether a call is a request or notification. An ordinary method requires a non-null `id`; a method declared with `Notification()` requires the `id` member to be absent. A mismatch receives Invalid Request before parameters are decoded or the service is called. Valid declared notifications run without producing a response. Empty string IDs and exact string or numeric IDs remain valid for ordinary methods. An invalid request object receives Invalid Request with `id: null`, including over server-sent events. A success with no value includes `"result": null`. Leading JSON whitespace no longer changes an array into a single request, `[]` returns one Invalid Request response, and valid and invalid batch members are handled independently. A batch cannot start a server stream: a valid streaming request receives Method Not Found with “Method is not available in a batch request.” In a server that has both ordinary and streaming methods, no `Accept` header or `*/*` permits each method's designed response format. A unary method requires an acceptable JSON media type, a streaming method requires an acceptable server-sent-event media type, and an unacceptable format returns HTTP 406. Media type spelling is case-insensitive and `q=0` rejects that format. Invalid method arguments map to Invalid Params; an undeclared service failure maps to Internal Error. | Regenerate clients and servers together. Custom peers must omit `id` only for methods declared as notifications and must send a non-null ID for every ordinary method, including streams. Valid ordinary requests and valid declared notifications keep their previous results. Clients that expected no error for an invalid object, one Parse Error for a mixed batch, an omitted `result`, an empty body for `[]`, or selection of a format the method cannot return must follow the generated contract and JSON-RPC 2.0 response rules. A custom caller of `RawRequest.UnmarshalJSON` must inspect `Invalid`: structurally invalid JSON-RPC objects and IDs other than strings, numbers, or null now set that field without returning a Go JSON decoding error. No data migration is needed. |
| HTTP and JSON-RPC response bodies | Generated clients close response bodies they fully consume and return read and close failures as decoding errors. When decoding and closing both fail, the returned error preserves both. A successful method that deliberately returns the raw body leaves it open for the caller. | Successful decoded responses are unchanged. Failure handling may now return a nonnil error, or `decoding_error` instead of a raw or request error, where older code discarded or mislabeled a read or close failure. |
| Nested HTTP validation paths | Generated validation errors now keep the complete field and array-index path while entering nested generated types instead of restarting the path at the nested value. | Invalid inputs may produce more precise error field names. Valid inputs and wire data are unchanged. |
| Maps stored in the complete HTTP query string | A payload map now decodes the raw query keys that generated clients send, such as `?a=1&b=2`. Released servers looked for bracketed keys such as `query[a]` even though released generated clients did not write that form. | Generated clients work with the new server without change. Custom clients that worked around the released decoder must send raw keys; otherwise `query[a]` becomes the literal map key. Rolling back restores the mismatched decoder. |
| HTTP float query text | Generated clients now use Go's shortest round-trip text for float32 and float64 query values. Ordinary values are unchanged, while very large or small values may use exponent form, such as `1e+100`, instead of a long decimal expansion. Servers decode the same number. | Systems that sign, cache, or compare the exact URL text must accept the compact form. Rolling back returns to the longer spelling. |
| gRPC metadata text | Generated metadata conversions are specialized for the designed type. Bytes now use their string contents, so `[]byte{65, 66}` is sent as `"AB"` instead of `"[65 66]"`. Floating-point values use the designed width. Other scalar text remains equivalent. | Regenerate both sides when a metadata consumer parses the old byte-slice display or depends on the old floating-point spelling. Rolling back restores the old text. |
| gRPC result views | Unary and streaming clients and servers now use the protobuf conversion and validation rules for the selected view. A client rejects a missing selected field before conversion instead of silently using a zero value or panicking; fields omitted by that view are not required. Dynamic server streams send the view in initial `goa-view` metadata, and dynamic clients require that metadata before decoding the first message. The protobuf schema is unchanged. An old stream server does not send this value, while an old stream client assumes the default view. Older conversions can also fail when a selected view omits a field used by the default conversion. A non-default view validator has a stable name such as `ValidateShowResponseTiny`; the default view keeps `ValidateShowResponse`. Identical top-level validators for one message and view share that name, so obsolete numeric duplicates such as `ValidateShowResponse2` disappear. Handwritten code that calls generated validators must use the validator for its selected view. | Regenerate and deploy both sides of a dynamic viewed gRPC stream together. Also regenerate both sides of any viewed method whose selected view omits default-view fields. Fixed-view methods otherwise keep their wire choice. |
| Nested gRPC validator names | Validators called directly by generated transport code remain exported as `Validate<Message>` or `Validate<Message><View>`. Validators used only for nested fields are private. Identical nested checks that receive the same argument and add the same error path share one function. Its ordinary name records the API, service, protobuf message, argument, and error context. If different checks would otherwise receive the same name, Goa also records the method, request or response role, selected view or error, and complete typed field path. Punctuation is written explicitly instead of being collapsed into the same Go spelling. Adding another field therefore does not renumber an existing helper. A root validator never shares a declaration with a nested helper. Validation checks and error paths are unchanged. | Regenerate the complete gRPC client and server packages. Handwritten code must call the exported top-level validator instead of a former nested helper such as `ValidateDetail` or `ValidateArrayOfString`. Plugins that write code into the generated package must use the planned validation declaration rather than rebuilding its name. No protobuf or stored data changes. |
| Protobuf wrappers around repeated and map values | Protobuf does not record whether a repeated or map field was omitted, so Goa no longer invents a missing-field error for the generated wrapper field. A nil wrapper used as a map value decodes as an empty collection instead of panicking. Authored rules still apply: for example, nil and present-empty wrappers both fail an authored minimum length of one, and item rules still check every present item. A wrapper with no authored rule no longer emits an otherwise empty exported validator. Handwritten code that called that unnecessary validator must remove the call. Generated function signatures and protobuf data are otherwise unchanged. | Rolling back can panic while converting a nil map-value wrapper and can reject an empty repeated value as missing. No data migration is needed. |
| Protobuf name collisions | Normal schemas keep their existing field encoding. A design whose protobuf declarations collide may receive stable numeric message, file, or generated Go suffixes instead of invalid source. Descriptor names, and a gRPC method path if its service or method itself needed a suffix, can therefore change for that formerly conflicting design. | Regenerate both sides from the same Goa version. A previously valid, collision-free schema needs no coordinated runtime rollout. |
| API error selection | API-level errors are reusable definitions. They no longer appear automatically on every service and method. A name-only `Error("busy")` on a service or method selects the parent definition and keeps all of its settings; supplying more arguments defines a separate local error. | Designs that relied on implicit API error lookup must add the explicit name-only selection before regenerating. Compiled peers and stored data do not need migration. Rolling back makes the API error visible through implicit lookup again. |

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
