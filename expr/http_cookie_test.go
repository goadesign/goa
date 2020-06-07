package expr_test

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestHTTPResponseCookie(t *testing.T) {
	type Props map[string]interface{}

	cases := []struct {
		Name  string
		DSL   func()
		Props Props
	}{
		{"cookie", testdata.CookieObjectResultDSL, nil},
		{"cookie", testdata.CookieStringResultDSL, nil},
		{"max-age", testdata.CookieMaxAgeDSL, Props{"cookie:max-age": testdata.CookieMaxAgeValue}},
		{"domain", testdata.CookieDomainDSL, Props{"cookie:domain": testdata.CookieDomainValue}},
		{"path", testdata.CookiePathDSL, Props{"cookie:path": testdata.CookiePathValue}},
		{"secure", testdata.CookieSecureDSL, Props{"cookie:secure": "Secure"}},
		{"http-only", testdata.CookieHTTPOnlyDSL, Props{"cookie:http-only": "HttpOnly"}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			e := root.API.HTTP.Services[len(root.API.HTTP.Services)-1].HTTPEndpoints[0]
			cookies := e.Responses[0].Cookies.AttributeExpr
			if len(*expr.AsObject(cookies.Type)) != 1 {
				t.Errorf("got %d cookie(s), expected exactly one", len(*expr.AsObject(cookies.Type)))
			} else {
				m := cookies.Meta
				for n, v := range c.Props {
					if len(m) != 1 {
						t.Errorf("got cookies metadata with length %d, expected 1", len(m))
					} else if len(m[n]) != 1 {
						t.Errorf("got cookies metadata %q with length %d, expected 1", n, len(m[n]))
					} else if m[n][0] != fmt.Sprintf("%v", v) {
						t.Errorf("got value %q for cookies metadata %q, expected %q", m[n][0], n, fmt.Sprintf("%v", v))
					}
				}
			}
		})
	}
}
