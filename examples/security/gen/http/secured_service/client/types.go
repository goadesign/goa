// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// secured_service HTTP client types
//
// Command:
// $ goa gen goa.design/plugins/security/examples/multi_auth/design

package client

import (
	securedservice "goa.design/plugins/security/examples/multi_auth/gen/secured_service"
)

// SigninUnauthorizedResponseBody is the type of the "secured_service" service
// "signin" endpoint HTTP response body for the "unauthorized" error.
type SigninUnauthorizedResponseBody string

// SecureUnauthorizedResponseBody is the type of the "secured_service" service
// "secure" endpoint HTTP response body for the "unauthorized" error.
type SecureUnauthorizedResponseBody string

// DoublySecureUnauthorizedResponseBody is the type of the "secured_service"
// service "doubly_secure" endpoint HTTP response body for the "unauthorized"
// error.
type DoublySecureUnauthorizedResponseBody string

// AlsoDoublySecureUnauthorizedResponseBody is the type of the
// "secured_service" service "also_doubly_secure" endpoint HTTP response body
// for the "unauthorized" error.
type AlsoDoublySecureUnauthorizedResponseBody string

// NewSigninUnauthorized builds a secured_service service signin endpoint
// unauthorized error.
func NewSigninUnauthorized(body SigninUnauthorizedResponseBody) securedservice.Unauthorized {
	v := securedservice.Unauthorized(body)
	return v
}

// NewSecureUnauthorized builds a secured_service service secure endpoint
// unauthorized error.
func NewSecureUnauthorized(body SecureUnauthorizedResponseBody) securedservice.Unauthorized {
	v := securedservice.Unauthorized(body)
	return v
}

// NewDoublySecureUnauthorized builds a secured_service service doubly_secure
// endpoint unauthorized error.
func NewDoublySecureUnauthorized(body DoublySecureUnauthorizedResponseBody) securedservice.Unauthorized {
	v := securedservice.Unauthorized(body)
	return v
}

// NewAlsoDoublySecureUnauthorized builds a secured_service service
// also_doubly_secure endpoint unauthorized error.
func NewAlsoDoublySecureUnauthorized(body AlsoDoublySecureUnauthorizedResponseBody) securedservice.Unauthorized {
	v := securedservice.Unauthorized(body)
	return v
}
