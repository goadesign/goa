// This test demonstrates all the possible usage of Type.
package dsl_test

import . "goa.design/goa.v2/design"
import . "goa.design/goa.v2/dsl"

// BasicType shows the basic usage for Type.
//
// Type is the DSL function used to describe user types. A user type is an
// object type that can be used to define request and response types or
// attributes of request and response types.
//
// Type takes the name of the type as first argument. This name must be unique
// across all types in a given design package.
var BasicType = Type("Name", func() {
	// Optional description used in code comments and docs.
	Description("Optional description")

	// Attribute defines a field of the type, see below for all possible
	// usage.
	Attribute("an_attribute", String)

	// Required lists the required attributes. Required takes one or more
	// attribute names and can appear one or more times.
	Required("an_attribute")
})

// BasicMediaType shows the basic usage for MediaType.
//
// MediaType is the DSL function used to describe media types. A media type is a
// special kind of type that adds the concept of views: A view defines a subset
// of the type attributes to be rendered. This is used to describe *response*
// types where an endpoint may render different attributes depending on the
// request state or when different endpoints render the type differently (for
// example a listing endpoint may render less attribute than a endpoint that
// retrieves a single value). All media types muse define a default view. The
// default view is the view named "default".
//
// Media type takes a media type identifier (as defined by RFC6 838) as first
// argument. This identifier must be unique across all media types in a given
// package.
var BasicMediaType = MediaType("application/vnd.goa.media", func() {
	// Optional description used in code comments and docs.
	Description("Optional description")

	// Attributes lists the media type attributes identically to Type.
	Attributes(func() {
		Attribute("an_attribute", String)
		Required("an_attribute")
	})

	// View defines a view. View may appear one or more times in a MediaType
	// expression.
	View("default", func() {
		// There is no need to repeat the attribute properties when
		// listing the view attributes, only the name is required.
		Attribute("an_attribute")
	})
})

// AllTypes is a type definition with attributes using all possible data types.
var AllTypes = Type("AllTypes", func() {
	Description("An object with attributes of all possible types")

	Attribute("string", String)
	Attribute("bytes", Bytes)
	Attribute("boolean", Boolean)
	Attribute("int32", Int32)
	Attribute("int64", Int64)
	Attribute("float32", Float32)
	Attribute("float64", Float64)
	Attribute("any", Any)
	Attribute("array", ArrayOf(String))
	Attribute("map", MapOf(String, String))
	Attribute("object", func() {
		Description("Inner type")
		Attribute("inner_attribute", String)
		Required("inner_attribute")
	})
	Attribute("user", AUserType)
	Attribute("media", AMediaType)
	Attribute("collection", CollectionOf(AMediaType))
})

// AUserType is a type used to define an attribute in AllTypes.
var AUserType = Type("UserType", func() {
	Description("Optional description")
	Attribute("required", String)
	Attribute("optional", String)
	Required("required")
})

// AMediaType is a media type used to define an attribute in AllTypes.
var AMediaType = MediaType("MediaType", func() {
	Description("Optional description")
	Attributes(func() {
		Attribute("optional", String)
		Attribute("required", String)
		Required("required")
	})
	View("default", func() {
		Attribute("optional", String)
		Attribute("required", String)
	})
	View("tiny", func() {
		Attribute("required", String)
	})
})

// ACollectionMedia type shows all the possible DSL of CollectionOf.
var ACollectionMedia = CollectionOf(AMediaType, func() {
	// View allows defining collection specific views.
	// The view is defined using the attributes of the element of the
	// collection.
	View("collection", func() {
		Attribute("optional")
	})

	// View can also refer to existing views defined in the element media
	// type. If no View is specified (i.e. no DSL argument is provided to
	// CollectionOf) then all the element media type views are inherited.
	View("tiny")
})

// AArrayType is a array type with a element validation.
var AArrayType = ArrayOf(String, func() {
	Pattern("regexp")
})

// AMapType is a map with element and key validations.
var AMapType = MapOf(String, String, func() {
	// Key is used to define validations that apply to the keys of the map.
	Key(func() {
		Pattern("keyregexp")
	})

	// Value is used to define validations that apply to the values of the map.
	Value(func() {
		Pattern("valueregexp")
	})
})

// Attrs is a type definition which demonstrates the different ways attributes
// may be defined.
var Attrs = Type("Attributes", func() {
	// Attribute defined with a name and a type
	Attribute("name", String)
	// Attribute defined with a name, a type and a description
	Attribute("name_2", String, "description")
	// Attribute defined with a name, a type and validations
	Attribute("name_3", String, func() {
		MinLength(10)
	})
	// Attribute defined with a name, a type, a description, validations, a
	// default value and an example.
	Attribute("name_4", String, "description", func() {
		MinLength(10)
		MaxLength(100)
		Default("default value")
		Example("example value")
		Example("another example value")
		Example("title", "example text")
	})
})

// Validations is a type definition with all possible validations.
var Validations = Type("Validations", func() {
	Description("An object with attributes with all possible validations")

	Attribute("string", String, func() {
		// MinLength specifies the minimum number of characters in the
		// string.
		MinLength(5)
		// MaxLength specifies the maximum number of characters in the
		// string.
		MaxLength(100)
		// Pattern specifies a regular expression that the value must
		// validate.
		Pattern(`^A.*@goa\.design`)
		// Format specifies a format the string must comply to.
		// See ValidationFormat constants in design package for the list
		// of supported formats.
		Format(FormatEmail)
		// Enum specifies the list of possible values (a real design
		// would probably not use other validations together with Enum)
		Enum("support@goa.design", "info@goa.design")
	})

	Attribute("bytes", Bytes, func() {
		MinLength(5)
		MaxLength(10)
		Enum([]byte{'1', '2'}, []byte{'3', '4'})
	})

	Attribute("number", Int64, func() {
		// Minimum specifies the minimum value for the number
		// (inclusive)
		Minimum(1)
		// Minimum specifies the maximum value for the number
		// (inclusive)
		Maximum(10)
		Enum(1, 2, 3, 4)
	})

	Attribute("array", ArrayOf(String), func() {
		MinLength(2)
		MaxLength(10)
		Enum([]string{"a", "b"}, []string{"c", "d"})
	})

	Attribute("map", ArrayOf(String), func() {
		MinLength(5)
		MaxLength(10)
	})
})

// Embedded is a type definition with embedded object attribute definitions.
var Embedded = Type("Embedded", func() {
	// Attribute accepts either a type or a DSL describing a type as second
	// argument.
	Attribute("embedded", func() {
		Description("Inner object description")
		Attribute("inner1", String)
		Attribute("inner2", Int32)
		Required("inner")
	})
})

// Recursive is a type definition with recursive attribute definitions.
var Recursive = Type("Recursive", func() {
	// Attribute allows specifying a type using its name rather than a Go
	// variable to make it possible to describe recursive data structures
	// without running into circular dependency compilation errors.
	Attribute("recursive", "Recursive")
	Attribute("recursives", ArrayOf("Recursive"))
})
