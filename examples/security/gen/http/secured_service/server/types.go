// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// secured_service HTTP server types
//
// Command:
// $ goa gen goa.design/plugins/security/examples/multi_auth/design

package server

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

// NewSigninUnauthorizedResponseBody builds the HTTP response body from the
// result of the "signin" endpoint of the "secured_service" service.
func NewSigninUnauthorizedResponseBody(res securedservice.Unauthorized) SigninUnauthorizedResponseBody {
	body := SigninUnauthorizedResponseBody(res)
	return body
}

// NewSecureUnauthorizedResponseBody builds the HTTP response body from the
// result of the "secure" endpoint of the "secured_service" service.
func NewSecureUnauthorizedResponseBody(res securedservice.Unauthorized) SecureUnauthorizedResponseBody {
	body := SecureUnauthorizedResponseBody(res)
	return body
}

// NewDoublySecureUnauthorizedResponseBody builds the HTTP response body from
// the result of the "doubly_secure" endpoint of the "secured_service" service.
func NewDoublySecureUnauthorizedResponseBody(res securedservice.Unauthorized) DoublySecureUnauthorizedResponseBody {
	body := DoublySecureUnauthorizedResponseBody(res)
	return body
}

// NewAlsoDoublySecureUnauthorizedResponseBody builds the HTTP response body
// from the result of the "also_doubly_secure" endpoint of the
// "secured_service" service.
func NewAlsoDoublySecureUnauthorizedResponseBody(res securedservice.Unauthorized) AlsoDoublySecureUnauthorizedResponseBody {
	body := AlsoDoublySecureUnauthorizedResponseBody(res)
	return body
}

// NewSigninSigninPayload builds a secured_service service signin endpoint
// payload.
func NewSigninSigninPayload() *securedservice.SigninPayload {
	return &securedservice.SigninPayload{}
}

// NewSecureSecurePayload builds a secured_service service secure endpoint
// payload.
func NewSecureSecurePayload(fail *bool, token *string) *securedservice.SecurePayload {
	return &securedservice.SecurePayload{
		Fail:  fail,
		Token: token,
	}
}

// NewDoublySecureDoublySecurePayload builds a secured_service service
// doubly_secure endpoint payload.
func NewDoublySecureDoublySecurePayload(key *string, token *string) *securedservice.DoublySecurePayload {
	return &securedservice.DoublySecurePayload{
		Key:   key,
		Token: token,
	}
}

// NewAlsoDoublySecureAlsoDoublySecurePayload builds a secured_service service
// also_doubly_secure endpoint payload.
func NewAlsoDoublySecureAlsoDoublySecurePayload(key *string, oauthToken *string, token *string) *securedservice.AlsoDoublySecurePayload {
	return &securedservice.AlsoDoublySecurePayload{
		Key:        key,
		OauthToken: oauthToken,
		Token:      token,
	}
}
