// This file verifies generated transport packages and example applications
// consume the same complete-path aliases selected during planning.
package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestGeneratedTransportPackagesCompileWithServiceAliasCollisions proves that
// client, server, protobuf, CLI, and service imports remain paired when their
// preferred qualifiers collide.
func TestGeneratedTransportPackagesCompileWithServiceAliasCollisions(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		interceptor := dsl.Interceptor("Trace", func() {})
		for _, name := range []string{"Foo", "Fooc", "Foosvr", "Foojssvr"} {
			dsl.Service(name, func() {
				if name == "Foo" {
					dsl.ClientInterceptor(interceptor)
				}
				dsl.Method("Read", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.HTTP(func() { dsl.POST("/" + strings.ToLower(name)) })
					dsl.GRPC(func() {})
				})
				dsl.Method("ReadJSON", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.JSONRPC(func() {})
				})
			})
		}
	})

	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, planTransportData(generation))
	require.NoError(t, generation.Freeze())
	files, err := Service(generation)
	require.NoError(t, err)
	transport, err := Transport(generation)
	require.NoError(t, err)
	files = append(files, transport...)
	examples, err := Example(generation)
	require.NoError(t, err)
	files = append(files, examples...)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	runGeneratedTests(t, dir)
}
