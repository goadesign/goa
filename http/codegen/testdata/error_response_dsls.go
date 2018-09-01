package testdata

import (
	. "goa.design/goa/dsl"
)

var DefaultErrorResponseDSL = func() {
	Service("ServiceDefaultErrorResponse", func() {
		Method("MethodDefaultErrorResponse", func() {
			Error("bad_request")
			HTTP(func() {
				GET("/one/two")
				Response("bad_request", StatusBadRequest)
			})
		})
	})
}

var PrimitiveErrorResponseDSL = func() {
	Service("ServicePrimitiveErrorResponse", func() {
		Method("MethodPrimitiveErrorResponse", func() {
			Error("bad_request", String)
			Error("internal_error", String)
			HTTP(func() {
				GET("/one/two")
				Response("bad_request", StatusBadRequest)
				Response("internal_error", StatusInternalServerError)
			})
		})
	})
}

var ServiceErrorResponseDSL = func() {
	Service("ServiceServiceErrorResponse", func() {
		Error("bad_request")
		HTTP(func() {
			Response("bad_request", StatusBadRequest)
		})
		Method("MethodServiceErrorResponse", func() {
			Error("internal_error")
			HTTP(func() {
				GET("/one/two")
				Response("internal_error", StatusInternalServerError)
			})
		})
	})
}
