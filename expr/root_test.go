// This file verifies root validation, including shared HTTP routes and
// exact-origin dependency traversal for explicitly relocated user types.
package expr

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"goa.design/goa/v3/eval"
)

type rootExternalType struct {
	Value string
}

func TestRelocatedDependenciesUseDeclarationOrigin(t *testing.T) {
	dependency := &UserTypeExpr{
		TypeName:      "Dependency",
		UID:           "shared-semantic-id",
		AttributeExpr: &AttributeExpr{Type: String},
	}
	relocated := &UserTypeExpr{
		TypeName: "Relocated",
		UID:      "shared-semantic-id",
		AttributeExpr: &AttributeExpr{
			Meta: MetaExpr{"struct:pkg:path": {"types"}},
			Type: &Object{&NamedAttributeExpr{
				Name:      "dependency",
				Attribute: &AttributeExpr{Type: dependency},
			}},
		},
	}
	root := &RootExpr{Types: []UserType{relocated, dependency}}

	errors := root.validateRelocatedUserTypes()
	if len(errors.Errors) != 1 {
		t.Fatalf("expected one relocated dependency error, got %d", len(errors.Errors))
	}
	if message := errors.Errors[0].Error(); !strings.Contains(message, "Dependency") {
		t.Errorf("expected dependency name in error, got %q", message)
	}
}

func TestRelocatedDependencyWalkStopsAtExactOriginCopy(t *testing.T) {
	relocated := &UserTypeExpr{
		TypeName: "Relocated",
		UID:      "relocated",
		AttributeExpr: &AttributeExpr{
			Meta: MetaExpr{"struct:pkg:path": {"types"}},
			Type: String,
		},
	}
	copy := relocated.Dup(DupAtt(relocated.Attribute()))
	relocated.AttributeExpr.Type = &Object{&NamedAttributeExpr{
		Name:      "self",
		Attribute: &AttributeExpr{Type: copy},
	}}
	root := &RootExpr{Types: []UserType{relocated}}

	errors := root.validateRelocatedUserTypes()
	if len(errors.Errors) != 0 {
		t.Errorf("expected exact origin copy to be treated as recursion, got %v", errors)
	}
}

func TestRootExprValidate(t *testing.T) {
	cases := map[string]struct {
		api      *APIExpr
		expected *eval.ValidationErrors
	}{
		"no error": {
			api: &APIExpr{
				Name: "foo",
			},
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"missing api declaration": {
			api: nil,
			expected: &eval.ValidationErrors{
				Errors: []error{fmt.Errorf("Missing API declaration")},
			},
		},
	}

	for k, tc := range cases {
		e := RootExpr{
			API: tc.api,
		}
		var actual *eval.ValidationErrors
		if errors.As(e.Validate(), &actual); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestRootExprValidateMixedHTTPRoutes(t *testing.T) {
	tests := []struct {
		name          string
		httpMethod    string
		httpPath      string
		jsonrpcMethod string
		jsonrpcPath   string
		servers       []*ServerExpr
		wantError     bool
	}{
		{
			name:          "same route on the default server",
			httpMethod:    "POST",
			httpPath:      "/tasks",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/tasks",
			wantError:     true,
		},
		{
			name:          "wildcards with different names",
			httpMethod:    "POST",
			httpPath:      "/tasks/{taskID}",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/tasks/{id}",
			wantError:     true,
		},
		{
			name:          "catch-all wildcards with different names",
			httpMethod:    "POST",
			httpPath:      "/assets/{*assetPath}",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/assets/{*path}",
			wantError:     true,
		},
		{
			name:          "different wildcard kinds",
			httpMethod:    "POST",
			httpPath:      "/assets/{asset}",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/assets/{*path}",
		},
		{
			name:          "different methods",
			httpMethod:    "GET",
			httpPath:      "/tasks",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/tasks",
		},
		{
			name:          "different paths",
			httpMethod:    "POST",
			httpPath:      "/tasks",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/rpc",
		},
		{
			name:          "same route on different servers",
			httpMethod:    "POST",
			httpPath:      "/tasks",
			jsonrpcMethod: "POST",
			jsonrpcPath:   "/tasks",
			servers: []*ServerExpr{
				{Name: "http", Services: []string{"http_service"}},
				{Name: "jsonrpc", Services: []string{"jsonrpc_service"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := mixedHTTPRoot(
				test.httpMethod,
				test.httpPath,
				test.jsonrpcMethod,
				test.jsonrpcPath,
				test.servers,
			)
			err := root.Validate()
			var validationErrors *eval.ValidationErrors
			errors.As(err, &validationErrors)
			if !test.wantError {
				if len(validationErrors.Errors) > 0 {
					t.Errorf("expected routes to coexist, got %v", err)
				}
				return
			}
			if len(validationErrors.Errors) == 0 {
				t.Errorf("expected conflicting routes to be rejected")
				return
			}
			for _, text := range []string{
				"server \"test\"",
				"ordinary HTTP route " + test.httpMethod,
				test.httpPath,
				"JSON-RPC route " + test.jsonrpcMethod,
				test.jsonrpcPath,
			} {
				if !strings.Contains(err.Error(), text) {
					t.Errorf("expected error to contain %q, got %v", text, err)
				}
			}
		})
	}
}

// TestRootExprValidateRejectsDuplicateTypeMappings catches two identical
// conversion or creation declarations that would emit the same receiver method.
func TestRootExprValidateRejectsDuplicateTypeMappings(t *testing.T) {
	user := &UserTypeExpr{
		TypeName:      "Value",
		UID:           "value",
		AttributeExpr: &AttributeExpr{Type: String},
	}
	for _, test := range []struct {
		name        string
		conversions []*TypeMap
		creations   []*TypeMap
	}{
		{
			name: "conversion",
			conversions: []*TypeMap{
				{User: user, External: rootExternalType{}},
				{User: user, External: rootExternalType{}},
			},
		},
		{
			name: "creation",
			creations: []*TypeMap{
				{User: user, External: rootExternalType{}},
				{User: user, External: rootExternalType{}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := &RootExpr{
				API:         &APIExpr{Name: "test"},
				Types:       []UserType{user},
				Conversions: test.conversions,
				Creations:   test.creations,
			}
			err := root.Validate()
			if err == nil || !strings.Contains(err.Error(), test.name+" from") || !strings.Contains(err.Error(), "defined twice") {
				t.Fatalf("expected precise duplicate %s error, got %v", test.name, err)
			}
		})
	}
}

func TestMetaExpr_Last(t *testing.T) {
	tt := map[string]struct {
		meta  MetaExpr
		value string
		ok    bool
	}{
		"no-key": {
			MetaExpr{},
			"",
			false,
		},
		"key-no-values": {
			MetaExpr{
				"test:key": []string{},
			},
			"",
			false,
		},
		"key-with-one-value": {
			MetaExpr{
				"test:key": []string{
					"value-one",
				},
			},
			"value-one",
			true,
		},
		"key-with-multiple-values": {
			MetaExpr{
				"test:key": []string{
					"value-one",
					"value-two",
					"value-n",
				},
			},
			"value-n",
			true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			value, ok := tc.meta.Last("test:key")
			if tc.ok != ok {
				t.Errorf("expected ok to be %v, got %v", tc.ok, ok)
			}
			if tc.value != value {
				t.Errorf("expected value to be %s, got %s", value, value)
			}
		})
	}
}

// mixedHTTPRoot returns a design with one ordinary HTTP route and one
// JSON-RPC route.
func mixedHTTPRoot(
	httpMethod, httpPath, jsonrpcMethod, jsonrpcPath string,
	servers []*ServerExpr,
) *RootExpr {
	api := NewAPIExpr("test", func() {})
	api.Servers = servers
	root := &RootExpr{API: api}

	addMixedHTTPRoute(root, api.HTTP, "http_service", httpMethod, httpPath)
	addMixedHTTPRoute(root, &api.JSONRPC.HTTPExpr, "jsonrpc_service", jsonrpcMethod, jsonrpcPath)
	return root
}

// addMixedHTTPRoute adds one service method to transport for route validation.
func addMixedHTTPRoute(root *RootExpr, transport *HTTPExpr, serviceName, method, routePath string) {
	service := &ServiceExpr{Name: serviceName}
	methodExpr := &MethodExpr{
		Name:    "run",
		Payload: &AttributeExpr{Type: Empty},
		Result:  &AttributeExpr{Type: Empty},
		Service: service,
		Stream:  NoStreamKind,
	}
	service.Methods = []*MethodExpr{methodExpr}
	root.Services = append(root.Services, service)
	httpService := transport.ServiceFor(service, transport)
	endpoint := httpService.EndpointFor(methodExpr)
	endpoint.Routes = []*RouteExpr{{
		Method:   method,
		Path:     routePath,
		Endpoint: endpoint,
	}}
}
