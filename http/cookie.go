package http

// CookieIntAttr returns a pointer to v unless v is zero, in which case it
// returns nil. The HTTP transport calls this from generated client decoders
// for pointer-bound CookieAttributes Max-Age bindings: net/http parses an
// absent Max-Age cookie attribute as the zero value, so surfacing the
// attribute as nil rather than &0 lets clients tell "no Max-Age" apart from
// the unrelated semantic of "Max-Age explicitly set to zero". The two cases
// are inherently indistinguishable through net/http; treating zero as absent
// matches the more common server intent.
func CookieIntAttr[T ~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64](v T) *T {
	if v == 0 {
		return nil
	}
	return &v
}

// CookieStringAttr returns a pointer to v unless v is the empty string, in
// which case it returns nil. The HTTP transport calls this from generated
// client decoders for pointer-bound Domain and Path CookieAttributes
// bindings: net/http parses an absent attribute as the empty string, and the
// helper folds that into nil.
func CookieStringAttr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// CookieBoolAttr returns a pointer to v unless v is false, in which case it
// returns nil. The HTTP transport calls this from generated client decoders
// for pointer-bound Secure and HttpOnly CookieAttributes bindings: those
// attributes are flag-only on the wire (Secure / HttpOnly is either present
// or absent — there is no "Secure=false"), so net/http reports false for an
// absent flag and the helper folds that into nil.
func CookieBoolAttr(v bool) *bool {
	if !v {
		return nil
	}
	return &v
}

// CookieSameSiteAttr returns a pointer to v unless v is empty or the
// canonical "default" sentinel that net/http produces for cookies without a
// SameSite attribute (or with the literal SameSite=Default token). The HTTP
// transport calls this from generated client decoders for pointer-bound
// SameSite bindings.
func CookieSameSiteAttr(v string) *string {
	if v == "" || v == "default" {
		return nil
	}
	return &v
}
