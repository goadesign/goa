package expr_test

import (
	"testing"

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
			DSL: testdata.GRPCEndpointWithAnyType,
			Errors: []string{`service "Service" gRPC endpoint "Method": Map key type is Any type which is not supported in gRPC
service "Service" gRPC endpoint "Method": Array element type is Any type which is not supported in gRPC
service "Service" gRPC endpoint "Method": Attribute "invalid_primitive" is Any type which is not supported in gRPC
service "Service" gRPC endpoint "Method": Array element type is Any type which is not supported in gRPC
service "Service" gRPC endpoint "Method": Error "invalid_error_type" type is Any type which is not supported in gRPC
service "Service" gRPC endpoint "Method": Map element type is Any type which is not supported in gRPC`,
			},
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
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.Errors == nil || len(c.Errors) == 0 {
				expr.RunDSL(t, c.DSL)
			} else {
				var errors []error

				err := expr.RunInvalidDSL(t, c.DSL)
				if err != nil {
					if merr, ok := err.(eval.MultiError); ok {
						for _, e := range merr {
							errors = append(errors, e.GoError)
						}
					} else {
						errors = append(errors, err)
					}
				}

				if len(c.Errors) != len(errors) {
					t.Errorf("%s: got %d, expected the number of error values to match %d", name, len(errors), len(c.Errors))
				} else {
					for i, err := range errors {
						if err.Error() != c.Errors[i] {
							t.Errorf("%s:\ngot \t%q,\nexpected\t%q at index %d", name, err.Error(), c.Errors[i], i)
						}
					}
				}
			}
		})
	}
}
