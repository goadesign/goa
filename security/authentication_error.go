// Generated service endpoints use this file to identify an authentication
// rejection before calling the business method. The original error remains
// available to transport encoders and callers through errors.Is and errors.As.
package security

type (
	// AuthenticationError means every permitted authentication alternative failed
	// and the generated endpoint did not call the business method. Authentication
	// callbacks and endpoint middleware may already have performed their own work.
	// This error does not by itself authorize a retry.
	AuthenticationError struct {
		cause error
	}
)

// NewAuthenticationError marks a non-nil authentication callback error after all
// permitted alternatives fail. Generated endpoints use it before method dispatch;
// business-method errors must not be marked with this constructor.
func NewAuthenticationError(err error) *AuthenticationError {
	return &AuthenticationError{cause: err}
}

// Error returns the original authentication callback's message.
func (e *AuthenticationError) Error() string {
	return e.cause.Error()
}

// Unwrap returns the original authentication callback error so its name, fields,
// transport mappings and cause remain available to callers and error encoders.
func (e *AuthenticationError) Unwrap() error {
	return e.cause
}
