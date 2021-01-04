package goa

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp/syntax"
	"testing"
	"time"
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
		"invalid date":               {"invalidDate", invalidDate, FormatDate, InvalidFormatError("invalidDate", invalidDate, FormatDate, &time.ParseError{Layout: "2006-01-02", Value: invalidDate, LayoutElem: "-", ValueElem: invalidDate[4:]})},
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
		"invalid hostname":   {"invalidHostname", invalidHostname, FormatHostname, InvalidFormatError("invalidHostname", invalidHostname, FormatHostname, fmt.Errorf("hostname value '%s' does not match %s", invalidHostname, `^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`))},
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
		if actual != tc.expected {
			// Compare only the messages because the error has always a new error ID.
			if actual == nil || tc.expected == nil || actual.Error() != tc.expected.Error() {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
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
		if actual != tc.expected {
			// Compare only the messages because the error has always a new error ID.
			if actual.Error() != tc.expected.Error() {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}
