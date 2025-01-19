package testdata

import . "goa.design/goa/v3/dsl"

var NoInterceptorsDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}

var ValidInterceptorsDSL = func() {
	Interceptor("api", func() {})
	API("API", func() {
		ServerInterceptor("api")
	})

	Service("Service", func() {
		Interceptor("service", func() {})
		ServerInterceptor("service")

		Method("Method", func() {
			Interceptor("method", func() {})
			ServerInterceptor("method")
		})
	})
}

var DuplicateInterceptorsDSL = func() {
	Interceptor("duplicate", func() {})
	API("API", func() {
		ServerInterceptor("duplicate")
	})

	Service("Service", func() {
		ServerInterceptor("duplicate")
		Method("Method", func() {
			ServerInterceptor("duplicate")
		})
	})
}

var MixedInterceptorsDSL = func() {
	Interceptor("api", func() {})
	Interceptor("api-client", func() {})
	API("API", func() {
		ServerInterceptor("api")
		ClientInterceptor("api-client")
	})

	Service("Service", func() {
		Interceptor("service", func() {})
		Interceptor("service-client", func() {})
		ServerInterceptor("service")
		ClientInterceptor("service-client")

		Method("Method", func() {
			Interceptor("method", func() {})
			Interceptor("method-client", func() {})
			ServerInterceptor("method")
			ClientInterceptor("method-client")
		})
	})
}

var UndefinedInterceptorDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			ServerInterceptor("undefined")
		})
	})
}

var EmptyInterceptorNameDSL = func() {
	Service("Service", func() {
		Method("Method", func() {
			ServerInterceptor("")
		})
	})
}
