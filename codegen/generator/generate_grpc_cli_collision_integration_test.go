// This file verifies that a generated gRPC command parser calls the exact
// client constructor selected for its generated client package.
package generator

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

type grpcClientCollisionOrder string

// ComparePackageName orders the declaration added by this collision test.
func (o grpcClientCollisionOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return strings.Compare(string(o), string(other.(grpcClientCollisionOrder)))
}

func TestGeneratedGRPCCLIUsesFinalClientConstructorName(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Records", func() {
			dsl.Method("Read", func() {
				dsl.Result(dsl.String)
				dsl.GRPC(func() {})
			})
		})
	})
	reserve := func(plan *Plan) error {
		clientPackage, err := plan.Generation().ClaimPackage(path.Join(
			"generated.local/gen", "grpc", "records", "client",
		))
		if err != nil {
			return err
		}
		return clientPackage.DeclareName(codegen.NewPreferredName(
			codegen.NameFunction,
			"NewClient",
			codegen.ExportedName,
			grpcClientCollisionOrder("plugin-client-constructor"),
		))
	}
	plan := mustTestPlan(
		t,
		"generated.local/gen",
		[]eval.Root{root},
		planServiceData,
		reserve,
		planTransportData,
	)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	runGeneratedTests(t, directory)
}
