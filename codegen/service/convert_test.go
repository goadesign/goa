package service

import (
	"os"
	"reflect"
	"testing"
	"time"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service/testdata"
	"goa.design/goa/design"
	"goa.design/goa/dsl"
	"goa.design/goa/eval"
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
			if cp != tc.ExpectedCommonPath {
				t.Errorf("got %s expected %s", cp, tc.ExpectedCommonPath)
			}
		})
	}
}

func TestPkgImport(t *testing.T) {
	cwd := os.Getenv("GOPATH") + "/src/goa.design/goa/codegen/service"
	goModCwd := "/home/user/project/goa/codegen/service"
	cases := []struct {
		Name           string
		Cwd            string
		Pkg            string
		ExpectedImport string
	}{
		{"root-pkg", cwd, "goa.design/goa", "goa.design/goa"},
		{"internal-pkg", cwd, "goa.design/goa/codegen", "goa.design/goa/codegen"},
		{"vendored-pkg", cwd, "goa.design/goa/vendor/github.com/some/pkg", "github.com/some/pkg"},
		{"external-pkg", cwd, "github.com/some/pkg", "github.com/some/pkg"},
		{"gomod-root-pkg", goModCwd, "goa.design/goa", "goa.design/goa"},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			pkgImport := getPkgImport(tc.Pkg, tc.Cwd)
			if pkgImport != tc.ExpectedImport {
				t.Errorf("got %s, expected %s", pkgImport, tc.ExpectedImport)
			}
		})
	}
}

func TestDesignType(t *testing.T) {
	var f bool
	cases := []struct {
		Name         string
		From         interface{}
		ExpectedType design.DataType
		ExpectedErr  string
	}{
		{"bool", false, design.Boolean, ""},
		{"int", 0, design.Int, ""},
		{"int32", int32(0), design.Int32, ""},
		{"int64", int64(0), design.Int64, ""},
		{"uint", uint(0), design.UInt, ""},
		{"uint32", uint32(0), design.UInt32, ""},
		{"uint64", uint64(0), design.UInt64, ""},
		{"float32", float32(0.0), design.Float32, ""},
		{"float64", 0.0, design.Float64, ""},
		{"string", "", design.String, ""},
		{"bytes", []byte{}, design.Bytes, ""},
		{"array", []string{}, dsl.ArrayOf(design.String), ""},
		{"map", map[string]string{}, dsl.MapOf(design.String, design.String), ""},
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
			var dt design.DataType
			err := buildDesignType(&dt, reflect.TypeOf(c.From), nil)

			// We didn't expect an error
			if c.ExpectedErr == "" {
				if err != nil {
					// but got one
					t.Errorf("got error %s, expected none", err)
				} else if !design.Equal(dt, c.ExpectedType) {
					t.Errorf("got %v expected %v", dt, c.ExpectedType)
				}
			}

			// We expected an error
			if c.ExpectedErr != "" {
				if err == nil {
					// but got none
					t.Errorf("got no error, expected %q", c.ExpectedErr)
				} else {
					if err.Error() != c.ExpectedErr {
						t.Errorf("got error %q, expected %q", err, c.ExpectedErr)
					}
				}
			}
		})
	}
}
func TestCompatible(t *testing.T) {
	cases := []struct {
		Name        string
		From        design.DataType
		To          interface{}
		ExpectedErr string
	}{
		{"bool", design.Boolean, false, ""},
		{"int", design.Int, 0, ""},
		{"int32", design.Int32, int32(0), ""},
		{"int64", design.Int64, int64(0), ""},
		{"uint", design.UInt, uint(0), ""},
		{"uint32", design.UInt32, uint32(0), ""},
		{"uint64", design.UInt64, uint64(0), ""},
		{"float32", design.Float32, float32(0.0), ""},
		{"float64", design.Float64, 0.0, ""},
		{"string", design.String, "", ""},
		{"bytes", design.Bytes, []byte{}, ""},
		{"array", dsl.ArrayOf(design.String), []string{}, ""},
		{"map", dsl.MapOf(design.String, design.String), map[string]string{}, ""},
		{"map-interface", dsl.MapOf(design.String, design.Any), map[string]interface{}{}, ""},
		{"object", obj, objT{}, ""},
		{"object-mapped", objMapped, objT{}, ""},
		{"object-ignored", objIgnored, objT{}, ""},
		{"object-extra", objIgnored, objExtraT{}, ""},
		{"object-recursive", objRecursive(), objRecursiveT{}, ""},
		{"array-object", dsl.ArrayOf(obj), []objT{{}}, ""},

		{"invalid-primitive", design.String, 0, "types don't match: type of <value> is int but type of corresponding attribute is string"},
		{"invalid-int", design.Int, 0.0, "types don't match: type of <value> is float64 but type of corresponding attribute is int"},
		{"invalid-float32", design.Float32, 0, "types don't match: type of <value> is int but type of corresponding attribute is float32"},
		{"invalid-array", dsl.ArrayOf(design.String), []int{0}, "types don't match: type of <value>[0] is int but type of corresponding attribute is string"},
		{"invalid-map-key", dsl.MapOf(design.String, design.String), map[int]string{0: ""}, "types don't match: type of <value>.key is int but type of corresponding attribute is string"},
		{"invalid-map-val", dsl.MapOf(design.String, design.String), map[string]int{"": 0}, "types don't match: type of <value>.value is int but type of corresponding attribute is string"},
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
			if err == nil {
				if c.ExpectedErr != "" {
					t.Errorf("got no error, expected %q", c.ExpectedErr)
				}
			} else {
				if c.ExpectedErr == "" {
					t.Errorf("got error %q, expected none", err)
				} else {
					if err.Error() != c.ExpectedErr {
						t.Errorf("got error %q, expected %q", err, c.ExpectedErr)
					}
				}
			}
		})
	}
}

func TestConvertFile(t *testing.T) {
	cases := []struct {
		Name         string
		DSL          func()
		SectionIndex int
		Code         string
	}{
		{"convert-string", testdata.ConvertStringDSL, 1, testdata.ConvertStringCode},
		{"convert-string-required", testdata.ConvertStringRequiredDSL, 1, testdata.ConvertStringRequiredCode},
		{"convert-string-pointer", testdata.ConvertStringPointerDSL, 1, testdata.ConvertStringPointerCode},
		{"convert-string-pointer-required", testdata.ConvertStringPointerRequiredDSL, 1, testdata.ConvertStringPointerRequiredCode},
		{"create-string", testdata.CreateStringDSL, 1, testdata.CreateStringCode},
		{"create-string-required", testdata.CreateStringRequiredDSL, 1, testdata.CreateStringRequiredCode},
		{"create-string-pointer", testdata.CreateStringPointerDSL, 1, testdata.CreateStringPointerCode},
		{"create-string-pointer-required", testdata.CreateStringPointerRequiredDSL, 1, testdata.CreateStringPointerRequiredCode},

		{"convert-external-name", testdata.ConvertExternalNameDSL, 1, testdata.ConvertExternalNameCode},
		{"convert-external-name-required", testdata.ConvertExternalNameRequiredDSL, 1, testdata.ConvertExternalNameRequiredCode},
		{"convert-external-name-pointer", testdata.ConvertExternalNamePointerDSL, 1, testdata.ConvertExternalNamePointerCode},
		{"convert-external-name-pointer-required", testdata.ConvertExternalNamePointerRequiredDSL, 1, testdata.ConvertExternalNamePointerRequiredCode},
		{"create-external-name", testdata.CreateExternalNameDSL, 1, testdata.CreateExternalNameCode},
		{"create-external-name-required", testdata.CreateExternalNameRequiredDSL, 1, testdata.CreateExternalNameRequiredCode},
		{"create-external-name-pointer", testdata.CreateExternalNamePointerDSL, 1, testdata.CreateExternalNamePointerCode},
		{"create-external-name-pointer-required", testdata.CreateExternalNamePointerRequiredDSL, 1, testdata.CreateExternalNamePointerRequiredCode},

		{"convert-array-string", testdata.ConvertArrayStringDSL, 1, testdata.ConvertArrayStringCode},
		{"convert-array-string-required", testdata.ConvertArrayStringRequiredDSL, 1, testdata.ConvertArrayStringRequiredCode},
		{"create-array-string", testdata.CreateArrayStringDSL, 1, testdata.CreateArrayStringCode},
		{"create-array-string-required", testdata.CreateArrayStringRequiredDSL, 1, testdata.CreateArrayStringRequiredCode},
		{"convert-object", testdata.ConvertObjectDSL, 1, testdata.ConvertObjectCode},
		{"convert-object-2", testdata.ConvertObjectDSL, 2, testdata.ConvertObjectHelperCode},
		{"convert-object-required", testdata.ConvertObjectRequiredDSL, 2, testdata.ConvertObjectRequiredHelperCode},
		{"convert-object-2-required", testdata.ConvertObjectRequiredDSL, 1, testdata.ConvertObjectRequiredCode},
		{"create-object", testdata.CreateObjectDSL, 1, testdata.CreateObjectCode},
		{"create-object-required", testdata.CreateObjectRequiredDSL, 1, testdata.CreateObjectRequiredCode},
		{"create-object-extra", testdata.CreateObjectExtraDSL, 1, testdata.CreateObjectExtraCode},
		{"create-external-convert", testdata.CreateExternalDSL, 0, testdata.CreateExternalConvert},
		{"create-alias-convert", testdata.CreateAliasDSL, 0, testdata.CreateAliasConvert},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := runDSL(t, c.DSL)
			for _, svc := range root.Services {
				f, err := ConvertFile(root, svc)
				if err != nil {
					t.Fatal(err)
				}
				if f == nil {
					t.Fatal("no file produced")
				}
				sections := f.SectionTemplates
				if len(sections) <= c.SectionIndex {
					t.Fatalf("got %d sections, expected at least %d", len(sections), c.SectionIndex+1)
				}

				var code string

				if c.SectionIndex == 0 {
					methodSection := sections[1]
					code = codegen.SectionCodeFromImportsAndMethods(t, sections[c.SectionIndex], methodSection)
				} else {
					code = codegen.SectionCode(t, sections[c.SectionIndex])
				}

				if code != c.Code {
					t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
				}
			}
		})
	}
}

// runDSL returns the DSL root resulting from running the given DSL.
func runDSL(t *testing.T, dsl func()) *design.RootExpr {
	// reset all roots and codegen data structures
	Services = make(ServicesData)
	eval.Reset()
	design.Root = new(design.RootExpr)
	eval.Register(design.Root)
	design.Root.API = &design.APIExpr{Name: "test api"}
	design.Root.API.Servers = []*design.ServerExpr{design.Root.API.DefaultServer()}

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}

	// return generated root
	return design.Root
}

// Test fixtures

var obj = &design.UserTypeExpr{
	AttributeExpr: &design.AttributeExpr{
		Type: &design.Object{
			{"Foo", &design.AttributeExpr{Type: design.String}},
			{"Bar", &design.AttributeExpr{Type: design.Int}},
			{"Baz", &design.AttributeExpr{Type: design.Boolean}},
			{"Goo", &design.AttributeExpr{Type: design.Float32}},
			{"Goo2", &design.AttributeExpr{Type: design.UInt}},
		},
	},
	TypeName: "objT",
}

var objMapped = &design.UserTypeExpr{
	AttributeExpr: &design.AttributeExpr{
		Type: &design.Object{
			{"Foo", &design.AttributeExpr{Type: design.String}},
			{"Bar", &design.AttributeExpr{Type: design.Int}},
			{"mapped", &design.AttributeExpr{Type: design.Boolean, Metadata: design.MetadataExpr{"struct.field.external": []string{"Baz"}}}},
		},
	},
	TypeName: "objT",
}

var objIgnored = &design.UserTypeExpr{
	AttributeExpr: &design.AttributeExpr{
		Type: &design.Object{
			{"Foo", &design.AttributeExpr{Type: design.String}},
			{"Bar", &design.AttributeExpr{Type: design.Int}},
			{"ignored", &design.AttributeExpr{Type: design.Boolean, Metadata: design.MetadataExpr{"struct.field.external": []string{"-"}}}},
		},
	},
	TypeName: "objT",
}

func objRecursive() *design.UserTypeExpr {
	res := &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{
			Type: &design.Object{
				{"Foo", &design.AttributeExpr{Type: design.String}},
				{"Bar", &design.AttributeExpr{Type: design.Int}},
			},
		},
		TypeName: "objRecursiveT",
	}
	obj := res.AttributeExpr.Type.(*design.Object)
	*obj = append(*obj, &design.NamedAttributeExpr{"Rec", &design.AttributeExpr{Type: res}})

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
