package goa

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp/syntax"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateFormat(t *testing.T) {
	var (
		validDate              = "2015-10-26"
		invalidDate            = "201510-26"
		validDateTime          = "2015-10-26T08:31:23Z"
		invalidDateTime        = "201510-26T08:31:23Z"
		validUUID              = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
		validUUIDWithBrace     = "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"
		validUUIDRaw           = "6ba7b8109dad11d180b400c04fd430c8"
		validUUIDWithURNPrefix = "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
		invalidUUID            = "96054a62-a9e45ed26688389b"
		invalidUUIDNonHex      = "abcdefgh-ijkl-mnop-qrst-uvqxyz012345" // UUID with characters other than hex digit
		validEmail             = "raphael@goa.design"

		// Re-enable once CircleCI uses Go 1.13
		// invalidEmail    = "foo"

		validHostname   = "goa.design"
		invalidHostname = "_hi_"
		validIPv4       = "192.168.0.1"
		invalidIPv4     = "192-168.0.1"
		validIPv6       = "::1"
		invalidIPv6     = "foo"
		validURI        = "hhp://goa.design/contact"
		invalidURI      = "foo_"
		validMAC        = "06-00-00-00-00-00"
		invalidMAC      = "bar"
		validCIDR       = "10.0.0.0/8"
		invalidCIDR     = "foo"
		validRegexp     = "^goa$"
		invalidRegexp   = "foo["
		validJSON       = `{"a":"b","c":2}`
		invalidJSON     = "{"
		validRFC1123    = "Mon, 04 Jun 2017 23:52:05 MST"
		invalidRFC1123  = "Mon 04 Jun 2017 23:52:05 MST"
	)
	cases := map[string]struct {
		name     string
		val      string
		format   Format
		expected error
	}{
		"valid date":                 {"validDate", validDate, FormatDate, nil},
		"invalid date":               {"invalidDate", invalidDate, FormatDate, InvalidFormatError("invalidDate", invalidDate, FormatDate, &time.ParseError{Layout: time.DateOnly, Value: invalidDate, LayoutElem: "-", ValueElem: invalidDate[4:]})},
		"valid date-time":            {"validDateTime", validDateTime, FormatDateTime, nil},
		"invalid date-time":          {"invalidDateTime", invalidDateTime, FormatDateTime, InvalidFormatError("invalidDateTime", invalidDateTime, FormatDateTime, &time.ParseError{Layout: time.RFC3339, Value: invalidDateTime, LayoutElem: "-", ValueElem: invalidDateTime[4:]})},
		"valid uuid":                 {"validUUID", validUUID, FormatUUID, nil},
		"valid uuid with brace":      {"validUUIDWithBrace", validUUIDWithBrace, FormatUUID, nil},
		"valid uuid with no dash":    {"validUUIDRaw", validUUIDRaw, FormatUUID, nil},
		"valid uuid with urn prefix": {"validUUIDWithURNPrefix", validUUIDWithURNPrefix, FormatUUID, nil},
		"invalid uuid":               {"invalidUUID", invalidUUID, FormatUUID, InvalidFormatError("invalidUUID", invalidUUID, FormatUUID, fmt.Errorf("uuid: %s: invalid UUID length: 25", invalidUUID))},
		"invalid uuid non hex":       {"invalidUUIDNonHex", invalidUUIDNonHex, FormatUUID, InvalidFormatError("invalidUUIDNonHex", invalidUUIDNonHex, FormatUUID, fmt.Errorf("uuid: %s: invalid UUID format", invalidUUIDNonHex))},

		"valid email": {"validEmail", validEmail, FormatEmail, nil},

		// Re-enable once CircleCI uses Go 1.13
		// "invalid email":      {"invalidEmail", invalidEmail, FormatEmail, InvalidFormatError("invalidEmail", invalidEmail, FormatEmail, errors.New("mail: missing '@' or angle-addr"))},

		"valid hostname":     {"validHostname", validHostname, FormatHostname, nil},
		"invalid hostname":   {"invalidHostname", invalidHostname, FormatHostname, InvalidFormatError("invalidHostname", invalidHostname, FormatHostname, fmt.Errorf("hostname contains invalid character '_'"))},
		"valid ipv4":         {"validIPv4", validIPv4, FormatIPv4, nil},
		"valid ipv6 as ipv4": {"validIPv6", validIPv6, FormatIPv4, InvalidFormatError("validIPv6", validIPv6, FormatIPv4, fmt.Errorf("\"%s\" is an invalid %s value", validIPv6, FormatIPv4))},
		"invalid ipv4":       {"invalidIPv4", invalidIPv4, FormatIPv4, InvalidFormatError("invalidIPv4", invalidIPv4, FormatIPv4, fmt.Errorf("\"%s\" is an invalid %s value", invalidIPv4, FormatIPv4))},
		"valid ipv6":         {"validIPv6", validIPv6, FormatIPv6, nil},
		"valid ipv4 as ipv6": {"validIPv4", validIPv4, FormatIPv6, InvalidFormatError("validIPv4", validIPv4, FormatIPv6, fmt.Errorf("\"%s\" is an invalid %s value", validIPv4, FormatIPv6))},
		"invalid ipv6":       {"invalidIPv6", invalidIPv6, FormatIPv6, InvalidFormatError("invalidIPv6", invalidIPv6, FormatIPv6, fmt.Errorf("\"%s\" is an invalid %s value", invalidIPv6, FormatIPv6))},
		"valid ipv4 as ip":   {"validIPv4", validIPv4, FormatIP, nil},
		"valid ipv6 as ip":   {"validIPv6", validIPv6, FormatIP, nil},
		"invalid ipv4 as ip": {"invalidIPv4", invalidIPv4, FormatIP, InvalidFormatError("invalidIPv4", invalidIPv4, FormatIP, fmt.Errorf("\"%s\" is an invalid %s value", invalidIPv4, FormatIP))},
		"invalid ipv6 as ip": {"invalidIPv6", invalidIPv6, FormatIP, InvalidFormatError("invalidIPv6", invalidIPv6, FormatIP, fmt.Errorf("\"%s\" is an invalid %s value", invalidIPv6, FormatIP))},
		"valid uri":          {"validURI", validURI, FormatURI, nil},
		"invalid uri":        {"invalidURI", invalidURI, FormatURI, InvalidFormatError("invalidURI", invalidURI, FormatURI, &url.Error{Op: "parse", URL: invalidURI, Err: errors.New("invalid URI for request")})},
		"valid mac":          {"validMAC", validMAC, FormatMAC, nil},
		"invalid mac":        {"invalidMAC", invalidMAC, FormatMAC, InvalidFormatError("invalidMAC", invalidMAC, FormatMAC, &net.AddrError{Err: "invalid MAC address", Addr: invalidMAC})},
		"valid cidr":         {"validCIDR", validCIDR, FormatCIDR, nil},
		"invalid cidr":       {"invalidCIDR", invalidCIDR, FormatCIDR, InvalidFormatError("invalidCIDR", invalidCIDR, FormatCIDR, &net.ParseError{Type: "CIDR address", Text: invalidCIDR})},
		"valid regexp":       {"validRegexp", validRegexp, FormatRegexp, nil},
		"invalid regexp":     {"invalidRegexp", invalidRegexp, FormatRegexp, InvalidFormatError("invalidRegexp", invalidRegexp, FormatRegexp, &syntax.Error{Code: syntax.ErrMissingBracket, Expr: invalidRegexp[3:4]})},
		"valid json":         {"validJSON", validJSON, FormatJSON, nil},
		"invalid json":       {"invalidJSON", invalidJSON, FormatJSON, InvalidFormatError("invalidJSON", invalidJSON, FormatJSON, fmt.Errorf("invalid JSON"))},
		"valid rfc1123":      {"validRFC1123", validRFC1123, FormatRFC1123, nil},
		"invalid rfc1123":    {"invalidRFC1123", invalidRFC1123, FormatRFC1123, InvalidFormatError("invalidRFC1123", invalidRFC1123, FormatRFC1123, &time.ParseError{Layout: time.RFC1123, Value: invalidRFC1123, LayoutElem: ", ", ValueElem: invalidRFC1123[3:]})},
	}

	for k, tc := range cases {
		actual := ValidateFormat(tc.name, tc.val, tc.format)
		if !errors.Is(actual, tc.expected) {
			// Compare only the messages because the error has always a new error ID.
			if actual == nil || tc.expected == nil || actual.Error() != tc.expected.Error() {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestValidateFormatURIReference(t *testing.T) {
	cases := []struct {
		name  string
		val   string
		valid bool
	}{
		{"absolute uri", "https://example.com/a?b=c", true},
		{"urn", "urn:isbn:0451450523", true},
		{"empty", "", true},
		{"absolute path", "/users/123", true},
		{"network path", "//example.com/a", true},
		{"relative path", "users/123", true},
		{"dot segments", "../a", true},
		{"query only", "?q=1", true},
		{"fragment only", "#frag", true},
		{"ipv6 host", "http://[::1]:8080/a", true},
		{"pct-encoded", "/a%20b", true},
		{"sub-delims", "/a;b=c,d!$&'()*+", true},
		{"space", "/a b", false},
		{"non-ascii", "/caf\u00e9", false},
		{"control character", "/a\x7f", false},
		{"backslash", `a\b`, false},
		{"invalid pct-encoding", "/a%zz", false},
		{"empty scheme", ":a", false},
		{"colon in first relative segment", "1a:b", false},
		{"unclosed ipv6 host", "http://[::1/a", false},
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

func TestValidatePattern(t *testing.T) {
	var (
		name      = "foo"
		pattern   = "^goa$"
		matched   = "goa"
		unmatched = "foo["
	)
	cases := map[string]struct {
		name     string
		val      string
		pattern  string
		expected error
	}{
		"matched value":   {name, matched, pattern, nil},
		"unmatched value": {name, unmatched, pattern, InvalidPatternError(name, unmatched, pattern)},
	}

	for k, tc := range cases {
		actual := ValidatePattern(tc.name, tc.val, tc.pattern)
		if !errors.Is(actual, tc.expected) {
			// Compare only the messages because the error has always a new error ID.
			if actual.Error() != tc.expected.Error() {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestValidateFormatHostname(t *testing.T) {
	label63 := strings.Repeat("a", 63)
	label64 := strings.Repeat("a", 64)
	host253 := label63 + "." + label63 + "." + label63 + "." + strings.Repeat("b", 61)
	host254 := label63 + "." + label63 + "." + label63 + "." + strings.Repeat("b", 62)

	cases := []struct {
		name    string
		val     string
		wantErr string
	}{
		{name: "two labels", val: "goa.design"},
		{name: "three labels", val: "api.example.com"},
		{name: "mixed case", val: "API.Example.COM"},
		{name: "single label word", val: "localhost"},
		{name: "single letter", val: "a"},
		{name: "single uppercase letter", val: "Z"},
		{name: "single label with hyphen and digit", val: "db-1"},
		{name: "single label starting with digit", val: "1a"},
		{name: "single label digits with hyphen", val: "1-2"},
		{name: "digit-leading label", val: "0.pool.ntp.org"},
		{name: "all-numeric non-final labels", val: "123.456.com"},
		{name: "interior hyphens", val: "my-service-01.prod-us-east-1.internal"},
		{name: "punycode first label", val: "xn--nxasmq6b.com"},
		{name: "punycode top-level label", val: "example.xn--p1ai"},
		{name: "max 63-char single label", val: label63},
		{name: "max 63-char first label", val: label63 + ".com"},
		{name: "max 253-char hostname", val: host253},

		{name: "empty", val: "", wantErr: "hostname cannot be empty"},
		{name: "254-char hostname", val: host254, wantErr: "hostname exceeds 253 characters"},
		{name: "64-char single label", val: label64, wantErr: "hostname label exceeds 63 characters"},
		{name: "64-char first label", val: label64 + ".com", wantErr: "hostname label exceeds 63 characters"},
		{name: "64-char final label", val: "example." + label64, wantErr: "hostname label exceeds 63 characters"},
		{name: "root dot", val: ".", wantErr: "hostname cannot contain empty label"},
		{name: "leading dot", val: ".example.com", wantErr: "hostname cannot contain empty label"},
		{name: "consecutive dots", val: "example..com", wantErr: "hostname cannot contain empty label"},
		{name: "trailing dot single label", val: "example.", wantErr: "hostname cannot end with '.'"},
		{name: "trailing dot multi label", val: "example.com.", wantErr: "hostname cannot end with '.'"},
		{name: "hyphen only", val: "-", wantErr: "hostname label cannot start with '-'"},
		{name: "leading hyphen first label", val: "-example.com", wantErr: "hostname label cannot start with '-'"},
		{name: "leading hyphen inner label", val: "api.-example.com", wantErr: "hostname label cannot start with '-'"},
		{name: "trailing hyphen single label", val: "a-", wantErr: "hostname label cannot end with '-'"},
		{name: "trailing hyphen first label", val: "example-.com", wantErr: "hostname label cannot end with '-'"},
		{name: "trailing hyphen final label", val: "example.com-", wantErr: "hostname label cannot end with '-'"},
		{name: "single digit", val: "1", wantErr: "hostname top-level label cannot be all-numeric"},
		{name: "multiple digits", val: "123", wantErr: "hostname top-level label cannot be all-numeric"},
		{name: "ipv4 literal", val: "127.0.0.1", wantErr: "hostname top-level label cannot be all-numeric"},
		{name: "all-numeric final label", val: "example.123", wantErr: "hostname top-level label cannot be all-numeric"},
		{name: "underscore", val: "_hi_", wantErr: "hostname contains invalid character '_'"},
		{name: "srv record underscore", val: "_sip._tcp.example.com", wantErr: "hostname contains invalid character '_'"},
		{name: "space path and query after alnum prefix", val: "ab internal-admin/secret?x=1", wantErr: "hostname contains invalid character ' '"},
		{name: "crlf injection after alnum prefix", val: "ab\r\nX-Injected: evil", wantErr: "hostname contains invalid character '\\r'"},
		{name: "special chars ending in letter", val: "@#$%^&*()_+=letMeIn/../../etc/passwdX", wantErr: "hostname contains invalid character '@'"},
		{name: "digit followed by at", val: "1@", wantErr: "hostname contains invalid character '@'"},
		{name: "host with port", val: "example.com:8080", wantErr: "hostname contains invalid character ':'"},
		{name: "scheme prefix", val: "https://example.com", wantErr: "hostname contains invalid character ':'"},
		{name: "userinfo prefix", val: "user@example.com", wantErr: "hostname contains invalid character '@'"},
		{name: "ipv6 literal", val: "::1", wantErr: "hostname contains invalid character ':'"},
		{name: "nul byte", val: "ab\x00cd.com", wantErr: "hostname contains invalid character '\\x00'"},
		{name: "del byte", val: "ab\x7fcd.com", wantErr: "hostname contains invalid character '\\x7f'"},
		{name: "non-ascii rune", val: "caf\u00e9.example.com", wantErr: "hostname contains invalid character 'é'"},
		{name: "invalid utf8 byte", val: "bad\xff.example.com", wantErr: "hostname contains non-ASCII byte 0xff"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFormat("host", tc.val, FormatHostname)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			var se *ServiceError
			if assert.ErrorAs(t, err, &se) {
				assert.Equal(t, InvalidFormat, se.Name)
				if assert.NotNil(t, se.Field) {
					assert.Equal(t, "host", *se.Field)
				}
			}
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
