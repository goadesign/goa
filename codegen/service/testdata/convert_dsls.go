package testdata

import (
	aliasd "goa.design/goa/v3/codegen/service/testdata/alias-external"
	"goa.design/goa/v3/codegen/service/testdata/external"
	. "goa.design/goa/v3/dsl"
)

var ConvertStringDSL = func() {
	var StringType = Type("StringType", func() {
		ConvertTo(StringT{})
		Attribute("String", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var ConvertStringRequiredDSL = func() {
	var StringType = Type("StringType", func() {
		ConvertTo(StringT{})
		Attribute("String", String)
		Required("String")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var ConvertStringPointerDSL = func() {
	var StringPointerType = Type("StringPointerType", func() {
		ConvertTo(StringPointerT{})
		Attribute("String", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringPointerType)
		})
	})
}

var ConvertStringPointerRequiredDSL = func() {
	var StringPointerType = Type("StringPointerType", func() {
		ConvertTo(StringPointerT{})
		Attribute("String", String)
		Required("String")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(StringPointerType)
		})
	})
}

var ConvertExternalNameDSL = func() {
	var ExternalNameType = Type("ExternalNameType", func() {
		ConvertTo(ExternalNameT{})
		Attribute("string", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ExternalNameType)
		})
	})
}

var ConvertExternalNameRequiredDSL = func() {
	var ExternalNameType = Type("ExternalNameType", func() {
		ConvertTo(ExternalNameT{})
		Attribute("string", String)
		Required("string")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ExternalNameType)
		})
	})
}

var ConvertExternalNamePointerDSL = func() {
	var ExternalNamePointerType = Type("ExternalNamePointerType", func() {
		ConvertTo(ExternalNamePointerT{})
		Attribute("string", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ExternalNamePointerType)
		})
	})
}

var ConvertExternalNamePointerRequiredDSL = func() {
	var ExternalNamePointerType = Type("ExternalNamePointerType", func() {
		ConvertTo(ExternalNamePointerT{})
		Attribute("string", String)
		Required("string")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ExternalNamePointerType)
		})
	})
}

var ConvertArrayStringDSL = func() {
	var ArrayStringType = Type("ArrayStringType", func() {
		ConvertTo(ArrayStringT{})
		Attribute("ArrayString", ArrayOf(String))
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ArrayStringType)
		})
	})
}

var ConvertArrayStringRequiredDSL = func() {
	var ArrayStringType = Type("ArrayStringType", func() {
		ConvertTo(ArrayStringT{})
		Attribute("ArrayString", ArrayOf(String))
		Required("ArrayString")
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ArrayStringType)
		})
	})
}

var ConvertObjectDSL = func() {
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
		ConvertTo(ObjectT{})
		Attribute("Object", ObjectField)
		Required("Object")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(ObjectType)
		})
	})
}

var ConvertObjectRequiredDSL = func() {
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
		ConvertTo(ObjectT{})
		Attribute("Object", ObjectField)
		Required("Object")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(ObjectType)
		})
	})
}

var ConvertExternalDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(external.ConvertModel{})
		Attribute("Foo", String)
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}

var ConvertAliasDSL = func() {
	var StringType = Type("StringType", func() {
		CreateFrom(aliasd.ConvertModel{})
		Attribute("Bar", String)
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(StringType)
		})
	})
}
