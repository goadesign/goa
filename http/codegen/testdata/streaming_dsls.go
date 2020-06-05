package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var MixedEndpointsDSL = func() {
	Service("MixedEndpoints", func() {
		Method("NonStreaming", func() {
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
		Method("Streaming", func() {
			StreamingResult(Int)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingMultipleServicesDSL = func() {
	Service("StreamingServiceA", func() {
		Method("Method", func() {
			StreamingResult(Int)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
	Service("StreamingServiceB", func() {
		Method("Method", func() {
			StreamingPayload(Int)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var Result = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingResultService", func() {
		Method("StreamingResultMethod", func() {
			Payload(Request)
			StreamingResult(Result)
			HTTP(func() {
				GET("/{x}")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultWithViewsDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var Result = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingResultWithViewsService", func() {
		Method("StreamingResultWithViewsMethod", func() {
			Payload(Request)
			StreamingResult(Result)
			HTTP(func() {
				GET("/{x}")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultWithExplicitViewDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var Result = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingResultWithExplicitViewService", func() {
		Method("StreamingResultWithExplicitViewMethod", func() {
			Payload(Request)
			StreamingResult(Result, func() {
				View("extended")
			})
			HTTP(func() {
				GET("/{x}")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultCollectionWithViewsDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var Result = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingResultCollectionWithViewsService", func() {
		Method("StreamingResultCollectionWithViewsMethod", func() {
			Payload(Request)
			StreamingResult(CollectionOf(Result))
			HTTP(func() {
				GET("/{x}")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultCollectionWithExplicitViewDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var Result = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingResultCollectionWithExplicitViewService", func() {
		Method("StreamingResultCollectionWithExplicitViewMethod", func() {
			Payload(Request)
			StreamingResult(CollectionOf(Result), func() {
				View("tiny")
			})
			HTTP(func() {
				GET("/{x}")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultNoPayloadDSL = func() {
	var Result = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingResultNoPayloadService", func() {
		Method("StreamingResultNoPayloadMethod", func() {
			StreamingResult(Result)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultPrimitiveDSL = func() {
	Service("StreamingResultPrimitiveService", func() {
		Method("StreamingResultPrimitiveMethod", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultPrimitiveArrayDSL = func() {
	Service("StreamingResultPrimitiveArrayService", func() {
		Method("StreamingResultPrimitiveArrayMethod", func() {
			StreamingResult(ArrayOf(Int32))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultPrimitiveMapDSL = func() {
	Service("StreamingResultPrimitiveMapService", func() {
		Method("StreamingResultPrimitiveMapMethod", func() {
			StreamingResult(MapOf(Int32, String))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultUserTypeArrayDSL = func() {
	var Result = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingResultUserTypeArrayService", func() {
		Method("StreamingResultUserTypeArrayMethod", func() {
			StreamingResult(ArrayOf(Result))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingResultUserTypeMapDSL = func() {
	var Result = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingResultUserTypeMapService", func() {
		Method("StreamingResultUserTypeMapMethod", func() {
			StreamingResult(MapOf(String, Result))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var PayloadType = Type("Payload", func() {
		Attribute("p", String)
		Attribute("q", String)
		Attribute("r", String)
	})
	var ResultType = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingPayloadService", func() {
		Method("StreamingPayloadMethod", func() {
			Payload(PayloadType)
			StreamingPayload(Request)
			Result(ResultType)
			HTTP(func() {
				GET("/{p}")
				Param("q")
				Header("r:Location")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadNoPayloadDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var ResultType = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("StreamingPayloadNoPayloadService", func() {
		Method("StreamingPayloadNoPayloadMethod", func() {
			StreamingPayload(Request)
			Result(ResultType)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadPrimitiveDSL = func() {
	Service("StreamingPayloadPrimitiveService", func() {
		Method("StreamingPayloadPrimitiveMethod", func() {
			StreamingPayload(String)
			Result(String)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadPrimitiveArrayDSL = func() {
	Service("StreamingPayloadPrimitiveArrayService", func() {
		Method("StreamingPayloadPrimitiveArrayMethod", func() {
			StreamingPayload(ArrayOf(Int32))
			Result(ArrayOf(String))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadPrimitiveMapDSL = func() {
	Service("StreamingPayloadPrimitiveMapService", func() {
		Method("StreamingPayloadPrimitiveMapMethod", func() {
			StreamingPayload(MapOf(String, Int32))
			Result(MapOf(Int, Int))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadUserTypeArrayDSL = func() {
	var RequestType = Type("RequestType", func() {
		Attribute("a", String)
	})
	Service("StreamingPayloadUserTypeArrayService", func() {
		Method("StreamingPayloadUserTypeArrayMethod", func() {
			StreamingPayload(ArrayOf(RequestType))
			Result(String)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadUserTypeMapDSL = func() {
	var RequestType = Type("RequestType", func() {
		Attribute("a", String)
	})
	Service("StreamingPayloadUserTypeMapService", func() {
		Method("StreamingPayloadUserTypeMapMethod", func() {
			StreamingPayload(MapOf(String, RequestType))
			Result(ArrayOf(String))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadNoResultDSL = func() {
	Service("StreamingPayloadNoResultService", func() {
		Method("StreamingPayloadNoResultMethod", func() {
			StreamingPayload(String)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadResultWithViewsDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingPayloadResultWithViewsService", func() {
		Method("StreamingPayloadResultWithViewsMethod", func() {
			StreamingPayload(Float32)
			Result(ResultT)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadResultWithExplicitViewDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingPayloadResultWithExplicitViewService", func() {
		Method("StreamingPayloadResultWithExplicitViewMethod", func() {
			StreamingPayload(Float32)
			Result(ResultT, func() {
				View("extended")
			})
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadResultCollectionWithViewsDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingPayloadResultCollectionWithViewsService", func() {
		Method("StreamingPayloadResultCollectionWithViewsMethod", func() {
			StreamingPayload(Any)
			Result(CollectionOf(ResultT))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var StreamingPayloadResultCollectionWithExplicitViewDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("StreamingPayloadResultCollectionWithExplicitViewService", func() {
		Method("StreamingPayloadResultCollectionWithExplicitViewMethod", func() {
			StreamingPayload(Any)
			Result(CollectionOf(ResultT), func() {
				View("tiny")
			})
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var PayloadType = Type("Payload", func() {
		Attribute("p", String)
		Attribute("q", String)
		Attribute("r", String)
	})
	var ResultType = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("BidirectionalStreamingService", func() {
		Method("BidirectionalStreamingMethod", func() {
			Payload(PayloadType)
			StreamingPayload(Request)
			StreamingResult(ResultType)
			HTTP(func() {
				GET("/{p}")
				Param("q")
				Header("r:Location")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingNoPayloadDSL = func() {
	var Request = Type("Request", func() {
		Attribute("x", String)
	})
	var ResultType = Type("UserType", func() {
		Attribute("a", String)
	})
	Service("BidirectionalStreamingNoPayloadService", func() {
		Method("BidirectionalStreamingNoPayloadMethod", func() {
			StreamingPayload(Request)
			StreamingResult(ResultType)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingPrimitiveDSL = func() {
	Service("BidirectionalStreamingPrimitiveService", func() {
		Method("BidirectionalStreamingPrimitiveMethod", func() {
			StreamingPayload(String)
			StreamingResult(String)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingPrimitiveArrayDSL = func() {
	Service("BidirectionalStreamingPrimitiveArrayService", func() {
		Method("BidirectionalStreamingPrimitiveArrayMethod", func() {
			StreamingPayload(ArrayOf(Int32))
			StreamingResult(ArrayOf(String))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingPrimitiveMapDSL = func() {
	Service("BidirectionalStreamingPrimitiveMapService", func() {
		Method("BidirectionalStreamingPrimitiveMapMethod", func() {
			StreamingPayload(MapOf(String, Int32))
			StreamingResult(MapOf(Int, Int))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingUserTypeArrayDSL = func() {
	var RequestType = Type("RequestType", func() {
		Attribute("a", String)
	})
	var ResultT = Type("ResultType", func() {
		Attribute("b", String)
	})
	Service("BidirectionalStreamingUserTypeArrayService", func() {
		Method("BidirectionalStreamingUserTypeArrayMethod", func() {
			StreamingPayload(ArrayOf(RequestType))
			StreamingResult(ArrayOf(ResultT))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingUserTypeMapDSL = func() {
	var RequestType = Type("RequestType", func() {
		Attribute("a", String)
	})
	var ResultT = Type("ResultType", func() {
		Attribute("b", String)
	})
	Service("BidirectionalStreamingUserTypeMapService", func() {
		Method("BidirectionalStreamingUserTypeMapMethod", func() {
			StreamingPayload(MapOf(String, RequestType))
			StreamingResult(MapOf(String, ResultT))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingResultWithViewsDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("BidirectionalStreamingResultWithViewsService", func() {
		Method("BidirectionalStreamingResultWithViewsMethod", func() {
			StreamingPayload(Float32)
			StreamingResult(ResultT)
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingResultWithExplicitViewDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("BidirectionalStreamingResultWithExplicitViewService", func() {
		Method("BidirectionalStreamingResultWithExplicitViewMethod", func() {
			StreamingPayload(Float32)
			StreamingResult(ResultT, func() {
				View("extended")
			})
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingResultCollectionWithViewsDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("BidirectionalStreamingResultCollectionWithViewsService", func() {
		Method("BidirectionalStreamingResultCollectionWithViewsMethod", func() {
			StreamingPayload(Any)
			StreamingResult(CollectionOf(ResultT))
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}

var BidirectionalStreamingResultCollectionWithExplicitViewDSL = func() {
	var ResultT = ResultType("UserType", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
			Attribute("c", String)
		})
		View("tiny", func() {
			Attribute("a", String)
		})
		View("extended", func() {
			Attribute("a")
			Attribute("b")
			Attribute("c")
		})
	})
	Service("BidirectionalStreamingResultCollectionWithExplicitViewService", func() {
		Method("BidirectionalStreamingResultCollectionWithExplicitViewMethod", func() {
			StreamingPayload(Any)
			StreamingResult(CollectionOf(ResultT), func() {
				View("tiny")
			})
			HTTP(func() {
				GET("/")
				Response(StatusOK)
			})
		})
	})
}
