package design

import (
	"mime"
	"regexp"
	"sort"
	"strings"

	"github.com/goadesign/goa/dslengine"
)

// MediaTypeRoot is the data structure that represents the additional DSL definition root
// that contains the media type definition set created by CollectionOf.
type MediaTypeRoot map[string]*MediaTypeDefinition

// List of all built-in response names.
const (
	Continue           = "Continue"
	SwitchingProtocols = "SwitchingProtocols"

	OK                   = "OK"
	Created              = "Created"
	Accepted             = "Accepted"
	NonAuthoritativeInfo = "NonAuthoritativeInfo"
	NoContent            = "NoContent"
	ResetContent         = "ResetContent"
	PartialContent       = "PartialContent"

	MultipleChoices   = "MultipleChoices"
	MovedPermanently  = "MovedPermanently"
	Found             = "Found"
	SeeOther          = "SeeOther"
	NotModified       = "NotModified"
	UseProxy          = "UseProxy"
	TemporaryRedirect = "TemporaryRedirect"

	BadRequest                   = "BadRequest"
	Unauthorized                 = "Unauthorized"
	PaymentRequired              = "PaymentRequired"
	Forbidden                    = "Forbidden"
	NotFound                     = "NotFound"
	MethodNotAllowed             = "MethodNotAllowed"
	NotAcceptable                = "NotAcceptable"
	ProxyAuthRequired            = "ProxyAuthRequired"
	RequestTimeout               = "RequestTimeout"
	Conflict                     = "Conflict"
	Gone                         = "Gone"
	LengthRequired               = "LengthRequired"
	PreconditionFailed           = "PreconditionFailed"
	RequestEntityTooLarge        = "RequestEntityTooLarge"
	RequestURITooLong            = "RequestURITooLong"
	UnsupportedMediaType         = "UnsupportedMediaType"
	RequestedRangeNotSatisfiable = "RequestedRangeNotSatisfiable"
	ExpectationFailed            = "ExpectationFailed"
	Teapot                       = "Teapot"
	UnprocessableEntity          = "UnprocessableEntity"

	InternalServerError     = "InternalServerError"
	NotImplemented          = "NotImplemented"
	BadGateway              = "BadGateway"
	ServiceUnavailable      = "ServiceUnavailable"
	GatewayTimeout          = "GatewayTimeout"
	HTTPVersionNotSupported = "HTTPVersionNotSupported"
)

var (
	// Design being built by DSL.
	Design *APIDefinition

	// GeneratedMediaTypes contains DSL definitions that were created by the design DSL and
	// need to be executed as a second pass.
	// An example of this are media types defined with CollectionOf: the element media type
	// must be defined first then the definition created by CollectionOf must execute.
	GeneratedMediaTypes MediaTypeRoot

	// WildcardRegex is the regular expression used to capture path parameters.
	WildcardRegex = regexp.MustCompile(`/(?::|\*)([a-zA-Z0-9_]+)`)

	// DefaultDecoders contains the decoding definitions used when no Consumes DSL is found.
	DefaultDecoders []*EncodingDefinition

	// DefaultEncoders contains the encoding definitions used when no Produces DSL is found.
	DefaultEncoders []*EncodingDefinition

	// KnownEncoders contains the list of encoding packages and factories known by goa indexed
	// by MIME type.
	KnownEncoders = map[string][3]string{
		"application/json":      {"json", "JSONEncoderFactory", "JSONDecoderFactory"},
		"application/xml":       {"xml", "XMLEncoderFactory", "XMLDecoderFactory"},
		"text/xml":              {"xml", "XMLEncoderFactory", "XMLDecoderFactory"},
		"application/gob":       {"gob", "GobEncoderFactory", "GobDecoderFactory"},
		"application/x-gob":     {"gob", "GobEncoderFactory", "GobDecoderFactory"},
		"application/binc":      {"github.com/goadesign/encoding/binc", "EncoderFactory", "DecoderFactory"},
		"application/x-binc":    {"github.com/goadesign/encoding/binc", "EncoderFactory", "DecoderFactory"},
		"application/x-cbor":    {"github.com/goadesign/encoding/cbor", "EncoderFactory", "DecoderFactory"},
		"application/cbor":      {"github.com/goadesign/encoding/cbor", "EncoderFactory", "DecoderFactory"},
		"application/msgpack":   {"github.com/goadesign/encoding/msgpack", "EncoderFactory", "DecoderFactory"},
		"application/x-msgpack": {"github.com/goadesign/encoding/msgpack", "EncoderFactory", "DecoderFactory"},
	}

	// JSONContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/json to unmarshal unless overwritten using SetDecoder
	JSONContentTypes = []string{"application/json"}

	// XMLContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/xml to unmarshal unless overwritten using SetDecoder
	XMLContentTypes = []string{"application/xml", "text/xml"}

	// GobContentTypes is a slice of default Content-Type headers that will use stdlib
	// encoding/gob to unmarshal unless overwritten using SetDecoder
	GobContentTypes = []string{"application/gob", "application/x-gob"}
)

func init() {
	var types []string
	types = append(types, JSONContentTypes...)
	types = append(types, XMLContentTypes...)
	types = append(types, GobContentTypes...)
	DefaultEncoders = []*EncodingDefinition{{MIMETypes: types}}
	DefaultDecoders = []*EncodingDefinition{{MIMETypes: types}}
}

// CanonicalIdentifier returns the media type identifier sans suffix
// which is what the DSL uses to store and lookup media types.
func CanonicalIdentifier(identifier string) string {
	base, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		return identifier
	}
	id := base
	if i := strings.Index(id, "+"); i != -1 {
		id = id[:i]
	}
	return mime.FormatMediaType(id, params)
}

// HasKnownEncoder returns true if the encoder for the given MIME type is known by goa.
// MIME types with unknown encoders must be associated with a package path explicitly in the DSL.
func HasKnownEncoder(mimeType string) bool {
	return KnownEncoders[mimeType][1] != ""
}

// IsGoaEncoder returns true if the encoder for the given MIME type is implemented in the goa
// package.
func IsGoaEncoder(pkgPath string) bool {
	return pkgPath == "json" || pkgPath == "xml" || pkgPath == "gob"
}

// ExtractWildcards returns the names of the wildcards that appear in path.
func ExtractWildcards(path string) []string {
	matches := WildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}

// DSLName is displayed to the user when the DSL executes.
func (r MediaTypeRoot) DSLName() string {
	return "Generated Media Types"
}

// DependsOn return the DSL roots the generated media types DSL root depends on, that's the API DSL.
func (r MediaTypeRoot) DependsOn() []dslengine.Root {
	return []dslengine.Root{Design}
}

// IterateSets iterates over the one generated media type definition set.
func (r MediaTypeRoot) IterateSets(iterator dslengine.SetIterator) {
	canonicalIDs := make([]string, len(r))
	i := 0
	for _, mt := range r {
		canonicalID := CanonicalIdentifier(mt.Identifier)
		Design.MediaTypes[canonicalID] = mt
		canonicalIDs[i] = canonicalID
		i++
	}
	sort.Strings(canonicalIDs)
	set := make([]dslengine.Definition, len(canonicalIDs))
	for i, cid := range canonicalIDs {
		set[i] = Design.MediaTypes[cid]
	}
	iterator(set)
}

// Reset deletes all the keys.
func (r MediaTypeRoot) Reset() {
	for k := range r {
		delete(r, k)
	}
}
