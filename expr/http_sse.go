package expr

import (
	"fmt"
	"strings"

	"slices"

	"goa.design/goa/v3/eval"
)

type (
	// HTTPSSEExpr describes a Server-Sent Events configuration for a HTTP endpoint.
	// It defines how a streaming endpoint should use the Server-Sent Events protocol
	// instead of WebSockets.
	HTTPSSEExpr struct {
		// RequestIDField is the name of the attribute in the Payload type
		// that provides the Last-Event-ID request header value.
		// If empty, no Last-Event-ID request header is included in the request.
		RequestIDField string
		// DataField is the name of the attribute in the StreamingResult type
		// that provides the data field for a Server-Sent Event.
		// If empty, the entire StreamingResult is used as the data field.
		DataField string
		// IDField is the name of the attribute in the StreamingResult type
		// that provides the id field for a Server-Sent Event.
		// If empty, no id field is included in the event.
		IDField string
		// EventField is the name of the attribute in the StreamingResult type
		// that provides the event field (event type) for a Server-Sent Event.
		// If empty, no event field is included in the event.
		EventField string
		// RetryField is the name of the attribute in the StreamingResult type
		// that provides the retry field for a Server-Sent Event.
		// If empty, no retry field is included in the event.
		RetryField string
	}
)

// EvalName returns the generic expression name used in error messages.
func (e *HTTPSSEExpr) EvalName() string {
	return "Server-Sent Events"
}

// Validate validates the Server-Sent Events expression against a specific result type.
func (e *HTTPSSEExpr) Validate(method *MethodExpr) error {
	if method == nil || method.Result == nil {
		return nil
	}

	verr := new(eval.ValidationErrors)
	if err := validateSSEField(method.Payload, e.RequestIDField, "request ID", []DataType{String}); err != nil {
		verr.Add(method, "%s", err.Error())
	}
	if err := validateSSEField(method.Result, e.DataField, "event data", nil); err != nil {
		verr.Add(method, "%s", err.Error())
	}
	if err := validateSSEField(method.Result, e.IDField, "event id", []DataType{String}); err != nil {
		verr.Add(method, "%s", err.Error())
	}
	if err := validateSSEField(method.Result, e.EventField, "event type", []DataType{String}); err != nil {
		verr.Add(method, "%s", err.Error())
	}
	if err := validateSSEField(method.Result, e.RetryField, "event retry", []DataType{Int, Int32, Int64, UInt, UInt32, UInt64}); err != nil {
		verr.Add(method, "%s", err.Error())
	}

	if len(verr.Errors) == 0 {
		return nil
	}
	return verr
}

// validateSSEField validates that the given field exists in the result type and has the expected type.
func validateSSEField(rt *AttributeExpr, field, desc string, expectedTypes []DataType) error {
	if field == "" {
		return nil
	}

	if rt == nil {
		return fmt.Errorf("cannot use %q for SSE %s field: result type is nil", field, desc)
	}

	obj := AsObject(rt.Type)
	if obj == nil {
		return fmt.Errorf("cannot use %q for SSE %s field: result type is not an object", field, desc)
	}

	att := obj.Attribute(field)
	if att == nil {
		return fmt.Errorf("cannot use %q for SSE %s field: attribute not found in result type", field, desc)
	}

	if len(expectedTypes) > 0 && !slices.Contains(expectedTypes, att.Type) {
		typeNames := make([]string, len(expectedTypes))
		for i, t := range expectedTypes {
			typeNames[i] = t.Name()
		}
		return fmt.Errorf("cannot use %q for SSE %s field: attribute type must be one of %s", field, desc, strings.Join(typeNames, ", "))
	}

	return nil
}
