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
		Name           string
		Type           expr.UserType
		Required       bool
		Pointer        bool
		DefaultPointer bool
		UseDefault     bool
		Code           string
	}{
		{"integer-required", integerT, true, false, false, false, testdata.IntegerRequiredValidationCode},
		{"integer-pointer", integerT, false, true, false, false, testdata.IntegerPointerValidationCode},
		{"integer-use-default", integerT, false, false, false, true, testdata.IntegerUseDefaultValidationCode},
		{"integer-default-pointer", integerT, false, false, true, true, testdata.IntegerDefaultPointerValidationCode},
		{"float-required", floatT, true, false, false, false, testdata.FloatRequiredValidationCode},
		{"float-pointer", floatT, false, true, false, false, testdata.FloatPointerValidationCode},
		{"float-use-default", floatT, false, false, false, true, testdata.FloatUseDefaultValidationCode},
		{"string-required", stringT, true, false, false, false, testdata.StringRequiredValidationCode},
		{"string-pointer", stringT, false, true, false, false, testdata.StringPointerValidationCode},
		{"string-use-default", stringT, false, false, false, true, testdata.StringUseDefaultValidationCode},
		{"user-type-required", userT, true, false, false, false, testdata.UserTypeRequiredValidationCode},
		{"user-type-pointer", userT, false, true, false, false, testdata.UserTypePointerValidationCode},
		{"user-type-default", userT, false, false, false, true, testdata.UserTypeUseDefaultValidationCode},
		{"array-required", arrayT, true, false, false, false, testdata.ArrayRequiredValidationCode},
		{"array-pointer", arrayT, false, true, false, false, testdata.ArrayPointerValidationCode},
		{"array-use-default", arrayT, false, false, false, true, testdata.ArrayUseDefaultValidationCode},
		{"map-required", mapT, true, false, false, false, testdata.MapRequiredValidationCode},
		{"map-pointer", mapT, false, true, false, false, testdata.MapPointerValidationCode},
		{"map-use-default", mapT, false, false, false, true, testdata.MapUseDefaultValidationCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			att := &expr.AttributeExpr{Type: c.Type}
			an := NewAttributeAnalyzer(att, c.Required, c.Pointer, c.DefaultPointer, c.UseDefault, "", scope)
			code := RecursiveValidationCode(an, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}
