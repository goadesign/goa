package service

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service/testdata"
	"goa.design/goa/design"
)

func TestClient(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"single", testdata.SingleEndpointDSL, testdata.SingleMethodClient},
		{"multiple", testdata.MultipleEndpointsDSL, testdata.MultipleMethodsClient},
		{"no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadMethodsClient},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			if len(design.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(design.Root.Services))
			}
			fs := ClientFile(design.Root.Services[0])
			if fs == nil {
				t.Fatalf("got nil file, expected not nil")
			}
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				if err := s.Write(buf); err != nil {
					t.Fatal(err)
				}
			}
			code := buf.String()
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs expected\n:%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
