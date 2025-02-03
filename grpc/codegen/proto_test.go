package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestProtoFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"protofiles-unary-rpcs", testdata.UnaryRPCsDSL, testdata.UnaryRPCsProtoCode},
		{"protofiles-unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL, testdata.UnaryRPCNoPayloadProtoCode},
		{"protofiles-unary-rpc-no-result", testdata.UnaryRPCNoResultDSL, testdata.UnaryRPCNoResultProtoCode},
		{"protofiles-server-streaming-rpc", testdata.ServerStreamingRPCDSL, testdata.ServerStreamingRPCProtoCode},
		{"protofiles-client-streaming-rpc", testdata.ClientStreamingRPCDSL, testdata.ClientStreamingRPCProtoCode},
		{"protofiles-bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL, testdata.BidirectionalStreamingRPCProtoCode},
		{"protofiles-same-service-and-message-name", testdata.MessageWithServiceNameDSL, testdata.MessageWithServiceNameProtoCode},
		{"protofiles-method-with-reserved-proto-name", testdata.MethodWithReservedNameDSL, testdata.MethodWithReservedNameProtoCode},
		{"protofiles-multiple-methods-same-return-type", testdata.MultipleMethodsSameResultCollectionDSL, testdata.MultipleMethodsSameResultCollectionProtoCode},
		{"protofiles-method-with-acronym", testdata.MethodWithAcronymDSL, testdata.MethodWithAcronymProtoCode},
		{"protofiles-custom-package-name", testdata.ServiceWithPackageDSL, testdata.ServiceWithPackageCode},
		{"protofiles-struct-meta-type", testdata.StructMetaTypeDSL, testdata.StructMetaTypePackageCode},
		{"protofiles-default-fields", testdata.DefaultFieldsDSL, testdata.DefaultFieldsPackageCode},
		{"protofiles-custom-message-name", testdata.CustomMessageNameDSL, testdata.CustomMessageNamePackageCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ProtoFiles("", expr.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].SectionTemplates
			require.GreaterOrEqual(t, len(sections), 3)
			code := sectionCode(t, sections[1:]...)
			if runtime.GOOS == "windows" {
				code = strings.ReplaceAll(code, "\r\n", "\n")
			}
			assert.Equal(t, c.Code, code)
			fpath := codegen.CreateTempFile(t, code)
			assert.NoError(t, protoc(defaultProtocCmd, fpath, nil), "error occurred when compiling proto file %q", fpath)
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
		{"user-type-with-alias", testdata.MessageUserTypeWithAliasMessageDSL, testdata.MessageUserTypeWithAliasMessageCode},
		{"user-type-with-nested-user-types", testdata.MessageUserTypeWithNestedUserTypesDSL, testdata.MessageUserTypeWithNestedUserTypesCode},
		{"result-type-collection", testdata.MessageResultTypeCollectionDSL, testdata.MessageResultTypeCollectionCode},
		{"user-type-with-collection", testdata.MessageUserTypeWithCollectionDSL, testdata.MessageUserTypeWithCollectionCode},
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
			require.Len(t, fs, 1)
			sections := fs[0].SectionTemplates
			require.GreaterOrEqual(t, len(sections), 3)
			code := sectionCode(t, sections[:2]...)
			msgCode := sectionCode(t, sections[3:]...)
			if runtime.GOOS == "windows" {
				msgCode = strings.ReplaceAll(msgCode, "\r\n", "\n")
			}
			assert.Equal(t, c.Code, msgCode)
			fpath := codegen.CreateTempFile(t, code+msgCode)
			assert.NoError(t, protoc(defaultProtocCmd, fpath, nil), "error occurred when compiling proto file %q", fpath)
		})
	}
}

func TestProtoc(t *testing.T) {
	const code = testdata.UnaryRPCsProtoCode

	fakeBin := filepath.Join(os.TempDir(), t.Name()+"-fakeprotoc")
	if runtime.GOOS == "windows" {
		fakeBin += ".exe"
	}
	out, err := exec.Command("go", "build", "-o", fakeBin, "./testdata/protoc").CombinedOutput()
	t.Log("go build output: ", string(out))
	require.NoError(t, err, "compile a fake protoc that requires a prefix")
	t.Cleanup(func() { assert.NoError(t, os.Remove(fakeBin)) })

	cases := []struct {
		Name string
		Cmd  []string
	}{
		{"protoc", defaultProtocCmd},
		{"fakepc", []string{fakeBin, "required-ignored-arg"}},
	}

	var firstOutput string

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", strings.ReplaceAll(t.Name(), "/", "-"))
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, os.RemoveAll(dir)) })
			fpath := filepath.Join(dir, "schema")
			require.NoError(t, os.WriteFile(fpath, []byte(code), 0o600), "error occured writing proto schema")
			require.NoError(t, protoc(c.Cmd, fpath, nil), "error occurred when compiling proto file with the standard protoc %q", fpath)

			fcontents, err := os.ReadFile(fpath + ".pb.go")
			require.NoError(t, err)

			if firstOutput == "" {
				firstOutput = string(fcontents)
				return
			}

			assert.Equal(t, firstOutput, string(fcontents))
		})
	}
}
