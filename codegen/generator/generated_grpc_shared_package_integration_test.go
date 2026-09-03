// This file checks two gRPC designs that write service, client, and server
// files into the same directories. Every written file must use the names
// chosen for both designs before generation starts.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedGRPCPackagesCompileAcrossSharedPackageRoots checks both design
// orders because input order must not change the names in shared directories.
func TestGeneratedGRPCPackagesCompileAcrossSharedPackageRoots(t *testing.T) {
	tests := []struct {
		name    string
		reverse bool
	}{
		{name: "forward"},
		{name: "reverse", reverse: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := grpcSharedPackageRoot(t, "Shared", "First")
			second := grpcSharedPackageRoot(t, "Shared", "Second")
			roots := []eval.Root{first, second}
			if test.reverse {
				roots[0], roots[1] = roots[1], roots[0]
			}

			reserveValidator := func(plan *Plan) error {
				pkg, err := plan.Generation().ClaimPackage("generated.local/gen/grpc/shared/server")
				if err != nil {
					return err
				}
				for _, name := range []string{"ValidateSyncRequest", "ValidateExchangeRequest", "ValidateExchangeStreamingRequest"} {
					if err := pkg.DeclareName(codegen.NewExactName(codegen.NameFunction, name)); err != nil {
						return err
					}
				}
				return nil
			}
			plan := mustTestPlan(t, "generated.local/gen", roots, planTransportData, reserveValidator)
			files, err := testServiceFiles(plan)
			require.NoError(t, err)
			transportFiles, err := testTransportFiles(plan)
			require.NoError(t, err)
			files = append(files, transportFiles...)
			files = append(files, &codegen.File{
				Path: "gen/grpc/shared/server/validator_owner.go",
				SectionTemplates: []*codegen.SectionTemplate{
					{
						Name: "validator-owner",
						Source: `package server

func ValidateSyncRequest() {}
func ValidateExchangeRequest() {}
func ValidateExchangeStreamingRequest() {}`,
					},
				},
			})
			files, err = mergeFilesByPath(files)
			require.NoError(t, err)

			dir := t.TempDir()
			writeGeneratedModule(t, dir, "generated.local")
			for _, file := range files {
				_, err := file.Render(dir)
				require.NoError(t, err)
			}
			runGeneratedTests(t, dir)
		})
	}
}

// grpcSharedPackageRoot returns one design whose service name chooses the
// output directories. typePrefix keeps its values separate from the other
// design that writes into those directories.
func grpcSharedPackageRoot(t *testing.T, serviceName, typePrefix string) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.API(typePrefix, func() {})
		node := dsl.Type(typePrefix+"Node", func() {
			dsl.Field(1, "name", dsl.String)
			dsl.Field(2, "next", typePrefix+"Node")
			dsl.Required("name")
		})
		payload := dsl.Type(typePrefix+"Payload", func() {
			dsl.Field(1, "id", dsl.String)
			dsl.Field(2, "node", node)
			dsl.Required("id", "node")
		})
		result := dsl.Type(typePrefix+"Result", func() {
			dsl.Field(1, "status", dsl.String)
			dsl.Field(2, "node", node)
			dsl.Required("status", "node")
		})
		failure := dsl.Type(typePrefix+"Failure", func() {
			dsl.Field(1, "message", dsl.String)
			dsl.Required("message")
		})

		dsl.Service(serviceName, func() {
			dsl.Error("failed", failure)
			dsl.GRPC(func() {
				dsl.Response("failed", dsl.CodeInvalidArgument)
			})
			dsl.Method("Sync", func() {
				dsl.Payload(payload)
				dsl.Result(result)
				dsl.Error("failed", failure)
				dsl.GRPC(func() {})
			})
			dsl.Method("Exchange", func() {
				dsl.Payload(payload)
				dsl.StreamingPayload(payload)
				dsl.StreamingResult(result)
				dsl.Error("failed", failure)
				dsl.GRPC(func() {})
			})
		})
	})
}
