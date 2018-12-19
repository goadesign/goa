package codegen

import (
	"goa.design/goa/expr"
)

type (
	protobufAnalyzer struct {
		*expr.Analyzer
	}
)

// newProtoBufAnalyzer returns an attribute analyzer for protocol buffer types.
func newProtoBufAnalyzer(att *expr.AttributeExpr, p *expr.AttributeProperties) expr.AttributeAnalyzer {
	return &protobufAnalyzer{
		Analyzer: &expr.Analyzer{AttributeExpr: att, AttributeProperties: p},
	}
}

// IsPointer returns true if the given attribute expression is a pointer type.
//
// In proto3 syntax, primitive fields are always non-pointers even when
// optional or has default values.
//
func (p *protobufAnalyzer) IsPointer() bool {
	return false
}
