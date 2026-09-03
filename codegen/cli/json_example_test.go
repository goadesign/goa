// This file checks that CLI examples turn every supported Goa map key into a
// distinct JSON object key.
package cli

import (
	"fmt"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen/testutil"
)

// TestJSONExampleFormatsPrimitiveMapKeys catches map entries being merged when
// generated help turns primitive and primitive-alias keys into JSON text.
func TestJSONExampleFormatsPrimitiveMapKeys(t *testing.T) {
	type (
		exampleBool   bool
		exampleInt    int32
		exampleUint   uint64
		exampleFloat  float64
		exampleString string
	)
	tests := []struct {
		name  string
		value any
	}{
		{"boolean alias", map[exampleBool]string{false: "disabled", true: "enabled"}},
		{"signed alias", map[exampleInt]string{-2: "negative", 10: "positive"}},
		{"unsigned", map[uint32]string{7: "seven", 42: "forty-two"}},
		{"unsigned alias", map[exampleUint]string{9: "nine", 11: "eleven"}},
		{"floating-point alias", map[exampleFloat]string{1.25: "one", 2.5: "two"}},
		{"string", map[string]string{"first": "one", "second": "two"}},
		{"string alias", map[exampleString]string{"left": "one", "right": "two"}},
		{"empty object", map[string]any{}},
	}

	var actual strings.Builder
	for _, test := range tests {
		fmt.Fprintf(&actual, "%s:\n%s\n", test.name, jsonExample(test.value))
	}
	testutil.AssertString(t, "testdata/golden/json_example_primitive_map_keys.golden", actual.String())
}
