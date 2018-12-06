package expr

import (
	"regexp"
	"testing"
	"unicode/utf8"
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
	r := NewRandom("test")
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			val := &ValidationExpr{Pattern: k.Pattern}
			att := AttributeExpr{Validation: val}

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
