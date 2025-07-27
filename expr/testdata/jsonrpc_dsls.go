package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// Valid JSON-RPC DSL scenarios

var ValidJSONRPCBasicDSL = func() {
	Service("calc", func() {
		JSONRPC(func() {
			POST("/rpc")
		})
		Method("add", func() {
			Payload(func() {
				Attribute("a", Int)
				Attribute("b", Int)
			})
			Result(Int)
			JSONRPC(func() {})
		})
	})
}

var JSONRPCWithErrorMappingDSL = func() {
	var ErrorResult = Type("ErrorResult", func() {
		Attribute("message", String)
		Required("message")
	})

	API("test", func() {
		Error("unauthorized", ErrorResult)
		JSONRPC(func() {
			Response("unauthorized", RPCInvalidRequest)
		})
	})
	Service("calc", func() {
		JSONRPC(func() {
			POST("/rpc")
		})
		Error("div_zero", ErrorResult)
		Method("divide", func() {
			Payload(func() {
				Attribute("dividend", Int)
				Attribute("divisor", Int)
			})
			Result(Float64)
			Error("div_zero")
			JSONRPC(func() {
				Response("div_zero", RPCInvalidParams)
			})
		})
	})
}

var JSONRPCWithIDMappingDSL = func() {
	Service("calc", func() {
		JSONRPC(func() {
			POST("/rpc")
		})
		Method("compute", func() {
			Payload(func() {
				ID("request_id", String)
				Attribute("expression", String)
			})
			Result(Float64)
			JSONRPC(func() {})
		})
	})
}

var JSONRPCWithSSEDSL = func() {
	Service("ticker", func() {
		JSONRPC(func() {
			POST("/rpc")
			ServerSentEvents()
		})
		Method("stream", func() {
			Payload(func() {
				ID("client_id", String)
				Attribute("last_event_id", String)
			})
			StreamingResult(func() {
				Attribute("event_id", String)
				Attribute("price", Float64)
			})
			JSONRPC(func() {
				ServerSentEvents(func() {
					SSERequestID("last_event_id")
					SSEEventID("event_id")
				})
			})
		})
	})
}

var JSONRPCWithHeadersAndCookiesDSL = func() {
	Service("auth", func() {
		JSONRPC(func() {
			POST("/rpc")
			Headers(func() {
				Header("X-API-Version", String)
				Required("X-API-Version")
			})
			Cookie("session", String)
		})
		Method("secure", func() {
			Payload(func() {
				Attribute("data", String)
			})
			Result(String)
			JSONRPC(func() {
				Headers(func() {
					Header("Authorization", String)
					Required("Authorization")
				})
			})
		})
	})
}

var JSONRPCNotificationDSL = func() {
	Service("events", func() {
		JSONRPC(func() {
			POST("/rpc")
		})
		Method("notify", func() {
			Payload(func() {
				Attribute("event", String)
				Attribute("data", Any)
			})
			// No Result() - automatically a notification
			JSONRPC(func() {})
		})
	})
}

var JSONRPCMultipleServicesDSL = func() {
	Service("calc", func() {
		JSONRPC(func() {
			POST("/calc-rpc")
		})
		Method("add", func() {
			Payload(func() {
				Attribute("a", Int)
				Attribute("b", Int)
			})
			Result(Int)
			JSONRPC(func() {})
		})
	})
	Service("ticker", func() {
		JSONRPC(func() {
			POST("/ticker-rpc")
		})
		Method("price", func() {
			Payload(func() {
				Attribute("symbol", String)
			})
			Result(Float64)
			JSONRPC(func() {})
		})
	})
}

// Invalid JSON-RPC DSL scenarios

var JSONRPCBasicMissingServiceDSL = func() {
	Service("calc", func() {
		Method("add", func() {
			Payload(func() {
				Attribute("a", Int)
				Attribute("b", Int)
			})
			Result(Int)
			JSONRPC(func() {})
		})
	})
}

var JSONRPCInvalidContextDSL = func() {
	Type("MyType", func() {
		JSONRPC(func() {}) // Invalid - JSONRPC can't be used in Type
	})
}

var JSONRPCNonExistentErrorDSL = func() {
	Service("calc", func() {
		JSONRPC(func() {
			Response("unknown_error", RPCInternalError) // Error not defined
		})
	})
}

var JSONRPCInvalidIDAttributeDSL = func() {
	Service("calc", func() {
		Method("compute", func() {
			Payload(func() {
				Attribute("data", String)
				ID("request_id", Int)
			})
			Result(Int)
			JSONRPC(func() {})
		})
	})
}

var JSONRPCNonPOSTRouteDSL = func() {
	Service("calc", func() {
		JSONRPC(func() {
			GET("/rpc") // JSON-RPC must use POST
		})
		Method("add", func() {
			Result(Int)
			JSONRPC(func() {})
		})
	})
}

var JSONRPCMixedStreamingDSL = func() {
	Service("mixed", func() {
		Method("stream1", func() {
			StreamingResult(String)
			JSONRPC(func() {
				ServerSentEvents()
			})
		})
		Method("stream2", func() {
			StreamingResult(String)
			JSONRPC(func() {
				// No SSE - defaults to WebSocket
			})
		})
	})
}

var JSONRPCSSEOnNonStreamingDSL = func() {
	Service("calc", func() {
		Method("regular", func() {
			Result(String)
			JSONRPC(func() {
				ServerSentEvents()
			})
		})
	})
}

var JSONRPCSSEOnBidirectionalDSL = func() {
	Service("chat", func() {
		Method("connect", func() {
			StreamingPayload(String)
			StreamingResult(String)
			JSONRPC(func() {
				ServerSentEvents()
			})
		})
	})
}

// Complex inheritance scenarios

var JSONRPCErrorInheritanceDSL = func() {
	var ErrorResult = Type("ErrorResult", func() {
		Attribute("message", String)
		Required("message")
	})

	API("test", func() {
		Error("api_error", ErrorResult)
		JSONRPC(func() {
			Response("api_error", RPCInternalError)
		})
	})
	Service("calc", func() {
		Error("service_error", ErrorResult)
		JSONRPC(func() {
			Response("service_error", RPCInvalidRequest)
		})
		Method("compute", func() {
			Result(Int)
			Error("api_error")     // Use API-level error
			Error("service_error") // Use service-level error
			Error("method_error", ErrorResult)
			JSONRPC(func() {
				Response("method_error", RPCInvalidParams)
			})
		})
	})
}

var JSONRPCSSEInheritanceDSL = func() {
	API("test", func() {
		JSONRPC(func() {
			ServerSentEvents()
		})
	})
	Service("ticker", func() {
		Method("stream1", func() {
			StreamingResult(Any)
			JSONRPC(func() {}) // Should inherit SSE from API
		})
	})
	Service("chat", func() {
		// Service-level SSE configuration takes precedence over API level
		JSONRPC(func() {
			ServerSentEvents(func() {
				SSEEventID("custom_id") // Different SSE config than API
			})
		})
		Method("stream2", func() {
			StreamingResult(func() {
				Attribute("custom_id", String)
				Attribute("data", Any)
			})
			JSONRPC(func() {}) // Should inherit SSE from service (not API)
		})
	})
}

var JSONRPCHeadersCookiesInheritanceDSL = func() {
	Service("api", func() {
		JSONRPC(func() {
			Headers(func() {
				Header("X-Service-Version", String)
			})
			Cookie("service_session", String)
		})
		Method("method1", func() {
			Result(String)
			JSONRPC(func() {}) // Should inherit headers and cookies
		})
		Method("method2", func() {
			Result(String)
			JSONRPC(func() {
				Headers(func() {
					Header("X-Method-Header", String)
				})
				Cookie("method_cookie", String)
			})
		})
	})
}
