// These benchmarks separate saved snapshot allocation from repeated audits for
// synthetic deep, wide, shared, and cyclic expression graphs.
package generator

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func BenchmarkPreparedDesignShape(b *testing.B) {
	for _, shape := range []string{"deep", "wide", "shared", "cyclic"} {
		for _, size := range []int{64, 256, 1024} {
			b.Run(fmt.Sprintf("%s/%d", shape, size), func(b *testing.B) {
				roots := benchmarkDesignRoots(shape, size)
				b.Run("snapshot", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						if _, err := snapshotPreparedDesign(roots); err != nil {
							b.Fatal(err)
						}
					}
				})
				snapshot, err := snapshotPreparedDesign(roots)
				if err != nil {
					b.Fatal(err)
				}
				b.Run("audit", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						changed, err := snapshot.changedPath(roots)
						if err != nil || changed != "" {
							b.Fatalf("unchanged design: path %q, error %v", changed, err)
						}
					}
				})
			})
		}
	}
}

// benchmarkDesignRoots constructs expression graphs whose size and sharing are
// explicit so benchmark results distinguish depth from repeated references.
func benchmarkDesignRoots(shape string, size int) []eval.Root {
	node := &expr.UserTypeExpr{TypeName: "Node", AttributeExpr: &expr.AttributeExpr{Type: expr.String}}
	root := node
	if shape == "deep" || shape == "cyclic" {
		for range size {
			node = &expr.UserTypeExpr{TypeName: "Node", AttributeExpr: &expr.AttributeExpr{
				Type: &expr.Object{{Name: "next", Attribute: &expr.AttributeExpr{Type: node}}},
			}}
		}
		if shape == "cyclic" {
			root.AttributeExpr.Type = node
		}
		root = node
	} else {
		object := make(expr.Object, size)
		for index := range size {
			attribute := node.AttributeExpr
			if shape == "wide" {
				attribute = &expr.AttributeExpr{Type: expr.String}
			}
			object[index] = &expr.NamedAttributeExpr{Name: fmt.Sprintf("field%d", index), Attribute: attribute}
		}
		root = &expr.UserTypeExpr{TypeName: "Record", AttributeExpr: &expr.AttributeExpr{Type: &object}}
	}
	return []eval.Root{&expr.RootExpr{Types: []expr.UserType{root}}}
}
