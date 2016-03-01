package dslengine

import (
	"fmt"
	"strings"
)

// ValidationErrors records the errors encountered when running Validate.
type ValidationErrors struct {
	Errors      []error
	Definitions []Definition
}

// Error implements the error interface.
func (verr *ValidationErrors) Error() string {
	msg := make([]string, len(verr.Errors))
	for i, err := range verr.Errors {
		msg[i] = fmt.Sprintf("%s: %s", verr.Definitions[i].Context(), err)
	}
	return strings.Join(msg, "\n")
}

// Merge merges validation errors into the target.
func (verr *ValidationErrors) Merge(err *ValidationErrors) {
	if err == nil {
		return
	}
	verr.Errors = append(verr.Errors, err.Errors...)
	verr.Definitions = append(verr.Definitions, err.Definitions...)
}

// Add adds a validation error to the target.
func (verr *ValidationErrors) Add(def Definition, format string, vals ...interface{}) {
	verr.AddError(def, fmt.Errorf(format, vals...))
}

// AddError adds a validation error to the target.
// AddError "flattens" validation errors so that the recorded errors are never ValidationErrors
// themselves.
func (verr *ValidationErrors) AddError(def Definition, err error) {
	if v, ok := err.(*ValidationErrors); ok {
		verr.Errors = append(verr.Errors, v.Errors...)
		verr.Definitions = append(verr.Definitions, v.Definitions...)
		return
	}
	verr.Errors = append(verr.Errors, err)
	verr.Definitions = append(verr.Definitions, def)
}

// AsError returns an error if there are validation errors, nil otherwise.
func (verr *ValidationErrors) AsError() *ValidationErrors {
	if len(verr.Errors) > 0 {
		return verr
	}
	return nil
}
