package testing

import (
	. "goa.design/goa.v2/design"
	. "goa.design/goa.v2/dsl/rest"
)

// The DSL function names follow the following pattern:
//
// (Payload|Result)(Query|Path|Body)+(Type)(Validate)?DSL
//
// Where Type is the type of the payload or result.

var PayloadQueryBoolDSL = func() {
	Service("ServiceQueryBool", func() {
		Endpoint("EndpointQueryBool", func() {
			Payload(Boolean)
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryBoolAttributeDSL = func() {
	Service("ServiceQueryBoolAttribute", func() {
		Endpoint("EndpointQueryBoolAttribute", func() {
			Payload(func() {
				Attribute("q", Boolean)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryBoolValidateDSL = func() {
	Service("ServiceQueryBoolValidate", func() {
		Endpoint("EndpointQueryBoolValidate", func() {
			Payload(func() {
				Attribute("q", Boolean, func() {
					Enum(true)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryIntDSL = func() {
	Service("ServiceQueryInt", func() {
		Endpoint("EndpointQueryInt", func() {
			Payload(func() {
				Attribute("q", Int)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryIntValidateDSL = func() {
	Service("ServiceQueryIntValidate", func() {
		Endpoint("EndpointQueryIntValidate", func() {
			Payload(func() {
				Attribute("q", Int, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryInt32DSL = func() {
	Service("ServiceQueryInt32", func() {
		Endpoint("EndpointQueryInt32", func() {
			Payload(func() {
				Attribute("q", Int32)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryInt32ValidateDSL = func() {
	Service("ServiceQueryInt32Validate", func() {
		Endpoint("EndpointQueryInt32Validate", func() {
			Payload(func() {
				Attribute("q", Int32, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryInt64DSL = func() {
	Service("ServiceQueryInt64", func() {
		Endpoint("EndpointQueryInt64", func() {
			Payload(func() {
				Attribute("q", Int64)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryInt64ValidateDSL = func() {
	Service("ServiceQueryInt64Validate", func() {
		Endpoint("EndpointQueryInt64Validate", func() {
			Payload(func() {
				Attribute("q", Int64, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUIntDSL = func() {
	Service("ServiceQueryUInt", func() {
		Endpoint("EndpointQueryUInt", func() {
			Payload(func() {
				Attribute("q", UInt)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUIntValidateDSL = func() {
	Service("ServiceQueryUIntValidate", func() {
		Endpoint("EndpointQueryUIntValidate", func() {
			Payload(func() {
				Attribute("q", UInt, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUInt32DSL = func() {
	Service("ServiceQueryUInt32", func() {
		Endpoint("EndpointQueryUInt32", func() {
			Payload(func() {
				Attribute("q", UInt32)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUInt32ValidateDSL = func() {
	Service("ServiceQueryUInt32Validate", func() {
		Endpoint("EndpointQueryUInt32Validate", func() {
			Payload(func() {
				Attribute("q", UInt32, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUInt64DSL = func() {
	Service("ServiceQueryUInt64", func() {
		Endpoint("EndpointQueryUInt64", func() {
			Payload(func() {
				Attribute("q", UInt64)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryUInt64ValidateDSL = func() {
	Service("ServiceQueryUInt64Validate", func() {
		Endpoint("EndpointQueryUInt64Validate", func() {
			Payload(func() {
				Attribute("q", UInt64, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryFloat32DSL = func() {
	Service("ServiceQueryFloat32", func() {
		Endpoint("EndpointQueryFloat32", func() {
			Payload(func() {
				Attribute("q", Float32)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryFloat32ValidateDSL = func() {
	Service("ServiceQueryFloat32Validate", func() {
		Endpoint("EndpointQueryFloat32Validate", func() {
			Payload(func() {
				Attribute("q", Float32, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryFloat64DSL = func() {
	Service("ServiceQueryFloat64", func() {
		Endpoint("EndpointQueryFloat64", func() {
			Payload(func() {
				Attribute("q", Float64)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryFloat64ValidateDSL = func() {
	Service("ServiceQueryFloat64Validate", func() {
		Endpoint("EndpointQueryFloat64Validate", func() {
			Payload(func() {
				Attribute("q", Float64, func() {
					Minimum(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryStringDSL = func() {
	Service("ServiceQueryString", func() {
		Endpoint("EndpointQueryString", func() {
			Payload(func() {
				Attribute("q", String)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryStringValidateDSL = func() {
	Service("ServiceQueryStringValidate", func() {
		Endpoint("EndpointQueryStringValidate", func() {
			Payload(func() {
				Attribute("q", String, func() {
					Enum("val")
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryBytesDSL = func() {
	Service("ServiceQueryBytes", func() {
		Endpoint("EndpointQueryBytes", func() {
			Payload(func() {
				Attribute("q", Bytes)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryBytesValidateDSL = func() {
	Service("ServiceQueryBytesValidate", func() {
		Endpoint("EndpointQueryBytesValidate", func() {
			Payload(func() {
				Attribute("q", Bytes, func() {
					MinLength(1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryAnyDSL = func() {
	Service("ServiceQueryAny", func() {
		Endpoint("EndpointQueryAny", func() {
			Payload(func() {
				Attribute("q", Any)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryAnyValidateDSL = func() {
	Service("ServiceQueryAnyValidate", func() {
		Endpoint("EndpointQueryAnyValidate", func() {
			Payload(func() {
				Attribute("q", Any, func() {
					Enum("val", 1)
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayBoolDSL = func() {
	Service("ServiceQueryArrayBool", func() {
		Endpoint("EndpointQueryArrayBool", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Boolean))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayBoolValidateDSL = func() {
	Service("ServiceQueryArrayBoolValidate", func() {
		Endpoint("EndpointQueryArrayBoolValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Boolean), func() {
					MinLength(1)
					Elem(func() {
						Enum(true)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayIntDSL = func() {
	Service("ServiceQueryArrayInt", func() {
		Endpoint("EndpointQueryArrayInt", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayIntValidateDSL = func() {
	Service("ServiceQueryArrayIntValidate", func() {
		Endpoint("EndpointQueryArrayIntValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayInt32DSL = func() {
	Service("ServiceQueryArrayInt32", func() {
		Endpoint("EndpointQueryArrayInt32", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int32))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayInt32ValidateDSL = func() {
	Service("ServiceQueryArrayInt32Validate", func() {
		Endpoint("EndpointQueryArrayInt32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayInt64DSL = func() {
	Service("ServiceQueryArrayInt64", func() {
		Endpoint("EndpointQueryArrayInt64", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int64))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayInt64ValidateDSL = func() {
	Service("ServiceQueryArrayInt64Validate", func() {
		Endpoint("EndpointQueryArrayInt64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUIntDSL = func() {
	Service("ServiceQueryArrayUInt", func() {
		Endpoint("EndpointQueryArrayUInt", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUIntValidateDSL = func() {
	Service("ServiceQueryArrayUIntValidate", func() {
		Endpoint("EndpointQueryArrayUIntValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUInt32DSL = func() {
	Service("ServiceQueryArrayUInt32", func() {
		Endpoint("EndpointQueryArrayUInt32", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt32))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUInt32ValidateDSL = func() {
	Service("ServiceQueryArrayUInt32Validate", func() {
		Endpoint("EndpointQueryArrayUInt32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUInt64DSL = func() {
	Service("ServiceQueryArrayUInt64", func() {
		Endpoint("EndpointQueryArrayUInt64", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt64))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayUInt64ValidateDSL = func() {
	Service("ServiceQueryArrayUInt64Validate", func() {
		Endpoint("EndpointQueryArrayUInt64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayFloat32DSL = func() {
	Service("ServiceQueryArrayFloat32", func() {
		Endpoint("EndpointQueryArrayFloat32", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float32))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayFloat32ValidateDSL = func() {
	Service("ServiceQueryArrayFloat32Validate", func() {
		Endpoint("EndpointQueryArrayFloat32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayFloat64DSL = func() {
	Service("ServiceQueryArrayFloat64", func() {
		Endpoint("EndpointQueryArrayFloat64", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float64))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayFloat64ValidateDSL = func() {
	Service("ServiceQueryArrayFloat64Validate", func() {
		Endpoint("EndpointQueryArrayFloat64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayStringDSL = func() {
	Service("ServiceQueryArrayString", func() {
		Endpoint("EndpointQueryArrayString", func() {
			Payload(func() {
				Attribute("q", ArrayOf(String))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayStringValidateDSL = func() {
	Service("ServiceQueryArrayStringValidate", func() {
		Endpoint("EndpointQueryArrayStringValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(String), func() {
					MinLength(1)
					Elem(func() {
						Enum("val")
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayBytesDSL = func() {
	Service("ServiceQueryArrayBytes", func() {
		Endpoint("EndpointQueryArrayBytes", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Bytes))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayBytesValidateDSL = func() {
	Service("ServiceQueryArrayBytesValidate", func() {
		Endpoint("EndpointQueryArrayBytesValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Bytes), func() {
					MinLength(1)
					Elem(func() {
						MinLength(2)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayAnyDSL = func() {
	Service("ServiceQueryArrayAny", func() {
		Endpoint("EndpointQueryArrayAny", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Any))
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryArrayAnyValidateDSL = func() {
	Service("ServiceQueryArrayAnyValidate", func() {
		Endpoint("EndpointQueryArrayAnyValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Any), func() {
					MinLength(1)
					Elem(func() {
						Enum("val", 1)
					})
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadPathStringDSL = func() {
	Service("ServicePathString", func() {
		Endpoint("EndpointPathString", func() {
			Payload(func() {
				Attribute("p", String)
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathStringValidateDSL = func() {
	Service("ServicePathStringValidate", func() {
		Endpoint("EndpointPathStringValidate", func() {
			Payload(func() {
				Attribute("p", String, func() {
					Enum("val")
				})
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathArrayStringDSL = func() {
	Service("ServicePathArrayString", func() {
		Endpoint("EndpointPathArrayString", func() {
			Payload(func() {
				Attribute("p", ArrayOf(String))
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathArrayStringValidateDSL = func() {
	Service("ServicePathArrayStringValidate", func() {
		Endpoint("EndpointPathArrayStringValidate", func() {
			Payload(func() {
				Attribute("p", ArrayOf(String), func() {
					Enum([]string{"val"})
				})
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadBodyStringDSL = func() {
	Service("ServiceBodyString", func() {
		Endpoint("EndpointBodyString", func() {
			Payload(func() {
				Attribute("b", String)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyStringValidateDSL = func() {
	Service("ServiceBodyStringValidate", func() {
		Endpoint("EndpointBodyStringValidate", func() {
			Payload(func() {
				Attribute("b", String, func() {
					Pattern("pattern")
				})
				Required("b")
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String)
	})
	Service("ServiceBodyUser", func() {
		Endpoint("EndpointBodyUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("apattern")
		})
	})
	Service("ServiceBodyUserValidate", func() {
		Endpoint("EndpointBodyUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyArrayStringDSL = func() {
	Service("ServiceBodyArrayString", func() {
		Endpoint("EndpointBodyArrayString", func() {
			Payload(func() {
				Attribute("b", ArrayOf(String))
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyArrayStringValidateDSL = func() {
	Service("ServiceBodyArrayStringValidate", func() {
		Endpoint("EndpointBodyArrayStringValidate", func() {
			Payload(func() {
				Attribute("b", ArrayOf(String), func() {
					MinLength(2)
					Elem(func() {
						MinLength(3)
					})
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyArrayUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("apattern")
		})
	})
	Service("ServiceBodyArrayUser", func() {
		Endpoint("EndpointBodyArrayUser", func() {
			Payload(func() {
				Attribute("b", ArrayOf(PayloadType))
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyArrayUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("apattern")
		})
	})
	Service("ServiceBodyArrayUserValidate", func() {
		Endpoint("EndpointBodyArrayUserValidate", func() {
			Payload(func() {
				Attribute("b", ArrayOf(PayloadType), func() {
					MinLength(2)
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyMapStringDSL = func() {
	Service("ServiceBodyMapString", func() {
		Endpoint("EndpointBodyMapString", func() {
			Payload(func() {
				Attribute("b", MapOf(String, String))
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyMapStringValidateDSL = func() {
	Service("ServiceBodyMapStringValidate", func() {
		Endpoint("EndpointBodyMapStringValidate", func() {
			Payload(func() {
				Attribute("b", MapOf(String, String), func() {
					Elem(func() {
						MinLength(2)
					})
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyMapUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("apattern")
		})
	})
	Service("ServiceBodyMapUser", func() {
		Endpoint("EndpointBodyMapUser", func() {
			Payload(func() {
				Attribute("b", MapOf(String, PayloadType))
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyMapUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("apattern")
		})
	})
	Service("ServiceBodyMapUserValidate", func() {
		Endpoint("EndpointBodyMapUserValidate", func() {
			Payload(func() {
				Attribute("b", MapOf(String, PayloadType), func() {
					Key(func() {
						MinLength(2)
					})
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyQueryObjectDSL = func() {
	Service("ServiceBodyQueryObject", func() {
		Endpoint("EndpointBodyQueryObject", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
			})
			HTTP(func() {
				POST("/")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryObjectValidateDSL = func() {
	Service("ServiceBodyQueryObjectValidate", func() {
		Endpoint("EndpointBodyQueryObjectValidate", func() {
			Payload(func() {
				Attribute("a", String, func() {
					Pattern("patterna")
				})
				Attribute("b", String, func() {
					Pattern("patternb")
				})
				Required("a", "b")
			})
			HTTP(func() {
				POST("/")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String)
		Attribute("b", String)
	})
	Service("ServiceBodyQueryUser", func() {
		Endpoint("EndpointBodyQueryUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a", "b")
	})
	Service("ServiceBodyQueryUserValidate", func() {
		Endpoint("EndpointBodyQueryUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
				Param("b")
			})
		})
	})
}

var PayloadBodyPathObjectDSL = func() {
	Service("ServiceBodyPathObject", func() {
		Endpoint("EndpointBodyPathObject", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
			})
			HTTP(func() {
				POST("/:b")
			})
		})
	})
}

var PayloadBodyPathObjectValidateDSL = func() {
	Service("ServiceBodyPathObjectValidate", func() {
		Endpoint("EndpointBodyPathObjectValidate", func() {
			Payload(func() {
				Attribute("a", String, func() {
					Pattern("patterna")
				})
				Attribute("b", String, func() {
					Pattern("patternb")
				})
				Required("a", "b")
			})
			HTTP(func() {
				POST("/:b")
			})
		})
	})
}

var PayloadBodyPathUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String)
		Attribute("b", String)
	})
	Service("ServiceBodyPathUser", func() {
		Endpoint("EndpointBodyPathUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/:b")
			})
		})
	})
}

var PayloadBodyPathUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a", "b")
	})
	Service("ServiceBodyPathUserValidate", func() {
		Endpoint("EndpointUserBodyPathValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/:b")
			})
		})
	})
}

var PayloadBodyQueryPathObjectDSL = func() {
	Service("ServiceBodyQueryPathObject", func() {
		Endpoint("EndpointBodyQueryPathObject", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
				Attribute("c", String)
			})
			HTTP(func() {
				POST("/{c}")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryPathObjectValidateDSL = func() {
	Service("ServiceBodyQueryPathObjectValidate", func() {
		Endpoint("EndpointBodyQueryPathObjectValidate", func() {
			Payload(func() {
				Attribute("a", String, func() {
					Pattern("patterna")
				})
				Attribute("b", String, func() {
					Pattern("patternb")
				})
				Attribute("c", String, func() {
					Pattern("patternc")
				})
				Required("a", "b", "c")
			})
			HTTP(func() {
				POST("/{c}")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryPathUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String)
		Attribute("b", String)
		Attribute("c", String)
	})
	Service("ServiceBodyQueryPathUser", func() {
		Endpoint("EndpointBodyQueryPathUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{c}")
				Param("b")
			})
		})
	})
}

var PayloadBodyQueryPathUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Attribute("c", String, func() {
			Pattern("patternc")
		})
		Required("a", "b", "c")
	})
	Service("ServiceBodyQueryPathUserValidate", func() {
		Endpoint("EndpointBodyQueryPathUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{c}")
				Param("b")
			})
		})
	})
}
