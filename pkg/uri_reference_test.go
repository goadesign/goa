package goa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateURIReferenceGrammar(t *testing.T) {
	cases := []struct {
		name  string
		val   string
		valid bool
	}{
		{"pct-encoded reg-name", "http://exa%41mple.com", true},
		{"ipvfuture host", "http://[v1.x]/", true},
		{"ipv4 mapped ipv6 host", "http://[::ffff:1.2.3.4]/", true},
		{"userinfo", "http://user:pass@example.com:8080/a", true},
		{"empty port", "http://example.com:/a", true},
		{"question mark in fragment", "?a#b?c", true},
		{"question mark in query", "/a?b?c", true},
		{"colon after first segment", "a/b:c", true},
		{"empty authority", "file:///etc/hosts", true},
		{"scheme only", "mailto:", true},
		{"malformed query escape", "?q=%zz", false},
		{"truncated query escape", "/a?b=%", false},
		{"truncated fragment escape", "#%4", false},
		{"malformed host escape", "http://exa%4gmple.com", false},
		{"bracket in path", "/a[b]", false},
		{"open bracket in path", "/a[", false},
		{"bracket in query", "?a[0]=1", false},
		{"second fragment", "a#b#c", false},
		{"at sign in host", "http://a@b@c/", false},
		{"ipv4 literal", "http://[1.2.3.4]/", false},
		{"ipv6 zone", "http://[fe80::1%25eth0]/", false},
		{"ipv6 without brackets", "http://::1/", false},
		{"text after ip literal", "http://[::1]x/", false},
		{"non-digit port", "http://example.com:80a/", false},
		{"empty ipvfuture version", "http://[v.x]/", false},
		{"empty ipvfuture address", "http://[v1.]/", false},
		{"invalid scheme character", "a_b:c", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFormat("ref", tc.val, FormatURIReference)
			if tc.valid {
				require.NoError(t, err)
				return
			}
			var se *ServiceError
			require.ErrorAs(t, err, &se)
			require.Equal(t, InvalidFormat, se.Name)
		})
	}
}
