package testdata

import . "goa.design/goa/dsl"

var TestTypesDSL = func() {
	var (
		Simple = Type("Simple", func() {
			Attribute("required_string", String)
			Attribute("default_bool", Boolean, func() {
				Default(true)
			})
			Attribute("integer", Int)
			Required("required_string")
		})

		_ = Type("Required", func() {
			Extend(Simple)
			Required("required_string", "default_bool", "integer")
		})

		_ = Type("Super", func() {
			Extend(Simple)
			Attribute("ignored_attr", Float32)
		})

		_ = Type("Default", func() {
			Attribute("required_string", String, func() {
				Default("foo")
			})
			Attribute("default_bool", Boolean, func() {
				Default(true)
			})
			Attribute("integer", Int, func() {
				Default(1)
			})
			Required("integer")
		})

		SimpleMap = Type("SimpleMap", func() {
			Attribute("simple", MapOf(String, Int))
		})

		_ = Type("RequiredMap", func() {
			Extend(SimpleMap)
			Required("simple")
		})

		_ = Type("DefaultMap", func() {
			Attribute("simple", MapOf(String, Int), func() {
				Default(map[string]int{"foo": 1})
			})
		})

		_ = Type("NestedMap", func() {
			Attribute("nested_map", MapOf(Float64, MapOf(Int, MapOf(Float64, UInt64))))
		})

		_ = Type("TypeMap", func() {
			Attribute("type_map", MapOf(String, SimpleMap))
		})

		_ = Type("ArrayMap", func() {
			Attribute("array_map", MapOf(UInt32, ArrayOf(Float32)))
		})

		SimpleArray = Type("SimpleArray", func() {
			Attribute("string_array", ArrayOf(String))
		})

		_ = Type("RequiredArray", func() {
			Extend(SimpleArray)
			Required("string_array")
		})

		_ = Type("DefaultArray", func() {
			Attribute("string_array", ArrayOf(String), func() {
				Default([]string{"foo", "bar"})
			})
		})

		_ = Type("NestedArray", func() {
			Attribute("nested_array", ArrayOf(ArrayOf(ArrayOf(Float64))))
		})

		_ = Type("TypeArray", func() {
			Attribute("type_array", ArrayOf(SimpleArray))
		})

		_ = Type("MapArray", func() {
			Attribute("map_array", ArrayOf(MapOf(Int, String)))
		})

		_ = Type("Composite", func() {
			Attribute("required_string", String)
			Attribute("default_int", Int)
			Attribute("type", Simple)
			Attribute("map", MapOf(Int, String))
			Attribute("array", ArrayOf(String))
		})

		_ = Type("CompositeWithCustomField", func() {
			Attribute("required_string", String, func() {
				Meta("struct:field:name", "my_string")
			})
			Attribute("default_int", Int, func() {
				Meta("struct:field:name", "my_int")
				Default(100)
			})
			Attribute("type", Simple, func() {
				Meta("struct:field:name", "my_type")
			})
			Attribute("map", MapOf(Int, String), func() {
				Meta("struct:field:name", "my_map")
			})
			Attribute("array", ArrayOf(String), func() {
				Meta("struct:field:name", "my_array")
			})
			Required("required_string", "type", "map", "array")
		})

		_ = Type("Recursive", func() {
			Attribute("required_string", String)
			Attribute("recursive", "Recursive")
			Required("required_string")
		})

		_ = ResultType("application/vnd.goa.example", func() {
			TypeName("ResultType")
			Attributes(func() {
				Attribute("int", Int)
				Attribute("map", MapOf(Int, String))
			})
		})
	)
}
