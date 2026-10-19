// This file measures planning work for synthetic response graphs whose fields
// share named definitions. Collection must retain every generated type record.
package codegen

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/expr"
)

func BenchmarkWirePlanningSharedGraph(b *testing.B) {
	for _, depth := range []int{4, 8, 12} {
		body := sharedPlanningGraph(depth)
		b.Run(fmt.Sprintf("collect/depth%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				catalog := newWireTypeCatalog()
				catalog.collect(body, wireResponseBody, wireTypePolicy{})
				if len(catalog.records) != depth+1 {
					b.Fatalf("records: %d", len(catalog.records))
				}
			}
		})
		b.Run(fmt.Sprintf("copy/depth%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				expr.DupAtt(body)
			}
		})
	}
}

// sharedPlanningGraph creates two fields referencing the same next definition
// at each level, so the number of definitions grows only with depth.
func sharedPlanningGraph(depth int) *expr.AttributeExpr {
	var next expr.DataType = expr.String
	for level := depth; level >= 0; level-- {
		fields := expr.Object{
			&expr.NamedAttributeExpr{Name: "left", Attribute: &expr.AttributeExpr{Type: next}},
			&expr.NamedAttributeExpr{Name: "right", Attribute: &expr.AttributeExpr{Type: next}},
		}
		name := fmt.Sprintf("Level%d", level)
		next = &expr.UserTypeExpr{TypeName: name, UID: name, AttributeExpr: &expr.AttributeExpr{Type: &fields}}
	}
	return &expr.AttributeExpr{Type: next}
}
