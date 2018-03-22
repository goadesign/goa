// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider HTTP client types
//
// Command:
// $ goa gen goa.design/goa/examples/error/design

package client

import (
	goa "goa.design/goa"
	dividersvc "goa.design/goa/examples/error/gen/divider"
)

// IntegerDivideHasRemainderResponseBody is the type of the "divider" service
// "integer_divide" endpoint HTTP response body for the "has_remainder" error.
type IntegerDivideHasRemainderResponseBody struct {
	// Name is the name of this class of errors.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// Is the error temporary?
	Temporary *bool `form:"temporary,omitempty" json:"temporary,omitempty" xml:"temporary,omitempty"`
	// Is the error a timeout?
	Timeout *bool `form:"timeout,omitempty" json:"timeout,omitempty" xml:"timeout,omitempty"`
}

// IntegerDivideDivByZeroResponseBody is the type of the "divider" service
// "integer_divide" endpoint HTTP response body for the "div_by_zero" error.
type IntegerDivideDivByZeroResponseBody struct {
	// Name is the name of this class of errors.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// Is the error temporary?
	Temporary *bool `form:"temporary,omitempty" json:"temporary,omitempty" xml:"temporary,omitempty"`
	// Is the error a timeout?
	Timeout *bool `form:"timeout,omitempty" json:"timeout,omitempty" xml:"timeout,omitempty"`
}

// DivideDivByZeroResponseBody is the type of the "divider" service "divide"
// endpoint HTTP response body for the "div_by_zero" error.
type DivideDivByZeroResponseBody struct {
	// Name is the name of this class of errors.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// Is the error temporary?
	Temporary *bool `form:"temporary,omitempty" json:"temporary,omitempty" xml:"temporary,omitempty"`
	// Is the error a timeout?
	Timeout *bool `form:"timeout,omitempty" json:"timeout,omitempty" xml:"timeout,omitempty"`
}

// NewIntegerDivideHasRemainder builds a divider service integer_divide
// endpoint has_remainder error.
func NewIntegerDivideHasRemainder(body *IntegerDivideHasRemainderResponseBody) *dividersvc.Error {
	v := &dividersvc.Error{
		Name:      *body.Name,
		ID:        *body.ID,
		Message:   *body.Message,
		Temporary: body.Temporary,
		Timeout:   body.Timeout,
	}
	return v
}

// NewIntegerDivideDivByZero builds a divider service integer_divide endpoint
// div_by_zero error.
func NewIntegerDivideDivByZero(body *IntegerDivideDivByZeroResponseBody) *dividersvc.Error {
	v := &dividersvc.Error{
		Name:      *body.Name,
		ID:        *body.ID,
		Message:   *body.Message,
		Temporary: body.Temporary,
		Timeout:   body.Timeout,
	}
	return v
}

// NewDivideDivByZero builds a divider service divide endpoint div_by_zero
// error.
func NewDivideDivByZero(body *DivideDivByZeroResponseBody) *dividersvc.Error {
	v := &dividersvc.Error{
		Name:      *body.Name,
		ID:        *body.ID,
		Message:   *body.Message,
		Temporary: body.Temporary,
		Timeout:   body.Timeout,
	}
	return v
}

// Validate runs the validations defined on
// IntegerDivideHasRemainderResponseBody
func (body *IntegerDivideHasRemainderResponseBody) Validate() (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Message == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("message", "body"))
	}
	return
}

// Validate runs the validations defined on IntegerDivideDivByZeroResponseBody
func (body *IntegerDivideDivByZeroResponseBody) Validate() (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Message == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("message", "body"))
	}
	return
}

// Validate runs the validations defined on DivideDivByZeroResponseBody
func (body *DivideDivByZeroResponseBody) Validate() (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Message == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("message", "body"))
	}
	return
}
