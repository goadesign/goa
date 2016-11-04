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
	Attribute("user", UserType)
})

// UserType is a type used to define an attribute in AllTypes.
var UserType = Type("UserType", func() {
	Attribute("string", String)
})

// Attributes is a type definition which demonstrates the different ways
// attributes may be defined.
var Attributes = Type("Attributes", func() {
	Attribute("type", String)
	Attribute("type_desc", String, "description")
	Attribute("type_validations", String, func() {
		MinLength(10)
	})
	Attribute("type_desc_validations", String, "description", func() {
		Description("with validations")
		MinLength(10)
	})
	Attribute("type_inline_desc_validations", String, "description", func() {
		MinLength(10)
	})
})

// Validations is a type definition with all possible validations.
var Validations = Type("Validations", func() {
	Description("An object with attributes with all possible validations")

	Attribute("string", String, func() {
		MinLength(5)
		MaxLength(10)
		Pattern(`^A.*@gmail\.com`)
		Format("email")
	})

	Attribute("bytes", Bytes, func() {
		MinLength(5)
		MaxLength(10)
	})

	Attribute("number", Int64, func() {
		Minimum(1)
		Maximum(10)
	})

	Attribute("array", ArrayOf(String), func() {
		MinLength(5)
		MaxLength(10)
	})

	Attribute("map", ArrayOf(String), func() {
		MinLength(5)
		MaxLength(10)
	})
})

// Embedded is a type definition with embedded object attribute definitions.
var Embedded = Type("Embedded", func() {
	Attribute("embedded", func() {
		Description("Inner object description")
		Attribute("inner1", String)
		Attribute("inner2", Int32)
		Required("inner")
	})
})

// Recursive is a type definition with recursive attribute definitions.
var Recursive = Type("Recursive", func() {
	Attribute("recursive", "Recursive")
	Attribute("recursives", ArrayOf("Recursive"))
})
