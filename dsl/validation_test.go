package dsl_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
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

func TestRequired(t *testing.T) {
	att := &expr.AttributeExpr{
		Type: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{
				Type: &expr.Object{
					{"foo", &expr.AttributeExpr{Type: String}},
					{"bar", &expr.AttributeExpr{Type: String}},
				},
			},
			TypeName: "Foo",
		},
	}
	eval.Context = &eval.DSLContext{}
	eval.Execute(func() { Required("foo") }, att)
	if eval.Context.Errors != nil {
		t.Errorf("Required failed unexpectedly with %s", eval.Context.Errors)
		return
	}
	if len(att.Validation.Required) == 0 {
		t.Errorf("Required not set on %+v", att)
		return
	}
	if att.Validation.Required[0] != "foo" {
		t.Errorf("Required invalid on %+v, expected foo, got %+v", att, att.Validation.Required)
	}
	uattr := att.Type.(*expr.UserTypeExpr).AttributeExpr
	if uattr.Validation == nil {
		t.Fatalf("Required not set on %+v", uattr)
	}
	if len(uattr.Validation.Required) == 0 {
		t.Errorf("Required not set on %+v, got %+v", uattr, uattr.Validation.Required)
	}
	if uattr.Validation.Required[0] != "foo" {
		t.Errorf("Required invalid on %+v, expected foo, got %+v", uattr, uattr.Validation.Required)
	}
}
