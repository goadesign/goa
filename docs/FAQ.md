# When is a generated struct field a pointer?

There are a few considerations taken into account by the code generation
algorithms to decide whether a generated struct field should be a pointer or
not. The goal is to avoid using pointers when not necessary as they tend to
complicate code and provide opportunity for errors. This discussion only affects
attributes using one of the primitive types. Fields that correspond to
attributes that are objects always use pointers. Fields that correspond to
attributes that are arrays or maps never use pointers.

The general idea is that if a type design specifies that a certain attribute is
required or has a default value then the corresponding field should never be nil
and therefore does not need to be a pointer. However the generated code that
decodes incoming HTTP requests or responses must account for the fact that these
fields may be missing (the request or response is malformed) and thus have to
use data structures that use pointers for all fields to test for nil values in
the unmarshaled data.

The table below lists whether fields generated for user type attributes that are
primitives use pointers (\*) or direct values (-).

* (s) : data structure generated for the server
* (c) : data structure generated for the client

| Properties / Data Structure | Payload / Result | Req. Body (s) | Resp. Body (s) | Req. Body (c) | Resp. Body (c) |
------------------------------|------------------|---------------|----------------|---------------|----------------|
| Required OR Default         | -                | *             | -              | -             | *              |
| Not Required, No Default    | *                | *             | *              | *             | *              |

# How are default values used?

The DSL allows for specifying default values for attributes. The default values
are used in two places by the code generators.

When generating marshaling code (server side to marshal the response or client
side to marshal the request) the default value is used to initialize the data
structure field if it is nil. As discussed in the previous section this cannot
happen if the attribute is defined with a primitive type since in this case the
field is not a pointer. However this can happen for attributes that are arrays
or maps.

When generating unmarshaling code (server side to unmarshal an incoming request
or client side to unmarshal a response) the default value is used to set the
value of missing fields. Note that if the attribute is required then the
generated code returns an error if the corresponding field is missing. So this
only applies for non required attributes with default values.
