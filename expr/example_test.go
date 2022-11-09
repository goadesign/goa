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
		{"max-len", "foo.*", 9},
		{"max-len-2", "^/api/example/[0-9]+$", 19},
	}
	r := expr.NewRandom("test")
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			val := &expr.ValidationExpr{Pattern: k.Pattern}
			att := expr.AttributeExpr{Validation: val}

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
	r := expr.NewRandom("test")
	example := att.Example(r).(string)
	if !regexp.MustCompile(`[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}`).MatchString(example) {
		t.Errorf("got %s, expected a match with `[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}`", example)
	}
}

func TestExample(t *testing.T) {
	cases := []struct {
		Name     string
		DSL      func()
		Expected interface{}
		Error    string
	}{
		{"with-example", testdata.WithExampleDSL, "example", ""},
		{"with-array-example", testdata.WithArrayExampleDSL, []int{1, 2}, ""},
		{"with-map-example", testdata.WithMapExampleDSL, map[string]int{"name": 1, "value": 2}, ""},
		{"with-multiple-examples", testdata.WithMultipleExamplesDSL, 100, ""},
		{"overriding-example", testdata.OverridingExampleDSL, map[string]interface{}{"name": "overridden"}, ""},
		{"with-extend", testdata.WithExtendExampleDSL, map[string]interface{}{"name": "example"}, ""},
		{"invalid-example-type", testdata.InvalidExampleTypeDSL, nil, "example value map[int]int{1:1} is incompatible with attribute of type map in attribute"},
		{"empty-example", testdata.EmptyExampleDSL, nil, "not enough arguments in attribute"},
		{"hiding-example", testdata.HidingExampleDSL, nil, ""},
		{"overriding-hidden-examples", testdata.OverridingHiddenExamplesDSL, "example", ""},
	}
	r := expr.NewRandom("test")
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			if k.Error == "" {
				expr.RunDSL(t, k.DSL)
				example := expr.Root.Services[0].Methods[0].Payload.Example(r)
				if !reflect.DeepEqual(example, k.Expected) {
					t.Errorf("invalid example: got %v, expected %v", example, k.Expected)
				}
			} else {
				if err := expr.RunInvalidDSL(t, k.DSL); err == nil {
					t.Error("the expected error was not returned")
				} else {
					if !strings.Contains(err.Error(), k.Error) {
						t.Errorf("invalid error: got %q, expected %q", err.Error(), k.Error)
					}
				}
			}
		})
	}
}
