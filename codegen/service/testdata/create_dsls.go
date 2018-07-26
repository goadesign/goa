package testdata

import (
	. "goa.design/goa/design"
	. "goa.design/goa/dsl"
)

var CreateStringDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(StringT{})
		Attribute("String", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var CreateStringRequiredDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(StringT{})
		Attribute("String", String)
		Required("String")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var CreateStringPointerDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(StringPointerT{})
		Attribute("String", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var CreateStringPointerRequiredDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(StringPointerT{})
		Attribute("String", String)
		Required("String")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var CreateArrayStringDSL = func() {
	var ArrayStringType = Type("ArrayStringType", func() {
		CreateFrom(ArrayStringT{})
		Attribute("ArrayString", ArrayOf(String))
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ArrayStringType)
		})
	})
}

var CreateArrayStringRequiredDSL = func() {
	var ArrayStringType = Type("ArrayStringType", func() {
		CreateFrom(ArrayStringT{})
		Attribute("ArrayString", ArrayOf(String))
		Required("ArrayString")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ArrayStringType)
		})
	})
}

var CreateObjectDSL = func() {
	var ObjectField = Type("ObjectField", func() {
		Attribute("Bool", Boolean)
		Attribute("Int", Int)
		Attribute("Int32", Int32)
		Attribute("Int64", Int64)
		Attribute("UInt", UInt)
		Attribute("UInt32", UInt32)
		Attribute("UInt64", UInt64)
		Attribute("Float32", Float32)
		Attribute("Float64", Float64)
		Attribute("Bytes", Bytes)
		Attribute("String", String)
		Attribute("Array", ArrayOf(Boolean))
		Attribute("Map", MapOf(String, Boolean))
	})

	var ObjectType = Type("ObjectType", func() {
		CreateFrom(ObjectT{})
		Attribute("Object", ObjectField)
		Required("Object")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(ObjectType)
		})
	})
}

var CreateObjectRequiredDSL = func() {
	var ObjectField = Type("ObjectField", func() {
		Attribute("Bool", Boolean)
		Attribute("Int", Int)
		Attribute("Int32", Int32)
		Attribute("Int64", Int64)
		Attribute("UInt", UInt)
		Attribute("UInt32", UInt32)
		Attribute("UInt64", UInt64)
		Attribute("Float32", Float32)
		Attribute("Float64", Float64)
		Attribute("Bytes", Bytes)
		Attribute("String", String)
		Attribute("Array", ArrayOf(Boolean))
		Attribute("Map", MapOf(String, Boolean))
		Required("Bool", "Int", "Int32", "Int64", "UInt", "UInt32",
			"UInt64", "Float32", "Float64", "Bytes", "String", "Array", "Map")
	})

	var ObjectType = Type("ObjectType", func() {
		CreateFrom(ObjectT{})
		Attribute("Object", ObjectField)
		Required("Object")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(ObjectType)
		})
	})
}

var CreateObjectExtraDSL = func() {
	var ObjectField = Type("ObjectField", func() {
		Attribute("Bool", Boolean)
		Attribute("Int", Int)
		Attribute("Int32", Int32)
		Attribute("Int64", Int64)
		Attribute("UInt", UInt)
		Attribute("UInt32", UInt32)
		Attribute("UInt64", UInt64)
		Attribute("Float32", Float32)
		Attribute("Float64", Float64)
		Attribute("Bytes", Bytes)
		Attribute("String", String)
		Attribute("Array", ArrayOf(Boolean))
		Attribute("Map", MapOf(String, Boolean))
	})

	var ObjectType = Type("ObjectType", func() {
		CreateFrom(ObjectExtraT{})
		Attribute("Object", ObjectField)
		Required("Object")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(ObjectType)
		})
	})
}
