# Testing the Planned Generation Preview

Goa `v3.31.0-preview.1` is an opt-in preview of a broad correction to code
generation. It is intended for application authors and plugin authors who can
regenerate their code, review the result, and report problems before the work
becomes a stable release.

This preview does not replace the current stable release. The Go command does
not select a pre-release version for `@latest` while a stable version exists.
Projects receive this preview only when they request its complete version.
See the Go documentation for [version queries][go-version-queries] and the
[pre-release workflow][go-prereleases].

The final stable version number has not been chosen. These changes include
intentional source breaks, so feedback from this preview will inform both the
final contract and whether the stable release requires a new major version.

## Why this preview exists

Goa used to make some generated-name and type decisions in separate passes.
Those passes could describe the same Go package while seeing different sets of
names. One pass could declare a validation function under one name while a
later pass called another name. Similar disagreements affected imports, union
types, conversions, result views, examples, and plugin output.

The new generator prepares the complete design first. It then chooses every
generated package, file, declaration, import, validation function, conversion,
and helper once. HTTP, gRPC, JSON-RPC, command, example, OpenAPI, and plugin
generation all use those recorded choices. Templates write the selected code;
generated programs do not inspect their own type names or choose between paths
that were already known during generation.

That architectural correction exposed places where the old output accepted
values that the design rejected, described transport behavior the transport
could not provide, or exposed generator details as public APIs. The preview
fixes those contracts together instead of preserving contradictory behavior.

## Who should test it

Please test the preview if your project has any of these characteristics:

- it uses a Goa code-generation plugin;
- it uses required primitive fields, `OneOf`, result views, defaults, repeated
  values, maps, or metadata with gRPC;
- it uses JSON-RPC, server-sent events, or WebSocket streams;
- it implements generated interceptors;
- it constructs generated transport or protobuf types directly;
- it replaces generated sections or calls Goa generator packages directly; or
- it has several services, generation roots, transports, or plugins that write
  into the same generated Go package.

Small HTTP-only services are also valuable tests. They help confirm that the
new planning work leaves ordinary generated APIs and wire behavior unchanged.

## Install the preview

Start from a branch with the current generated tree committed. Until the next
preview is published, install both the Goa module and the `goa` command from
the preview branch:

```bash
GOPROXY=direct go get goa.design/goa/v3@fix/goa-generation-plan
GOPROXY=direct go install goa.design/goa/v3/cmd/goa@fix/goa-generation-plan
goa version
```

The direct module lookup is required because the public Go module proxy rejects
branch queries that contain `/` instead of asking the Git repository.

The Go command records the exact branch commit as a pseudo-version in
`go.mod`. The installed command reports the preview release line:

```text
Goa version v3.31.0-preview.1
```

Do not use an older `goa` command with the preview module. Also install the
protobuf generators covered by this release if the design uses gRPC:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
```

Goa checks these two program versions before writing gRPC files.

## Regenerate and test an application

1. Record the existing generated output in version control.
2. Update the Goa module and command as shown above.
3. Run the same complete `goa gen` command used by the project. `goa gen`
   deletes and recreates the complete `gen` directory.
4. Run `goa example` only if the project normally maintains Goa starter files.
   This command creates missing files but does not replace handwritten files.
5. Read every generation error. Several invalid or ambiguous designs now stop
   before source is written; the sections below explain how to update them.
6. Compile the complete project and update handwritten code for intentional
   source changes.
7. Run unit, integration, and client/server tests. Compare generated `.proto`
   and OpenAPI files with the committed versions, especially field numbers,
   routes, selected views, errors, and examples.
8. Run generation a second time and confirm that it produces no further diff.

Never copy individual files from the old generated tree into the new one. A
declaration and every use of it must come from the same generation run.

## Changes most application authors will see

### Required gRPC primitive fields now preserve presence

A required boolean, number, string, enum, byte slice, or alias must distinguish
an omitted value from an explicit zero value. The generated protobuf schema now
uses proto3 presence for these singular fields. In generated protobuf Go code,
booleans, numbers, strings, enums, and their aliases therefore become pointers.
Byte slices remain slices, and messages remain pointers. Goa service types keep
their existing field layout.

The protobuf field numbers and binary wire types do not change. The behavior
does: generated clients and servers reject an omitted required field instead of
converting it to the service type's zero value. Explicit `false`, `0`, an empty
string, an empty byte slice, and protobuf null remain valid when supplied.

Code that constructs protobuf messages directly must set the new pointers or
use generated getters when reading them. Regenerate both peers when required
zero or empty values matter. An old client cannot prove to a new server that it
explicitly selected a zero value because its schema did not record presence.

Defaults now use the same rule. Goa applies an authored default only when the
incoming value is absent. An explicit zero or empty value stays explicit.
Service results are sent exactly as returned; result conversion does not add
input defaults.

### `OneOf` values have one explicit selected branch

Generated union branch storage is now private. Construct, update, and read a
union with its generated constructor, setter, accessor, and `Kind` method:

```go
value := NewValueText("hello")
value.SetCount(3)
count, ok := value.AsCount()
```

The exact names use the union and branch names from the design. Compiler-made
branch types have descriptive names such as `ValueBranchText`; discovery-order
names such as `Value2` are gone.

A union holds exactly one valid branch. Selecting a new branch replaces the
old selection. A required union rejects no selected branch, a typed wrapper
whose value is nil, and a selected nil message, byte slice, or `Any` value. A
selected nonnil empty message remains valid. JSON and protobuf data keep the
same union representation.

Update handwritten code that set or read public branch fields. If two authored
unions would need the same public Go name in one package, give them distinct
`TypeName` values; generation now reports the collision instead of silently
renumbering public types.

### Required array items reject JSON `null`

`ArrayOfRequired` marks every array item as required. Incoming HTTP and
JSON-RPC body types now use pointers for primitive items and primitive aliases
so the decoder can distinguish `null` from a zero value. A request containing
`[null]` returns a validation error. Valid requests still become ordinary
value slices in the service layer, and generated response bodies remain value
slices.

Only code that constructs incoming generated transport bodies directly needs
to use pointers for these items. Service implementations do not change.

### Interceptor information is read-only

Generated interceptor methods now receive `NameInfo`, a read-only interface,
instead of `*NameInfo`, a pointer to a public struct whose fields were already
private. The interface provides the same public accessor methods for the
service, method, call kind, and typed value.

Update each handwritten interceptor signature by removing the `*`. Code that
used the public accessors otherwise remains the same. The generated private
implementation is specific to the method and call kind, so it does not inspect
method names or switch on payload types while the service is running.

### Selected gRPC views use their exact contract

Every selected result view now has matching conversion and validation. A field
omitted by a view is not treated as required, while a missing field selected by
that view is rejected before conversion. A dynamic server stream sends the
selected view in initial `goa-view` metadata, and the generated client reads it
before decoding the first message.

Regenerate and deploy both sides of dynamic viewed streams together. Do the
same for a viewed method where a non-default view omits fields used by the old
default conversion. Fixed-view methods otherwise keep their wire choice.

Top-level validators keep stable exported names such as
`ValidateShowResponse` and `ValidateShowResponseTiny`. Validators used only by
another validator are private. Handwritten code must call the top-level
validator for its selected view rather than an old numbered or nested helper.

### API-level errors are definitions, not automatic method errors

An error declared at API level defines a reusable error. It no longer means
that every service and method can return that error. Select it where it is
actually returned:

```go
var _ = API("calc", func() {
    Error("busy", Busy)
})

var _ = Service("calculator", func() {
    Error("busy")
})
```

The name-only service or method declaration selects the API definition and
keeps its type, validation, default, description, and fault settings. Supplying
another argument defines a separate local error. Add explicit selections to
designs that relied on the old automatic lookup.

### Custom errors use the exact declared type

Goa adds `Error`, `ErrorName`, and `GoaErrorName` methods only to the exact
authored type passed to `Error`. A named type nested inside that error remains
an ordinary generated type and no longer receives error methods accidentally.
Update application code that returned or inspected the nested type as though it
were the declared service error.

When the exact error type is also used as a payload or result, Goa emits one Go
declaration and adds the error methods beside it. The same rule applies when
`struct:pkg:path` places the type in a shared generated package: every service
imports the one declaration instead of generating competing data and error
copies. This is a generated Go source correction. It does not change the
transport body or require a persisted-data migration.

One authored error type must also have one static error name across the complete
generation. Goa now rejects a type used as, for example, `not_found` in one
service and `missing` in another service, even when the services are evaluated
through different generation roots. The old generator checked only errors in
the same method, so its generated `ErrorName` result could depend on which
service happened to emit the shared type.

If one type intentionally represents several names, add an `ErrorName` field to
the type and set the name on each error value. Otherwise, define a separate type
for each static name. Regenerate the complete design after making the choice;
generating services separately does not create separate names for a type placed
in one shared Go package.

## Transport changes

### JSON-RPC supports unary calls and explicit server-sent-event streams

JSON-RPC methods now have one of two contracts:

- one HTTP request followed by one JSON response; or
- one HTTP request followed by server results through
  `ServerSentEvents()`.

Design validation rejects client streams, bidirectional streams, JSON-RPC
WebSocket streams, server streams without `ServerSentEvents()`, and a method
that defines both `Result` and `StreamingResult`. Move an unsupported method to
unary JSON-RPC, JSON-RPC server-sent events, ordinary HTTP WebSocket, or gRPC.

Request and response handling now follows JSON-RPC 2.0 for IDs,
notifications, invalid objects, batches, `result: null`, invalid parameters,
internal errors, accepted media types, and streams in batches. A method
declared as a notification requires an absent `id`; every ordinary method
requires a non-null string or numeric `id`. Regenerate custom peers together
and update them to follow the declared method kind.

A result whose view is chosen by the caller uses this value inside JSON-RPC's
top-level `result` member:

```json
{
  "view": "detailed",
  "body": {
    "name": "example"
  }
}
```

Goa adds this envelope only when the caller chooses among views. Unviewed and
fixed-view results keep their body shape. The value is valid JSON-RPC, which
allows any JSON value in `result`; a separate protocol built on JSON-RPC must
still use the result shape required by that protocol.

Designed JSON-RPC errors keep the authored JSON-RPC code and place their Goa
error name and body in the error's `data` member. Custom clients that decode
that data must accept the new `{ "name": ..., "body": ... }` shape.

### HTTP decoding, validation, and streams are more exact

- Multipart decoders fill and validate the request body before constructing
  the service payload. Generated multipart decoder callback signatures may
  change, and invalid nested fields now report their complete field and array
  paths.
- Exclusive maximum validation rejects the maximum itself as well as larger
  values.
- A map assigned to the complete query string reads raw keys such as
  `?a=1&b=2`, matching generated clients. A custom client that used bracketed
  keys must change.
- Float query values use Go's shortest text that converts back to the same
  value. Systems that sign, cache, or compare exact URL text must accept the
  compact spelling.
- Server-sent events write primitive values as raw event text, preserve the
  difference between an absent optional value and a present empty string, and
  return write and flush errors.
- An empty successful WebSocket stream now completes the handshake, sends a
  normal close frame, and closes once.
- Generated clients return body read and close errors for bodies they consume.
  A method that deliberately returns the raw body still leaves it open for the
  caller.

### Other gRPC corrections

- Byte metadata uses its actual string contents. For example, `[]byte{65, 66}`
  becomes `"AB"`, not `"[65 66]"`. Floating-point metadata uses the width
  declared by the design.
- Protobuf cannot distinguish an omitted repeated or map field from an empty
  one. Generated validators no longer report a false missing-field error.
  Authored length and item rules still run.
- A nil repeated or map wrapper used as a map value converts to an empty
  collection instead of panicking.
- A validator that has no work is no longer generated. Remove handwritten
  calls to such a function.
- A schema with a real protobuf name collision may receive stable suffixes.
  Regenerate both sides of a formerly conflicting schema because its descriptor
  names or gRPC method path can change.

## Generated starter code, examples, and OpenAPI

Generated commands now execute the endpoint, receive stream values, print
them, and return endpoint, stream, output, and close errors. Review handwritten
command code against the new starter when updating it. gRPC flags for complete
messages now decode protobuf JSON.

Generated example values now belong to the design declaration that authored
them. Another service or transport can no longer consume shared random state
and change those values. This causes a large one-time text change in some
examples, but a second generation must be identical.

OpenAPI output now has consistent base64 byte examples, empty security scope
arrays instead of JSON null, independent server-variable examples, selected
view schemas, and server-sent-event data schemas. Review these changes as
contract corrections rather than accepting the complete generated diff
without inspection.

## Plugin and generator-library migration

The released four-argument plugin registration functions, `Genfunc`, the
replaceable `Generators` variable, and the exported `Service`, `Transport`,
`OpenAPI`, and `Example` functions remain. Their released ordering and support
for repeated callback names also remain. Common generated-name fields such as
`MountHandler`, `HandlerInit`, constructors, codecs, validators, multipart
helpers, server-sent-event names, WebSocket names, and gRPC names remain
available for existing templates.

A plugin that edits ordinary generated values or files should usually continue
to work. A plugin that declares a package-level name or chooses a generated
name must move that work to `generator.PluginFactory` and declare it during
`Plugin.Plan`. This lets Goa check the plugin's names with every other name in
the generated package before templates run.

A preparation plugin that adds a service must attach it to the owning root and
call `EvaluateAttachedServices`. Public helpers that directly ran plugin
callbacks or rebuilt Goa's private service or transport analysis are removed;
the complete generation run now owns those steps once.

Several public planning and template-data structures gained private state or
typed declaration records. As a result:

- positional struct literals must become named-field literals;
- values containing maps or function-planning state can no longer be compared
  with `==` or used as map keys;
- plugins that replaced interceptor section templates must use the new typed
  section data; and
- plugins that render viewed gRPC conversions must use the per-view conversion
  lists rather than one default conversion for every view.

The complete table of changed generator APIs and their replacements is in
[Code Generation Architecture](codegen/ARCHITECTURE.md#generator-library-migration).
That document is the detailed contract for plugin authors; this guide is the
application-facing summary.

## Changes that require coordinated deployment

Updating the Goa module does not change a compiled application. The changes
take effect only after regeneration and compilation. There is no persisted-data
migration.

Regenerate, deploy, and roll back both client and server together for:

- caller-selected JSON-RPC views;
- JSON-RPC server-sent-event streams;
- dynamic caller-selected gRPC streams;
- viewed gRPC methods whose chosen view omits fields used by the old default
  conversion;
- HTTP server-sent-event streams with optional primitive data; and
- gRPC metadata whose byte or floating-point text matters to another program.

Required protobuf fields keep compatible field numbers and binary types, but a
new server may reject a required zero value from an old client because the old
message did not record whether that zero was present.

Most other changes are source changes in generated Go packages or stricter
validation of previously invalid input. They require regeneration and a normal
application build, but not a data migration.

## Return to the stable release

To stop testing the preview:

```bash
go get goa.design/goa/v3@v3.30.0
go install goa.design/goa/v3/cmd/goa@v3.30.0
goa gen YOUR_MODULE/design
```

Regenerate the complete `gen` directory with the stable command. Do not keep a
mixture of stable and preview files. If a changed transport shape was deployed,
return both client and server to stable output together. No persisted data
needs to be changed or restored.

## Report what you find

Use [pull request #3971][preview-pr] for feedback about the preview as a whole.
Open a [GitHub issue][issues] for a reproducible bug. Use
[GitHub Discussions][discussions] for design and migration questions.

A useful report includes:

- the preview version, Go version, and operating system;
- `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` versions for gRPC;
- the affected transport and whether a plugin participates;
- the smallest design that reproduces the behavior;
- the exact `goa gen` or `goa example` command;
- the relevant before-and-after generated diff;
- the complete generation, compilation, or runtime error; and
- whether both client and server were regenerated.

Please remove credentials and private application data. A small public
reproduction is ideal.

[discussions]: https://github.com/goadesign/goa/discussions
[go-prereleases]: https://go.dev/doc/modules/release-workflow#pre-release
[go-version-queries]: https://go.dev/ref/mod#version-queries
[issues]: https://github.com/goadesign/goa/issues
[preview-pr]: https://github.com/goadesign/goa/pull/3971
