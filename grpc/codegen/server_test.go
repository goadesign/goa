package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestServerGRPCInterface(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL},
		{"unary-rpc-with-errors", testdata.UnaryRPCWithErrorsDSL},
		{"unary-rpc-with-overriding-errors", testdata.UnaryRPCWithOverridingErrorsDSL},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
		{"bidirectional-streaming-rpc-with-errors", testdata.BidirectionalStreamingRPCWithErrorsDSL},
		{"client-streaming-rpc-with-payload-legacy-compat", testdata.ClientStreamingRPCWithPayloadLegacyCompatDSL},
		{"bidirectional-streaming-rpc-with-payload-legacy-compat", testdata.BidirectionalStreamingRPCWithPayloadLegacyCompatDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := serverFiles(services)
			require.Len(t, fs, 2)
			sections := fs[0].Section("server-grpc-interface")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/server_grpc_interface_"+c.Name+".go.golden", code)
		})
	}
}

func TestServerHandlerInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := serverFiles(services)
			require.Len(t, fs, 2)
			sections := fs[0].Section("grpc-handler-init")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/server_handler_init_"+c.Name+".go.golden", code)
		})
	}
}

func TestRequestDecoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"request-decoder-payload-user-type", testdata.MessageUserTypeWithNestedUserTypesDSL},
		{"request-decoder-payload-array", testdata.UnaryRPCNoResultDSL},
		{"request-decoder-payload-map", testdata.MessageMapDSL},
		{"request-decoder-payload-primitive", testdata.ServerStreamingRPCDSL},
		{"request-decoder-payload-primitive-with-streaming-payload", testdata.ClientStreamingRPCWithPayloadDSL},
		{"request-decoder-payload-user-type-with-streaming-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
		{"request-decoder-metadata-only-payload-with-streaming-payload", testdata.ClientStreamingRPCWithMetadataOnlyPayloadDSL},
		{"request-decoder-payload-primitive-with-streaming-payload-legacy-compat", testdata.ClientStreamingRPCWithPayloadLegacyCompatDSL},
		{"request-decoder-payload-user-type-with-streaming-payload-legacy-compat", testdata.BidirectionalStreamingRPCWithPayloadLegacyCompatDSL},
		{"request-decoder-payload-with-metadata-with-streaming-payload-legacy-compat", testdata.BidirectionalStreamingRPCWithMetadataLegacyCompatDSL},
		{"request-decoder-payload-with-metadata", testdata.MessageWithMetadataDSL},
		{"request-decoder-payload-with-validate", testdata.MessageWithValidateDSL},
		{"request-decoder-payload-with-security-attributes", testdata.MessageWithSecurityAttrsDSL},
		{"request-decoder-named-security-metadata", testdata.NamedSecurityMetadataDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := serverFiles(services)
			require.Len(t, fs, 2)
			sections := fs[1].Section("request-decoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/request_decoder_"+c.Name+".go.golden", code)
		})
	}
}

func TestResponseEncoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"response-encoder-empty-result", testdata.UnaryRPCNoResultDSL},
		{"response-encoder-result-with-views", testdata.MessageResultTypeWithViewsDSL},
		{"response-encoder-result-with-explicit-view", testdata.MessageResultTypeWithExplicitViewDSL},
		{"response-encoder-result-array", testdata.MessageArrayDSL},
		{"response-encoder-result-primitive", testdata.UnaryRPCNoPayloadDSL},
		{"response-encoder-result-with-metadata", testdata.MessageWithMetadataDSL},
		{"response-encoder-result-with-validate", testdata.MessageWithValidateDSL},
		{"response-encoder-result-collection", testdata.MessageResultTypeCollectionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := serverFiles(services)
			require.Len(t, fs, 2)
			sections := fs[1].Section("response-encoder")
			require.NotEmpty(t, sections)
			code := codegen.SectionsCode(t, sections)
			testutil.AssertGo(t, "testdata/golden/response_encoder_"+c.Name+".go.golden", code)
		})
	}
}
