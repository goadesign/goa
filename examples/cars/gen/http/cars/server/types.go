// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// cars HTTP server types
//
// Command:
// $ goa gen goa.design/goa/examples/cars/design -o
// $(GOPATH)/src/goa.design/goa/examples/cars

package server

import (
	carssvc "goa.design/goa/examples/cars/gen/cars"
)

// ListResponseBody is the type of the "cars" service "list" endpoint HTTP
// response body.
type ListResponseBody struct {
	// The make of the car
	Make *string `form:"make,omitempty" json:"make,omitempty" xml:"make,omitempty"`
	// The car model
	Model *string `form:"model,omitempty" json:"model,omitempty" xml:"model,omitempty"`
	// The car body style
	BodyStyle *string `form:"body_style,omitempty" json:"body_style,omitempty" xml:"body_style,omitempty"`
}

// LoginUnauthorizedResponseBody is the type of the "cars" service "login"
// endpoint HTTP response body for the "unauthorized" error.
type LoginUnauthorizedResponseBody string

// ListInvalidScopesResponseBody is the type of the "cars" service "list"
// endpoint HTTP response body for the "invalid-scopes" error.
type ListInvalidScopesResponseBody string

// ListUnauthorizedResponseBody is the type of the "cars" service "list"
// endpoint HTTP response body for the "unauthorized" error.
type ListUnauthorizedResponseBody string

// NewListResponseBody builds the HTTP response body from the result of the
// "list" endpoint of the "cars" service.
func NewListResponseBody(res *carssvc.Car) *ListResponseBody {
	body := &ListResponseBody{
		Make:      &res.Make,
		Model:     &res.Model,
		BodyStyle: &res.BodyStyle,
	}
	return body
}

// NewLoginUnauthorizedResponseBody builds the HTTP response body from the
// result of the "login" endpoint of the "cars" service.
func NewLoginUnauthorizedResponseBody(res carssvc.Unauthorized) LoginUnauthorizedResponseBody {
	body := LoginUnauthorizedResponseBody(res)
	return body
}

// NewListInvalidScopesResponseBody builds the HTTP response body from the
// result of the "list" endpoint of the "cars" service.
func NewListInvalidScopesResponseBody(res carssvc.InvalidScopes) ListInvalidScopesResponseBody {
	body := ListInvalidScopesResponseBody(res)
	return body
}

// NewListUnauthorizedResponseBody builds the HTTP response body from the
// result of the "list" endpoint of the "cars" service.
func NewListUnauthorizedResponseBody(res carssvc.Unauthorized) ListUnauthorizedResponseBody {
	body := ListUnauthorizedResponseBody(res)
	return body
}

// NewLoginLoginPayload builds a cars service login endpoint payload.
func NewLoginLoginPayload() *carssvc.LoginPayload {
	return &carssvc.LoginPayload{}
}

// NewListListPayload builds a cars service list endpoint payload.
func NewListListPayload(style string, token string) *carssvc.ListPayload {
	return &carssvc.ListPayload{
		Style: style,
		Token: token,
	}
}
