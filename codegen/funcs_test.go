package codegen

import (
	"testing"
)

func TestSnakeCase(t *testing.T) {
	cases := map[string]struct {
		str      string
		expected string
	}{
		"all lower":             {"aaa", "aaa"},
		"start upper":           {"Aaa", "aaa"},
		"start upper 2":         {"AAAaa", "aa_aaa"},
		"mid upper":             {"aAa", "a_aa"},
		"end upper":             {"aaA", "aa_a"},
		"sequential uppers":     {"aaAAaa", "aa_a_aaa"},
		"sequential uppers 2":   {"aaAAAa", "aa_aa_aa"},
		"end sequential uppers": {"aaAA", "aa_aa"},
		"multiple_uppers":       {"aaAaaAaa", "aa_aaa_aaa"},
		"dashes":                {"aa-Aaa-Aaa", "aa_aaa_aaa"},
		"dashes 2":              {"aa-AAaaAA-Aaa", "aa_a_aaa_aa_aaa"},
		"underscores":           {"aa_Aaa_Aaa", "aa_aaa_aaa"},
		"underscores 2":         {"aa_AAaaAA_Aaa", "aa_a_aaa_aa_aaa"},
		"numbers":               {"aa1", "aa1"},
		"blank spaces":          {"  aa  AA aa   ", "aa_aa_aa"},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			actual := SnakeCase(tc.str)
			if actual != tc.expected {
				t.Errorf("got %q, expected %q", actual, tc.expected)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	cases := map[string]struct {
		str        string
		firstUpper bool
		useAcronym bool
		expected   string
	}{
		"all lower":                     {"aaa", false, true, "aaa"},
		"all lower first upper":         {"aaa", true, true, "Aaa"},
		"start upper":                   {"Aaa", false, true, "aaa"},
		"mid upper":                     {"a_aa", false, true, "aAa"},
		"end upper":                     {"aa_a", false, true, "aaA"},
		"sequential uppers":             {"aa_aaaa", false, true, "aaAaaa"},
		"end sequential uppers":         {"aa_aa", false, true, "aaAa"},
		"multiple_uppers":               {"aa_aaa_aaa", false, true, "aaAaaAaa"},
		"underscores":                   {"aa_aaa_aaa", false, true, "aaAaaAaa"},
		"acronym":                       {"aa_id", false, true, "aaID"},
		"lower camel case":              {"aa_id", false, false, "aaId"},
		"upper camel case":              {"aaID", false, true, "aaID"},
		"lower camel case with acronym": {"aaId", false, true, "aaID"},

		"disable acronym":                    {"aa_id", false, false, "aaId"},
		"disable acronym first upper":        {"aaID", true, false, "AaId"},
		"disable acronym upper case acronym": {"aa_ID", false, false, "aaId"},
		"disable acronym upper camel case":   {"aaID", false, false, "aaId"},
		"disable acronym lower camel case":   {"aaId", false, false, "aaId"},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			actual := CamelCase(tc.str, tc.firstUpper, tc.useAcronym)
			if actual != tc.expected {
				t.Errorf("got %q, expected %q", actual, tc.expected)
			}
		})
	}
}

func TestKebabCase(t *testing.T) {
	cases := map[string]struct {
		str      string
		expected string
	}{
		"all lower":             {"aaa", "aaa"},
		"start upper":           {"Aaa", "aaa"},
		"start upper 2":         {"AAAaa", "aa-aaa"},
		"mid upper":             {"aAa", "a-aa"},
		"end upper":             {"aaA", "aa-a"},
		"sequential uppers":     {"aaAAaa", "aa-a-aaa"},
		"sequential uppers 2":   {"aaAAAa", "aa-aa-aa"},
		"end sequential uppers": {"aaAA", "aa-aa"},
		"multiple_uppers":       {"aaAaaAaa", "aa-aaa-aaa"},
		"underscores":           {"aa_Aaa_Aaa", "aa-aaa-aaa"},
		"underscores 2":         {"aa_AAaaAA_Aaa", "aa-a-aaa-aa-aaa"},
		"dashes":                {"aa-Aaa-Aaa", "aa-aaa-aaa"},
		"dashes 2":              {"aa-AAaaAA-Aaa", "aa-a-aaa-aa-aaa"},
		"numbers":               {"aa1", "aa1"},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			actual := KebabCase(tc.str)
			if actual != tc.expected {
				t.Errorf("got %q, expected %q", actual, tc.expected)
			}
		})
	}
}

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
