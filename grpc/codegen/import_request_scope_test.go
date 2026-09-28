// These tests distinguish a named Go reference from code that constructs its
// nested values, including types placed through an enclosing shared declaration.
package codegen

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGRPCImportRequestsPreserveTraversal(t *testing.T) {
	tests := []struct {
		name        string
		reference   bool
		complete    bool
		wantImports []string
	}{
		{
			name:        "named reference",
			reference:   true,
			wantImports: []string{"generated.local/gen/types"},
		},
		{
			name:        "complete value",
			complete:    true,
			wantImports: []string{"generated.local/gen/types", "generated.local/gen/details", "time"},
		},
		{
			name:        "complete wins overlap",
			reference:   true,
			complete:    true,
			wantImports: []string{"generated.local/gen/types", "generated.local/gen/details", "time"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, detail := grpcInheritedErrorImportRoot(t)
			generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
			plans, err := newPlans(
				generation,
				fixedProtobufToolResolver(),
				PlanInput{Root: root, Service: services[0]},
			)
			require.NoError(t, err)
			servicePlan := plans[0].servicesPlan[0]
			attribute := servicePlan.endpoints[0].expression.MethodExpr.Payload
			input := grpcFileImportInput{endpoints: servicePlan.endpoints}
			if test.reference {
				input.typeDefinitions = []*expr.AttributeExpr{attribute}
			}
			if test.complete {
				input.typeReferences = []*expr.AttributeExpr{attribute}
			}
			filePath := path.Join(codegen.Gendir, "grpc", servicePlan.packages.pathName, "server", "selection.go")
			output := generation.Package(path.Join(generation.GenPkg(), "grpc", servicePlan.packages.pathName, "server"))
			require.NoError(t, recordGRPCFileImports(plans[0], filePath, output, input))
			imports := plans[0].fileImports[grpcFilePathKey(filePath)]
			require.ElementsMatch(t, test.wantImports, imports.Paths())
			require.NotContains(t, detail.Attribute().Meta, "struct:pkg:path")
			require.NoError(t, generation.Freeze())
			require.NoError(t, services[0].Link())
			require.NoError(t, plans[0].Link())
			linked := make([]string, 0, len(imports.Imports()))
			for _, imported := range imports.Imports() {
				linked = append(linked, imported.Path)
			}
			require.ElementsMatch(t, test.wantImports, linked)
		})
	}
}

func TestGRPCInheritedErrorImportsCompile(t *testing.T) {
	root, detail := grpcInheritedErrorImportRoot(t)
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: services[0]})
	require.NoError(t, err)
	require.NotContains(t, detail.Attribute().Meta, "struct:pkg:path")
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())

	header := sectionCode(t, plans[0].ServerFiles()[0].SectionTemplates[0])
	require.Contains(t, header, `types "generated.local/gen/types"`)
	require.NotContains(t, header, `"generated.local/gen/details"`)
	require.NotContains(t, header, `"time"`)
	for _, files := range [][]*codegen.File{plans[0].ClientTypeFiles(), plans[0].ServerTypeFiles()} {
		header := sectionCode(t, files[0].SectionTemplates[0])
		require.Contains(t, header, `types "generated.local/gen/types"`)
		require.Contains(t, header, `details "generated.local/gen/details"`)
		require.Contains(t, header, `"time"`)
	}
	compileProtobufMethodServer(t, plans[0], services)
	require.NotContains(t, detail.Attribute().Meta, "struct:pkg:path")
}

// grpcInheritedErrorImportRoot returns an error with an unannotated child in the
// parent's package. The child's fields need separate generated and custom imports.
func grpcInheritedErrorImportRoot(t *testing.T) (*expr.RootExpr, expr.UserType) {
	t.Helper()
	var detail expr.UserType
	root := expr.RunDSL(t, func() {
		code := dsl.Type("DetailCode", func() {
			dsl.Meta("struct:pkg:path", "details")
			dsl.Field(1, "value", dsl.String)
		})
		detail = dsl.Type("Detail", func() {
			dsl.Field(1, "code", code)
			dsl.Field(2, "delay", dsl.Int64, func() {
				dsl.Meta("struct:field:type", "time.Duration", "time")
			})
			dsl.Required("code", "delay")
		})
		failure := dsl.Type("Failure", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Field(1, "detail", detail)
			dsl.Required("detail")
		})
		dsl.Service("Imports", func() {
			dsl.Method("Read", func() {
				dsl.Payload(detail)
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeFailedPrecondition)
				})
			})
		})
	})
	return root, detail
}
