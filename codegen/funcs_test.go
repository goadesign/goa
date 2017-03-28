package codegen

import (
	"testing"
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
