package middleware

type (
	// private type used to define context keys
	ctxKey int
)

const (
	// RequestMethodKey is the request context key used to store r.Method created by
	// the PopulateRequestContext middleware.
	RequestMethodKey ctxKey = iota + 1

	// RequestURIKey is the request context key used to store r.RequestURI created by
	// the PopulateRequestContext middleware.
	RequestURIKey

	// RequestPathKey is the request context key used to store r.URL.Path created by
	// the PopulateRequestContext middleware.
	RequestPathKey

	// RequestProtoKey is the request context key used to store r.Proto created by
	// the PopulateRequestContext middleware.
	RequestProtoKey

	// RequestHostKey is the request context key used to store r.Host created by
	// the PopulateRequestContext middleware.
	RequestHostKey

	// RequestRemoteAddrKey is the request context key used to store r.RemoteAddr
	// created by the PopulateRequestContext middleware.
	RequestRemoteAddrKey

	// RequestXForwardedForKey is the request context key used to store the
	// X-Forwarded-For header created by the PopulateRequestContext middleware.
	RequestXForwardedForKey

	// RequestXForwardedProtoKey is the request context key used to store the
	// X-Forwarded-Proto header created by the PopulateRequestContext middleware.
	RequestXForwardedProtoKey

	// RequestXRealIPKey is the request context key used to store the
	// X-Real-IP header created by the PopulateRequestContext middleware.
	RequestXRealIPKey

	// RequestAuthorizationKey is the request context key used to store the
	// Authorization header created by the PopulateRequestContext middleware.
	RequestAuthorizationKey

	// RequestRefererKey is the request context key used to store Referer header
	// created by the PopulateRequestContext middleware.
	RequestRefererKey

	// RequestUserAgentKey is the request context key used to store the User-Agent
	// header created by the PopulateRequestContext middleware.
	RequestUserAgentKey

	// RequestXRequestIDKey is the request context key used to store the X-Request-Id
	// header created by the PopulateRequestContext middleware.
	RequestXRequestIDKey

	// RequestAcceptKey is the request context key used to store the Accept header
	// created by the PopulateRequestContext middleware.
	RequestAcceptKey

	// RequestXCSRFTokenKey is the request context key used to store X-Csrf-Token header
	// created by the PopulateRequestContext middleware.
	RequestXCSRFTokenKey
)
