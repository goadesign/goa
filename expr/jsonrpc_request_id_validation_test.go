// This file checks that a JSON-RPC request ID can accept values created by
// generated clients when callers leave an optional ID unset.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestJSONRPCGeneratedRequestIDValidation checks that an optional request ID
// accepts every UUID that a generated client may create for it.
func TestJSONRPCGeneratedRequestIDValidation(t *testing.T) {
	tests := []struct {
		name       string
		validation func()
		want       string
	}{
		{
			name: "enum",
			validation: func() {
				dsl.Enum("known-id")
			},
			want: `enum`,
		},
		{
			name: "non-UUID format",
			validation: func() {
				dsl.Format(dsl.FormatEmail)
			},
			want: `format "email"`,
		},
		{
			name: "pattern",
			validation: func() {
				dsl.Pattern(`^req-`)
			},
			want: `pattern`,
		},
		{
			name: "minimum length",
			validation: func() {
				dsl.MinLength(37)
			},
			want: `minimum length 37`,
		},
		{
			name: "maximum length",
			validation: func() {
				dsl.MaxLength(35)
			},
			want: `maximum length 35`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, jsonRPCRequestIDValidationDSL(dsl.String, test.validation, false, nil))
			require.Contains(t, err.Error(), `JSON-RPC request ID field "request_id" is optional and has no default, so generated clients create a UUID when it is absent`)
			require.Contains(t, err.Error(), `the following validation rules may reject that UUID: `+test.want)
			require.Contains(t, err.Error(), `make the field required or give it a default`)
		})
	}
}

// TestJSONRPCGeneratedRequestIDInheritedValidation checks that named string
// types cannot hide a rule that rejects a client-created UUID.
func TestJSONRPCGeneratedRequestIDInheritedValidation(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		requestID := dsl.Type("RequestID", dsl.String, func() {
			dsl.Pattern(`^req-`)
		})
		jsonRPCRequestIDValidationService(requestID, nil, false, nil)
	})
	require.Contains(t, err.Error(), `the following validation rules may reject that UUID: pattern`)
}

// TestJSONRPCGeneratedRequestIDCompatibleValidation checks the rules that every
// generated UUID satisfies and the cases where Goa never creates the ID.
func TestJSONRPCGeneratedRequestIDCompatibleValidation(t *testing.T) {
	tests := []struct {
		name       string
		typ        func() expr.DataType
		validation func()
		required   bool
		defaultVal any
	}{
		{
			name: "optional UUID",
			typ: func() expr.DataType {
				return dsl.String
			},
			validation: func() {
				dsl.Format(dsl.FormatUUID)
				dsl.MinLength(36)
				dsl.MaxLength(36)
			},
		},
		{
			name: "inherited UUID",
			typ: func() expr.DataType {
				return dsl.Type("RequestID", dsl.String, func() {
					dsl.Format(dsl.FormatUUID)
					dsl.MinLength(1)
					dsl.MaxLength(64)
				})
			},
		},
		{
			name: "required pattern",
			typ: func() expr.DataType {
				return dsl.String
			},
			validation: func() {
				dsl.Pattern(`^req-`)
			},
			required: true,
		},
		{
			name: "defaulted enum",
			typ: func() expr.DataType {
				return dsl.String
			},
			validation: func() {
				dsl.Enum("known-id")
			},
			defaultVal: "known-id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr.RunDSL(t, jsonRPCRequestIDValidationDSL(test.typ(), test.validation, test.required, test.defaultVal))
		})
	}
}

// jsonRPCRequestIDValidationDSL defines one request ID with the given field
// contract.
func jsonRPCRequestIDValidationDSL(typ expr.DataType, validation func(), required bool, defaultVal any) func() {
	return func() {
		jsonRPCRequestIDValidationService(typ, validation, required, defaultVal)
	}
}

// jsonRPCRequestIDValidationService exposes one method through JSON-RPC so its
// request ID rules are evaluated.
func jsonRPCRequestIDValidationService(typ expr.DataType, validation func(), required bool, defaultVal any) {
	dsl.Service("records", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("lookup", func() {
			dsl.Payload(func() {
				dsl.ID("request_id", typ, func() {
					if validation != nil {
						validation()
					}
					if defaultVal != nil {
						dsl.Default(defaultVal)
					}
				})
				if required {
					dsl.Required("request_id")
				}
			})
			dsl.JSONRPC(func() {})
		})
	})
}
