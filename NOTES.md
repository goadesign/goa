## NonZeroAttributes

This is used in v1 to mark attributes that are used to define path parameters.
The ideas was that such attributes cannot have the zero value because by definition
they get initialized if a path matches. However the implementation is pretty ugly
and pervasive. More importantly there is actually a case where such a parameter
may be nil: when an action has multiple routes and some routes have path parameters
that others don't.

So in v2 the `AttributeExpr` data structure does not contain a `NonZeroAttributes`
field anymore.
