package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var MultiNoPayloadDSL = func() {
	Service("ServiceMultiNoPayload1", func() {
		Method("MethodServiceNoPayload11", func() {
			HTTP(func() {
				GET("/11")
			})
		})
		Method("MethodServiceNoPayload12", func() {
			HTTP(func() {
				GET("/12")
			})
		})
	})
	Service("ServiceMultiNoPayload2", func() {
		Method("MethodServiceNoPayload21", func() {
			HTTP(func() {
				GET("/21")
			})
		})
		Method("MethodServiceNoPayload22", func() {
			HTTP(func() {
				GET("/22")
			})
		})
	})
}

var MultiSimpleDSL = func() {
	Service("ServiceMultiSimple1", func() {
		Method("MethodMultiSimpleNoPayload", func() {
			HTTP(func() {
				GET("/")
			})
		})
		Method("MethodMultiSimplePayload", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
	Service("ServiceMultiSimple2", func() {
		Method("MethodMultiSimpleNoPayload", func() {
			HTTP(func() {
				GET("/2")
			})
		})
		Method("MethodMultiSimplePayload", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			HTTP(func() {
				POST("/2")
			})
		})
	})
}

var MultiRequiredPayloadDSL = func() {
	Service("ServiceMultiRequired1", func() {
		Method("MethodMultiRequiredPayload", func() {
			Payload(func() {
				Attribute("a", Boolean)
				Required("a")
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
	Service("ServiceMultiRequired2", func() {
		Method("MethodMultiRequiredNoPayload", func() {
			HTTP(func() {
				GET("/2")
			})
		})
		Method("MethodMultiRequiredPayload", func() {
			Payload(func() {
				Attribute("a", Boolean)
				Required("a")
			})
			HTTP(func() {
				POST("/2")
				Param("a")
			})
		})
	})
}

var MultiDSL = func() {
	var UserType = Type("UserType", func() {
		Attribute("att", Boolean)
		Attribute("att2", Int)
		Attribute("att3", Int32)
		Attribute("att4", Int64)
		Attribute("att5", UInt)
		Attribute("att6", UInt32)
		Attribute("att7", UInt64)
		Attribute("att8", Float32)
		Attribute("att9", Float64)
		Attribute("att10", String)
		Attribute("att11", Bytes)
		Attribute("att12", Any)
		Attribute("att13", ArrayOf(String))
		Attribute("att14", MapOf(String, String))
		Attribute("att15", func() {
			Attribute("inline")
		})
		Attribute("att16", "UserType")
	})

	Service("ServiceMulti", func() {
		Method("MethodMultiNoPayload", func() {
			HTTP(func() {
				GET("/")
			})
		})
		Method("MethodMultiPayload", func() {
			Payload(func() {
				Attribute("a", Boolean)
				Attribute("b", String)
				Attribute("c", UserType)
			})
			HTTP(func() {
				Header("a")
				Param("b")
				POST("/")
			})
		})
	})
}
