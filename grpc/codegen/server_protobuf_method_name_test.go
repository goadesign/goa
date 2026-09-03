// This file checks that Goa's gRPC server methods use the names written by
// protoc when two design method names produce the same Go spelling.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestServerUsesProtobufMethodNames checks both source orders. Each generated
// server method must match the protobuf method for the same endpoint.
func TestServerUsesProtobufMethodNames(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "underscored first"
		if reverse {
			name = "camel case first"
		}
		t.Run(name, func(t *testing.T) {
			root := RunGRPCDSL(t, collidingProtobufMethodDSL(reverse))
			generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
			plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlans[0].Link())
			require.NoError(t, plans[0].Link())
			services := plans[0].services
			service := services.Get("Values")
			sections := serverFiles(services)[0].Section("server-grpc-interface")
			require.Len(t, sections, len(service.Endpoints))

			foundDifferentName := false
			for index, endpoint := range service.Endpoints {
				foundDifferentName = foundDifferentName || endpoint.GRPCMethodName != endpoint.Method.VarName
				code := codegen.SectionCode(t, sections[index])
				require.Contains(t, code, "func (s *"+endpoint.ServerStruct+") "+endpoint.GRPCMethodName+"(")
			}
			require.True(t, foundDifferentName, "test methods did not exercise different service and protobuf names")
			compileProtobufMethodServer(t, plans[0], servicePlans)
		})
	}
}

// collidingProtobufMethodDSL defines two methods that produce the same initial
// Go name but use different request and response messages.
func collidingProtobufMethodDSL(reverse bool) func() {
	return func() {
		underscored := func() {
			dsl.Method("read_value", func() {
				dsl.Payload(func() { dsl.Field(1, "text", dsl.String) })
				dsl.Result(func() { dsl.Field(1, "text", dsl.String) })
				dsl.GRPC(func() {})
			})
		}
		camelCase := func() {
			dsl.Method("readValue", func() {
				dsl.Payload(func() { dsl.Field(1, "number", dsl.Int) })
				dsl.Result(func() { dsl.Field(1, "number", dsl.Int) })
				dsl.GRPC(func() {})
			})
		}
		dsl.Service("Values", func() {
			if reverse {
				camelCase()
				underscored()
				return
			}
			underscored()
			camelCase()
		})
	}
}

// compileProtobufMethodServer writes the service and transport files and asks
// Go to check that the server implements the generated protobuf interface.
func compileProtobufMethodServer(t *testing.T, plan *Plan, servicePlans []*service.Plan) {
	t.Helper()
	files, err := service.Files(servicePlans...)
	require.NoError(t, err)
	files = append(files, plan.ServerFiles()...)
	files = append(files, plan.ClientFiles()...)
	files = append(files, plan.ServerTypeFiles()...)
	files = append(files, plan.ClientTypeFiles()...)
	files = append(files, plan.ProtoFiles()...)
	moduleDir := t.TempDir()
	writeProtobufDescriptorModule(t, moduleDir)
	for _, file := range files {
		_, err := file.Render(moduleDir)
		require.NoError(t, err)
	}
	compileProtobufDescriptorModule(t, moduleDir)
}
