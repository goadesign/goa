package codegen

import (
	"testing"

	"goa.design/goa/codegen/testdata"
	"goa.design/goa/expr"
)

func TestRecursiveValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	var (
		integerT = root.UserType("Integer")
		stringT  = root.UserType("String")
		floatT   = root.UserType("Float")
		userT    = root.UserType("UserType")
		arrayT   = root.UserType("Array")
		mapT     = root.UserType("Map")

		pointerP  = &expr.AttributeProperties{Pointer: true}
		requiredP = &expr.AttributeProperties{Required: true}
		defaultP  = &expr.AttributeProperties{UseDefault: true}
	)
	cases := []struct {
		Name       string
		Type       expr.UserType
		Properties *expr.AttributeProperties
		Code       string
	}{
		{"integer-required", integerT, requiredP, testdata.IntegerRequiredValidationCode},
		{"integer-pointer", integerT, pointerP, testdata.IntegerPointerValidationCode},
		{"integer-use-default", integerT, defaultP, testdata.IntegerUseDefaultValidationCode},
		{"float-required", floatT, requiredP, testdata.FloatRequiredValidationCode},
		{"float-pointer", floatT, pointerP, testdata.FloatPointerValidationCode},
		{"float-use-default", floatT, defaultP, testdata.FloatUseDefaultValidationCode},
		{"string-required", stringT, requiredP, testdata.StringRequiredValidationCode},
		{"string-pointer", stringT, pointerP, testdata.StringPointerValidationCode},
		{"string-use-default", stringT, defaultP, testdata.StringUseDefaultValidationCode},
		{"user-type-required", userT, requiredP, testdata.UserTypeRequiredValidationCode},
		{"user-type-pointer", userT, pointerP, testdata.UserTypePointerValidationCode},
		{"user-type-default", userT, defaultP, testdata.UserTypeUseDefaultValidationCode},
		{"array-required", arrayT, requiredP, testdata.ArrayRequiredValidationCode},
		{"array-pointer", arrayT, pointerP, testdata.ArrayPointerValidationCode},
		{"array-use-default", arrayT, defaultP, testdata.ArrayUseDefaultValidationCode},
		{"map-required", mapT, requiredP, testdata.MapRequiredValidationCode},
		{"map-pointer", mapT, pointerP, testdata.MapPointerValidationCode},
		{"map-use-default", mapT, defaultP, testdata.MapUseDefaultValidationCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			att := &expr.AttributeExpr{Type: c.Type}
			an := expr.NewAttributeAnalyzer(att, c.Properties)
			code := RecursiveValidationCode(an, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}
