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
// Add "flattens" validation errors so that the recorded errors are never ValidationErrors
// themselves.
func (verr *ValidationErrors) Add(def Definition, format string, vals ...interface{}) {
	err := fmt.Errorf(format, vals...)
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

// CanUse returns nil if the provider supports all the versions supported by the client or if the
// provider is unversioned.
func CanUse(client, provider Versioned) error {
	if provider.Versions() == nil {
		return nil
	}
	versions := client.Versions()
	if versions == nil {
		return fmt.Errorf("cannot use versioned %s from unversioned %s", provider.Context(),
			client.Context())
	}
	providerVersions := provider.Versions()
	if len(versions) > len(providerVersions) {
		return fmt.Errorf("cannot use %s from %s: incompatible set of supported API versions",
			provider.Context(), client.Context())
	}
	for _, v := range versions {
		found := false
		for _, pv := range providerVersions {
			if v == pv {
				found = true
			}
			break
		}
		if !found {
			return fmt.Errorf("cannot use %s from %s: incompatible set of supported API versions",
				provider.Context(), client.Context())
		}
	}
	return nil
}
