package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientEndpointInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL, testdata.UnaryRPCsClientEndpointInitCode},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL, testdata.UnaryRPCNoPayloadClientEndpointInitCode},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL, testdata.UnaryRPCNoResultClientEndpointInitCode},
		{"unary-rpc-with-errors", testdata.UnaryRPCWithErrorsDSL, testdata.UnaryRPCWithErrorsClientEndpointInitCode},
		{"unary-rpc-acronym", testdata.UnaryRPCAcronymDSL, testdata.UnaryRPCAcronymClientEndpointInitCode},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL, testdata.ServerStreamingRPCClientEndpointInitCode},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingRPCClientEndpointInitCode},
		{"client-streaming-rpc-no-result", testdata.ClientStreamingNoResultDSL, testdata.ClientStreamingNoResultClientEndpointInitCode},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.ClientStreamingRPCWithPayloadClientEndpointInitCode},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingRPCClientEndpointInitCode},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.BidirectionalStreamingRPCWithPayloadClientEndpointInitCode},
		{"bidirectional-streaming-rpc-with-errors", testdata.BidirectionalStreamingRPCWithErrorsDSL, testdata.BidirectionalStreamingRPCWithErrorsClientEndpointInitCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].Section("client-endpoint-init")
			if len(sections) == 0 {
				t.Fatalf("got zero sections, expected at least one")
			}
			code := codegen.SectionsCode(t, sections)
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

func TestRequestEncoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"payload-user-type", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.PayloadUserTypeRequestEncoderCode},
		{"payload-array", testdata.UnaryRPCNoResultDSL, testdata.PayloadArrayRequestEncoderCode},
		{"payload-map", testdata.MessageMapDSL, testdata.PayloadMapRequestEncoderCode},
		{"payload-primitive", testdata.ServerStreamingRPCDSL, testdata.PayloadPrimitiveRequestEncoderCode},
		{"payload-primitive-with-streaming-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.PayloadPrimitiveWithStreamingPayloadRequestEncoderCode},
		{"payload-user-type-with-streaming-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.PayloadUserTypeWithStreamingPayloadRequestEncoderCode},
		{"payload-with-metadata", testdata.MessageWithMetadataDSL, testdata.PayloadWithMetadataRequestEncoderCode},
		{"payload-with-validate", testdata.MessageWithValidateDSL, testdata.PayloadWithValidateRequestEncoderCode},
		{"payload-with-security-attributes", testdata.MessageWithSecurityAttrsDSL, testdata.PayloadWithSecurityAttrsRequestEncoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].Section("request-encoder")
			if len(sections) == 0 {
				t.Fatalf("got zero sections, expected at least one")
			}
			code := codegen.SectionsCode(t, sections)
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

func TestResponseDecoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"result-with-views", testdata.MessageResultTypeWithViewsDSL, testdata.ResultWithViewsResponseDecoderCode},
		{"result-with-explicit-view", testdata.MessageResultTypeWithExplicitViewDSL, testdata.ResultWithExplicitViewResponseDecoderCode},
		{"result-array", testdata.MessageArrayDSL, testdata.ResultArrayResponseDecoderCode},
		{"result-primitive", testdata.UnaryRPCNoPayloadDSL, testdata.ResultPrimitiveResponseDecoderCode},
		{"result-with-metadata", testdata.MessageWithMetadataDSL, testdata.ResultWithMetadataResponseDecoderCode},
		{"result-with-validate", testdata.MessageWithValidateDSL, testdata.ResultWithValidateResponseDecoderCode},
		{"result-collection", testdata.MessageResultTypeCollectionDSL, testdata.ResultCollectionResponseDecoderCode},
		{"server-streaming", testdata.ServerStreamingUserTypeDSL, testdata.ServerStreamingResponseDecoderCode},
		{"server-streaming-result-with-views", testdata.ServerStreamingResultWithViewsDSL, testdata.ServerStreamingResultWithViewsResponseDecoderCode},
		{"client-streaming", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingResponseDecoderCode},
		{"bidirectional-streaming", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingResponseDecoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].Section("response-decoder")
			if len(sections) == 0 {
				t.Fatalf("got zero sections, expected at least one")
			}
			code := codegen.SectionsCode(t, sections)
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
