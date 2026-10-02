// These tests return complete declarations and independent failures through
// generated handlers. They inspect the wire status before client decoding,
// then check which fields, concrete error, and RPC cause the caller receives.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	declaredOwnershipCase struct {
		name   string
		err    error
		code   codes.Code
		custom *gencatalog.Denied
		owner  *goa.ServiceError
		prefix *status.Status
	}
)

func TestGeneratedDeclaredErrorOwnership(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, test := range declaredOwnershipCases(t, reverse) {
			for _, wrapped := range []bool{false, true} {
				for _, method := range []string{"ReadErrors", "RetryErrors", "Watch", "Upload", "Exchange", "Collect"} {
					t.Run(fmt.Sprintf("%s/reverse=%t/wrapped=%t/%s", test.name, reverse, wrapped, method), func(t *testing.T) {
						input := test.err
						if wrapped {
							input = fmt.Errorf("catalog call: %w", input)
						}
						var calls atomic.Int32
						var original error
						var transport grpc.ClientStream
						completed := make(chan struct{}, 2)
						client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
							calls.Add(1)
							defer func() { completed <- struct{}{} }()
							return nil, input
						},
							grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
								original = invoke(ctx, method, request, reply, conn, opts...)
								return original
							}),
							grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
								stream, err := streamer(ctx, desc, conn, method, opts...)
								if err != nil {
									return nil, err
								}
								transport = stream
								return &receivedStatus{stream, &original}, nil
							}),
						)
						ctx := catalogContext(t)
						err := callDeclaredFailure(t, ctx, client, method)
						require.NoError(t, ctx.Err())
						require.NotErrorIs(t, err, context.Canceled)
						require.NotErrorIs(t, err, context.DeadlineExceeded)
						require.Equal(t, test.code, status.Code(original))
						message := input.Error()
						if test.prefix != nil && !wrapped {
							message = test.prefix.Message()
						}
						raw := status.Convert(original)
						require.Equal(t, message, raw.Message())
						details := raw.Proto().Details
						prefixCount := 0
						if test.prefix != nil {
							prefix := test.prefix.Proto().Details
							prefixCount = len(prefix)
							require.Len(t, details, prefixCount+1)
							require.Equal(t, prefix, details[:prefixCount])
						} else {
							require.Len(t, details, 1)
						}
						if test.custom != nil {
							require.True(t, proto.Equal(deniedWireDetail(method, test.custom), goagrpc.DecodeError(original)))
							decoded, ok := err.(*gencatalog.Denied)
							require.True(t, ok, "received %T", err)
							require.Equal(t, test.custom, decoded)
							require.Equal(t, "", decoded.Error())
							require.Nil(t, errors.Unwrap(err))
						} else {
							response, ok := raw.Details()[prefixCount].(*goapb.ErrorResponse)
							require.True(t, ok)
							requireDeclaredResponse(t, response, input, test.owner)
							first := goagrpc.DecodeError(original)
							if response, ok := first.(*goapb.ErrorResponse); ok {
								decoded := requireGenericCause(t, err, test.code)
								require.Equal(t, response.Msg, decoded.Error())
								require.Same(t, original, errors.Unwrap(err))
							} else if method == "ReadErrors" {
								decoded, ok := err.(*goa.ServiceError)
								require.True(t, ok)
								require.Equal(t, original.Error(), decoded.Error())
								require.Nil(t, errors.Unwrap(err))
							} else if method == "RetryErrors" {
								require.Same(t, original, errors.Unwrap(err))
							} else {
								require.Same(t, original, err)
							}
						}
						require.EqualValues(t, 1, calls.Load())
						select {
						case <-completed:
						case <-ctx.Done():
							t.Error("the generated handler did not finish")
						}
						if transport != nil {
							select {
							case <-transport.Context().Done():
							case <-ctx.Done():
								t.Error("the terminal status did not close the stream")
							}
						}
					})
				}
			}
		}
	}
}

// TestGeneratedDeclaredNoDeclarationControls keeps methods without a declared
// decoder on their existing generic unary or raw stream error path.
func TestGeneratedDeclaredNoDeclarationControls(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			for _, method := range []string{"Read", "Count", "WatchRaw", "CollectRaw"} {
				t.Run(fmt.Sprintf("reverse=%t/wrapped=%t/%s", reverse, wrapped, method), func(t *testing.T) {
					var input error = errors.Join(&gencatalog.Denied{Reason: "not selected"}, status.Error(codes.Canceled, "independent stop"))
					if reverse {
						input = errors.Join(status.Error(codes.Canceled, "independent stop"), &gencatalog.Denied{Reason: "not selected"})
					}
					if wrapped {
						input = fmt.Errorf("catalog call: %w", input)
					}
					var calls atomic.Int32
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						return nil, input
					})
					ctx := catalogContext(t)
					err := callDeclaredFailure(t, ctx, client, method)
					require.Equal(t, codes.Unknown, status.Code(err))
					if method == "Read" || method == "Count" {
						decoded := requireGenericCause(t, err, codes.Unknown)
						require.Equal(t, input.Error(), decoded.Message)
					} else {
						require.Equal(t, input.Error(), status.Convert(err).Message())
						require.Nil(t, errors.Unwrap(err))
					}
					require.EqualValues(t, 1, calls.Load())
					require.NoError(t, ctx.Err())
				})
			}
		}
	}
}

// requireDeclaredResponse compares owned fields exactly and requires a fresh
// whole fault when independent causes supply no outer Goa field owner.
func requireDeclaredResponse(t *testing.T, response *goapb.ErrorResponse, input error, owner *goa.ServiceError) {
	t.Helper()
	if owner != nil {
		require.Equal(t, owner.Name, response.Name)
		require.Equal(t, owner.ID, response.Id)
		require.Equal(t, owner.Message, response.Msg)
		require.Equal(t, owner.Timeout, response.Timeout)
		require.Equal(t, owner.Temporary, response.Temporary)
		require.Equal(t, owner.Fault, response.Fault)
		require.Equal(t, goagrpc.NewErrorResponse(owner).History, response.History)
		return
	}
	require.Equal(t, "fault", response.Name)
	require.NotEmpty(t, response.Id)
	require.Equal(t, input.Error(), response.Msg)
	require.True(t, response.Fault)
	require.False(t, response.Timeout)
	require.False(t, response.Temporary)
	require.Empty(t, response.History)
	var child *goa.ServiceError
	if errors.As(input, &child) {
		require.NotEqual(t, child.ID, response.Id)
	}
}

// callDeclaredFailure receives the actual terminal response for each generated
// unary, receive, result-close, and no-result-close method.
func callDeclaredFailure(t *testing.T, ctx context.Context, client *genclient.Client, method string) error {
	t.Helper()
	payload := &gencatalog.Selection{Key: "book"}
	switch method {
	case "ReadErrors":
		_, err := client.ReadErrors()(ctx, payload)
		return err
	case "RetryErrors":
		_, err := client.RetryErrors()(ctx, payload)
		return err
	case "Read":
		_, err := client.Read()(ctx, payload)
		return err
	case "Count":
		_, err := client.Count()(ctx, payload)
		return err
	case "Watch":
		result, err := client.Watch()(ctx, payload)
		require.NoError(t, err)
		_, err = result.(gencatalog.WatchClientStream).Recv()
		return err
	case "Exchange":
		result, err := client.Exchange()(ctx, payload)
		require.NoError(t, err)
		_, err = result.(gencatalog.ExchangeClientStream).Recv()
		return err
	case "Upload":
		result, err := client.Upload()(ctx, payload)
		require.NoError(t, err)
		_, err = result.(gencatalog.UploadClientStream).CloseAndRecv()
		return err
	case "Collect":
		result, err := client.Collect()(ctx, nil)
		require.NoError(t, err)
		return result.(gencatalog.CollectClientStream).Close()
	case "WatchRaw":
		result, err := client.WatchRaw()(ctx, payload)
		require.NoError(t, err)
		_, err = result.(gencatalog.WatchRawClientStream).Recv()
		return err
	case "CollectRaw":
		result, err := client.CollectRaw()(ctx, nil)
		require.NoError(t, err)
		return result.(gencatalog.CollectRawClientStream).Close()
	default:
		t.Fatalf("unknown method %q", method)
		return nil
	}
}

func deniedWireDetail(method string, denied *gencatalog.Denied) proto.Message {
	reason := proto.String(denied.Reason)
	switch method {
	case "ReadErrors":
		return &genpb.ReadErrorsDeniedError{Reason: reason}
	case "RetryErrors":
		return &genpb.RetryErrorsDeniedError{Reason: reason}
	case "Watch":
		return &genpb.WatchDeniedError{Reason: reason}
	case "Upload":
		return &genpb.UploadDeniedError{Reason: reason}
	case "Exchange":
		return &genpb.ExchangeDeniedError{Reason: reason}
	default:
		return &genpb.CollectDeniedError{Reason: reason}
	}
}
