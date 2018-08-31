package expressions_test

import (
	"testing"

	. "goa.design/goa/expr"
	. "goa.design/goa/dsl"
)

func TestHTTPRouteValidation(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"valid", validRouteDSL, ""},
		{"invalid", duplicateWCRouteDSL, `route POST "/{id}" of service "InvalidRoute" HTTP endpoint "Method": Wildcard "id" appears multiple times in full path "/{id}/{id}"`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				RunHTTPDSL(t, c.DSL)
			} else {
				err := RunInvalidHTTPDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

var validRouteDSL = func() {
	Service("ValidRoute", func() {
		HTTP(func() {
			Path("/{base_id}")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("base_id", String)
				Attribute("id", String)
			})
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}

var duplicateWCRouteDSL = func() {
	Service("InvalidRoute", func() {
		HTTP(func() {
			Path("/{id}")
		})
		Method("Method", func() {
			Payload(func() {
				Attribute("id", String)
			})
			HTTP(func() {
				POST("/{id}")
			})
		})
	})
}
