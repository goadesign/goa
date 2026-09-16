// This file checks that generated starter servers pass transport arguments in
// the same order accepted by their generated helper functions.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestJSONRPCOnlyExampleServerCompiles checks a server whose two services are
// both exposed only through JSON-RPC.
func TestJSONRPCOnlyExampleServerCompiles(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		for _, name := range []string{"First", "Second"} {
			dsl.Service(name, func() {
				dsl.Method("Read", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.JSONRPC(func() {})
				})
			})
		}
	})
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	files = append(files, exampleFiles...)
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	runGeneratedTests(t, directory)
}
