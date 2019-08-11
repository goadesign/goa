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
