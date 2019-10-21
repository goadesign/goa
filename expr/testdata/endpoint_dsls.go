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
			Payload(func() {
				Attribute("child_id", Int)
			})
			HTTP(func() {
				GET("/{child_id}")
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

