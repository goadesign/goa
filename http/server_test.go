package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplace(t *testing.T) {
	var (
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Path", r.URL.Path)
			w.Header().Set("X-RawPath", r.URL.RawPath)
		})
	)
	cases := []struct {
		old     string
		nw      string
		reqPath string
		path    string // If empty we want a 404.
		rawPath string
	}{
		{"/foo/bar", "", "/foo/bar/qux", "/qux", ""},
		{"/foo/bar", "", "/foo/bar%2Fqux", "/qux", "%2Fqux"},
		{"/foo/bar", "", "/foo%2Fbar/qux", "", ""}, // Escaped prefix does not match.
		{"/foo/bar", "", "/bar", "", ""},           // No prefix match.
		{"/foo/bar", "/baz", "/foo/bar/qux", "/baz/qux", ""},
		{"/foo/bar", "/baz", "/foo/bar%2Fqux", "/baz/qux", "/baz%2Fqux"},
		{"/foo/bar", "/baz", "/foo%2Fbar/qux", "", ""}, // Escaped prefix does not match.
		{"/foo/bar", "/baz", "/bar", "", ""},           // No prefix match.
		{"", "/baz/baz/baz", "/foo/bar/qux", "/baz/baz/baz", ""},
		{"", "/baz/baz/baz", "/foo/bar%2Fqux", "/baz/baz/baz", "/baz/baz/baz"},
		{"", "/baz/baz/baz", "/foo%2Fbar/qux", "/baz/baz/baz", "/baz/baz/baz"},
		{"", "/baz/baz/baz", "/bar", "/baz/baz/baz", ""},
	}
	for _, tc := range cases {
		t.Run(tc.reqPath, func(t *testing.T) {
			ts := httptest.NewServer(Replace(tc.old, tc.nw, h))
			defer ts.Close()
			c := ts.Client()
			res, err := c.Get(ts.URL + tc.reqPath)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
			if tc.path == "" {
				if res.StatusCode != http.StatusNotFound {
					t.Errorf("got %q, want 404 Not Found", res.Status)
				}
				return
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("got %q, want 200 OK", res.Status)
			}
			if g, w := res.Header.Get("X-Path"), tc.path; g != w {
				t.Errorf("got Path %q, want %q", g, w)
			}
			if g, w := res.Header.Get("X-RawPath"), tc.rawPath; g != w {
				t.Errorf("got RawPath %q, want %q", g, w)
			}
		})
	}
}
