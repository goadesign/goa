package dsls

import . "goa.design/goa/dsl"

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
