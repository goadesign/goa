package dsl_test

import (
	"testing"

	"goa.design/goa/expr"
	. "goa.design/goa/dsl"
	"goa.design/goa/eval"
)

func TestFormat(t *testing.T) {
	cases := map[string]struct {
		Format expr.ValidationFormat
	}{
		"date":      {expr.FormatDate},
		"date-time": {expr.FormatDateTime},
		"uuid":      {expr.FormatUUID},
		"email":     {expr.FormatEmail},
		"hostname":  {expr.FormatHostname},
		"ipv4":      {expr.FormatIPv4},
		"ipv6":      {expr.FormatIPv6},
		"ip":        {expr.FormatIP},
		"uri":       {expr.FormatURI},
		"mac":       {expr.FormatMAC},
		"cidr":      {expr.FormatCIDR},
		"regexp":    {expr.FormatRegexp},
		"json":      {expr.FormatJSON},
		"rfc1123":   {expr.FormatRFC1123},
	}

	for k, tc := range cases {
		eval.Context = &eval.DSLContext{}
		expr := &expr.AttributeExpr{}
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
