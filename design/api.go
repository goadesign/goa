package design

import (
	"mime"
	"regexp"
	"sort"
	"strings"

	"github.com/goadesign/goa/dslengine"
)

// MediaTypeRoot is the data structure that represents the additional DSL definition root
// that contains the media type definition set created by CollectionOf index by canonical id.
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

	// ProjectedMediaTypes is a cache used by the MediaType strut Project method.
	ProjectedMediaTypes MediaTypeRoot

	// WildcardRegex is the regular expression used to capture path parameters.
	WildcardRegex = regexp.MustCompile(`/(?::|\*)([a-zA-Z0-9_]+)`)

	// DefaultDecoders contains the decoding definitions used when no Consumes DSL is found.
	DefaultDecoders []*EncodingDefinition

	// DefaultEncoders contains the encoding definitions used when no Produces DSL is found.
	DefaultEncoders []*EncodingDefinition

	// KnownEncoders contains the list of encoding packages and factories known by goa indexed
	// by MIME type.
	KnownEncoders = map[string]string{
		"application/json":      "github.com/goadesign/goa",
		"application/xml":       "github.com/goadesign/goa",
		"application/gob":       "github.com/goadesign/goa",
		"application/x-gob":     "github.com/goadesign/goa",
		"application/binc":      "github.com/goadesign/goa/encoding/binc",
		"application/x-binc":    "github.com/goadesign/goa/encoding/binc",
		"application/cbor":      "github.com/goadesign/goa/encoding/cbor",
		"application/x-cbor":    "github.com/goadesign/goa/encoding/cbor",
		"application/msgpack":   "github.com/goadesign/goa/encoding/msgpack",
		"application/x-msgpack": "github.com/goadesign/goa/encoding/msgpack",
	}

	// KnownEncoderFunctions contains the list of encoding encoder and decoder functions known
	// by goa indexed by MIME type.
	KnownEncoderFunctions = map[string][2]string{
		"application/json":      {"NewJSONEncoder", "NewJSONDecoder"},
		"application/xml":       {"NewXMLEncoder", "NewXMLDecoder"},
		"application/gob":       {"NewGobEncoder", "NewGobDecoder"},
		"application/x-gob":     {"NewGobEncoder", "NewGobDecoder"},
		"application/binc":      {"NewEncoder", "NewDecoder"},
		"application/x-binc":    {"NewEncoder", "NewDecoder"},
		"application/cbor":      {"NewEncoder", "NewDecoder"},
		"application/x-cbor":    {"NewEncoder", "NewDecoder"},
		"application/msgpack":   {"NewEncoder", "NewDecoder"},
		"application/x-msgpack": {"NewEncoder", "NewDecoder"},
	}

	// JSONContentTypes list the Content-Type header values that cause goa to encode or decode
	// JSON by default.
	JSONContentTypes = []string{"application/json"}

	// XMLContentTypes list the Content-Type header values that cause goa to encode or decode
	// XML by default.
	XMLContentTypes = []string{"application/xml"}

	// GobContentTypes list the Content-Type header values that cause goa to encode or decode
	// Gob by default.
	GobContentTypes = []string{"application/gob", "application/x-gob"}

	// ErrorMediaIdentifier is the media type identifier used for error responses.
	ErrorMediaIdentifier = "application/vnd.goa.error"

	// ErrorMedia is the built-in media type for error responses.
	ErrorMedia = &MediaTypeDefinition{
		UserTypeDefinition: &UserTypeDefinition{
			AttributeDefinition: &AttributeDefinition{
				Type:        errorMediaType,
				Description: "Error response media type",
				Example: map[string]interface{}{
					"id":     "3F1FKVRR",
					"status": "400",
					"code":   "invalid_value",
					"detail": "Value of ID must be an integer",
					"meta":   map[string]interface{}{"timestamp": 1458609066},
				},
			},
			TypeName: "error",
		},
		Identifier: ErrorMediaIdentifier,
		Views:      map[string]*ViewDefinition{"default": errorMediaView},
	}

	errorMediaType = Object{
		"id": &AttributeDefinition{
			Type:        String,
			Description: "a unique identifier for this particular occurrence of the problem.",
			Example:     "3F1FKVRR",
		},
		"status": &AttributeDefinition{
			Type:        String,
			Description: "the HTTP status code applicable to this problem, expressed as a string value.",
			Example:     "400",
		},
		"code": &AttributeDefinition{
			Type:        String,
			Description: "an application-specific error code, expressed as a string value.",
			Example:     "invalid_value",
		},
		"detail": &AttributeDefinition{
			Type:        String,
			Description: "a human-readable explanation specific to this occurrence of the problem.",
			Example:     "Value of ID must be an integer",
		},
		"meta": &AttributeDefinition{
			Type: &Hash{
				KeyType:  &AttributeDefinition{Type: String},
				ElemType: &AttributeDefinition{Type: Any},
			},
			Description: "a meta object containing non-standard meta-information about the error.",
			Example:     map[string]interface{}{"timestamp": 1458609066},
		},
	}

	errorMediaView = &ViewDefinition{
		AttributeDefinition: &AttributeDefinition{Type: errorMediaType},
		Name:                "default",
	}
)

func init() {
	goa := "github.com/goadesign/goa"
	DefaultEncoders = []*EncodingDefinition{
		{MIMETypes: JSONContentTypes, PackagePath: goa, Function: "NewJSONEncoder"},
		{MIMETypes: XMLContentTypes, PackagePath: goa, Function: "NewXMLEncoder"},
		{MIMETypes: GobContentTypes, PackagePath: goa, Function: "NewGobEncoder"},
	}
	DefaultDecoders = []*EncodingDefinition{
		{MIMETypes: JSONContentTypes, PackagePath: goa, Function: "NewJSONDecoder"},
		{MIMETypes: XMLContentTypes, PackagePath: goa, Function: "NewXMLDecoder"},
		{MIMETypes: GobContentTypes, PackagePath: goa, Function: "NewGobDecoder"},
	}
	errorMediaView.Parent = ErrorMedia
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
	return KnownEncoders[mimeType] != ""
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
