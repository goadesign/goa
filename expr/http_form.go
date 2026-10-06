// This file validates URL-encoded request bodies before Goa plans codecs.
// A form is a flat collection of typed values; unsupported shapes fail design
// evaluation rather than leave client or server code to guess an encoding.
package expr

import "goa.design/goa/v3/eval"

// formCustomType follows named scalar and collection definitions so a custom
// Go field type cannot be hidden behind an alias and produce invalid codecs.
func formCustomType(attribute *AttributeExpr) bool {
	if _, custom := attribute.Meta["struct:field:type"]; custom {
		return true
	}
	if user, named := attribute.Type.(UserType); named {
		return formCustomType(user.Attribute())
	}
	return false
}

// validateFormBody checks the remaining HTTP body after transport mappings.
// JSON-RPC envelopes, alternate body codecs and nested form values are rejected;
// accepted fields use the same generated body types and validators as HTTP.
func (e *HTTPEndpointExpr) validateFormBody(body *AttributeExpr, errors *eval.ValidationErrors) {
	if e.IsJSONRPC() || e.MultipartRequest || e.SkipRequestBodyEncodeDecode {
		errors.Add(e, "FormRequest cannot be combined with JSON-RPC, MultipartRequest or SkipRequestBodyEncodeDecode.")
	}
	object := AsObject(body.Type)
	if object == nil || len(*object) == 0 {
		errors.Add(e, "FormRequest requires a nonempty object request body.")
		return
	}
	for _, field := range *object {
		attribute := field.Attribute
		if array := AsArray(attribute.Type); array != nil {
			attribute = array.ElemType
		}
		if primitive := defaultPrimitive(attribute.Type); primitive == 0 || primitive == Any {
			errors.Add(e, "FormRequest field %q must be a primitive or an array of primitives.", field.Name)
			continue
		}
		if formCustomType(field.Attribute) || formCustomType(attribute) {
			errors.Add(e, "FormRequest field %q cannot use a custom Go field type.", field.Name)
		}
	}
}
