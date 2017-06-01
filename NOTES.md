## NonZeroAttributes

This is used in v1 to mark attributes that are used to define path parameters.
The ideas was that such attributes cannot have the zero value because by definition
they get initialized if a path matches. However the implementation is pretty ugly
and pervasive. More importantly there is actually a case where such a parameter
may be nil: when an action has multiple routes and some routes have path parameters
that others don't.

So in v2 the `AttributeExpr` data structure does not contain a `NonZeroAttributes`
field anymore.

## No HTTP Request or Response in Context

goa v1 makes it possible to access the underlying HTTP request and response
directly from the context. This is convenient mainly to deal with HTTP headers
(read them and write them) since the request body is already read.

In v2 the header handling is made explicit in the design removing the need for
the above. Other needs for accessing the request and response seem to fall in
the realm of middlewares which can be endpoint specific in v2. On top of this
carrying the HTTP request and response in the context is problematic as the
context may be passed down to modules that have no business writing to the HTTP
response for example.

v2 does not set the HTTP request and response in the context.

## Optional Payload

There is no such thing as an optional payload. It is possible to define an
object payload with optional attributes and an optional attribute may be used to
define the content of the request body or any of its header or params.

The payload may also be design.Empty for methods that don't take arguments.
