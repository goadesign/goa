package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMuxRegexp(t *testing.T) {
	cases := []struct{ Name, Pattern, Expected string }{
		{"empty", "", ""},
		{"no capture", "a", "a"},
		{"no capture 2", "/a", "/a"},
		{"no capture 3", "/a/b", "/a/b"},
		{"no capture 4", ":a", ":a"},
		{"no capture 5", ":*a", ":*a"},
		{"segment", "/{a}", "/{a}"},
		{"segment 2", "/a/{b}", "/a/{b}"},
		{"segment 3", "/{a}/b", "/{a}/b"},
		{"segment 4", "/a/{b}/c", "/a/{b}/c"},
		{"wildcard", "/{*a}", "/{a:.*}"},
		{"wildcard 2", "/a/{*b}", "/a/{b:.*}"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			actual := wildPath.ReplaceAllString(c.Pattern, "/{$1:.*}")
			assert.Equal(t, c.Expected, actual)
		})
	}
}

func TestMiddlewares(t *testing.T) {
	m1 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("m1")) // nolint: errcheck
			h.ServeHTTP(w, r)
		})
	}
	m2 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("m2")) // nolint: errcheck
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
		t.Run(c.Name, func(t *testing.T) {
			m := NewMuxer()
			for _, mw := range c.Middlewares {
				m.Use(mw)
			}
			m.Handle("GET", "/", func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte("hello")) // nolint: errcheck
			})
			r, _ := http.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			m.ServeHTTP(w, r)
			assert.Equal(t, fmt.Sprintf("%shello", c.BodyPrefix), w.Body.String())
		})
	}
}

func TestVars(t *testing.T) {
	cases := []struct {
		Name     string
		Pattern  string
		URL      string
		Expected map[string]string
	}{
		{
			Name:    "simple",
			Pattern: "/users/{id}",
			URL:     "/users/123",
			Expected: map[string]string{
				"id": "123",
			},
		},
		{
			Name:    "multiple",
			Pattern: "/users/{id}/posts/{post_id}",
			URL:     "/users/123/posts/456",
			Expected: map[string]string{
				"id":      "123",
				"post_id": "456",
			},
		},
		{
			Name:    "wildcard",
			Pattern: "/users/{id}/posts/{*post_id}",
			URL:     "/users/123/posts/456/789",
			Expected: map[string]string{
				"id":      "123",
				"post_id": "456/789",
			},
		},
		{
			Name:     "no var",
			Pattern:  "/users",
			URL:      "/users",
			Expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var called bool
			mux := NewMuxer()
			mux.Handle("GET", c.Pattern, func(_ http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				assert.Equal(t, c.Expected, vars)
				called = true
			})
			req, _ := http.NewRequest("GET", c.URL, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.True(t, called)
		})
	}
}
