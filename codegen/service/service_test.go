package service

import (
	"bytes"
	"go/format"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
)

func TestService(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL, testdata.NamesWithSpaces},
		{"service-single", testdata.SingleMethodDSL, testdata.SingleMethod},
		{"service-multiple", testdata.MultipleMethodsDSL, testdata.MultipleMethods},
		{"service-union", testdata.UnionMethodDSL, testdata.UnionMethod},
		{"service-multi-union", testdata.MultiUnionMethodDSL, testdata.MultiUnionMethod},
		{"service-no-payload-no-result", testdata.EmptyMethodDSL, testdata.EmptyMethod},
		{"service-payload-no-result", testdata.EmptyResultMethodDSL, testdata.EmptyResultMethod},
		{"service-no-payload-result", testdata.EmptyPayloadMethodDSL, testdata.EmptyPayloadMethod},
		{"service-payload-result-with-default", testdata.WithDefaultDSL, testdata.WithDefault},
		{"service-result-with-multiple-views", testdata.MultipleMethodsResultMultipleViewsDSL, testdata.MultipleMethodsResultMultipleViews},
		{"service-result-with-explicit-and-default-views", testdata.WithExplicitAndDefaultViewsDSL, testdata.WithExplicitAndDefaultViews},
		{"service-result-collection-multiple-views", testdata.ResultCollectionMultipleViewsMethodDSL, testdata.ResultCollectionMultipleViewsMethod},
		{"service-result-with-other-result", testdata.ResultWithOtherResultMethodDSL, testdata.ResultWithOtherResultMethod},
		{"service-result-with-result-collection", testdata.ResultWithResultCollectionMethodDSL, testdata.ResultWithResultCollectionMethod},
		{"service-result-with-dashed-mime-type", testdata.ResultWithDashedMimeTypeMethodDSL, testdata.ResultWithDashedMimeTypeMethod},
		{"service-result-with-one-of-type", testdata.ResultWithOneOfTypeMethodDSL, testdata.ResultWithOneOfTypeMethod},
		{"service-result-with-inline-validation", testdata.ResultWithInlineValidationDSL, testdata.ResultWithInlineValidation},
		{"service-service-level-error", testdata.ServiceErrorDSL, testdata.ServiceError},
		{"service-custom-errors", testdata.CustomErrorsDSL, testdata.CustomErrors},
		{"service-custom-errors-custom-field", testdata.CustomErrorsCustomFieldDSL, testdata.CustomErrorsCustomField},
		{"service-force-generate-type", testdata.ForceGenerateTypeDSL, testdata.ForceGenerateType},
		{"service-force-generate-type-explicit", testdata.ForceGenerateTypeExplicitDSL, testdata.ForceGenerateTypeExplicit},
		{"service-streaming-result", testdata.StreamingResultMethodDSL, testdata.StreamingResultMethod},
		{"service-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethod},
		{"service-streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewMethodDSL, testdata.StreamingResultWithExplicitViewMethod},
		{"service-streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL, testdata.StreamingResultNoPayloadMethod},
		{"service-streaming-payload", testdata.StreamingPayloadMethodDSL, testdata.StreamingPayloadMethod},
		{"service-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethod},
		{"service-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethod},
		{"service-streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsMethodDSL, testdata.StreamingPayloadResultWithViewsMethod},
		{"service-streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewMethodDSL, testdata.StreamingPayloadResultWithExplicitViewMethod},
		{"service-bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL, testdata.BidirectionalStreamingMethod},
		{"service-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethod},
		{"service-bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsMethodDSL, testdata.BidirectionalStreamingResultWithViewsMethod},
		{"service-bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewMethodDSL, testdata.BidirectionalStreamingResultWithExplicitViewMethod},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			require.Len(t, expr.Root.Services, 1)
			files := Files("goa.design/goa/example", expr.Root.Services[0], make(map[string][]string))
			require.Greater(t, len(files), 0)
			validateFile(t, files[0], files[0].Path, c.Code)
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
		SvcCodes  []string
		TypeFiles []string
		TypeCodes []string
	}{
		{"none", testdata.SingleMethodDSL, []string{testdata.SingleMethod}, nil, nil},
		{"single", testdata.PkgPathDSL, []string{testdata.PkgPath}, []string{fooPath}, []string{testdata.PkgPathFoo}},
		{"array", testdata.PkgPathArrayDSL, []string{testdata.PkgPathArray}, []string{fooPath}, []string{testdata.PkgPathArrayFoo}},
		{"recursive", testdata.PkgPathRecursiveDSL, []string{testdata.PkgPathRecursive}, []string{fooPath, recursiveFooPath}, []string{testdata.PkgPathRecursiveFooFoo, testdata.PkgPathRecursiveFoo}},
		{"multiple", testdata.PkgPathMultipleDSL, []string{testdata.PkgPathMultiple}, []string{barPath, bazPath}, []string{testdata.PkgPathBar, testdata.PkgPathBaz}},
		{"nopkg", testdata.PkgPathNoDirDSL, []string{testdata.PkgPathNoDir}, nil, nil},
		{"dupes", testdata.PkgPathDupeDSL, []string{testdata.PkgPathDupe1, testdata.PkgPathDupe2}, []string{fooPath}, []string{testdata.PkgPathFooDupe}},
		{"payload_attribute", testdata.PkgPathPayloadAttributeDSL, []string{testdata.PkgPathPayloadAttribute}, []string{fooPath}, []string{testdata.PkgPathPayloadAttributeFoo}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			userTypePkgs := make(map[string][]string)
			codegen.RunDSL(t, c.DSL)
			if len(expr.Root.Services) != len(c.SvcCodes) {
				t.Fatalf("got %d services, expected %d", len(expr.Root.Services), len(c.SvcCodes))
			}
			files := Files("goa.design/goa/example", expr.Root.Services[0], userTypePkgs)
			if len(files) != len(c.TypeFiles)+1 {
				t.Fatalf("got %d files, expected %d", len(files), len(c.TypeFiles)+1)
			}
			validateFile(t, files[0], files[0].Path, c.SvcCodes[0])
			for i, f := range c.TypeFiles {
				validateFile(t, files[i+1], f, c.TypeCodes[i])
			}
			if len(c.SvcCodes) > 1 {
				files = Files("goa.design/goa/example", expr.Root.Services[1], userTypePkgs)
				require.Len(t, files, 1)
				validateFile(t, files[0], files[0].Path, c.SvcCodes[1])
			}
		})
	}
}

func validateFile(t *testing.T, f *codegen.File, path, code string) {
	if f.Path != path {
		t.Errorf("got %q, expected %q", f.Path, path)
	}
	buf := new(bytes.Buffer)
	for _, s := range f.SectionTemplates[1:] {
		require.NoError(t, s.Write(buf))
	}
	bs, err := format.Source(buf.Bytes())
	require.NoError(t, err, buf.String())
	actual := string(bs)
	actual = strings.ReplaceAll(actual, "\r\n", "\n")
	assert.Equal(t, code, actual)
}
