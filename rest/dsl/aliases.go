//************************************************************************//
// Aliased goa DSL Functions
//
// Generated with aliaser
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package dsl

import (
	"goa.design/goa.v2/design"
	goadsl "goa.design/goa.v2/dsl"
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
//        TermsOfAPI("terms")           // Terms of use
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
func API(name string, dsl func()) *design.APIExpr {
	return goadsl.API(name, dsl)
}

// ArrayOf creates an array type from its element type.
//
// ArrayOf may be used wherever types can.
// The first argument of ArrayOf is the type of the array elements specified by
// name or by reference.
// The second argument of ArrayOf is an optional DSL that defines validations
// for the array elements.
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
func ArrayOf(v interface{}, dsl ...func()) *design.Array {
	return goadsl.ArrayOf(v, dsl...)
}

// Attribute describes a field of an object.
//
// An attribute has a name, a type and optionally a default value, an example
// value and validation rules.
//
// The type of an attribute can be one of:
//
// * The primitive types Boolean, Float32, Float64, Int32, Int64, UInt32,
//   UInt64, String or Bytes.
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
//    Attribute(name, dsl)  // Attribute of type object with inline field
//                          // definitions, description, validations, default
//                          // and/or example value
//
//    Attribute(name, type) // Attribute with no description, no validation,
//                          // no default or example value
//
//    Attribute(name, type, dsl) // Attribute with description, validations,
//                               // default and/or example value
//
//    Attribute(name, type, description)      // Attribute with no validation,
//                                            // default or example value
//
//    Attribute(name, type, description, dsl) // Attribute with description,
//                                            // validations, default and/or
//                                            // example value
//
// Where name is a string indicating the name of the attribute, type specifies
// the attribute type (see above for the possible values), description a string
// providing a human description of the attribute and dsl the defining DSL if
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
	goadsl.Attribute(name, args...)
}

// Attributes implements the media type Attributes DSL. See MediaType.
func Attributes(dsl func()) {
	goadsl.Attributes(dsl)
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
	return goadsl.CollectionOf(v, adsl...)
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	goadsl.Contact(dsl)
}

// ContentType sets the value of the Content-Type response header. By default
// the ID of the media type is used.
//
//    ContentType("application/json")
//
func ContentType(typ string) {
	goadsl.ContentType(typ)
}

// Default sets the default value for an attribute.
func Default(def interface{}) {
	goadsl.Default(def)
}

// Email sets the contact email.
func Email(email string) {
	goadsl.Email(email)
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
//        Request(Operands)
//        Response(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Endpoint(name string, dsl func()) {
	goadsl.Endpoint(name, dsl)
}

// Enum adds a "enum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor76.
func Enum(vals ...interface{}) {
	goadsl.Enum(vals...)
}

// Error describes an endpoint error response. The description includes a unique
// name (in the scope of the endpoint), an optional type, description and DSL
// that further describes the type. If no type is specified then the goa
// ErrorMedia type is used. The DSL syntax is identical to the Attribute DSL.
// Transport specific DSL may further describe the mapping between the error
// type attributes and the serialized response.
//
// Error may appear in the Service (to define error responses that apply to all
// the service endpoints) or Endppoint expressions.
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorMedia
//
//        // Endpoint which uses the default type for its response.
//        Endpoint("divide", func() {
//            Request(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	goadsl.Error(name, args...)
}

// Example sets the example of an attribute to be used for the documentation.
// If no example is explicitly provided then a random example is generated
// unless the "swagger:example" metadata is set to "false". See Metadata.
//
// Example may appear in a Attribute expression.
// Example takes one argument: the example value.
//
// Example:
//
//	Attributes(func() {
//		Attribute("ID", Int64, func() {
//			Example(1)
//		})
//	})
//
func Example(exp interface{}) {
	goadsl.Example(exp)
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
	goadsl.Field(tag, name, args...)
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
	goadsl.Format(f)
}

// Key makes it possible to specify validations for map keys.
func Key(dsl func()) {
	goadsl.Key(dsl)
}

// License sets the API license information.
func License(dsl func()) {
	goadsl.License(dsl)
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
func MapOf(k, v interface{}, dsl ...func()) *design.Map {
	return goadsl.MapOf(k, v, dsl...)
}

// MaxLength adds a "maxItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor42.
func MaxLength(val int) {
	goadsl.MaxLength(val)
}

// Maximum adds a "maximum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor17.
func Maximum(val interface{}) {
	goadsl.Maximum(val)
}

// MediaType describes a media type used to desribe an endpoint response.
//
// Media types are defined with a unique identifier as described in RFC6838. The
// identifier defines the default value for the Content-Type header of HTTP
// responses.
//
// The media type expression includes a listing of all the response attributes.
// Views specify which of the attributes are actually rendered so that the same
// media type expression may represent multiple rendering of a given response.
//
// All media types must define a view named "default". This view is used to
// render the media type in responses when no other view is specified.
//
// MediaType is a top level DSL.
// MediaType accepts two arguments: the media type identifier and the defining
// DSL.
//
// Example:
//
//    var BottleMT = MediaType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleMedia")         // Override generated name
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
//        View("default", func() {        // Define default view
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
func MediaType(identifier string, dsl func()) *design.MediaTypeExpr {
	return goadsl.MediaType(identifier, dsl)
}

// MinLength adss a "minItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor45.
func MinLength(val int) {
	goadsl.MinLength(val)
}

// Minimum adds a "minimum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor21.
func Minimum(val interface{}) {
	goadsl.Minimum(val)
}

// Name sets the contact or license name.
func Name(name string) {
	goadsl.Name(name)
}

// Pattern adds a "pattern" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor33.
func Pattern(p string) {
	goadsl.Pattern(p)
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
	goadsl.Reference(t)
}

// Request defines the data type which lists the request parameters in its
// attributes. Transport specific DSL may provide a mapping between the
// attributes and incoming request state (e.g. which attributes are initialized
// from HTTP headers, query string values or body fields in the case of HTTP)
//
// Request may appear in a Endpoint expression.
//
// Request takes one or two arguments. The first argument is either a reference
// to a type, the name of a type or a DSL function.
// If the first argument is a type or the name of a type then an optional DSL
// may be passed as second argument that further specializes the type by
// providing additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("add", func() {
//     // Define request type inline
//     Request(func() {
//         Attribute("left", Int32, "Left operand")
//         Attribute("right", Int32, "Left operand")
//         Required("left", "right")
//     })
// })
//
// Endpoint("add", func() {
//     // Define request type by reference to user type
//     Request(Operands)
// })
//
// Endpoint("divide", func() {
//     // Specify required attributes on user type
//     Request(Operands, func() {
//         Required("left", "right")
//     })
// })
//
func Request(val interface{}, dsls ...func()) {
	goadsl.Request(val, dsls...)
}

// Required adds a "required" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor61.
func Required(names ...string) {
	goadsl.Required(names...)
}

// Server defines an API host.
func Server(url string, dsl ...func()) {
	goadsl.Server(url, dsl...)
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
func Service(name string, dsl func()) *design.ServiceExpr {
	return goadsl.Service(name, dsl)
}

// TermsOfAPI describes the API terms of services or links to them.
func TermsOfAPI(terms string) {
	goadsl.TermsOfAPI(terms)
}

// Title sets the API title used by the generated documentation and code comments.
func Title(val string) {
	goadsl.Title(val)
}

// Type describes a user type.
//
// Type is a top level definition.
// Type takes two arguments: the type name and the defining DSL.
//
// Example:
//
//     var SumPayload = Type("SumPayload", func() {
//         Description("Type sent to add endpoint")
//
//         Attribute("a", String)                 // string field "a"
//         Attribute("b", Int32, "operand")       // field with description
//         Attribute("operands", ArrayOf(Int32))  // array field
//         Attribute("ops", MapOf(String, Int32)) // map field
//         Attribute("c", SumMod)                 // field using user type
//         Attribute("len", Int64, func() {       // field with validation
//             Minimum(1)
//         })
//
//         Required("a")                          // Required fields
//         Required("b", "c")
//     })
//
func Type(name string, dsl func()) design.UserType {
	return goadsl.Type(name, dsl)
}

// TypeName makes it possible to set the Go struct name for a type or media type
// in the generated code. By default goagen uses the name (type) or identifier
// (media type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
func TypeName(name string) {
	goadsl.TypeName(name)
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
	goadsl.URL(url)
}

// Value makes it possible to specify validations for map values.
func Value(dsl func()) {
	goadsl.Value(dsl)
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	goadsl.Version(ver)
}

// View adds a new view to a media type. A view has a name and lists attributes
// that are rendered when the view is used to produce a response. The attribute
// names must appear in the media type expression. If an attribute is itself a
// media type then the view may specify which view to use when rendering the
// attribute using the View function in the View adsl. If not specified then the
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
	goadsl.View(name, adsl...)
}