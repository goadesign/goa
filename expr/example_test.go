// This file exercises attribute example generation across validation rules and
// confirms every configured generator is anchored to its owning expression.
package expr_test

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestByPattern(t *testing.T) {
	cases := []struct {
		Name           string
		Pattern        string
		ExpectedMaxLen int
	}{
		{"not-a-regexp", "foo", 3},
		{"max-len", "foo[a-z]+", 9},
		{"max-len-2", "^/api/example/[0-9]+$", 19},
	}
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			val := &expr.ValidationExpr{Pattern: k.Pattern}
			att := expr.AttributeExpr{Validation: val}
			r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
				expr.MethodPayloadExampleIdentity(exampleMethod("pattern", k.Name)),
			)

			example := att.Example(r).(string)

			if match, _ := regexp.MatchString(k.Pattern, example); !match {
				t.Errorf("got %s, expected a match for %s", example, k.Pattern)
			}
			if utf8.RuneCountInString(example) > k.ExpectedMaxLen {
				t.Errorf("got %s (len %d) exceeded expected len of %d", example, len(example), k.ExpectedMaxLen)
			}
		})
	}
}

func TestByFormatUUID(t *testing.T) {
	val := &expr.ValidationExpr{Format: expr.FormatUUID}
	att := expr.AttributeExpr{Validation: val}
	r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(exampleMethod("format", "uuid")),
	)
	example := att.Example(r).(string)
	if !regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).MatchString(example) {
		t.Errorf("got %s, expected a match with `[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}`", example)
	}
}

func TestExample(t *testing.T) {
	cases := []struct {
		Name     string
		DSL      func()
		Expected any
		Error    string
	}{
		{"with-example", testdata.WithExampleDSL, "example", ""},
		{"with-array-example", testdata.WithArrayExampleDSL, []int{1, 2}, ""},
		{"with-map-example", testdata.WithMapExampleDSL, map[string]int{"name": 1, "value": 2}, ""},
		{"with-multiple-examples", testdata.WithMultipleExamplesDSL, 100, ""},
		{"overriding-example", testdata.OverridingExampleDSL, map[string]any{"name": "overridden"}, ""},
		{"with-extend", testdata.WithExtendExampleDSL, map[string]any{"name": "example"}, ""},
		{"invalid-example-type", testdata.InvalidExampleTypeDSL, nil, "service \"InvalidExampleType\" method \"Method\": payload - example value map[int]int{1:1} is incompatible with type map"},
		{"empty-example", testdata.EmptyExampleDSL, nil, "too few arguments given to Example in attribute"},
		{"hiding-example", testdata.HidingExampleDSL, nil, ""},
		{"openapi-generate-false-array-example", testdata.OpenAPIGenerateFalseArrayExampleDSL, map[string]any{"items": []map[string]any{{"name": "example"}}}, ""},
		{"overriding-hidden-examples", testdata.OverridingHiddenExamplesDSL, "example", ""},
	}
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			if k.Error == "" {
				expr.RunDSL(t, k.DSL)
				method := expr.Root.Services[0].Methods[0]
				r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
					expr.MethodPayloadExampleIdentity(method),
				)
				example := method.Payload.Example(r)
				if !reflect.DeepEqual(example, k.Expected) {
					t.Errorf("invalid example: got %v, expected %v", example, k.Expected)
				}
			} else {
				if err := expr.RunInvalidDSL(t, k.DSL); err == nil {
					t.Error("the expected error was not returned")
				} else if !strings.Contains(err.Error(), k.Error) {
					t.Errorf("invalid error: got %q, expected %q", err.Error(), k.Error)
				}
			}
		})
	}
}

// TestByLengthWithAliasType tests that alias types with length validations
// can generate examples correctly. Previously, this would panic because the
// code checked a.Type.Kind() instead of the underlying type's kind.
func TestByLengthWithAliasType(t *testing.T) {
	// Create an alias type based on String with length validation
	// We need to use the DSL package properly
	root := expr.RunDSL(t, testdata.AliasLengthValidationDSL)

	aliasType := root.UserType("ValidatedString")
	att := aliasType.Attribute()
	r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.UserTypeExampleIdentity(aliasType),
	)

	// This should not panic and should generate a string example
	// The key test is that byLength handles alias types correctly by unaliasing
	// before checking the kind. Previously this would panic with:
	// "invalid type for length validation: ValidatedString"
	example := att.Example(r)
	str, ok := example.(string)
	if !ok {
		t.Fatalf("Expected string example, got %T", example)
	}
	// Verify it's a valid non-empty string (exact length may vary due to randomness)
	if len(str) == 0 {
		t.Error("Expected non-empty string example")
	}
}

// TestByLengthWithAliasArray tests that alias array types with length
// validations generate examples correctly.
func TestByLengthWithAliasArray(t *testing.T) {
	root := expr.RunDSL(t, testdata.AliasArrayLengthValidationDSL)

	aliasType := root.UserType("StringArray")
	att := aliasType.Attribute()
	r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.UserTypeExampleIdentity(aliasType),
	)

	// This should not panic and should generate an array example
	example := att.Example(r)
	arr, ok := example.([]string)
	if !ok {
		t.Fatalf("Expected []string example, got %T", example)
	}

	// Verify the length is within the validation constraints
	if len(arr) < 2 || len(arr) > 5 {
		t.Errorf("Generated example has length %d, expected between 2 and 5", len(arr))
	}
}
