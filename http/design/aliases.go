//************************************************************************//
// Code generated with aliaser, DO NOT EDIT.
//
// Aliased DSL Constants
//************************************************************************//

package design

import (
	design "goa.design/goa/design"
)

const (
	// FormatDateTime describes RFC3339 date time values.
	FormatDateTime = design.FormatDateTime
	// FormatUUID describes RFC4122 UUID values.
	FormatUUID = design.FormatUUID
	// FormatEmail describes RFC5322 email addresses.
	FormatEmail = design.FormatEmail
	// FormatHostname describes RFC1035 Internet hostnames.
	FormatHostname = design.FormatHostname
	// FormatIPv4 describes RFC2373 IPv4 address values.
	FormatIPv4 = design.FormatIPv4
	// FormatIPv6 describes RFC2373 IPv6 address values.
	FormatIPv6 = design.FormatIPv6
	// FormatIP describes RFC2373 IPv4 or IPv6 address values.
	FormatIP = design.FormatIP
	// FormatURI describes RFC3986 URI values.
	FormatURI = design.FormatURI
	// FormatMAC describes IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address values.
	FormatMAC = design.FormatMAC
	// FormatCIDR describes RFC4632 and RFC4291 CIDR notation IP address values.
	FormatCIDR = design.FormatCIDR
	// FormatRegexp describes regular expression syntax accepted by RE2.
	FormatRegexp = design.FormatRegexp
	// FormatJSON describes JSON text.
	FormatJSON = design.FormatJSON
	// FormatRFC1123 describes RFC1123 date time values.
	FormatRFC1123 = design.FormatRFC1123
)

const (
	// DefaultView is the name of the default result type view.
	DefaultView = design.DefaultView
)

const (
	// OAuth2Kind identifies a "OAuth2" security scheme.
	OAuth2Kind = design.OAuth2Kind
	// BasicAuthKind means "basic" security scheme.
	BasicAuthKind = design.BasicAuthKind
	// APIKeyKind means "apiKey" security scheme.
	APIKeyKind = design.APIKeyKind
	// JWTKind means an "apiKey" security scheme, with support for
	// TokenPath and Scopes.
	JWTKind = design.JWTKind
	// NoKind means to have no security for this endpoint.
	NoKind = design.NoKind
)

const (
	// AuthorizationCodeFlowKind identifies a OAuth2 authorization code
	// flow.
	AuthorizationCodeFlowKind = design.AuthorizationCodeFlowKind
	// ImplicitFlowKind identifiers a OAuth2 implicit flow.
	ImplicitFlowKind = design.ImplicitFlowKind
	// PasswordFlowKind identifies a Resource Owner Password flow.
	PasswordFlowKind = design.PasswordFlowKind
	// ClientCredentialsFlowKind identifies a OAuth Client Credentials flow.
	ClientCredentialsFlowKind = design.ClientCredentialsFlowKind
)

const (
	// BooleanKind represents a boolean.
	BooleanKind = design.BooleanKind
	// IntKind represents a signed integer.
	IntKind = design.IntKind
	// Int32Kind represents a signed 32-bit integer.
	Int32Kind = design.Int32Kind
	// Int64Kind represents a signed 64-bit integer.
	Int64Kind = design.Int64Kind
	// UIntKind represents an unsigned integer.
	UIntKind = design.UIntKind
	// UInt32Kind represents an unsigned 32-bit integer.
	UInt32Kind = design.UInt32Kind
	// UInt64Kind represents an unsigned 64-bit integer.
	UInt64Kind = design.UInt64Kind
	// Float32Kind represents a 32-bit floating number.
	Float32Kind = design.Float32Kind
	// Float64Kind represents a 64-bit floating number.
	Float64Kind = design.Float64Kind
	// StringKind represents a JSON string.
	StringKind = design.StringKind
	// BytesKind represent a series of bytes (binary data).
	BytesKind = design.BytesKind
	// ArrayKind represents a JSON array.
	ArrayKind = design.ArrayKind
	// ObjectKind represents a JSON object.
	ObjectKind = design.ObjectKind
	// MapKind represents a JSON object where the keys are not known in
	// advance.
	MapKind = design.MapKind
	// UserTypeKind represents a user type.
	UserTypeKind = design.UserTypeKind
	// ResultTypeKind represents a result type.
	ResultTypeKind = design.ResultTypeKind
	// AnyKind represents an unknown type.
	AnyKind = design.AnyKind
)

const (
	// Boolean is the type for a JSON boolean.
	Boolean = design.Boolean
	// Int is the type for a signed integer.
	Int = design.Int
	// Int32 is the type for a signed 32-bit integer.
	Int32 = design.Int32
	// Int64 is the type for a signed 64-bit integer.
	Int64 = design.Int64
	// UInt is the type for an unsigned integer.
	UInt = design.UInt
	// UInt32 is the type for an unsigned 32-bit integer.
	UInt32 = design.UInt32
	// UInt64 is the type for an unsigned 64-bit integer.
	UInt64 = design.UInt64
	// Float32 is the type for a 32-bit floating number.
	Float32 = design.Float32
	// Float64 is the type for a 64-bit floating number.
	Float64 = design.Float64
	// String is the type for a JSON string.
	String = design.String
	// Bytes is the type for binary data.
	Bytes = design.Bytes
	// Any is the type for an arbitrary JSON value (interface{} in Go).
	Any = design.Any
)
