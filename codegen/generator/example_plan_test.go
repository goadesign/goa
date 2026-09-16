// This file checks that plugins receive a separate copy of the example server
// description retained for the exact prepared design root.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

func TestExampleReturnsSeparateCopyForExactRoot(t *testing.T) {
	root := &expr.RootExpr{}
	transport := &example.TransportData{
		Type:     example.TransportHTTP,
		Name:     "HTTP",
		Services: []string{"calc"},
	}
	variable := &example.VariableData{
		Name:         "version",
		Description:  "API version",
		DefaultValue: "v1",
		Values:       []string{"v1", "v2"},
	}
	planned := &example.Root{
		APIName:  "calc",
		Services: []string{"calc"},
		Servers: []*example.Data{{
			Name:        "edge",
			Description: "public server",
			Services:    []string{"calc"},
			Schemes:     []string{"http"},
			Variables:   []*example.VariableData{variable},
			Transports:  []*example.TransportData{transport},
			Hosts: []*example.HostData{{
				Name:      "development",
				Schemes:   []string{"http"},
				Variables: []*example.VariableData{variable},
				URIs: []*example.URIData{{
					URL:       "http://localhost/{version}",
					Scheme:    "http",
					Port:      "80",
					Transport: transport,
					HandlerArgs: []example.HandlerArg{{
						Service:  "calc",
						Endpoint: true,
					}},
				}},
			}},
		}},
	}
	plan := &Plan{example: []*examplePlanEntry{{source: root, root: planned}}}

	got, ok := plan.Example(root)
	require.True(t, ok)
	require.Equal(t, planned, got)
	require.NotSame(t, planned, got)
	require.NotSame(t, planned.Servers[0], got.Servers[0])
	require.NotSame(t, planned.Servers[0].Hosts[0], got.Servers[0].Hosts[0])
	require.NotSame(t, variable, got.Servers[0].Variables[0])
	require.Same(t, got.Servers[0].Variables[0], got.Servers[0].Hosts[0].Variables[0])
	require.NotSame(t, transport, got.Servers[0].Transports[0])
	require.Same(t, got.Servers[0].Transports[0], got.Servers[0].Hosts[0].URIs[0].Transport)

	got.Services[0] = "changed"
	got.Servers[0].Services[0] = "changed"
	got.Servers[0].Variables[0].Values[0] = "changed"
	got.Servers[0].Hosts[0].URIs[0].HandlerArgs[0].Service = "changed"
	require.Equal(t, "calc", planned.Services[0])
	require.Equal(t, "calc", planned.Servers[0].Services[0])
	require.Equal(t, "v1", variable.Values[0])
	require.Equal(t, "calc", planned.Servers[0].Hosts[0].URIs[0].HandlerArgs[0].Service)

	got, ok = plan.Example(&expr.RootExpr{})
	require.False(t, ok)
	require.Nil(t, got)
}
