package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var TestUnionDSL = func() {
	var (
		SomeType = Type("SomeType", func() {
			Attribute("someField", String)
		})

		UnionString = Type("UnionString", func() {
			OneOf("UnionString", func() {
				Attribute("String", String)
			})
		})
		UnionString2 = Type("UnionString2", func() {
			OneOf("UnionString2", func() {
				Attribute("String", String)
			})
		})
		UnionStringInt = Type("UnionStringInt", func() {
			OneOf("UnionStringInt", func() {
				Attribute("String", String)
				Attribute("Int", Int)
			})
		})
		UnionStringInt2 = Type("UnionStringInt2", func() {
			OneOf("UnionStringInt2", func() {
				Attribute("String", String)
				Attribute("Int", Int)
			})
		})
		UnionSomeType = Type("UnionSomeType", func() {
			OneOf("UnionSomeType", func() {
				Attribute("SomeType", SomeType)
			})
		})

		_ = Type("Container", func() {
			Attribute("UnionString", UnionString)
			Attribute("UnionString2", UnionString2)
			Attribute("UnionStringInt", UnionStringInt)
			Attribute("UnionStringInt2", UnionStringInt2)
			Attribute("UnionSomeType", UnionSomeType)
		})

		_ = Type("UnionUserType", func() {
			Attribute("Type", String)
			Attribute("Value", String)
			Required("Type", "Value")
		})
	)
}
