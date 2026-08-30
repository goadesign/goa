package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

// TestNewErrorResponseHistory tests that history is correctly included for merged errors
func TestNewErrorResponseHistory(t *testing.T) {
	// Test simple error - should not have history
	simpleErr := goa.MissingFieldError("username", "body")
	resp := NewErrorResponse(simpleErr)
	assert.Nil(t, resp.History)

	// Test merged error - should have history
	mergedErr := goa.MergeErrors(
		goa.MissingFieldError("username", "body"),
		goa.InvalidFormatError("data", "{invalid}", goa.FormatJSON, fmt.Errorf("invalid JSON")),
	)
	mergedResp := NewErrorResponse(mergedErr)
	assert.NotNil(t, mergedResp.History)
	assert.Len(t, mergedResp.History, 2)
	assert.Equal(t, goa.MissingField, mergedResp.History[0].Name)
	assert.Equal(t, "username", mergedResp.History[0].Field)
	assert.Equal(t, goa.InvalidFormat, mergedResp.History[1].Name)
	assert.Equal(t, "data", mergedResp.History[1].Field)
}

// TestEncodeErrorStatusCodes tests that validation errors get mapped to InvalidArgument
func TestEncodeErrorStatusCodes(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "missing_field",
			err:          goa.MissingFieldError("username", "body"),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid_format",
			err:          goa.InvalidFormatError("data", "{invalid}", goa.FormatJSON, fmt.Errorf("invalid JSON")),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "decode_payload",
			err:          &goa.ServiceError{Name: "decode_payload", Message: "failed to decode"},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "timeout",
			err:          &goa.ServiceError{Name: "timeout", Message: "timed out", Timeout: true},
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name:         "fault",
			err:          goa.Fault("internal error"),
			expectedCode: codes.Internal,
		},
		{
			name:         "temporary",
			err:          &goa.ServiceError{Name: "unavailable", Message: "service unavailable", Temporary: true},
			expectedCode: codes.Unavailable,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			encoded := EncodeError(c.err)
			st, ok := status.FromError(encoded)
			assert.True(t, ok)
			assert.Equal(t, c.expectedCode, st.Code())

			// Check that details are included
			details := st.Details()
			assert.Len(t, details, 1)
			_, ok = details[0].(*goapb.ErrorResponse)
			assert.True(t, ok)
		})
	}
}

func TestInvalidLengthErrorFitsGRPCHeaders(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnknownServiceHandler(func(any, grpc.ServerStream) error {
		return EncodeError(goa.InvalidLengthError("payload", make([]byte, 1<<20), 1<<20, 1024, false))
	}))
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	conn, err := grpc.NewClient(
		"passthrough:///bounded-validation-error",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithMaxHeaderListSize(8*1024),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close())
		server.Stop()
		serveResult := <-serveErr
		if serveResult != nil {
			require.ErrorIs(t, serveResult, grpc.ErrServerStopped)
		}
	})

	err = conn.Invoke(
		context.Background(),
		"/test.Validation/Check",
		&emptypb.Empty{},
		&emptypb.Empty{},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	response, ok := DecodeError(err).(*goapb.ErrorResponse)
	require.True(t, ok)
	require.Equal(t, goa.InvalidLength, response.Name)
	require.Equal(t, "length of payload must be at most 1024 but got 1048576", response.Msg)
}

func TestNewTransportError(t *testing.T) {
	cases := []struct {
		name      string
		code      codes.Code
		temporary bool
		timeout   bool
	}{
		{
			name:      "unavailable",
			code:      codes.Unavailable,
			temporary: true,
		},
		{
			name:    "deadline exceeded",
			code:    codes.DeadlineExceeded,
			timeout: true,
		},
		{
			name: "internal",
			code: codes.Internal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cause := status.Error(tc.code, tc.name)

			err := NewTransportError(cause)

			assert.Equal(t, tc.temporary, err.Temporary)
			assert.Equal(t, tc.timeout, err.Timeout)
			assert.True(t, err.Fault)
			assert.ErrorIs(t, err, cause)
		})
	}
}

func TestDecodeErrorUnknownDetail(t *testing.T) {
	transportErr := status.FromProto(&statuspb.Status{
		Code:    int32(codes.Unknown),
		Message: "unknown remote detail",
		Details: []*anypb.Any{{
			TypeUrl: "type.googleapis.com/example.UnknownError",
		}},
	}).Err()

	assert.NotPanics(t, func() {
		assert.Nil(t, DecodeError(transportErr))
	})
}

func TestContextError(t *testing.T) {
	transportErr := errors.New("transport failed")

	t.Run("active context has no context error", func(t *testing.T) {
		require.NoError(t, ContextError(context.Background(), transportErr))
	})

	t.Run("matching canceled status preserves text and context error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		transportStatus, statusErr := status.New(codes.Canceled, "transport canceled").WithDetails(
			&goapb.ErrorResponse{Name: "remote_canceled", Msg: "remote detail"},
		)
		require.NoError(t, statusErr)
		transportErr := transportStatus.Err()
		err := ContextError(ctx, transportErr)
		require.ErrorContains(t, err, transportErr.Error())
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, codes.Canceled, status.Code(err))
		require.Equal(t, transportStatus.Proto(), status.Convert(err).Proto())
	})

	t.Run("matching deadline status preserves text and context error", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer cancel()

		transportErr := status.Error(codes.DeadlineExceeded, "transport deadline")
		err := ContextError(ctx, transportErr)
		require.ErrorContains(t, err, transportErr.Error())
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, codes.DeadlineExceeded, status.Code(err))
	})

	t.Run("independent transport failure remains separate", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		require.NoError(t, ContextError(ctx, status.Error(codes.Internal, "server failed")))
	})

	t.Run("mismatched context and transport statuses remain separate", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		require.NoError(t, ContextError(
			canceledCtx,
			status.Error(codes.DeadlineExceeded, "transport deadline"),
		))

		deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer deadlineCancel()
		require.NoError(t, ContextError(
			deadlineCtx,
			status.Error(codes.Canceled, "transport canceled"),
		))
	})

	t.Run("non-gRPC context error remains separate", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		require.NoError(t, ContextError(ctx, context.Canceled))
	})
}
