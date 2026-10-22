// These tests forward existing gRPC responses through the server encoder.
// Their status and ordered details must reach callers without invented fields.
package grpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	gengoapb "goa.design/goa/v3/grpc/pb"
)

// TestEncodeErrorForwardedStatus preserves the peer's complete error response
// for native errors, ordinary wrappers, and joins with one effective cause.
func TestEncodeErrorForwardedStatus(t *testing.T) {
	for code := codes.Canceled; code <= codes.Unauthenticated; code++ {
		for _, withDetails := range []bool{false, true} {
			original := status.New(code, "downstream operation failed")
			if withDetails {
				var err error
				original, err = original.WithDetails(wrapperspb.String("first detail"), &gengoapb.ErrorResponse{
					Name: "rejected", Id: "downstream-error", Msg: "declared downstream rejection",
				})
				require.NoError(t, err)
			}
			for _, test := range []struct {
				name string
				err  error
			}{
				{"raw", original.Err()},
				{"wrapped", fmt.Errorf("read downstream: %w", original.Err())},
				{"single join", errors.Join(original.Err())},
			} {
				t.Run(fmt.Sprintf("%s/details=%t/%s", code, withDetails, test.name), func(t *testing.T) {
					expected, ok := status.FromError(test.err)
					require.True(t, ok)
					encoded := status.Convert(EncodeError(test.err))
					assert.True(t, proto.Equal(expected.Proto(), encoded.Proto()), "the complete existing response must be preserved")
				})
			}
		}
	}
}
