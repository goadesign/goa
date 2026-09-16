package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientEndpointInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL},
		{"unary-rpc-with-errors", testdata.UnaryRPCWithErrorsDSL},
		{"unary-rpc-acronym", testdata.UnaryRPCAcronymDSL},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL},
		{"client-streaming-rpc-no-result", testdata.ClientStreamingNoResultDSL},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
		{"bidirectional-streaming-rpc-with-errors", testdata.BidirectionalStreamingRPCWithErrorsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := clientFiles(services)
			require.Len(t, fs, 2)
			sections := fs[0].Section("client-endpoint-init")
			if len(sections) == 0 {
				t.Fatalf("got zero sections, expected at least one")
			}
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/client_endpoint_init_"+c.Name+".go.golden", code)
		})
	}
}

func TestRequestEncoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"request-encoder-payload-user-type", testdata.MessageUserTypeWithNestedUserTypesDSL},
		{"request-encoder-payload-array", testdata.UnaryRPCNoResultDSL},
		{"request-encoder-payload-map", testdata.MessageMapDSL},
		{"request-encoder-payload-primitive", testdata.ServerStreamingRPCDSL},
		{"request-encoder-payload-primitive-with-streaming-payload", testdata.ClientStreamingRPCWithPayloadDSL},
		{"request-encoder-payload-user-type-with-streaming-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
		{"request-encoder-payload-user-type-with-streaming-payload-legacy-compat", testdata.BidirectionalStreamingRPCWithPayloadLegacyCompatDSL},
		{"request-encoder-payload-with-metadata", testdata.MessageWithMetadataDSL},
		{"request-encoder-payload-with-validate", testdata.MessageWithValidateDSL},
		{"request-encoder-payload-with-security-attributes", testdata.MessageWithSecurityAttrsDSL},
		{"request-encoder-named-security-metadata", testdata.NamedSecurityMetadataDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := clientFiles(services)
			require.Len(t, fs, 2)
			sections := fs[1].Section("request-encoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/request_encoder_"+c.Name+".go.golden", code)
		})
	}
}

func TestResponseDecoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"response-decoder-result-with-views", testdata.MessageResultTypeWithViewsDSL},
		{"response-decoder-result-with-explicit-view", testdata.MessageResultTypeWithExplicitViewDSL},
		{"response-decoder-result-array", testdata.MessageArrayDSL},
		{"response-decoder-result-primitive", testdata.UnaryRPCNoPayloadDSL},
		{"response-decoder-result-with-metadata", testdata.MessageWithMetadataDSL},
		{"response-decoder-result-with-validate", testdata.MessageWithValidateDSL},
		{"response-decoder-result-collection", testdata.MessageResultTypeCollectionDSL},
		{"response-decoder-server-streaming", testdata.ServerStreamingUserTypeDSL},
		{"response-decoder-server-streaming-result-with-views", testdata.ServerStreamingResultWithViewsDSL},
		{"response-decoder-client-streaming", testdata.ClientStreamingRPCDSL},
		{"response-decoder-bidirectional-streaming", testdata.BidirectionalStreamingRPCDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := clientFiles(services)
			require.Len(t, fs, 2)
			sections := fs[1].Section("response-decoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/response_decoder_"+c.Name+".go.golden", code)
		})
	}
}
