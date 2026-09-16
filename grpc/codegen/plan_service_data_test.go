// This file checks that retained gRPC plans link the exact service expressions
// and do not invent protobuf messages for payloads built only from metadata.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
)

func TestPlanServiceDataUsesExactExpressionAfterLink(t *testing.T) {
	roots := grpcPlanRoots(t, "Calc")
	generation, services := grpcServicePlans(t, roots)
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: roots[0], Service: services[0]},
	)
	require.NoError(t, err)
	require.PanicsWithValue(t, "gRPC files requested before plan linking", func() {
		plans[0].ServiceData(roots[0].API.GRPC.Services[0])
	})

	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())

	data, ok := plans[0].ServiceData(roots[0].API.GRPC.Services[0])
	require.True(t, ok)
	require.Equal(t, "Calc", data.Name)
	require.NotEmpty(t, data.ClientStruct)
	require.NotEmpty(t, data.ServerStruct)

	foreign := grpcPlanRoots(t, "Calc")
	data, ok = plans[0].ServiceData(foreign[0].API.GRPC.Services[0])
	require.False(t, ok)
	require.Nil(t, data)
}

// TestPlanLinksMetadataOnlyStreamingPayload verifies that a payload built only
// from metadata does not require a protobuf request message.
func TestPlanLinksMetadataOnlyStreamingPayload(t *testing.T) {
	root := RunGRPCDSL(t, func() {
		dsl.Service("Chatter", func() {
			dsl.Method("Echo", func() {
				dsl.Payload(func() {
					dsl.Field(1, "token", dsl.String)
					dsl.Required("token")
				})
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.GRPC(func() {
					dsl.Metadata(func() {
						dsl.Attribute("token")
					})
				})
			})
		})
	})

	services := CreateGRPCServices(root)
	request := services.Get("Chatter").Endpoints[0].Request
	require.Nil(t, request.PayloadMessage)
	require.NotNil(t, request.ServerConvert)
	require.Empty(t, request.ServerConvert.SrcName)
	require.Empty(t, request.ServerConvert.SrcRef)
	require.Nil(t, request.ServerConvert.Validation)
}
