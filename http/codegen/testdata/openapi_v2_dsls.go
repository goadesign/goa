package testdata

import . "goa.design/goa/v3/dsl"

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
	Service("httpService", func() {
		Method("httpEndpoint", func() {
			HTTP(func() { GET("/") })
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

var FileServiceDSL = func() {
	var _ = Service("service-name", func() {
		Files("path1", "filename")
		Files("path2", "filename", func() {
			Meta("swagger:tag:user-tag")
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
	var PayloadT = Type("Payload", func() {
		Attribute("string", String, func() {
			Example("")
			Meta("swagger:extension:x-test-schema", "Payload")
		})
	})
	var ResultT = Type("Result", func() {
		Attribute("string", String, func() {
			Example("")
			Meta("swagger:extension:x-test-schema", "Result")
		})
	})
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
		Meta("swagger:extension:x-test-api", "API")
		Meta("swagger:tag:Backend")
		Meta("swagger:tag:Backend:desc", "Description of Backend")
		Meta("swagger:tag:Backend:url", "http://example.com")
		Meta("swagger:tag:Backend:url:desc", "See more docs here")
		Meta("swagger:tag:Backend:extension:x-data", `{"foo":"bar"}`)
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(PayloadT)
			Result(ResultT)
			HTTP(func() {
				POST("/")
				Meta("swagger:extension:x-test-foo", "bar")
			})
			Meta("swagger:extension:x-test-operation", "Operation")
		})
	})
}

var SecurityDSL = func() {
	var JWTAuth = JWTSecurity("jwt", func() {
		Description(`Secures endpoint by requiring a valid JWT token retrieved via the signin endpoint. Supports scopes "api:read" and "api:write".`)
		Scope("api:read", "Read-only access")
		Scope("api:write", "Read and write access")
	})

	var APIKeyAuth = APIKeySecurity("api_key", func() {
		Description("Secures endpoint by requiring an API key.")
	})

	var BasicAuth = BasicAuthSecurity("basic", func() {
		Description("Basic authentication used to authenticate security principal during signin")
	})

	var OAuth2Auth = OAuth2Security("oauth2", func() {
		AuthorizationCodeFlow("http://goa.design/authorization", "http://goa.design/token", "http://goa.design/refresh")
		Description(`Secures endpoint by requiring a valid OAuth2 token retrieved via the signin endpoint. Supports scopes "api:read" and "api:write".`)
		Scope("api:read", "Read-only access")
		Scope("api:write", "Read and write access")
	})

	Service("testService", func() {
		Method("testEndpointA", func() {
			Security(BasicAuth, OAuth2Auth, JWTAuth, APIKeyAuth, func() {
				Scope("api:read")
			})
			Payload(func() {
				Username("username", String)
				Password("password", String)
				APIKey("api_key", "key", String)
				Token("token", String)
				AccessToken("oauth_token", String)
				Required("username", "password", "key", "token", "oauth_token")
			})
			HTTP(func() {
				GET("/")
				Header("oauth_token:Token")
				Param("key:k")
				Header("token:X-Authorization")
			})
		})
		Method("testEndpointB", func() {
			Security(APIKeyAuth)
			Security(OAuth2Auth, func() {
				Scope("api:read")
				Scope("api:write")
			})
			Payload(func() {
				APIKey("api_key", "key", String)
				AccessToken("oauth_token", String)
				Required("key", "oauth_token")
			})
			HTTP(func() {
				POST("/")
				Param("oauth_token:auth")
				Header("key:Authorization")
			})
		})
	})
}

var ServerHostWithVariablesDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://{version}.goa.design")
				Variable("version", String, "API Version", func() {
					Default("v1")
				})
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var WithSpacesDSL = func() {
	var Bar = Type("bar", func() {
		Attribute("string", String, func() {
			Example("")
		})
	})
	var FooBar = ResultType("application/vnd.goa.foobar", func() {
		TypeName("Foo Bar")
		Attribute("foo", String, func() {
			Example("")
		})
		Attribute("bar", ArrayOf(Bar))
	})
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(Bar)
			Result(FooBar)
			HTTP(func() {
				POST("/")
				Response(StatusOK)
				Response(StatusNotFound)
			})
		})
	})
}

var WithMapDSL = func() {
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("int_map", MapOf(Int, String))
				Attribute("uint_map", MapOf(UInt, String))
			})
			Result(func() {
				Attribute("uint32_map", MapOf(UInt32, String))
				Attribute("uint64_map", MapOf(UInt64, String))
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var PathWithWildcardDSL = func() {
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				POST("/{*int_map}")
			})
		})
	})
}
