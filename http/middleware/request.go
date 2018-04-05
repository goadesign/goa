package middleware

import (
	"context"
	"net/http"
)

// PopulateRequestContext returns a middleware which populates a number of standard HTTP reqeust
// values into the request context. Those values may be extracted using the
// corresponding ContextKey type in this package.
func PopulateRequestContext() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for k, v := range map[ctxKey]string{
				RequestMethodKey:          r.Method,
				RequestURIKey:             r.RequestURI,
				RequestPathKey:            r.URL.Path,
				RequestProtoKey:           r.Proto,
				RequestHostKey:            r.Host,
				RequestRemoteAddrKey:      r.RemoteAddr,
				RequestXForwardedForKey:   r.Header.Get("X-Forwarded-For"),
				RequestXRealIPKey:         r.Header.Get("X-Real-Ip"),
				RequestXForwardedProtoKey: r.Header.Get("X-Forwarded-Proto"),
				RequestAuthorizationKey:   r.Header.Get("Authorization"),
				RequestRefererKey:         r.Header.Get("Referer"),
				RequestUserAgentKey:       r.Header.Get("User-Agent"),
				RequestXRequestIDKey:      r.Header.Get("X-Request-Id"),
				RequestXCSRFTokenKey:      r.Header.Get("X-Csrf-Token"),
				RequestAcceptKey:          r.Header.Get("Accept"),
			} {
				ctx = context.WithValue(ctx, k, v)
			}
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
