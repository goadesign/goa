package service

import (
	"bytes"
	"go/format"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/codegen/testutil"
)

func TestService(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL},
		{"service-single", testdata.SingleMethodDSL},
		{"service-multiple", testdata.MultipleMethodsDSL},
		{"service-union", testdata.UnionMethodDSL},
		{"service-multi-union", testdata.MultiUnionMethodDSL},
		{"service-union-alias-cross-pkg", testdata.UnionWithAliasCrossPkgDSL},
		{"service-no-payload-no-result", testdata.EmptyMethodDSL},
		{"service-payload-no-result", testdata.EmptyResultMethodDSL},
		{"service-no-payload-result", testdata.EmptyPayloadMethodDSL},
		{"service-payload-result-with-default", testdata.WithDefaultDSL},
		{"service-result-with-multiple-views", testdata.MultipleMethodsResultMultipleViewsDSL},
		{"service-result-with-explicit-and-default-views", testdata.WithExplicitAndDefaultViewsDSL},
		{"service-result-collection-multiple-views", testdata.ResultCollectionMultipleViewsMethodDSL},
		{"service-result-with-other-result", testdata.ResultWithOtherResultMethodDSL},
		{"service-result-with-result-collection", testdata.ResultWithResultCollectionMethodDSL},
		{"service-result-with-dashed-mime-type", testdata.ResultWithDashedMimeTypeMethodDSL},
		{"service-result-with-one-of-type", testdata.ResultWithOneOfTypeMethodDSL},
		{"service-result-with-inline-validation", testdata.ResultWithInlineValidationDSL},
		{"service-service-level-error", testdata.ServiceErrorDSL},
		{"service-custom-errors", testdata.CustomErrorsDSL},
		{"service-custom-errors-custom-field", testdata.CustomErrorsCustomFieldDSL},
		{"service-force-generate-type", testdata.ForceGenerateTypeDSL},
		{"service-force-generate-type-explicit", testdata.ForceGenerateTypeExplicitDSL},
		{"service-streaming-result", testdata.StreamingResultMethodDSL},
		{"service-mixed-results", testdata.MixedResultsEndpointDSL},
		{"service-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL},
		{"service-streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewMethodDSL},
		{"service-streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL},
		{"service-streaming-payload", testdata.StreamingPayloadMethodDSL},
		{"service-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL},
		{"service-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL},
		{"service-streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsMethodDSL},
		{"service-streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewMethodDSL},
		{"service-bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL},
		{"service-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL},
		{"service-bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsMethodDSL},
		{"service-bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewMethodDSL},
		{"service-multiple-api-key-security", testdata.MultipleAPIKeySecurityDSL},
		{"service-mixed-and-multiple-api-key-security", testdata.MixedAndMultipleAPIKeySecurityDSL},
		{"service-bearer-security", testdata.BearerSecurityDSL},
		{"service-raw-object-payload-type-name-collision", testdata.RawObjectPayloadTypeNameCollisionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			services := NewServicesData(root)
			require.Len(t, root.Services, 1)
			files := Files("goa.design/goa/example", root.Services[0], services, make(map[string][]string))
			require.Greater(t, len(files), 0)

			// Generate the code
			buf := new(bytes.Buffer)
			for _, s := range files[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := string(bs)

			// Compare with golden file
			testutil.AssertGo(t, "testdata/golden/service_"+c.Name+".go.golden", code)
		})
	}
}

func TestStructPkgPath(t *testing.T) {
	fooPath := filepath.Join("gen", "foo", "foo.go")
	recursiveFooPath := filepath.Join("gen", "foo", "recursive_foo.go")
	barPath := filepath.Join("gen", "bar", "bar.go")
	bazPath := filepath.Join("gen", "baz", "baz.go")
	cases := []struct {
		Name      string
		DSL       func()
		TypeFiles []string
	}{
		{"none", testdata.SingleMethodDSL, nil},
		{"single", testdata.PkgPathDSL, []string{fooPath}},
		{"array", testdata.PkgPathArrayDSL, []string{fooPath}},
		{"recursive", testdata.PkgPathRecursiveDSL, []string{fooPath, recursiveFooPath}},
		{"multiple", testdata.PkgPathMultipleDSL, []string{barPath, bazPath}},
		{"nopkg", testdata.PkgPathNoDirDSL, nil},
		{"dupes", testdata.PkgPathDupeDSL, []string{fooPath}},
		{"payload_attribute", testdata.PkgPathPayloadAttributeDSL, []string{fooPath}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			userTypePkgs := make(map[string][]string)
			root := codegen.RunDSL(t, c.DSL)
			services := NewServicesData(root)
			files := Files("goa.design/goa/example", root.Services[0], services, userTypePkgs)

			// Check file count
			expectedFiles := len(c.TypeFiles) + 1
			require.Len(t, files, expectedFiles, "unexpected number of files")

			// First file is always the service file
			buf := new(bytes.Buffer)
			for _, s := range files[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err)
			testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_service.go.golden", string(bs))

			// Type files
			for i, typeFile := range c.TypeFiles {
				buf := new(bytes.Buffer)
				for _, s := range files[i+1].SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err)
				goldenName := filepath.Base(typeFile)
				testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_"+goldenName+".golden", string(bs))
			}

			// For dupes case, test the second service
			if c.Name == "dupes" && len(root.Services) > 1 {
				files = Files("goa.design/goa/example", root.Services[1], services, userTypePkgs)
				require.Len(t, files, 1)
				buf := new(bytes.Buffer)
				for _, s := range files[0].SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err)
				testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_service2.go.golden", string(bs))
			}
		})
	}
}

func TestStructPkgPath_UnionImportsJSON(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionDSL)
	services := NewServicesData(root)
	require.Len(t, root.Services, 1)

	files := Files("goa.design/goa/example", root.Services[0], services, make(map[string][]string))
	require.GreaterOrEqual(t, len(files), 2, "expected at least service.go + one struct:pkg:path file")

	var typeFile *codegen.File
	for _, f := range files {
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "type_with_union.go")) {
			typeFile = f
			break
		}
	}
	require.NotNil(t, typeFile, "expected generated type file for struct:pkg:path type_with_union")

	buf := new(bytes.Buffer)
	for _, s := range typeFile.SectionTemplates {
		require.NoError(t, s.Write(buf))
	}
	code := buf.String()
	require.Contains(t, code, "\"encoding/json\"", "expected encoding/json import in generated file:\n%s", code)
}

func TestStructPkgPath_UnionJSONFieldBranchesGenerateAliases(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionJSONFieldDSL)
	services := NewServicesData(root)
	require.Len(t, root.Services, 1)

	files := Files("goa.design/goa/example", root.Services[0], services, make(map[string][]string))
	require.GreaterOrEqual(t, len(files), 2, "expected at least service.go + one struct:pkg:path file")

	var typeFile *codegen.File
	for _, f := range files {
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "type_with_json_field_union.go")) {
			typeFile = f
			break
		}
	}
	require.NotNil(t, typeFile, "expected generated type file for struct:pkg:path type_with_json_field_union")

	buf := new(bytes.Buffer)
	for _, s := range typeFile.SectionTemplates {
		require.NoError(t, s.Write(buf))
	}
	code := buf.String()
	require.Contains(t, code, "A ValuesA", "expected union field A to use generated alias type:\n%s", code)
	require.Contains(t, code, "B ValuesB", "expected union field B to use generated alias type:\n%s", code)

	var hasValuesAFile, hasValuesBFile bool
	for _, f := range files {
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "values_a.go")) {
			hasValuesAFile = true
		}
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "values_b.go")) {
			hasValuesBFile = true
		}
	}
	require.True(t, hasValuesAFile, "expected generated alias file in struct:pkg:path package: gen/types/values_a.go")
	require.True(t, hasValuesBFile, "expected generated alias file in struct:pkg:path package: gen/types/values_b.go")
}
