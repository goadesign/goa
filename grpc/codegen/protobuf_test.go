package codegen

import "testing"

func TestProtobufify(t *testing.T) {
	cases := []struct {
		Name       string
		String     string
		FirstUpper bool
		Acronym    bool
		Expected   string
	}{{
		"AllLower", "lower", false, false, "lower",
	}, {
		"AllLowerFirstUpper", "lower", true, false, "Lower",
	}, {
		"AllUpper", "UPPER", false, false, "uPPER",
	}, {
		"AllUpperFirstUpper", "UPPER", true, false, "UPPER",
	}, {
		"StartUpperThenLower", "Upper", false, false, "upper",
	}, {
		"StartUpperThenLowerFirstUpper", "Upper", true, false, "Upper",
	}, {
		"StartsWithUnderscore", "_foo", false, false, "foo",
	}, {
		"EndsWithUnderscore", "foo_", false, false, "foo",
	}, {
		"ContainsUnderscore", "foo_bar", false, false, "fooBar",
	}, {
		"StartsWithDigits", "123foo", false, false, "123Foo",
	}, {
		"EndsWithDigits", "foo123", false, false, "foo123",
	}, {
		"ContainsDigits", "foo123bar", false, false, "foo123Bar",
	}, {
		"ContainsIgnoredAcronym", "foo_jwt", false, false, "fooJwt",
	}, {
		"ContainsAcronym", "foo_jwt", false, true, "fooJWT",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoBufify(c.String, c.FirstUpper, c.Acronym)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}
