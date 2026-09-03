// This file verifies gRPC endpoint preparation and validation, including the
// native primitive contract required by request and response metadata.
package expr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestGRPCEndpointValidation(t *testing.T) {
	cases := map[string]struct {
		DSL    func()
		Errors []string
	}{
		"endpoint-with-any-type": {
			DSL:    testdata.GRPCEndpointWithAnyType,
			Errors: []string{}, // Any type is now supported in gRPC
		},
		"endpoint-with-mixed-results": {
			DSL: testdata.GRPCEndpointWithMixedResults,
			Errors: []string{
				`service "Service" gRPC endpoint "Method": gRPC method "Method" cannot define both Result and StreamingResult because one gRPC call cannot return a separate result after its response stream`,
			},
		},
		"endpoint-with-matching-mixed-results": {
			DSL: testdata.GRPCEndpointWithMatchingMixedResults,
			Errors: []string{
				`service "Service" gRPC endpoint "Method": gRPC method "Method" cannot define both Result and StreamingResult because one gRPC call cannot return a separate result after its response stream`,
			},
		},
		"endpoint-with-streaming-result": {
			DSL:    testdata.GRPCEndpointWithStreamingResult,
			Errors: []string{},
		},
		"endpoint-with-untagged-fields": {
			DSL: testdata.GRPCEndpointWithUntaggedFields,
			Errors: []string{`service "Service" gRPC endpoint "Method": attribute "req_not_field" does not have "rpc:tag" defined in the meta, use "Field" to define the attribute of a type used in a gRPC method
service "Service" gRPC endpoint "Method": attribute "resp_not_field" does not have "rpc:tag" defined in the meta, use "Field" to define the attribute of a type used in a gRPC method`,
			},
		},
		"endpoint-with-repeated-field-tags": {
			DSL: testdata.GRPCEndpointWithRepeatedFieldTags,
			Errors: []string{`service "Service" gRPC endpoint "Method": field number 1 in attribute "key_dup_id" already exists for attribute "key"
service "Service" gRPC endpoint "Method": field number 2 in attribute "key_dup_id" already exists for attribute "key"`,
			},
		},
		"endpoint-with-reference-types-field-inheritance": {
			DSL:    testdata.GRPCEndpointWithReferenceTypes,
			Errors: []string{},
		},
		"endpoint-with-extended-types": {
			DSL:    testdata.GRPCEndpointWithExtendedTypes,
			Errors: []string{},
		},
		"endpoint-with-composite-metadata": {
			DSL: testdata.GRPCEndpointWithCompositeMetadata,
			Errors: []string{`service "Service" gRPC endpoint "Method": Request metadata attribute "object" must be a primitive or an array of primitives, got MetadataObject
service "Service" gRPC endpoint "Method": Request metadata attribute "mapping" must be a primitive or an array of primitives, got map
service "Service" gRPC endpoint "Method": Request metadata attribute "choice" must be a primitive or an array of primitives, got choice
service "Service" gRPC endpoint "Method": Response metadata attribute "object" must be a primitive or an array of primitives, got MetadataObject
service "Service" gRPC endpoint "Method": Response metadata attribute "mapping" must be a primitive or an array of primitives, got map
service "Service" gRPC endpoint "Method": Response metadata attribute "choice" must be a primitive or an array of primitives, got choice`},
		},
		"endpoint-with-inherit-error": {
			DSL:    testdata.GRPCEndpointWithInheritErrorDSL,
			Errors: []string{},
		},
		"endpoint-stream-compat": {
			DSL:    testdata.GRPCEndpointStreamCompat,
			Errors: []string{},
		},
		"endpoint-stream-compat-service-level": {
			DSL:    testdata.GRPCEndpointStreamCompatServiceLevel,
			Errors: []string{},
		},
		"endpoint-stream-compat-bad-value": {
			DSL:    testdata.GRPCEndpointStreamCompatBadValue,
			Errors: []string{`service "Service" gRPC endpoint "Method": invalid "grpc:stream:compat" meta value "v2": only "v1" is supported`},
		},
		"endpoint-stream-compat-no-streaming-payload": {
			DSL:    testdata.GRPCEndpointStreamCompatNoStreamingPayload,
			Errors: []string{`service "Service" gRPC endpoint "Method": "grpc:stream:compat" meta requires the method to define both Payload and StreamingPayload`},
		},
		"endpoint-stream-compat-union-payload": {
			DSL:    testdata.GRPCEndpointStreamCompatUnionPayload,
			Errors: []string{`service "Service" gRPC endpoint "Method": attribute "version_ref" of the method payload must be a primitive or an array of primitives to satisfy the "grpc:stream:compat" meta`},
		},
		"endpoint-union-containing-any": {
			DSL: testdata.GRPCEndpointWithUnionContainingAny,
			Errors: []string{
				`service "Service" method "MethodUnion": union type choice has array elements, not supported by gRPC
service "Service" method "MethodUnion": union type choice has map elements, not supported by gRPC`,
				// Any type error removed as Any is now supported
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if len(c.Errors) == 0 {
				expr.RunDSL(t, c.DSL)
			} else {
				var errs []error

				err := expr.RunInvalidDSL(t, c.DSL)
				if err != nil {
					var merr eval.MultiError
					if errors.As(err, &merr) {
						for _, e := range merr {
							errs = append(errs, e.GoError)
						}
					} else {
						errs = append(errs, err)
					}
				}

				if len(c.Errors) != len(errs) {
					t.Errorf("%s: got %d, expected the number of error values to match %d", name, len(errs), len(c.Errors))
				} else {
					for i, err := range errs {
						got := stripValidationLocations(err.Error())
						if got != c.Errors[i] {
							t.Errorf("%s:\ngot \t%q,\nexpected\t%q at index %d", name, got, c.Errors[i], i)
						}
					}
				}
			}
		})
	}
}

func TestGRPCEndpointStreamingPayloadKeepsInitialRequest(t *testing.T) {
	root := expr.RunDSL(t, testdata.GRPCEndpointWithStreamingPayloadInitialRequest)
	grpcSvc := root.API.GRPC.Service("Service")
	require.NotNil(t, grpcSvc)
	require.Len(t, grpcSvc.GRPCEndpoints, 1)

	endpoint := grpcSvc.GRPCEndpoints[0]
	req := expr.AsObject(endpoint.Request.Type)
	require.NotNil(t, req)
	require.NotNil(t, req.Attribute("repository_id"))
	require.NotNil(t, req.Attribute("version_ref"))
	require.True(t, endpoint.Metadata.IsEmpty())
}

func TestGRPCEndpointLegacyStreamCompat(t *testing.T) {
	cases := []struct {
		Name     string
		DSL      func()
		Expected bool
	}{
		{"method-level", testdata.GRPCEndpointStreamCompat, true},
		{"service-level", testdata.GRPCEndpointStreamCompatServiceLevel, true},
		{"not-set", testdata.GRPCEndpointWithStreamingPayloadInitialRequest, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			grpcSvc := root.API.GRPC.Service("Service")
			require.NotNil(t, grpcSvc)
			require.Len(t, grpcSvc.GRPCEndpoints, 1)
			require.Equal(t, c.Expected, grpcSvc.GRPCEndpoints[0].LegacyStreamCompat())
		})
	}
}
