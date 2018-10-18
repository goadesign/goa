package codegen

import (
	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

type (
	protobufAttribute struct{}
)

// newProtoBufAttributeHelper returns an AttributeHelper for protocol buffer
// types.
func newProtoBufAttributeHelper() codegen.AttributeHelper {
	return &protobufAttribute{}
}

// IsPointer returns true if the given attribute expression is a pointer type.
//
// In proto3 syntax, primitive fields are always non-pointers even when
// optional or has default values.
//
func (p *protobufAttribute) IsPointer(att *expr.AttributeExpr, required, pointer, useDefault bool) bool {
	return false
}
