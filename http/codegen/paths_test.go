package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
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
			require.Len(t, expr.Root.API.HTTP.Services, 1)
			fs := serverPath(expr.Root.API.HTTP.Services[0])
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[1])
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestPathTrailingShash(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"slash_with_base_path_no_trailing", testdata.BasePathNoTrailing_SlashWithBasePathNoTrailingDSL, testdata.BasePathNoTrailing_SlashWithBasePathNoTrailingCode},
		{"trailing_with_base_path_no_trailing", testdata.BasePathNoTrailing_TrailingWithBasePathNoTrailingDSL, testdata.BasePathNoTrailing_TrailingWithBasePathNoTrailingCode},
		{"slash_with_base_path_with_trailing", testdata.BasePathWithTrailingSlash_WithBasePathWithTrailingDSL, testdata.BasePathWithTrailingSlash_WithBasePathWithTrailingCode},
		{"slash_no_base_path", testdata.NoBasePath_SlashNoBasePathDSL, testdata.NoBasePath_SlashNoBasePathCode},
		{"path-trailing_no_base_path", testdata.NoBasePath_TrailingNoBasePathDSL, testdata.NoBasePath_TrailingNoBasePathCode},
		{"add-trailing-slash-to-base-path", testdata.BasePath_SpecialTrailingSlashDSL, testdata.BasePath_SpecialTrailingSlashCode},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			require.Len(t, expr.Root.API.HTTP.Services, 1)
			fs := serverPath(expr.Root.API.HTTP.Services[0])
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[1])
			assert.Equal(t, c.Code, code)
		})
	}
}
