package design_test

import (
	"testing"

	"goa.design/goa/http/design"
	"goa.design/goa/http/design/testdata"
)

func TestResponseValidation(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"empty", testdata.EmptyResultEmptyResponseDSL, ""},
		{"non empty result", testdata.NonEmptyResultEmptyResponseDSL, ""},
		{"non empty response", testdata.EmptyResultNonEmptyResponseDSL, ""},
		// {"string result", testdata.StringResultResponseWithHeadersDSL, ""},
		{"object result", testdata.ObjectResultResponseWithHeadersDSL, ""},
		{"array result", testdata.ArrayResultResponseWithHeadersDSL, ""},
		{"map result", testdata.MapResultResponseWithHeadersDSL, ""},
		{"invalid", testdata.EmptyResultResponseWithHeadersDSL, `HTTP response of service "EmptyResultResponseWithHeaders" HTTP endpoint "Method": response defines headers but result is empty`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				design.RunHTTPDSL(t, c.DSL)
			} else {
				err := design.RunInvalidHTTPDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
				}
			}
		})
	}
}
