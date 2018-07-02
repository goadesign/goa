package testdata

import (
	. "goa.design/goa/http/design"
	. "goa.design/goa/http/dsl"
)

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
				GET("/")
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
		Attribute("a", String)
	})
	Service("StreamingResultWithViewsService", func() {
		Method("StreamingResultWithViewsMethod", func() {
			Payload(Request)
			StreamingResult(Result)
			HTTP(func() {
				GET("/")
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
