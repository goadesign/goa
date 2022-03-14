package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var TestUnionDSL = func() {
	var (
		SomeType = Type("SomeType", func() {
			Attribute("someField", String)
		})

		UnionString    = OneOf("UnionString", func() { Attribute("String", String) })
		UnionString2   = OneOf("UnionString2", func() { Attribute("String", String) })
		UnionStringInt = OneOf("UnionStringInt", func() { Attribute("String", String); Attribute("Int", Int) })
		UnionSomeType  = OneOf("UnionSomeType", func() { Attribute("SomeType", SomeType) })

		_ = Type("Container", func() {
			Attribute("UnionString", UnionString)
			Attribute("UnionString2", UnionString2)
			Attribute("UnionStringInt", UnionStringInt)
			Attribute("UnionSomeType", UnionSomeType)
		})

		_ = Type("UnionUserType", func() {
			Attribute("Type", String)
			Attribute("Value", Any)
			Required("Type", "Value")
		})
	)
}
