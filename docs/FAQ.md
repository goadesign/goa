# When are validations defined in the design enforced?

There is a trade-off between performance and robustness wrt enforcing the validations
defined in the service design. goa trusts that the user code does the right thing
and only validates external data ("input"). This means goa validates incoming 
requests server side and responses client side. This way your code is always
guaranteed to get valid data but doesn't have to pay the price of validation
for each response being written server side or request being sent client side.

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

# How are views for a result type computed?

Views can be defined on a result type. If a method returns a result type
* the service method returns an extra view along with the result and error if
  the result type has more than one view. The generated endpoint function uses
  this view to create a viewed result type.
* a views package is generated at the service level which defines a viewed
  result type for each method result. This viewed result type has identical
  field names and types but uses pointers for all fields so that view specific
  validation logic may be generated. Constructors are generated in the service
  package to convert a result type to a viewed result type and vice versa.

The server side response marshaling code marshals the viewed result type
returned by the endpoint into a server type omitting any nil attributes.
The view used to render the result type is passed to the client in "Goa-View"
header.

The client side response unmarshaling code unmarshals the response into the
client type which is then transformed to the viewed result type and sets the
view attribute of the viewed result type from the Goa-View header. It validates
the attributes in the viewed result type as defined by the view and converts
the viewed result type into the service result type using the appropriate
constructor.

NOTE: If a result type is defined without any views, a "default" view is added
to the result type by goa. If you don't care about views, you can define a
method result using the `Type` DSL which will bypass the view-specific logic.
