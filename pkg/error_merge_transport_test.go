// These tests use the public runtime and existing transport helpers to verify
// whole response fields, original issue messages, and retained concrete causes.
// The decoders keep their existing loss of local contribution facts.
package goa_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	goagrpc "goa.design/goa/v3/grpc"
	generrorpb "goa.design/goa/v3/grpc/pb"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type (
	// mergeTransportCause retains a status object so custom As inspection can
	// find the exact transport error behind ContextError after a merge.
	mergeTransportCause struct {
		transportStatus *status.Status
	}
)

func TestMergeErrorsTransportWholeFields(t *testing.T) {
	for leftTraits := 0; leftTraits < 8; leftTraits++ {
		for rightTraits := 0; rightTraits < 8; rightTraits++ {
			t.Run(fmt.Sprintf("%d/%d", leftTraits, rightTraits), func(t *testing.T) {
				left := &goa.ServiceError{
					Name: "left", ID: "left-id", Message: "left message",
					Timeout: leftTraits&1 != 0, Temporary: leftTraits&2 != 0, Fault: leftTraits&4 != 0,
				}
				right := &goa.ServiceError{
					Name: "right", ID: "right-id", Message: "right message",
					Timeout: rightTraits&1 != 0, Temporary: rightTraits&2 != 0, Fault: rightTraits&4 != 0,
				}
				merged := goa.MergeErrors(left, right)
				whole := &goa.ServiceError{
					Name: left.Name, ID: left.ID, Message: "left message; right message",
					Timeout: left.Timeout && right.Timeout, Temporary: left.Temporary && right.Temporary,
					Fault: left.Fault && right.Fault,
				}
				expectedHTTP := goahttp.NewErrorResponse(context.Background(), whole)
				actualHTTP := goahttp.NewErrorResponse(context.Background(), merged)
				require.Equal(t, expectedHTTP, actualHTTP)
				require.Equal(t, expectedHTTP.StatusCode(), actualHTTP.StatusCode())
				expectedBody, actualBody := httptest.NewRecorder(), httptest.NewRecorder()
				ctx := context.Background()
				require.NoError(t, goahttp.ResponseEncoder(ctx, expectedBody).Encode(expectedHTTP))
				require.NoError(t, goahttp.ResponseEncoder(ctx, actualBody).Encode(actualHTTP))
				require.Equal(t, expectedBody.Body.Bytes(), actualBody.Body.Bytes())
				require.Equal(t, expectedBody.Header(), actualBody.Header())

				expectedGRPC := goagrpc.NewErrorResponse(whole)
				expectedGRPC.History = []*generrorpb.ErrorField{
					{Name: left.Name, Msg: left.Message},
					{Name: right.Name, Msg: right.Message},
				}
				actualGRPC := goagrpc.NewErrorResponse(merged)
				require.True(t, proto.Equal(expectedGRPC, actualGRPC))
				encoded := goagrpc.EncodeError(merged)
				require.Equal(t, status.Code(goagrpc.EncodeError(whole)), status.Code(encoded))
				require.Equal(t, whole.Message, status.Convert(encoded).Message())
				require.True(t, proto.Equal(expectedGRPC, goagrpc.DecodeError(encoded)))
			})
		}
	}
}

func TestMergeErrorsTransportOriginalFieldIssues(t *testing.T) {
	left := goa.MissingFieldError("title", "body")
	right := goa.InvalidLengthError("description", nil, 0, 1, true)
	merged := goa.MergeErrors(goa.MergeErrors(left, right), left)
	var whole *goa.ServiceError
	require.ErrorAs(t, merged, &whole)
	history := whole.History()
	require.Len(t, history, 3)
	require.Equal(t, history[0].ID, history[2].ID)
	require.NotEqual(t, history[0].ID, history[1].ID)

	response := goagrpc.NewErrorResponse(merged)
	expected := &generrorpb.ErrorResponse{
		Name: goa.MissingField, Id: history[0].ID,
		Msg: left.Error() + "; " + right.Error() + "; " + left.Error(),
		History: []*generrorpb.ErrorField{
			{Name: goa.MissingField, Field: "title", Msg: left.Error()},
			{Name: goa.InvalidLength, Field: "description", Msg: right.Error()},
			{Name: goa.MissingField, Field: "title", Msg: left.Error()},
		},
	}
	require.True(t, proto.Equal(expected, response))
	encoded := goagrpc.EncodeError(merged)
	require.Equal(t, codes.InvalidArgument, status.Code(encoded))
	require.True(t, proto.Equal(expected, goagrpc.DecodeError(encoded)))

	decoded := goagrpc.NewServiceError(response)
	decodedWithCause := goagrpc.NewServiceErrorWithCause(encoded, response)
	for _, value := range []*goa.ServiceError{decoded, decodedWithCause} {
		require.Equal(t, whole.Name, value.Name)
		require.Equal(t, whole.ID, value.ID)
		require.Equal(t, whole.Message, value.Message)
		require.Equal(t, whole.Timeout, value.Timeout)
		require.Equal(t, whole.Temporary, value.Temporary)
		require.Equal(t, whole.Fault, value.Fault)
		require.Nil(t, value.Field)
		require.Len(t, value.History(), 1)
		require.Equal(t, whole.Message, value.History()[0].Message)
	}
	require.Nil(t, decoded.Unwrap())
	require.Same(t, encoded, decodedWithCause.Unwrap())
	require.Equal(t, status.Convert(encoded).Proto().Details,
		status.Convert(decodedWithCause).Proto().Details)
	require.Equal(t, http.StatusBadRequest,
		goahttp.NewErrorResponse(context.Background(), merged).StatusCode())
}

func TestMergeErrorsRetainsStatusAndCustomTransportCause(t *testing.T) {
	for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
		t.Run(code.String(), func(t *testing.T) {
			transportStatus, err := status.New(code, "transport diagnostic").WithDetails(
				wrapperspb.String("first detail"),
				wrapperspb.Int64(42),
			)
			require.NoError(t, err)
			original := &mergeTransportCause{transportStatus: transportStatus}
			var ctx context.Context
			if code == codes.Canceled {
				canceled, cancel := context.WithCancel(context.Background())
				cancel()
				ctx = canceled
			} else {
				deadline, cancel := context.WithTimeout(context.Background(), 0)
				defer cancel()
				ctx = deadline
			}
			correlated := goagrpc.ContextError(ctx, original)
			require.NotNil(t, correlated)
			left := goa.NewServiceError(correlated, "interrupted", true, false, false)
			right := goa.NewServiceError(ctx.Err(), "context", true, false, false)
			merged := goa.MergeErrors(left, right)
			require.Same(t, correlated, left.Unwrap())
			require.Equal(t, original.Error(), left.Message)
			require.ErrorIs(t, merged, ctx.Err())
			require.ErrorIs(t, merged, original)
			var got *mergeTransportCause
			require.ErrorAs(t, merged, &got)
			require.Same(t, original, got)
			require.Same(t, transportStatus, got.GRPCStatus())
			require.Equal(t, code, status.Code(merged))
			require.Equal(t, transportStatus.Proto().Details, status.Convert(merged).Proto().Details)
			require.Equal(t, merged.Error(), status.Convert(merged).Message())
			var whole *goa.ServiceError
			require.ErrorAs(t, merged, &whole)
			require.Same(t, correlated, whole.History()[0].Unwrap())
			var missing *goagrpc.ClientError
			require.False(t, errors.As(merged, &missing))
			require.NotErrorIs(t, merged, errors.New("absent"))
		})
	}
}

func (e *mergeTransportCause) Error() string {
	return e.transportStatus.Err().Error()
}

func (e *mergeTransportCause) GRPCStatus() *status.Status {
	return e.transportStatus
}
