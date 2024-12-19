package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var NoInterceptorExampleDSL = func() {
	var _ = Service("NoInterceptorService", func() {
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var ServerInterceptorExampleDSL = func() {
	var SInterceptor = Interceptor("test")
	var _ = Service("ServerInterceptorService", func() {
		ServerInterceptor(SInterceptor)
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var ClientInterceptorExampleDSL = func() {
	var CInterceptor = Interceptor("test")
	var _ = Service("ClientInterceptorService", func() {
		ClientInterceptor(CInterceptor)
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var ServerInterceptorByNameExampleDSL = func() {
	var _ = Interceptor("test")
	var _ = Service("ServerInterceptorByNameService", func() {
		ServerInterceptor("test")
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var MultipleInterceptorsExampleDSL = func() {
	var _ = Interceptor("test")
	var _ = Interceptor("test2")
	var SInterceptor = Interceptor("test3")
	var CInterceptor = Interceptor("test4")
	var _ = Service("MultipleInterceptorsService", func() {
		ServerInterceptor("test", SInterceptor)
		ClientInterceptor("test2", CInterceptor)
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var MultipleServicesInterceptorsExampleDSL = func() {
	var _ = Interceptor("test")
	var _ = Interceptor("test2")
	var SInterceptor = Interceptor("test3")
	var CInterceptor = Interceptor("test4")
	var _ = Service("MultipleServicesInterceptorsService", func() {
		ServerInterceptor("test", SInterceptor)
		ClientInterceptor("test2", CInterceptor)
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
	var _ = Service("MultipleServicesInterceptorsService2", func() {
		ServerInterceptor("test", SInterceptor)
		ClientInterceptor("test2", CInterceptor)
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var APIInterceptorExampleDSL = func() {
	var APIInterceptor = Interceptor("api")
	var _ = API("test", func() {
		ServerInterceptor(APIInterceptor)
		ClientInterceptor(APIInterceptor)
	})
	var _ = Service("APIInterceptorService", func() {
		Method("Method", func() { HTTP(func() { GET("/") }) })
	})
}

var ChainedInterceptorExampleDSL = func() {
	var APIInterceptor = Interceptor("api")
	var ServiceInterceptor = Interceptor("service")
	var MethodInterceptor = Interceptor("method")

	var _ = API("test", func() {
		ServerInterceptor(APIInterceptor)
		ClientInterceptor(APIInterceptor)
	})

	var _ = Service("ChainedInterceptorService", func() {
		ServerInterceptor(ServiceInterceptor)
		ClientInterceptor(ServiceInterceptor)
		Method("Method", func() {
			ServerInterceptor(MethodInterceptor)
			ClientInterceptor(MethodInterceptor)
			HTTP(func() { GET("/") })
		})
	})
}
