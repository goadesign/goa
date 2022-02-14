package dsls

import . "goa.design/goa/v3/dsl"

func StringBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Payload(String)
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func AliasStringBodyDSL(svcName, metName string) func() {
	return func() {
		var UUID = Type("UUID", String, func() {
			Format(FormatUUID)
		})
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Payload(UUID)
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func ObjectBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Payload(func() {
					Attribute("name")
					Attribute("age", Int)
				})
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func StringResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Result(String)
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func ObjectResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Result(func() {
					Attribute("name")
					Attribute("age", Int)
				})
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func StringStreamingResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				StreamingResult(String)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

func ObjectStreamingResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				StreamingResult(func() {
					Attribute("name")
					Attribute("age", Int)
				})
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

func StringErrorResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Error("bad", String)
				HTTP(func() {
					POST("/")
					Response("bad", StatusBadRequest)
				})
			})
		})
	}
}

func ObjectErrorResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Error("bad", func() {
					Attribute("name")
					Attribute("age", Int)
				})
				HTTP(func() {
					POST("/")
					Response("bad", StatusBadRequest)
				})
			})
		})
	}
}
