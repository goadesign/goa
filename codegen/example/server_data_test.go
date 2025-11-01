package example

import (
	"testing"

	"goa.design/goa/v3/expr"
)

func TestComputeHandlerArgsForURI_JSONRPCOrdering(t *testing.T) {
	method := &expr.MethodExpr{Name: "Run"}
	httpSvc := &expr.HTTPServiceExpr{
		ServiceExpr: &expr.ServiceExpr{
			Name:    "orchestrator",
			Methods: []*expr.MethodExpr{method},
		},
		HTTPEndpoints: []*expr.HTTPEndpointExpr{{MethodExpr: method}},
	}
	mcpMethod := &expr.MethodExpr{Name: "ListTools"}
	jsonrpcOrchestrator := &expr.HTTPServiceExpr{ServiceExpr: &expr.ServiceExpr{Name: "orchestrator"}}
	jsonrpcMCPAssistant := &expr.HTTPServiceExpr{
		ServiceExpr: &expr.ServiceExpr{
			Name:    "mcp_assistant",
			Methods: []*expr.MethodExpr{mcpMethod},
		},
		HTTPEndpoints: []*expr.HTTPEndpointExpr{{MethodExpr: mcpMethod}},
	}
	root := &expr.RootExpr{
		API: &expr.APIExpr{
			HTTP: &expr.HTTPExpr{
				Services: []*expr.HTTPServiceExpr{httpSvc},
			},
			JSONRPC: &expr.JSONRPCExpr{
				HTTPExpr: expr.HTTPExpr{
					Services: []*expr.HTTPServiceExpr{jsonrpcOrchestrator, jsonrpcMCPAssistant},
				},
			},
		},
		Services: []*expr.ServiceExpr{
			{Name: "orchestrator", Methods: []*expr.MethodExpr{method}},
			{Name: "mcp_assistant", Methods: []*expr.MethodExpr{mcpMethod}},
		},
	}
	server := &Data{
		Services: []string{"orchestrator", "mcp_assistant"},
		Transports: []*TransportData{
			{Type: TransportHTTP, Services: []string{"orchestrator", "mcp_assistant"}},
		},
	}
	uri := &URIData{Transport: &TransportData{Type: TransportHTTP}}

	args := computeHandlerArgsForURI(uri, server, root)

	want := []HandlerArg{
		{Endpoint: "orchestratorEndpoints"},
		{Service: "orchestratorSvc"},
		{Service: "mcpAssistantSvc"},
		{Endpoint: "mcpAssistantEndpoints"},
	}
	if len(args) != len(want) {
		t.Fatalf("expected %d handler args, got %d (%v)", len(want), len(args), args)
	}
	for i, arg := range want {
		if args[i] != arg {
			t.Fatalf("handler arg %d: expected %+v, got %+v", i, arg, args[i])
		}
	}
}
