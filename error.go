/*
Package goa standardizes on structured error responses: a request that fails
because of an invalid input or an unexpected condition produces a response that
contains a structured error.

By default, the error data structures returned to clients contains three fields:
an id, a status and a message.

* The id is unique for each occurrence of the error, it helps correlate the
  content of the response with the content of the service logs for example.

* The status carries the error handling semantic, whether the error can be
  retried for example.

* The message contains is specific to the error occurrence and is intended for
  human consumption.

Instances of Error can be created via error class functions. New error classes
can be created with NewErrorClass.

All instance of errors created via a error class implement the Error interface.
This interface is leveraged by the error handler middleware to produce the error
responses. The middleware takes care of mapping back any error returned by
previously called middleware or action handler into transport specific
responses.
*/
package goa

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const (
	// StatusInvalid identifies badly formed requests.
	StatusInvalid ErrorStatus = iota + 1
	// StatusUnauthorized indicates the request is not authorized.
	StatusUnauthorized
	// StatusTimeout identifies a request that timed out.
	StatusTimeout
	// StatusBug indicates an uncaught error.
	StatusBug
)

var (
	// ErrInvalid is the error class used to build responses for requests
	// that fail validations.
	ErrInvalid = NewErrorClass(StatusInvalid)
	// ErrUnauthorized is a generic unauthorized error.
	ErrUnauthorized = NewErrorClass(StatusUnauthorized)
	// ErrTimeout is the error class used to build responses for requests
	// that timed out.
	ErrTimeout = NewErrorClass(StatusTimeout)
	// ErrBug is the class of error used for uncaught errors.
	ErrBug = NewErrorClass(StatusBug)
)

type (
	// ErrorStatus defines the semantic attached to the error to inform
	// error handlers. For example clients may use the status to determine
	// whether a request should be retried.
	ErrorStatus int

	// ErrorClass is an error generating function.
	// It accepts a format and values a la fmt.Fprintf.
	ErrorClass func(format string, v ...interface{}) error

	// Error is the interface implemented by all errors created using a
	// ErrorClass function.
	Error interface {
		// Error extends the error interface
		error
		// ID is a unique value associated with the occurrence of the
		// error.
		ID() string
		// Status of error, informs clients on how to perform handling.
		Status() ErrorStatus
		// Message contains the occurrence specific error details.
		Message() string
	}

	// serviceError is the error type used by the goa package.
	serviceError struct {
		id      string
		status  ErrorStatus
		message string
	}
)

// NewErrorClass creates a error class with the given status.
func NewErrorClass(status ErrorStatus) ErrorClass {
	return func(format string, v ...interface{}) error {
		return &serviceError{
			id:      newErrorID(),
			status:  status,
			message: fmt.Sprintf(format, v...),
		}
	}
}

// MissingPayloadError is the error produced when a request is missing a
// required payload.
func MissingPayloadError() error {
	return ErrInvalid("missing required payload")
}

// InvalidFieldTypeError is the error produced when the type of a payload field
// does not match the type defined in the design.
func InvalidFieldTypeError(name string, val interface{}, expected string) error {
	return ErrInvalid("invalid value %#v for %#v, must be a %s", val, name, expected)
}

// MissingFieldError is the error produced when a payload is missing a required
// field.
func MissingFieldError(name, context string) error {
	return ErrInvalid("%#v is missing from %s", name, context)
}

// InvalidEnumValueError is the error produced when the value of a payload field
// does not match one the values defined in the design Enum validation.
func InvalidEnumValueError(name string, val interface{}, allowed []interface{}) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	return ErrInvalid("value of %s must be one of %s but got value %#v", name, strings.Join(elems, ", "), val)
}

// InvalidFormatError is the error produced when the value of a payload field
// does not match the format validation defined in the design.
func InvalidFormatError(name, target string, format Format, formatError error) error {
	return ErrInvalid("%s must be formatted as a %s but got value %#v, %s", name, format, target, formatError.Error())
}

// InvalidPatternError is the error produced when the value of a payload field
// does not match the pattern validation defined in the design.
func InvalidPatternError(name, target string, pattern string) error {
	return ErrInvalid("%s must match the regexp %#v but got value %#v", name, pattern, target)
}

// InvalidRangeError is the error produced when the value of a payload field does
// not match the range validation defined in the design. value may be an int or
// a float64.
func InvalidRangeError(name string, target interface{}, value interface{}, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return ErrInvalid("%s must be %s than %d but got value %#v", name, comp, value, target)
}

// InvalidLengthError is the error produced when the value of a payload field
// does not match the length validation defined in the design.
func InvalidLengthError(name string, target interface{}, ln, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return ErrInvalid("length of %s must be %s than %d but got value %#v (len=%d)", name, comp, value, target, ln)
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
	e := asError(err).(*serviceError)
	o := asError(other)
	switch {
	case e.status == StatusBug || o.Status() == StatusBug:
		if e.status != StatusBug {
			e.status = StatusBug
		}
	case e.status != o.Status():
		e.status = StatusInvalid
	}
	e.message = e.message + "; " + o.Message()

	return e
}

func asError(err error) Error {
	e, ok := err.(*serviceError)
	if !ok {
		return &serviceError{
			id:      newErrorID(),
			status:  StatusBug,
			message: err.Error(),
		}
	}
	return e
}

// Error returns the error occurrence details.
func (e *serviceError) Error() string {
	return fmt.Sprintf("[%s] %d: %s", e.id, e.status, e.message)
}

func (e *serviceError) Status() ErrorStatus { return e.status }
func (e *serviceError) ID() string          { return e.id }
func (e *serviceError) Message() string     { return e.message }

// If you're curious - simplifying a bit - the probability of 2 values being
// equal for n 6-bytes values is n^2 / 2^49. For n = 1 million this gives around
// 1 chance in 500. 6 bytes seems to be a good trade-off between probability of
// clashes and length of ID (6 * 4/3 = 8 chars) since clashes are not
// catastrophic.
func newErrorID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}
