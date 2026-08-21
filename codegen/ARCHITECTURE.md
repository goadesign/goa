# Code Generation Architecture

This document defines how Goa turns one evaluated design into generated files.
It focuses on ownership rules that are easy to violate when a declaration is
written outside its service package.

## Generation lifecycle

The `goa` command compiles and runs a temporary generator for one design. The
evaluation package registers exactly one core `*expr.RootExpr`; its evaluation
name is `design`, and duplicate evaluation names are rejected. A generation run
then follows this order:

1. Evaluate and validate the design.
2. Let preparation plugins amend the evaluated expression roots.
3. Normalize the roots.
4. Create one generation context for the normalized roots.
5. Let every selected core generator and plugin declare the generated service
   types it may emit, then freeze their package names.
6. Let the core service, HTTP, gRPC, JSON-RPC, and OpenAPI generators render
   files using the frozen declarations.
7. Let post-generation plugins render additional files using the same context.
8. Merge contributions with the same output path and render the files.

The core service generator therefore reasons about one real design root. A
temporary root created by a plugin is a separate analysis unless the plugin is
explicitly given the active generation context.

## Ownership

| Concern | Owner | Consumers |
| --- | --- | --- |
| Design identity and structural equality | `expr` | validation and code generation |
| Go identifiers in a generated package | the generated-package record for its output path | service and transport rendering |
| Relocated user-type and union declarations | the same generated-package record | file rendering |
| HTTP, gRPC, and JSON-RPC wire types | each transport generator | transport templates |
| Output-path merging | the generator | all file-producing plugins |

`expr.Union.Hash()` describes expression identity. It must not change merely
because generated Go source needs a different notion of equality. Code
generation uses a separate typed union identity containing every property that
changes the emitted union declaration, including discriminator keys, branch
order, branch type shape, and relocated branch packages.

## Generated packages

The generation context owns a catalog keyed by the actual generated import
path. Each catalog entry represents one Go package and owns:

- one `codegen.NameScope` for every package-level identifier;
- the relocated user types rendered into that package;
- the structurally distinct unions rendered into that package; and
- the final names of each union type, discriminator type, constants, and
  constructors.

The generation copies its evaluated roots and generated module import path at
construction. Callers can read those values but cannot replace the registered
roots or redirect output packages after planning begins. Planning and rendering
therefore test membership against the same root snapshot.

The generation also owns one import qualifier for each complete import path.
Imports referenced by static templates have required qualifiers, generated
service and views packages have preferred qualifiers, and design metadata has
lower-priority preferences. Required qualifiers are allocated first and
conflicting requirements are rejected; generated and metadata qualifiers may
receive deterministic suffixes. Each generated file still imports only the
paths used by the declarations and references it renders.

Planning a declaration returns its canonical record. Once every selected
generator and plugin has planned its output, the context freezes the catalog.
Rendering may only look up those records; a late attempt to add a declaration
is an error. Code that declares a type and code that refers to it must consume
the same record or the package-aware attribute scope backed by it. A transport
must never recreate a union name from a service-local `NameScope`.

For example, suppose two services place types in `gen/types`. Both contain a
nested union whose natural Go name is `Value`, but the unions have different
branches. The package catalog may assign `Value` and `Value2`. The user-type
definitions, service methods, HTTP transforms, and gRPC transforms must all read
those exact assignments from the `gen/types` catalog.

Relocated declared user types keep the Go form of their declared name. If two
declared names in one output package become the same Go identifier, such as
`foo-bar` and `foo_bar` both becoming `FooBar`, generation rejects the design
before rendering. Silently assigning `FooBar2` would make a public declaration
depend on unrelated traversal order.

Relocated user types are emitted in their metadata-selected files. Relocated
unions are emitted once in `unions.go` in the owning package, independent of
which service first referred to them.

## Attribute naming during transforms

`codegen.AttributeContext` asks an `Attributor` for names and references. A
service-local context uses the service package record. A context that transforms
a relocated type uses a package-aware attributor backed by the generation
catalog:

- a declared user type selects the package named by its `struct:pkg:path`;
- a nested union selects the package of the enclosing generated declaration;
- a local type selects the service package scope.

This is the only supported route for resolving generated service types inside
HTTP, gRPC, JSON-RPC, conversion, and validation helpers. Transport-specific
scopes still own transport-only wire declarations.

## Plugin and file assembly contracts

A plugin that can emit generated service types plans them before the catalog is
frozen and renders them with the active generation context. A plugin that
analyzes a temporary root must either plan those types in the active context or
emit to packages isolated from every other participant. Independent analyses
may not coordinate through package names, process-global maps, decorated
strings, or render-order assumptions.

Standalone or selective generation creates a fresh context containing exactly
the roots, generators, and plugins selected for that output. It runs the same
plan, freeze, and render phases. An API that accepts only roots and reconstructs
its own scope cannot safely contribute to a larger generation.

`codegen.SectionTemplate.Name` labels a template for diagnostics. It is not a
declaration identity. Output merging appends same-path sections and merges
imports; it must not discard sections merely because their diagnostic labels
match. Package owners remove identical declaration contributions before they
become file sections. Conflicting declarations remain visible and fail with a
generation or Go compilation error instead of disappearing silently.

## Review gate

Before changing type naming, relocated declarations, generation roots, plugin
files, or file merging, trace one declaration through all of these stages:

1. the single evaluated root;
2. service analysis;
3. planning and catalog freeze;
4. the owning generated-package record;
5. service declaration rendering;
6. HTTP and gRPC references;
7. post-generation plugin contributions; and
8. final files after output-path merging.

A service-only render test is insufficient. The regression must compile a real
generated module with both HTTP and gRPC enabled whenever those transports can
refer to the declaration.
