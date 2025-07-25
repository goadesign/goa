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

var ConvertExternalNameWithInitialismDSL = func() {
	var ExternalNameWithInitialismType = Type("ExternalNameWithInitialismType", func() {
		ConvertTo(ApiNameT{})
		Attribute("string", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(ExternalNameWithInitialismType)
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

// Test cases for multiple package struct:pkg:path with ConvertTo/CreateFrom

// Mock external types for testing
type TestFilterConfig struct {
	Name    string
	Enabled bool
	Value   int
}

type TestUserConfig struct {
	Username string
	Active   bool
}

// ConvertMultiPkgDSL tests conversion functions with different struct:pkg:path values
var ConvertMultiPkgDSL = func() {
	var FilterConfigType = Type("FilterConfig", func() {
		Meta("struct:pkg:path", "types")
		CreateFrom(TestFilterConfig{})
		ConvertTo(TestFilterConfig{})
		Attribute("name", String)
		Attribute("enabled", Boolean)
		Attribute("value", Int)
		Required("name", "enabled", "value")
	})

	var UserConfigType = Type("UserConfig", func() {
		Meta("struct:pkg:path", "models")
		CreateFrom(TestUserConfig{})
		ConvertTo(TestUserConfig{})
		Attribute("username", String)
		Attribute("active", Boolean)
		Required("username", "active")
	})

	Service("MultiPkgService", func() {
		Method("FilterMethod", func() {
			Payload(FilterConfigType)
			Result(FilterConfigType)
		})
		Method("UserMethod", func() {
			Payload(UserConfigType)
			Result(UserConfigType)
		})
	})
}

// ConvertSinglePkgDSL tests conversion functions with single custom package
var ConvertSinglePkgDSL = func() {
	var FilterConfigType = Type("FilterConfig", func() {
		Meta("struct:pkg:path", "types")
		CreateFrom(TestFilterConfig{})
		ConvertTo(TestFilterConfig{})
		Attribute("name", String)
		Attribute("enabled", Boolean)
		Attribute("value", Int)
		Required("name", "enabled", "value")
	})

	Service("SinglePkgService", func() {
		Method("FilterMethod", func() {
			Payload(FilterConfigType)
			Result(FilterConfigType)
		})
	})
}

// ConvertMixedPkgDSL tests mix of custom paths and default service package
var ConvertMixedPkgDSL = func() {
	var FilterConfigType = Type("FilterConfig", func() {
		Meta("struct:pkg:path", "types")
		CreateFrom(TestFilterConfig{})
		ConvertTo(TestFilterConfig{})
		Attribute("name", String)
		Required("name")
	})

	var SystemConfigType = Type("SystemConfig", func() {
		// No Meta - should go to service package
		CreateFrom(TestUserConfig{})
		ConvertTo(TestUserConfig{})
		Attribute("username", String)
		Required("username")
	})

	Service("MixedPkgService", func() {
		Method("FilterMethod", func() {
			Payload(FilterConfigType)
		})
		Method("SystemMethod", func() {
			Payload(SystemConfigType)
		})
	})
}
