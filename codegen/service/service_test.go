package service

import (
	"bytes"
	"fmt"
	"go/format"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service/testdata"
	"goa.design/goa/expr"
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
		{"payload-result-with-default", testdata.WithDefaultDSL, testdata.WithDefault},
		{"result-with-multiple-views", testdata.MultipleMethodsResultMultipleViewsDSL, testdata.MultipleMethodsResultMultipleViews},
		{"result-collection-multiple-views", testdata.ResultCollectionMultipleViewsMethodDSL, testdata.ResultCollectionMultipleViewsMethod},
		{"result-with-other-result", testdata.ResultWithOtherResultMethodDSL, testdata.ResultWithOtherResultMethod},
		{"service-level-error", testdata.ServiceErrorDSL, testdata.ServiceError},
		{"force-generate-type", testdata.ForceGenerateTypeDSL, testdata.ForceGenerateType},
		{"force-generate-type-explicit", testdata.ForceGenerateTypeExplicitDSL, testdata.ForceGenerateTypeExplicit},
		{"streaming-result", testdata.StreamingResultMethodDSL, testdata.StreamingResultMethod},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethod},
		{"streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewMethodDSL, testdata.StreamingResultWithExplicitViewMethod},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL, testdata.StreamingResultNoPayloadMethod},
		{"streaming-payload", testdata.StreamingPayloadMethodDSL, testdata.StreamingPayloadMethod},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethod},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethod},
		{"streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsMethodDSL, testdata.StreamingPayloadResultWithViewsMethod},
		{"streaming-payload-result-with-explict-view", testdata.StreamingPayloadResultWithExplicitViewMethodDSL, testdata.StreamingPayloadResultWithExplicitViewMethod},
		{"bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL, testdata.BidirectionalStreamingMethod},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethod},
		{"bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsMethodDSL, testdata.BidirectionalStreamingResultWithViewsMethod},
		{"bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewMethodDSL, testdata.BidirectionalStreamingResultWithExplicitViewMethod},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSLWithFunc(t, c.DSL, func() {
				expr.Root.Types = []expr.UserType{testdata.APayload, testdata.BPayload, testdata.AResult, testdata.BResult, testdata.ParentType, testdata.ChildType}
			})
			if len(expr.Root.Services) != 1 {
				t.Fatalf("got %d services, expected 1", len(expr.Root.Services))
			}
			fs := File("goa.design/goa/example", expr.Root.Services[0])
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
