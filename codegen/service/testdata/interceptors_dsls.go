package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var NoInterceptorsDSL = func() {
	Service("NoInterceptors", func() {
		Method("Method", func() {
			HTTP(func() { GET("/") })
		})
	})
}

var SingleAPIServerInterceptorDSL = func() {
	Interceptor("logging")
	API("SingleAPIServerInterceptor", func() {
		ServerInterceptor("logging")
	})
	Service("SingleAPIServerInterceptor", func() {
		Method("Method", func() {
			HTTP(func() { GET("/1") })
		})
		Method("Method2", func() {
			HTTP(func() { GET("/2") })
		})
	})
}

var SingleServiceServerInterceptorDSL = func() {
	Interceptor("logging")
	Service("SingleServerInterceptor", func() {
		ServerInterceptor("logging")
		Method("Method", func() {
			HTTP(func() {
				GET("/1")
			})
		})
		Method("Method2", func() {
			HTTP(func() {
				GET("/2")
			})
		})
	})
}

var SingleMethodServerInterceptorDSL = func() {
	Interceptor("logging")
	Service("SingleMethodServerInterceptor", func() {
		Method("Method", func() {
			ServerInterceptor("logging")
			HTTP(func() { GET("/1") })
		})
		Method("Method2", func() {
			HTTP(func() { GET("/2") })
		})
	})
}

var SingleClientInterceptorDSL = func() {
	Interceptor("tracing")
	Service("SingleClientInterceptor", func() {
		ClientInterceptor("tracing")
		Method("Method", func() {
			Payload(func() {
				Attribute("id", Int)
			})
			Result(func() {
				Attribute("value", String)
			})
			HTTP(func() { GET("/") })
		})
	})
}

var MultipleInterceptorsDSL = func() {
	Interceptor("logging")
	Interceptor("tracing")
	Interceptor("metrics")
	Service("MultipleInterceptors", func() {
		ServerInterceptor("logging")
		ServerInterceptor("tracing")
		ClientInterceptor("metrics")
		Method("Method", func() {
			Payload(func() {
				Attribute("query", String)
			})
			Result(func() {
				Attribute("data", String)
			})
			HTTP(func() { GET("/") })
		})
	})
}

var InterceptorWithReadPayloadDSL = func() {
	Interceptor("validation", func() {
		ReadPayload(func() {
			Attribute("name")
		})
	})
	Service("InterceptorWithReadPayload", func() {
		ServerInterceptor("validation")
		ClientInterceptor("validation")
		Method("Method", func() {
			Payload(func() {
				Attribute("name", String)
				Required("name")
			})
			HTTP(func() { POST("/") })
		})
	})
}

var InterceptorWithWritePayloadDSL = func() {
	Interceptor("validation", func() {
		WritePayload(func() {
			Attribute("name")
		})
	})
	Service("InterceptorWithWritePayload", func() {
		ServerInterceptor("validation")
		ClientInterceptor("validation")
		Method("Method", func() {
			Payload(func() {
				Attribute("name", String)
				Required("name")
			})
			HTTP(func() { POST("/") })
		})
	})
}

var InterceptorWithReadWritePayloadDSL = func() {
	Interceptor("validation", func() {
		ReadPayload(func() {
			Attribute("name")
		})
		WritePayload(func() {
			Attribute("name")
		})
	})
	Service("InterceptorWithReadWritePayload", func() {
		ServerInterceptor("validation")
		ClientInterceptor("validation")
		Method("Method", func() {
			Payload(func() {
				Attribute("name", String)
				Required("name")
			})
			HTTP(func() { POST("/") })
		})
	})
}

var InterceptorWithReadResultDSL = func() {
	Interceptor("caching", func() {
		ReadResult(func() {
			Attribute("data")
		})
	})
	Service("InterceptorWithReadResult", func() {
		ServerInterceptor("caching")
		ClientInterceptor("caching")
		Method("Method", func() {
			Result(func() {
				Attribute("data", String)
				Required("data")
			})
			HTTP(func() { GET("/") })
		})
	})
}

var InterceptorWithWriteResultDSL = func() {
	Interceptor("caching", func() {
		WriteResult(func() {
			Attribute("data")
		})
	})
	Service("InterceptorWithWriteResult", func() {
		ServerInterceptor("caching")
		ClientInterceptor("caching")
		Method("Method", func() {
			Result(func() {
				Attribute("data", String)
				Required("data")
			})
			HTTP(func() { GET("/") })
		})
	})
}

var InterceptorWithReadWriteResultDSL = func() {
	Interceptor("caching", func() {
		ReadResult(func() {
			Attribute("data")
		})
		WriteResult(func() {
			Attribute("data")
		})
	})
	Service("InterceptorWithReadWriteResult", func() {
		ServerInterceptor("caching")
		ClientInterceptor("caching")
		Method("Method", func() {
			Result(func() {
				Attribute("data", String)
				Required("data")
			})
			HTTP(func() { GET("/") })
		})
	})
}

var StreamingInterceptorsDSL = func() {
	Interceptor("logging")
	Service("StreamingInterceptors", func() {
		ServerInterceptor("logging")
		Method("Method", func() {
			StreamingPayload(func() {
				Attribute("chunk", String)
			})
			StreamingResult(func() {
				Attribute("data", String)
			})
			HTTP(func() { GET("/stream") })
		})
	})
}

var StreamingInterceptorsWithReadPayloadDSL = func() {
	Interceptor("logging", func() {
		ReadPayload(func() {
			Attribute("initial")
		})
	})
	Service("StreamingInterceptorsWithReadPayload", func() {
		ServerInterceptor("logging")
		Method("Method", func() {
			Payload(func() {
				Attribute("initial", String)
			})
			StreamingPayload(func() {
				Attribute("chunk", String)
			})
			HTTP(func() {
				Header("initial")
				GET("/stream")
			})
		})
	})
}

var StreamingInterceptorsWithReadResultDSL = func() {
	Interceptor("logging", func() {
		ReadResult(func() {
			Attribute("data")
		})
	})
	Service("StreamingInterceptorsWithReadResult", func() {
		ServerInterceptor("logging")
		Method("Method", func() {
			Payload(func() {
				Attribute("initial", String)
			})
			StreamingPayload(func() {
				Attribute("chunk", String)
			})
			Result(func() {
				Attribute("data", String)
			})
			HTTP(func() {
				Header("initial")
				GET("/stream")
			})
		})
	})
}

// Invalid DSL
var StreamingResultInterceptorDSL = func() {
	Interceptor("logging", func() {
		ReadResult(func() {
			Attribute("data")
		})
	})
	Service("StreamingResultInterceptor", func() {
		ServerInterceptor("logging")
		Method("Method", func() {
			StreamingResult(func() {
				Attribute("data", String)
			})
			HTTP(func() { GET("/stream") })
		})
	})
}
