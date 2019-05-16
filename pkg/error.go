package goa

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type (
	// ServiceError is the default error type used by the goa package to
	// encode and decode error responses.
	ServiceError struct {
		// Name is a name for that class of errors.
		Name string
		// ID is a unique value for each occurrence of the error.
		ID string
		// Message contains the specific error details.
		Message string
		// Is the error a timeout?
		Timeout bool
		// Is the error temporary?
		Temporary bool
		// Is the error a server-side fault?
		Fault bool
	}
)

// Fault creates an error given a format and values a la fmt.Printf. The error
// has the Fault field set to true.
func Fault(format string, v ...interface{}) *ServiceError {
	return newError("fault", false, false, true, format, v...)
}

// PermanentError creates an error given a name and a format and values a la
// fmt.Printf.
func PermanentError(name, format string, v ...interface{}) *ServiceError {
	return newError(name, false, false, false, format, v...)
}

// TemporaryError is an error class that indicates that the error is temporary
// and that retrying the request may be successful. TemporaryError creates an
// error given a name and a format and values a la fmt.Printf. The error has the
// Temporary field set to true.
func TemporaryError(name, format string, v ...interface{}) *ServiceError {
	return newError(name, false, true, false, format, v...)
}

// PermanentTimeoutError creates an error given a name and a format and values a
// la fmt.Printf. The error has the Timeout field set to true.
func PermanentTimeoutError(name, format string, v ...interface{}) *ServiceError {
	return newError(name, true, false, false, format, v...)
}

// TemporaryTimeoutError creates an error given a name and a format and values a
// la fmt.Printf. The error has both the Timeout and Temporary fields set to
// true.
func TemporaryTimeoutError(name, format string, v ...interface{}) *ServiceError {
	return newError(name, true, true, false, format, v...)
}

// MissingPayloadError is the error produced by the generated code when a
// request is missing a required payload.
func MissingPayloadError() error {
	return PermanentError("missing_payload", "missing required payload")
}

// DecodePayloadError is the error produced by the generated code when a request
// body cannot be decoded successfully.
func DecodePayloadError(msg string) error {
	return PermanentError("decode_payload", msg)
}

// InvalidFieldTypeError is the error produced by the generated code when the
// type of a payload field does not match the type defined in the design.
func InvalidFieldTypeError(name string, val interface{}, expected string) error {
	return PermanentError("invalid_field_type", "invalid value %#v for %q, must be a %s", val, name, expected)
}

// MissingFieldError is the error produced by the generated code when a payload
// is missing a required field.
func MissingFieldError(name, context string) error {
	return PermanentError("missing_field", "%q is missing from %s", name, context)
}

// InvalidEnumValueError is the error produced by the generated code when the
// value of a payload field does not match one the values defined in the design
// Enum validation.
func InvalidEnumValueError(name string, val interface{}, allowed []interface{}) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	return PermanentError("invalid_enum_value", "value of %s must be one of %s but got value %#v", name, strings.Join(elems, ", "), val)
}

// InvalidFormatError is the error produced by the generated code when the value
// of a payload field does not match the format validation defined in the
// design.
func InvalidFormatError(name, target string, format Format, formatError error) error {
	return PermanentError("invalid_format", "%s must be formatted as a %s but got value %q, %s", name, format, target, formatError.Error())
}

// InvalidPatternError is the error produced by the generated code when the
// value of a payload field does not match the pattern validation defined in the
// design.
func InvalidPatternError(name, target string, pattern string) error {
	return PermanentError("invalid_pattern", "%s must match the regexp %q but got value %q", name, pattern, target)
}

// InvalidRangeError is the error produced by the generated code when the value
// of a payload field does not match the range validation defined in the design.
// value may be an int or a float64.
func InvalidRangeError(name string, target interface{}, value interface{}, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return PermanentError("invalid_range", "%s must be %s than %d but got value %#v", name, comp, value, target)
}

// InvalidLengthError is the error produced by the generated code when the value
// of a payload field does not match the length validation defined in the
// design.
func InvalidLengthError(name string, target interface{}, ln, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return PermanentError("invalid_length", "length of %s must be %s than %d but got value %#v (len=%d)", name, comp, value, target, ln)
}

// NewErrorID creates a unique 8 character ID that is well suited to use as an
// error identifier.
func NewErrorID() string {
	// for the curious - simplifying a bit - the probability of 2 values
	// being equal for n 6-bytes values is n^2 / 2^49. For n = 1 million
	// this gives around 1 chance in 500. 6 bytes seems to be a good
	// trade-off between probability of clashes and length of ID (6 * 4/3 =
	// 8 chars) since clashes are not catastrophic.
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// MergeErrors updates an error by merging another into it. It first converts
// other into a ServiceError if not already one. The merge algorithm then:
//
// * uses the name of err if a ServiceError, the name of other otherwise.
//
// * appends both error messages.
//
// * computes Timeout and Temporary by "and"ing the fields of both errors.
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
	if e.Name == "error" {
		e.Name = o.Name
	}
	e.Message = e.Message + "; " + o.Message
	e.Timeout = e.Timeout && o.Timeout
	e.Temporary = e.Temporary && o.Temporary
	e.Fault = e.Fault && o.Fault

	return e
}

// Error returns the error message.
func (s *ServiceError) Error() string { return s.Message }

// ErrorName returns the error name.
func (s *ServiceError) ErrorName() string { return s.Name }

func newError(name string, timeout, temporary, fault bool, format string, v ...interface{}) *ServiceError {
	return &ServiceError{
		Name:      name,
		ID:        NewErrorID(),
		Message:   fmt.Sprintf(format, v...),
		Timeout:   timeout,
		Temporary: temporary,
		Fault:     fault,
	}
}

func asError(err error) *ServiceError {
	e, ok := err.(*ServiceError)
	if !ok {
		return &ServiceError{
			Name:    "error",
			ID:      NewErrorID(),
			Message: err.Error(),
			Fault:   true, // Default to fault for unexpected errors
		}
	}
	return e
}
