package version

import (
	"regexp"
	"strconv"
	"testing"
)

func TestVersionFormat(t *testing.T) {
	semver := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-.+)?`)

	cases := map[string]struct{ Build, Regexp string }{
		"default": {"", `\.9999-dev$`},
		"custom":  {"1-custom", `\.1-custom$`},
	}

	for n, tc := range cases {
		oldBuild := Build
		Build = tc.Build
		ver := String()
		if !semver.MatchString(ver) {
			t.Errorf("%s: invalid version format %#v using build component %#v, not a valid semver", n, ver, tc.Build)
		}
		if !regexp.MustCompile(tc.Regexp).MatchString(ver) {
			t.Errorf("%s: invalid version format %#v using build component %#v, does not match %#v", n, ver, tc.Build, tc.Regexp)
		}
		Build = oldBuild
	}
}

func TestCompatibilty(t *testing.T) {
	cases := map[string]struct {
		Other  string
		Compat bool
		Err    bool
	}{
		"compatible":   {"v" + strconv.Itoa(Major) + ".12.13", true, false},
		"incompatible": {"v99999123129999.1.9", false, false},
		"error":        {"v99999123129999.1", false, true},
	}

	for n, tc := range cases {
		compat, err := Compatible(tc.Other)
		if compat != tc.Compat {
			t.Errorf("%s: expected Compatible to return %v for %#v", n, compat, tc.Other)
		}
		if (err != nil) != tc.Err {
			var not string
			if !tc.Err {
				not = "not "
			}
			t.Errorf("%s: expected Compatible to "+not+"return an error, but it returned %#v", n, err)
		}
	}
}
