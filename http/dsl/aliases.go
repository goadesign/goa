//************************************************************************//
// Code generated with aliaser, DO NOT EDIT.
//
// Aliased DSL Functions
//************************************************************************//

package dsl

import (
	"goa.design/goa/design"
	dsl "goa.design/goa/dsl"
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

// APIKey defines the attribute used to provide the API key to an endpoint
// secured with API keys. The parameters and usage of APIKey are the same as the
// goa DSL Attribute function except that it accepts an extra first argument
// corresponding to the name of the API key security scheme.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to set the API key value.
//
// APIKey must appear in Payload or Type.
//
// Example:
//
//    Method("secured_read", func() {
//        Security(APIKeyAuth)
//        Payload(func() {
//            APIKey("api_key", "key", String, "API key used to perform authorization")
//            Required("key")
//        })
//        Result(String)
//        HTTP(func() {
//            GET("/")
//            Param("key:k") // Provide the key as a query string param "k"
//        })
//    })
//
//    Method("secured_write", func() {
//        Security(APIKeyAuth)
//        Payload(func() {
//            APIKey("api_key", "key", String, "API key used to perform authorization")
//            Attribute("data", String, "Data to be written")
//            Required("key", "data")
//        })
//        HTTP(func() {
//            POST("/")
//            Header("key:Authorization") // Provide the key in Authorization header (default)
//        })
//    })
//
func APIKey(scheme, name string, args ...interface{}) {
	dsl.APIKey(scheme, name, args...)
}

// APIKeySecurity defines an API key security scheme where a key must be
// provided by the client to perform authorization.
//
// APIKeySecurity is a top level DSL.
//
// APIKeySecurity takes a name as first argument and an optional DSL as
// second argument.
//
// Example:
//
//    var APIKey = APIKeySecurity("key", func() {
//          Description("Shared secret")
//    })
//
func APIKeySecurity(name string, fn ...func()) *design.SchemeExpr {
	return dsl.APIKeySecurity(name, fn...)
}

// AccessToken defines the attribute used to provide the access token to an
// endpoint secured with OAuth2. The parameters and usage of AccessToken are the
// same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
//
// AccessToken must appear in Payload or Type.
//
// Example:
//
//    Method("secured", func() {
//        Security(OAuth2)
//        Payload(func() {
//            AccessToken("token", String, "OAuth2 access token used to perform authorization")
//            Required("token")
//        })
//        Result(String)
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            GET("/")
//        })
//    })
//
func AccessToken(name string, args ...interface{}) {
	dsl.AccessToken(name, args...)
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
// a result type where ArrayOf returns a user type. In general you want to use
// CollectionOf if the argument is a result type and ArrayOf if it is a user
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
// Attribute must appear in ResultType, Type, Attribute or Attributes.
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

// Attributes implements the result type Attributes DSL. See ResultType.
func Attributes(fn func()) {
	dsl.Attributes(fn)
}

// AuthorizationCodeFlow defines an authorizationCode OAuth2 flow as described
// in section 1.3.1 of RFC 6749.
//
// AuthorizationCodeFlow must be used in OAuth2Security.
//
// AuthorizationCodeFlow accepts three arguments: the authorization, token and
// refresh URLs.
func AuthorizationCodeFlow(authorizationURL, tokenURL, refreshURL string) {
	dsl.AuthorizationCodeFlow(authorizationURL, tokenURL, refreshURL)
}

// BasicAuthSecurity defines a basic authentication security scheme.
//
// BasicAuthSecurity is a top level DSL.
//
// BasicAuthSecurity takes a name as first argument and an optional DSL as
// second argument.
//
// Example:
//
//     var Basic = BasicAuthSecurity("basicauth", func() {
//         Description("Use your own password!")
//     })
//
func BasicAuthSecurity(name string, fn ...func()) *design.SchemeExpr {
	return dsl.BasicAuthSecurity(name, fn...)
}

// ClientCredentialsFlow defines an clientCredentials OAuth2 flow as described
// in section 1.3.4 of RFC 6749.
//
// ClientCredentialsFlow must be used in OAuth2Security.
//
// ClientCredentialsFlow accepts two arguments: the token and refresh URLs.
func ClientCredentialsFlow(tokenURL, refreshURL string) {
	dsl.ClientCredentialsFlow(tokenURL, refreshURL)
}

// CollectionOf creates a collection result type from its element result type. A
// collection result type represents the content of responses that return a
// collection of values such as listings. The expression accepts an optional DSL
// as second argument that allows specifying which view(s) of the original result
// type apply.
//
// The resulting result type identifier is built from the element result type by
// appending the result type parameter "type" with value "collection".
//
// CollectionOf must appear wherever ResultType can.
//
// CollectionOf takes the element result type as first argument and an optional
// DSL as second argument.
//
// Example:
//
//     var DivisionResult = ResultType("application/vnd.goa.divresult", func() {
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
func CollectionOf(v interface{}, adsl ...func()) *design.ResultTypeExpr {
	return dsl.CollectionOf(v, adsl...)
}

// Contact sets the API contact information.
func Contact(fn func()) {
	dsl.Contact(fn)
}

// ConvertTo specifies an external type that instances of the generated struct
// are converted into. The generated struct is equipped with a method that makes
// it possible to instantiate the external type. The default algorithm used to
// match the external type fields to the design attributes is as follows:
//
//    1. Look for an attribute with the same name as the field
//    2. Look for an attribute with the same name as the field but with the
//       first letter being lowercase
//    3. Look for an attribute with a name corresponding to the snake_case
//       version of the field name
//
// This algorithm does not apply if the attribute is equipped with the
// "struct.field.external" metadata. In this case the matching is done by
// looking up the field with a name corresponding to the value of the metadata.
// If the value of the metadata is "-" the attribute isn't matched and no
// conversion code is generated for it. In all other cases it is an error if no
// match is found or if the matching field type does not correspond to the
// attribute type.
//
// ConvertTo must appear in Type or ResutType.
//
// ConvertTo accepts one arguments: an instance of the external type.
//
// Example:
//
// Service design:
//
//    var Bottle = Type("bottle", func() {
//        Description("A bottle")
//        ConvertTo(models.Bottle{})
//        // The "rating" attribute is matched to the external
//        // typ "Rating" field.
//        Attribute("rating", Int)
//        Attribute("name", String, func() {
//            // The "name" attribute is matched to the external
//            // type "MyName" field.
//            Metadata("struct.field.external", "MyName")
//        })
//        Attribute("vineyard", String, func() {
//            // The "vineyard" attribute is not converted.
//            Metadata("struct.field.external", "-")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        Rating int
//        // Mapped field
//        MyName string
//        // Additional fields are OK
//        Description string
//    }
//
func ConvertTo(obj interface{}) {
	dsl.ConvertTo(obj)
}

// CreateFrom specifies an external type that instances of the generated struct
// can be initialized from. The generated struct is equipped with a method that
// initializes its fields from an instance of the external type. The default
// algorithm used to match the external type fields to the design attributes is
// as follows:
//
//    1. Look for an attribute with the same name as the field
//    2. Look for an attribute with the same name as the field but with the
//       first letter being lowercase
//    3. Look for an attribute with a name corresponding to the snake_case
//       version of the field name
//
// This algorithm does not apply if the attribute is equipped with the
// "struct.field.external" metadata. In this case the matching is done by
// looking up the field with a name corresponding to the value of the metadata.
// If the value of the metadata is "-" the attribute isn't matched and no
// conversion code is generated for it. In all other cases it is an error if no
// match is found or if the matching field type does not correspond to the
// attribute type.
//
// CreateFrom must appear in Type or ResutType.
//
// CreateFrom accepts one arguments: an instance of the external type.
//
// Example:
//
// Service design:
//
//    var Bottle = Type("bottle", func() {
//        Description("A bottle")
//        CreateFrom(models.Bottle{})
//        Attribute("rating", Int)
//        Attribute("name", String, func() {
//            // The "name" attribute is matched to the external
//            // type "MyName" field.
//            Metadata("struct.field.external", "MyName")
//        })
//        Attribute("vineyard", String, func() {
//            // The "vineyard" attribute is not initialized by the
//            // generated constructor method.
//            Metadata("struct.field.external", "-")
//        })
//    })
//
// External (i.e. non design) package:
//
//    package model
//
//    type Bottle struct {
//        Rating int
//        // Mapped field
//        MyName string
//        // Additional fields are OK
//        Description string
//    }
//
func CreateFrom(obj interface{}) {
	dsl.CreateFrom(obj)
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

// Enum adds a "enum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor76.
func Enum(vals ...interface{}) {
	dsl.Enum(vals...)
}

// Error describes a method error return value. The description includes a
// unique name (in the scope of the method), an optional type, description and
// DSL that further describes the type. If no type is specified then the
// built-in ErrorResult type is used. The DSL syntax is identical to the
// Attribute DSL.
//
// Error must appear in the Service (to define error responses that apply to all
// the service methods) or Method expressions.
//
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorResult
//
//        // Method which uses the default type for its response.
//        Method("divide", func() {
//            Payload(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	dsl.Error(name, args...)
}

// Example provides an example value for a type, a parameter, a header or any
// attribute. Example supports two syntaxes: one syntax accepts two arguments
// where the first argument is a summary describing the example and the second a
// value provided directly or via a DSL which may also specify a long
// description. The other syntax accepts a single argument and is equivalent to
// using the first syntax where the summary is the string "default".
//
// If no example is explicitly provided in an attribute expression then a random
// example is generated unless the "swagger:example" metadata is set to "false".
// See Metadata.
//
// Example must appear in a Attributes or Attribute expression DSL.
//
// Example takes one or two arguments: an optional summary and the example value
// or defining DSL.
//
// Examples:
//
//	Params(func() {
//		Param("ZipCode:zip-code", String, "Zip code filter", func() {
//			Example("Santa Barbara", "93111")
//			Example("93117") // same as Example("default", "93117")
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
func Example(args ...interface{}) {
	dsl.Example(args...)
}

// Extend adds the parameter type attributes to the type using Extend. The
// parameter type must be an object.
//
// Extend may be used in Type or ResultType. Extend accepts a single argument:
// the type or result type containing the attributes to be copied.
//
// Example:
//
//    var CreateBottlePayload = Type("CreateBottlePayload", func() {
//       Attribute("name", String, func() {
//          MinLength(3)
//       })
//       Attribute("vintage", Int32, func() {
//          Minimum(1970)
//       })
//    })
//
//    var UpdateBottlePayload = Type("UpatePayload", func() {
//        Atribute("id", String, "ID of bottle to update")
//        Extend(CreateBottlePayload) // Adds attributes "name" and "vintage"
//    })
//
func Extend(t design.DataType) {
	dsl.Extend(t)
}

// Field is syntactic sugar to define an attribute with the "rpc:tag" metadata
// set with the value of the first argument.
//
// Field must appear wherever Attribute can.
//
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
// FormatUUID: RFC4122 uuid
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
//
// FormatJSON: JSON text
//
// FormatRFC1123: RFC1123 date time
//
func Format(f design.ValidationFormat) {
	dsl.Format(f)
}

// ImplicitFlow defines an implicit OAuth2 flow as described in section 1.3.2
// of RFC 6749.
//
// ImplicitFlow must be used in OAuth2Security.
//
// ImplicitFlow accepts two arguments: the authorization and refresh URLs.
func ImplicitFlow(authorizationURL, refreshURL string) {
	dsl.ImplicitFlow(authorizationURL, refreshURL)
}

// JWTSecurity defines an HTTP security scheme where a JWT is passed in the
// request Authorization header as a bearer token to perform auth. This scheme
// supports defining scopes that endpoint may require to authorize the request.
// The scheme also supports specifying a token URL used to retrieve token
// values.
//
// Since scopes are not compatible with the Swagger specification, the swagger
// generator inserts comments in the description of the different elements on
// which they are defined.
//
// JWTSecurity is a top level DSL.
//
// JWTSecurity takes a name as first argument and an optional DSL as second
// argument.
//
// Example:
//
//    var JWT = JWTSecurity("jwt", func() {
//        Scope("system:write", "Write to the system")
//        Scope("system:read", "Read anything in there")
//    })
//
func JWTSecurity(name string, fn ...func()) *design.SchemeExpr {
	return dsl.JWTSecurity(name, fn...)
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

// Method defines a single service method.
//
// Method must appear in a Service expression.
//
// Method takes two arguments: the name of the method and the defining DSL.
//
// Example:
//
//    Method("add", func() {
//        Description("The add method returns the sum of A and B")
//        Docs(func() {
//            Description("Add docs")
//            URL("http//adder.goa.design/docs/endpoints/add")
//        })
//        Payload(Operands)
//        Result(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Method(name string, fn func()) {
	dsl.Method(name, fn)
}

// MinLength adds a "minItems" validation to the attribute.
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

// NoSecurity removes the need for an endpoint to perform authorization.
//
// NoSecurity must appear in Method.
func NoSecurity() {
	dsl.NoSecurity()
}

// OAuth2Security defines an OAuth2 security scheme. The DSL provided as second
// argument defines the specific flows supported by the scheme. The supported
// flow types are ImplicitFlow, PasswordFlow, ClientCredentialsFlow, and
// AuthorizationCodeFlow. The DSL also defines the scopes that may be
// associated with the incoming request tokens.
//
// OAuth2Security is a top level DSL.
//
// OAuth2Security takes a name as first argument and a DSL as second argument.
//
// Example:
//
//    var OAuth2 = OAuth2Security("googauth", func() {
//        ImplicitFlow("/authorization")
//
//        Scope("api:write", "Write acess")
//        Scope("api:read", "Read access")
//    })
//
func OAuth2Security(name string, fn ...func()) *design.SchemeExpr {
	return dsl.OAuth2Security(name, fn...)
}

// Password defines the attribute used to provide the password to an endpoint
// secured with basic authentication. The parameters and usage of Password are
// the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to compute the basic authentication Authorization header value.
//
// Password must appear in Payload or Type.
//
// Example:
//
//    Method("login", func() {
//        Security(Basic)
//        Payload(func() {
//            Username("user", String)
//            Password("pass", String)
//        })
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            POST("/login")
//        })
//    })
//
func Password(name string, args ...interface{}) {
	dsl.Password(name, args...)
}

// PasswordFlow defines an Resource Owner Password Credentials OAuth2 flow as
// described in section 1.3.3 of RFC 6749.
//
// PasswordFlow must be used in OAuth2Security.
//
// PasswordFlow accepts two arguments: the token and refresh URLs.
func PasswordFlow(tokenURL, refreshURL string) {
	dsl.PasswordFlow(tokenURL, refreshURL)
}

// Pattern adds a "pattern" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor33.
func Pattern(p string) {
	dsl.Pattern(p)
}

// Payload defines the data type of an method input. Payload also makes the
// input required.
//
// Payload must appear in a Method expression.
//
// Payload takes one to three arguments. The first argument is either a type or
// a DSL function. If the first argument is a type then an optional description
// may be passed as second argument. Finally a DSL may be passed as last
// argument that further specializes the type by providing additional
// validations (e.g. list of required attributes)
//
// The valid usage for Payload are thus:
//
//    Payload(Type)
//
//    Payload(func())
//
//    Payload(Type, "description")
//
//    Payload(Type, func())
//
//    Payload(Type, "description", func())
//
// Examples:
//
//    Method("upper"), func() {
//        // Use primitive type.
//        Payload(String)
//    }
//
//    Method("upper"), func() {
//        // Use primitive type.and description
//        Payload(String, "string to convert to uppercase")
//    }
//
//    Method("upper"), func() {
//        // Use primitive type, description and validations
//        Payload(String, "string to convert to uppercase", func() {
//            Pattern("^[a-z]")
//        })
//    }
//
//    Method("add", func() {
//        // Define payload data structure inline
//        Payload(func() {
//            Description("Left and right operands to add")
//            Attribute("left", Int32, "Left operand")
//            Attribute("right", Int32, "Left operand")
//            Required("left", "right")
//        })
//    })
//
//    Method("add", func() {
//        // Define payload type by reference to user type
//        Payload(Operands)
//    })
//
//    Method("divide", func() {
//        // Specify additional required attributes on user type.
//        Payload(Operands, func() {
//            Required("left", "right")
//        })
//    })
//
func Payload(val interface{}, args ...interface{}) {
	dsl.Payload(val, args...)
}

// Reference sets a type or result type reference. The value itself can be a
// type or a result type. The reference type attributes define the default
// properties for attributes with the same name in the type using the reference.
//
// Reference may be used in Type or ResultType, it may appear multiple times in
// which case attributes are looked up in each reference in order of appearance
// in the DSL.
//
// Reference accepts a single argument: the type or result type containing the
// attributes that define the default properties of the attributes of the type
// or result type that uses Reference.
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
//	var BottleResult = ResultType("vnd.goa.bottle", func() {
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

// Result defines the data type of a method output.
//
// Result must appear in a Method expression.
//
// Result takes one to three arguments. The first argument is either a type or a
// DSL function. If the first argument is a type then an optional description
// may be passed as second argument. Finally a DSL may be passed as last
// argument that further specializes the type by providing additional
// validations (e.g. list of required attributes) The DSL may also specify a
// view when the first argument is a result type corresponding to the view
// rendered by this method. If no view is specified then the generated code
// defines response methods for all views.
//
// The valid syntax for Result is thus:
//
//    Result(Type)
//
//    Result(func())
//
//    Result(Type, "description")
//
//    Result(Type, func())
//
//    Result(Type, "description", func())
//
// Examples:
//
//    // Define result using primitive type
//    Method("add", func() {
//        Result(Int32)
//    })
//
//    // Define result using primitive type and description
//    Method("add", func() {
//        Result(Int32, "Resulting sum")
//    })
//
//    // Define result using primitive type, description and validations.
//    Method("add", func() {
//        Result(Int32, "Resulting sum", func() {
//            Minimum(0)
//        })
//    })
//
//    // Define result using object defined inline
//    Method("add", func() {
//        Result(func() {
//            Description("Result defines a single field which is the sum.")
//            Attribute("value", Int32, "Resulting sum")
//            Required("value")
//        })
//    })
//
//    // Define result type using user type
//    Method("add", func() {
//        Result(Sum)
//    })
//
//    // Specify view and required attributes on result type
//    Method("add", func() {
//        Result(Sum, func() {
//            View("default")
//            Required("value")
//        })
//    })
//
func Result(val interface{}, args ...interface{}) {
	dsl.Result(val, args...)
}

// ResultType defines a result type used to describe a method response.
//
// Result types have a unique identifier as described in RFC 6838. The
// identifier defines the default value for the Content-Type header of HTTP
// responses.
//
// The result type expression includes a listing of all the response attributes.
// Views specify which of the attributes are actually rendered so that the same
// result type expression may represent multiple rendering of a given response.
//
// All result types have a view named "default". This view is used to render the
// result type in responses when no other view is specified. If the default view
// is not explicitly described in the DSL then one is created that lists all the
// result type attributes.
//
// ResultType is a top level DSL.
//
// ResultType accepts two arguments: the result type identifier and the defining
// DSL.
//
// Example:
//
//    var BottleMT = ResultType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleResult")         // Override generated type name
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
func ResultType(identifier string, fn func()) *design.ResultTypeExpr {
	return dsl.ResultType(identifier, fn)
}

// Scope has two uses: in JWTSecurity or OAuth2Security it defines a scope
// supported by the scheme. In Security it lists required scopes.
//
// Scope must appear in Security, JWTSecurity or OAuth2Security.
//
// Scope accepts one or two arguments: the first argument is the scope name and
// when used in JWTSecurity or OAuth2Security the second argument is a
// description.
//
// Example:
//
//    var JWT = JWTSecurity("JWT", func() {
//        Scope("api:read", "Read access") // Defines a scope
//        Scope("api:write", "Write access")
//    })
//
//    Method("secured", func() {
//        Security(JWT, func() {
//            Scope("api:read") // Required scope for auth
//        })
//    })
//
func Scope(name string, desc ...string) {
	dsl.Scope(name, desc...)
}

// Security defines authentication requirements to access an API, a service or a
// service method.
//
// The requirement refers to one or more OAuth2Security, BasicAuthSecurity,
// APIKeySecurity or JWTSecurity security scheme. If the schemes include a
// OAuth2Security or JWTSecurity scheme then required scopes may be listed by
// name in the Security DSL. All the listed schemes must be validated by the
// client for the request to be authorized. Security may appear multiple times
// in the same scope in which case the client may validate any one of the
// requirements for the request to be authorized.
//
// Security must appear in a API, Service or Method expression.
//
// Security accepts an arbitrary number of security schemes as argument
// specified by name or by reference and an optional DSL function as last
// argument.
//
// Examples:
//
//    var _ = API("calc", func() {
//        // All API endpoints are secured via basic auth by default.
//        Security(BasicAuth)
//    })
//
//    var _ = Service("calculator", func() {
//        // Override default API security requirements. Accept either basic
//        // auth or OAuth2 access token with "api:read" scope.
//        Security(BasicAuth)
//        Security("oauth2", func() {
//            Scope("api:read")
//        })
//
//        Method("add", func() {
//            Description("Add two operands")
//
//            // Override default service security requirements. Require
//            // both basic auth and OAuth2 access token with "api:write"
//            // scope.
//            Security(BasicAuth, "oauth2", func() {
//                Scope("api:write")
//            })
//
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorResult)
//        })
//
//        Method("health-check", func() {
//            Description("Check health")
//
//            // Remove need for authorization for this endpoint.
//            NoSecurity()
//
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorResult)
//        })
//    })
//
func Security(args ...interface{}) {
	dsl.Security(args...)
}

// Server defines an API host.
func Server(url string, fn ...func()) {
	dsl.Server(url, fn...)
}

// Service defines a group of related methods. Refer to the transport specific
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
//                                  // methods. Also defines default
//                                  // properties (type, description and
//                                  // validations) for attributes with
//                                  // identical names in request types.
//
//        Error("Unauthorized", Unauthorized) // Error response that applies to
//                                            // all methods
//
//        Method("divide", func() {     // Defines a single method
//            Description("The divide method returns the division of A and B")
//            Request(DivideRequest)    // Request type listing all request
//                                      // parameters in its attributes.
//            Response(DivideResponse)  // Response type.
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

// Temporary qualifies an error type as describing temporary (i.e. retryable)
// errors.
//
// Temporary must appear in a Error expression.
//
// Temporary takes no argument.
//
// Example:
//
// var _ = Service("divider", func() {
//      Error("request_timeout", func() {
//              Temporary()
//      })
// })
func Temporary() {
	dsl.Temporary()
}

// TermsOfService describes the API terms of services or links to them.
func TermsOfService(terms string) {
	dsl.TermsOfService(terms)
}

// Timeout qualifies an error type as describing errors due to timeouts.
//
// Timeout must appear in a Error expression.
//
// Timeout takes no argument.
//
// Example:
//
// var _ = Service("divider", func() {
//	Error("request_timeout", func() {
//		Timeout()
//	})
// })
func Timeout() {
	dsl.Timeout()
}

// Title sets the API title used by the generated documentation and code comments.
func Title(val string) {
	dsl.Title(val)
}

// Token defines the attribute used to provide the JWT to an endpoint secured
// via JWT. The parameters and usage of Token are the same as the goa DSL
// Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
//
// Example:
//
//    Method("secured", func() {
//        Security(JWT)
//        Payload(func() {
//            Token("token", String, "JWT token used to perform authorization")
//            Required("token")
//        })
//        Result(String)
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            GET("/")
//        })
//    })
//
func Token(name string, args ...interface{}) {
	dsl.Token(name, args...)
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
//         Description("Type sent to add method")
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

// TypeName makes it possible to set the Go struct name for a type or result
// type in the generated code. By default goa uses the name (type) or identifier
// (result type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
func TypeName(name string) {
	dsl.TypeName(name)
}

// URL sets the contact, license or external documentation URL.
//
// URL must appear in Contact, License or Docs
//
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

// Username defines the attribute used to provide the username to an endpoint
// secured with basic authentication. The parameters and usage of Username are
// the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to compute the basic authentication Authorization header value.
//
// Username must appear in Payload or Type.
//
// Example:
//
//    Method("login", func() {
//        Security(Basic)
//        Payload(func() {
//            Username("user", String)
//            Password("pass", String)
//        })
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            POST("/login")
//        })
//    })
//
func Username(name string, args ...interface{}) {
	dsl.Username(name, args...)
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	dsl.Version(ver)
}

// View adds a new view to a result type. A view has a name and lists attributes
// that are rendered when the view is used to produce a response. The attribute
// names must appear in the result type expression. If an attribute is itself a
// result type then the view may specify which view to use when rendering the
// attribute using the View function in the View DSL. If not specified then the
// view named "default" is used.
//
// View must appear in a ResultType expression.
//
// View accepts two arguments: the view name and its defining DSL.
//
// Examples:
//
//	View("default", func() {
//              // "id" and "name" must be result type attributes
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
