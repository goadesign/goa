package goa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sync"
	"time"
)

// Format defines a validation format.
type Format string

const (
	// FormatDate describes RFC3339 date values.
	FormatDate Format = "date"

	// FormatDateTime describes RFC3339 date time values.
	FormatDateTime Format = "date-time"

	// FormatUUID describes RFC4122 UUID values.
	FormatUUID = "uuid"

	// FormatEmail describes RFC5322 email addresses.
	FormatEmail = "email"

	// FormatHostname describes RFC1035 Internet hostnames.
	FormatHostname = "hostname"

	// FormatIPv4 describes RFC2373 IPv4 address values.
	FormatIPv4 = "ipv4"

	// FormatIPv6 describes RFC2373 IPv6 address values.
	FormatIPv6 = "ipv6"

	// FormatIP describes RFC2373 IPv4 or IPv6 address values.
	FormatIP = "ip"

	// FormatURI describes RFC3986 URI values.
	FormatURI = "uri"

	// FormatMAC describes IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address values.
	FormatMAC = "mac"

	// FormatCIDR describes RFC4632 and RFC4291 CIDR notation IP address values.
	FormatCIDR = "cidr"

	// FormatRegexp describes regular expression syntax accepted by RE2.
	FormatRegexp = "regexp"

	// FormatJSON describes JSON text.
	FormatJSON = "json"

	// FormatRFC1123 describes RFC1123 date time values.
	FormatRFC1123 = "rfc1123"
)

var (
	hostnameRegex  = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
	ipv4Regex      = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	uuidURNPrefix  = []byte("urn:uuid:")
	uuidByteGroups = []int{8, 4, 4, 4, 12}
)

// ValidateFormat validates val against f. It returns nil if the string conforms
// to the format, an error otherwise. name is the name of the variable used in
// error messages. where in a data structure the error occurred if any. The
// format specification follows the json schema draft 4 validation extension.
// see http://json-schema.org/latest/json-schema-validation.html#anchor105
// Supported formats are:
//
//     - "date": RFC3339 date value
//     - "date-time": RFC3339 date time value
//     - "email": RFC5322 email address
//     - "hostname": RFC1035 Internet host name
//     - "ipv4", "ipv6", "ip": RFC2673 and RFC2373 IP address values
//     - "uri": RFC3986 URI value
//     - "mac": IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address value
//     - "cidr": RFC4632 and RFC4291 CIDR notation IP address value
//     - "regexp": Regular expression syntax accepted by RE2
//     - "rfc1123": RFC1123 date time value
func ValidateFormat(name string, val string, f Format) error {
	var err error
	switch f {
	case FormatDate:
		_, err = time.Parse("2006-01-02", val)
	case FormatDateTime:
		_, err = time.Parse(time.RFC3339, val)
	case FormatUUID:
		err = validateUUID(val)
	case FormatEmail:
		_, err = mail.ParseAddress(val)
	case FormatHostname:
		if !hostnameRegex.MatchString(val) {
			err = fmt.Errorf("hostname value '%s' does not match %s",
				val, hostnameRegex.String())
		}
	case FormatIPv4, FormatIPv6, FormatIP:
		ip := net.ParseIP(val)
		if ip == nil {
			err = fmt.Errorf("\"%s\" is an invalid %s value", val, f)
		}
		if f == FormatIPv4 {
			if !ipv4Regex.MatchString(val) {
				err = fmt.Errorf("\"%s\" is an invalid ipv4 value", val)
			}
		}
		if f == FormatIPv6 {
			if ipv4Regex.MatchString(val) {
				err = fmt.Errorf("\"%s\" is an invalid ipv6 value", val)
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
	case FormatJSON:
		if !json.Valid([]byte(val)) {
			err = fmt.Errorf("invalid JSON")
		}
	case FormatRFC1123:
		_, err = time.Parse(time.RFC1123, val)
	default:
		return fmt.Errorf("unknown format %#v", f)
	}
	if err != nil {
		return InvalidFormatError(name, val, f, err)
	}
	return nil
}

// knownPatterns records the compiled patterns.
// TBD: refactor all this so that the generated code initializes the map on start to get rid of the
// need for a RW mutex.
var knownPatterns = make(map[string]*regexp.Regexp)

// knownPatternsLock is the mutex used to access knownPatterns
var knownPatternsLock = &sync.RWMutex{}

// ValidatePattern returns an error if val does not match the regular expression
// p. It makes an effort to minimize the number of times the regular expression
// needs to be compiled. name is the name of the variable used in error messages.
func ValidatePattern(name, val, p string) error {
	knownPatternsLock.RLock()
	r, ok := knownPatterns[p]
	knownPatternsLock.RUnlock()
	if !ok {
		r = regexp.MustCompile(p) // DSL validation makes sure regexp is valid
		knownPatternsLock.Lock()
		knownPatterns[p] = r
		knownPatternsLock.Unlock()
	}
	if !r.MatchString(val) {
		return InvalidPatternError(name, val, p)
	}
	return nil
}

// The following formats are supported:
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
// "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
// "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
func validateUUID(uuid string) error {
	if len(uuid) < 32 {
		return fmt.Errorf("uuid: UUID string too short: %s", uuid)
	}
	t := []byte(uuid)
	braced := false
	if bytes.Equal(t[:9], uuidURNPrefix) {
		t = t[9:]
	} else if t[0] == '{' {
		t = t[1:]
		braced = true
	}
	for i, byteGroup := range uuidByteGroups {
		if i > 0 {
			if t[0] != '-' {
				return fmt.Errorf("uuid: invalid string format")
			}
			t = t[1:]
		}
		if len(t) < byteGroup {
			return fmt.Errorf("uuid: UUID string too short: %s", uuid)
		}
		if i == 4 && len(t) > byteGroup &&
			((braced && t[byteGroup] != '}') || len(t[byteGroup:]) > 1 || !braced) {
			return fmt.Errorf("uuid: UUID string too long: %s", uuid)
		}
		t = t[byteGroup:]
	}

	return nil
}
