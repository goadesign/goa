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

func TestClient(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"single", testdata.SingleEndpointDSL, testdata.SingleMethodClient},
		{"use", testdata.UseEndpointDSL, testdata.UseMethodClient},
		{"multiple", testdata.MultipleEndpointsDSL, testdata.MultipleMethodsClient},
		{"no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadMethodsClient},
		{"with-result", testdata.WithResultEndpointDSL, testdata.WithResultMethodClient},
		{"streaming-result", testdata.StreamingResultMethodDSL, testdata.StreamingResultMethodClient},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL, testdata.StreamingResultNoPayloadMethodClient},
		{"streaming-payload", testdata.StreamingPayloadMethodDSL, testdata.StreamingPayloadMethodClient},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodClient},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodClient},
		{"bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL, testdata.BidirectionalStreamingMethodClient},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodClient},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			if len(expr.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(expr.Root.Services))
			}
			fs := ClientFile("test/gen", expr.Root.Services[0])
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
				t.Errorf("%s: got\n%s\ngot vs expected\n:%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
