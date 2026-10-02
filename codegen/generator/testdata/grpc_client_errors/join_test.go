// These tests return independent failures from generated Catalog handlers and
// inspect both the raw RPC status and the error exposed by generated clients.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// operationStatus gives the complete operation a status independent of its
	// retained causes, so the encoder must preserve its ordered details.
	operationStatus struct {
		cause error
		st    *status.Status
	}

	// receivedStatus records the error before the generated stream decodes it.
	receivedStatus struct {
		grpc.ClientStream
		original *error
	}

	rpcEncodingCase struct {
		name  string
		err   error
		code  codes.Code
		owner *goa.ServiceError
		st    *status.Status
	}
)

func TestGeneratedWholeOperationEncoding(t *testing.T) {
	for _, test := range rpcEncodingCases(t) {
		for _, wrap := range []bool{false, true} {
			for _, method := range []string{"Read", "Count", "Watch", "Retry"} {
				t.Run(fmt.Sprintf("%s/wrapped=%t/%s", test.name, wrap, method), func(t *testing.T) {
					input := test.err
					if wrap {
						input = fmt.Errorf("catalog operation: %w", input)
					}
					var calls atomic.Int32
					var original error
					var transportStream grpc.ClientStream
					completed := make(chan struct{})
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						if calls.Add(1) == 1 {
							defer close(completed)
						}
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
							transportStream = stream
							return &receivedStatus{stream, &original}, nil
						}),
					)
					ctx := catalogContext(t)
					err := callCatalogFailure(t, ctx, client, method)
					require.NoError(t, ctx.Err(), "a server failure does not end the caller context")
					require.NotErrorIs(t, err, context.Canceled)
					require.NotErrorIs(t, err, context.DeadlineExceeded)
					expectedCalls := 1
					if method == "Retry" && test.st != nil {
						if response, ok := goagrpc.DecodeError(test.st.Err()).(*goapb.ErrorResponse); ok && response.Temporary {
							expectedCalls = 2
						}
					}
					require.EqualValues(t, expectedCalls, calls.Load())
					if method == "Watch" {
						select {
						case <-transportStream.Context().Done():
						default:
							t.Error("the transport stream did not close after its final status")
						}
					}
					select {
					case <-completed:
					default:
						t.Error("the service operation did not finish")
					}
					require.Equal(t, test.code, status.Code(original))
					raw := status.Convert(original)
					rawMessage := input.Error()
					prefix := 0
					if test.st != nil {
						if !wrap {
							rawMessage = test.st.Message()
						}
						prefix = len(test.st.Proto().Details)
						require.Equal(t, test.st.Proto().Details, raw.Proto().Details[:prefix])
					}
					require.Equal(t, rawMessage, raw.Message())
					require.Len(t, raw.Details(), prefix+1)
					if prefix > 0 {
						// The original first detail still controls decoding.
						// Appending a generic detail does not replace it.
						first := goagrpc.DecodeError(original)
						require.True(t, proto.Equal(goagrpc.DecodeError(test.st.Err()), first))
						appendedMessage := input.Error()
						if test.owner != nil {
							appendedMessage = test.owner.Message
						}
						require.Equal(t, appendedMessage, raw.Details()[prefix].(*goapb.ErrorResponse).Msg)
						if response, ok := first.(*goapb.ErrorResponse); ok {
							decoded := requireGenericCause(t, err, test.code)
							require.True(t, proto.Equal(response, goagrpc.NewErrorResponse(decoded)))
							require.Equal(t, response.Msg, err.Error())
							require.Same(t, original, errors.Unwrap(err))
						} else if method == "Watch" {
							require.Same(t, original, err)
							require.Equal(t, original.Error(), err.Error())
						} else {
							decoded, ok := err.(*goa.ServiceError)
							require.True(t, ok)
							require.Equal(t, "fault", decoded.Name)
							require.True(t, decoded.Fault)
							require.False(t, decoded.Timeout)
							require.False(t, decoded.Temporary)
							require.Equal(t, original.Error(), err.Error())
							if method == "Retry" {
								require.Same(t, original, errors.Unwrap(err))
								require.Equal(t, test.code, status.Code(err))
							} else {
								require.Nil(t, errors.Unwrap(err))
								require.Equal(t, codes.Unknown, status.Code(err))
							}
						}
						return
					}
					decoded := requireGenericCause(t, err, test.code)
					require.Same(t, original, errors.Unwrap(err))
					require.Equal(t, raw.Proto(), status.Convert(errors.Unwrap(err)).Proto())
					if test.owner != nil {
						require.Equal(t, test.owner.Name, decoded.Name)
						require.Equal(t, test.owner.ID, decoded.ID)
						require.Equal(t, test.owner.Message, decoded.Message)
						require.Equal(t, test.owner.Timeout, decoded.Timeout)
						require.Equal(t, test.owner.Temporary, decoded.Temporary)
						require.Equal(t, test.owner.Fault, decoded.Fault)
					} else {
						require.Equal(t, "fault", decoded.Name)
						require.NotEmpty(t, decoded.ID)
						require.Equal(t, input.Error(), decoded.Message)
						require.False(t, decoded.Timeout)
						require.False(t, decoded.Temporary)
						require.True(t, decoded.Fault)
						var child *goa.ServiceError
						if errors.As(input, &child) {
							require.NotEqual(t, child.ID, decoded.ID)
						}
					}
					require.Equal(t, decoded.Message, err.Error())
				})
			}
		}
	}
}

// TestGeneratedDeclaredJoinsUseCompleteResult keeps direct declared errors
// typed, while independent joins return the full text and original RPC cause.
func TestGeneratedDeclaredJoinsUseCompleteResult(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, wrap := range []bool{false, true} {
			for _, joined := range []bool{false, true} {
				t.Run(fmt.Sprintf("reverse=%t/wrapped=%t/joined=%t", reverse, wrap, joined), func(t *testing.T) {
					var input error = &gencatalog.Denied{Reason: "catalog access rejected"}
					if joined {
						cause := status.Error(codes.Canceled, "independent operation stopped")
						if reverse {
							input = errors.Join(cause, input)
						} else {
							input = errors.Join(input, cause)
						}
					}
					if wrap {
						input = fmt.Errorf("catalog operation: %w", input)
					}
					var calls atomic.Int32
					var original error
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						return nil, input
					}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						original = invoke(ctx, method, request, reply, conn, opts...)
						return original
					}))
					_, err := client.ReadErrors()(catalogContext(t), &gencatalog.Selection{Key: "book"})
					code := codes.PermissionDenied
					if joined {
						code = codes.Unknown
					}
					require.Equal(t, code, status.Code(original))
					require.Equal(t, input.Error(), status.Convert(original).Message())
					if joined {
						decoded := requireGenericCause(t, err, code)
						require.Equal(t, input.Error(), decoded.Message)
						require.Equal(t, "fault", decoded.Name)
						require.Same(t, original, errors.Unwrap(err))
					} else {
						denied, ok := err.(*gencatalog.Denied)
						require.True(t, ok)
						require.Equal(t, "catalog access rejected", denied.Reason)
						require.Nil(t, errors.Unwrap(err))
					}
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedEndedClientKeepsLocalCause lets an active request reach the
// service, ends the client context, and then waits for both sides to finish.
// The caller receives its local cancellation rather than server join details.
func TestGeneratedEndedClientKeepsLocalCause(t *testing.T) {
	for _, method := range []string{"Read", "Count", "Watch"} {
		t.Run(method, func(t *testing.T) {
			started, finished := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			client := newEncodingCatalogClient(t, func(ctx context.Context, _ any) (any, error) {
				calls.Add(1)
				close(started)
				<-ctx.Done()
				defer close(finished)
				return nil, errors.Join(status.Error(codes.Canceled, "server stopped"), goa.Fault("cleanup failed"))
			})
			ctx, cancel := context.WithCancel(catalogContext(t))
			defer cancel()
			result := make(chan error, 1)
			go func() {
				result <- callCatalogFailure(t, ctx, client, method)
			}()
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("request did not reach the service")
			}
			cancel()
			select {
			case err := <-result:
				require.ErrorIs(t, err, context.Canceled)
				require.Equal(t, codes.Canceled, status.Code(err))
			case <-catalogContext(t).Done():
				t.Fatal("client did not finish after cancellation")
			}
			select {
			case <-finished:
			case <-catalogContext(t).Done():
				t.Fatal("server operation did not finish after cancellation")
			}
			require.EqualValues(t, 1, calls.Load())
		})
	}
}

func rpcEncodingCases(t *testing.T) []rpcEncodingCase {
	t.Helper()
	canceled := status.Error(codes.Canceled, "remote operation stopped")
	deadline := status.Error(codes.DeadlineExceeded, "remote deadline")
	fault := goa.Fault("cleanup failed")
	whole, err := status.New(codes.Canceled, "explicit whole status").WithDetails(
		&emptypb.Empty{}, &goapb.ErrorResponse{Name: "later", Id: "later-id", Msg: "later detail"},
	)
	require.NoError(t, err)
	wholeGeneric, err := status.New(codes.DeadlineExceeded, "explicit complete deadline").WithDetails(
		&goapb.ErrorResponse{Name: "rejected", Id: "whole-id", Msg: "whole detail", Timeout: true, Temporary: true},
		&emptypb.Empty{},
	)
	require.NoError(t, err)
	later, err := anypb.New(&goapb.ErrorResponse{Name: "later", Id: "later-id", Msg: "later detail"})
	require.NoError(t, err)
	malformed := status.FromProto(&statuspb.Status{
		Code: int32(codes.Canceled), Message: "malformed first detail",
		Details: []*anypb.Any{
			{TypeUrl: "type.googleapis.com/goa.ErrorResponse", Value: []byte{0xff}}, later,
		},
	})
	unknown := status.FromProto(&statuspb.Status{
		Code: int32(codes.Canceled), Message: "unknown first detail",
		Details: []*anypb.Any{
			{TypeUrl: "type.googleapis.com/example.Unknown"}, later,
		},
	})
	owner := goa.NewServiceError(canceled, "bad_request", false, false, false)
	owner.ID, owner.Message = "request-id", "catalog selection rejected"
	cases := []rpcEncodingCase{
		{"owned single", owner, codes.Canceled, owner, nil},
		{"raw context duplicates", errors.Join(context.Canceled, context.Canceled), codes.Unknown, nil, nil},
	}
	for _, reverse := range []bool{false, true} {
		join := func(left, right error) error {
			if reverse {
				return errors.Join(right, left)
			}
			return errors.Join(left, right)
		}
		owned := goa.NewServiceError(join(canceled, fault), "bad_request", false, false, true)
		for _, test := range []rpcEncodingCase{
			{"mixed status fault", join(canceled, fault), codes.Unknown, nil, nil},
			{"different codes", join(canceled, deadline), codes.Unknown, nil, nil},
			{"same code", join(canceled, status.Error(codes.Canceled, "another operation stopped")), codes.Canceled, nil, nil},
			{"ordinary join", join(errors.New("read failed"), errors.New("close failed")), codes.Unknown, nil, nil},
			{"named join", join(fault, goa.Fault("write failed")), codes.Internal, nil, nil},
			{"temporary join", join(goa.TemporaryError("busy", "catalog unavailable"), goa.TemporaryError("pending", "catalog pending")), codes.Unavailable, nil, nil},
			{"timeout join", join(goa.PermanentTimeoutError("expired", "read expired"), goa.TemporaryTimeoutError("wait_expired", "wait expired")), codes.DeadlineExceeded, nil, nil},
			{"owned join", owned, codes.Internal, owned, nil},
			{"explicit whole join", &operationStatus{join(canceled, fault), whole}, codes.Canceled, nil, whole},
			{"explicit generic whole", &operationStatus{join(canceled, fault), wholeGeneric}, codes.DeadlineExceeded, nil, wholeGeneric},
			{"malformed first whole", &operationStatus{join(canceled, fault), malformed}, codes.Canceled, nil, malformed},
			{"unknown first whole", &operationStatus{join(canceled, fault), unknown}, codes.Canceled, nil, unknown},
		} {
			test.name += fmt.Sprintf("/reverse=%t", reverse)
			cases = append(cases, test)
		}
	}
	return cases
}

// callCatalogFailure invokes each generated method and receives the final
// server-stream error, which closes the transport stream.
func callCatalogFailure(t *testing.T, ctx context.Context, client *genclient.Client, method string) error {
	t.Helper()
	payload := &gencatalog.Selection{Key: "book"}
	switch method {
	case "Read":
		_, err := client.Read()(ctx, payload)
		return err
	case "Count":
		_, err := client.Count()(ctx, payload)
		return err
	case "Retry":
		_, err := client.Retry()(ctx, payload)
		return err
	default:
		result, err := client.Watch()(ctx, payload)
		if err != nil {
			return err
		}
		stream := result.(gencatalog.WatchClientStream)
		_, err = stream.Recv()
		return err
	}
}

func (e *operationStatus) Error() string {
	return e.cause.Error()
}

func (e *operationStatus) GRPCStatus() *status.Status {
	return e.st
}

func (e *operationStatus) Unwrap() error {
	return e.cause
}

func (s *receivedStatus) RecvMsg(message any) error {
	err := s.ClientStream.RecvMsg(message)
	if err != nil {
		*s.original = err
	}
	return err
}
