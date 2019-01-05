package testdata

import . "goa.design/goa/dsl"

var SimpleDSL = func() {
	var PayloadT = Type("Payload", func() {
		Attribute("string", String, func() {
			Example("")
		})
	})
	var ResultT = Type("Result", func() {
		Attribute("string", String, func() {
			Example("")
		})
	})
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(PayloadT)
			Result(ResultT)
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var MultipleServicesDSL = func() {
	var PayloadT = Type("Payload", func() {
		Attribute("string", String, func() {
			Example("")
		})
	})
	var ResultT = Type("Result", func() {
		Attribute("string", String, func() {
			Example("")
		})
	})
	var _ = API("test", func() {
		Server("test", func() {
			Services("testService", "anotherTestService")
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(PayloadT)
			Result(ResultT)
			HTTP(func() {
				GET("/")
			})
		})
	})
	Service("anotherTestService", func() {
		Method("testEndpoint", func() {
			Payload(PayloadT)
			Result(ResultT)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var MultipleViewsDSL = func() {
	var ResultT = ResultType("application/json", func() {
		ContentType("application/vnd.custom+json")
		TypeName("Result")
		Attributes(func() {
			Attribute("string", String, func() {
				Example("")
			})
			Attribute("int", Int, func() {
				Example(1)
			})
		})
		View("default", func() {
			Attribute("string")
			Attribute("int")
		})
		View("tiny", func() {
			Attribute("string")
		})
	})
	Service("testService", func() {
		Method("testEndpointDefault", func() {
			Result(ResultT)
			HTTP(func() {
				GET("/")
				Response(StatusOK, func() {
					ContentType("application/custom+json")
				})
			})
		})
		Method("testEndpointTiny", func() {
			Result(ResultT)
			HTTP(func() {
				GET("/tiny")
			})
		})
	})
}

var ExplicitViewDSL = func() {
	var ResultT = ResultType("application/json", func() {
		TypeName("Result")
		Attributes(func() {
			Attribute("string", String, func() {
				Example("")
			})
			Attribute("int", Int, func() {
				Example(1)
			})
		})
		View("tiny", func() {
			Attribute("string")
		})
	})
	Service("testService", func() {
		Method("testEndpointDefault", func() {
			Result(ResultT, func() {
				View("default")
			})
			HTTP(func() {
				GET("/")
			})
		})
		Method("testEndpointTiny", func() {
			Result(ResultT, func() {
				View("tiny")
			})
			HTTP(func() {
				GET("/tiny")
			})
		})
	})
}

var InvalidDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("http://[::1]:namedport") // invalid URL
			})
		})
	})
}

var EmptyDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
}

var StringValidationDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(String, func() {
				MinLength(0)
				MaxLength(42)
				Example("")
			})
			Result(String, func() {
				MinLength(0)
				MaxLength(42)
				Example("")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var IntValidationDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Int, func() {
				Minimum(0)
				Maximum(42)
				Example(1)
			})
			Result(Int, func() {
				Minimum(0)
				Maximum(42)
				Example(1)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ArrayValidationDSL = func() {
	var Bar = Type("bar", func() {
		Attribute("string", String, func() {
			MinLength(0)
			MaxLength(42)
			Example("")
		})
	})
	var FooBar = Type("foobar", func() {
		Attribute("foo", ArrayOf(String), func() {
			MinLength(0)
			MaxLength(42)
		})
		Attribute("bar", ArrayOf(Bar), func() {
			MinLength(0)
			MaxLength(42)
		})
	})
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(ArrayOf(FooBar))
			Result(String, func() {
				MinLength(0)
				MaxLength(42)
				Example("")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ExtensionDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				POST("/")
				Meta("swagger:extension:x-test-foo", "bar")
			})
		})
	})
}
