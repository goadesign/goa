package expr_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestHTTPResponseCookie(t *testing.T) {
	type Props map[string]any

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
		{"same-site", testdata.CookieSameSiteDSL, Props{"cookie:same-site": testdata.CookieSameSiteValue}},
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
					switch {
					case len(m) != 1:
						t.Errorf("got cookies metadata with length %d, expected 1", len(m))
					case len(m[n]) != 1:
						t.Errorf("got cookies metadata %q with length %d, expected 1", n, len(m[n]))
					case m[n][0] != fmt.Sprintf("%v", v):
						t.Errorf("got value %q for cookies metadata %q, expected %q", m[n][0], n, fmt.Sprintf("%v", v))
					}
				}
			}
		})
	}
}

func TestHTTPResponseCookieAttrBindings(t *testing.T) {
	root := expr.RunDSL(t, testdata.CookieAttrBindingsDSL)
	cookies := root.API.HTTP.Services[len(root.API.HTTP.Services)-1].HTTPEndpoints[0].Responses[0].Cookies
	obj := expr.AsObject(cookies.Type)
	if len(*obj) != 1 {
		t.Fatalf("got %d cookies, expected 1", len(*obj))
	}
	cookie := (*obj)[0].Attribute
	cases := map[string]string{
		"cookie:max-age:from":   "expiresIn",
		"cookie:domain:from":    "cookieDomain",
		"cookie:path:from":      "cookiePath",
		"cookie:secure:from":    "isSecure",
		"cookie:http-only:from": "isHTTPOnly",
		"cookie:same-site:from": "sameSite",
	}
	for k, want := range cases {
		got, ok := cookie.Meta[k]
		if !ok {
			t.Errorf("cookie metadata %q missing", k)
			continue
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("cookie metadata %q = %v, want [%q]", k, got, want)
		}
	}
}

func TestCookieSameSiteConstantsAreLowercase(t *testing.T) {
	cases := map[expr.CookieSameSiteValue]string{
		expr.CookieSameSiteStrict:  "strict",
		expr.CookieSameSiteLax:     "lax",
		expr.CookieSameSiteNone:    "none",
		expr.CookieSameSiteDefault: "default",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("CookieSameSite constant = %q, want %q (the SameSiteFrom binding contract and the http codegen partials key on these exact lower-case values)", string(got), want)
		}
	}
}

func TestHTTPResponseBodyExcludesCookieAttrBindings(t *testing.T) {
	root := expr.RunDSL(t, testdata.CookieAttrBindingsDSL)
	resp := root.API.HTTP.Services[len(root.API.HTTP.Services)-1].HTTPEndpoints[0].Responses[0]
	if resp.Body == nil {
		t.Fatalf("expected response body to be computed")
	}
	bound := []string{"expiresIn", "cookieDomain", "cookiePath", "isSecure", "isHTTPOnly", "sameSite"}
	if obj := expr.AsObject(resp.Body.Type); obj != nil {
		for _, nat := range *obj {
			for _, b := range bound {
				if nat.Name == b {
					t.Errorf("response body still contains cookie-bound attribute %q", b)
				}
			}
		}
	}
	cookieValue := "cookie"
	if obj := expr.AsObject(resp.Body.Type); obj != nil {
		for _, nat := range *obj {
			if nat.Name == cookieValue {
				t.Errorf("response body still contains cookie value attribute %q", cookieValue)
			}
		}
	}
}

func TestHTTPErrorCookieAttrBindings(t *testing.T) {
	root := expr.RunDSL(t, testdata.CookieAttrBindingErrorDSL)
	httpErr := root.API.HTTP.Services[len(root.API.HTTP.Services)-1].HTTPEndpoints[0].HTTPErrors[0]
	obj := expr.AsObject(httpErr.Response.Cookies.Type)
	if len(*obj) != 1 {
		t.Fatalf("got %d cookies, expected 1", len(*obj))
	}
	cookie := (*obj)[0].Attribute
	got, ok := cookie.Meta["cookie:max-age:from"]
	if !ok {
		t.Fatalf("cookie metadata %q missing", "cookie:max-age:from")
	}
	if len(got) != 1 || got[0] != "retryAfter" {
		t.Errorf("cookie metadata %q = %v, want [%q]", "cookie:max-age:from", got, "retryAfter")
	}
}

func TestHTTPResponseCookieAttrBindingValidation(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Want string
	}{
		{
			"missing-attr",
			testdata.CookieAttrBindingMissingAttrDSL,
			"binds Max-Age to attribute \"doesNotExist\"",
		},
		{
			"wrong-type",
			testdata.CookieAttrBindingWrongTypeDSL,
			"binds Max-Age to attribute \"expiresIn\" but it must be an integer",
		},
		{
			"undeclared-cookie",
			testdata.CookieAttrBindingUndeclaredDSL,
			"CookieAttributes references cookie \"notDeclared\"",
		},
		{
			"error-missing-attr",
			testdata.CookieAttrBindingErrorMissingAttrDSL,
			"binds Max-Age to attribute \"doesNotExist\" which has no equivalent attribute in error type",
		},
		{
			"error-wrong-type",
			testdata.CookieAttrBindingErrorWrongTypeDSL,
			"binds Max-Age to attribute \"retryAfter\" but it must be an integer",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, c.DSL)
			if err == nil {
				t.Fatalf("expected validation error containing %q", c.Want)
			}
			var msg string
			var verr *eval.ValidationErrors
			if errors.As(err, &verr) {
				msgs := make([]string, len(verr.Errors))
				for i, e := range verr.Errors {
					msgs[i] = e.Error()
				}
				msg = strings.Join(msgs, "\n")
			} else {
				msg = err.Error()
			}
			if !strings.Contains(msg, c.Want) {
				t.Fatalf("expected error to contain %q, got: %s", c.Want, msg)
			}
		})
	}
}
