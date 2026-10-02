// This file constructs runtime and validation errors for generated code and
// service implementations. Merges return a new whole error and keep original
// contribution values; callers can inspect the existing causes through Unwrap.
package goa

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

type (
	// ServiceError is the default error type used by the goa package to
	// encode and decode error responses. Finish setting its public fields and
	// constructing its causes before sharing it with readers. MergeErrors does
	// not modify its inputs, but direct field writes and external cause behavior
	// remain the caller's responsibility.
	ServiceError struct {
		// Name is a name for that class of errors.
		Name string
		// ID is a unique value for each occurrence of the error.
		ID string
		// Pointer to the field that caused this error, if appropriate
		Field *string
		// Message contains the specific error details.
		Message string
		// Is the error a timeout?
		Timeout bool
		// Is the error temporary?
		Temporary bool
		// Is the error a server-side fault?
		Fault bool
		// history holds original contribution values in merge order. Each value
		// has its own Field string and no nested history, so History can return
		// editable copies without exposing these saved values.
		history []ServiceError
		// err holds the original error if exists.
		err error
	}

	// GoaErrorNamer is an interface implemented by generated error structs that
	// exposes the name of the error as defined in the design.
	GoaErrorNamer interface {
		GoaErrorName() string
	}
)

const (
	// InvalidFieldType is the error name for invalid field type errors.
	InvalidFieldType = "invalid_field_type"
	// MissingField is the error name for missing field errors.
	MissingField = "missing_field"
	// InvalidEnumValue is the error name for invalid enum value errors.
	InvalidEnumValue = "invalid_enum_value"
	// InvalidFormat is the error name for invalid format errors.
	InvalidFormat = "invalid_format"
	// InvalidPattern is the error name for invalid pattern errors.
	InvalidPattern = "invalid_pattern"
	// InvalidRange is the error name for invalid range errors.
	InvalidRange = "invalid_range"
	// InvalidLength is the error name for invalid length errors.
	InvalidLength = "invalid_length"
	// UnsupportedMediaType is the error name returned by the Goa decoder
	// when the content type of the HTTP request body is not supported.
	UnsupportedMediaType = "unsupported_media_type"
	// DecodePayload is the error name for decode payload errors.
	DecodePayload = "decode_payload"
	// MissingPayload is the error name for missing payload errors.
	MissingPayload = "missing_payload"
)

// NewServiceError creates an error.
func NewServiceError(err error, name string, timeout, temporary, fault bool) *ServiceError {
	return &ServiceError{
		Name:      name,
		ID:        NewErrorID(),
		Message:   err.Error(),
		Timeout:   timeout,
		Temporary: temporary,
		Fault:     fault,
		err:       err,
	}
}

// Fault creates an error given a format and values a la fmt.Printf. The error
// has the Fault field set to true.
func Fault(format string, v ...any) *ServiceError {
	return newError("fault", false, false, true, format, v...)
}

// PermanentError creates an error given a name and a format and values a la
// fmt.Printf.
func PermanentError(name, format string, v ...any) *ServiceError {
	return newError(name, false, false, false, format, v...)
}

// TemporaryError is an error class that indicates that the error is temporary
// and that retrying the request may be successful. TemporaryError creates an
// error given a name and a format and values a la fmt.Printf. The error has the
// Temporary field set to true.
func TemporaryError(name, format string, v ...any) *ServiceError {
	return newError(name, false, true, false, format, v...)
}

// PermanentTimeoutError creates an error given a name and a format and values a
// la fmt.Printf. The error has the Timeout field set to true.
func PermanentTimeoutError(name, format string, v ...any) *ServiceError {
	return newError(name, true, false, false, format, v...)
}

// TemporaryTimeoutError creates an error given a name and a format and values a
// la fmt.Printf. The error has both the Timeout and Temporary fields set to
// true.
func TemporaryTimeoutError(name, format string, v ...any) *ServiceError {
	return newError(name, true, true, false, format, v...)
}

// MissingPayloadError is the error produced by the generated code when a
// request is missing a required payload.
func MissingPayloadError() error {
	return PermanentError(MissingPayload, "missing required payload")
}

// DecodePayloadError is the error produced by the generated code when a request
// body cannot be decoded successfully.
func DecodePayloadError(msg string) error {
	return PermanentError(DecodePayload, "%s", msg)
}

// UnsupportedMediaTypeError is the error produced by the Goa decoder when the
// content type of the HTTP request body is not supported.
func UnsupportedMediaTypeError(ct string) error {
	return PermanentError(UnsupportedMediaType, "unsupported media type %s", ct)
}

// InvalidFieldTypeError is the error produced by the generated code when the
// type of a payload field does not match the type defined in the design.
func InvalidFieldTypeError(name string, val any, expected string) error {
	return withField(name, PermanentError(
		InvalidFieldType, "invalid value %#v for %q, must be a %s", val, name, expected))
}

// MissingFieldError is the error produced by the generated code when a payload
// is missing a required field.
func MissingFieldError(name, context string) error {
	return withField(name, PermanentError(
		MissingField, "%q is missing from %s", name, context))
}

// InvalidEnumValueError is the error produced by the generated code when the
// value of a payload field does not match one the values defined in the design
// Enum validation.
func InvalidEnumValueError(name string, val any, allowed []any) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	return withField(name, PermanentError(
		InvalidEnumValue, "value of %s must be one of %s but got value %#v", name, strings.Join(elems, ", "), val))
}

// InvalidFormatError is the error produced by the generated code when the value
// of a payload field does not match the format validation defined in the
// design.
func InvalidFormatError(name, target string, format Format, formatError error) error {
	return withField(name, PermanentError(
		InvalidFormat, "%s must be formatted as a %s but got value %q, %s", name, format, target, formatError.Error()))
}

// InvalidPatternError is the error produced by the generated code when the
// value of a payload field does not match the pattern validation defined in the
// design.
func InvalidPatternError(name, target, pattern string) error {
	return withField(name, PermanentError(
		InvalidPattern, "%s must match the regexp %q but got value %q", name, pattern, target))
}

// InvalidRangeError is the error produced by the generated code when the value
// of a payload field does not match the range validation defined in the design.
// value may be an int or a float64.
func InvalidRangeError(name string, target, value any, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return withField(name, PermanentError(
		InvalidRange, "%s must be %s than %d but got value %#v", name, comp, value, target))
}

// InvalidLengthError is the error produced by the generated code when the value
// of a payload field does not match the length validation defined in the
// design.
func InvalidLengthError(name string, _ any, ln, value int, min bool) error {
	comp := "at least"
	if !min {
		comp = "at most"
	}
	return withField(name, PermanentError(
		InvalidLength, "length of %s must be %s %d but got %d", name, comp, value, ln))
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
	io.ReadFull(rand.Reader, b) // nolint: errcheck
	return base64.RawURLEncoding.EncodeToString(b)
}

// MergeErrors returns a fresh *ServiceError for two nonnil inputs without
// modifying either input. Callers must use the returned error. Each input is
// selected with errors.As; an input with no *ServiceError becomes an error-named
// fault containing its original text and cause.
//
// The result keeps the selected left error's ID and Field pointer. It keeps the
// left Name unless that name is "error", in which case it uses the right Name.
// Its Message joins the selected messages with "; ". Timeout, Temporary and
// Fault are each true only when both selected errors have that trait.
//
// The result joins the existing causes in left-to-right order and retains the
// original contributions for History, including duplicates. No cause is copied
// or modified. Stable finite input cause graphs remain finite because the new
// result cannot be reached from an older input. This does not repair existing
// cycles or control methods and mutations on caller-owned causes.
//
// If either input is nil, MergeErrors returns the other input exactly; if both
// are nil, it returns nil.
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

	// Save both inputs' contributions before combining whole fields. A new
	// slice keeps subsequent merges from writing into either input's history.
	left := errorContributions(e)
	right := errorContributions(o)
	merged := *e
	merged.history = make([]ServiceError, 0, len(left)+len(right))
	merged.history = append(merged.history, left...)
	merged.history = append(merged.history, right...)
	if merged.Name == "error" {
		merged.Name = o.Name
	}
	merged.Message = e.Message + "; " + o.Message
	merged.Timeout = e.Timeout && o.Timeout
	merged.Temporary = e.Temporary && o.Temporary
	merged.Fault = e.Fault && o.Fault
	merged.err = errors.Join(e.err, o.err)

	return &merged
}

// History returns detached copies of the original, unmerged contributions in
// left-to-right merge order, preserving duplicates. Each entry retains its
// original Name, ID, Message, traits and cause, has a copied Field value, and
// contains no nested history. Changing the returned slice, entries or Field
// strings does not change saved contributions or the whole error.
//
// For an unmerged error, History returns a detached singleton with the error's
// current fields. A merged error retains contributions as they were included;
// later input or whole-field edits do not rewrite them. Causes remain the exact
// original objects, with behavior owned by their callers.
func (e *ServiceError) History() []*ServiceError {
	contributions := errorContributions(e)
	history := make([]*ServiceError, len(contributions))
	for i, original := range contributions {
		entry := copyErrorContribution(original)
		history[i] = &entry
	}
	return history
}

// Error returns the error message.
func (e *ServiceError) Error() string { return e.Message }

// ErrorName returns the error name.
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e *ServiceError) ErrorName() string { return e.Name }

// GoaErrorName returns the error name.
func (e *ServiceError) GoaErrorName() string { return e.ErrorName() }

// Unwrap returns the existing cause, or the ordered joined causes of a merge.
func (e *ServiceError) Unwrap() error { return e.err }

// copyErrorContribution copies an error's fields and Field string, removes its
// nested history, and retains its exact cause so the copy describes one original
// contribution without exposing an owned Field value.
func copyErrorContribution(original ServiceError) ServiceError {
	original.history = nil
	if original.Field != nil {
		field := *original.Field
		original.Field = &field
	}
	return original
}

// errorContributions supplies the saved original values for a merged error or
// captures an unmerged error's current fields. Its callers read this list and
// allocate their own slices before storing or returning entries.
func errorContributions(e *ServiceError) []ServiceError {
	if len(e.history) > 0 {
		return e.history
	}
	return []ServiceError{copyErrorContribution(*e)}
}

func withField(field string, err *ServiceError) *ServiceError {
	err.Field = &field
	return err
}

func newError(name string, timeout, temporary, fault bool, format string, v ...any) *ServiceError {
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
	var e *ServiceError
	if !errors.As(err, &e) {
		return &ServiceError{
			Name:    "error",
			ID:      NewErrorID(),
			Message: err.Error(),
			Fault:   true, // Default to fault for unexpected errors
			err:     err,
		}
	}
	return e
}
