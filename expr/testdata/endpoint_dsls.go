package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var ValidRouteDSL = func() {
	Service("ValidRoute", func() {
		HTTP(func() {
			Path("/{base_id}")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("base_id", String)
				Attribute("id", String)
			})
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var DuplicateWCRouteDSL = func() {
	Service("InvalidRoute", func() {
		HTTP(func() {
			Path("/{id}")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var DisallowResponseBodyHeadDSL = func() {
	Service("DisallowResponseBody", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("id", String)
			})
			Error("not_found")
			HTTP(func() {
				HEAD("/")
				Response("not_found", StatusNotFound)
			})
		})
	})
}

var EndpointWithParentDSL = func() {
	Service("Parent", func() {
		Method("show", func() {
			Payload(func() {
				Attribute("pparam", String)
				Attribute("pheader", String)
				Attribute("pcookie", String)
			})
			HTTP(func() {
				POST("/{pparam}")
				Header("pheader")
				Cookie("pcookie")
			})
		})
	})
	Service("Child", func() {
		HTTP(func() {
			Parent("Parent")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("param", String)
				Attribute("header", String)
				Attribute("pheader", String)
				Attribute("cookie", String)
				Attribute("pcookie", String)
			})
			HTTP(func() {
				POST("/{param}")
				Header("header")
				Cookie("cookie")
			})
		})
	})
}

var EndpointWithParentRevertDSL = func() {
	Service("Child", func() {
		HTTP(func() {
			Parent("Parent")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("param", String)
				Attribute("header", String)
				Attribute("pheader", String)
				Attribute("cookie", String)
				Attribute("pcookie", String)
			})
			HTTP(func() {
				POST("/{param}")
				Header("header")
				Cookie("cookie")
			})
		})
	})
	Service("Parent", func() {
		Method("show", func() {
			Payload(func() {
				Attribute("pparam", String)
				Attribute("pheader", String)
				Attribute("pcookie", String)
			})
			HTTP(func() {
				POST("/{pparam}")
				Header("pheader")
				Cookie("pcookie")
			})
		})
	})
}

var EndpointRecursiveParentDSL = func() {
	Service("Parent", func() {
		HTTP(func() {
			Parent("Child")
		})
		Method("show", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
	Service("Child", func() {
		HTTP(func() {
			Parent("Parent")
		})
		Method("show", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var EndpointBodyAsPayloadProp = func() {
	Service("Service", func() {
		Method("Method", func() {
			Payload(func() {
				Attribute("id", String)
				Attribute("name", String)
			})
			HTTP(func() {
				POST("/{id}")
				Body("name")
			})
		})
	})
}

var EndpointBodyAsMissedPayloadProp = func() {
	Service("Service", func() {
		Method("Method", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				POST("/{id}")
				Body("name")
			})
		})
	})
}

var EndpointBodyExtendPayload = func() {
	Service("Service", func() {
		Method("Method", func() {
			Payload(func() {
				Attribute("id", String)
				Attribute("name", String)
			})
			HTTP(func() {
				POST("/{id}")
				Body(func() {
					Attribute("name")
				})
			})
		})
	})
}

var EndpointBodyAsUserType = func() {
	var Entity = Type("Entity", func() {
		Attribute("id", String)
		Attribute("name", String)
	})

	var EntityData = Type("EntityData", func() {
		Attribute("name", String)
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(Entity)
			HTTP(func() {
				POST("/{id}")
				Body(EntityData)
			})
		})
	})
}

var EndpointMissingToken = func() {
	var Entity = Type("Entity", func() {
		Attribute("id", String)
		Attribute("name", String)
	})
	var JWT = JWTSecurity("JWT", func() {
		Scope("api:read", "Read access")
	})
	Service("Service", func() {
		Security(JWT, func() {
			Scope("api:read")
		})
		Method("Method", func() {
			Payload(Entity)
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var EndpointMissingTokenPayload = func() {
	var JWT = JWTSecurity("JWT", func() {
		Scope("api:read", "Read access")
	})
	Service("Service", func() {
		Security(JWT, func() {
			Scope("api:read")
		})
		Method("Method", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var EndpointExtendToken = func() {
	var CommonAttributes = Type("Common", func() {
		Token("token", String)
	})
	var Entity = Type("Entity", func() {
		Extend(CommonAttributes)
		Attribute("id", String)
		Attribute("name", String)
	})
	var JWT = JWTSecurity("JWT", func() {
		Scope("api:read", "Read access")
	})
	Service("Service", func() {
		Security(JWT, func() {
			Scope("api:read")
		})
		Method("Method", func() {
			Payload(Entity)
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var EndpointHasParent = func() {
	Service("Ancestor", func() {
		HTTP(func() {
			Path("/ancestor")
			CanonicalMethod("Method")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("ancestor_id", Int)
				Attribute("query_0", String)
				Required("ancestor_id")
			})
			HTTP(func() {
				GET("/{ancestor_id}")
				Param("query_0")
			})
		})

	})
	Service("Parent", func() {
		HTTP(func() {
			Path("/parents")
			CanonicalMethod("Method")
			Parent("Ancestor")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("parent_id", Int)
				Attribute("query_1", String)
				Required("parent_id")
			})
			HTTP(func() {
				GET("/{parent_id}")
				Param("query_1")
			})
		})
	})
	Service("Child", func() {
		HTTP(func() {
			Path("/children")
			Parent("Parent")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("child_id", Int)
			})
			HTTP(func() {
				GET("/{child_id}")
			})
		})
	})
}

var EndpointHasParentAndOther = func() {
	Service("Parent", func() {
		HTTP(func() {
			Path("/parents")
			CanonicalMethod("Method")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("parent_id", Int)
				Attribute("query_1", String)
			})
			HTTP(func() {
				GET("/{parent_id}")
				Param("query_1")
			})
		})
	})
	Service("Child", func() {
		HTTP(func() {
			Path("/children")
			Parent("Parent")
		})
		Method("Method", func() {
			HTTP(func() {
				GET("")
			})
		})
	})
	Service("Other", func() {
		HTTP(func() {
			Path("/others")
		})
		Method("Method", func() {
			HTTP(func() {
				GET("")
			})
		})
	})

}

var EndpointHasSkipRequestEncodeAndPayloadStreaming = func() {
	Service("Service", func() {
		Method("Method", func() {
			StreamingPayload(String)
			HTTP(func() {
				GET("/")
				SkipRequestBodyEncodeDecode()
			})
		})
	})
}

var EndpointHasSkipRequestEncodeAndResultStreaming = func() {
	Service("Service", func() {
		Method("Method", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				SkipRequestBodyEncodeDecode()
			})
		})
	})
}

var EndpointHasSkipResponseEncodeAndPayloadStreaming = func() {
	Service("Service", func() {
		Method("Method", func() {
			StreamingPayload(String)
			HTTP(func() {
				GET("/")
				SkipResponseBodyEncodeDecode()
			})
		})
	})
}

var EndpointHasSkipResponseEncodeAndResultStreaming = func() {
	Service("Service", func() {
		Method("Method", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				SkipResponseBodyEncodeDecode()
			})
		})
	})
}

var EndpointHasSkipEncodeAndGRPC = func() {
	Service("Service", func() {
		Method("Method", func() {
			Payload(func() {
				Field(1, "param", Int)
				Field(2, "query", String)
			})
			HTTP(func() {
				GET("/{param}")
				Param("query")
				SkipRequestBodyEncodeDecode()
			})
			GRPC(func() {})
		})
	})
}

var EndpointPayloadMissingRequired = func() {
	Service("Service", func() {
		Method("Method", func() {
			Payload(func() {
				Attribute("nonreq")
			})
			HTTP(func() {
				POST("/")
				Body(func() {
					Attribute("nonreq")
					Required("nonreq")
				})
			})
		})
	})
}

var StreamingEndpointRequestBody = func() {
	var PT = Type("Payload", func() {
		Attribute("foo", String)
	})
	Service("Service", func() {
		Method("MethodA", func() {
			Payload(func() {
				Attribute("bar", String)
			})
			StreamingResult(String)
			HTTP(func() {
				GET("/")
			})
		})
		Method("MethodB", func() {
			Payload(func() {
				Extend(PT)
			})
			StreamingResult(String)
			HTTP(func() {
				GET("/")
			})
		})
		Method("MethodC", func() {
			Payload(String)
			StreamingResult(String)
			HTTP(func() {
				GET("/")
			})
		})
		Method("MethodD", func() {
			Payload(func() {
				Attribute("bar", String)
			})
			StreamingResult(String)
			HTTP(func() {
				GET("/{bar}")
			})
		})
		Method("MethodE", func() {
			Payload(func() {
				Extend(PT)
			})
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				Param("foo")
			})
		})
		Method("MethodF", func() {
			Payload(String)
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				Header("foo")
			})
		})
	})
}

var FinalizeEndpointBodyAsExtendedTypeDSL = func() {
	var EntityData = Type("EntityData", func() {
		Attribute("name", String)
	})

	var Entity = Type("Entity", func() {
		Attribute("id", String)
		Extend(EntityData)
		Required("id")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(Entity)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var FinalizeEndpointBodyAsPropWithExtendedTypeDSL = func() {
	var OAuth2 = OAuth2Security("authCode")

	var EntityData = Type("EntityData", func() {
		Attribute("name", String)
	})

	var Entity = Type("Entity", func() {
		Attribute("id", String)
		Extend(EntityData)
		Required("id")
	})

	Service("Service", func() {
		Method("Method", func() {
			Security(OAuth2)
			Payload(func() {
				AccessToken("token", String)
				Attribute("payload", Entity)
			})
			HTTP(func() {
				POST("/")
				Body("payload")
			})
		})
	})
}

var ExplicitAuthHeaderDSL = func() {
	var OAuth2 = OAuth2Security("authCode")
	Service("Service", func() {
		Method("Method", func() {
			Security(OAuth2)
			Payload(func() {
				AccessToken("token", String)
				Attribute("payload", String)
			})
			HTTP(func() {
				POST("/")
				Header("token")
			})
		})
	})
}

var ImplicitAuthHeaderDSL = func() {
	var OAuth2 = OAuth2Security("authCode")
	Service("Service", func() {
		Method("Method", func() {
			Security(OAuth2)
			Payload(func() {
				AccessToken("token", String)
				Attribute("payload", String)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var GRPCEndpointWithAnyType = func() {
	var Recursive = Type("Recursive", func() {
		Field(1, "invalid_map_key", MapOf(Any, "Recursive"))
		Field(3, "invalid_array", ArrayOf(ArrayOf(Any)))
	})
	var InvalidRT = ResultType("application/vnd.result", func() {
		TypeName("RT")
		Attributes(func() {
			Field(1, "invalid_primitive", Any)
			Field(2, "invalid_array", ArrayOf(Any))
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(Recursive)
			Result(CollectionOf(InvalidRT))
			Error("invalid_error_type", Any)
			Error("invalid_map_type", MapOf(Int, Any))
			GRPC(func() {})
		})
	})
}

var GRPCEndpointWithUntaggedFields = func() {
	var Req = Type("Req", func() {
		Attribute("req_not_field", String)
	})
	var Resp = Type("Resp", func() {
		Attribute("resp_not_field", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(Req)
			Result(Resp)
			GRPC(func() {})
		})
	})
}

var GRPCEndpointWithRepeatedFieldTags = func() {
	var Req = Type("Req", func() {
		Field(1, "key", String)
		Field(1, "key_dup_id", String)
	})
	var Resp = Type("Resp", func() {
		Field(2, "key", String)
		Field(2, "key_dup_id", String)
	})
	Service("Service", func() {
		Method("Method", func() {
			Payload(Req)
			Result(Resp)
			GRPC(func() {})
		})
	})
}

var GRPCEndpointWithReferenceTypes = func() {
	var EntityReference = Type("EntityReference", func() {
		Field(1, "name", String)
	})

	var Entity = Type("Entity", func() {
		Reference(EntityReference)
		Field(1, "id", String)
		Field(2, "name")
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(Entity)
			GRPC(func() {})
		})
	})
}

var GRPCEndpointWithExtendedTypes = func() {
	var EntityExtended = Type("EntityExtended", func() {
		Field(1, "name", String)
		Field(2, "id", String)
	})

	var Entity = Type("Entity", func() {
		Extend(EntityExtended)
	})

	Service("Service", func() {
		Method("Method", func() {
			Payload(Entity)
			Result(func() {
				Extend(Entity)
			})
			GRPC(func() {
				Metadata(func() {
					Attribute("name")
				})
				Message(func() {
					Attribute("id")
				})
				Response(func() {
					Headers(func() {
						Attribute("name")
					})
					Message(func() {
						Attribute("id")
					})
				})
			})
		})
	})
}
