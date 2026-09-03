package expr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestHTTPSSEExprValidation(t *testing.T) {
	cases := map[string]struct {
		SSE          *expr.HTTPSSEExpr
		Payload      *expr.AttributeExpr
		Result       *expr.AttributeExpr
		StreamResult *expr.AttributeExpr
		ExpectedErrs []string
	}{
		"valid-empty": {
			SSE:     &expr.HTTPSSEExpr{},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result:  &expr.AttributeExpr{Type: &expr.Object{}},
		},
		"valid-with-fields": {
			SSE: &expr.HTTPSSEExpr{
				RequestIDField: "request_id",
				DataField:      "data",
				IDField:        "id",
				EventField:     "event",
				RetryField:     "retry",
			},
			Payload: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "request_id", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			},
			Result: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "data", Attribute: &expr.AttributeExpr{Type: expr.String}},
					&expr.NamedAttributeExpr{Name: "id", Attribute: &expr.AttributeExpr{Type: expr.String}},
					&expr.NamedAttributeExpr{Name: "event", Attribute: &expr.AttributeExpr{Type: expr.String}},
					&expr.NamedAttributeExpr{Name: "retry", Attribute: &expr.AttributeExpr{Type: expr.Int}},
				},
			},
		},
		"valid-named-fields": {
			SSE: &expr.HTTPSSEExpr{
				RequestIDField: "request_id",
				IDField:        "id",
				EventField:     "event",
				RetryField:     "retry",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "request_id", Attribute: &expr.AttributeExpr{Type: namedPrimitive("RequestID", expr.String)}},
			}},
			Result: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "id", Attribute: &expr.AttributeExpr{Type: namedPrimitive("EventID", expr.String)}},
				&expr.NamedAttributeExpr{Name: "event", Attribute: &expr.AttributeExpr{Type: namedPrimitive("EventType", expr.String)}},
				&expr.NamedAttributeExpr{Name: "retry", Attribute: &expr.AttributeExpr{Type: namedPrimitive("Retry", expr.UInt32)}},
			}},
		},
		"invalid-id-field-type": {
			SSE: &expr.HTTPSSEExpr{
				IDField: "id",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "id", Attribute: &expr.AttributeExpr{Type: expr.Int}},
				},
			},
			ExpectedErrs: []string{"cannot use \"id\" for SSE event id field: attribute type must be one of string"},
		},
		"invalid-request-id-field-type": {
			SSE: &expr.HTTPSSEExpr{
				RequestIDField: "request_id",
			},
			Payload: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "request_id", Attribute: &expr.AttributeExpr{Type: expr.Int}},
				},
			},
			Result:       &expr.AttributeExpr{Type: &expr.Object{}},
			ExpectedErrs: []string{"cannot use \"request_id\" for SSE request ID field: attribute type must be one of string"},
		},
		"invalid-event-field-type": {
			SSE: &expr.HTTPSSEExpr{
				EventField: "event",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "event", Attribute: &expr.AttributeExpr{Type: expr.Int}},
				},
			},
			ExpectedErrs: []string{"cannot use \"event\" for SSE event type field: attribute type must be one of string"},
		},
		"invalid-retry-field-type": {
			SSE: &expr.HTTPSSEExpr{
				RetryField: "retry",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "retry", Attribute: &expr.AttributeExpr{Type: expr.Boolean}},
				},
			},
			ExpectedErrs: []string{"cannot use \"retry\" for SSE event retry field: attribute type must be one of int, int32, int64, uint, uint32, uint64"},
		},
		"missing-field": {
			SSE: &expr.HTTPSSEExpr{
				DataField: "missing",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "data", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			},
			ExpectedErrs: []string{"cannot use \"missing\" for SSE event data field: attribute not found in result type"},
		},
		"valid-streaming-result-field": {
			SSE:     &expr.HTTPSSEExpr{DataField: "event_data"},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "final_data", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
			StreamResult: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "event_data", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
		},
		"invalid-ordinary-result-field": {
			SSE:     &expr.HTTPSSEExpr{DataField: "final_data"},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
			Result: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "final_data", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
			StreamResult: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "event_data", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
			ExpectedErrs: []string{"cannot use \"final_data\" for SSE event data field: attribute not found in result type"},
		},
		"missing-request-id-field": {
			SSE: &expr.HTTPSSEExpr{
				RequestIDField: "missing",
			},
			Payload: &expr.AttributeExpr{
				Type: &expr.Object{
					&expr.NamedAttributeExpr{Name: "request_id", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			},
			Result:       &expr.AttributeExpr{Type: &expr.Object{}},
			ExpectedErrs: []string{"cannot use \"missing\" for SSE request ID field: attribute not found in payload"},
		},
		"nil-result-type": {
			SSE: &expr.HTTPSSEExpr{
				DataField: "data",
			},
			Payload: &expr.AttributeExpr{Type: &expr.Object{}},
		},
		"nil-payload-type": {
			SSE: &expr.HTTPSSEExpr{
				RequestIDField: "request_id",
			},
		},
		"empty-result-type": {
			SSE: &expr.HTTPSSEExpr{
				DataField: "data",
			},
			Payload:      &expr.AttributeExpr{Type: &expr.Object{}},
			Result:       &expr.AttributeExpr{Type: expr.Int},
			ExpectedErrs: []string{"cannot use \"data\" for SSE event data field: result type is not an object"},
		},
		"non-object-result-type": {
			SSE: &expr.HTTPSSEExpr{
				DataField: "data",
			},
			Payload:      &expr.AttributeExpr{Type: &expr.Object{}},
			Result:       &expr.AttributeExpr{Type: expr.String},
			ExpectedErrs: []string{"cannot use \"data\" for SSE event data field: result type is not an object"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create a mock HTTP endpoint expression for validation
			methodExpr := &expr.MethodExpr{
				Name:            "TestMethod",
				Payload:         tc.Payload,
				Result:          tc.Result,
				StreamingResult: tc.StreamResult,
				Stream:          expr.ServerStreamKind, // Must be a streaming method for SSE
			}

			// Run validation
			err := tc.SSE.Validate(methodExpr)

			// Check results
			if len(tc.ExpectedErrs) == 0 {
				require.NoError(t, err, "expected no error")
			} else {
				require.Error(t, err, "expected error, got none")
				for _, expected := range tc.ExpectedErrs {
					assert.Contains(t, err.Error(), expected, "error should contain expected message")
				}
			}
		})
	}
}

func namedPrimitive(name string, primitive expr.Primitive) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: primitive},
		TypeName:      name,
		UID:           name,
	}
}
