// This file checks that services added by plugins receive every declaration
// needed by their generated JSON-RPC server-sent-event code.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestAttachedJSONRPCSSEServiceCompiles runs every generation step after a
// plugin adds one method that returns a value and another that streams values.
func TestAttachedJSONRPCSSEServiceCompiles(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("attached-stream", func() {})
	})
	registry := newDefaultRegistry()
	registry.registerPlugin("attached-stream", "gen", pluginNormal, func() Plugin {
		return Plugin{Prepare: func(_ string, roots []eval.Root) error {
			return attachJSONRPCSSEService(roots[0].(*expr.RootExpr))
		}}
	})
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	files, err := mergeFilesByPath(result.files)
	require.NoError(t, err)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	serviceCode, err := os.ReadFile(filepath.Join(dir, "gen", "attached_stream", "service.go"))
	require.NoError(t, err)
	generated := string(serviceCode)
	require.Contains(t, generated, "Watch(context.Context, WatchServerStream) (err error)")
	require.Contains(t, generated, "Send(string) error")
	require.Contains(t, generated, "SendWithContext(context.Context, string) error")
	require.Contains(t, generated, "Close() error")
	require.NotContains(t, generated, "SendAndClose")
	require.NotContains(t, generated, "SendError")
	require.NotContains(t, generated, "RequestID")
	require.NotContains(t, generated, "isWatchEvent")
	require.NotContains(t, generated, "Send(ctx context.Context")
	runGeneratedTests(t, dir)
}

// attachJSONRPCSSEService adds one evaluated service and its JSON-RPC route to
// a design that completed DSL evaluation before the plugin ran.
func attachJSONRPCSSEService(root *expr.RootExpr) error {
	read := &expr.MethodExpr{
		Name:    "Read",
		Payload: &expr.AttributeExpr{Type: expr.Empty},
		Result:  &expr.AttributeExpr{Type: expr.String},
		Meta:    expr.MetaExpr{"jsonrpc": []string{}},
	}
	watch := &expr.MethodExpr{
		Name:            "Watch",
		Payload:         &expr.AttributeExpr{Type: expr.Empty},
		StreamingResult: &expr.AttributeExpr{Type: expr.String},
		Stream:          expr.ServerStreamKind,
		Meta:            expr.MetaExpr{"jsonrpc": []string{}},
	}
	service := &expr.ServiceExpr{
		Name:    "AttachedStream",
		Methods: []*expr.MethodExpr{read, watch},
		Meta:    expr.MetaExpr{"jsonrpc:service": []string{}},
	}
	read.Service = service
	watch.Service = service

	transport := &expr.HTTPServiceExpr{
		ServiceExpr: service,
		JSONRPCRoute: &expr.RouteExpr{
			Method: "POST",
			Path:   "/rpc",
		},
		SSE: &expr.HTTPSSEExpr{},
	}
	transport.Root = &root.API.JSONRPC.HTTPExpr
	transport.JSONRPCRoute.Endpoint = &expr.HTTPEndpointExpr{Service: transport}
	for _, method := range service.Methods {
		endpoint := &expr.HTTPEndpointExpr{
			MethodExpr: method,
			Service:    transport,
			Body:       method.Payload,
			Params:     expr.NewEmptyMappedAttributeExpr(),
			Headers:    expr.NewEmptyMappedAttributeExpr(),
			Cookies:    expr.NewEmptyMappedAttributeExpr(),
			Meta:       expr.MetaExpr{"jsonrpc": []string{}},
		}
		if method.IsResultStreaming() {
			endpoint.SSE = &expr.HTTPSSEExpr{}
		}
		endpoint.Routes = []*expr.RouteExpr{{Method: "POST", Path: "/rpc", Endpoint: endpoint}}
		transport.HTTPEndpoints = append(transport.HTTPEndpoints, endpoint)
	}

	root.Services = append(root.Services, service)
	root.API.JSONRPC.Services = append(root.API.JSONRPC.Services, transport)
	return root.EvaluateAttachedServices([]*expr.ServiceExpr{service})
}
