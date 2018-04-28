package service

import (
	"bytes"
	"fmt"
	"go/format"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service/testdata"
	"goa.design/goa/design"
)

func TestService(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"single", testdata.SingleMethodDSL, testdata.SingleMethod},
		{"multiple", testdata.MultipleMethodsDSL, testdata.MultipleMethods},
		{"no-payload-no-result", testdata.EmptyMethodDSL, testdata.EmptyMethod},
		{"payload-no-result", testdata.EmptyResultMethodDSL, testdata.EmptyResultMethod},
		{"no-payload-result", testdata.EmptyPayloadMethodDSL, testdata.EmptyPayloadMethod},
		{"result-with-multiple-views", testdata.MultipleMethodsResultMultipleViewsDSL, testdata.MultipleMethodsResultMultipleViews},
		{"result-with-other-result", testdata.ResultWithOtherResultMethodDSL, testdata.ResultWithOtherResultMethod},
		{"service-level-error", testdata.ServiceErrorDSL, testdata.ServiceError},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSLWithFunc(t, c.DSL, func() {
				design.Root.Types = []design.UserType{testdata.APayload, testdata.BPayload, testdata.AResult, testdata.BResult, testdata.ParentType, testdata.ChildType}
			})
			design.Root.GeneratedTypes = &design.GeneratedRoot{}
			if len(design.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(design.Root.Services))
			}
			fs := File("goa.design/goa/example", design.Root.Services[0])
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
