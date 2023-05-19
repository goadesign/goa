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
			Result(ResultT, func() {
				View("tiny")
			})
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
			Meta("openapi:tag:user-tag")
		})
	})
}

var FileServiceSwaggerDSL = func() {
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
			Meta("openapi:extension:x-test-schema", "Payload")
		})
	})
	var ResultT = Type("Result", func() {
		Attribute("string", String, func() {
			Example("")
			Meta("openapi:extension:x-test-schema", "Result")
		})
	})
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
		Meta("openapi:extension:x-test-api", "API")
		Meta("openapi:tag:Backend")
		Meta("openapi:tag:Backend:desc", "Description of Backend")
		Meta("openapi:tag:Backend:url", "http://example.com")
		Meta("openapi:tag:Backend:url:desc", "See more docs here")
		Meta("openapi:tag:Backend:extension:x-data", `{"foo":"bar"}`)
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(PayloadT)
			Result(ResultT)
			HTTP(func() {
				POST("/")
				Meta("openapi:extension:x-test-foo", "bar")
			})
			Meta("openapi:extension:x-test-operation", "Operation")
		})
	})
}

var ExtensionSwaggerDSL = func() {
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
			Payload(func() {
				Attribute("int_map", MapOf(String, Int, func() {
					Key(func() { Example("") })
					Elem(func() { Example(1) })
				}))
				Attribute("uint_map", MapOf(String, UInt, func() {
					Key(func() { Example("") })
					Elem(func() { Example(uint(1)) })
				}))
				Attribute("type_map", MapOf(String, Bar), func() {
					Key(func() { Example("") })
				})
			})
			Result(func() {
				Attribute("uint32_map", MapOf(String, UInt32, func() {
					Key(func() { Example("") })
					Elem(func() { Example(uint32(1)) })
				}))
				Attribute("uint64_map", MapOf(String, UInt64, func() {
					Key(func() { Example("") })
					Elem(func() { Example(uint64(1)) })
				}))
				Attribute("resulttype_map", MapOf(String, FooBar, func() {
					Key(func() { Example("") })
				}))
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

var WithTagsDSL = func() {
	Service("test service", func() {
		HTTP(func() {
			Meta("openapi:tag:SomeTag:desc", "Endpoint description")
			Meta("openapi:tag:SomeTag:url", "Endpoint URL")
			Meta("openapi:tag:AnotherTag:desc", "Endpoint description")
			Meta("openapi:tag:AnotherTag:url", "Endpoint URL")
		})
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("openapi:tag:SomeTag")
				POST("/{*int_map}")
			})
		})
		Method("another test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("openapi:generate", "false")
				Meta("openapi:tag:AnotherTag")
				POST("/{*int_map}")
			})
		})
	})
	Service("another test service", func() {
		Meta("openapi:generate", "false")
		HTTP(func() {
			Meta("openapi:tag:AnotherService:desc", "Another service description")
		})
		Method("another test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("openapi:tag:AnotherService")
				POST("/{*int_map}")
			})
		})
	})
}

var WithTagsSwaggerDSL = func() {
	Service("test service", func() {
		HTTP(func() {
			Meta("swagger:tag:SomeTag:desc", "Endpoint description")
			Meta("swagger:tag:SomeTag:url", "Endpoint URL")
			Meta("swagger:tag:AnotherTag:desc", "Endpoint description")
			Meta("swagger:tag:AnotherTag:url", "Endpoint URL")
		})
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("swagger:tag:SomeTag")
				POST("/{*int_map}")
			})
		})
		Method("another test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("swagger:generate", "false")
				Meta("swagger:tag:AnotherTag")
				POST("/{*int_map}")
			})
		})
	})
	Service("another test service", func() {
		Meta("swagger:generate", "false")
		HTTP(func() {
			Meta("swagger:tag:AnotherService:desc", "Another service description")
		})
		Method("another test endpoint", func() {
			Payload(func() {
				Attribute("int_map", Int)
			})
			HTTP(func() {
				Meta("swagger:tag:AnotherService")
				POST("/{*int_map}")
			})
		})
	})
}

var TypenameDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})

	var Foo = Type("Foo", func() {
		Meta("openapi:typename", "FooPayload")
		Attribute("value", String, func() {
			Example("")
		})
	})

	var Bar = ResultType("application/vnd.goa.example.bar", func() {
		TypeName("Bar")
		Meta("openapi:typename", "BarResult")
		Attribute("value", String, func() {
			Example("")
		})
	})

	var _ = Service("testService", func() {
		Method("foo", func() {
			Payload(Foo)
			Result(Bar, func() {
				Meta("openapi:typename", "FooResult")
			})
			HTTP(func() {
				POST("/foo")
			})
		})
		Method("bar", func() {
			Payload(Foo, func() {
				Meta("openapi:typename", "BarPayload")
			})
			Result(Bar)
			HTTP(func() {
				POST("/bar")
			})
		})
		Method("baz", func() {
			Payload(func() {
				Meta("openapi:typename", "BazPayload")
				Attribute("value", String, func() {
					Example("")
				})
			})
			Result(func() {
				Meta("openapi:typename", "BazResult")
				Attribute("value", String, func() {
					Example("")
				})
			})
			HTTP(func() {
				POST("/baz")
			})
		})
	})
}

var SkipResponseBodyEncodeDecodeDSL = func() {
	Service("testService", func() {
		Method("empty", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				GET("/empty")
			})
		})
		Method("empty_ok", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				GET("/empty/ok")
				Response(StatusOK)
			})
		})
		Method("binary", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				GET("/binary")
				SkipResponseBodyEncodeDecode()
				Response(StatusOK, func() {
					ContentType("image/png")
				})
			})
		})
	})
}
