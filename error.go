/*
Package goa standardizes on structured error responses: a request that fails
because of an invalid input or an unexpected condition produces a response that
contains a structured error.

By default, the error data structures returned to clients contain four fields:
an id, a message and two boolean values indicating whether the error is
temporary and/or a timeout.

* The id is unique for each occurrence of the error, it helps correlate the
  content of the response with the content of the service logs for example.

* The message contains is specific to the error occurrence and is intended for
  human consumption.

* The temporary and timeout fields helps clients determine whether the request
  should be retried.

Instances of Error can be created via the NewXXXError functions. The generated
code uses these functions to produce the error responses in case of request
validation errors.
*/
package goa

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type (
	// ServiceError is the error type used by the goa package to encode and
	// decode error responses.
	ServiceError struct {
		// ID is a unique value for each occurrence of the error.
		ID string
		// Message contains the specific error details.
		Message string
		// Is the error a timeout?
		Timeout bool
		// Is the error temporary?
		Temporary bool
	}
)

// PermanentError is an error class that indicates that the error is
// definitive and that retrying the request is not needed.
func PermanentError(format string, v ...interface{}) *ServiceError {
	return newError(false, false, format, v...)
}

// TemporaryError is an error class that indicates that the error is
// temporary and that retrying the request may be successful.
func TemporaryError(format string, v ...interface{}) *ServiceError {
	return newError(false, true, format, v...)
}

// PermanentTimeoutError is an error class that indicates a timeout
// and that retrying the request is not needed.
func PermanentTimeoutError(format string, v ...interface{}) *ServiceError {
	return newError(true, false, format, v...)
}

// TemporaryTimeoutError is an error class that indicates a timeout
// and that retrying the request may be successful.
func TemporaryTimeoutError(format string, v ...interface{}) *ServiceError {
	return newError(true, true, format, v...)
}

// MissingPayloadError is the error produced when a request is missing a
// required payload.
func MissingPayloadError() error {
	return PermanentError("missing required payload")
}

// DecodePayloadError is the error produced when a request body cannot be
// decoded successfully.
func DecodePayloadError(msg string) error {
	return PermanentError(msg)
}

// InvalidFieldTypeError is the error produced when the type of a payload field
// does not match the type defined in the design.
func InvalidFieldTypeError(name string, val interface{}, expected string) error {
	return PermanentError("invalid value %#v for %q, must be a %s", val, name, expected)
}

// MissingFieldError is the error produced when a payload is missing a required
// field.
func MissingFieldError(name, context string) error {
	return PermanentError("%q is missing from %s", name, context)
}

// InvalidEnumValueError is the error produced when the value of a payload field
// does not match one the values defined in the design Enum validation.
func InvalidEnumValueError(name string, val interface{}, allowed []interface{}) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	return PermanentError("value of %s must be one of %s but got value %#v", name, strings.Join(elems, ", "), val)
}

// InvalidFormatError is the error produced when the value of a payload field
// does not match the format validation defined in the design.
func InvalidFormatError(name, target string, format Format, formatError error) error {
	return PermanentError("%s must be formatted as a %s but got value %q, %s", name, format, target, formatError.Error())
}

// InvalidPatternError is the error produced when the value of a payload field
// does not match the pattern validation defined in the design.
func InvalidPatternError(name, target string, pattern string) error {
	return PermanentError("%s must match the regexp %q but got value %q", name, pattern, target)
}

// InvalidRangeError is the error produced when the value of a payload field does
// not match the range validation defined in the design. value may be an int or
// a float64.
func InvalidRangeError(name string, target interface{}, value interface{}, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return PermanentError("%s must be %s than %d but got value %#v", name, comp, value, target)
}

// InvalidLengthError is the error produced when the value of a payload field
// does not match the length validation defined in the design.
func InvalidLengthError(name string, target interface{}, ln, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return PermanentError("length of %s must be %s than %d but got value %#v (len=%d)", name, comp, value, target, ln)
}

// MergeErrors updates an error by merging another into it. It first converts
// other into a Error if not already one - producing an internal error in that
// case. The merge algorithm then:
//
// * produces an internal error if any of e or other is an internal error
//
// * produces a bad request error if the status or code of e and other do not
//   match
//
// The Message field is updated by concatenating the fields of e and other
// separated by a semi-colon.
//
// Merge returns the updated error. This makes it possible to return other when
// err is nil.
func MergeErrors(err, other error) error {
	if err == nil {
		if other == nil {
			return nil
		}
		return other
	}
	if other == nil {
		return err
	}
	e := asError(err)
	o := asError(other)
	e.Message = e.Message + "; " + o.Message
	e.Timeout = e.Timeout && o.Timeout
	e.Temporary = e.Temporary && o.Temporary

	return e
}

// Error returns the service error message.
func (s *ServiceError) Error() string { return s.Message }

func newError(timeout, temporary bool, format string, v ...interface{}) *ServiceError {
	return &ServiceError{
		ID:        newErrorID(),
		Message:   fmt.Sprintf(format, v...),
		Timeout:   timeout,
		Temporary: temporary,
	}
}

func asError(err error) *ServiceError {
	e, ok := err.(*ServiceError)
	if !ok {
		return &ServiceError{
			ID:      newErrorID(),
			Message: err.Error(),
		}
	}
	return e
}

// If you're curious - simplifying a bit - the probability of 2 values being
// equal for n 6-bytes values is n^2 / 2^49. For n = 1 million this gives around
// 1 chance in 500. 6 bytes seems to be a good trade-off between probability of
// clashes and length of ID (6 * 4/3 = 8 chars) since clashes are not
// catastrophic.
func newErrorID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
