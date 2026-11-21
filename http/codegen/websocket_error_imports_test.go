// This file checks that WebSocket-only HTTP servers import context in the
// server codec only when it renders an error encoder.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestHTTPWebSocketErrorEncoderImportsContext checks that a service whose only
// endpoint is a WebSocket stream imports context when it renders an error
// encoder.
func TestHTTPWebSocketErrorEncoderImportsContext(t *testing.T) {
	tests := []struct {
		name    string
		mapErr  bool
		context bool
	}{
		{name: "mapped_error", mapErr: true, context: true},
		{name: "no_error", mapErr: false, context: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := webSocketErrorRoot(t, test.mapErr)
			plan := linkedHTTPPlanForRoot(t, root)
			imports := plannedHTTPFileImports(t, plan.ServerFiles(), "/server/encode_decode.go")
			require.Equal(t, test.context, slices.Contains(imports, "context"))

			serviceFiles, err := service.Files(plan.servicePlan)
			require.NoError(t, err)
			files := slices.Clone(serviceFiles)
			files = append(files, plan.ClientFiles()...)
			files = append(files, plan.ClientTypeFiles()...)
			files = append(files, plan.ServerFiles()...)
			files = append(files, plan.ServerTypeFiles()...)
			files = append(files, plan.PathFiles()...)
			runGeneratedMixedSSECompile(t, files, "./gen/...")
		})
	}
}

// webSocketErrorRoot defines one service with a single WebSocket streaming
// endpoint, optionally mapping an error to an HTTP status.
func webSocketErrorRoot(t *testing.T, mapErr bool) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("watcher", func() {
			dsl.Method("watch", func() {
				dsl.Payload(func() {
					dsl.Attribute("id", dsl.String)
					dsl.Required("id")
				})
				dsl.StreamingResult(dsl.Int)
				if mapErr {
					dsl.Error("not_found")
				}
				dsl.HTTP(func() {
					dsl.GET("/watch/{id}")
					if mapErr {
						dsl.Response("not_found", dsl.StatusNotFound)
					}
				})
			})
		})
	})
}
