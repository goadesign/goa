package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/grpc/codegen/testdata"
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
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL, testdata.ServerStreamingRPCClientEndpointInitCode},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingRPCClientEndpointInitCode},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.ClientStreamingRPCWithPayloadClientEndpointInitCode},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingRPCClientEndpointInitCode},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.BidirectionalStreamingRPCWithPayloadClientEndpointInitCode},
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
		{"result-with-views", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.ResultWithViewsResponseDecoderCode},
		{"result-array", testdata.MessageArrayDSL, testdata.ResultArrayResponseDecoderCode},
		{"result-primitive", testdata.UnaryRPCNoPayloadDSL, testdata.ResultPrimitiveResponseDecoderCode},
		{"result-with-metadata", testdata.MessageWithMetadataDSL, testdata.ResultWithMetadataResponseDecoderCode},
		{"result-collection", testdata.MessageResultTypeCollectionDSL, testdata.ResultCollectionResponseDecoderCode},
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
