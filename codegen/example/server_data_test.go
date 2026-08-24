package example

import (
	"testing"

	"github.com/stretchr/testify/require"

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
	jsonrpcUnhosted := &expr.HTTPServiceExpr{
		ServiceExpr: &expr.ServiceExpr{
			Name:    "unhosted",
			Methods: []*expr.MethodExpr{{Name: "Ignore"}},
		},
		HTTPEndpoints: []*expr.HTTPEndpointExpr{{MethodExpr: &expr.MethodExpr{Name: "Ignore"}}},
	}
	root := &expr.RootExpr{
		API: &expr.APIExpr{
			HTTP: &expr.HTTPExpr{
				Services: []*expr.HTTPServiceExpr{httpSvc},
			},
			JSONRPC: &expr.JSONRPCExpr{
				HTTPExpr: expr.HTTPExpr{
					Services: []*expr.HTTPServiceExpr{jsonrpcOrchestrator, jsonrpcMCPAssistant, jsonrpcUnhosted},
				},
			},
		},
		Services: []*expr.ServiceExpr{
			{Name: "orchestrator", Methods: []*expr.MethodExpr{method}},
			{Name: "mcp_assistant", Methods: []*expr.MethodExpr{mcpMethod}},
			{Name: "unhosted", Methods: []*expr.MethodExpr{{Name: "Ignore"}}},
		},
	}
	server := &Data{
		Services: []string{"orchestrator", "mcp_assistant"},
		Transports: []*TransportData{
			{Type: TransportHTTP, Services: []string{"orchestrator", "mcp_assistant"}},
		},
	}
	uri := &URIData{Transport: &TransportData{Type: TransportHTTP}}

	args := planHandlerArgsForURI(uri, server, root)

	want := []HandlerArg{
		{Service: "orchestrator", Endpoint: true},
		{Service: "orchestrator"},
		{Service: "mcp_assistant"},
		{Service: "mcp_assistant", Endpoint: true},
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

// TestPlanHandlerArgsForJSONRPCOnlyServer checks that the generated main and
// HTTP helper can use the same service-by-service argument order.
func TestPlanHandlerArgsForJSONRPCOnlyServer(t *testing.T) {
	firstMethod := &expr.MethodExpr{Name: "First"}
	secondMethod := &expr.MethodExpr{Name: "Second"}
	first := &expr.ServiceExpr{Name: "first", Methods: []*expr.MethodExpr{firstMethod}}
	second := &expr.ServiceExpr{Name: "second", Methods: []*expr.MethodExpr{secondMethod}}
	root := &expr.RootExpr{
		API: &expr.APIExpr{
			HTTP: &expr.HTTPExpr{},
			JSONRPC: &expr.JSONRPCExpr{HTTPExpr: expr.HTTPExpr{Services: []*expr.HTTPServiceExpr{
				{
					ServiceExpr:   first,
					HTTPEndpoints: []*expr.HTTPEndpointExpr{{MethodExpr: firstMethod}},
				},
				{
					ServiceExpr:   second,
					HTTPEndpoints: []*expr.HTTPEndpointExpr{{MethodExpr: secondMethod}},
				},
			}}},
		},
		Services: []*expr.ServiceExpr{first, second},
	}
	server := &Data{
		Services: []string{"first", "second"},
		Transports: []*TransportData{{
			Type: TransportHTTP,
		}},
	}
	uri := &URIData{Transport: &TransportData{Type: TransportHTTP}}

	require.Equal(t, []HandlerArg{
		{Service: "first"},
		{Service: "first", Endpoint: true},
		{Service: "second"},
		{Service: "second", Endpoint: true},
	}, planHandlerArgsForURI(uri, server, root))
}
