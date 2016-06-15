# goa Roadmap

This document summarizes the areas of focus for the next release (v2). These cover improvements and
new features that introduce non-backwards compatible changes.

## Breakout the Context

The generated context data structure contains both the request and response state. While this makes
it convenient to write controller code it does not make sense to pass these objects all the way down
in all layers. It also creates hidden dependencies.

The proposal is to break up the generated context into two data structures, one that contains the
request state and the other the response state. Concretely today the following design:

```go
Action("update", func() {
	Routing(
		PATCH("/:bottleID"),
	)
	Params(func() {
		Param("bottleID", Integer)
	})
	Payload(BottlePayload)
	Response(NoContent)
	Response(NotFound)
})
```

produces the following context data structure:

```go
// UpdateBottleContext provides the bottle update action context.
type UpdateBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	AccountID int
	BottleID  int
	Payload   *UpdateBottlePayload
	service *goa.Service
}
```

This proposal would break it down into:

```go
// UpdateBottleRequest provides the bottle update action request context.
type UpdateBottleContext struct {
	*goa.RequestData
	AccountID int
	BottleID  int
	Payload   *UpdateBottlePayload
}

// UpdateBottleResponse provides the bottle update action response context.
type UpdateBottleResponse struct {
	*goa.ResponseData
	service *goa.Service
}
```

The response functions would move from the `UpdateBottleContext` struct to the
`UpdateBottleResponse` struct.

Both the response and request state structs would be given to the controller action function. Today
the design above generates:

```go
Update(*UpdateBottleContext) error
```

Under this proposal the generated code would be:

```go
Update(ctx context.Context, resp *UpdateBottleResponse, req *UpdateBottleRequest) error
```

### Error Handling Improvements

Move from a model using structs to implement goa errors and check for them to a model using checking
for behavior.

Consider changing the goa request handler signature from:

```go
func (context.Context, http.ResponseWriter, *http.Request) error
```

to the standard:

```go
func (context.Context, http.ResponseWriter, *http.Request)
```

And keep the error in the response struct instead. This means tweaking how error handling is done so
that the error handler middleware knows where to look for errors.

## Code Generation Improvements

### Generator Hooks

The first planned improvement made to code generation is the ability for arbitrary plugins to hook
into the generated code of another generator. The main scenario that this should enable is the
ability to inject code in the built-in generator outputs (in particular `gen_app`). This would make
it possible to move the security and CORS generation in plugins and help keep the main generator
streamlined. Details on how this is enabled TBD. One possibility is to cater specifically to the use
case above and have the `gen_app` generator expose explicit hooks. Would be nice to come up with a
more generic approach though.

### Easier Plugin Use

Make it easier to invoke plugins on the command line. One possibility could be to have a hosted plugin
registry that goagen could look up where plugins could register with a simple moniker. This registry
could also provide discoverability of plugins (via something like `goagen plugins`). The registry
would only contain metadata (moniker, description, author, last update data etc.) and point to the
actual plugin package.

### Intermediary Representation

Generators start from the design data structures then massage that data into a shape suitable for
code generation. Today this is done in a ad-hoc way - each generator having its own strategy. For
example the `gen_app` generator uses `Template` data structures except for test code generation that
use a different kind of structs. Other generators rely on functions called at generation time to
perform the transformations which makes it hard to follow.

The proposal is to implement a standard strategy of first creating explicit structs representing the
intermediary representation used by the code generation templates. This would help clarify what the
templates expect and keep that up-to-date as the definitions evolve. This should be done in a way
that doesn't force generators to use it (some simple generators may not have a need for such a IR)
but it should encourage it. For example there could be an additional optional interface that
generators could implement that would get called by the generation engine.

## Cleanup Type System

Move `DateTime` and `UUID` to formats on the `String` type. This is to keep the set of basic types
small and remain consistent with how `Integer` support formats for various underlying Go data types
(see below). This also will allow removing a lot of cruft that supporting these types introduced.
From the need to "alias" them to string in code generators to the special cases they introduce with
gopherjs support.

## Protobuf / gRPC

Look into generating .proto files from the design and invoke `protoc` on them during code
generation. Integrate the generated code with the generated data structures.

This also requires the ability to specify the bit length of numerical values. The proposal here
would be to add support for the `Format` DSL to the `Integer` and `Number` primitive types. The
format would indicate bitness and whether the generated integer should be signed.

## Client Improvements

Try to remove the signers from the client package and instead make it possible to integrate with 3rd
party client packages. Make the client function generate http.Request to enable that integration.
