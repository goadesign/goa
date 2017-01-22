/*
Package goa standardizes on structured error responses: a request that fails
because of an invalid input or an unexpected condition produces a response that
contains a structured error.

By default, the error data structures returned to clients contains five fields:
a token, a code, a status, a detail and additional data.

* The token is unique for each occurrence of the error, it helps correlate the
  content of the response with the content of the service logs for example.

* The status carries the error handling semantic: whether the error can be
  retried, whether it should bubble up to the transport layer etc. Status maps
  to HTTP status codes when the underlying transport is HTTP.

* The code defines the class of error (e.g. "invalid_parameter_type"). This is
  meant to help clients easily (programmatically) identify classes of errors.

* The detail contains a message specific to the error occurrence intended for
  human consumption.

* The metadata contains key/value pairs that provide contextual information such
  as the name of parameters, the value of an invalid parameter etc.

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
	// CodeInvalid is the code of the error class used to build responses
	// for invalid requests.
	CodeInvalid = "invalid_request"
	// CodeTimeout is the code of the error class used to build responses for
	// requests that timed out.
	CodeTimeout = "timeout"
	// CodeBug is the code of the error class used to build responses for
	// uncaught errors.
	CodeBug = "internal_error"
)

const (
	// StatusInvalid identifies badly formed requests.
	StatusInvalid ErrorStatus = iota + 1
	// StatusTimeout identifies a request that timed out.
	StatusTimeout
	// StatusBug indicates an uncaught error.
	StatusBug
)

var (
	// ErrInvalid is the error class used to build responses for requests
	// that fail validations.
	ErrInvalid = NewErrorClass(CodeInvalid, StatusInvalid)
	// ErrTimeout is the error class used to build responses for requests
	// that timed out.
	ErrTimeout = NewErrorClass(CodeTimeout, StatusTimeout)
)

type (
	// ErrorStatus defines the semantic attached to the error to inform
	// error handlers. For example clients may use the status to determine
	// whether a request should be retried or the library code may use it to
	// determine whether the error is a transport layer error.
	ErrorStatus int

	// ErrorClass is an error generating function.
	//
	// It accepts a message and optional key value pairs and produces errors
	// that implement Error:
	//
	// - If the message is a string or a fmt.Stringer then the string value
	//   is used to set the message.
	//
	// - If the message is an error then the string returned by Error() is
	//   used to set the message.
	//
	// - Otherwise the string produced using fmt.Sprintf("%v") is used.
	//
	// The optional key value pairs are intended to provide additional
	// contextual information and are returned to the client unchanged.
	ErrorClass func(message interface{}, keyvals ...interface{}) error

	// Error is the interface implemented by all errors created using a
	// ErrorClass function. It is an interface so that client library
	// packages can define their own compatible interface to handle errors
	// created by goa. This makes it possible to write library packages that
	// consume goa errors but don't force their own clients to import the goa
	// package.
	Error interface {
		// Error extends the error interface
		error
		// Status of error, informs whether error should bubble up to
		// transport.
		Status() ErrorStatus
		// Code is the code of the error class used to create this error.
		Code() string
		// Token is a unique value associated with the occurrence of the
		// error.
		Token() string
		// Detail contains the occurrence specific error message.
		Detail() string
		// Data returns additional key/value pairs that may be useful to
		// clients for error handling. It uses a slice to guarantee
		// ordering.
		Data() []map[string]interface{}
	}

	// serviceError is the error type used by the goa package.
	serviceError struct {
		status ErrorStatus
		token  string
		code   string
		detail string
		data   []map[string]interface{}
	}
)

// NewErrorClass creates a error class with the given code and status.
//
// The code identifies the newly created error class and must be unique in a
// given service process. The status indicates the semantic of the error to all
// error handlers: whether the error is retryable, whether it should bubble up
// to the transport layer and trigger cicuit breaker etc.
//
// It is the responsibility of the client to guarantee uniqueness of code.
func NewErrorClass(code string, status ErrorStatus) ErrorClass {
	return func(message interface{}, keyvals ...interface{}) error {
		var msg string
		switch actual := message.(type) {
		case string:
			msg = actual
		case error:
			msg = actual.Error()
		case fmt.Stringer:
			msg = actual.String()
		default:
			msg = fmt.Sprintf("%v", actual)
		}
		data := make([]map[string]interface{}, (len(keyvals)+1)/2)
		for i := 0; i < len(keyvals); i += 2 {
			k := keyvals[i]
			var v interface{} = "MISSING"
			if i+1 < len(keyvals) {
				v = keyvals[i+1]
			}
			data[i/2] = map[string]interface{}{fmt.Sprintf("%v", k): v}
		}
		return &serviceError{
			token:  newErrorToken(),
			status: status,
			code:   code,
			detail: msg,
			data:   data,
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
	msg := fmt.Sprintf("invalid value %#v for %#v, must be a %s", val, name, expected)
	return ErrInvalid(msg, "field", name, "value", val, "expected", expected)
}

// MissingFieldError is the error produced when a payload is missing a required
// field.
func MissingFieldError(name string) error {
	msg := fmt.Sprintf("%#v is missing", name)
	return ErrInvalid(msg, "field", name)
}

// InvalidEnumValueError is the error produced when the value of a payload field
// does not match one the values defined in the design Enum validation.
func InvalidEnumValueError(ctx string, val interface{}, allowed []interface{}) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	msg := fmt.Sprintf("value of %s must be one of %s but got value %#v", ctx, strings.Join(elems, ", "), val)
	return ErrInvalid(msg, "field", ctx, "value", val, "expected", strings.Join(elems, ", "))
}

// InvalidFormatError is the error produced when the value of a payload field
// does not match the format validation defined in the design.
func InvalidFormatError(ctx, target string, format Format, formatError error) error {
	msg := fmt.Sprintf("%s must be formatted as a %s but got value %#v, %s", ctx, format, target, formatError.Error())
	return ErrInvalid(msg, "field", ctx, "value", target, "expected", format, "error", formatError.Error())
}

// InvalidPatternError is the error produced when the value of a payload field
// does not match the pattern validation defined in the design.
func InvalidPatternError(ctx, target string, pattern string) error {
	msg := fmt.Sprintf("%s must match the regexp %#v but got value %#v", ctx, pattern, target)
	return ErrInvalid(msg, "field", ctx, "value", target, "regexp", pattern)
}

// InvalidRangeError is the error produced when the value of a payload field does
// not match the range validation defined in the design. value may be an int or
// a float64.
func InvalidRangeError(ctx string, target interface{}, value interface{}, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	msg := fmt.Sprintf("%s must be %s than %d but got value %#v", ctx, comp, value, target)
	return ErrInvalid(msg, "field", ctx, "value", target, "comp", comp, "expected", value)
}

// InvalidLengthError is the error produced when the value of a payload field
// does not match the length validation defined in the design.
func InvalidLengthError(ctx string, target interface{}, ln, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	msg := fmt.Sprintf("length of %s must be %s than %d but got value %#v (len=%d)", ctx, comp, value, target, ln)
	return ErrInvalid(msg, "field", ctx, "value", target, "len", ln, "comp", comp, "expected", value)
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
// The Detail field is updated by concatenating the Detail fields of e and other
// separated by a semi-colon. The MetaValues field of is updated by merging the
// map of other MetaValues into e's where values in e with identical keys to
// values in other get overwritten.
//
// Merge returns the updated error. This is makes it possible to return other
// when err is nil.
func MergeErrors(err, other error) error {
	if err == nil {
		if other == nil {
			return nil
		}
		return asError(other)
	}
	if other == nil {
		return asError(err)
	}
	e := asError(err).(*serviceError)
	o := asError(other)
	switch {
	case e.status == StatusBug || o.Status() == StatusBug:
		if e.status != StatusBug {
			e.status = StatusBug
			e.code = CodeBug
		}
	case e.status != o.Status() || e.code != o.Code():
		e.status = StatusInvalid
		e.code = CodeInvalid
	}
	e.detail = e.detail + "; " + o.Detail()

	for _, val := range o.Data() {
		for k, v := range val {
			e.data = append(e.data, map[string]interface{}{k: v})
		}
	}
	return e
}

// Error returns the error occurrence details.
func (e *serviceError) Error() string {
	msg := fmt.Sprintf("[%s] %d %s: %s", e.token, e.status, e.code, e.detail)
	for _, val := range e.data {
		for k, v := range val {
			msg += ", " + fmt.Sprintf("%s: %v", k, v)
		}
	}
	return msg
}

func (e *serviceError) Status() ErrorStatus            { return e.status }
func (e *serviceError) Token() string                  { return e.token }
func (e *serviceError) Code() string                   { return e.code }
func (e *serviceError) Detail() string                 { return e.detail }
func (e *serviceError) Data() []map[string]interface{} { return e.data }

func asError(err error) Error {
	e, ok := err.(*serviceError)
	if !ok {
		return &serviceError{
			status: StatusBug,
			code:   CodeBug,
			token:  newErrorToken(),
			detail: err.Error(),
		}
	}
	return e
}

// If you're curious - simplifying a bit - the probability of 2 values being
// equal for n 6-bytes values is n^2 / 2^49. For n = 1 million this gives around
// 1 chance in 500. 6 bytes seems to be a good trade-off between probability of
// clashes and length of ID (6 * 4/3 = 8 chars) since clashes are not
// catastrophic.
func newErrorToken() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}
