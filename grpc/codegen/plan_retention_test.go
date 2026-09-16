// This file proves gRPC files use the service and endpoint values saved by
// NewPlans even when a caller changes the evaluated design before Link.
package codegen

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

type grpcPlanRetentionFixture struct {
	root     *expr.RootExpr
	service  *expr.GRPCServiceExpr
	endpoint *expr.GRPCEndpointExpr
	method   *expr.MethodExpr
}

// TestGRPCPlanIgnoresDesignChangesAfterPlanning checks endpoint membership,
// messages, metadata, validation, and streaming decisions separately.
func TestGRPCPlanIgnoresDesignChangesAfterPlanning(t *testing.T) {
	baselineFixture := grpcPlanRetentionDSL(t)
	baseline := renderGRPCPlanRetentionFixture(t, baselineFixture, nil)

	tests := []struct {
		name   string
		mutate func(*grpcPlanRetentionFixture)
	}{
		{"service membership", func(f *grpcPlanRetentionFixture) {
			f.root.API.GRPC.Services = nil
		}},
		{"server membership", func(f *grpcPlanRetentionFixture) {
			f.root.API.Servers = nil
		}},
		{"endpoint membership", func(f *grpcPlanRetentionFixture) {
			f.service.GRPCEndpoints = append(f.service.GRPCEndpoints, &expr.GRPCEndpointExpr{})
		}},
		{"request messages", func(f *grpcPlanRetentionFixture) {
			f.endpoint.Request.Type = expr.Empty
			f.endpoint.StreamingRequest.Type = expr.Empty
		}},
		{"response message and metadata", func(f *grpcPlanRetentionFixture) {
			f.endpoint.Response.Message.Type = expr.Empty
			f.endpoint.Response.Headers = expr.NewEmptyMappedAttributeExpr()
			f.endpoint.Response.Trailers = expr.NewEmptyMappedAttributeExpr()
			f.endpoint.Response.StatusCode = 13
		}},
		{"request metadata", func(f *grpcPlanRetentionFixture) {
			f.endpoint.Metadata = expr.NewEmptyMappedAttributeExpr()
		}},
		{"validation", func(f *grpcPlanRetentionFixture) {
			f.endpoint.Request.Validation.Required = nil
			field := expr.AsObject(f.method.Payload.Type).Attribute("value")
			field.Validation.MinLength = nil
		}},
		{"imports", func(f *grpcPlanRetentionFixture) {
			field := expr.AsObject(f.method.Payload.Type).Attribute("value")
			field.Meta["struct:field:type"] = []string{"time.Time", "time"}
		}},
		{"streaming method", func(f *grpcPlanRetentionFixture) {
			f.method.Stream = expr.NoStreamKind
			f.method.StreamingPayload.Type = expr.Empty
			f.method.StreamingResult.Type = expr.Empty
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := grpcPlanRetentionDSL(t)
			actual := renderGRPCPlanRetentionFixture(t, fixture, test.mutate)
			require.Equal(t, baseline, actual)
		})
	}
}

// TestGRPCPlanCopiesMissingStreamingResult checks that a unary method keeps
// its allowed nil streaming result when NewPlans copies the method.
func TestGRPCPlanCopiesMissingStreamingResult(t *testing.T) {
	root := grpcPlanRoots(t, "Unary")[0]
	require.Nil(t, root.API.GRPC.Services[0].GRPCEndpoints[0].MethodExpr.StreamingResult)
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: root, Service: services[0]},
	)
	require.NoError(t, err)
	require.Nil(t, plans[0].servicesPlan[0].expression.GRPCEndpoints[0].MethodExpr.StreamingResult)
}

// TestGRPCPlanCopiesMethodErrors checks that each copied gRPC error uses the
// matching copied method error, including Goa's built-in error value.
func TestGRPCPlanCopiesMethodErrors(t *testing.T) {
	root := RunGRPCDSL(t, testdata.UnaryRPCWithErrorsDSL)
	service, err := copyGRPCService(root.API.GRPC.Services[0])
	require.NoError(t, err)
	endpoint := service.expression.GRPCEndpoints[0]
	for _, grpcError := range endpoint.GRPCErrors {
		require.Same(t, endpoint.MethodExpr.Error(grpcError.Name), grpcError.ErrorExpr)
	}
	require.Same(t, expr.ErrorResult, endpoint.MethodExpr.Error("timeout").Type)
}

// TestGRPCPlanKeepsEmptyPayloadResponseConversion checks that a method without
// a payload still saves and renders its result conversion.
func TestGRPCPlanKeepsEmptyPayloadResponseConversion(t *testing.T) {
	baseline := renderGRPCPlanRetentionFixture(t, grpcEmptyPayloadFixture(t), nil)
	actual := renderGRPCPlanRetentionFixture(t, grpcEmptyPayloadFixture(t), func(f *grpcPlanRetentionFixture) {
		f.method.Result.Type = expr.Empty
		f.endpoint.Response.Message.Type = expr.Empty
	})
	require.Equal(t, baseline, actual)
}

// grpcPlanRetentionDSL creates one streaming endpoint with request and
// response metadata and message validation.
func grpcPlanRetentionDSL(t *testing.T) *grpcPlanRetentionFixture {
	t.Helper()
	fixture := new(grpcPlanRetentionFixture)
	fixture.root = expr.RunDSL(t, func() {
		payload := dsl.Type("SavedPayload", func() {
			dsl.Field(1, "value", dsl.String, func() { dsl.MinLength(2) })
			dsl.Field(2, "token", dsl.String)
			dsl.Required("value", "token")
		})
		result := dsl.Type("SavedResult", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Field(2, "count", dsl.Int)
			dsl.Required("value", "count")
		})
		stream := dsl.Type("SavedStream", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("SavedTransport", func() {
			fixture.method = dsl.Method("Watch", func() {
				dsl.Payload(payload)
				dsl.StreamingPayload(stream)
				dsl.StreamingResult(result)
				dsl.GRPC(func() {
					dsl.Metadata(func() { dsl.Attribute("token:authorization") })
					dsl.Response(dsl.CodeOK, func() {
						dsl.Headers(func() { dsl.Attribute("count:x-count") })
						dsl.Trailers(func() { dsl.Attribute("value:x-value") })
					})
				})
			})
		})
	})
	fixture.service = fixture.root.API.GRPC.Services[0]
	fixture.endpoint = fixture.service.GRPCEndpoints[0]
	return fixture
}

// grpcEmptyPayloadFixture creates one unary method with no payload and a
// custom result that needs a protobuf conversion.
func grpcEmptyPayloadFixture(t *testing.T) *grpcPlanRetentionFixture {
	t.Helper()
	fixture := new(grpcPlanRetentionFixture)
	fixture.root = expr.RunDSL(t, func() {
		result := dsl.Type("SavedEmptyPayloadResult", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("SavedEmptyPayload", func() {
			fixture.method = dsl.Method("Read", func() {
				dsl.Result(result)
				dsl.GRPC(func() {})
			})
		})
	})
	fixture.service = fixture.root.API.GRPC.Services[0]
	fixture.endpoint = fixture.service.GRPCEndpoints[0]
	return fixture
}

// renderGRPCPlanRetentionFixture saves the design, applies one later change,
// and renders every non-example gRPC file.
func renderGRPCPlanRetentionFixture(t *testing.T, fixture *grpcPlanRetentionFixture, mutate func(*grpcPlanRetentionFixture)) []string {
	t.Helper()
	generation, services := grpcServicePlans(t, []*expr.RootExpr{fixture.root})
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: fixture.root, Service: services[0]},
	)
	require.NoError(t, err)
	if mutate != nil {
		mutate(fixture)
	}
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())

	files := plans[0].ProtoFiles()
	files = append(files, plans[0].ServerFiles()...)
	files = append(files, plans[0].ClientFiles()...)
	files = append(files, plans[0].ServerTypeFiles()...)
	files = append(files, plans[0].ClientTypeFiles()...)
	files = append(files, plans[0].ClientCLIFiles()...)
	result := make([]string, len(files))
	for index, file := range files {
		result[index] = file.Path + "\n" + sectionCode(t, file.SectionTemplates...)
	}
	sort.Strings(result)
	return result
}
