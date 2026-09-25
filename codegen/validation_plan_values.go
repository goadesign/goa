// Validation planning copies mutable rule values so later design edits cannot
// change the checks written by a retained plan.
package codegen

import "goa.design/goa/v3/expr"

// copyValidationFloat copies one optional scalar rule value.
func copyValidationFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// copyValidationInt copies one optional length rule value.
func copyValidationInt(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// copyValidationValue detaches the mutable collection shapes accepted by Goa
// enum validations. Primitive values are immutable and remain shared.
func copyValidationValue(value any) any {
	switch actual := value.(type) {
	case expr.Val:
		copied := make(expr.Val, len(actual))
		for name, child := range actual {
			copied[name] = copyValidationValue(child)
		}
		return copied
	case expr.ArrayVal:
		copied := make(expr.ArrayVal, len(actual))
		for index, child := range actual {
			copied[index] = copyValidationValue(child)
		}
		return copied
	case expr.MapVal:
		copied := make(expr.MapVal, len(actual))
		for key, child := range actual {
			copied[copyValidationValue(key)] = copyValidationValue(child)
		}
		return copied
	case []any:
		copied := make([]any, len(actual))
		for index, child := range actual {
			copied[index] = copyValidationValue(child)
		}
		return copied
	case []byte:
		return append([]byte(nil), actual...)
	case map[string]any:
		copied := make(map[string]any, len(actual))
		for name, child := range actual {
			copied[name] = copyValidationValue(child)
		}
		return copied
	case map[any]any:
		copied := make(map[any]any, len(actual))
		for key, child := range actual {
			copied[copyValidationValue(key)] = copyValidationValue(child)
		}
		return copied
	default:
		return actual
	}
}
