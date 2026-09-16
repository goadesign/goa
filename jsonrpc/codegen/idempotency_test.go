package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestIdempotentJSONRPCEndpointCodegen(t *testing.T) {
	root := expr.RunDSL(t, func() {
		API("retry", func() {
			JSONRPC(func() {})
		})
		Service("Retry", func() {
			JSONRPC(func() {
				POST("/rpc")
			})
			Method("read", func() {
				Idempotent()
				Error("busy", func() {
					Temporary()
				})
				JSONRPC(func() {
					Response("busy", RPCInternalError)
				})
			})
		})
	})
	plan := CreateJSONRPCPlan(root)
	clientFiles := plan.ClientFiles()
	require.NotEmpty(t, clientFiles)

	var clientCode string
	for _, file := range clientFiles {
		clientCode += codegen.SectionsCode(t, file.Section("jsonrpc-client-endpoint-init"))
	}

	assert.Contains(t, clientCode, `goa.RetryEndpoint(endpoint, "busy")`)
}
