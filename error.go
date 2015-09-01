package goa

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	// ErrorKind is an enum listing the possible types of errors.
	ErrorKind int

	// TypedError describes an error that can be returned in a HTTP response.
	TypedError struct {
		Kind ErrorKind
		Mesg string
	}

	// MultiError records multiple errors.
	MultiError []error
)

const (
	// ErrInvalidParam is the error used when the parameter is of the wrong type
	ErrInvalidParam = iota + 1

	// ErrMissingParam is the error used when a required parameter is missing
	ErrMissingParam

	// ErrInvalidPayload is the error used when the payload is of the wrong type
	ErrInvalidPayload

	// ErrMissingPayloadField is the error used when a required payload field is missing
	ErrMissingPayloadField

	// ErrInvalidPayloadField is the error used when a payload field is of the wrong type
	ErrInvalidPayloadField

	// ErrInvalidPayloadFieldValue is the error used when a Payload field has an invalid value (not in enum)
	ErrInvalidPayloadFieldValue
)

// Title returns a human friendly error title
func (k ErrorKind) Title() string {
	switch k {
	case ErrMissingParam:
		return "missing required parameter"
	case ErrInvalidParam:
		return "invalid parameter value"
	case ErrInvalidPayload:
		return "invalid payload"
	case ErrMissingPayloadField:
		return "missing required payload field"
	case ErrInvalidPayloadField:
		return "invalid payload field"
	case ErrInvalidPayloadFieldValue:
		return "invalid payload field value"
	}
	panic("unknown kind")
}

// Error builds an error message from the typed error details.
func (t *TypedError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString(`{"kind":"`)
	buffer.WriteString(t.Kind.Title())
	buffer.WriteString(`","msg":"`)
	buffer.WriteString(t.Mesg)
	buffer.WriteString(`"}`)
	return buffer.String()
}

// Error summarizes all the underlying error messages in one JSON array.
func (m MultiError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i, err := range m {
		txt := err.Error()
		if _, ok := err.(*TypedError); !ok {
			txt = fmt.Sprintf(`"%s"`, txt)
		}
		buffer.WriteString(txt)
		if i < len(m)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	return buffer.String()
}

// InvalidParamValue coerces the given error into a MultiError then appends a typed error with
// kind ErrInvalidParam and returns the resulting MultiError.
func InvalidParamValue(name string, val interface{}, expected string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidParam,
		Mesg: fmt.Sprintf("invalid value '%v' for parameter %s, must be a %s",
			val, name, expected),
	}
	return appendError(err, &terr)
}

// MissingParam coerces the given error into a MultiError then appends a typed error with
// kind ErrMissingParam and returns the resuling MultiError.
func MissingParam(name string, err error) error {
	terr := TypedError{
		Kind: ErrMissingParam,
		Mesg: fmt.Sprintf("missing required parameter %s", name),
	}
	return appendError(err, &terr)
}

// InvalidPayload coerces the given error into a MultiError then appends a typed error with
// kind ErrInvalidPayload and returns the resulting MultiError.
func InvalidPayload(expected string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidPayload,
		Mesg: fmt.Sprintf("invalid payload, must be a %s", expected),
	}
	return appendError(err, &terr)
}

// MissingPayloadField coerces the given error into a MultiError then appends a typed error with
// kind ErrMissingPayloadField and returns the resulting MultiError.
func MissingPayloadField(name string, err error) error {
	terr := TypedError{
		Kind: ErrMissingPayloadField,
		Mesg: fmt.Sprintf("missing required payload field %s", name),
	}
	return appendError(err, &terr)
}

// InvalidPayloadField coerces the given error into a MultiError then appends a typed error with
// kind ErrInvalidPayloadField and returns the resulting MultiError.
func InvalidPayloadField(name string, val interface{}, expected string, err error) error {
	if name == "" {
		return InvalidPayload(expected, err)
	}
	terr := TypedError{
		Kind: ErrInvalidPayloadField,
		Mesg: fmt.Sprintf("invalid value '%v' for payload field %s, must be a %s",
			val, name, expected),
	}
	return appendError(err, &terr)
}

// InvalidPayloadFieldValue coerces the given error into a MultiError then appends a typed error with
// kind ErrInvalidParam and returns the resulting MultiError.
func InvalidPayloadFieldValue(name string, val interface{}, expected []string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidPayloadFieldValue,
		Mesg: fmt.Sprintf("invalid value '%v' for parameter %s, must be one of %s",
			val, name, strings.Join(expected, ", ")),
	}
	return appendError(err, &terr)
}

// IncompatibleTypeError is the error produced by the generated code when a payload type does not
// match the design definition.
func IncompatibleTypeError(ctx string, val interface{}, expected string) error {
	return fmt.Errorf("type of %s must be %s but got value %v", ctx, expected, val)
}

// appendError coerces the first argument into a MultiError then appends the second argument and
// returns the resulting MultiError.
func appendError(err error, err2 error) error {
	if err == nil {
		return MultiError{err2}
	}
	if merr, ok := err.(MultiError); ok {
		return append(merr, err2)
	}
	return MultiError{err, err2}
}
