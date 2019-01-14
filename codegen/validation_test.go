package codegen

import (
	"testing"

	"goa.design/goa/codegen/testdata"
	"goa.design/goa/expr"
)

func TestRecursiveValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	var (
		scope = NewNameScope()

		integerT = root.UserType("Integer")
		stringT  = root.UserType("String")
		floatT   = root.UserType("Float")
		userT    = root.UserType("UserType")
		arrayT   = root.UserType("Array")
		mapT     = root.UserType("Map")
	)
	cases := []struct {
		Name       string
		Type       expr.UserType
		Required   bool
		Pointer    bool
		UseDefault bool
		Code       string
	}{
		{"integer-required", integerT, true, false, false, testdata.IntegerRequiredValidationCode},
		{"integer-pointer", integerT, false, true, false, testdata.IntegerPointerValidationCode},
		{"integer-use-default", integerT, false, false, true, testdata.IntegerUseDefaultValidationCode},
		{"float-required", floatT, true, false, false, testdata.FloatRequiredValidationCode},
		{"float-pointer", floatT, false, true, false, testdata.FloatPointerValidationCode},
		{"float-use-default", floatT, false, false, true, testdata.FloatUseDefaultValidationCode},
		{"string-required", stringT, true, false, false, testdata.StringRequiredValidationCode},
		{"string-pointer", stringT, false, true, false, testdata.StringPointerValidationCode},
		{"string-use-default", stringT, false, false, true, testdata.StringUseDefaultValidationCode},
		{"user-type-required", userT, true, false, false, testdata.UserTypeRequiredValidationCode},
		{"user-type-pointer", userT, false, true, false, testdata.UserTypePointerValidationCode},
		{"user-type-default", userT, false, false, true, testdata.UserTypeUseDefaultValidationCode},
		{"array-required", arrayT, true, false, false, testdata.ArrayRequiredValidationCode},
		{"array-pointer", arrayT, false, true, false, testdata.ArrayPointerValidationCode},
		{"array-use-default", arrayT, false, false, true, testdata.ArrayUseDefaultValidationCode},
		{"map-required", mapT, true, false, false, testdata.MapRequiredValidationCode},
		{"map-pointer", mapT, false, true, false, testdata.MapPointerValidationCode},
		{"map-use-default", mapT, false, false, true, testdata.MapUseDefaultValidationCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ca := &ContextualAttribute{
				Attribute:  NewGoAttribute(&expr.AttributeExpr{Type: c.Type}, "", scope),
				Required:   c.Required,
				Pointer:    c.Pointer,
				UseDefault: c.UseDefault}
			code := RecursiveValidationCode(ca, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}
