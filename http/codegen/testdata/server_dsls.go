package testdata

import (
	. "goa.design/goa/dsl"
)

var ServerNoPayloadNoResultDSL = func() {
	Service("ServiceNoPayloadNoResult", func() {
		Method("MethodNoPayloadNoResult", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ServerPayloadNoResultDSL = func() {
	Service("ServicePayloadNoResult", func() {
		Method("MethodPayloadNoResult", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ServerNoPayloadResultDSL = func() {
	Service("ServiceNoPayloadResult", func() {
		Method("MethodNoPayloadResult", func() {
			Result(func() {
				Attribute("b", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ServerPayloadResultDSL = func() {
	Service("ServicePayloadResult", func() {
		Method("MethodPayloadResult", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			Result(func() {
				Attribute("b", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var ServerPayloadResultErrorDSL = func() {
	Service("ServicePayloadResultError", func() {
		Method("MethodPayloadResultError", func() {
			Payload(func() {
				Attribute("a", Boolean)
			})
			Result(func() {
				Attribute("b", Boolean)
			})
			Error("e", func() {
				Attribute("c", Boolean)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK)
				Response("e", func() {
					Code(StatusConflict)
				})
			})
		})
	})
}

var ServerMultiBasesDSL = func() {
	Service("ServiceMultiBases", func() {
		HTTP(func() {
			Path("/base_1")
			Path("/base_2")
		})
		Method("MethodMultiBases", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				GET("/{id}")
			})
		})
	})
}

var ServerMultiEndpointsDSL = func() {
	Service("ServiceMultiEndpoints", func() {
		HTTP(func() {
			Path("/server_multi_endpoints")
		})
		Method("MethodMultiEndpoints1", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				GET("/{id}")
			})
		})
		Method("MethodMultiEndpoints2", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ServerFileServerDSL = func() {
	Service("ServiceFileServer", func() {
		HTTP(func() {
			Path("/server_file_server")
		})
		Files("/file1.json", "/path/to/file1.json")
		Files("/file2.json", "/path/to/file2.json")
		Files("/file3.json", "/path/to/file3.json")
	})
}

var ServerMixedDSL = func() {
	Service("ServerMixed", func() {
		Method("MethodMixed", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				GET("/{id}")
			})
		})
		Files("/file1.json", "/path/to/file1.json")
		Files("/file2.json", "/path/to/file2.json")
	})
}

var ServerMultipartDSL = func() {
	Service("ServiceMultipart", func() {
		Method("MethodMultiBases", func() {
			Payload(String)
			HTTP(func() {
				GET("/")
				MultipartRequest()
			})
		})
	})
}
