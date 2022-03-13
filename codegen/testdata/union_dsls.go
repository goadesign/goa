package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var TestUnionDSL = func() {
	var (
		SomeType = Type("SomeType", func() {
			Attribute("someField", String)
		})

		UnionString        = OneOf("UnionString", String)
		UnionString2       = OneOf("UnionString2", String)
		UnionStringInt     = OneOf("UnionStringInt", String, Int)
		UnionSomeType      = OneOf("UnionSomeType", SomeType)
		UnionArray         = OneOf("UnionArray", ArrayOf(String))
		UnionArrayUserType = OneOf("UnionArrayUserType", ArrayOf(SomeType))
		UnionMap           = OneOf("UnionMap", MapOf(String, String))
		UnionMapUserType   = OneOf("UnionMapUserType", MapOf(String, SomeType))

		_ = Type("Container", func() {
			Attribute("UnionString", UnionString)
			Attribute("UnionString2", UnionString2)
			Attribute("UnionStringInt", UnionStringInt)
			Attribute("UnionSomeType", UnionSomeType)
			Attribute("UnionArray", UnionArray)
			Attribute("UnionArrayUserType", UnionArrayUserType)
			Attribute("UnionMap", UnionMap)
			Attribute("UnionMapUserType", UnionMapUserType)
		})

		_ = Type("UnionUserType", func() {
			Attribute("Type", String)
			Attribute("Value", Any)
			Required("Type", "Value")
		})
	)
}
