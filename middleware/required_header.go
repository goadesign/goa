package middleware

import (
	"net/http"
	"regexp"

	"github.com/goadesign/goa"

	"context"
)

// RequireHeader requires a request header to match a value pattern. If the
// header is missing or does not match then the failureStatus is the response
// (e.g. http.StatusUnauthorized). If pathPattern is nil then any path is
// included. If requiredHeaderValue is nil then any value is accepted so long as
// the header is non-empty.
func RequireHeader(
	service *goa.Service,
	pathPattern *regexp.Regexp,
	requiredHeaderName string,
	requiredHeaderValue *regexp.Regexp,
	failureStatus int) goa.Middleware {

	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			if pathPattern == nil || pathPattern.MatchString(req.URL.Path) {
				matched := false
				headerValue := req.Header.Get(requiredHeaderName)
				if len(headerValue) > 0 {
					if requiredHeaderValue == nil {
						matched = true
					} else {
						matched = requiredHeaderValue.MatchString(headerValue)
					}
				}
				if matched {
					err = h(ctx, rw, req)
				} else {
					err = service.Send(ctx, failureStatus, http.StatusText(failureStatus))
				}
			} else {
				err = h(ctx, rw, req)
			}
			return
		}
	}
}
