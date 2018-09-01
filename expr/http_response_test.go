package expr_test

import (
	"testing"

	. "goa.design/goa/dsl"
	"goa.design/goa/expr"
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
		{"object result", objectResultResponseWithHeadersDSL, ""},
		{"array result", arrayResultResponseWithHeadersDSL, ""},
		{"map result", mapResultResponseWithHeadersDSL, ""},
		{"invalid", emptyResultResponseWithHeadersDSL, `HTTP response of service "EmptyResultResponseWithHeaders" HTTP endpoint "Method": response defines headers but result is empty`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunHTTPDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidHTTPDSL(t, c.DSL)
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
