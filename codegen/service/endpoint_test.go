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
		{"single", testdata.SingleEndpointDSL, testdata.SingleEndpoint},
		{"use", testdata.UseEndpointDSL, testdata.UseEndpoint},
		{"multiple", testdata.MultipleEndpointsDSL, testdata.MultipleEndpoints},
		{"no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadEndpoint},
		{"with-result", testdata.WithResultEndpointDSL, testdata.WithResultEndpoint},
		{"with-result-multiple-views", testdata.WithResultMultipleViewsEndpointDSL, testdata.WithResultMultipleViewsEndpoint},
		{"streaming-result", testdata.StreamingResultEndpointDSL, testdata.StreamingResultMethodEndpoint},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadEndpointDSL, testdata.StreamingResultNoPayloadMethodEndpoint},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethodEndpoint},
		{"streaming-payload", testdata.StreamingPayloadEndpointDSL, testdata.StreamingPayloadMethodEndpoint},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodEndpoint},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodEndpoint},
		{"bidirectional-streaming", testdata.BidirectionalStreamingEndpointDSL, testdata.BidirectionalStreamingMethodEndpoint},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodEndpoint},
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
