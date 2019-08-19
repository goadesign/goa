package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/expr"
)

func TestRecursiveValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	var (
		scope = NewNameScope()

		integerT = root.UserType("Integer")
		stringT  = root.UserType("String")
		floatT   = root.UserType("Float")
		userT    = root.UserType("UserType")
		arrayUT  = root.UserType("ArrayUserType")
		arrayT   = root.UserType("Array")
		mapT     = root.UserType("Map")
	)
	cases := []struct {
		Name         string
		Type         expr.UserType
		Required     bool
		Pointer      bool
		UseDefault   bool
		UseZeroValue bool
		Code         string
	}{
		{"integer-required", integerT, true, false, false, false, testdata.IntegerRequiredValidationCode},
		{"integer-pointer", integerT, false, true, false, false, testdata.IntegerPointerValidationCode},
		{"integer-use-default", integerT, false, false, true, false, testdata.IntegerUseDefaultValidationCode},
		{"integer-use-zero", integerT, false, false, false, true, testdata.IntegerUseZeroValueValidationCode},
		{"float-required", floatT, true, false, false, false, testdata.FloatRequiredValidationCode},
		{"float-pointer", floatT, false, true, false, false, testdata.FloatPointerValidationCode},
		{"float-use-default", floatT, false, false, true, false, testdata.FloatUseDefaultValidationCode},
		{"float-use-zero", floatT, false, false, false, true, testdata.FloatUseZeroValueValidationCode},
		{"string-required", stringT, true, false, false, false, testdata.StringRequiredValidationCode},
		{"string-pointer", stringT, false, true, false, false, testdata.StringPointerValidationCode},
		{"string-use-default", stringT, false, false, true, false, testdata.StringUseDefaultValidationCode},
		{"string-use-zero", stringT, false, false, false, true, testdata.StringUseZeroValueValidationCode},
		{"user-type-required", userT, true, false, false, false, testdata.UserTypeRequiredValidationCode},
		{"user-type-pointer", userT, false, true, false, false, testdata.UserTypePointerValidationCode},
		{"user-type-default", userT, false, false, true, false, testdata.UserTypeUseDefaultValidationCode},
		{"user-type-use-zero", userT, false, false, false, true, testdata.UserTypeUseZeroValueValidationCode},
		{"user-type-array-required", arrayUT, true, true, false, false, testdata.UserTypeArrayValidationCode},
		{"array-required", arrayT, true, false, false, false, testdata.ArrayRequiredValidationCode},
		{"array-pointer", arrayT, false, true, false, false, testdata.ArrayPointerValidationCode},
		{"array-use-default", arrayT, false, false, true, false, testdata.ArrayUseDefaultValidationCode},
		{"array-use-zero", arrayT, false, false, false, true, testdata.ArrayUseZeroValueValidationCode},
		{"map-required", mapT, true, false, false, false, testdata.MapRequiredValidationCode},
		{"map-pointer", mapT, false, true, false, false, testdata.MapPointerValidationCode},
		{"map-use-default", mapT, false, false, true, false, testdata.MapUseDefaultValidationCode},
		{"map-use-zero", mapT, false, false, false, true, testdata.MapUseZeroValueValidationCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := NewAttributeContext(c.Pointer, false, c.UseDefault, "", scope)
			code := RecursiveValidationCode(&expr.AttributeExpr{Type: c.Type}, ctx, c.Required, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}
