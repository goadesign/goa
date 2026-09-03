// This file verifies generated view declarations, validators, and converters.
package service

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
)

var updateViewGolden = flag.Bool("update-views", false, "update view code expectations")

func TestViews(t *testing.T) {
	cases := []struct {
		Name     string
		Constant string
		DSL      func()
		Code     string
	}{
		{"result-with-multiple-views", "ResultWithMultipleViewsCode", testdata.ResultWithMultipleViewsDSL, testdata.ResultWithMultipleViewsCode},
		{"result-collection-multiple-views", "ResultCollectionMultipleViewsCode", testdata.ResultCollectionMultipleViewsDSL, testdata.ResultCollectionMultipleViewsCode},
		{"result-with-user-type", "ResultWithUserTypeCode", testdata.ResultWithUserTypeDSL, testdata.ResultWithUserTypeCode},
		{"result-with-result-type", "ResultWithResultTypeCode", testdata.ResultWithResultTypeDSL, testdata.ResultWithResultTypeCode},
		{"result-with-recursive-result-type", "ResultWithRecursiveResultTypeCode", testdata.ResultWithRecursiveResultTypeDSL, testdata.ResultWithRecursiveResultTypeCode},
		{"result-type-with-custom-fields", "ResultWithCustomFieldsCode", testdata.ResultWithCustomFieldsDSL, testdata.ResultWithCustomFieldsCode},
		{"result-with-recursive-collection-of-result-type", "ResultWithRecursiveCollectionOfResultTypeCode", testdata.ResultWithRecursiveCollectionOfResultTypeDSL, testdata.ResultWithRecursiveCollectionOfResultTypeCode},
		{"result-with-multiple-methods", "ResultWithMultipleMethodsCode", testdata.ResultWithMultipleMethodsDSL, testdata.ResultWithMultipleMethodsCode},
		{"result-with-enum-type", "ResultWithEnumType", testdata.ResultWithEnumTypeDSL, testdata.ResultWithEnumType},
		{"result-with-pkg-path", "ResultWithPkgPathCode", testdata.ResultWithPkgPathDSL, testdata.ResultWithPkgPathCode},
		{"result-with-oneof-in-result-type", "ResultWithOneOfInResultTypeCode", testdata.ResultWithOneOfInResultTypeDSL, testdata.ResultWithOneOfInResultTypeCode},
	}
	updates := make(map[string]string, len(cases))
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			fs := viewsFile(plan, plan.facts.services[0])
			require.NotNil(t, fs)
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := string(bs)
			code = strings.ReplaceAll(code, "\r\n", "\n")
			if *updateViewGolden {
				updates[c.Constant] = code
				return
			}
			assert.Equal(t, c.Code, code)
		})
	}
	if *updateViewGolden {
		updateViewCodeExpectations(t, updates)
	}
}

// updateViewCodeExpectations replaces only the named string literals and
// leaves the surrounding test fixtures unchanged.
func updateViewCodeExpectations(t *testing.T, updates map[string]string) {
	t.Helper()
	path := "testdata/views_code.go"
	source, err := os.ReadFile(path)
	require.NoError(t, err)
	files := token.NewFileSet()
	parsed, err := parser.ParseFile(files, path, source, 0)
	require.NoError(t, err)
	type replacement struct {
		start int
		end   int
		value string
	}
	var replacements []replacement
	for _, declaration := range parsed.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.CONST {
			continue
		}
		for _, specification := range generic.Specs {
			value := specification.(*ast.ValueSpec)
			for index, name := range value.Names {
				updated, exists := updates[name.Name]
				if !exists {
					continue
				}
				literal := value.Values[index].(*ast.BasicLit)
				replacements = append(replacements, replacement{
					start: files.Position(literal.Pos()).Offset,
					end:   files.Position(literal.End()).Offset,
					value: viewCodeLiteral(updated),
				})
			}
		}
	}
	require.Len(t, replacements, len(updates))
	slices.SortFunc(replacements, func(left, right replacement) int {
		return right.start - left.start
	})
	for _, replacement := range replacements {
		source = append(source[:replacement.start], append([]byte(replacement.value), source[replacement.end:]...)...)
	}
	require.NoError(t, os.WriteFile(path, source, 0o644))
}

// viewCodeLiteral keeps readable raw strings unless generated Go tags require
// an interpreted string.
func viewCodeLiteral(source string) string {
	if !strings.Contains(source, "`") {
		return "`" + source + "`"
	}
	return strconv.Quote(source)
}
