package testdata

import (
	. "goa.design/goa/dsl"
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
			})
			HTTP(func() {
				POST("/{id}")
				Body(func() {
					Attribute("name", String)
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
