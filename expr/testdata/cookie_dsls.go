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

const CookieSameSiteValue = CookieSameSiteStrict

var CookieSameSiteDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieSameSite(CookieSameSiteValue)
				})
			})
		})
	})
}

var CookieAttrBindingsDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
				Attribute("expiresIn", Int)
				Attribute("cookieDomain", String)
				Attribute("cookiePath", String)
				Attribute("isSecure", Boolean)
				Attribute("isHTTPOnly", Boolean)
				Attribute("sameSite", String)
				Required("cookie", "expiresIn", "cookieDomain", "cookiePath",
					"isSecure", "isHTTPOnly", "sameSite")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieAttributes("cookie", func() {
						MaxAge("expiresIn")
						Domain("cookieDomain")
						Path("cookiePath")
						Secure("isSecure")
						HTTPOnly("isHTTPOnly")
						SameSite("sameSite")
					})
				})
			})
		})
	})
}

var CookieAttrBindingMissingAttrDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
				Required("cookie")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieAttributes("cookie", func() {
						MaxAge("doesNotExist")
					})
				})
			})
		})
	})
}

var CookieAttrBindingWrongTypeDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
				Attribute("expiresIn", String)
				Required("cookie", "expiresIn")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieAttributes("cookie", func() {
						MaxAge("expiresIn")
					})
				})
			})
		})
	})
}

var CookieAttrBindingUndeclaredDSL = func() {
	Service("CookieSvc", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("cookie", String)
				Required("cookie")
			})
			HTTP(func() {
				POST("/")
				Response(StatusOK, func() {
					Cookie("cookie")
					CookieAttributes("notDeclared", func() {
						MaxAge("cookie")
					})
				})
			})
		})
	})
}

var CookieAttrBindingErrorDSL = func() {
	var SessionInvalid = Type("SessionInvalid", func() {
		ErrorName("name")
		Attribute("name", String)
		Attribute("reason", String)
		Attribute("retryAfter", Int)
		Required("name", "reason", "retryAfter")
	})
	Service("CookieSvc", func() {
		Method("Method", func() {
			Error("session_invalid", SessionInvalid)
			HTTP(func() {
				GET("/")
				Response("session_invalid", StatusUnauthorized, func() {
					Cookie("reason")
					CookieAttributes("reason", func() {
						MaxAge("retryAfter")
					})
				})
			})
		})
	})
}

var CookieAttrBindingErrorMissingAttrDSL = func() {
	var SessionInvalid = Type("SessionInvalid", func() {
		ErrorName("name")
		Attribute("name", String)
		Attribute("reason", String)
		Required("name", "reason")
	})
	Service("CookieSvc", func() {
		Method("Method", func() {
			Error("session_invalid", SessionInvalid)
			HTTP(func() {
				GET("/")
				Response("session_invalid", StatusUnauthorized, func() {
					Cookie("reason")
					CookieAttributes("reason", func() {
						MaxAge("doesNotExist")
					})
				})
			})
		})
	})
}

var CookieAttrBindingErrorWrongTypeDSL = func() {
	var SessionInvalid = Type("SessionInvalid", func() {
		ErrorName("name")
		Attribute("name", String)
		Attribute("reason", String)
		Attribute("retryAfter", String)
		Required("name", "reason", "retryAfter")
	})
	Service("CookieSvc", func() {
		Method("Method", func() {
			Error("session_invalid", SessionInvalid)
			HTTP(func() {
				GET("/")
				Response("session_invalid", StatusUnauthorized, func() {
					Cookie("reason")
					CookieAttributes("reason", func() {
						MaxAge("retryAfter")
					})
				})
			})
		})
	})
}
