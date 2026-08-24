// This file verifies generated protobuf services and message declarations,
// including package-owned naming collisions, by compiling each schema.
package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestProtoFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"protofiles-unary-rpcs", testdata.UnaryRPCsDSL},
		{"protofiles-unary-rpc-no-payload", testdata.UnaryRPCNoPayloadDSL},
		{"protofiles-unary-rpc-no-result", testdata.UnaryRPCNoResultDSL},
		{"protofiles-server-streaming-rpc", testdata.ServerStreamingRPCDSL},
		{"protofiles-client-streaming-rpc", testdata.ClientStreamingRPCDSL},
		{"protofiles-bidirectional-streaming-rpc", testdata.BidirectionalStreamingRPCDSL},
		{"protofiles-same-service-and-message-name", testdata.MessageWithServiceNameDSL},
		{"protofiles-method-with-reserved-proto-name", testdata.MethodWithReservedNameDSL},
		{"protofiles-multiple-methods-same-return-type", testdata.MultipleMethodsSameResultCollectionDSL},
		{"protofiles-method-with-acronym", testdata.MethodWithAcronymDSL},
		{"protofiles-custom-package-name", testdata.ServiceWithPackageDSL},
		{"protofiles-struct-meta-type", testdata.StructMetaTypeDSL},
		{"protofiles-default-fields", testdata.DefaultFieldsDSL},
		{"protofiles-required-primitive-presence", testdata.RequiredPrimitivePresenceDSL},
		{"protofiles-custom-message-name", testdata.CustomMessageNameDSL},
		{"protofiles-distinct-custom-message-names", testdata.DistinctCustomMessageNamesDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := protoFiles(services)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].SectionTemplates
			require.GreaterOrEqual(t, len(sections), 3)
			code := sectionCode(t, sections[1:]...)
			// testutil.AssertString handles line ending normalization internally
			testutil.AssertString(t, "testdata/golden/proto_"+c.Name+".proto.golden", code)
			fpath := codegen.CreateTempFile(t, code)
			assert.NoError(t, protoc(defaultProtocCmd, fpath), "error occurred when compiling proto file %q", fpath)
		})
	}
}

func TestMessageDefSection(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"user-type-with-primitives", testdata.MessageUserTypeWithPrimitivesDSL},
		{"user-type-with-alias", testdata.MessageUserTypeWithAliasMessageDSL},
		{"user-type-with-nested-user-types", testdata.MessageUserTypeWithNestedUserTypesDSL},
		{"result-type-collection", testdata.MessageResultTypeCollectionDSL},
		{"user-type-with-collection", testdata.MessageUserTypeWithCollectionDSL},
		{"array", testdata.MessageArrayDSL},
		{"map", testdata.MessageMapDSL},
		{"primitive", testdata.MessagePrimitiveDSL},
		{"with-metadata", testdata.MessageWithMetadataDSL},
		{"with-security-attributes", testdata.MessageWithSecurityAttrsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := protoFiles(services)
			require.Len(t, fs, 1)
			sections := fs[0].SectionTemplates
			require.GreaterOrEqual(t, len(sections), 3)
			code := sectionCode(t, sections[:2]...)
			msgCode := sectionCode(t, sections[3:]...)
			// testutil.AssertString handles line ending normalization internally
			testutil.AssertString(t, "testdata/golden/proto_"+c.Name+".proto.golden", code+msgCode)
			fpath := codegen.CreateTempFile(t, code+msgCode)
			assert.NoError(t, protoc(defaultProtocCmd, fpath), "error occurred when compiling proto file %q", fpath)
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
			require.NoError(t, os.WriteFile(fpath, []byte(code), 0o600), "error occurred writing proto schema")
			require.NoError(t, protoc(c.Cmd, fpath), "error occurred when compiling proto file with the standard protoc %q", fpath)

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
