package dsls

import . "goa.design/goa/v3/dsl"

var DescOnly = func(svc, met, desc string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				Description(desc)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var RequestStringBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				Payload(String, func() {
					Description("body")
				})
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

var RequestObjectBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				Payload(func() {
					Attribute("name")
				})
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

var RequestStreamingStringBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				StreamingPayload(String)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var RequestStreamingObjectBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				StreamingPayload(func() {
					Description("body")
					Attribute("name")
				})
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var ResponseStringBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				Result(String)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var ResponseObjectBody = func(svc, met string) func() {
	return func() {
		var _ = Service(svc, func() {
			Method(met, func() {
				Result(func() { Attribute("name") })
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var ResponseArrayOfString = func(svc, met string) func() {
	return func() {
		var arrayOfType = Type("arrayOfString", func() {
			Attribute("children", ArrayOf(String))
		})

		var arrResultType = ResultType("arrResultType", func() {
			Attribute("result", arrayOfType)
		})

		var _ = Service(svc, func() {
			Method(met, func() {
				Result(arrResultType)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var ResponseRecursiveUserType = func(svc, met string) func() {
	return func() {
		var recursiveType = Type("recursiveType", func() {
			Attribute("recursive", "recursiveType")
		})

		var _ = Service(svc, func() {
			Method(met, func() {
				Result(recursiveType)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}

var ResponseRecursiveArrayUserType = func(svc, met string) func() {
	return func() {
		var recursiveType = Type("recursiveType", func() {
			Attribute("children", ArrayOf("recursiveType"))
		})

		var arrResultType = ResultType("recursiveArrayResultType", func() {
			Attribute("result", recursiveType)
		})

		var _ = Service(svc, func() {
			Method(met, func() {
				Result(arrResultType)
				HTTP(func() {
					GET("/")
				})
			})
		})
	}
}
