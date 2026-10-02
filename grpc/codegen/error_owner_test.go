// These tests check that declared-error helpers use planned private names and
// imports, and that designs without declarations do not emit unused helpers.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGRPCDeclaredErrorOwnerPlanning(t *testing.T) {
	for _, custom := range []bool{false, true} {
		name := "shared detail"
		if custom {
			name = "custom detail"
		}
		t.Run(name, func(t *testing.T) {
			root := expr.RunDSL(t, func() {
				var errType expr.DataType = expr.ErrorResult
				if custom {
					errType = dsl.Type("Denied", func() {
						dsl.Field(1, "reason", dsl.String)
					})
				}
				dsl.Service("Status", func() {
					dsl.Method("Read", func() {
						dsl.Error("denied", errType)
						dsl.GRPC(func() {
							dsl.Response("denied", dsl.CodePermissionDenied)
						})
					})
				})
			})
			generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
			server, err := generation.ClaimPackage("generated.local/gen/grpc/status/server")
			require.NoError(t, err)
			for _, name := range []string{"errorOwner", "nextError"} {
				require.NoError(t, server.DeclareName(codegen.NewExactName(codegen.NameFunction, name)))
			}
			plans, err := newPlans(generation, fixedProtobufToolResolver(), PlanInput{Root: root, Service: services[0]})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			require.NoError(t, services[0].Link())
			require.NoError(t, plans[0].Link())
			data, ok := plans[0].ServiceData(root.API.GRPC.Services[0])
			require.True(t, ok)
			require.NotEqual(t, "errorOwner", data.serverErrors.Owner.Name())
			require.NotEqual(t, "nextError", data.serverErrors.Next.Name())
			files := plans[0].ServerFiles()
			owner := codegen.SectionsCode(t, files[0].Section("server-error-owner"))
			method := codegen.SectionsCode(t, files[0].Section("server-grpc-interface"))
			require.Contains(t, owner, "func "+data.serverErrors.Owner.Name()+"(")
			require.Contains(t, owner, data.serverErrors.Next.Name()+"(current)")
			require.Contains(t, method, data.serverErrors.Owner.Name()+"(err)")
			require.Contains(t, sectionCode(t, files[0].SectionTemplates[0]), `"google.golang.org/grpc/status"`)
			require.Contains(t, owner, "GRPCStatus() *status.Status")
			require.Equal(t, custom, hasCustomErrors(data.Endpoints[0].Errors))
			if custom {
				require.Contains(t, method, "if joined")
			} else {
				require.Contains(t, method, "en, _, _")
				require.NotContains(t, method, "if joined")
			}
		})
	}
	root := RunGRPCDSL(t, func() {
		dsl.Service("Ordinary", func() {
			dsl.Method("Read", func() { dsl.GRPC(func() {}) })
		})
	})
	files := serverFiles(CreateGRPCServices(root))
	require.Empty(t, files[0].Section("server-error-owner"))
	require.NotContains(t, sectionCode(t, files[0].SectionTemplates[0]), `"google.golang.org/grpc/status"`)
}
