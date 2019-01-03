package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/grpc/codegen/testdata"
)

func TestServerGRPCInterface(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL, testdata.UnaryRPCsServerInterfaceCode},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL, testdata.UnaryRPCNoPayloadServerInterfaceCode},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL, testdata.UnaryRPCNoResultServerInterfaceCode},
		{"unary-rpc-with-errors", testdata.UnaryRPCWithErrorsDSL, testdata.UnaryRPCWithErrorsServerInterfaceCode},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL, testdata.ServerStreamingRPCServerInterfaceCode},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingRPCServerInterfaceCode},
		{"client-streaming-rpc-with-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.ClientStreamingRPCWithPayloadServerInterfaceCode},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingRPCServerInterfaceCode},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.BidirectionalStreamingRPCWithPayloadServerInterfaceCode},
		{"bidirectional-streaming-rpc-with-errors", testdata.BidirectionalStreamingRPCWithErrorsDSL, testdata.BidirectionalStreamingRPCWithErrorsServerInterfaceCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].Section("server-grpc-interface")
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

func TestRequestDecoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"payload-user-type", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.PayloadUserTypeRequestDecoderCode},
		{"payload-array", testdata.UnaryRPCNoResultDSL, testdata.PayloadArrayRequestDecoderCode},
		{"payload-map", testdata.MessageMapDSL, testdata.PayloadMapRequestDecoderCode},
		{"payload-primitive", testdata.ServerStreamingRPCDSL, testdata.PayloadPrimitiveRequestDecoderCode},
		{"payload-primitive-with-streaming-payload", testdata.ClientStreamingRPCWithPayloadDSL, testdata.PayloadPrimitiveWithStreamingPayloadRequestDecoderCode},
		{"payload-user-type-with-streaming-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL, testdata.PayloadUserTypeWithStreamingPayloadRequestDecoderCode},
		{"payload-with-metadata", testdata.MessageWithMetadataDSL, testdata.PayloadWithMetadataRequestDecoderCode},
		{"payload-with-security-attributes", testdata.MessageWithSecurityAttrsDSL, testdata.PayloadWithSecurityAttrsRequestDecoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].Section("request-decoder")
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

func TestResponseEncoder(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"result-with-views", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.ResultWithViewsResponseEncoderCode},
		{"result-array", testdata.MessageArrayDSL, testdata.ResultArrayResponseEncoderCode},
		{"result-primitive", testdata.UnaryRPCNoPayloadDSL, testdata.ResultPrimitiveResponseEncoderCode},
		{"result-with-metadata", testdata.MessageWithMetadataDSL, testdata.ResultWithMetadataResponseEncoderCode},
		{"result-collection", testdata.MessageResultTypeCollectionDSL, testdata.ResultCollectionResponseEncoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].Section("response-encoder")
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
