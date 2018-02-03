package dsl_test

import (
	"testing"

	"goa.design/goa/design"
	. "goa.design/goa/dsl"
	"goa.design/goa/eval"
)

func TestFormat(t *testing.T) {
	cases := map[string]struct {
		Format design.ValidationFormat
	}{
		"date-time": {design.FormatDateTime},
		"uuid":      {design.FormatUUID},
		"email":     {design.FormatEmail},
		"hostname":  {design.FormatHostname},
		"ipv4":      {design.FormatIPv4},
		"ipv6":      {design.FormatIPv6},
		"ip":        {design.FormatIP},
		"uri":       {design.FormatURI},
		"mac":       {design.FormatMAC},
		"cidr":      {design.FormatCIDR},
		"regexp":    {design.FormatRegexp},
		"json":      {design.FormatJSON},
		"rfc1123":   {design.FormatRFC1123},
	}

	for k, tc := range cases {
		eval.Context = &eval.DSLContext{}
		expr := &design.AttributeExpr{}
		eval.Execute(func() { Format(tc.Format) }, expr)
		if eval.Context.Errors != nil {
			t.Errorf("%s: Format failed unexpectedly with %s", k, eval.Context.Errors)
		}
		if expr.Validation == nil {
			t.Errorf("%s: Format not initialized Validation in %+v", k, expr)
		} else {
			if expr.Validation.Format != tc.Format {
				t.Errorf("%s: Format not set on %+v, expected %s, got %+v", k, expr, tc.Format, expr.Validation.Format)
			}
		}
	}
}
