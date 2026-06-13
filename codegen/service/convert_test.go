package service

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	inner struct {
		foo string
	}

	hasNonPtrFields struct {
		inner inner
	}
)

func TestCommonPath(t *testing.T) {
	cases := map[string]struct {
		Paths              []string
		ExpectedCommonPath string
	}{
		"common-path-exist": {
			Paths: []string{
				"/home/user1/tmp/coverage/test",
				"/home/user1/tmp/covert/operator",
				"/home/user1/tmp/coven/members",
				"/home//user1/tmp/coventry",
				"/home/user1/././tmp/covertly/foo",
				"/home/bob/../user1/tmp/coved/bar",
			},
			ExpectedCommonPath: "/home/user1/tmp",
		},
		"common-path-does-not-exist": {
			Paths: []string{
				"/home1/user1/tmp/coverage/test",
				"/home/user1/tmp/covert/operator",
			},
			ExpectedCommonPath: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cp := commonPath('/', tc.Paths...)
			assert.Equal(t, tc.ExpectedCommonPath, cp)
		})
	}
}

func TestPkgImport(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	cwd := gopath + "/src/goa.design/goa/codegen/service"
	goModCwd := "/home/user/project/goa/codegen/service"
	cases := []struct {
		Name           string
		Cwd            string
		Pkg            string
		ExpectedImport string
	}{
		{"goa-pkg", cwd, "goa.design/goa/v3/pkg", "goa.design/goa/v3/pkg"},
		{"internal-pkg", cwd, "goa.design/goa/v3/codegen", "goa.design/goa/v3/codegen"},
		{"vendored-pkg", cwd, "goa.design/goa/vendor/github.com/some/pkg", "github.com/some/pkg"},
		{"external-pkg", cwd, "github.com/some/pkg", "github.com/some/pkg"},
		{"gomod-goa-pkg", goModCwd, "goa.design/goa/v3/pkg", "goa.design/goa/v3/pkg"},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			pkgImport := getPkgImport(tc.Pkg, tc.Cwd)
			assert.Equal(t, tc.ExpectedImport, pkgImport)
		})
	}
}

func TestDesignType(t *testing.T) {
	var f bool
	cases := []struct {
		Name         string
		From         any
		ExpectedType expr.DataType
		ExpectedErr  string
	}{
		{"bool", false, expr.Boolean, ""},
		{"int", 0, expr.Int, ""},
		{"int32", int32(0), expr.Int32, ""},
		{"int64", int64(0), expr.Int64, ""},
		{"uint", uint(0), expr.UInt, ""},
		{"uint32", uint32(0), expr.UInt32, ""},
		{"uint64", uint64(0), expr.UInt64, ""},
		{"float32", float32(0.0), expr.Float32, ""},
		{"float64", 0.0, expr.Float64, ""},
		{"string", "", expr.String, ""},
		{"bytes", []byte{}, expr.Bytes, ""},
		{"array", []string{}, dsl.ArrayOf(expr.String), ""},
		{"map", map[string]string{}, dsl.MapOf(expr.String, expr.String), ""},
		{"object", objT{}, obj, ""},
		{"array-object", []objT{{}}, dsl.ArrayOf(obj), ""},

		{"invalid-bool", &f, nil, "*(<value>): only pointer to struct can be converted"},
		{"invalid-array", []*bool{&f}, nil, "*(<value>[0]): only pointer to struct can be converted"},
		{"invalid-map-key", map[*bool]string{&f: ""}, nil, "*(<value>.key): only pointer to struct can be converted"},
		{"invalid-map-val", map[string]*bool{"": &f}, nil, "*(<value>.value): only pointer to struct can be converted"},
		{"invalid-struct", hasNonPtrFields{inner: inner{foo: "foo"}}, nil, "<value>.inner: fields of type struct must use pointers"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var dt expr.DataType
			err := buildDesignType(&dt, reflect.TypeOf(c.From), nil)

			if c.ExpectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, c.ExpectedType, dt)
			}

			if c.ExpectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, c.ExpectedErr, err.Error())
			}
		})
	}
}
func TestCompatible(t *testing.T) {
	cases := []struct {
		Name        string
		From        expr.DataType
		To          any
		ExpectedErr string
	}{
		{"bool", expr.Boolean, false, ""},
		{"int", expr.Int, 0, ""},
		{"int32", expr.Int32, int32(0), ""},
		{"int64", expr.Int64, int64(0), ""},
		{"uint", expr.UInt, uint(0), ""},
		{"uint32", expr.UInt32, uint32(0), ""},
		{"uint64", expr.UInt64, uint64(0), ""},
		{"float32", expr.Float32, float32(0.0), ""},
		{"float64", expr.Float64, 0.0, ""},
		{"string", expr.String, "", ""},
		{"bytes", expr.Bytes, []byte{}, ""},
		{"array", dsl.ArrayOf(expr.String), []string{}, ""},
		{"map", dsl.MapOf(expr.String, expr.String), map[string]string{}, ""},
		{"map-interface", dsl.MapOf(expr.String, expr.Any), map[string]any{}, ""},
		{"object", obj, objT{}, ""},
		{"object-mapped", objMapped, objT{}, ""},
		{"object-ignored", objIgnored, objT{}, ""},
		{"object-extra", objIgnored, objExtraT{}, ""},
		{"object-recursive", objRecursive(), objRecursiveT{}, ""},
		{"array-object", dsl.ArrayOf(obj), []objT{{}}, ""},

		{"invalid-primitive", expr.String, 0, "types don't match: type of <value> is int but type of corresponding attribute is string"},
		{"invalid-int", expr.Int, 0.0, "types don't match: type of <value> is float64 but type of corresponding attribute is int"},
		{"invalid-float32", expr.Float32, 0, "types don't match: type of <value> is int but type of corresponding attribute is float32"},
		{"invalid-array", dsl.ArrayOf(expr.String), []int{0}, "types don't match: type of <value>[0] is int but type of corresponding attribute is string"},
		{"invalid-map-key", dsl.MapOf(expr.String, expr.String), map[int]string{0: ""}, "types don't match: type of <value>.key is int but type of corresponding attribute is string"},
		{"invalid-map-val", dsl.MapOf(expr.String, expr.String), map[string]int{"": 0}, "types don't match: type of <value>.value is int but type of corresponding attribute is string"},
		{"invalid-obj", obj, "", "types don't match: <value> is a string, expected a struct"},
		{"invalid-obj-2", obj, objT2{}, "types don't match: type of <value>.Bar is string but type of corresponding attribute is int"},
		{"invalid-obj-3", obj, objT3{}, "types don't match: type of <value>.Goo is int but type of corresponding attribute is float32"},
		{"invalid-obj-4", obj, objT4{}, "types don't match: type of <value>.Goo2 is float32 but type of corresponding attribute is uint"},
		{"invalid-obj-5", obj, objT5{}, "types don't match: could not find field \"Baz\" of external type \"objT5\" matching attribute \"Baz\" of type \"objT\""},
		{"invalid-array-object", dsl.ArrayOf(obj), []objT2{{}}, "types don't match: type of <value>[0].Bar is string but type of corresponding attribute is int"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := compatible(c.From, reflect.TypeOf(c.To))
			if c.ExpectedErr == "" {
				assert.NoError(t, err)
			}
			if c.ExpectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, c.ExpectedErr, err.Error())
			}
		})
	}
}

// Test fixtures
var obj = &expr.UserTypeExpr{
	AttributeExpr: &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "Bar", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			{Name: "Baz", Attribute: &expr.AttributeExpr{Type: expr.Boolean}},
			{Name: "Goo", Attribute: &expr.AttributeExpr{Type: expr.Float32}},
			{Name: "Goo2", Attribute: &expr.AttributeExpr{Type: expr.UInt}},
		},
		Validation: &expr.ValidationExpr{
			Required: []string{"Foo", "Bar", "Baz", "Goo", "Goo2"},
		},
	},
	TypeName: "objT",
	UID:      "goa.design/goa/v3/codegen/service#objT",
}

var objMapped = &expr.UserTypeExpr{
	AttributeExpr: &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "Bar", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			{Name: "mapped", Attribute: &expr.AttributeExpr{Type: expr.Boolean, Meta: expr.MetaExpr{"struct:field:external": []string{"Baz"}}}},
			{Name: "mapped (deprecated syntax)", Attribute: &expr.AttributeExpr{Type: expr.Boolean, Meta: expr.MetaExpr{"struct.field.external": []string{"Baz"}}}},
		},
	},
	TypeName: "objT",
}

var objIgnored = &expr.UserTypeExpr{
	AttributeExpr: &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "Bar", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			{Name: "ignored", Attribute: &expr.AttributeExpr{Type: expr.Boolean, Meta: expr.MetaExpr{"struct:field:external": []string{"-"}}}},
			{Name: "ignored (deprecated syntax)", Attribute: &expr.AttributeExpr{Type: expr.Boolean, Meta: expr.MetaExpr{"struct.field.external": []string{"-"}}}},
		},
	},
	TypeName: "objT",
}

func objRecursive() *expr.UserTypeExpr {
	res := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
				{Name: "Bar", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			},
		},
		TypeName: "objRecursiveT",
	}
	obj := res.Type.(*expr.Object)
	*obj = append(*obj, &expr.NamedAttributeExpr{
		Name:      "Rec",
		Attribute: &expr.AttributeExpr{Type: res}})

	return res
}

type objT struct {
	Foo  string
	Bar  int
	Baz  bool
	Goo  float32
	Goo2 uint
}

type objExtraT struct {
	Foo   string
	Bar   int
	Baz   bool
	Goo   float32
	Goo2  uint
	Extra time.Time
}

type objRecursiveT struct {
	Foo  string
	Bar  int
	Goo  float32
	Goo2 uint
	Rec  *objRecursiveT
}
type objT2 struct {
	Foo  string
	Bar  string
	Baz  bool
	Goo  float32
	Goo2 uint
}

type objT3 struct {
	Foo  string
	Bar  int
	Baz  bool
	Goo  int
	Goo2 uint
}

type objT4 struct {
	Foo  string
	Bar  int
	Baz  bool
	Goo  float32
	Goo2 float32
}

type objT5 struct {
	Foo  string
	Bar  int
	Goo  float32
	Goo2 uint
}

func TestConvertFiles(t *testing.T) {
	cases := []struct {
		Name          string
		DSL           func()
		ExpectedFiles map[string]int // path -> number of sections
	}{
		{
			"multi-package-conversions",
			testdata.ConvertMultiPkgDSL,
			map[string]int{
				"gen/types/convert.go":  3, // header + convert-to + create-from sections
				"gen/models/convert.go": 3, // header + convert-to + create-from sections
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := runDSL(t, c.DSL)
			services := NewServicesData(root)

			for _, svc := range root.Services {
				files, err := ConvertFiles(root, svc, services)
				require.NoError(t, err)

				// Check expected number of files
				require.Equal(t, len(c.ExpectedFiles), len(files))

				// Verify each expected file
				for expectedPath, expectedSections := range c.ExpectedFiles {
					found := false
					// Normalize expected path for cross-platform compatibility
					normalizedExpected := filepath.FromSlash(expectedPath)
					for _, file := range files {
						if strings.HasSuffix(file.Path, normalizedExpected) {
							found = true
							require.Equal(t, expectedSections, len(file.SectionTemplates))
							// First section should be header
							require.Equal(t, "source-header", file.SectionTemplates[0].Name)
							break
						}
					}
					require.True(t, found, "Expected file %s not found", expectedPath)
				}
			}
		})
	}
}
