// This file selects form fields from the HTTP body and its retained Go layout.
// Client and server codecs read these static choices; generated programs do not
// inspect types, tags or design expressions when sending or receiving forms.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// formFieldData retains one wire field and its generated Go selector.
	formFieldData struct {
		*AttributeData
		// Item describes one decoded array value when this field is an array.
		Item *AttributeData
	}
)

// formBodyFields reads the retained body layout and original service fields.
// Mapped attributes select wire names and pointer rules; original occurrences
// retain the generated selectors and service aliases needed for conversion.
func formBodyFields(body, payload *expr.AttributeExpr, context *codegen.AttributeContext) []*formFieldData {
	mapped := expr.NewMappedAttributeExpr(body)
	original := expr.AsObject(body.Type)
	if origin, selected := body.Meta["origin:attribute"]; selected {
		payload = payload.Find(origin[0])
	}
	object := expr.AsObject(mapped.Type)
	fields := make([]*formFieldData, 0, len(*object))
	for index, field := range *object {
		attribute := (*original)[index].Attribute
		fieldType := payload.Find(field.Name).Type
		data := &formFieldData{AttributeData: &AttributeData{
			Name:      mapped.ElemName(field.Name),
			FieldName: context.Scope.Field(attribute, mapped.ElemName(field.Name), true),
			Type:      attribute.Type,
			FieldType: fieldType,
			TypeRef:   context.Scope.Ref(attribute, ""),
			Pointer:   context.IsFieldPointer(field.Name, mapped.AttributeExpr),
			VarName:   "formValue",
		}}
		if array := expr.AsArray(attribute.Type); array != nil {
			data.Item = &AttributeData{
				Name:      mapped.ElemName(field.Name),
				Type:      array.ElemType.Type,
				FieldType: expr.AsArray(fieldType).ElemType.Type,
				TypeRef:   context.Scope.Ref(array.ElemType, ""),
				Pointer:   context.IsArrayElementPointer(array),
				VarName:   "formItem",
			}
		}
		fields = append(fields, data)
	}
	return fields
}

// formCodecImportPaths reserves only the packages named by the selected form
// fields. Collection runs before import names freeze, including for strings,
// numeric values and byte values in services that also contain JSON endpoints.
func formCodecImportPaths(service *expr.HTTPServiceExpr, client bool) []string {
	var paths []string
	for _, endpoint := range service.HTTPEndpoints {
		if !endpoint.FormRequest {
			continue
		}
		if len(paths) == 0 {
			if client {
				paths = append(paths, "io", "strings")
			} else {
				paths = append(paths, "mime", "net/url", "unicode/utf8")
			}
		}
		for _, field := range *expr.AsObject(endpoint.Body.Type) {
			dataType := underlyingDataType(field.Attribute.Type)
			if array := expr.AsArray(dataType); array != nil {
				dataType = underlyingDataType(array.ElemType.Type)
			}
			if dataType == expr.Bytes {
				paths = append(paths, "encoding/base64")
			}
			if dataTypeHasTextConvertedValue(dataType) {
				paths = append(paths, "strconv")
			}
		}
	}
	return paths
}
