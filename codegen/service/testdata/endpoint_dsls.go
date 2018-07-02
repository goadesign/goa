package testdata

import (
	. "goa.design/goa/design"
	. "goa.design/goa/dsl"
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
			Result(ViewType)
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
