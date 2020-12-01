package testdata

import . "goa.design/goa/v3/dsl"

var NoServerDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var SameAPIServiceNameDSL = func() {
	API("Service", func() {})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var SingleServerSingleHostDSL = func() {
	API("SingleServerSingleHost", func() {
		Server("SingleHost", func() {
			Services("Service")
			Host("dev", func() {
				URI("http://example:8090")
				URI("https://example:80")
				URI("grpc://example:8080")
				URI("http://[::1]:8080")
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var SingleServerSingleHostWithVariablesDSL = func() {
	API("SingleServerSingleHostWithVariables", func() {
		Server("SingleHost", func() {
			Services("Service")
			Host("dev", func() {
				URI("http://example-{int}-{uint}-{float32}:8090")
				Variable("int", Int, func() {
					Default(1)
				})
				Variable("uint", UInt, func() {
					Default(1)
				})
				Variable("float32", Float32, func() {
					Default(1.1)
				})
				URI("https://example-{int32}-{int64}-{uint32}-{uint64}-{float64}:80/{bool}")
				Variable("int32", Int32, func() {
					Default(1)
				})
				Variable("int64", Int64, func() {
					Default(1)
				})
				Variable("uint32", UInt32, func() {
					Default(1)
				})
				Variable("uint64", UInt64, func() {
					Default(1)
				})
				Variable("float64", Float64, func() {
					Default(1)
				})
				Variable("bool", Boolean, func() {
					Default(true)
				})
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var ServerHostingServiceWithFileServerDSL = func() {
	API("ServerHostingServiceWithFileServer", func() {
		Server("SingleHost", func() {
			Services("Service")
		})
	})
	Service("Service", func() {
		Files("/file.json", "path.json")
	})
}

var ServerHostingServiceSubsetDSL = func() {
	API("ServerHostingServiceSubset", func() {
		Server("SingleHost", func() {
			Services("Service")
			Host("dev", func() {
				URI("http://example:8090")
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
	Service("IgnoredService", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var ServerHostingMultipleServicesDSL = func() {
	API("ServerHostingMultipleServices", func() {
		Server("SingleHost", func() {
			Services("Service", "AnotherService")
			Host("dev", func() {
				URI("http://example:8090")
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
	Service("AnotherService", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var SingleServerMultipleHostsDSL = func() {
	API("SingleServerMultipleHosts", func() {
		Server("MultipleHosts", func() {
			Services("Service")
			Host("dev", func() {
				URI("http://example:8090")
			})
			Host("stage", func() {
				URI("https://example")
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var SingleServerMultipleHostsWithVariablesDSL = func() {
	API("SingleServerMultipleHostsWithVariables", func() {
		Server("MultipleHostsWithVariables", func() {
			Services("Service")
			Host("dev", func() {
				URI("http://example-{version}:8090")
				Variable("version", String, "Version", func() {
					Enum("v1", "v2")
				})
			})
			Host("stage", func() {
				URI("https://example-{domain}:{port}")
				Variable("domain", String, "Domain", func() {
					Default("test")
				})
				Variable("port", String, "Port", func() {
					Default("8080")
				})
			})
		})
	})
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var ServiceForOnlyHTTPDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var ServiceForOnlyGRPCDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			GRPC(func() {})
		})
	})
}

var ServiceForHTTPAndPartOfGRPCDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
	Service("AnotherService", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var ConflictWithAPINameAndServiceNamesIncludingMultipartDSL = func() {
	var _ = API("aloha", func() {
		Title("conflict with API name and service names including multipart")
	})
	var _ = Service("aloha", func() { // same as API name
		Method("create", func() {
			Payload(func() {
				Attribute("price", Int)
			})
			HTTP(func() {
				POST("/aloha")
				MultipartRequest()
			})
		})
	})
	var _ = Service("alohaapi", func() { // API name + 'api' suffix
		Method("create", func() {
			Payload(func() {
				Attribute("price", Int)
			})
			HTTP(func() {
				POST("/aloha")
			})
		})
	})
	var _ = Service("alohaapi1", func() { // API name + 'api' suffix + sequential no.
		Method("create", func() {
			Payload(func() {
				Attribute("price", Int)
			})
			HTTP(func() {
				POST("/aloha")
			})
		})
	})
}
