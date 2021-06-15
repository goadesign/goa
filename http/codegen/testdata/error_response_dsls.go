package testdata

import (
	. "goa.design/goa/v3/dsl"
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

var DefaultErrorResponseWithContentTypeDSL = func() {
	Service("ServiceDefaultErrorResponse", func() {
		Method("MethodDefaultErrorResponse", func() {
			Error("bad_request")
			HTTP(func() {
				GET("/one/two")
				Response("bad_request", StatusBadRequest, func() {
					ContentType("application/xml")
				})
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
				Response(StatusInternalServerError, "internal_error")
			})
		})
	})
}

var PrimitiveErrorInResponseHeaderDSL = func() {
	Service("ServicePrimitiveErrorInResponseHeader", func() {
		Method("MethodPrimitiveErrorInResponseHeader", func() {
			Error("bad_request", String)
			Error("internal_error", Int)
			HTTP(func() {
				GET("/one/two")
				Response("bad_request", StatusBadRequest, func() {
					Header("string")
				})
				Response(StatusInternalServerError, "internal_error", func() {
					Header("int")
				})
			})
		})
	})
}

var APIPrimitiveErrorResponseDSL = func() {
	var _ = API("test", func() {
		Error("bad_request", String)
		HTTP(func() {
			Response(StatusBadRequest, "bad_request")
		})
	})
	Service("ServiceAPIPrimitiveErrorResponse", func() {
		Method("MethodAPIPrimitiveErrorResponse", func() {
			Error("bad_request", String)
			Error("internal_error")
			HTTP(func() {
				GET("/one/two")
				Response("internal_error", StatusInternalServerError)
			})
		})
	})
}

var ServiceErrorResponseDSL = func() {
	Service("ServiceServiceErrorResponse", func() {
		Error("bad_request")
		HTTP(func() {
			Response(StatusBadRequest, "bad_request")
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

var APIErrorResponseDSL = func() {
	var _ = API("test", func() {
		Error("bad_request")
		HTTP(func() {
			Response(StatusBadRequest, "bad_request")
		})
	})
	Service("ServiceServiceErrorResponse", func() {
		Method("MethodServiceErrorResponse", func() {
			Error("bad_request")
			Error("internal_error")
			HTTP(func() {
				GET("/one/two")
				Response("internal_error", StatusInternalServerError)
			})
		})
	})
}

var APIErrorResponseWithContentTypeDSL = func() {
	var _ = API("test", func() {
		Error("bad_request")
		HTTP(func() {
			Response(StatusBadRequest, "bad_request", func() {
				ContentType("application/xml")
			})
		})
	})
	Service("ServiceServiceErrorResponse", func() {
		Method("MethodServiceErrorResponse", func() {
			Error("bad_request")
			Error("internal_error")
			HTTP(func() {
				GET("/one/two")
				Response("internal_error", StatusInternalServerError)
			})
		})
	})
}

var APINoBodyErrorResponseDSL = func() {
	var StringError = Type("StringError", func() { Attribute("header") })
	var _ = API("test", func() {
		Error("bad_request", StringError)
		HTTP(func() {
			Response("bad_request", StatusBadRequest, func() {
				Header("header")
			})
		})
	})
	Service("ServiceNoBodyErrorResponse", func() {
		Error("bad_request")
		Method("MethodServiceErrorResponse", func() {
			HTTP(func() {
				GET("/one/two")
			})
		})
	})
}

var APINoBodyErrorResponseWithContentTypeDSL = func() {
	var StringError = ResultType("application/vnd.string.error", func() {
		ContentType("application/xml")
		Attribute("header")
	})
	var _ = API("test", func() {
		Error("bad_request", StringError)
		HTTP(func() {
			Response("bad_request", StatusBadRequest, func() {
				Header("header")
			})
		})
	})
	Service("ServiceNoBodyErrorResponse", func() {
		Error("bad_request")
		Method("MethodServiceErrorResponse", func() {
			HTTP(func() {
				GET("/one/two")
			})
		})
	})
}

var NoBodyErrorResponseDSL = func() {
	var StringError = Type("StringError", func() { Attribute("header") })
	Service("ServiceNoBodyErrorResponse", func() {
		Error("bad_request", StringError)
		HTTP(func() {
			Response("bad_request", StatusBadRequest, func() {
				Header("header")
			})
		})
		Method("MethodServiceErrorResponse", func() {
			HTTP(func() {
				GET("/one/two")
			})
		})
	})
}

var NoBodyErrorResponseWithContentTypeDSL = func() {
	var StringError = ResultType("application/vnd.string.error", func() {
		ContentType("application/xml")
		Attribute("header")
	})
	Service("ServiceNoBodyErrorResponse", func() {
		Error("bad_request", StringError)
		HTTP(func() {
			Response("bad_request", StatusBadRequest, func() {
				Header("header")
			})
		})
		Method("MethodServiceErrorResponse", func() {
			HTTP(func() {
				GET("/one/two")
			})
		})
	})
}
