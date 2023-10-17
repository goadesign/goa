package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	goahttp "goa.design/goa/v3/http"
)

func TestSmartRedirectSlashes(t *testing.T) {
	cases := []struct {
		Pattern  string
		URL      string
		Status   int
		Location string
	}{
		{"/users", "/users", http.StatusOK, ""},
		{"/users", "/users/", http.StatusMovedPermanently, "/users"},
		{"/users/", "/users/", http.StatusOK, ""},
		{"/users/", "/users", http.StatusMovedPermanently, "/users/"},
		{"/users/{id}", "/users/123", http.StatusOK, ""},
		{"/users/{id}", "/users/123/", http.StatusMovedPermanently, "/users/123"},
		{"/users/{id}/", "/users/123/", http.StatusOK, ""},
		{"/users/{id}/", "/users/123", http.StatusMovedPermanently, "/users/123/"},
		{"/users/{id}/posts/{post_id}", "/users/123/posts/456", http.StatusOK, ""},
		{"/users/{id}/posts/{post_id}", "/users/123/posts/456/", http.StatusMovedPermanently, "/users/123/posts/456"},
		{"/users/{id}/posts/{post_id}/", "/users/123/posts/456/", http.StatusOK, ""},
		{"/users/{id}/posts/{post_id}/", "/users/123/posts/456", http.StatusMovedPermanently, "/users/123/posts/456/"},
		{"/users/{id}/posts/{*post_id}", "/users/123/posts/456/789", http.StatusOK, ""},
		{"/users/{id}/posts/{*post_id}", "/users/123/posts/456/789/", http.StatusOK, ""},
		{"/users", "/users?name=foo", http.StatusOK, ""},
		{"/users", "/users/?name=foo", http.StatusMovedPermanently, "/users?name=foo"},
		{"/users/", "/users/?name=foo", http.StatusOK, ""},
		{"/users/", "/users?name=foo", http.StatusMovedPermanently, "/users/?name=foo"},
	}

	for _, c := range cases {
		t.Run(c.Pattern, func(t *testing.T) {
			var called bool
			mux := goahttp.NewMuxer()
			mux.Use(SmartRedirectSlashes)
			handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				called = true
			})
			mux.Handle("GET", c.Pattern, handler)
			req, _ := http.NewRequest("GET", c.URL, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.Equal(t, c.Status, w.Code)
			assert.Equal(t, w.Code == http.StatusOK, called)
			if w.Code == http.StatusMovedPermanently {
				assert.Equal(t, c.Location, w.Header().Get("Location"))
			}
		})
	}
}
