package testdata

import (
	. "goa.design/goa/dsl"
)

var UnaryRPCsDSL = func() {
	var PayloadA = Type("PayloadA", func() {
		Field(1, "Int", Int)
		Field(2, "String", String)
	})
	var PayloadB = Type("PayloadB", func() {
		Field(1, "UInt", UInt)
		Field(2, "Float32", Float32)
	})
	var ResultT = ResultType("application/vnd.goa.resultt", func() {
		TypeName("ResultT")
		Attributes(func() {
			Field(1, "ArrayField", ArrayOf(Boolean))
			Field(2, "MapField", MapOf(String, Float64))
		})
	})
	Service("ServiceUnaryRPCs", func() {
		Method("MethodUnaryRPCA", func() {
			Payload(PayloadA)
			Result(ResultT)
			GRPC(func() {})
		})
		Method("MethodUnaryRPCB", func() {
			Payload(PayloadB)
			Result(ResultT)
			GRPC(func() {})
		})
	})
}

var UnaryRPCNoPayloadDSL = func() {
	Service("ServiceUnaryRPCNoPayload", func() {
		Method("MethodUnaryRPCNoPayload", func() {
			Result(String)
			GRPC(func() {})
		})
	})
}

var UnaryRPCNoResultDSL = func() {
	Service("ServiceUnaryRPCNoResult", func() {
		Method("MethodUnaryRPCNoResult", func() {
			Payload(ArrayOf(String))
			GRPC(func() {})
		})
	})
}

var UnaryRPCWithErrorsDSL = func() {
	Service("ServiceUnaryRPCWithErrorsNoResult", func() {
		Method("MethodUnaryRPCWithErrorsNoResult", func() {
			Payload(String)
			Result(String)
			Error("timeout")
			Error("internal")
			Error("bad_request")
			GRPC(func() {
				Response("timeout", CodeCanceled)
				Response("internal", CodeUnknown)
				Response("bad_request", CodeInvalidArgument)
			})
		})
	})
}

var ServerStreamingRPCDSL = func() {
	Service("ServiceServerStreamingRPC", func() {
		Method("MethodServerStreamingRPC", func() {
			Payload(Int)
			StreamingResult(String)
			GRPC(func() {})
		})
	})
}

var ServerStreamingUserTypeDSL = func() {
	var UT = Type("UserType", func() {
		Attribute("IntField", Int)
	})
	Service("ServiceServerStreamingUserTypeRPC", func() {
		Method("MethodServerStreamingUserTypeRPC", func() {
			StreamingResult(UT)
			GRPC(func() {})
		})
	})
}

var ServerStreamingArrayDSL = func() {
	Service("ServiceServerStreamingArray", func() {
		Method("MethodServerStreamingArray", func() {
			StreamingResult(ArrayOf(Int))
			GRPC(func() {})
		})
	})
}

var ServerStreamingMapDSL = func() {
	var UT = Type("UserType", func() {
		Attribute("IntField", Int)
	})
	Service("ServiceServerStreamingMap", func() {
		Method("MethodServerStreamingMap", func() {
			StreamingResult(MapOf(String, UT))
			GRPC(func() {})
		})
	})
}

var ServerStreamingResultWithViewsDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("ResultType")
		Attributes(func() {
			Attribute("IntField", Int)
			Attribute("DoubleField", Float64)
		})
		View("default", func() {
			Attribute("IntField")
			Attribute("DoubleField")
		})
		View("tiny", func() {
			Attribute("IntField")
		})
	})
	Service("ServiceServerStreamingUserTypeRPC", func() {
		Method("MethodServerStreamingUserTypeRPC", func() {
			StreamingResult(RT)
			GRPC(func() {})
		})
	})
}

var ServerStreamingResultCollectionWithExplicitViewDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("ResultType")
		Attributes(func() {
			Attribute("IntField", Int)
			Attribute("DoubleField", Float64)
		})
		View("default", func() {
			Attribute("IntField")
			Attribute("DoubleField")
		})
		View("tiny", func() {
			Attribute("IntField")
		})
	})
	Service("ServiceServerStreamingResultTypeCollectionWithExplicitView", func() {
		Method("MethodServerStreamingResultTypeCollectionWithExplicitView", func() {
			StreamingResult(CollectionOf(RT), func() {
				View("tiny")
			})
			GRPC(func() {})
		})
	})
}

var ClientStreamingRPCDSL = func() {
	Service("ServiceClientStreamingRPC", func() {
		Method("MethodClientStreamingRPC", func() {
			StreamingPayload(Int)
			Result(String)
			GRPC(func() {})
		})
	})
}

var ClientStreamingRPCWithPayloadDSL = func() {
	Service("ServiceClientStreamingRPCWithPayload", func() {
		Method("MethodClientStreamingRPCWithPayload", func() {
			Payload(Int)
			StreamingPayload(Int)
			Result(String)
			GRPC(func() {})
		})
	})
}

var ClientStreamingNoResultDSL = func() {
	Service("ServiceClientStreamingNoResult", func() {
		Method("MethodClientStreamingNoResult", func() {
			StreamingPayload(Int)
			GRPC(func() {})
		})
	})
}

var BidirectionalStreamingRPCDSL = func() {
	var RT = ResultType("id", func() {
		Attributes(func() {
			Field(1, "a", Int)
			Field(2, "b", String)
		})
	})
	Service("ServiceBidirectionalStreamingRPC", func() {
		Method("MethodBidirectionalStreamingRPC", func() {
			StreamingPayload(Int)
			StreamingResult(RT)
			GRPC(func() {})
		})
	})
}

var BidirectionalStreamingRPCWithPayloadDSL = func() {
	var PT = Type("Payload", func() {
		Field(1, "a", Int)
		Field(2, "b", String)
	})
	Service("ServiceBidirectionalStreamingRPCWithPayload", func() {
		Method("MethodBidirectionalStreamingRPCWithPayload", func() {
			Payload(PT)
			StreamingPayload(Int)
			StreamingResult(UInt)
			GRPC(func() {})
		})
	})
}

var BidirectionalStreamingRPCWithErrorsDSL = func() {
	Service("ServiceBidirectionalStreamingRPCWithErrors", func() {
		Method("MethodBidirectionalStreamingRPCWithErrors", func() {
			StreamingPayload(Int)
			StreamingResult(Int)
			Error("timeout")
			Error("internal")
			Error("bad_request")
			GRPC(func() {
				Response("timeout", CodeCanceled)
				Response("internal", CodeUnknown)
				Response("bad_request", CodeInvalidArgument)
			})
		})
	})
}

var MessageUserTypeWithPrimitivesDSL = func() {
	var PayloadT = Type("PayloadT", func() {
		Field(1, "BooleanField", Boolean)
		Field(2, "IntField", Int)
		Field(3, "Int32Field", Int32)
		Field(4, "Int64Field", Int64)
		Field(5, "UIntField", UInt)
		Field(6, "UInt32Field", UInt32)
		Field(7, "UInt64Field", UInt64)
	})
	var ResultT = ResultType("application/vnd.goa.resultt", func() {
		TypeName("ResultT")
		Attributes(func() {
			Attribute("Float32Field", Float32, func() {
				Meta("rpc:tag", "1")
			})
			Attribute("Float64Field", Float64, func() {
				Meta("rpc:tag", "2")
			})
			Attribute("StringField", String, func() {
				Meta("rpc:tag", "3")
			})
			Attribute("BytesField", Bytes, func() {
				Meta("rpc:tag", "4")
			})
		})
	})
	Service("ServiceMessageUserTypeWithPrimitives", func() {
		Method("MethodMessageUserTypeWithPrimitives", func() {
			Payload(PayloadT)
			Result(ResultT)
			GRPC(func() {})
		})
	})
}

var MessageUserTypeWithNestedUserTypesDSL = func() {
	var UTLevel2 = Type("UTLevel2", func() {
		Field(2, "Int64Field", Int64)
	})
	var UTLevel1 = Type("UTLevel1", func() {
		Field(1, "Int32Field", Int32)
		Field(2, "Int64Field", Int64)
		Field(3, "UTLevel2", UTLevel2)
	})
	var UT = Type("UT", func() {
		Field(1, "BooleanField", Boolean)
		Field(2, "IntField", Int)
		Field(3, "UTLevel1", UTLevel1)
	})
	var RecursiveT = ResultType("application/vnd.goa.recursivet", func() {
		TypeName("RecursiveT")
		Attributes(func() {
			Field(1, "Recursive", "RecursiveT")
		})
	})
	Service("ServiceMessageUserTypeWithNestedUserTypes", func() {
		Method("MethodMessageUserTypeWithNestedUserTypes", func() {
			Payload(UT)
			Result(RecursiveT)
			GRPC(func() {})
		})
	})
}

var MessageResultTypeCollectionDSL = func() {
	var RT = ResultType("application/vnd.goa.rt", func() {
		TypeName("RT")
		Attributes(func() {
			Field(1, "IntField", Int)
		})
	})
	Service("ServiceMessageUserTypeWithNestedUserTypes", func() {
		Method("MethodMessageUserTypeWithNestedUserTypes", func() {
			Result(CollectionOf(RT))
			GRPC(func() {})
		})
	})
}

var MessageArrayDSL = func() {
	var UT = Type("UT", func() {
		Field(1, "ArrayOfPrimitives", ArrayOf(UInt))
		Field(2, "TwoDArray", ArrayOf(ArrayOf(Bytes)))
		Field(3, "ThreeDArray", ArrayOf(ArrayOf(ArrayOf(Bytes))))
		Field(4, "ArrayOfMaps", ArrayOf(MapOf(String, Float64)))
	})
	Service("ServiceMessageArray", func() {
		Method("MethodMessageArray", func() {
			Payload(UT)
			Result(ArrayOf(UT))
			GRPC(func() {})
		})
	})
}

var MessageMapDSL = func() {
	var UTLevel1 = Type("UTLevel1", func() {
		Field(1, "MapOfMapOfPrimitives", MapOf(String, MapOf(Int, UInt)))
	})
	var UT = Type("UT", func() {
		Field(1, "MapOfPrimitives", MapOf(UInt, Boolean))
		Field(2, "MapOfPrimitiveUTArray", MapOf(Int32, ArrayOf(UTLevel1)))
	})
	Service("ServiceMessageMap", func() {
		Method("MethodMessageMap", func() {
			Payload(MapOf(Int, UT))
			Result(UT)
			GRPC(func() {})
		})
	})
}

var MessagePrimitiveDSL = func() {
	Service("ServiceMessagePrimitive", func() {
		Method("MethodMessagePrimitive", func() {
			Payload(UInt)
			Result(Int)
			GRPC(func() {})
		})
	})
}

var MessageWithMetadataDSL = func() {
	var UTLevel1 = Type("UTLevel1", func() {
		Field(1, "Int32Field", Int32)
		Field(2, "Int64Field", Int64)
	})
	var RequestUT = Type("RequestUT", func() {
		Field(1, "BooleanField", Boolean)
		Field(2, "InMetadata", Int)
		Field(3, "UTLevel1", UTLevel1)
	})
	var ResponseUT = Type("ResponseUT", func() {
		Field(1, "InTrailer", Boolean)
		Field(2, "InHeader", Int)
		Field(3, "UTLevel1", UTLevel1)
	})
	Service("ServiceMessageWithMetadata", func() {
		Method("MethodMessageWithMetadata", func() {
			Payload(RequestUT)
			Result(ResponseUT)
			GRPC(func() {
				Metadata(func() {
					Attribute("InMetadata:Authorization")
				})
				Response(CodeOK, func() {
					Headers(func() {
						Attribute("InHeader:Location")
					})
					Trailers(func() {
						Attribute("InTrailer")
					})
				})
			})
		})
	})
}

var MessageWithSecurityAttrsDSL = func() {
	var JWTAuth = JWTSecurity("jwt", func() {
		Scope("api:read", "Read-only access")
	})
	var APIKeyAuth = APIKeySecurity("api_key", func() {})
	var BasicAuth = BasicAuthSecurity("basic", func() {})
	var OAuth2Auth = OAuth2Security("oauth2", func() {
		Scope("api:write", "Read and write access")
	})
	var RequestUT = Type("RequestUT", func() {
		Field(1, "BooleanField", Boolean)
		TokenField(2, "token", String)
		AccessTokenField(3, "oauth_token", String)
		APIKey("api_key", "key", String)
		Username("username", String)
		Password("password", String)
	})
	Service("ServiceMessageWithSecurity", func() {
		Method("MethodMessageWithSecurity", func() {
			Security(JWTAuth, OAuth2Auth, APIKeyAuth, BasicAuth)
			Payload(RequestUT)
			GRPC(func() {
				Message(func() {
					Attribute("oauth_token")
				})
			})
		})
	})
}
