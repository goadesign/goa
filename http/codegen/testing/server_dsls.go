package testing

import (
	. "goa.design/goa/http/design"
	. "goa.design/goa/http/dsl"
)

var ServerNoPayloadNoResultDSL = func() {
	Service("ServiceNoPayloadNoResult", func() {
		Method("MethodNoPayloadNoResult", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ServerPayloadNoResultDSL = func() {
	Service("ServicePayloadNoResult", func() {
		Method("MethodPayloadNoResult", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ServerNoPayloadResultDSL = func() {
	Service("ServiceNoPayloadResult", func() {
		Method("MethodNoPayloadResult", func() {
			Result(func() {
				Attribute("b", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ServerPayloadResultDSL = func() {
	Service("ServicePayloadResult", func() {
		Method("MethodPayloadResult", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			Result(func() {
				Attribute("b", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ServerPayloadResultErrorDSL = func() {
	Service("ServicePayloadResultError", func() {
		Method("MethodPayloadResultError", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			Result(func() {
				Attribute("b", Boolean)
			})
			Error("e", func() {
				Attribute("c", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
				Response("e", func() {
					Code(StatusConflict)
				})
			})
		})
	})
}
