// This file checks services that generators add after the design DSL has
// finished running.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluateAttachedServicesUsesOwningRoot(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	other := newRootExprForTest("other_header")
	Root = other

	service, _, endpoint := attachTestService(owner, "generated")
	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))

	headers := AsObject(endpoint.Headers.Type)
	require.NotNil(t, headers.Attribute("owner_header"))
	require.Nil(t, headers.Attribute("other_header"))
}

func TestEvaluateAttachedServicesDoesNotReadPackageRoot(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	service, _, endpoint := attachTestService(owner, "generated")
	Root = nil

	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))
	require.NotNil(t, AsObject(endpoint.Headers.Type).Attribute("owner_header"))
}

func TestEvaluateAttachedGRPCServiceUsesOwningRootForAPIErrors(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	service := attachTestGRPCServiceWithAPIError(owner, "generated")
	Root = nil

	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))
}

func TestEvaluateAttachedGRPCServiceIgnoresPackageRootAPIErrors(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	service := attachTestGRPCServiceWithAPIError(owner, "generated")
	other := newRootExprForTest("other_header")
	other.Errors = append(other.Errors, &ErrorExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		Name:          "failed",
	})
	Root = other

	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))
}

func TestEvaluateAttachedServiceFinishesFileServerWithOwningRoot(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	owner.API.HTTP.Path = "/owner"
	service, transport, _ := attachTestService(owner, "generated")
	transport.Paths = []string{"/generated"}
	fileServer := &HTTPFileServerExpr{
		Service:      transport,
		FilePath:     "./public",
		RequestPaths: []string{"/assets/{*path}"},
	}
	transport.FileServers = append(transport.FileServers, fileServer)
	Root = nil

	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))
	require.Equal(t, []string{"/owner/generated/assets/{*path}"}, fileServer.RequestPaths)
}

func TestEvaluateAttachedServiceIgnoresPackageRootForFileServer(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	owner := newRootExprForTest("owner_header")
	owner.API.HTTP.Path = "/owner"
	service, transport, _ := attachTestService(owner, "generated")
	transport.Paths = []string{"/generated"}
	fileServer := &HTTPFileServerExpr{
		Service:      transport,
		FilePath:     "./public",
		RequestPaths: []string{"/assets/{*path}"},
	}
	transport.FileServers = append(transport.FileServers, fileServer)
	other := newRootExprForTest("other_header")
	other.API.HTTP.Path = "/other"
	Root = other

	require.NoError(t, owner.EvaluateAttachedServices([]*ServiceExpr{service}))
	require.Equal(t, []string{"/owner/generated/assets/{*path}"}, fileServer.RequestPaths)
}

func TestEvaluateAttachedServicesChecksAllBeforeFinishing(t *testing.T) {
	root := newRootExprForTest("owner_header")
	first, firstTransport, _ := attachTestService(root, "first")
	second, _, _ := attachTestService(root, "second")
	second.Methods[0].Payload.Validation = &ValidationExpr{Required: []string{"missing"}}

	err := root.EvaluateAttachedServices([]*ServiceExpr{first, second})

	require.Error(t, err)
	require.Empty(t, firstTransport.Paths)
}

func TestEvaluateAttachedServicesRejectsMixedHTTPRouteCollision(t *testing.T) {
	root := newRootExprForTest("owner_header")
	_, _, httpEndpoint := attachTestService(root, "ordinary")
	httpEndpoint.Routes[0].Path = "/tasks/{taskID}"
	jsonrpcService := attachTestJSONRPCService(root, "generated", "/tasks/{id}")

	err := root.EvaluateAttachedServices([]*ServiceExpr{jsonrpcService})

	require.ErrorContains(t, err, "ordinary HTTP route POST \"/tasks/{taskID}\"")
	require.ErrorContains(t, err, "JSON-RPC route POST \"/tasks/{id}\"")
}

func TestEvaluateAttachedServicesRunsDifferentRootsTogether(t *testing.T) {
	originalRoot := Root
	t.Cleanup(func() {
		Root = originalRoot
	})
	firstRoot := newRootExprForTest("first_header")
	firstService, _, firstEndpoint := attachTestService(firstRoot, "first")
	secondRoot := newRootExprForTest("second_header")
	secondService, _, secondEndpoint := attachTestService(secondRoot, "second")
	Root = nil

	errors := make(chan error, 2)
	go func() {
		errors <- firstRoot.EvaluateAttachedServices([]*ServiceExpr{firstService})
	}()
	go func() {
		errors <- secondRoot.EvaluateAttachedServices([]*ServiceExpr{secondService})
	}()
	require.NoError(t, <-errors)
	require.NoError(t, <-errors)

	require.NotNil(t, AsObject(firstEndpoint.Headers.Type).Attribute("first_header"))
	require.Nil(t, AsObject(firstEndpoint.Headers.Type).Attribute("second_header"))
	require.NotNil(t, AsObject(secondEndpoint.Headers.Type).Attribute("second_header"))
	require.Nil(t, AsObject(secondEndpoint.Headers.Type).Attribute("first_header"))
}

// newRootExprForTest returns a design with one API header.
func newRootExprForTest(header string) *RootExpr {
	api := NewAPIExpr("test", func() {})
	obj := &Object{}
	obj.Set(header, &AttributeExpr{Type: String})
	api.HTTP.Headers = NewMappedAttributeExpr(&AttributeExpr{Type: obj})
	return &RootExpr{API: api}
}

// attachTestService adds one HTTP service whose payload contains the API
// header used by the test.
func attachTestService(root *RootExpr, name string) (*ServiceExpr, *HTTPServiceExpr, *HTTPEndpointExpr) {
	payload := &Object{}
	for _, header := range *AsObject(root.API.HTTP.Headers.Type) {
		payload.Set(header.Name, header.Attribute)
	}
	service := &ServiceExpr{Name: name}
	method := &MethodExpr{
		Name:    "run",
		Payload: &AttributeExpr{Type: payload},
		Result:  &AttributeExpr{Type: Empty},
		Service: service,
		Stream:  NoStreamKind,
	}
	service.Methods = []*MethodExpr{method}
	root.Services = append(root.Services, service)
	transport := root.API.HTTP.ServiceFor(service, root.API.HTTP)
	endpoint := transport.EndpointFor(method)
	endpoint.Routes = []*RouteExpr{{
		Method:   "POST",
		Path:     "/run",
		Endpoint: endpoint,
	}}
	return service, transport, endpoint
}

// attachTestJSONRPCService adds one JSON-RPC method at routePath.
func attachTestJSONRPCService(root *RootExpr, name, routePath string) *ServiceExpr {
	service := &ServiceExpr{Name: name}
	method := &MethodExpr{
		Name:    "run",
		Payload: &AttributeExpr{Type: Empty},
		Result:  &AttributeExpr{Type: Empty},
		Service: service,
		Stream:  NoStreamKind,
	}
	service.Methods = []*MethodExpr{method}
	root.Services = append(root.Services, service)
	transport := root.API.JSONRPC.ServiceFor(service, &root.API.JSONRPC.HTTPExpr)
	endpoint := transport.EndpointFor(method)
	endpoint.Routes = []*RouteExpr{{
		Method:   "POST",
		Path:     routePath,
		Endpoint: endpoint,
	}}
	return service
}

// attachTestGRPCServiceWithAPIError adds one gRPC method that uses an error
// response defined for the complete API.
func attachTestGRPCServiceWithAPIError(root *RootExpr, name string) *ServiceExpr {
	apiError := &ErrorExpr{
		AttributeExpr: &AttributeExpr{Type: ErrorResult},
		Name:          "failed",
	}
	root.Errors = append(root.Errors, apiError)
	response := &GRPCResponseExpr{
		StatusCode: 13,
		Parent:     root.API.GRPC,
	}
	response.Prepare()
	root.API.GRPC.Errors = append(root.API.GRPC.Errors, &GRPCErrorExpr{
		Name:     "failed",
		Response: response,
	})

	service := &ServiceExpr{Name: name}
	method := &MethodExpr{
		Name:    "run",
		Payload: &AttributeExpr{Type: Empty},
		Result:  &AttributeExpr{Type: Empty},
		Errors: []*ErrorExpr{{
			AttributeExpr: &AttributeExpr{Type: ErrorResult},
			Name:          "failed",
		}},
		Service: service,
		Stream:  NoStreamKind,
	}
	service.Methods = []*MethodExpr{method}
	root.Services = append(root.Services, service)
	transport := root.API.GRPC.ServiceFor(service)
	transport.EndpointFor(method.Name, method)
	return service
}
