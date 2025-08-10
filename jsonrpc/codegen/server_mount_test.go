package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func TestServerMount_MixedTransport(t *testing.T) {
	// When a service has both regular JSON-RPC and SSE endpoints (mixed transport),
	// the Mount function should mount ServeHTTP, not handleSSE
	
	// Setup
	svc := &expr.ServiceExpr{Name: "TestService"}
	
	regularMethod := &expr.MethodExpr{
		Name:    "Regular",
		Service: svc,
	}
	
	streamingMethod := &expr.MethodExpr{
		Name:    "Streaming",
		Service: svc,
		Stream:  expr.ServerStreamKind,
	}
	
	httpSvc := &expr.HTTPServiceExpr{
		ServiceExpr: svc,
	}
	
	httpSvc.HTTPEndpoints = []*expr.HTTPEndpointExpr{
		{
			MethodExpr: regularMethod,
			Service:    httpSvc,
			Body:       &expr.AttributeExpr{Type: expr.String},
			Meta:       expr.MetaExpr{"jsonrpc": []string{}},
		},
		{
			MethodExpr: streamingMethod,
			Service:    httpSvc,
			Body:       &expr.AttributeExpr{Type: expr.String},
			SSE:        &expr.HTTPSSEExpr{},
			Meta:       expr.MetaExpr{"jsonrpc": []string{}},
		},
	}
	
	sd := &httpcodegen.ServiceData{
		Service: &service.Data{
			Name:     "TestService",
			VarName:  "testservice",
			PkgName:  "testservice",
			PathName: "testservice",
			Scope:    codegen.NewNameScope(),
		},
		ServerStruct: "Server",
		MountServer:  "Mount",
		Endpoints: []*httpcodegen.EndpointData{
			{
				Routes: []*httpcodegen.RouteData{
					{Verb: "POST", Path: "/rpc"},
				},
			},
		},
	}
	
	services := &httpcodegen.ServicesData{
		ServicesData: &service.ServicesData{
			Services: map[string]*service.Data{
				"TestService": sd.Service,
			},
		},
		HTTPData: map[string]*httpcodegen.ServiceData{
			"TestService": sd,
		},
	}
	
	// Test
	hasMixed := hasMixedJSONRPCTransports(httpSvc, services)
	hasSSE := hasJSONRPCSSE(httpSvc, services)
	
	assert.True(t, hasMixed, "should detect mixed transport")
	assert.True(t, hasSSE, "should detect SSE")
	
	mountData := struct {
		*httpcodegen.ServiceData
		HasSSE   bool
		HasMixed bool
	}{
		ServiceData: sd,
		HasSSE:      hasSSE,
		HasMixed:    hasMixed,
	}
	
	section := &codegen.SectionTemplate{
		Name:   "server-mount",
		Source: jsonrpcTemplates.Read(serverMountT),
		Data:   mountData,
	}
	
	var buf bytes.Buffer
	err := section.Write(&buf)
	require.NoError(t, err, "template should render without error")
	
	code := buf.String()
	
	// Verify
	assert.Contains(t, code, "h.ServeHTTP", "mixed transport should mount ServeHTTP")
	assert.NotContains(t, code, "h.handleSSE", "mixed transport should not mount handleSSE")
	assert.Contains(t, code, "Mount ServeHTTP which handles both regular JSON-RPC and SSE", "should include explanatory comment")
}

func TestServerMount_SSEOnly(t *testing.T) {
	// When a service has only SSE endpoints (no regular JSON-RPC),
	// the Mount function should mount handleSSE
	
	// Setup
	svc := &expr.ServiceExpr{Name: "TestService"}
	
	streamingMethod := &expr.MethodExpr{
		Name:    "Streaming",
		Service: svc,
		Stream:  expr.ServerStreamKind,
	}
	
	httpSvc := &expr.HTTPServiceExpr{
		ServiceExpr: svc,
	}
	
	httpSvc.HTTPEndpoints = []*expr.HTTPEndpointExpr{
		{
			MethodExpr: streamingMethod,
			Service:    httpSvc,
			Body:       &expr.AttributeExpr{Type: expr.String},
			SSE:        &expr.HTTPSSEExpr{},
			Meta:       expr.MetaExpr{"jsonrpc": []string{}},
		},
	}
	
	sd := &httpcodegen.ServiceData{
		Service: &service.Data{
			Name:     "TestService",
			VarName:  "testservice",
			PkgName:  "testservice",
			PathName: "testservice",
			Scope:    codegen.NewNameScope(),
		},
		ServerStruct: "Server",
		MountServer:  "Mount",
		Endpoints: []*httpcodegen.EndpointData{
			{
				Routes: []*httpcodegen.RouteData{
					{Verb: "POST", Path: "/rpc"},
				},
			},
		},
	}
	
	services := &httpcodegen.ServicesData{
		ServicesData: &service.ServicesData{
			Services: map[string]*service.Data{
				"TestService": sd.Service,
			},
		},
		HTTPData: map[string]*httpcodegen.ServiceData{
			"TestService": sd,
		},
	}
	
	// Test
	hasMixed := hasMixedJSONRPCTransports(httpSvc, services)
	hasSSE := hasJSONRPCSSE(httpSvc, services)
	
	assert.False(t, hasMixed, "should not detect mixed transport for SSE-only")
	assert.True(t, hasSSE, "should detect SSE")
	
	mountData := struct {
		*httpcodegen.ServiceData
		HasSSE   bool
		HasMixed bool
	}{
		ServiceData: sd,
		HasSSE:      hasSSE,
		HasMixed:    hasMixed,
	}
	
	section := &codegen.SectionTemplate{
		Name:   "server-mount",
		Source: jsonrpcTemplates.Read(serverMountT),
		Data:   mountData,
	}
	
	var buf bytes.Buffer
	err := section.Write(&buf)
	require.NoError(t, err, "template should render without error")
	
	code := buf.String()
	
	// Verify
	assert.Contains(t, code, "h.handleSSE", "SSE-only should mount handleSSE")
	assert.NotContains(t, code, "h.ServeHTTP", "SSE-only should not mount ServeHTTP")
	assert.Contains(t, code, "Mount SSE handler for all endpoint routes", "should include SSE comment")
}

func TestServerMount_RegularOnly(t *testing.T) {
	// When a service has only regular JSON-RPC endpoints (no SSE),
	// the Mount function should mount ServeHTTP
	
	// Setup
	svc := &expr.ServiceExpr{Name: "TestService"}
	
	regularMethod := &expr.MethodExpr{
		Name:    "Regular",
		Service: svc,
	}
	
	httpSvc := &expr.HTTPServiceExpr{
		ServiceExpr: svc,
	}
	
	httpSvc.HTTPEndpoints = []*expr.HTTPEndpointExpr{
		{
			MethodExpr: regularMethod,
			Service:    httpSvc,
			Body:       &expr.AttributeExpr{Type: expr.String},
			Meta:       expr.MetaExpr{"jsonrpc": []string{}},
		},
	}
	
	sd := &httpcodegen.ServiceData{
		Service: &service.Data{
			Name:     "TestService",
			VarName:  "testservice",
			PkgName:  "testservice",
			PathName: "testservice",
			Scope:    codegen.NewNameScope(),
		},
		ServerStruct: "Server",
		MountServer:  "Mount",
		Endpoints: []*httpcodegen.EndpointData{
			{
				Routes: []*httpcodegen.RouteData{
					{Verb: "POST", Path: "/rpc"},
				},
			},
		},
	}
	
	services := &httpcodegen.ServicesData{
		ServicesData: &service.ServicesData{
			Services: map[string]*service.Data{
				"TestService": sd.Service,
			},
		},
		HTTPData: map[string]*httpcodegen.ServiceData{
			"TestService": sd,
		},
	}
	
	// Test
	hasMixed := hasMixedJSONRPCTransports(httpSvc, services)
	hasSSE := hasJSONRPCSSE(httpSvc, services)
	
	assert.False(t, hasMixed, "should not detect mixed transport")
	assert.False(t, hasSSE, "should not detect SSE")
	
	mountData := struct {
		*httpcodegen.ServiceData
		HasSSE   bool
		HasMixed bool
	}{
		ServiceData: sd,
		HasSSE:      hasSSE,
		HasMixed:    hasMixed,
	}
	
	section := &codegen.SectionTemplate{
		Name:   "server-mount",
		Source: jsonrpcTemplates.Read(serverMountT),
		Data:   mountData,
	}
	
	var buf bytes.Buffer
	err := section.Write(&buf)
	require.NoError(t, err, "template should render without error")
	
	code := buf.String()
	
	// Verify
	assert.Contains(t, code, "h.ServeHTTP", "regular-only should mount ServeHTTP")
	assert.NotContains(t, code, "h.handleSSE", "regular-only should not mount handleSSE")
}