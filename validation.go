package goa

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"time"
)

// Format defines a validation format.
type Format string

const (
	// FormatDateTime defines RFC3339 date time values.
	FormatDateTime Format = "date-time"

	// FormatEmail defines RFC5322 email addresses.
	FormatEmail = "email"

	// FormatHostname defines RFC1035 Internet host names.
	FormatHostname = "hostname"

	// FormatIPv4 defines RFC2373 IPv4 address values.
	FormatIPv4 = "ipv4"

	// FormatIPv6 defines RFC2373 IPv6 address values.
	FormatIPv6 = "ipv6"

	// FormatURI defines RFC3986 URI values.
	FormatURI = "uri"

	// FormatMAC defines IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address values.
	FormatMAC = "mac"

	// FormatCIDR defines RFC4632 and RFC4291 CIDR notation IP address values.
	FormatCIDR = "cidr"

	// FormatRegexp Regexp defines regular expression syntax accepted by RE2.
	FormatRegexp = "regexp"
)

var (
	// Regular expression used to validate RFC1035 hostnames*/
	hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

	// Simple regular expression for IPv4 values, more rigorous checking is done via net.ParseIP
	ipv4Regex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
)

// ValidateFormat validates a string against a standard format.
// It returns nil if the string conforms to the format, an error otherwise.
// The format specification follows the json schema draft 4 validation extension.
// see http://json-schema.org/latest/json-schema-validation.html#anchor105
// Supported formats are:
//
//     - "date-time": RFC3339 date time value
//     - "email": RFC5322 email address
//     - "hostname": RFC1035 Internet host name
//     - "ipv4" and "ipv6": RFC2673 and RFC2373 IP address values
//     - "uri": RFC3986 URI value
//     - "mac": IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address value
//     - "cidr": RFC4632 and RFC4291 CIDR notation IP address value
//     - "regexp": Regular expression syntax accepted by RE2
func ValidateFormat(f Format, val string) error {
	var err error
	switch f {
	case FormatDateTime:
		_, err = time.Parse(time.RFC3339, val)
	case FormatEmail:
		_, err = mail.ParseAddress(val)
	case FormatHostname:
		if !hostnameRegex.MatchString(val) {
			err = fmt.Errorf("hostname value '%s' does not match %s",
				val, hostnameRegex.String())
		}
	case FormatIPv4, FormatIPv6:
		ip := net.ParseIP(val)
		if ip == nil {
			err = fmt.Errorf("\"%s\" is an invalid %s value", val, f)
		}
		if f == "ipv4" {
			if !ipv4Regex.MatchString(val) {
				err = fmt.Errorf("\"%s\" is an invalid ipv4 value", val)
			}
		}
	case FormatURI:
		_, err = url.ParseRequestURI(val)
	case FormatMAC:
		_, err = net.ParseMAC(val)
	case FormatCIDR:
		_, _, err = net.ParseCIDR(val)
	case FormatRegexp:
		_, err = regexp.Compile(val)
	default:
		return fmt.Errorf("unknown format %#v", f)
	}
	if err != nil {
		go IncrCounter([]string{"goa", "validation", "error", string(f)}, 1.0)
		return fmt.Errorf("invalid %s value, %s", f, err)
	}
	return nil
}

// knownPatterns records the compiled patterns.
var knownPatterns = make(map[string]*regexp.Regexp)

// ValidatePattern returns an error if val does not match the regular expression p.
// It makes an effort to minimize the number of times the regular expression needs to be compiled.
func ValidatePattern(p string, val string) bool {
	r, ok := knownPatterns[p]
	if !ok {
		r = regexp.MustCompile(p) // DSL validation makes sure regexp is valid
		knownPatterns[p] = r
	}
	return r.MatchString(val)
}
