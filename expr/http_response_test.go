package expr_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestHTTPResponseValidation(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"empty", emptyResultEmptyResponseDSL, ""},
		{"non empty result", nonEmptyResultEmptyResponseDSL, ""},
		{"non empty response", emptyResultNonEmptyResponseDSL, ""},
		{"string result", stringResultResponseWithHeadersDSL, ""},
		{"string result", stringResultResponseWithTextContentTypeDSL, ""},
		{"object result", objectResultResponseWithHeadersDSL, ""},
		{"array result", arrayResultResponseWithHeadersDSL, ""},
		{"map result", mapResultResponseWithHeadersDSL, `service "MapResultResponseWithHeaders" HTTP endpoint "Method": attribute "foo" used in HTTP headers must be a primitive type or an array of primitive types.`},
		{"invalid", emptyResultResponseWithHeadersDSL, `HTTP response of service "EmptyResultResponseWithHeaders" HTTP endpoint "Method": response defines headers but result is empty`},
		{"implicit object in header", implicitObjectResultResponseWithHeadersDSL, `service "ArrayObjectResultResponseWithHeaders" HTTP endpoint "Method": attribute "foo" used in HTTP headers must be a primitive type or an array of primitive types.`},
		{"array of object in header", arrayObjectResultResponseWithHeadersDSL, `service "ArrayObjectResultResponseWithHeaders" HTTP endpoint "Method": Array result is mapped to an HTTP header but is not an array of primitive types.`},
		{"not string or []byte", intResultResponseWithTextContentTypeDSL, `HTTP response of service "StringResultResponseWithHeaders" HTTP endpoint "Method": Result type must be String or Bytes when ContentType is 'text/plain'`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

var emptyResultEmptyResponseDSL = func() {
	Service("EmptyResultEmptyResponse", func() {
		Method("Method", func() {
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var nonEmptyResultEmptyResponseDSL = func() {
	Service("NonEmptyResultEmptyResponse", func() {
		Method("Method", func() {
			Result(String)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var emptyResultNonEmptyResponseDSL = func() {
	Service("EmptyResultNonEmptyResponse", func() {
		Method("Method", func() {
			HTTP(func() {
				POST("/")
				Response(StatusOK)
			})
		})
	})
}

var stringResultResponseWithHeadersDSL = func() {
	Service("StringResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(String)
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("Location")
				})
			})
		})
	})
}

var stringResultResponseWithTextContentTypeDSL = func() {
	Service("StringResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(String)
			HTTP(func() {
				POST("/")
				Response(func() {
					ContentType("text/plain")
				})
			})
		})
	})
}

var objectResultResponseWithHeadersDSL = func() {
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

var implicitObjectResultResponseWithHeadersDSL = func() {
	Service("ArrayObjectResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(func() {
				Attribute("foo", func() {
					Attribute("bar", String)
					Attribute("baz", String)
				})
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

var arrayObjectResultResponseWithHeadersDSL = func() {
	var Obj = Type("Obj", func() {
		Attribute("foo", String)
	})
	Service("ArrayObjectResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(ArrayOf(Obj))
			HTTP(func() {
				POST("/")
				Response(func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var arrayResultResponseWithHeadersDSL = func() {
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

var mapResultResponseWithHeadersDSL = func() {
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

var emptyResultResponseWithHeadersDSL = func() {
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

var intResultResponseWithTextContentTypeDSL = func() {
	Service("StringResultResponseWithHeaders", func() {
		Method("Method", func() {
			Result(Int)
			HTTP(func() {
				POST("/")
				Response(func() {
					ContentType("text/plain")
				})
			})
		})
	})
}
