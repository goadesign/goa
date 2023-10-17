package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SmartRedirectSlashes is a middleware that matches the request path with
// patterns added to the router and redirects it.
//
// If a pattern is added to the router with a trailing slash, any matches on
// that pattern without a trailing slash will be redirected to the version with
// the slash. If a pattern does not have a trailing slash, matches on that
// pattern with a trailing slash will be redirected to the version without.
//
// This middleware depends on chi, so it needs to be mounted on chi's router.
// It make the router behavior similar to httptreemux.
func SmartRedirectSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			var path string
			if rctx.RoutePath != "" {
				path = rctx.RoutePath
			} else {
				path = r.URL.Path
			}
			var method string
			if rctx.RouteMethod != "" {
				method = rctx.RouteMethod
			} else {
				method = r.Method
			}
			if len(path) > 1 {
				if rctx.Routes != nil {
					if !rctx.Routes.Match(chi.NewRouteContext(), method, path) {
						if path[len(path)-1] == '/' {
							path = path[:len(path)-1]
						} else {
							path += "/"
						}
						if rctx.Routes.Match(chi.NewRouteContext(), method, path) {
							if r.URL.RawQuery != "" {
								path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
							}
							redirectURL := fmt.Sprintf("//%s%s", r.Host, path)
							http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
							return
						}
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
