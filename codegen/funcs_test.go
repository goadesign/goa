package codegen

import (
	"testing"

	"goa.design/goa/expr"
)

func TestWrapText(t *testing.T) {
	cases := map[string]struct {
		maxChars int
		str      string
		expected string
	}{
		"long_text": {
			maxChars: 50,
			str:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			expected: "Lorem ipsum dolor sit amet, consectetur\n" +
				"adipiscing elit, sed do eiusmod tempor incididunt\n" +
				"ut labore et dolore magna aliqua. Ut enim ad\n" +
				"minim veniam, quis nostrud exercitation ullamco\n" +
				"laboris nisi ut aliquip ex ea commodo consequat.",
		},
		"long_text_with_newlines": {
			maxChars: 50,
			str: "Lorem ipsum dolor sit amet,\n" +
				"consectetur adipiscing elit,\n" +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n" +
				"Ut enim ad minim veniam,\n" +
				"quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			expected: "Lorem ipsum dolor sit amet,\n" +
				"consectetur adipiscing elit,\n" +
				"sed do eiusmod tempor incididunt ut labore et\n" +
				"dolore magna aliqua.\n" +
				"Ut enim ad minim veniam,\n" +
				"quis nostrud exercitation ullamco laboris nisi ut\n" +
				"aliquip ex ea commodo consequat.",
		},
		"respect_empty_lines": {
			maxChars: 50,
			str: "Lorem ipsum dolor sit amet,\n" +
				"\n" +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			expected: "Lorem ipsum dolor sit amet,\n" +
				"\n" +
				"sed do eiusmod tempor incididunt ut labore et\n" +
				"dolore magna aliqua.",
		},
		"too_long_word_to_wrap_intact": {
			maxChars: 15,
			str:      "Who said supercalifragilisticexpialidocious in a movie",
			expected: "Who said\n" +
				"supercalifragilisticexpialidocious\n" +
				"in a movie",
		},
		"empty": {
			maxChars: 50,
			str:      "",
			expected: "",
		},
	}

	for k, tc := range cases {
		actual := WrapText(tc.str, tc.maxChars)

		if actual != tc.expected {
			t.Errorf("%s: got `%s`, expected `%s`", k, actual, tc.expected)
		}
	}
}

func TestAPIPkg(t *testing.T) {
	cases := map[string]struct {
		root     *expr.RootExpr
		expected string
	}{
		"distinct-API-name": {
			root: &expr.RootExpr{
				API: &expr.APIExpr{Name: "API"},
				Services: []*expr.ServiceExpr{
					{Name: "Service"},
				},
			},
			expected: "api",
		},
		"conflicting-API-name": {
			root: &expr.RootExpr{
				API: &expr.APIExpr{Name: "Service"},
				Services: []*expr.ServiceExpr{
					{Name: "Service"},
				},
			},
			expected: "serviceapi",
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			got := APIPkg(tc.root)
			if got != tc.expected {
				t.Errorf("invalid API pkg name: got %q, expected %q", got, tc.expected)
			}
		})
	}
}
