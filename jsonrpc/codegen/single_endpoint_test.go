package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestJSONRPCSingleEndpoint(t *testing.T) {
	t.Run("service-level JSONRPC", func(t *testing.T) {
		root := RunJSONRPCDSL(t, func() {
			dsl.Service("calc", func() {
				dsl.JSONRPC(func() {
					dsl.POST("/rpc")
				})

				dsl.Method("add", func() {
					dsl.Payload(func() {
						dsl.ID("id")
						dsl.Attribute("a", dsl.Int)
						dsl.Attribute("b", dsl.Int)
					})
					dsl.Result(func() {
						dsl.ID("id")
						dsl.Attribute("sum", dsl.Int)
					})
					dsl.JSONRPC(func() {})
				})

				dsl.Method("subtract", func() {
					dsl.Payload(func() {
						dsl.ID("id")
						dsl.Attribute("a", dsl.Int)
						dsl.Attribute("b", dsl.Int)
					})
					dsl.Result(func() {
						dsl.ID("id")
						dsl.Attribute("diff", dsl.Int)
					})
					dsl.JSONRPC(func() {})
				})
			})
		})

		// Check service is marked as JSON-RPC
		svc := root.Service("calc")
		require.NotNil(t, svc)
		require.NotNil(t, svc.Meta)
		assert.NotNil(t, svc.Meta["jsonrpc:service"], "service should be marked as JSON-RPC")

		// Check HTTP service
		httpSvc := root.API.JSONRPC.Service("calc")
		require.NotNil(t, httpSvc)
		
		// Prepare the service to create routes
		httpSvc.Prepare()

		// Check that all JSON-RPC endpoints have exactly one route with the same path
		var firstRoute *expr.RouteExpr
		jsonrpcEndpointCount := 0
		for _, e := range httpSvc.HTTPEndpoints {
			if e.IsJSONRPC() {
				jsonrpcEndpointCount++
				assert.Equal(t, 1, len(e.Routes), "each JSON-RPC endpoint should have exactly one route")
				if len(e.Routes) > 0 {
					if firstRoute == nil {
						firstRoute = e.Routes[0]
					} else {
						// Verify all endpoints share the same route configuration
						assert.Equal(t, firstRoute.Path, e.Routes[0].Path, "all JSON-RPC endpoints should share the same path")
						assert.Equal(t, firstRoute.Method, e.Routes[0].Method, "all JSON-RPC endpoints should share the same HTTP method")
					}
				}
			}
		}
		
		assert.Greater(t, jsonrpcEndpointCount, 0, "should have at least one JSON-RPC endpoint")
		assert.NotNil(t, firstRoute, "should have found a route")
		assert.Equal(t, "/rpc", firstRoute.Path, "route path should be /rpc")
		assert.Equal(t, "POST", firstRoute.Method, "route method should be POST")
	})

	t.Run("method-level JSONRPC auto-enables service", func(t *testing.T) {
		root := RunJSONRPCDSL(t, func() {
			dsl.Service("calc2", func() {
				// No service-level JSONRPC

				dsl.Method("multiply", func() {
					dsl.Payload(func() {
						dsl.ID("id")
						dsl.Attribute("a", dsl.Int)
						dsl.Attribute("b", dsl.Int)
					})
					dsl.Result(func() {
						dsl.ID("id")
						dsl.Attribute("product", dsl.Int)
					})
					dsl.JSONRPC(func() {}) // Should auto-enable service
				})
			})
		})

		// Check service is auto-marked as JSON-RPC
		svc := root.Service("calc2")
		require.NotNil(t, svc)
		require.NotNil(t, svc.Meta)
		assert.NotNil(t, svc.Meta["jsonrpc:service"], "service should be auto-marked as JSON-RPC")
	})

	t.Run("WebSocket forces GET", func(t *testing.T) {
		root := RunJSONRPCDSL(t, func() {
			dsl.Service("stream", func() {
				dsl.JSONRPC(func() {})

				dsl.Method("echo", func() {
					dsl.StreamingPayload(func() {
						dsl.ID("id")
						dsl.Attribute("msg", dsl.String)
					})
					dsl.StreamingResult(func() {
						dsl.ID("id")
						dsl.Attribute("echo", dsl.String)
					})
					dsl.JSONRPC(func() {})
				})
			})
		})

		// Check route method
		httpSvc := root.API.JSONRPC.Service("stream")
		require.NotNil(t, httpSvc)
		
		// Prepare the service to create routes
		httpSvc.Prepare()
		
		// Find first endpoint with route
		var route *expr.RouteExpr
		for _, e := range httpSvc.HTTPEndpoints {
			if e.IsJSONRPC() && len(e.Routes) > 0 {
				route = e.Routes[0]
				break
			}
		}
		
		require.NotNil(t, route)
		assert.Equal(t, "GET", route.Method, "WebSocket should force GET method")
	})
}