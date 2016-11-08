// This test demonstrates all the possible usage of Type.
package dsl_test

import . "github.com/goadesign/goa/design"
import . "github.com/goadesign/goa/design/dsl"

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
})

// AUserType is a type used to define an attribute in AllTypes.
var AUserType = Type("UserType", func() {
	Description("Optional description")
	Attribute("required", String)
	Attribute("optional", String)
	Required("required")
})

// AMediaType is a media type used to define an attribute in AllTypes.
var AMediaType = Type("MediaType", func() {
	Description("Optional description")
	Attributes(func() {
		Attribute("optional", String)
		Attribute("required", String)
		Required("required")
	})
	View(Default, func() {
		Attribute("optional", String)
		Attribute("required", String)
	})
})

// Attributes is a type definition which demonstrates the different ways
// attributes may be defined.
var Attributes = Type("Attributes", func() {
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
		DefaultValue("default value")
		Example("example value")
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
