package codegen

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/http/codegen/testing"
	httpdesign "goa.design/goa.v2/http/design"
)

func TestPaths(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"single-path-no-param", PathNoParamDSL, PathNoParamCode},
		{"single-path-one-param", PathOneParamDSL, PathOneParamCode},
		{"single-path-multiple-params", PathMultipleParamsDSL, PathMultipleParamsCode},
		{"alternative-paths", PathAlternativesDSL, PathAlternativesCode},
		{"path-with-string-slice-param", PathStringSliceParamDSL, PathStringSliceParamCode},
		{"path-with-int-slice-param", PathIntSliceParamDSL, PathIntSliceParamCode},
		{"path-with-int32-slice-param", PathInt32SliceParamDSL, PathInt32SliceParamCode},
		{"path-with-int64-slice-param", PathInt64SliceParamDSL, PathInt64SliceParamCode},
		{"path-with-uint-slice-param", PathUintSliceParamDSL, PathUintSliceParamCode},
		{"path-with-uint32-slice-param", PathUint32SliceParamDSL, PathUint32SliceParamCode},
		{"path-with-uint64-slice-param", PathUint64SliceParamDSL, PathUint64SliceParamCode},
		{"path-with-float33-slice-param", PathFloat32SliceParamDSL, PathFloat32SliceParamCode},
		{"path-with-float64-slice-param", PathFloat64SliceParamDSL, PathFloat64SliceParamCode},
		{"path-with-bool-slice-param", PathBoolSliceParamDSL, PathBoolSliceParamCode},
		{"path-with-interface-slice-param", PathInterfaceSliceParamDSL, PathInterfaceSliceParamCode},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			if len(httpdesign.Root.HTTPServices) != 1 {
				t.Fatalf("got %d file(s), expected 1", len(httpdesign.Root.HTTPServices))
			}
			fs := serverPath(httpdesign.Root.HTTPServices[0])
			sections := fs.Sections("")
			code := codegen.SectionCode(t, sections[1])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
