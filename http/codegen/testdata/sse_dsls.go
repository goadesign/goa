package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var SSEStringDSL = func() {
	Service("SSEStringService", func() {
		Method("SSEStringMethod", func() {
			StreamingResult(String, func() {
				Example("event")
			})
			HTTP(func() {
				GET("/string")
				ServerSentEvents()
			})
		})
	})
}

var SSEIntDSL = func() {
	Service("SSEIntService", func() {
		Method("SSEIntMethod", func() {
			StreamingResult(Int)
			HTTP(func() {
				GET("/int")
				ServerSentEvents()
			})
		})
	})
}

var SSEBoolDSL = func() {
	Service("SSEBoolService", func() {
		Method("SSEBoolMethod", func() {
			StreamingResult(Boolean)
			HTTP(func() {
				GET("/bool")
				ServerSentEvents()
			})
		})
	})
}

var SSEObjectDSL = func() {
	Service("SSEObjectService", func() {
		Method("SSEObjectMethod", func() {
			StreamingResult(func() {
				Attribute("id", String, func() {
					Example("event")
				})
				Attribute("value", Int, func() {
					Example(1)
				})
				Attribute("flag", Boolean, func() {
					Example(true)
				})
			})
			HTTP(func() {
				GET("/object")
				ServerSentEvents()
			})
		})
	})
}

var SSEDataFieldDSL = func() {
	Service("SSEDataFieldService", func() {
		Method("SSEDataFieldMethod", func() {
			StreamingResult(func() {
				Attribute("data", String, func() {
					Example("event")
				})
				Attribute("flag", Boolean, func() {
					Example(true)
				})
			})
			HTTP(func() {
				GET("/data-field")
				ServerSentEvents("data")
			})
		})
	})
}

var SSEDataIDFieldDSL = func() {
	Service("SSEDataIDFieldService", func() {
		Method("SSEDataIDFieldMethod", func() {
			StreamingResult(func() {
				Attribute("data", String)
				Attribute("id", String)
			})
			HTTP(func() {
				GET("/data-id-field")
				ServerSentEvents("data", func() {
					SSEEventID("id")
				})
			})
		})
	})
}

var SSERequestIDDSL = func() {
	Service("SSERequestIDService", func() {
		Method("SSERequestIDMethod", func() {
			Payload(func() {
				Attribute("id", String, func() {
					Example("request")
				})
			})
			StreamingResult(String, func() {
				Example("event")
			})
			HTTP(func() {
				GET("/request-id")
				ServerSentEvents(func() {
					SSERequestID("id")
				})
			})
		})
	})
}

var SSEAllFieldsDSL = func() {
	Service("SSEAllFieldsService", func() {
		Method("SSEAllFieldsMethod", func() {
			Payload(func() {
				Attribute("id", String, func() {
					Example("request")
				})
			})
			StreamingResult(func() {
				Attribute("id", String, func() {
					Example("123")
				})
				Attribute("event", String, func() {
					Example("update")
				})
				Attribute("retry", Int, func() {
					Example(3000)
				})
				Attribute("data", func() {
					Attribute("message", String, func() {
						Example("Hello, world!")
					})
				})
			})
			HTTP(func() {
				GET("/all-fields")
				ServerSentEvents(func() {
					SSERequestID("id")
					SSEEventID("id")
					SSEEventType("event")
					SSEEventRetry("retry")
					SSEEventData("data")
				})
			})
		})
	})
}
