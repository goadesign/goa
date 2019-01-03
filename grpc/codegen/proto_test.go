package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/grpc/codegen/testdata"
)

func TestProtoFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL, testdata.UnaryRPCsProtoCode},
		{"unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL, testdata.UnaryRPCNoPayloadProtoCode},
		{"unary-rpc-no-result", testdata.UnaryRPCNoResultDSL, testdata.UnaryRPCNoResultProtoCode},
		{"server-streaming-rpc", testdata.ServerStreamingRPCDSL, testdata.ServerStreamingRPCProtoCode},
		{"client-streaming-rpc", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingRPCProtoCode},
		{"bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingRPCProtoCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ProtoFiles("", expr.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 3 {
				t.Fatalf("got %d sections, expected at least three", len(sections))
			}
			hdrCode := sectionCode(t, sections[0])
			code := sectionCode(t, sections[1:]...)
			if code != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, code, codegen.Diff(t, code, c.Code))
			}
			fpath := codegen.CreateTempFile(t, hdrCode+code)
			if err := protoc(fpath); err != nil {
				t.Fatalf("error occurred when compiling proto file %q: %s", fpath, err)
			}
		})
	}
}

func TestMessageDefSection(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"user-type-with-primitives", testdata.MessageUserTypeWithPrimitivesDSL, testdata.MessageUserTypeWithPrimitivesMessageCode},
		{"user-type-with-nested-user-types", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.MessageUserTypeWithNestedUserTypesCode},
		{"result-type-collection", testdata.MessageResultTypeCollectionDSL, testdata.MessageResultTypeCollectionCode},
		{"array", testdata.MessageArrayDSL, testdata.MessageArrayCode},
		{"map", testdata.MessageMapDSL, testdata.MessageMapCode},
		{"primitive", testdata.MessagePrimitiveDSL, testdata.MessagePrimitiveCode},
		{"with-metadata", testdata.MessageWithMetadataDSL, testdata.MessageWithMetadataCode},
		{"with-security-attributes", testdata.MessageWithSecurityAttrsDSL, testdata.MessageWithSecurityAttrsCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ProtoFiles("", expr.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 3 {
				t.Fatalf("got %d sections, expected at least three", len(sections))
			}
			code := sectionCode(t, sections[:1]...)
			msgCode := sectionCode(t, sections[2:]...)
			if msgCode != c.Code {
				t.Errorf("%s: got\n%s\ngot vs. expected:\n%s", c.Name, msgCode, codegen.Diff(t, msgCode, c.Code))
			}
			fpath := codegen.CreateTempFile(t, code+msgCode)
			if err := protoc(fpath); err != nil {
				t.Fatalf("error occurred when compiling proto file %q: %s", fpath, err)
			}
		})
	}
}
