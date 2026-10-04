// These tests inspect authentication rejections through the public error API.
// Wrapping adds method-dispatch information while preserving the callback's
// original cause, named error and native transport status.
package security_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	goagrpc "goa.design/goa/v3/grpc"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

type (
	callbackError struct{}
)

func (callbackError) Error() string {
	return "credential rejected"
}

func (callbackError) GoaErrorName() string {
	return "denied"
}

func TestAuthenticationErrorPreservesCause(t *testing.T) {
	for _, cause := range []error{
		errors.New("credential rejected"),
		callbackError{},
		goa.PermanentError("denied", "credential rejected"),
		goa.Fault("authentication service failed"),
		context.Canceled,
		context.DeadlineExceeded,
		status.Error(codes.PermissionDenied, "credential rejected"),
		errors.Join(errors.New("one rejected"), errors.New("another rejected")),
	} {
		t.Run(fmt.Sprintf("%T/%s", cause, cause.Error()), func(t *testing.T) {
			rejection := security.NewAuthenticationError(cause)
			require.Equal(t, cause, errors.Unwrap(rejection))
			require.ErrorIs(t, rejection, cause)
			require.Equal(t, cause.Error(), rejection.Error())
			require.Equal(t, status.Code(goagrpc.EncodeError(cause)), status.Code(goagrpc.EncodeError(rejection)))
			require.Equal(t, goahttp.NewErrorResponse(t.Context(), cause).StatusCode(), goahttp.NewErrorResponse(t.Context(), rejection).StatusCode())
			var originalName, wrappedName goa.GoaErrorNamer
			require.Equal(t, errors.As(cause, &originalName), errors.As(rejection, &wrappedName))
			if originalName != nil {
				require.Equal(t, originalName.GoaErrorName(), wrappedName.GoaErrorName())
			}
			var outer *security.AuthenticationError
			require.ErrorAs(t, fmt.Errorf("endpoint middleware: %w", rejection), &outer)
			require.Same(t, rejection, outer)
		})
	}
}
