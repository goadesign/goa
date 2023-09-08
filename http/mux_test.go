package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		actual := wildPath.ReplaceAllString(c.Pattern, "/{$1:.*}")
		if actual != c.Expected {
			t.Errorf("%s: expected %#v, got %#v", c.Name, c.Expected, actual)
		}
	}
}

func TestMiddlewares(t *testing.T) {
	m1 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("m1"))
			h.ServeHTTP(w, r)
		})
	}
	m2 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("m2"))
			h.ServeHTTP(w, r)
		})
	}
	cases := []struct {
		Name        string
		Middlewares []func(http.Handler) http.Handler
		BodyPrefix  string
	}{
		{"empty", nil, ""},
		{"one", []func(http.Handler) http.Handler{m1}, "m1"},
		{"two", []func(http.Handler) http.Handler{m1, m2}, "m1m2"},
	}
	for _, c := range cases {
		m := NewMuxer()
		for _, mw := range c.Middlewares {
			m.Use(mw)
		}
		m.Handle("GET", "/", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("hello"))
		})
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, r)
		if w.Body.String() != fmt.Sprintf("%shello", c.BodyPrefix) {
			t.Errorf("%s: got %s, expected %s", c.Name, w.Body.String(), fmt.Sprintf("%shello", c.BodyPrefix))
		}
	}
}
