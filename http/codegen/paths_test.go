package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestPaths(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"single-path-no-param", testdata.PathNoParamDSL},
		{"single-path-one-param", testdata.PathOneParamDSL},
		{"single-path-multiple-params", testdata.PathMultipleParamsDSL},
		{"alternative-paths", testdata.PathAlternativesDSL},
		{"path-with-string-slice-param", testdata.PathStringSliceParamDSL},
		{"path-with-int-slice-param", testdata.PathIntSliceParamDSL},
		{"path-with-int32-slice-param", testdata.PathInt32SliceParamDSL},
		{"path-with-int64-slice-param", testdata.PathInt64SliceParamDSL},
		{"path-with-uint-slice-param", testdata.PathUintSliceParamDSL},
		{"path-with-uint32-slice-param", testdata.PathUint32SliceParamDSL},
		{"path-with-uint64-slice-param", testdata.PathUint64SliceParamDSL},
		{"path-with-float33-slice-param", testdata.PathFloat32SliceParamDSL},
		{"path-with-float64-slice-param", testdata.PathFloat64SliceParamDSL},
		{"path-with-bool-slice-param", testdata.PathBoolSliceParamDSL},
		{"path-with-interface-slice-param", testdata.PathInterfaceSliceParamDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			require.Len(t, root.API.HTTP.Services, 1)
			services := CreateHTTPServices(root)
			fs := serverPath(root.API.HTTP.Services[0], services)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[1])
			testutil.AssertGo(t, "testdata/golden/paths_"+c.Name+".go.golden", code)
		})
	}
}

func TestPathTrailingShash(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"slash_with_base_path_no_trailing", testdata.BasePathNoTrailing_SlashWithBasePathNoTrailingDSL},
		{"trailing_with_base_path_no_trailing", testdata.BasePathNoTrailing_TrailingWithBasePathNoTrailingDSL},
		{"slash_with_base_path_with_trailing", testdata.BasePathWithTrailingSlash_WithBasePathWithTrailingDSL},
		{"slash_no_base_path", testdata.NoBasePath_SlashNoBasePathDSL},
		{"path-trailing_no_base_path", testdata.NoBasePath_TrailingNoBasePathDSL},
		{"add-trailing-slash-to-base-path", testdata.BasePath_SpecialTrailingSlashDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			require.Len(t, root.API.HTTP.Services, 1)
			services := CreateHTTPServices(root)
			fs := serverPath(root.API.HTTP.Services[0], services)
			sections := fs.SectionTemplates
			code := codegen.SectionCode(t, sections[1])
			testutil.AssertGo(t, "testdata/golden/paths_"+c.Name+".go.golden", code)
		})
	}
}
