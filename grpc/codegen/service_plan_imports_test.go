// This file checks that copied gRPC values keep every package needed by their
// fields, even when the copies came from the same Goa type.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// TestCollectGRPCProtobufImportsVisitsEachCopiedType checks that one copied
// type cannot hide an external protobuf field used by another copy.
func TestCollectGRPCProtobufImportsVisitsEachCopiedType(t *testing.T) {
	first, second := grpcImportCopies()
	protobufField := expr.AsObject(second).Attribute("value")
	protobufField.Meta["struct:field:proto"] = []string{
		"google.protobuf.Timestamp",
		"google/protobuf/timestamp.proto",
		"Timestamp",
		"google.golang.org/protobuf/types/known/timestamppb",
	}

	for _, roots := range [][2]expr.UserType{{first, second}, {second, first}} {
		plan := &protobufServicePlan{messages: []*protobufEndpointMessages{{
			request:  &expr.AttributeExpr{Type: roots[0]},
			response: &expr.AttributeExpr{Type: roots[1]},
		}}}
		protoImports, goImports := collectGRPCProtobufImports(plan)

		require.Contains(t, protoImports, "google/protobuf/timestamp.proto")
		require.Contains(t, goImports, codegen.NewImport(
			"timestamppb",
			"google.golang.org/protobuf/types/known/timestamppb",
		))
	}
}

// TestGRPCServiceImportPathsVisitsEachCopiedType checks that one copied type
// cannot hide a Go field package used by another copy.
func TestGRPCServiceImportPathsVisitsEachCopiedType(t *testing.T) {
	first, second := grpcImportCopies()
	goField := expr.AsObject(second).Attribute("value")
	goField.Meta["struct:field:type"] = []string{"time.Time", "time"}
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)

	for _, roots := range [][2]expr.UserType{{first, second}, {second, first}} {
		service := &expr.GRPCServiceExpr{
			ServiceExpr: &expr.ServiceExpr{Name: "Imports"},
			GRPCEndpoints: []*expr.GRPCEndpointExpr{{MethodExpr: &expr.MethodExpr{
				Payload:          &expr.AttributeExpr{Type: roots[0]},
				Result:           &expr.AttributeExpr{Type: roots[1]},
				StreamingPayload: &expr.AttributeExpr{Type: expr.Empty},
				StreamingResult:  &expr.AttributeExpr{Type: expr.Empty},
			}}},
		}

		require.Contains(t, grpcServiceImportPaths(generation, service), "time")
	}
}

// grpcImportCopies returns two independent copies of one Goa type.
func grpcImportCopies() (expr.UserType, expr.UserType) {
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	return expr.Dup(original).(expr.UserType), expr.Dup(original).(expr.UserType)
}
