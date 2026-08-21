package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestIdempotentRPCCodegen(t *testing.T) {
	root := RunGRPCDSL(t, testdata.IdempotentRPCsDSL)
	services := CreateGRPCServices(root)

	protoFiles := ProtoFiles(services)
	require.Len(t, protoFiles, 1)
	protoCode := sectionCode(t, protoFiles[0].SectionTemplates[1:]...)
	assert.Equal(t, 2, strings.Count(protoCode, "option idempotency_level = IDEMPOTENT;"))
	protoPath := codegen.CreateTempFile(t, protoCode)
	assert.NoError(t, protoc(defaultProtocCmd, protoPath, nil))

	clientFiles := ClientFiles(services)
	require.Len(t, clientFiles, 2)
	clientCode := codegen.SectionsCode(t, clientFiles[0].Section("client-endpoint-init"))
	assert.Contains(t, clientCode, `goa.RetryEndpoint(endpoint, "busy")`)
	assert.Equal(t, 1, strings.Count(clientCode, "goa.RetryEndpoint("))
}
