package testdata

import . "goa.design/goa/v3/dsl"

// RecursiveValidationDSL defines types for testing recursive validation
// with proper cycle detection.
var RecursiveValidationDSL = func() {
	// Recursive type that references itself directly
	_ = Type("RecursiveType", func() {
		Attribute("name", String, func() {
			MinLength(1)
		})
		Attribute("child", "RecursiveType")
		Required("name")
	})

	// Type with nested recursive type in array
	_ = Type("ContainerWithRecursiveArray", func() {
		Attribute("items", ArrayOf("RecursiveType"))
	})

	// Type with nested recursive type in map
	_ = Type("ContainerWithRecursiveMap", func() {
		Attribute("items", MapOf(String, "RecursiveType"))
	})

	// Type with multiple levels of nesting before recursion
	_ = Type("NestedRecursive", func() {
		Attribute("name", String)
		Attribute("inner", func() {
			Attribute("name", String)
			Attribute("deep", "NestedRecursive")
			Required("name")
		})
		Required("name")
	})
}
