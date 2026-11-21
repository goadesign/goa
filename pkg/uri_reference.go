package goa

import (
	"fmt"
	"net/netip"
	"strings"
)

// validateURIReference returns an error if val is not an RFC 3986
// URI-reference. It splits val into components as described in RFC 3986
// Appendix B and checks each component against its ABNF rule, including
// pct-encoded triplets, IP-literal hosts, and the restriction that the first
// segment of a relative path reference may not contain a colon. The empty
// string is a valid (same-document) reference.
func validateURIReference(val string) error {
	rest := val
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		if err := validateURIComponent("fragment", rest[i+1:], ":@/?"); err != nil {
			return err
		}
		rest = rest[:i]
	}
	if i := strings.IndexByte(rest, '?'); i >= 0 {
		if err := validateURIComponent("query", rest[i+1:], ":@/?"); err != nil {
			return err
		}
		rest = rest[:i]
	}
	// A colon before the first slash can only end a scheme: path-noscheme
	// forbids colons in the first segment of a relative reference.
	if i := strings.IndexByte(rest, ':'); i >= 0 && !strings.Contains(rest[:i], "/") {
		if err := validateURIScheme(rest[:i]); err != nil {
			return err
		}
		rest = rest[i+1:]
	}
	if authority, ok := strings.CutPrefix(rest, "//"); ok {
		i := strings.IndexByte(authority, '/')
		if i < 0 {
			i = len(authority)
		}
		if err := validateURIAuthority(authority[:i]); err != nil {
			return err
		}
		rest = authority[i:]
	}
	return validateURIComponent("path", rest, ":@/")
}

// validateURIScheme checks scheme = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." ).
func validateURIScheme(scheme string) error {
	if scheme == "" || !isURIAlpha(scheme[0]) {
		return fmt.Errorf("invalid scheme %q: must start with a letter", scheme)
	}
	for i := 1; i < len(scheme); i++ {
		c := scheme[i]
		if !isURIAlpha(c) && !isURIDigit(c) && c != '+' && c != '-' && c != '.' {
			return fmt.Errorf("invalid character %q in scheme", c)
		}
	}
	return nil
}

// validateURIAuthority checks authority = [ userinfo "@" ] host [ ":" port ].
func validateURIAuthority(authority string) error {
	hostport := authority
	if userinfo, rest, ok := strings.Cut(authority, "@"); ok {
		if err := validateURIComponent("userinfo", userinfo, ":"); err != nil {
			return err
		}
		hostport = rest
	}
	var port string
	if literal, ok := strings.CutPrefix(hostport, "["); ok {
		host, rest, ok := strings.Cut(literal, "]")
		if !ok {
			return fmt.Errorf("missing ']' in host %q", hostport)
		}
		if err := validateURIIPLiteral(host); err != nil {
			return err
		}
		if rest != "" {
			p, ok := strings.CutPrefix(rest, ":")
			if !ok {
				return fmt.Errorf("unexpected %q after IP literal", rest)
			}
			port = p
		}
	} else {
		host, p, _ := strings.Cut(hostport, ":")
		if err := validateURIComponent("host", host, ""); err != nil {
			return err
		}
		port = p
	}
	for i := 0; i < len(port); i++ {
		if !isURIDigit(port[i]) {
			return fmt.Errorf("invalid port %q", port)
		}
	}
	return nil
}

// validateURIIPLiteral checks the content of IP-literal = "[" ( IPv6address /
// IPvFuture ) "]" without the brackets.
func validateURIIPLiteral(host string) error {
	if host != "" && (host[0] == 'v' || host[0] == 'V') {
		version, address, ok := strings.Cut(host[1:], ".")
		if !ok || version == "" || address == "" {
			return fmt.Errorf("invalid IPvFuture host %q", host)
		}
		for i := 0; i < len(version); i++ {
			if !isURIHex(version[i]) {
				return fmt.Errorf("invalid IPvFuture version %q", version)
			}
		}
		for i := 0; i < len(address); i++ {
			c := address[i]
			if !isURIUnreserved(c) && !isURISubDelim(c) && c != ':' {
				return fmt.Errorf("invalid character %q in IPvFuture host", c)
			}
		}
		return nil
	}
	// RFC 3986 has no zone identifiers, so reject what netip would accept.
	if strings.IndexByte(host, '%') >= 0 {
		return fmt.Errorf("invalid IPv6 host %q", host)
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("invalid IPv6 host %q: %w", host, err)
	}
	if !addr.Is6() {
		return fmt.Errorf("invalid IPv6 host %q", host)
	}
	return nil
}

// validateURIComponent checks that every byte of s is an unreserved character,
// a sub-delim, a byte listed in extra, or the start of a pct-encoded triplet.
func validateURIComponent(component, s, extra string) error {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' {
			if i+2 >= len(s) || !isURIHex(s[i+1]) || !isURIHex(s[i+2]) {
				return fmt.Errorf("invalid percent-encoding in %s %q", component, s)
			}
			i += 2
			continue
		}
		if !isURIUnreserved(c) && !isURISubDelim(c) && strings.IndexByte(extra, c) < 0 {
			return fmt.Errorf("invalid character %q in %s", c, component)
		}
	}
	return nil
}

// isURIUnreserved reports whether c is an RFC 3986 unreserved character.
func isURIUnreserved(c byte) bool {
	return isURIAlpha(c) || isURIDigit(c) || c == '-' || c == '.' || c == '_' || c == '~'
}

// isURISubDelim reports whether c is an RFC 3986 sub-delims character.
func isURISubDelim(c byte) bool {
	return strings.IndexByte("!$&'()*+,;=", c) >= 0
}

// isURIAlpha reports whether c is an ASCII letter.
func isURIAlpha(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// isURIDigit reports whether c is an ASCII digit.
func isURIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// isURIHex reports whether c is an ASCII hexadecimal digit.
func isURIHex(c byte) bool {
	return isURIDigit(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}
