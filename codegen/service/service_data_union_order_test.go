// This file verifies retained service union declarations keep deterministic
// names regardless of design traversal order.
package service

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

func TestServicePlanUnionNamesAreIndependentOfObjectOrder(t *testing.T) {
	forward := retainedUnionNames(t, false)
	reverse := retainedUnionNames(t, true)

	require.Len(t, forward, 2)
	require.Equal(t, forward, reverse)
}

// retainedUnionNames plans two same-base unions in the requested field order
// and indexes their frozen names by their ordered branch contract.
func retainedUnionNames(t *testing.T, reverse bool) map[string]string {
	t.Helper()
	root := codegen.RunDSL(t, func() {
		signals := dsl.Type("Signals", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("source", func() {
				dsl.TypeName("SignalSource")
				dsl.Attribute("physical_point", dsl.String)
				dsl.Attribute("synthetic_series", dsl.String)
			})
		})
		inputs := dsl.Type("Inputs", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("source", func() {
				dsl.TypeName("InputSource")
				dsl.Attribute("time_series", dsl.String)
				dsl.Attribute("energy_rates", dsl.String)
			})
		})
		dsl.Service("test", func() {
			dsl.Method("read", func() {
				dsl.Payload(func() {
					if reverse {
						dsl.Attribute("beta", inputs)
						dsl.Attribute("alpha", signals)
					} else {
						dsl.Attribute("alpha", signals)
						dsl.Attribute("beta", inputs)
					}
				})
			})
		})
	})
	plan := mustServicePlan(t, root)
	names := make(map[string]string)
	for _, union := range plan.Services().Get("test").unions {
		branches := make([]string, len(union.Fields))
		for index, field := range union.Fields {
			branches[index] = field.Name
		}
		sort.Strings(branches)
		names[strings.Join(branches, ",")] = union.Name
	}
	return names
}
