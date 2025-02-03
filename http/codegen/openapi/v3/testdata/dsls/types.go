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

func StringEnumBodyDSL() func() {
	return func() {
		var T1 = Type("T1", func() {
			Attribute("my_attr", String, func() {
				Enum("a", "b", "c")
			})
		})

		var T2 = Type("T2", func() {
			Attribute("my_attr", String, func() {
				Enum("d", "e")
			})
		})

		var _ = Service("svc_enum_1", func() {
			Method("method_enum", func() {
				Payload(func() {
					Reference(T1)
					Attribute("my_attr")
					Required("my_attr")
				})
				HTTP(func() {
					POST("/")
				})
			})
		})

		var _ = Service("svc_enum_2", func() {
			Method("method_enum", func() {
				Payload(func() {
					Reference(T2)
					Attribute("my_attr")
					Required("my_attr")
				})

				HTTP(func() {
					POST("/other")
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

func MapBodyDSL(svcName, metName string) func() {
	return func() {
		var _ = Service(svcName, func() {
			Method(metName, func() {
				Payload(func() {
					Attribute("map", MapOf(String, Any))
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
					Attribute("misc", Any)
				})
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func MultiCookieResponseBodyDSL(svcName, metName string) func() {
	return func() {
		var U = Type("U", func() {
			Attribute("name")
			Attribute("cookie")
		})
		var T = Type("T", func() {
			Attribute("name")
		})
		var _ = Service(svcName, func() {
			Method("other", func() {
				Result(U)
				HTTP(func() {
					GET("/cookie")
					Response(StatusOK, func() { Cookie("cookie") })
				})
			})
			Method(metName, func() {
				Result(T)
				HTTP(func() {
					GET("/")
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

func ForcedTypeDSL(svcName, metName string) func() {
	return func() {
		var _ = Type("Forced", func() {
			Attribute("foo")
			Meta("type:generate:force")
		})
		var _ = Service(svcName, func() {
			Method(metName, func() {
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}

func ForcedResultTypeDSL(svcName, metName string) func() {
	return func() {
		var _ = ResultType("Forced", func() {
			Attributes(func() {
				Attribute("foo")
			})
			Meta("type:generate:force")
		})
		var _ = Service(svcName, func() {
			Method(metName, func() {
				HTTP(func() {
					POST("/")
				})
			})
		})
	}
}
