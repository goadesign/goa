package testdata

import (
	. "goa.design/goa/design"
	. "goa.design/goa/http/dsl"
)

// The DSL function names follow the following pattern:
//
// (Payload|Result)(Query|Path|Body)+(Type)(Validate)?DSL
//
// Where Type is the type of the payload or result.

var PayloadQueryBoolDSL = func() {
	Service("ServiceQueryBool", func() {
		Method("MethodQueryBool", func() {
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
		Method("MethodQueryBoolValidate", func() {
			Payload(func() {
				Attribute("q", Boolean, func() {
					Enum(true)
				})
				Required("q")
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
		Method("MethodQueryInt", func() {
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
		Method("MethodQueryIntValidate", func() {
			Payload(func() {
				Attribute("q", Int, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryInt32", func() {
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
		Method("MethodQueryInt32Validate", func() {
			Payload(func() {
				Attribute("q", Int32, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryInt64", func() {
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
		Method("MethodQueryInt64Validate", func() {
			Payload(func() {
				Attribute("q", Int64, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryUInt", func() {
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
		Method("MethodQueryUIntValidate", func() {
			Payload(func() {
				Attribute("q", UInt, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryUInt32", func() {
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
		Method("MethodQueryUInt32Validate", func() {
			Payload(func() {
				Attribute("q", UInt32, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryUInt64", func() {
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
		Method("MethodQueryUInt64Validate", func() {
			Payload(func() {
				Attribute("q", UInt64, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryFloat32", func() {
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
		Method("MethodQueryFloat32Validate", func() {
			Payload(func() {
				Attribute("q", Float32, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryFloat64", func() {
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
		Method("MethodQueryFloat64Validate", func() {
			Payload(func() {
				Attribute("q", Float64, func() {
					Minimum(1)
				})
				Required("q")
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
		Method("MethodQueryString", func() {
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
		Method("MethodQueryStringValidate", func() {
			Payload(func() {
				Attribute("q", String, func() {
					Enum("val")
				})
				Required("q")
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
		Method("MethodQueryBytes", func() {
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
		Method("MethodQueryBytesValidate", func() {
			Payload(func() {
				Attribute("q", Bytes, func() {
					MinLength(1)
				})
				Required("q")
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
		Method("MethodQueryAny", func() {
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
		Method("MethodQueryAnyValidate", func() {
			Payload(func() {
				Attribute("q", Any, func() {
					Enum("val", 1)
				})
				Required("q")
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
		Method("MethodQueryArrayBool", func() {
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
		Method("MethodQueryArrayBoolValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Boolean), func() {
					MinLength(1)
					Elem(func() {
						Enum(true)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayInt", func() {
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
		Method("MethodQueryArrayIntValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayInt32", func() {
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
		Method("MethodQueryArrayInt32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayInt64", func() {
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
		Method("MethodQueryArrayInt64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Int64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayUInt", func() {
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
		Method("MethodQueryArrayUIntValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayUInt32", func() {
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
		Method("MethodQueryArrayUInt32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayUInt64", func() {
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
		Method("MethodQueryArrayUInt64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(UInt64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayFloat32", func() {
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
		Method("MethodQueryArrayFloat32Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float32), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayFloat64", func() {
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
		Method("MethodQueryArrayFloat64Validate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Float64), func() {
					MinLength(1)
					Elem(func() {
						Minimum(1)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayString", func() {
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
		Method("MethodQueryArrayStringValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(String), func() {
					MinLength(1)
					Elem(func() {
						Enum("val")
					})
				})
				Required("q")
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
		Method("MethodQueryArrayBytes", func() {
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
		Method("MethodQueryArrayBytesValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Bytes), func() {
					MinLength(1)
					Elem(func() {
						MinLength(2)
					})
				})
				Required("q")
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
		Method("MethodQueryArrayAny", func() {
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
		Method("MethodQueryArrayAnyValidate", func() {
			Payload(func() {
				Attribute("q", ArrayOf(Any), func() {
					MinLength(1)
					Elem(func() {
						Enum("val", 1)
					})
				})
				Required("q")
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryPrimitiveStringValidateDSL = func() {
	Service("ServiceQueryPrimitiveStringValidate", func() {
		Method("MethodQueryPrimitiveStringValidate", func() {
			Payload(String, func() {
				Enum("val")
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryPrimitiveBoolValidateDSL = func() {
	Service("ServiceQueryPrimitiveBoolValidate", func() {
		Method("MethodQueryPrimitiveBoolValidate", func() {
			Payload(Boolean, func() {
				Enum(true)
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryPrimitiveArrayStringValidateDSL = func() {
	Service("ServiceQueryPrimitiveArrayStringValidate", func() {
		Method("MethodQueryPrimitiveArrayStringValidate", func() {
			Payload(ArrayOf(String), func() {
				MinLength(1)
				Elem(func() {
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

var PayloadQueryPrimitiveArrayBoolValidateDSL = func() {
	Service("ServiceQueryPrimitiveArrayBoolValidate", func() {
		Method("MethodQueryPrimitiveArrayBoolValidate", func() {
			Payload(ArrayOf(Boolean), func() {
				MinLength(1)
				Elem(func() {
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

var PayloadQueryStringMappedDSL = func() {
	Service("ServiceQueryStringMapped", func() {
		Method("MethodQueryStringMapped", func() {
			Payload(func() {
				Attribute("query")
			})
			HTTP(func() {
				GET("/")
				Param("query:q")
			})
		})
	})
}

var PayloadQueryStringDefaultDSL = func() {
	Service("ServiceQueryStringDefault", func() {
		Method("MethodQueryStringDefault", func() {
			Payload(func() {
				Attribute("q", func() {
					Default("def")
				})
			})
			HTTP(func() {
				GET("/")
				Param("q")
			})
		})
	})
}

var PayloadQueryPrimitiveStringDefaultDSL = func() {
	Service("ServiceQueryPrimitiveStringDefault", func() {
		Method("MethodQueryPrimitiveStringDefault", func() {
			Payload(String, func() {
				Default("def")
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
		Method("MethodPathString", func() {
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
		Method("MethodPathStringValidate", func() {
			Payload(func() {
				Attribute("p", String, func() {
					Enum("val")
				})
				Required("p")
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathArrayStringDSL = func() {
	Service("ServicePathArrayString", func() {
		Method("MethodPathArrayString", func() {
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
		Method("MethodPathArrayStringValidate", func() {
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

var PayloadPathPrimitiveStringValidateDSL = func() {
	Service("ServicePathPrimitiveStringValidate", func() {
		Method("MethodPathPrimitiveStringValidate", func() {
			Payload(String, func() {
				Enum("val")
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathPrimitiveBoolValidateDSL = func() {
	Service("ServicePathPrimitiveBoolValidate", func() {
		Method("MethodPathPrimitiveBoolValidate", func() {
			Payload(Boolean, func() {
				Enum(true)
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathPrimitiveArrayStringValidateDSL = func() {
	Service("ServicePathPrimitiveArrayStringValidate", func() {
		Method("MethodPathPrimitiveArrayStringValidate", func() {
			Payload(ArrayOf(String), func() {
				MinLength(1)
				Elem(func() {
					Enum("val")
				})
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadPathPrimitiveArrayBoolValidateDSL = func() {
	Service("ServicePathPrimitiveArrayBoolValidate", func() {
		Method("MethodPathPrimitiveArrayBoolValidate", func() {
			Payload(ArrayOf(Boolean), func() {
				MinLength(1)
				Elem(func() {
					Enum(true)
				})
			})
			HTTP(func() {
				GET("/{p}")
			})
		})
	})
}

var PayloadHeaderStringDSL = func() {
	Service("ServiceHeaderString", func() {
		Method("MethodHeaderString", func() {
			Payload(func() {
				Attribute("h", String)
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderStringValidateDSL = func() {
	Service("ServiceHeaderStringValidate", func() {
		Method("MethodHeaderStringValidate", func() {
			Payload(func() {
				Attribute("h", String, func() {
					Pattern("header")
				})
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderArrayStringDSL = func() {
	Service("ServiceHeaderArrayString", func() {
		Method("MethodHeaderArrayString", func() {
			Payload(func() {
				Attribute("h", ArrayOf(String))
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderArrayStringValidateDSL = func() {
	Service("ServiceHeaderArrayStringValidate", func() {
		Method("MethodHeaderArrayStringValidate", func() {
			Payload(func() {
				Attribute("h", ArrayOf(String, func() {
					Enum("val")
				}))
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderPrimitiveStringValidateDSL = func() {
	Service("ServiceHeaderPrimitiveStringValidate", func() {
		Method("MethodHeaderPrimitiveStringValidate", func() {
			Payload(String, func() {
				Enum("val")
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderPrimitiveBoolValidateDSL = func() {
	Service("ServiceHeaderPrimitiveBoolValidate", func() {
		Method("MethodHeaderPrimitiveBoolValidate", func() {
			Payload(Boolean, func() {
				Enum(true)
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderPrimitiveArrayStringValidateDSL = func() {
	Service("ServiceHeaderPrimitiveArrayStringValidate", func() {
		Method("MethodHeaderPrimitiveArrayStringValidate", func() {
			Payload(ArrayOf(String), func() {
				MinLength(1)
				Elem(func() {
					Pattern("val")
				})
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderPrimitiveArrayBoolValidateDSL = func() {
	Service("ServiceHeaderPrimitiveArrayBoolValidate", func() {
		Method("MethodHeaderPrimitiveArrayBoolValidate", func() {
			Payload(ArrayOf(Boolean), func() {
				MinLength(1)
				Elem(func() {
					Enum(true)
				})
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderStringDefaultDSL = func() {
	Service("ServiceHeaderStringDefault", func() {
		Method("MethodHeaderStringDefault", func() {
			Payload(func() {
				Attribute("h", String, func() {
					Default("def")
				})
			})
			HTTP(func() {
				GET("/")
				Header("h")
			})
		})
	})
}

var PayloadHeaderPrimitiveStringDefaultDSL = func() {
	Service("ServiceHeaderPrimitiveStringDefault", func() {
		Method("MethodHeaderPrimitiveStringDefault", func() {
			Payload(String, func() {
				Default("def")
			})
			HTTP(func() {
				GET("")
				Header("h")
			})
		})
	})
}

var PayloadBodyStringDSL = func() {
	Service("ServiceBodyString", func() {
		Method("MethodBodyString", func() {
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
		Method("MethodBodyStringValidate", func() {
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
		Method("MethodBodyUser", func() {
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
		Method("MethodBodyUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyArrayStringDSL = func() {
	Service("ServiceBodyArrayString", func() {
		Method("MethodBodyArrayString", func() {
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
		Method("MethodBodyArrayStringValidate", func() {
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
		Method("MethodBodyArrayUser", func() {
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
		Method("MethodBodyArrayUserValidate", func() {
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
		Method("MethodBodyMapString", func() {
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
		Method("MethodBodyMapStringValidate", func() {
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
		Method("MethodBodyMapUser", func() {
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
		Method("MethodBodyMapUserValidate", func() {
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

var PayloadBodyPrimitiveStringValidateDSL = func() {
	Service("ServiceBodyPrimitiveStringValidate", func() {
		Method("MethodBodyPrimitiveStringValidate", func() {
			Payload(String, func() {
				Enum("val")
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyPrimitiveBoolValidateDSL = func() {
	Service("ServiceBodyPrimitiveBoolValidate", func() {
		Method("MethodBodyPrimitiveBoolValidate", func() {
			Payload(Boolean, func() {
				Enum(true)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyPrimitiveArrayStringValidateDSL = func() {
	Service("ServiceBodyPrimitiveArrayStringValidate", func() {
		Method("MethodBodyPrimitiveArrayStringValidate", func() {
			Payload(ArrayOf(String), func() {
				MinLength(1)
				Elem(func() {
					Enum("val")
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyPrimitiveArrayBoolValidateDSL = func() {
	Service("ServiceBodyPrimitiveArrayBoolValidate", func() {
		Method("MethodBodyPrimitiveArrayBoolValidate", func() {
			Payload(ArrayOf(Boolean), func() {
				MinLength(1)
				Elem(func() {
					Enum(true)
				})
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyPrimitiveArrayUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("pattern")
		})
		Required("a")
	})
	Service("ServiceBodyPrimitiveArrayUserValidate", func() {
		Method("MethodBodyPrimitiveArrayUserValidate", func() {
			Payload(ArrayOf(PayloadType), func() {
				MinLength(1)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyPrimitiveFieldEmptyDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", ArrayOf(String))
	})
	Service("ServiceBodyPrimitiveArrayUser", func() {
		Method("MethodBodyPrimitiveArrayUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
				Param("a")
				Body(Empty)
			})
		})
	})
}

var PayloadBodyPrimitiveFieldArrayUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", ArrayOf(String))
	})
	Service("ServiceBodyPrimitiveArrayUser", func() {
		Method("MethodBodyPrimitiveArrayUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
				Body("a")
			})
		})
	})
}

var PayloadBodyPrimitiveFieldArrayUserValidateDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", ArrayOf(String), func() {
			MinLength(1)
			Elem(func() {
				Pattern("pattern")
			})
		})
		Required("a")
	})
	Service("ServiceBodyPrimitiveArrayUserValidate", func() {
		Method("MethodBodyPrimitiveArrayUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
				Body("a")
			})
		})
	})
}

var PayloadBodyQueryObjectDSL = func() {
	Service("ServiceBodyQueryObject", func() {
		Method("MethodBodyQueryObject", func() {
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
		Method("MethodBodyQueryObjectValidate", func() {
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
		Method("MethodBodyQueryUser", func() {
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
		Method("MethodBodyQueryUserValidate", func() {
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
		Method("MethodBodyPathObject", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
			})
			HTTP(func() {
				POST("/{b}")
			})
		})
	})
}

var PayloadBodyPathObjectValidateDSL = func() {
	Service("ServiceBodyPathObjectValidate", func() {
		Method("MethodBodyPathObjectValidate", func() {
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
				POST("/{b}")
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
		Method("MethodBodyPathUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{b}")
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
		Method("MethodUserBodyPathValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{b}")
			})
		})
	})
}

var PayloadBodyQueryPathObjectDSL = func() {
	Service("ServiceBodyQueryPathObject", func() {
		Method("MethodBodyQueryPathObject", func() {
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
		Method("MethodBodyQueryPathObjectValidate", func() {
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
		Method("MethodBodyQueryPathUser", func() {
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
		Method("MethodBodyQueryPathUserValidate", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{c}")
				Param("b")
			})
		})
	})
}

var PayloadBodyUserInnerDSL = func() {
	var InnerType = Type("InnerType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a")
	})
	var PayloadType = Type("PayloadType", func() {
		Attribute("inner", InnerType)
	})
	Service("ServiceBodyUserInner", func() {
		Method("MethodBodyUserInner", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyUserInnerDefaultDSL = func() {
	var InnerType = Type("InnerType", func() {
		Attribute("a", String, func() {
			Default("defaulta")
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Default("defaultb")
			Pattern("patternb")
		})
		Required("a")
	})
	var PayloadType = Type("PayloadType", func() {
		Attribute("inner", InnerType)
	})
	Service("ServiceBodyUserInnerDefault", func() {
		Method("MethodBodyUserInnerDefault", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyInlineArrayUserDSL = func() {
	var ElemType = Type("ElemType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a")
	})
	Service("ServiceBodyInlineArrayUser", func() {
		Method("MethodBodyInlineArrayUser", func() {
			Payload(ArrayOf(ElemType))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyInlineMapUserDSL = func() {
	var KeyType = Type("KeyType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a")
	})
	var ElemType = Type("ElemType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Required("a")
	})
	Service("ServiceBodyInlineMapUser", func() {
		Method("MethodBodyInlineMapUser", func() {
			Payload(MapOf(KeyType, ElemType))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PayloadBodyInlineRecursiveUserDSL = func() {
	var PayloadType = Type("PayloadType", func() {
		Attribute("a", String, func() {
			Pattern("patterna")
		})
		Attribute("b", String, func() {
			Pattern("patternb")
		})
		Attribute("c", "PayloadType")
		Required("a", "c")
	})

	Service("ServiceBodyInlineRecursiveUser", func() {
		Method("MethodBodyInlineRecursiveUser", func() {
			Payload(PayloadType)
			HTTP(func() {
				POST("/{a}")
				Param("b")
			})
		})
	})
}
