package expr_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestHTTPErrorResponseValidation(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"header string error", stringErrorResponseWithHeadersDSL, ""},
		{"header object result", objectErrorResponseWithHeadersDSL, ""},
		{"header array result", arrayErrorResponseWithHeadersDSL, ""},
		{"header map result", mapErrorResponseWithHeadersDSL, `HTTP response of service "MapErrorResponseWithHeaders" HTTP endpoint "Method": attribute "foo" used in HTTP headers must be a primitive type or an array of primitive types.`},
		{"implicit object in header", implicitObjectErrorResponseWithHeadersDSL, `HTTP response of service "ArrayObjectErrorResponseWithHeaders" HTTP endpoint "Method": attribute "foo" used in HTTP headers must be a primitive type or an array of primitive types.`},
		{"array of object in header", arrayObjectErrorResponseWithHeadersDSL, `HTTP response of service "ArrayObjectErrorResponseWithHeaders" HTTP endpoint "Method": Array error type is mapped to an HTTP header but is not an array of primitive types.`},
		{"map in header", mapErrorTypeResponseWithHeadersDSL, `HTTP response of service "MapErrorTypeResponseWithHeaders" HTTP endpoint "Method": error type must be a primitive type or an array of primitive types.`},
		{"missing header result attribute", missingHeaderErrorAttributeDSL, `HTTP response of service "MissingHeaderErrorAttribute" HTTP endpoint "Method": header "bar" has no equivalent attribute in error type, use notation 'attribute_name:header_name' to identify corresponding error type attribute.`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("\ngot error %q\nexpected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

var stringErrorResponseWithHeadersDSL = func() {
	Service("StringErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", String)
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("Location")
				})
			})
		})
	})
}

var objectErrorResponseWithHeadersDSL = func() {
	Service("ObjectErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", func() {
				Attribute("foo", String)
			})
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var implicitObjectErrorResponseWithHeadersDSL = func() {
	Service("ArrayObjectErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", func() {
				Attribute("foo", func() {
					Attribute("bar", String)
					Attribute("baz", String)
				})
			})
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var arrayObjectErrorResponseWithHeadersDSL = func() {
	var Obj = Type("Obj", func() {
		Attribute("foo", String)
	})
	Service("ArrayObjectErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", ArrayOf(Obj))
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var mapErrorTypeResponseWithHeadersDSL = func() {
	Service("MapErrorTypeResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", MapOf(String, Int))
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("Location")
				})
			})
		})
	})
}

var arrayErrorResponseWithHeadersDSL = func() {
	Service("ArrayErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", func() {
				Attribute("foo", ArrayOf(String))
			})
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var mapErrorResponseWithHeadersDSL = func() {
	Service("MapErrorResponseWithHeaders", func() {
		Method("Method", func() {
			Error("error", func() {
				Attribute("foo", MapOf(String, String))
			})
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("foo:Location")
				})
			})
		})
	})
}

var missingHeaderErrorAttributeDSL = func() {
	Service("MissingHeaderErrorAttribute", func() {
		Method("Method", func() {
			Error("error")
			HTTP(func() {
				POST("/")
				Response("error", func() {
					Header("bar")
				})
			})
		})
	})
}
