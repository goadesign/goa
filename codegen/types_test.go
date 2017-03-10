package codegen

import (
	"testing"
)

func TestGoify(t *testing.T) {
	cases := map[string]struct {
		str        string
		firstUpper bool
		expected   string
	}{
		"with first upper false":                                 {"blue_id", false, "blueID"},
		"with first upper false normal identifier all lower":     {"blue", false, "blue"},
		"with first upper false and UUID":                        {"blue_uuid", false, "blueUUID"},
		"with first upper true":                                  {"blue_id", true, "BlueID"},
		"with first upper true and UUID":                         {"blue_uuid", true, "BlueUUID"},
		"with first upper true normal identifier all lower":      {"blue", true, "Blue"},
		"with first upper false normal identifier":               {"Blue", false, "blue"},
		"with first upper true normal identifier":                {"Blue", true, "Blue"},
		"with invalid identifier":                                {"Blue%50", true, "Blue50"},
		"with invalid identifier firstupper false":               {"Blue%50", false, "blue50"},
		"with only UUID and firstupper false":                    {"UUID", false, "uuid"},
		"with consecutives invalid identifiers firstupper false": {"[[fields___type]]", false, "fieldsType"},
		"with consecutives invalid identifiers":                  {"[[fields___type]]", true, "FieldsType"},
		"with all invalid identifiers":                           {"[[", false, ""},
	}

	for k, tc := range cases {
		actual := Goify(tc.str, tc.firstUpper)

		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
