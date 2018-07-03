package testdata

import (
	. "goa.design/goa/http/design"
	. "goa.design/goa/http/dsl"
)

var EmptyResultEmptyResponseDSL = func() {
	Service("EmptyResultEmptyResponse", func() {
		Method("Method", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var NonEmptyResultEmptyResponseDSL = func() {
	Service("NonEmptyResultEmptyResponse", func() {
		Method("Method", func() {
			Result(String)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var EmptyResultNonEmptyResponseDSL = func() {
	Service("EmptyResultNonEmptyResponse", func() {
		Method("Method", func() {
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

// var StringResultResponseWithHeadersDSL = func() {
// 	Service("StringResultResponseWithHeaders", func() {
// 		Method("Method", func() {
// 			Result(String)
// 			HTTP(func() {
// 				POST("/")
// 				Response(func() {
// 					Header("Location")
// 				})
// 			})
// 		})
// 	})
// }

var ObjectResultResponseWithHeadersDSL = func() {
	Service("ObjectResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("foo", String)
			})
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var ArrayResultResponseWithHeadersDSL = func() {
	Service("ArrayResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("foo", ArrayOf(String))
			})
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var MapResultResponseWithHeadersDSL = func() {
	Service("MapResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("foo", MapOf(String, String))
			})
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var EmptyResultResponseWithHeadersDSL = func() {
	Service("EmptyResultResponseWithHeaders", func() {
		Method("Method", func() {
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("foo:Location")
				})
			})
		})
	})
}
