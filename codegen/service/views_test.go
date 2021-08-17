package service

import (
	"bytes"
	"fmt"
	"go/format"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
)

func TestViews(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"result-with-multiple-views", testdata.ResultWithMultipleViewsDSL, testdata.ResultWithMultipleViewsCode},
		{"result-collection-multiple-views", testdata.ResultCollectionMultipleViewsDSL, testdata.ResultCollectionMultipleViewsCode},
		{"result-with-user-type", testdata.ResultWithUserTypeDSL, testdata.ResultWithUserTypeCode},
		{"result-with-result-type", testdata.ResultWithResultTypeDSL, testdata.ResultWithResultTypeCode},
		{"result-with-recursive-result-type", testdata.ResultWithRecursiveResultTypeDSL, testdata.ResultWithRecursiveResultTypeCode},
		{"result-type-with-custom-fields", testdata.ResultWithCustomFieldsDSL, testdata.ResultWithCustomFieldsCode},
		{"result-with-recursive-collection-of-result-type", testdata.ResultWithRecursiveCollectionOfResultTypeDSL, testdata.ResultWithRecursiveCollectionOfResultTypeCode},
		{"result-with-multiple-methods", testdata.ResultWithMultipleMethodsDSL, testdata.ResultWithMultipleMethodsCode},
		{"result-with-enum-type", testdata.ResultWithEnumTypeDSL, testdata.ResultWithEnumType},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			if len(expr.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(expr.Root.Services))
			}
			fs := ViewsFile("goa.design/goa/example", expr.Root.Services[0])
			if fs == nil {
				t.Fatalf("got nil file, expected not nil")
			}
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				if err := s.Write(buf); err != nil {
					t.Fatal(err)
				}
			}
			bs, err := format.Source(buf.Bytes())
			if err != nil {
				fmt.Println(buf.String())
				t.Fatal(err)
			}
			code := string(bs)
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
