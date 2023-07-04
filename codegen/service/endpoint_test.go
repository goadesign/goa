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

func TestEndpoint(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"endpoint-single", testdata.SingleEndpointDSL, testdata.SingleEndpoint},
		{"endpoint-use", testdata.UseEndpointDSL, testdata.UseEndpoint},
		{"endpoint-multiple", testdata.MultipleEndpointsDSL, testdata.MultipleEndpoints},
		{"endpoint-no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadEndpoint},
		{"endpoint-with-result", testdata.WithResultEndpointDSL, testdata.WithResultEndpoint},
		{"endpoint-with-result-multiple-views", testdata.WithResultMultipleViewsEndpointDSL, testdata.WithResultMultipleViewsEndpoint},
		{"endpoint-streaming-result", testdata.StreamingResultEndpointDSL, testdata.StreamingResultMethodEndpoint},
		{"endpoint-streaming-result-no-payload", testdata.StreamingResultNoPayloadEndpointDSL, testdata.StreamingResultNoPayloadMethodEndpoint},
		{"endpoint-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethodEndpoint},
		{"endpoint-streaming-payload", testdata.StreamingPayloadEndpointDSL, testdata.StreamingPayloadMethodEndpoint},
		{"endpoint-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodEndpoint},
		{"endpoint-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodEndpoint},
		{"endpoint-bidirectional-streaming", testdata.BidirectionalStreamingEndpointDSL, testdata.BidirectionalStreamingMethodEndpoint},
		{"endpoint-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodEndpoint},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			expr.Root.GeneratedTypes = &expr.GeneratedRoot{}
			if len(expr.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(expr.Root.Services))
			}
			fs := EndpointFile("goa.design/goa/example", expr.Root.Services[0])
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
