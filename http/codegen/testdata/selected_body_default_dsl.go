// This file defines selected HTTP body defaults used by constructor goldens.
package testdata

import . "goa.design/goa/v3/dsl"

// PayloadSelectedBodyDefaultsDSL defines composite, byte, and OneOf defaults
// whose request body may be omitted.
var PayloadSelectedBodyDefaultsDSL = func() {
	empty := Type("SelectedDefaultEmpty", func() {})
	details := Type("SelectedDefaultDetails", func() {
		Attribute("name", String)
		Required("name")
	})
	labels := Type("SelectedDefaultLabels", MapOf(String, String))
	Service("SelectedBodyDefaults", func() {
		Method("Object", func() {
			Payload(func() {
				Attribute("value", details, func() {
					Default(map[string]any{"name": "safe"})
				})
			})
			HTTP(func() {
				POST("/object")
				Body("value")
			})
		})
		Method("Map", func() {
			Payload(func() {
				Attribute("value", labels, func() {
					Default(map[string]string{"mode": "safe"})
				})
			})
			HTTP(func() {
				POST("/map")
				Body("value")
			})
		})
		Method("Bytes", func() {
			Payload(func() {
				Attribute("value", Bytes, func() {
					Default("safe")
				})
			})
			HTTP(func() {
				POST("/bytes")
				Body("value")
			})
		})
		Method("Union", func() {
			Payload(func() {
				OneOf("value", func() {
					TypeName("SelectedDefaultValue")
					Field(1, "name", String)
					Field(2, "inactive", empty)
					Default(map[string]any{"type": "inactive", "value": map[string]any{}})
				})
			})
			HTTP(func() {
				POST("/union")
				Body("value")
			})
		})
	})
}

// PayloadSelectedBodyObjectDefaultDSL defines an optional selected object body
// whose omitted value initializes a required child field.
var PayloadSelectedBodyObjectDefaultDSL = func() {
	details := Type("SelectedObjectDefault", func() {
		Attribute("name", String)
		Required("name")
	})
	Service("SelectedObjectBodyDefault", func() {
		Method("Object", func() {
			Payload(func() {
				Attribute("value", details, func() {
					Default(map[string]any{"name": "safe"})
				})
			})
			HTTP(func() {
				POST("/object")
				Body("value")
			})
		})
	})
}

// PayloadSelectedBodyMapDefaultDSL defines an optional selected map body whose
// omitted value initializes the authored entries.
var PayloadSelectedBodyMapDefaultDSL = func() {
	labels := Type("SelectedMapDefault", MapOf(String, String))
	Service("SelectedMapBodyDefault", func() {
		Method("Map", func() {
			Payload(func() {
				Attribute("value", labels, func() {
					Default(map[string]string{"mode": "safe"})
				})
			})
			HTTP(func() {
				POST("/map")
				Body("value")
			})
		})
	})
}

// PayloadSelectedBodyBytesDefaultDSL defines an optional selected byte body
// whose omitted value initializes the authored text as bytes.
var PayloadSelectedBodyBytesDefaultDSL = func() {
	Service("SelectedBytesBodyDefault", func() {
		Method("Bytes", func() {
			Payload(func() {
				Attribute("value", Bytes, func() {
					Default("safe")
				})
			})
			HTTP(func() {
				POST("/bytes")
				Body("value")
			})
		})
	})
}
