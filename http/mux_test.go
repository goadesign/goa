package http

import "testing"

func TestMuxRegexp(t *testing.T) {
	cases := []struct{ Name, Pattern, Expected string }{
		{"empty", "", ""},
		{"no capture", "a", "a"},
		{"no capture 2", "/a", "/a"},
		{"no capture 3", "/a/b", "/a/b"},
		{"no capture 4", "{a}", "{a}"},
		{"no capture 5", "{*a}", "{*a}"},
		{"segment", "/{a}", "/:a"},
		{"segment 2", "/a/{b}", "/a/:b"},
		{"segment 3", "/{a}/b", "/:a/b"},
		{"segment 4", "/a/{b}/c", "/a/:b/c"},
		{"path", "/{*a}", "/*a"},
		{"path 2", "/a/{*b}", "/a/*b"},
	}
	for _, c := range cases {
		actual := treemuxify(c.Pattern)
		if actual != c.Expected {
			t.Errorf("%s: expected %#v, got %#v", c.Name, c.Expected, actual)
		}
	}
}
