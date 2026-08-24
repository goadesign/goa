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

// BytesExampleDSL defines a response whose OpenAPI example must remain a
// string in both JSON and YAML documents.
var BytesExampleDSL = func() {
	var _ = API("bytes", func() {
		Server("bytes", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("bytes", func() {
		Method("download", func() {
			Result(Bytes, func() {
				Example([]byte("hello"))
			})
			HTTP(func() {
				GET("/download")
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

// ReleasedResponseCollectionNamesDSL exercises the public OpenAPI component
// names for response collections whose elements use fixed views.
var ReleasedResponseCollectionNamesDSL = func() {
	var StoredBottle = ResultType("application/vnd.stored-bottle", func() {
		TypeName("StoredBottle")
		Attributes(func() {
			Attribute("name", String, func() {
				Example("Blue's Cuvee")
			})
			Attribute("vintage", UInt32, func() {
				Example(2003)
			})
			Required("name", "vintage")
		})
		View("default", func() {
			Attribute("name")
			Attribute("vintage")
		})
		View("tiny", func() {
			Attribute("name")
		})
	})

	Service("storage", func() {
		Method("list_default", func() {
			Result(CollectionOf(StoredBottle), func() {
				View("default")
			})
			HTTP(func() {
				GET("/default")
			})
		})
		Method("list_tiny", func() {
			Result(CollectionOf(StoredBottle), func() {
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

// FileServiceWildcardDSL defines a service with a file server using a wildcard path.
var FileServiceWildcardDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("http://localhost:80")
			})
		})
	})
	var _ = Service("front", func() {
		Files("/ui/{*filepath}", "ui/dist", func() {
			Meta("openapi:summary", "Download ui/dist")
			Meta("openapi:tag:front")
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
		Example(Val{"string": "item"})
		Attribute("string", String, func() {
			MinLength(0)
			MaxLength(42)
			Example("item")
		})
	})
	var FooBar = Type("foobar", func() {
		Example(Val{"foo": []any{"item"}, "bar": []any{Val{"string": "item"}}})
		Attribute("foo", ArrayOf(String, func() {
			Example("item")
		}), func() {
			MinLength(0)
			MaxLength(42)
			Example([]any{"item"})
		})
		Attribute("bar", ArrayOf(Bar), func() {
			MinLength(0)
			MaxLength(42)
			Example([]any{Val{"string": "item"}})
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
			Payload(ArrayOf(FooBar), func() {
				Example([]any{Val{"foo": []any{"item"}, "bar": []any{Val{"string": "item"}}}})
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

var BearerSecurityDSL = func() {
	var BearerAuth = BearerSecurity("bearer", func() {
		Description("Opaque session token or trusted OIDC access token.")
		Scope("api:read", "Read-only access")
	})

	var FormattedBearerAuth = BearerSecurity("formatted_bearer", func() {
		Description("JWT access token.")
		BearerFormat("JWT")
	})

	Service("testService", func() {
		Method("plainBearer", func() {
			Security(BearerAuth, func() {
				Scope("api:read")
			})
			Payload(func() {
				BearerToken("token", String)
				Required("token")
			})
			HTTP(func() {
				GET("/plain")
			})
		})
		Method("formattedBearer", func() {
			Security(FormattedBearerAuth)
			Payload(func() {
				BearerToken("token", String)
				Required("token")
			})
			HTTP(func() {
				GET("/formatted")
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
		Example(Val{"string": "item"})
		Attribute("string", String, func() {
			Example("")
		})
	})
	var FooBar = ResultType("application/vnd.goa.foobar", func() {
		TypeName("Foo Bar")
		Attribute("foo", String, func() {
			Example("")
		})
		Attribute("bar", ArrayOf(Bar), func() {
			Example([]any{Val{"string": "item"}})
		})
	})
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(Bar)
			Result(FooBar, func() {
				Example(Val{"foo": "", "bar": []any{Val{"string": "item"}}})
			})
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

var WithAnyDSL = func() {
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(func() {
				Example(Val{"any": "", "any_array": []any{""}, "any_map": Val{"key": ""}})
				Attribute("any", Any, func() {
					Example("")
				})
				Attribute("any_array", ArrayOf(Any, func() {
					Example("")
				}), func() {
					Example([]any{""})
				})
				Attribute("any_map", MapOf(String, Any), func() {
					Example(Val{"key": ""})
					Key(func() { Example("") })
					Elem(func() { Example("") })
				})
			})
			Result(func() {
				Example(Val{"any": "", "any_array": []any{""}, "any_map": Val{"key": ""}})
				Attribute("any", Any, func() {
					Example("")
				})
				Attribute("any_array", ArrayOf(Any, func() {
					Example("")
				}), func() {
					Example([]any{""})
				})
				Attribute("any_map", MapOf(String, Any), func() {
					Example(Val{"key": ""})
					Key(func() { Example("") })
					Elem(func() { Example("") })
				})
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
				Attribute("int_map", Int, func() {
					Example(1)
				})
			})
			HTTP(func() {
				POST("/{*int_map}")
			})
		})
	})
}

var PathWithMultipleWildcardDSL = func() {
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("foo", Int, func() {
					Example(1)
				})
				Attribute("bar", Int, func() {
					Example(2)
				})
			})
			HTTP(func() {
				POST("/{bar}")
			})
		})
		HTTP(func() {
			Path("/{foo}")
		})
	})
}

var PathWithMultipleExplicitWildcardDSL = func() {
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("foo", Int, func() {
					Example(1)
				})
				Attribute("bar", Int, func() {
					Example(2)
				})
			})
			HTTP(func() {
				POST("/{bar}")
				Param("bar")
			})
		})
		HTTP(func() {
			Path("/{foo}")
			Param("foo")
		})
	})
}

var HeadersDSL = func() {
	Service("test service", func() {
		Method("test endpoint", func() {
			Payload(func() {
				Attribute("foo", Int, func() {
					Example(1)
				})
				Attribute("bar", Int, func() {
					Example(2)
				})
			})
			HTTP(func() {
				POST("/")
				Header("bar")
			})
		})
		HTTP(func() {
			Header("foo")
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
				Attribute("int_map", Int, func() {
					Example(1)
				})
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
				Attribute("int_map", Int, func() {
					Example(1)
				})
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

var NotGenerateServerDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
			Meta("openapi:generate", "false")
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Result(String, func() {
				Example("ok")
			})
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var NotGenerateHostDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
				Meta("openapi:generate", "false")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Result(String, func() {
				Example("ok")
			})
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var NotGenerateAttributeDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	var PayloadT = Type("Payload", func() {
		Attribute("int", Int, func() {
			Meta("openapi:generate", "false")
		})
		Attribute("string", String, func() {
			Example("")
		})
		Attribute("required_int", Int, func() {
			Meta("openapi:generate", "false")
		})
		Attribute("required_string", String, func() {
			Example("")
		})
		Required("required_int", "required_string")
	})
	var ResultT = Type("Result", func() {
		Attribute("int", Int, func() {
			Example(0)
		})
		Attribute("string", String, func() {
			Meta("openapi:generate", "false")
		})
		Attribute("required_int", Int, func() {
			Example(0)
		})
		Attribute("required_string", String, func() {
			Meta("openapi:generate", "false")
		})
		Required("required_int", "required_string")
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

var JSONPrefixDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
		Meta("openapi:json:prefix", "  ")
	})
	var _ = Service("service-name", func() {
		Files("path1", "filename")
		Files("path2", "filename", func() {
			Meta("openapi:tag:user-tag")
		})
	})
}

var JSONIndentDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
		Meta("openapi:json:indent", "  ")
	})
	var _ = Service("service-name", func() {
		Files("path1", "filename")
		Files("path2", "filename", func() {
			Meta("openapi:tag:user-tag")
		})
	})
}

var JSONPrefixIndentDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
		Meta("openapi:json:prefix", " ")
		Meta("openapi:json:indent", "  ")
	})
	var _ = Service("service-name", func() {
		Files("path1", "filename")
		Files("path2", "filename", func() {
			Meta("openapi:tag:user-tag")
		})
	})
}

var AdditionalPropertiesTypeDSL = func() {
	var PayloadT = Type("Payload", func() {
		Attribute("string", String, func() {
			Example("")
		})
		Meta("openapi:additionalProperties", "false")
	})
	var ResultT = Type("Result", func() {
		Attribute("string", String, func() {
			Example("")
		})
		Meta("openapi:additionalProperties", "false")
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

var AdditionalPropertiesPayloadResultDSL = func() {
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
			Payload(PayloadT, func() {
				Meta("openapi:additionalProperties", "false")
			})
			Result(ResultT, func() {
				Meta("openapi:additionalProperties", "false")
			})
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var AdditionalPropertiesEmbeddedPayloadResultDSL = func() {
	var _ = API("test", func() {
		Server("test", func() {
			Host("localhost", func() {
				URI("https://goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(func() {
				Attribute("string", String, func() {
					Example("")
				})
				Meta("openapi:additionalProperties", "false")
			})
			Result(func() {
				Attribute("string", String, func() {
					Example("")
				})
				Meta("openapi:additionalProperties", "false")
			})
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var OpenAPIV32MetaDSL = func() {
	var _ = API("test", func() {
		Meta("openapi:info:summary", "Test API summary")
		Meta("openapi:tag:Users:desc", "Operations about users")
		Meta("openapi:tag:Users:summary", "Users")
		Meta("openapi:tag:Users:parent", "Accounts")
		Meta("openapi:tag:Users:kind", "nav")
		Server("multi", func() {
			Host("dev", func() {
				URI("http://localhost:8080")
			})
			Host("prod", func() {
				URI("https://goa.design")
			})
		})
		Server("single", func() {
			Host("only", func() {
				URI("https://single.goa.design")
			})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Empty)
			Result(Empty)
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var OpenAPIVersionsSubsetDSL = func() {
	var _ = API("test", func() {
		Meta("openapi:versions", "3.2")
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var OpenAPIPathOverrideDSL = func() {
	var _ = API("test", func() {
		Meta("openapi:path:3.2", "docs/openapi")
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var OpenAPIInvalidVersionDSL = func() {
	var _ = API("test", func() {
		Meta("openapi:versions", "4.0")
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var TypeExtensionDSL = func() {
	var Notification = Type("Notification", func() {
		Meta("openapi:extension:x-test-include", "true")
		Attribute("id", String, func() {
			Example("notice")
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Notification)
			Result(Notification)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var AliasTypeDSL = func() {
	var Stage = Type("Stage", String, func() {
		Description("Setup stage.")
		Enum("who", "when", "where", "what")
		Example("who")
	})
	var Setup = Type("Setup", func() {
		Example(Val{"current": "who", "completed": []any{"when"}})
		Attribute("current", Stage)
		Attribute("completed", ArrayOf(Stage), func() {
			Example([]any{"when"})
		})
	})
	Service("testService", func() {
		Method("testEndpoint", func() {
			Payload(Setup)
			Result(Setup)
			HTTP(func() {
				POST("/")
			})
		})
	})
}
