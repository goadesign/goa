// This file verifies gRPC conversions share strict private functions when the
// generated source and target types are the same.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
)

// TestTransformHelpersShareRequiredAndOptionalCalls checks that a required
// field and an optional field in one conversion plan share one function. The
// optional call remains inside its nil check.
func TestTransformHelpersShareRequiredAndOptionalCalls(t *testing.T) {
	root := RunGRPCDSL(t, sharedGRPCTransformHelperDSL)
	services := CreateGRPCServices(root)
	service := services.Get("SharedHelpers")
	require.Len(t, service.clientTransformHelpers, 2)
	require.Len(t, service.serverTransformHelpers, 1)
	var clientHelper *codegen.TransformFunctionData
	for _, helper := range service.clientTransformHelpers {
		if strings.HasPrefix(helper.ParamTypeRef, "*sharedhelpers.") {
			clientHelper = helper
		}
	}
	require.NotNil(t, clientHelper)
	serverHelper := service.serverTransformHelpers[0]
	require.NotContains(t, clientHelper.Code, "if v == nil")
	require.NotContains(t, serverHelper.Code, "if v == nil")

	request := service.Endpoints[0].Request
	clientCode := request.ClientConvert.Init.Code
	require.Contains(t, clientCode, "message.Left = "+clientHelper.Declaration.Name()+"(payload.Left)")
	require.Contains(t, clientCode, "if payload.Right != nil {")
	require.Contains(t, clientCode, "message.Right = "+clientHelper.Declaration.Name()+"(payload.Right)")
	serverCode := request.ServerConvert.Init.Code
	require.Contains(t, serverCode, "v.Left = "+serverHelper.Declaration.Name()+"(message.Left)")
	require.Contains(t, serverCode, "if message.Right != nil {")
	require.Contains(t, serverCode, "v.Right = "+serverHelper.Declaration.Name()+"(message.Right)")

	client := codegen.SectionsCode(t, clientTypeFiles(services)[0].SectionTemplates[1:])
	testutil.AssertGo(t, "testdata/golden/transform_helper_shared.go.golden", client)
}

// sharedGRPCTransformHelperDSL creates one request with required and optional
// fields that use the same named type.
func sharedGRPCTransformHelperDSL() {
	child := dsl.Type("SharedChild", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Required("value")
	})
	dsl.Service("SharedHelpers", func() {
		dsl.Method("Store", func() {
			dsl.Payload(func() {
				dsl.Field(1, "left", child)
				dsl.Field(2, "right", child)
				dsl.Required("left")
			})
			dsl.GRPC(func() {})
		})
	})
}
