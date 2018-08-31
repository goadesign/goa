package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestPaths(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"single-path-no-param", testdata.PathNoParamDSL, testdata.PathNoParamCode},
		{"single-path-one-param", testdata.PathOneParamDSL, testdata.PathOneParamCode},
		{"single-path-multiple-params", testdata.PathMultipleParamsDSL, testdata.PathMultipleParamsCode},
		{"alternative-paths", testdata.PathAlternativesDSL, testdata.PathAlternativesCode},
		{"path-with-string-slice-param", testdata.PathStringSliceParamDSL, testdata.PathStringSliceParamCode},
		{"path-with-int-slice-param", testdata.PathIntSliceParamDSL, testdata.PathIntSliceParamCode},
		{"path-with-int32-slice-param", testdata.PathInt32SliceParamDSL, testdata.PathInt32SliceParamCode},
		{"path-with-int64-slice-param", testdata.PathInt64SliceParamDSL, testdata.PathInt64SliceParamCode},
		{"path-with-uint-slice-param", testdata.PathUintSliceParamDSL, testdata.PathUintSliceParamCode},
		{"path-with-uint32-slice-param", testdata.PathUint32SliceParamDSL, testdata.PathUint32SliceParamCode},
		{"path-with-uint64-slice-param", testdata.PathUint64SliceParamDSL, testdata.PathUint64SliceParamCode},
		{"path-with-float33-slice-param", testdata.PathFloat32SliceParamDSL, testdata.PathFloat32SliceParamCode},
		{"path-with-float64-slice-param", testdata.PathFloat64SliceParamDSL, testdata.PathFloat64SliceParamCode},
		{"path-with-bool-slice-param", testdata.PathBoolSliceParamDSL, testdata.PathBoolSliceParamCode},
		{"path-with-interface-slice-param", testdata.PathInterfaceSliceParamDSL, testdata.PathInterfaceSliceParamCode},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			if len(expr.Root.HTTPServices) != 1 {
				t.Fatalf("got %d file(s), expected 1", len(expr.Root.HTTPServices))
			}
			fs := serverPath(expr.Root.HTTPServices[0])
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[1])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
