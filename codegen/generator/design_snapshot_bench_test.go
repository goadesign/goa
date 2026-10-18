// These benchmarks measure repeated audits of an unchanged prepared design.
package generator

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func BenchmarkPreparedDesignAudit(b *testing.B) {
	for _, size := range []int{64, 512} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			object := make(expr.Object, 0, size)
			for index := range size {
				object = append(object, &expr.NamedAttributeExpr{
					Name: fmt.Sprintf("field%d", index),
					Attribute: &expr.AttributeExpr{
						Type:       expr.String,
						Meta:       expr.MetaExpr{"description": {"unchanged"}},
						Validation: &expr.ValidationExpr{Pattern: "[a-z]+"},
					},
				})
			}
			roots := []eval.Root{&expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
				TypeName: "Record", AttributeExpr: &expr.AttributeExpr{Type: &object},
			}}}}
			snapshot, err := snapshotPreparedDesign(roots)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				changed, err := snapshot.changedPath(roots)
				if err != nil || changed != "" {
					b.Fatalf("unchanged design: path %q, error %v", changed, err)
				}
			}
		})
	}
}
