// This file measures declaration collection with long shared ordering prefixes
// in several discovery orders. Every operation collects one fresh package.
package codegen

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

type (
	benchmarkNameOrder       string
	benchmarkSecondNameOrder string
	benchmarkThirdNameOrder  string
	benchmarkFourthNameOrder string
)

var benchmarkNameComparisons uint64

func (o benchmarkNameOrder) ComparePackageName(other PackageNameOrder) int {
	benchmarkNameComparisons++
	return strings.Compare(string(o), string(other.(benchmarkNameOrder)))
}

func (o benchmarkSecondNameOrder) ComparePackageName(other PackageNameOrder) int {
	benchmarkNameComparisons++
	return strings.Compare(string(o), string(other.(benchmarkSecondNameOrder)))
}

func (o benchmarkThirdNameOrder) ComparePackageName(other PackageNameOrder) int {
	benchmarkNameComparisons++
	return strings.Compare(string(o), string(other.(benchmarkThirdNameOrder)))
}

func (o benchmarkFourthNameOrder) ComparePackageName(other PackageNameOrder) int {
	benchmarkNameComparisons++
	return strings.Compare(string(o), string(other.(benchmarkFourthNameOrder)))
}

func BenchmarkDeclareNameOrderLookup(b *testing.B) {
	for _, size := range []int{32, 512, 8192} {
		for _, families := range []int{1, 4} {
			for _, discovery := range []string{"ascending", "reverse", "shuffled"} {
				b.Run(fmt.Sprintf("names=%d/families=%d/%s", size, families, discovery), func(b *testing.B) {
					orders := benchmarkPackageNameOrders(size, families, discovery)
					b.ReportAllocs()
					b.ResetTimer()
					benchmarkNameComparisons = 0
					for range b.N {
						pkg := newGeneratedPackage("generated.local/gen/names", "generated.local/gen/names", "names")
						for _, order := range orders {
							declaration := NewPreferredName(NameFunction, "Build", ExportedName, order)
							if err := pkg.DeclareName(declaration); err != nil {
								b.Fatal(err)
							}
						}
					}
					b.ReportMetric(float64(benchmarkNameComparisons)/float64(b.N), "comparisons/op")
				})
			}
		}
	}
}

// benchmarkPackageNameOrders creates immutable order values before timing, so
// measurements cover collecting declarations rather than constructing fixtures.
func benchmarkPackageNameOrders(size, families int, discovery string) []PackageNameOrder {
	orders := make([]PackageNameOrder, size)
	prefix := strings.Repeat("shared-ordering-prefix/", 16)
	for i := range orders {
		value := fmt.Sprintf("%s%08d", prefix, i)
		switch i % families {
		case 0:
			orders[i] = benchmarkNameOrder(value)
		case 1:
			orders[i] = benchmarkSecondNameOrder(value)
		case 2:
			orders[i] = benchmarkThirdNameOrder(value)
		case 3:
			orders[i] = benchmarkFourthNameOrder(value)
		}
	}
	switch discovery {
	case "reverse":
		slices.Reverse(orders)
	case "shuffled":
		random := rand.New(rand.NewPCG(1, 2))
		random.Shuffle(len(orders), func(i, j int) {
			orders[i], orders[j] = orders[j], orders[i]
		})
	}
	return orders
}
