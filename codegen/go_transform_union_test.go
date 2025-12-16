package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

func TestGoTransformUnion(t *testing.T) {
	root := RunDSL(t, testdata.TestUnionDSL)
	var (
		scope = NewNameScope()

		// types to test
		unionString     = root.UserType("Container").Attribute().Find("UnionString").Find("UnionString")
		unionString2    = root.UserType("Container").Attribute().Find("UnionString2").Find("UnionString2")
		unionStringInt  = root.UserType("Container").Attribute().Find("UnionStringInt").Find("UnionStringInt")
		unionStringInt2 = root.UserType("Container").Attribute().Find("UnionStringInt2").Find("UnionStringInt2")
		defaultCtx      = NewAttributeContext(false, false, true, "", scope)
	)
	tc := []struct {
		Name   string
		Source *expr.AttributeExpr
		Target *expr.AttributeExpr
	}{
		{"UnionString to UnionString2", unionString, unionString2},
		{"UnionStringInt to UnionStringInt2", unionStringInt, unionStringInt2},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			code, _, err := GoTransform(c.Source, c.Target, "source", "target", defaultCtx, defaultCtx, "", true)
			require.NoError(t, err)
			code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
			testutil.AssertGo(t, "testdata/golden/go_transform_union_"+c.Name+".go.golden", code)
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
