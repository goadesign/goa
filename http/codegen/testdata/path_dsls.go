package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var PathNoParamDSL = func() {
	Service("ServicePathNoParam", func() {
		Method("MethodPathNoParam", func() {
			HTTP(func() {
				GET("/one/two")
			})
		})
	})
}

var BasePathNoTrailing_SlashWithBasePathNoTrailingDSL = func() {
	Service("BasePathNoTrailing", func() {
		HTTP(func() {
			Path("foo")
		})
		Method("SlashWithBasePathNoTrailing", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var BasePathNoTrailing_TrailingWithBasePathNoTrailingDSL = func() {
	Service("BasePathNoTrailing", func() {
		HTTP(func() {
			Path("foo")
		})
		Method("TrailingWithBasePathNoTrailing", func() {
			HTTP(func() {
				GET("/bar/")
			})
		})
	})
}

var BasePathWithTrailingSlash_WithBasePathWithTrailingDSL = func() {
	Service("BasePathWithTrailing", func() {
		HTTP(func() {
			Path("foo/")
		})
		Method("SlashWithBasePathWithTrailing", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var NoBasePath_SlashNoBasePathDSL = func() {
	Service("NoBasePath", func() {
		Method("SlashNoBasePath", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var NoBasePath_TrailingNoBasePathDSL = func() {
	Service("NoBasePath", func() {
		Method("TrailingNoBasePath", func() {
			HTTP(func() {
				GET("/foo/")
			})
		})
	})
}

var BasePath_SpecialTrailingSlashDSL = func() {
	Service("BasePath", func() {
		HTTP(func() {
			Path("/foo")
		})
		Method("SpecialTrailingSlash", func() {
			HTTP(func() {
				GET("/./")
			})
		})
	})
}

var PathOneParamDSL = func() {
	Service("ServicePathOneParam", func() {
		Method("MethodPathOneParam", func() {
			Payload(String)
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathMultipleParamsDSL = func() {
	Service("ServicePathMultipleParam", func() {
		Method("MethodPathMultipleParam", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
			})
			HTTP(func() {
				GET("one/{a}/two/{b}/three")
			})
		})
	})
}

var PathAlternativesDSL = func() {
	Service("ServicePathAlternatives", func() {
		Method("MethodPathAlternatives", func() {
			Payload(func() {
				Attribute("a", String)
				Attribute("b", String)
			})
			HTTP(func() {
				GET("one/{a}/two/{b}/three")
				POST("one/two/{b}/three/{a}")
			})
		})
	})
}

var PathStringSliceParamDSL = func() {
	Service("ServicePathStringSliceParam", func() {
		Method("MethodPathStringSliceParam", func() {
			Payload(ArrayOf(String))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathIntSliceParamDSL = func() {
	Service("ServicePathIntSliceParam", func() {
		Method("MethodPathIntSliceParam", func() {
			Payload(ArrayOf(Int))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathInt32SliceParamDSL = func() {
	Service("ServicePathInt32SliceParam", func() {
		Method("MethodPathInt32SliceParam", func() {
			Payload(ArrayOf(Int32))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathInt64SliceParamDSL = func() {
	Service("ServicePathInt64SliceParam", func() {
		Method("MethodPathInt64SliceParam", func() {
			Payload(ArrayOf(Int64))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathUintSliceParamDSL = func() {
	Service("ServicePathUintSliceParam", func() {
		Method("MethodPathUintSliceParam", func() {
			Payload(ArrayOf(UInt))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathUint32SliceParamDSL = func() {
	Service("ServicePathUint32SliceParam", func() {
		Method("MethodPathUint32SliceParam", func() {
			Payload(ArrayOf(UInt32))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathUint64SliceParamDSL = func() {
	Service("ServicePathUint64SliceParam", func() {
		Method("MethodPathUint64SliceParam", func() {
			Payload(ArrayOf(UInt64))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathFloat32SliceParamDSL = func() {
	Service("ServicePathFloat32SliceParam", func() {
		Method("MethodPathFloat32SliceParam", func() {
			Payload(ArrayOf(Float32))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathFloat64SliceParamDSL = func() {
	Service("ServicePathFloat64SliceParam", func() {
		Method("MethodPathFloat64SliceParam", func() {
			Payload(ArrayOf(Float64))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathBoolSliceParamDSL = func() {
	Service("ServicePathBoolSliceParam", func() {
		Method("MethodPathBoolSliceParam", func() {
			Payload(ArrayOf(Boolean))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}

var PathInterfaceSliceParamDSL = func() {
	Service("ServicePathInterfaceSliceParam", func() {
		Method("MethodPathInterfaceSliceParam", func() {
			Payload(ArrayOf(Any))
			HTTP(func() {
				GET("one/{a}/two")
			})
		})
	})
}
