// These benchmarks measure complete conversions of objects whose fields carry
// metadata, including the copied field mappings used during rendering.
package codegen

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/expr"
)

func BenchmarkTransformMappedObject(b *testing.B) {
	for _, count := range []int{16, 128} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			object := &expr.Object{}
			for index := range count {
				object.Set(fmt.Sprintf("field%d:wire%d", index, index), &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"description": {"field metadata"}},
				})
			}
			source := &expr.AttributeExpr{Type: object}
			target := expr.DupAtt(source)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				scope := NewNameScope()
				_, _, err := GoTransform(source, target, "source", "target",
					NewAttributeContext(true, false, false, "", scope),
					NewAttributeContext(false, false, true, "", scope), "", true)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
