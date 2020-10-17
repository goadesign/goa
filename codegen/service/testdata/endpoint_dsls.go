package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var SingleEndpointDSL = func() {
	var AType = Type("AType", func() {
		Attribute("a", String)
	})
	Service("SingleEndpoint", func() {
		Method("A", func() {
			Payload(AType)
		})
	})
}

var UseEndpointDSL = func() {
	Service("UseEndpoint", func() {
		Method("Use", func() {
			Payload(String)
		})
	})
}

var MultipleEndpointsDSL = func() {
	var BType = Type("BType", func() {
		Attribute("b", String)
	})
	var CType = Type("CType", func() {
		Attribute("c", String)
	})
	Service("MultipleEndpoints", func() {
		Method("B", func() {
			Payload(BType)
		})
		Method("C", func() {
			Payload(CType)
		})
	})
}

var NoPayloadEndpointDSL = func() {
	Service("NoPayload", func() {
		Method("NoPayload", func() {
		})
	})
}

var WithResultEndpointDSL = func() {
	var RType = ResultType("application/vnd.withresult", func() {
		TypeName("Rtype")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
		})
	})
	Service("WithResult", func() {
		Method("A", func() {
			Result(RType)
		})
	})
}

var WithResultMultipleViewsEndpointDSL = func() {
	var ViewType = ResultType("application/vnd.withresult.multiple.views", func() {
		TypeName("Viewtype")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
			Attribute("b")
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	Service("WithResultMultipleViews", func() {
		Method("A", func() {
			Result(ViewType, func() {
				View("tiny")
			})
		})
		Method("B", func() {
			Result(ViewType, func() {
				View("default")
			})
		})
	})
}

var StreamingResultEndpointDSL = func() {
	var AType = Type("AType", func() {
		Attribute("a", String)
	})
	var RType = ResultType("application/vnd.withresult", func() {
		TypeName("Rtype")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
		})
	})
	Service("StreamingResultEndpoint", func() {
		Method("StreamingResultMethod", func() {
			Payload(AType)
			StreamingResult(RType)
		})
	})
}

var StreamingPayloadEndpointDSL = func() {
	var AType = Type("AType", func() {
		Attribute("a", String)
	})
	var BType = Type("BType", func() {
		Attribute("x", String)
	})
	Service("StreamingPayloadEndpoint", func() {
		Method("StreamingPayloadMethod", func() {
			Payload(BType)
			StreamingPayload(AType)
			Result(AResult)
		})
	})
}

var StreamingResultNoPayloadEndpointDSL = func() {
	var RType = ResultType("application/vnd.withresult", func() {
		TypeName("Rtype")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
		})
	})
	Service("StreamingResultNoPayloadEndpoint", func() {
		Method("StreamingResultNoPayloadMethod", func() {
			StreamingResult(RType)
		})
	})
}

var BidirectionalStreamingEndpointDSL = func() {
	var AType = Type("AType", func() {
		Attribute("a", String)
	})
	var BType = Type("BType", func() {
		Attribute("x", String)
	})
	Service("BidirectionalStreamingEndpoint", func() {
		Method("BidirectionalStreamingMethod", func() {
			Payload(AType)
			StreamingPayload(BType)
			StreamingResult(AResult)
		})
	})
}
