// These benchmarks measure repeated field lookups in one retained Go type.
package codegen

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/expr"
)

func BenchmarkGoTypeLayoutLookups(b *testing.B) {
	for _, size := range []int{64, 512} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			object := make(expr.Object, 0, size)
			for index := range size {
				object = append(object, &expr.NamedAttributeExpr{Name: fmt.Sprintf("field%d", index), Attribute: &expr.AttributeExpr{Type: expr.String}})
			}
			plan, err := PlanGoType(&expr.AttributeExpr{Type: &object}, GoTypePlanOptions{Owner: "example.test/service"})
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				resolver := &goTypeLayoutAttributor{
					walk:   &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
					layout: plan.Link("example.test/service", nil),
				}
				for _, field := range object {
					for range 4 {
						if len(resolver.findAll(field.Attribute)) != 1 {
							b.Fatal("field occurrence missing")
						}
					}
				}
			}
		})
	}
}
