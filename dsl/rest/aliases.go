//************************************************************************//
// Aliased goa DSL Functions
//
// Generated with aliaser
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package rest

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/dsl"
)

// API defines a network service API. It provides the API name, description and other global
// properties. There may only be one API declaration in a given design package.
//
// API is a top level DSL.
// API takes two arguments: the name of the API and the defining DSL.
//
// Example:
//
//    var _ = API("adder", func() {
//        Title("title")                // Title used in documentation
//        Description("description")    // Description used in documentation
//        Version("2.0")                // Version of API
//        TermsOfService("terms")       // Terms of use
//        Contact(func() {              // Contact info
//            Name("contact name")
//            Email("contact email")
//            URL("contact URL")
//        })
//        License(func() {              // License
//            Name("license name")
//            URL("license URL")
//        })
//        Docs(func() {                 // Documentation links
//            Description("doc description")
//            URL("doc URL")
//        })
//    }
//
func API(name string, fn func()) *design.APIExpr {
	return dsl.API(name, fn)
}

// ArrayOf creates an array type from its element type.
//
// ArrayOf may be used wherever types can.
// The first argument of ArrayOf is the type of the array elements specified by
// name or by reference.
// The second argument of ArrayOf is an optional function that defines
// validations for the array elements.
//
// Examples:
//
//    var Names = ArrayOf(String, func() {
//        Pattern("[a-zA-Z]+") // Validates elements of the array
//    })
//
//    var Account = Type("Account", func() {
//        Attribute("bottles", ArrayOf(Bottle), "Account bottles", func() {
//            MinLength(1) // Validates array as a whole
//        })
//    })
//
// Note: CollectionOf and ArrayOf both return array types. CollectionOf returns
// a media type where ArrayOf returns a user type. In general you want to use
// CollectionOf if the argument is a media type and ArrayOf if it is a user
// type.
func ArrayOf(v interface{}, fn ...func()) *design.Array {
	return dsl.ArrayOf(v, fn...)
}

// Attribute describes a field of an object.
//
// An attribute has a name, a type and optionally a default value, an example
// value and validation rules.
//
// The type of an attribute can be one of:
//
// * The primitive types Boolean, Float32, Float64, Int, Int32, Int64, UInt,
//   UInt32, UInt64, String or Bytes.
//
// * A user type defined via the Type function.
//
// * An array defined using the ArrayOf function.
//
// * An map defined using the MapOf function.
//
// * An object defined inline using Attribute to define the type fields
//   recursively.
//
// * The special type Any to indicate that the attribute may take any of the
//   types listed above.
//
// Attribute may appear in MediaType, Type, Attribute or Attributes.
//
// Attribute accepts one to four arguments, the valid usages of the function
// are:
//
//    Attribute(name)       // Attribute of type String with no description, no
//                          // validation, default or example value
//
//    Attribute(name, fn)   // Attribute of type object with inline field
//                          // definitions, description, validations, default
//                          // and/or example value
//
//    Attribute(name, type) // Attribute with no description, no validation,
//                          // no default or example value
//
//    Attribute(name, type, fn) // Attribute with description, validations,
//                              // default and/or example value
//
//    Attribute(name, type, description)     // Attribute with no validation,
//                                           // default or example value
//
//    Attribute(name, type, description, fn) // Attribute with description,
//                                           // validations, default and/or
//                                           // example value
//
// Where name is a string indicating the name of the attribute, type specifies
// the attribute type (see above for the possible values), description a string
// providing a human description of the attribute and fn the defining DSL if
// any.
//
// When defining the type inline using Attribute recursively the function takes
// the second form (name and DSL defining the type). The description can be
// provided using the Description function in this case.
//
// Examples:
//
//    Attribute("name")
//
//    Attribute("driver", Person)         // Use type defined with Type function
//
//    Attribute("driver", "Person")       // May also use the type name
//
//    Attribute("name", String, func() {
//        Pattern("^foo")                 // Adds a validation rule
//    })
//
//    Attribute("driver", Person, func() {
//        Required("name")                // Add required field to list of
//    })                                  // fields already required in Person
//
//    Attribute("name", String, func() {
//        Default("bob")                  // Sets a default value
//    })
//
//    Attribute("name", String, "name of driver") // Sets a description
//
//    Attribute("age", Int32, "description", func() {
//        Minimum(2)                       // Sets both a description and
//                                         // validations
//    })
//
// The definition below defines an attribute inline. The resulting type
// is an object with three attributes "name", "age" and "child". The "child"
// attribute is itself defined inline and has one child attribute "name".
//
//    Attribute("driver", func() {           // Define type inline
//        Description("Composite attribute") // Set description
//
//        Attribute("name", String)          // Child attribute
//        Attribute("age", Int32, func() {   // Another child attribute
//            Description("Age of driver")
//            Default(42)
//            Minimum(2)
//        })
//        Attribute("child", func() {        // Defines a child attribute
//            Attribute("name", String)      // Grand-child attribute
//            Required("name")
//        })
//
//        Required("name", "age")            // List required attributes
//    })
//
func Attribute(name string, args ...interface{}) {
	dsl.Attribute(name, args...)
}

// Attributes implements the media type Attributes DSL. See MediaType.
func Attributes(fn func()) {
	dsl.Attributes(fn)
}

// CollectionOf creates a collection media type from its element media type. A
// collection media type represents the content of responses that return a
// collection of values such as listings. The expression accepts an optional DSL
// as second argument that allows specifying which view(s) of the original media
// type apply.
//
// The resulting media type identifier is built from the element media type by
// appending the media type parameter "type" with value "collection".
//
// CollectionOf takes the element media type as first argument and an optional
// DSL as second argument.
// CollectionOf may appear wherever MediaType can.
//
// Example:
//
//     var DivisionResult = MediaType("application/vnd.goa.divresult", func() {
//         Attributes(func() {
//             Attribute("value", Float64)
//         })
//         View("default", func() {
//             Attribute("value")
//         })
//     })
//
//     var MultiResults = CollectionOf(DivisionResult)
//
func CollectionOf(v interface{}, adsl ...func()) *design.MediaTypeExpr {
	return dsl.CollectionOf(v, adsl...)
}

// Contact sets the API contact information.
func Contact(fn func()) {
	dsl.Contact(fn)
}

// ContentType sets the value of the Content-Type response header. By default
// the ID of the media type is used.
//
//    ContentType("application/json")
//
func ContentType(typ string) {
	dsl.ContentType(typ)
}

// Default sets the default value for an attribute.
func Default(def interface{}) {
	dsl.Default(def)
}

// Elem makes it possible to specify validations for array and map values.
func Elem(fn func()) {
	dsl.Elem(fn)
}

// Email sets the contact email.
func Email(email string) {
	dsl.Email(email)
}

// Endpoint defines a single service endpoint.
//
// Endpoint may appear in a Service expression.
// Endpoint takes two arguments: the name of the endpoint and the defining DSL.
//
// Example:
//
//    Endpoint("add", func() {
//        Description("The add endpoint returns the sum of A and B")
//        Docs(func() {
//            Description("Add docs")
//            URL("http//adder.goa.design/docs/actions/add")
//        })
//        Payload(Operands)
//        Result(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Endpoint(name string, fn func()) {
	dsl.Endpoint(name, fn)
}

// Enum adds a "enum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor76.
func Enum(vals ...interface{}) {
	dsl.Enum(vals...)
}

// Error describes an endpoint error response. The description includes a unique
// name (in the scope of the endpoint), an optional type, description and DSL
// that further describes the type. If no type is specified then the goa
// ErrorMedia type is used. The DSL syntax is identical to the Attribute DSL.
// Transport specific DSL may further describe the mapping between the error
// type attributes and the serialized response.
//
// goa has a few predefined error names for the common cases, see ErrBadRequest
// for example.
//
// Error may appear in the Service (to define error responses that apply to all
// the service endpoints) or Endpoint expressions.
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorMedia
//
//        // Endpoint which uses the default type for its response.
//        Endpoint("divide", func() {
//            Payload(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	dsl.Error(name, args...)
}

// Example provides an example value for a type, a parameter, a header or any
// attribute. Example supports two syntaxes, both syntaxes accept two arguments
// and in both cases the first argument is a summary describing the example. The
// second argument provides the value of the example either directly or via a
// DSL that can also specify a long description.
//
// If no example is explicitly provided then a random example is generated
// unless the "swagger:example" metadata is set to "false". See Metadata.
//
// Example may appear in a Attributes or Attribute expression DSL.
// Example takes two arguments: a summary and the example value or defining DSL.
//
// Examples:
//
//	Params(func() {
//		Param("ZipCode:zip-code", String, "Zip code filter", func() {
//			Example("Santa Barbara", "93111")
//		})
//	})
//
//	Attributes(func() {
//		Attribute("ID", Int64, "ID is the unique bottle identifier")
//		Example("The first bottle", func() {
//			Description("This bottle has an ID set to 1")
//			Value(Val{"ID": 1})
//		})
//		Example("Another bottle", func() {
//			Description("This bottle has an ID set to 5")
//			Value(Val{"ID": 5})
//		})
//	})
//
func Example(summary string, arg interface{}) {
	dsl.Example(summary, arg)
}

// Field is syntactic sugar to define an attribute with the "rpc:tag" metadata
// set with the value of the first argument.
//
// Field may appear wherever Attribute can.
// Field takes the same arguments as Attribute with the addition of the tag
// value as first argument.
//
// Example:
//
//     Field(1, "ID", String, func() {
//         Pattern("[0-9]+")
//     })
//
func Field(tag interface{}, name string, args ...interface{}) {
	dsl.Field(tag, name, args...)
}

// Format adds a "format" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor104.
// The formats supported by goa are:
//
// FormatDateTime: RFC3339 date time
//
// FormatEmail: RFC5322 email address
//
// FormatHostname: RFC1035 internet host name
//
// FormatIPv4, FormatIPv6, FormatIP: RFC2373 IPv4, IPv6 address or either
//
// FormatURI: RFC3986 URI
//
// FormatMAC: IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address
//
// FormatCIDR: RFC4632 or RFC4291 CIDR notation IP address
//
// FormatRegexp: RE2 regular expression
func Format(f design.ValidationFormat) {
	dsl.Format(f)
}

// Key makes it possible to specify validations for map keys.
func Key(fn func()) {
	dsl.Key(fn)
}

// License sets the API license information.
func License(fn func()) {
	dsl.License(fn)
}

// MapOf creates a map from its key and element types.
//
// MapOf may be used wherever types can.
// MapOf takes two arguments: the key and value types either by name of by reference.
//
// Example:
//
//    var ReviewByID = MapOf(Int64, String, func() {
//        Key(func() {
//            Minimum(1)           // Validates keys of the map
//        })
//        Value(func() {
//            Pattern("[a-zA-Z]+") // Validates values of the map
//        })
//    })
//
//    var Review = Type("Review", func() {
//        Attribute("ratings", MapOf(Bottle, Int32), "Bottle ratings")
//    })
//
func MapOf(k, v interface{}, fn ...func()) *design.Map {
	return dsl.MapOf(k, v, fn...)
}

// MaxLength adds a "maxItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor42.
func MaxLength(val int) {
	dsl.MaxLength(val)
}

// Maximum adds a "maximum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor17.
func Maximum(val interface{}) {
	dsl.Maximum(val)
}

// MediaType defines a media type used to describe an endpoint response.
//
// Media types have a unique identifier as described in RFC6838. The identifier
// defines the default value for the Content-Type header of HTTP responses.
//
// The media type expression includes a listing of all the response attributes.
// Views specify which of the attributes are actually rendered so that the same
// media type expression may represent multiple rendering of a given response.
//
// All media types have a view named "default". This view is used to render the
// media type in responses when no other view is specified. If the default view
// is not explicitly described in the DSL then one is created that lists all the
// media type attributes.
//
// MediaType is a top level DSL.
// MediaType accepts two arguments: the media type identifier and the defining
// DSL.
//
// Example:
//
//    var BottleMT = MediaType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleMedia")         // Override generated type name
//        ContentType("application/json") // Override Content-Type header
//
//        Attributes(func() {
//            Attribute("id", Integer, "ID of bottle")
//            Attribute("href", String, "API href of bottle")
//            Attribute("account", Account, "Owner account")
//            Attribute("origin", Origin, "Details on wine origin")
//            Required("id", "href")
//        })
//
//        View("default", func() {        // Explicitly define default view
//            Attribute("id")
//            Attribute("href")
//        })
//
//        View("extended", func() {       // Define "extended" view
//            Attribute("id")
//            Attribute("href")
//            Attribute("account")
//            Attribute("origin")
//        })
//     })
//
func MediaType(identifier string, fn func()) *design.MediaTypeExpr {
	return dsl.MediaType(identifier, fn)
}

// MinLength adss a "minItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor45.
func MinLength(val int) {
	dsl.MinLength(val)
}

// Minimum adds a "minimum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor21.
func Minimum(val interface{}) {
	dsl.Minimum(val)
}

// Name sets the contact or license name.
func Name(name string) {
	dsl.Name(name)
}

// Pattern adds a "pattern" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor33.
func Pattern(p string) {
	dsl.Pattern(p)
}

// Payload defines the data type of an endpoint input. Payload also makes the
// input required. Use PayloadOpt to describe the type of an optional input.
//
// Payload may appear in a Endpoint expression.
//
// Payload takes one or two arguments. The first argument is either a type or a
// DSL function. If the first argument is a type then an optional DSL may be
// passed as second argument that further specializes the type by providing
// additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("save"), func() {
//	// Use primitive type.
//	Payload(String)
// }
//
// Endpoint("add", func() {
//     // Define payload data structure inline.
//     Payload(func() {
//         Attribute("left", Int32, "Left operand")
//         Attribute("right", Int32, "Left operand")
//         Required("left", "right")
//     })
// })
//
// Endpoint("add", func() {
//     // Define payload type by reference to user type.
//     Payload(Operands)
// })
//
// Endpoint("divide", func() {
//     // Specify additional required attributes on user type.
//     Payload(Operands, func() {
//         Required("left", "right")
//     })
// })
//
func Payload(val interface{}, fns ...func()) {
	dsl.Payload(val, fns...)
}

// PayloadOpt defines the data type of an endpoint input. PayloadOpt also makes
// the input optional. Use Payload to describe the type of a required input.
//
// See Payload for usage and examples.
func PayloadOpt(val interface{}, fns ...func()) {
	dsl.PayloadOpt(val, fns...)
}

// Reference sets a type or media type reference. The value itself can be a type
// or a media type.  The reference type attributes define the default properties
// for attributes with the same name in the type using the reference.
//
// Reference may be used in Type or MediaType.
// Reference accepts a single argument: the type or media type containing the
// attributes that define the default properties of the attributes of the type
// or media type that uses Reference.
//
// Example:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name", String, func() {
//			MinLength(3)
//		})
//		Attribute("vintage", Int32, func() {
//			Minimum(1970)
//		})
//		Attribute("somethingelse", String)
//	})
//
//	var BottleMedia = MediaType("vnd.goa.bottle", func() {
//		Reference(Bottle)
//		Attributes(func() {
//			Attribute("id", UInt64, "ID is the bottle identifier")
//
//                      // The type and validation of "name" and "vintage" are
//                      // inherited from the Bottle type "name" and "vintage"
//                      // attributes.
//			Attribute("name")
//			Attribute("vintage")
//		})
//	})
//
func Reference(t design.DataType) {
	dsl.Reference(t)
}

// Required adds a "required" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor61.
func Required(names ...string) {
	dsl.Required(names...)
}

// Result describes and endpoint result type.
//
// Result may appear in a Endpoint expression.
//
// Result accepts a type as first argument. This argument is optional in which
// case the type must be described inline (see below).
//
// Result accepts an optional DSL function as second argument. This function may
// define the result type inline using Attribute or may further specialize the
// type passed as first argument e.g. by providing additional validations (e.g.
// list of required attributes). The DSL may also specify a view when the first
// argument is a media type corresponding to the view rendered by this endpoint.
// Note that specifying a view when the result type is a media type is optional
// and only useful in cases the endpoint renders a single view.
//
// The valid syntax for Result is thus:
//
//    Result(dsltype)
//
//    Result(func())
//
//    Result(dsltype, func())
//
// Examples:
//
//    // Define result using primitive type
//    Endpoint("add", func() {
//        Result(Int32)
//    })
//
//    // Define result using object defined inline
//    Endpoint("add", func() {
//        Result(func() {
//            Attribute("value", Int32, "Resulting sum")
//            Required("value")
//        })
//    })
//
//    // Define result type using user type
//    Endpoint("add", func() {
//        Result(Sum)
//    })
//
//    // Specify view and required attributes on media type
//    Endpoint("add", func() {
//        Result(Sum, func() {
//            View("default")
//            Required("value")
//        })
//    })
//
func Result(val interface{}, fns ...func()) {
	dsl.Result(val, fns...)
}

// Server defines an API host.
func Server(url string, fn ...func()) {
	dsl.Server(url, fn...)
}

// Service defines a group of related endpoints. Refer to the transport specific
// DSLs to learn how to provide transport specific information.
//
// Service is as a top level expression.
// Service accepts two arguments: the name of the service (which must be unique
// in the design package) and its defining DSL.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Description("divider service") // Optional description
//
//        DefaultType(DivideResult) // Default response type for the service
//                                  // endpoints. Also defines default
//                                  // properties (type, description and
//                                  // validations) for attributes with
//                                  // identical names in request types.
//
//        Error("Unauthorized", Unauthorized) // Error response that applies to
//                                            // all endpoints
//
//        Endpoint("divide", func() {     // Defines a single endpoint
//            Description("The divide endpoint returns the division of A and B")
//            Request(DivideRequest)      // Request type listing all request
//                                        // parameters in its attributes.
//            Response(DivideResponse)    // Response type.
//            Error("DivisionByZero", DivByZero) // Error, has a name and
//                                               // optionally a type
//                                               // (DivByZero) describes the
//                                               // error response.
//        })
//    })
//
func Service(name string, fn func()) *design.ServiceExpr {
	return dsl.Service(name, fn)
}

// TermsOfService describes the API terms of services or links to them.
func TermsOfService(terms string) {
	dsl.TermsOfService(terms)
}

// Title sets the API title used by the generated documentation and code comments.
func Title(val string) {
	dsl.Title(val)
}

// Type defines a user type. A user type has a unique name and may be an alias
// to an existing type or may describe a completely new type using a list of
// attributes (object fields). Attribute types may themselves be user type.
// When a user type is defined as an alias to another type it may define
// additional validations - for example it a user type which is an alias of
// String may define a validation pattern that all instances of the type
// must match.
//
// Type is a top level definition.
//
// Type takes two or three arguments: the first argument is the name of the type.
// The name must be unique. The second argument is either another type or a
// function. If the second argument is a type then there may be a function passed
// as third argument.
//
// Example:
//
//     // simple alias
//     var MyString = Type("MyString", String)
//
//     // alias with description and additional validation
//     var Hostname = Type("Hostname", String, func() {
//         Description("A host name")
//         Format(FormatHostname)
//     })
//
//     // new type
//     var SumPayload = Type("SumPayload", func() {
//         Description("Type sent to add endpoint")
//
//         Attribute("a", String)                 // string attribute "a"
//         Attribute("b", Int32, "operand")       // attribute with description
//         Attribute("operands", ArrayOf(Int32))  // array attribute
//         Attribute("ops", MapOf(String, Int32)) // map attribute
//         Attribute("c", SumMod)                 // attribute using user type
//         Attribute("len", Int64, func() {       // attribute with validation
//             Minimum(1)
//         })
//
//         Required("a")                          // Required attributes
//         Required("b", "c")
//     })
//
func Type(name string, args ...interface{}) design.UserType {
	return dsl.Type(name, args...)
}

// TypeName makes it possible to set the Go struct name for a type or media type
// in the generated code. By default goagen uses the name (type) or identifier
// (media type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
func TypeName(name string) {
	dsl.TypeName(name)
}

// URL sets the contact, license or external documentation URL.
//
// URL may appear in Contact, License or Docs
// URL accepts a single argument which is the URL.
//
// Example:
//
//    Docs(func() {
//        Description("Additional information")
//        URL("https://goa.design")
//    })
//
func URL(url string) {
	dsl.URL(url)
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	dsl.Version(ver)
}

// View adds a new view to a media type. A view has a name and lists attributes
// that are rendered when the view is used to produce a response. The attribute
// names must appear in the media type expression. If an attribute is itself a
// media type then the view may specify which view to use when rendering the
// attribute using the View function in the View DSL. If not specified then the
// view named "default" is used. Examples:
//
//	View("default", func() {
//              // "id" and "name" must be media type attributes
//		Attribute("id")
//		Attribute("name")
//	})
//
//	View("extended", func() {
//		Attribute("id")
//		Attribute("name")
//		Attribute("origin", func() {
//			// Use view "extended" to render attribute "origin"
//			View("extended")
//		})
//	})
//
func View(name string, adsl ...func()) {
	dsl.View(name, adsl...)
}