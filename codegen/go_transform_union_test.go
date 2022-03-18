package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/expr"
)

func TestGoTransformUnion(t *testing.T) {
	root := RunDSL(t, testdata.TestUnionDSL)
	var (
		scope = NewNameScope()

		// types to test
		unionString    = root.UserType("Container").Attribute().Find("UnionString").Find("UnionString")
		unionString2   = root.UserType("Container").Attribute().Find("UnionString2").Find("UnionString2")
		unionStringInt = root.UserType("Container").Attribute().Find("UnionStringInt").Find("UnionStringInt")
		unionSomeType  = root.UserType("Container").Attribute().Find("UnionSomeType").Find("UnionSomeType")
		userType       = &expr.AttributeExpr{Type: root.UserType("UnionUserType")}
		defaultCtx     = NewAttributeContext(false, false, true, "", scope)
	)
	tc := []struct {
		Name     string
		Source   *expr.AttributeExpr
		Target   *expr.AttributeExpr
		Expected string
	}{
		{"UnionString to UnionString2", unionString, unionString2, unionToUnionCode},

		{"UnionString to User Type", unionString, userType, unionStringToUserTypeCode},
		{"UnionStringInt to User Type", unionStringInt, userType, unionStringIntToUserTypeCode},
		{"UnionSomeType to User Type", unionSomeType, userType, unionSomeTypeToUserTypeCode},

		{"User Type to UnionString", userType, unionString, userTypeToUnionStringCode},
		{"User Type to UnionStringInt", userType, unionStringInt, userTypeToUnionStringIntCode},
		{"User Type to UnionSomeType", userType, unionSomeType, userTypeToUnionSomeTypeCode},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			code, _, err := GoTransform(c.Source, c.Target, "source", "target", defaultCtx, defaultCtx, "", true)
			if err != nil {
				t.Errorf("unexpected error %s", err)
				return
			}
			code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
			if code != c.Expected {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Expected))
			}
		})
	}
}

func TestGoTransformUnionError(t *testing.T) {
	root := RunDSL(t, testdata.TestUnionDSL)
	var (
		scope = NewNameScope()

		// types to test
		unionString    = root.UserType("Container").Attribute().Find("UnionString").Find("UnionString")
		unionStringInt = root.UserType("Container").Attribute().Find("UnionStringInt").Find("UnionStringInt")
		unionSomeType  = root.UserType("Container").Attribute().Find("UnionSomeType").Find("UnionSomeType")
		defaultCtx     = NewAttributeContext(false, false, true, "", scope)
	)
	tc := []struct {
		Name   string
		Source *expr.AttributeExpr
		Target *expr.AttributeExpr
		Error  string
	}{
		{"UnionString to UnionStringInt", unionString, unionStringInt, "cannot transform union: number of union types differ (UnionString has 1, UnionStringInt has 2)"},
		{"UnionString to UnionSomeType", unionString, unionSomeType, "cannot transform union UnionString to UnionSomeType: type at index 0: source is a string but target type is object"},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			_, _, err := GoTransform(c.Source, c.Target, "source", "target", defaultCtx, defaultCtx, "", true)
			if err == nil {
				t.Errorf("unexpected success")
				return
			}
			if err.Error() != c.Error {
				t.Errorf("unexpected error, got: %s, expected: %s", err, c.Error)
			}
		})
	}
}

const unionToUnionCode = `func transform() {
	var target *UnionString2
	switch actual := source.(type) {
	case UnionStringString:
		val := UnionString2String(actual)
		target = UnionString2{Value: val}
	}
}
`

const unionStringToUserTypeCode = `func transform() {
	var target *UnionUserType
	var name string
	switch source.(type) {
	case UnionStringString:
		name = "UnionStringString"
	}
	target = &UnionUserType{
		Type:  name,
		Value: source,
	}
}
`

const unionStringIntToUserTypeCode = `func transform() {
	var target *UnionUserType
	var name string
	switch source.(type) {
	case UnionStringIntString:
		name = "UnionStringIntString"
	case UnionStringIntInt:
		name = "UnionStringIntInt"
	}
	target = &UnionUserType{
		Type:  name,
		Value: source,
	}
}
`

const unionSomeTypeToUserTypeCode = `func transform() {
	var target *UnionUserType
	var name string
	switch source.(type) {
	case *SomeType:
		name = "SomeType"
	}
	target = &UnionUserType{
		Type:  name,
		Value: source,
	}
}
`

const userTypeToUnionStringCode = `func transform() {
	var target *UnionString
	switch source.Type {
	case "UnionStringString":
		target = source.Value.(UnionStringString)
	}
}
`

const userTypeToUnionStringIntCode = `func transform() {
	var target *UnionStringInt
	switch source.Type {
	case "UnionStringIntString":
		target = source.Value.(UnionStringIntString)
	case "UnionStringIntInt":
		target = source.Value.(UnionStringIntInt)
	}
}
`

const userTypeToUnionSomeTypeCode = `func transform() {
	var target *UnionSomeType
	switch source.Type {
	case "SomeType":
		target = source.Value.(*SomeType)
	}
}
`
