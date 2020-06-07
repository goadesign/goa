package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var CookieObjectResultDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
				})
			})
		})
	})
}

var CookieStringResultDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(String)
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
				})
			})
		})
	})
}

const CookieMaxAgeValue = 3600

var CookieMaxAgeDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieMaxAge(CookieMaxAgeValue)
				})
			})
		})
	})
}

const CookieDomainValue = "goa.design"

var CookieDomainDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieDomain(CookieDomainValue)
				})
			})
		})
	})
}

const CookiePathValue = "/path"

var CookiePathDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookiePath(CookiePathValue)
				})
			})
		})
	})
}

var CookieSecureDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieSecure()
				})
			})
		})
	})
}

var CookieHTTPOnlyDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieHTTPOnly()
				})
			})
		})
	})
}
