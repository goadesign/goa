package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			require.Len(t, fs, 2)
			sections := fs[0].Section("client-endpoint-init")
			if len(sections) == 0 {
				t.Fatalf("got zero sections, expected at least one")
			}
			code := codegen.SectionsCode(t, sections)
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestRequestEncoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"request-encoder-payload-user-type", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.PayloadUserTypeRequestEncoderCode},
		{"request-encoder-payload-array", testdata.UnaryRPCNoResultDSL, testdata.PayloadArrayRequestEncoderCode},
		{"request-encoder-payload-map", testdata.MessageMapDSL, testdata.PayloadMapRequestEncoderCode},
		{"request-encoder-payload-primitive", testdata.ServerStreamingRPCDSL, testdata.PayloadPrimitiveRequestEncoderCode},
		{"request-encoder-payload-primitive-with-streaming-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.PayloadPrimitiveWithStreamingPayloadRequestEncoderCode},
		{"request-encoder-payload-user-type-with-streaming-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.PayloadUserTypeWithStreamingPayloadRequestEncoderCode},
		{"request-encoder-payload-with-metadata", testdata.MessageWithMetadataDSL, testdata.PayloadWithMetadataRequestEncoderCode},
		{"request-encoder-payload-with-validate", testdata.MessageWithValidateDSL, testdata.PayloadWithValidateRequestEncoderCode},
		{"request-encoder-payload-with-security-attributes", testdata.MessageWithSecurityAttrsDSL, testdata.PayloadWithSecurityAttrsRequestEncoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].Section("request-encoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestResponseDecoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"response-decoder-result-with-views", testdata.MessageResultTypeWithViewsDSL, testdata.ResultWithViewsResponseDecoderCode},
		{"response-decoder-result-with-explicit-view", testdata.MessageResultTypeWithExplicitViewDSL, testdata.ResultWithExplicitViewResponseDecoderCode},
		{"response-decoder-result-array", testdata.MessageArrayDSL, testdata.ResultArrayResponseDecoderCode},
		{"response-decoder-result-primitive", testdata.UnaryRPCNoPayloadDSL, testdata.ResultPrimitiveResponseDecoderCode},
		{"response-decoder-result-with-metadata", testdata.MessageWithMetadataDSL, testdata.ResultWithMetadataResponseDecoderCode},
		{"response-decoder-result-with-validate", testdata.MessageWithValidateDSL, testdata.ResultWithValidateResponseDecoderCode},
		{"response-decoder-result-collection", testdata.MessageResultTypeCollectionDSL, testdata.ResultCollectionResponseDecoderCode},
		{"response-decoder-server-streaming", testdata.ServerStreamingUserTypeDSL, testdata.ServerStreamingResponseDecoderCode},
		{"response-decoder-server-streaming-result-with-views", testdata.ServerStreamingResultWithViewsDSL, testdata.ServerStreamingResultWithViewsResponseDecoderCode},
		{"response-decoder-client-streaming", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingResponseDecoderCode},
		{"response-decoder-bidirectional-streaming", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingResponseDecoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].Section("response-decoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			assert.Equal(t, c.Code, code)
		})
	}
}
