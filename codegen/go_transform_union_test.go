package codegen

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/expr"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestGoTransformUnion(t *testing.T) {
	root := RunDSL(t, testdata.TestUnionDSL)
	var (
		scope = NewNameScope()

		// types to test
		unionString    = root.UserType("Container").Attribute().Find("UnionString")
		unionString2   = root.UserType("Container").Attribute().Find("UnionString2")
		unionStringInt = root.UserType("Container").Attribute().Find("UnionStringInt")
		unionSomeType  = root.UserType("Container").Attribute().Find("UnionSomeType")
		userType       = &expr.AttributeExpr{Type: root.UserType("UnionUserType")}
		defaultCtx     = NewAttributeContext(false, false, true, "", scope)
	)
	tc := []struct {
		Name   string
		Source *expr.AttributeExpr
		Target *expr.AttributeExpr
		Error  string
	}{
		{"UnionString to UnionString2", unionString, unionString2, ""},

		{"UnionString to User Type", unionString, userType, ""},
		{"UnionStringInt to User Type", unionStringInt, userType, ""},
		{"UnionSomeType to User Type", unionSomeType, userType, ""},

		{"User Type to UnionString", userType, unionString, ""},
		{"User Type to UnionStringInt", userType, unionStringInt, ""},
		{"User Type to UnionSomeType", userType, unionSomeType, ""},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			code, _, err := GoTransform(c.Source, c.Target, "source", "target", defaultCtx, defaultCtx, "", true)
			if err != nil {
				t.Errorf("unexpected error %s", err)
				return
			}
			code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
			path := filepath.Join("testdata", strings.Replace(c.Name, " ", "_", -1))
			if *update {
				if err := ioutil.WriteFile(path+".golden", []byte(code), 0644); err != nil {
					t.Error(err)
				}
				return
			}
			expected, err := ioutil.ReadFile(path + ".golden")
			if err != nil {
				t.Error(err)
				return
			}
			if code != string(expected) {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, string(expected)))
			}
		})
	}
}

func TestGoTransformUnionError(t *testing.T) {
	root := RunDSL(t, testdata.TestUnionDSL)
	var (
		scope = NewNameScope()

		// types to test
		unionString    = root.UserType("Container").Attribute().Find("UnionString")
		unionStringInt = root.UserType("Container").Attribute().Find("UnionStringInt")
		unionSomeType  = root.UserType("Container").Attribute().Find("UnionSomeType")
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
