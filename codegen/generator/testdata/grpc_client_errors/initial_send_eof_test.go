// These tests make the initial request send return EOF while the real server's
// reply remains unread. Generated clients must keep the stream so callers can
// receive the server's error without making successful opening wait for data.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// initialSendProbe controls the initial SendMsg result. For EOF, it waits
	// for the real server's terminal headers without consuming its response.
	// The generated receiver must still decode that response for the caller.
	initialSendProbe struct {
		grpc.ClientStream
		sendError error
		beforeEOF func()
		sends     int
		receives  int
		headers   int
	}
)

// TestGeneratedInitialSendEOFErrors preserves terminal errors when the initial
// envelope send reports EOF before the generated endpoint returns its stream.
func TestGeneratedInitialSendEOFErrors(t *testing.T) {
	for _, method := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		detail   func(*string) proto.Message
		receive  func(any) error
	}{
		{"bidirectional", (*genclient.Client).Exchange,
			func(reason *string) proto.Message {
				return &genpb.ExchangeDeniedError{Reason: reason}
			},
			func(result any) error {
				_, err := result.(gencatalog.ExchangeClientStream).Recv()
				return err
			}},
		{"client streaming", (*genclient.Client).Upload,
			func(reason *string) proto.Message {
				return &genpb.UploadDeniedError{Reason: reason}
			},
			func(result any) error {
				_, err := result.(gencatalog.UploadClientStream).CloseAndRecv()
				return err
			}},
	} {
		for _, test := range []struct {
			name   string
			detail proto.Message
			want   string
			cancel bool
		}{
			{"declared", method.detail(proto.String("not available")), "denied", false},
			{"malformed declared", method.detail(nil), goa.MissingField, false},
			{"generic", &goapb.ErrorResponse{
				Name: "rejected", Id: "initial-send-error", Msg: "not available",
			}, "rejected", false},
			{"generic before caller cancellation", &goapb.ErrorResponse{
				Name: "rejected", Id: "initial-send-error", Msg: "not available",
			}, "rejected", true},
		} {
			for _, wrapped := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/wrapped=%t", method.name, test.name, wrapped), func(t *testing.T) {
					ctx, cancel := context.WithCancel(catalogContext(t))
					defer cancel()
					remoteError := statusWithDetail(t, codes.PermissionDenied, test.detail)
					sendError := io.EOF
					if wrapped {
						sendError = fmt.Errorf("initial send: %w", io.EOF)
					}
					var calls atomic.Int32
					var probe *initialSendProbe
					var opens int
					client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
						calls.Add(1)
						return sendCatalogReply(stream, catalogReply{err: remoteError})
					}, grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, name string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
						opens++
						stream, err := streamer(ctx, desc, conn, name, opts...)
						if err != nil {
							return nil, err
						}
						probe = &initialSendProbe{
							ClientStream: stream,
							sendError:    sendError,
							beforeEOF: func() {
								if test.cancel {
									cancel()
								}
							},
						}
						return probe, nil
					}))
					result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, 1, probe.sends)
					require.Zero(t, probe.receives, "opening must not consume the terminal response")
					require.Zero(t, probe.headers, "opening must not add a header-read protocol")
					if test.cancel {
						require.ErrorIs(t, ctx.Err(), context.Canceled)
					} else {
						require.NoError(t, ctx.Err())
					}

					err = method.receive(result)
					require.NotErrorIs(t, err, io.EOF)
					require.NotErrorIs(t, err, context.Canceled)
					switch test.want {
					case "denied":
						var denied *gencatalog.Denied
						require.ErrorAs(t, err, &denied)
						require.Equal(t, "denied", denied.GoaErrorName())
						require.Equal(t, "not available", denied.Reason)
					case goa.MissingField:
						validation := requireValidation(t, err, goa.MissingField)
						require.NotNil(t, validation.Field)
						require.Equal(t, "reason", *validation.Field)
					case "rejected":
						decoded := requireGenericCause(t, err, codes.PermissionDenied)
						require.Equal(t, "rejected", decoded.Name)
						require.Equal(t, "initial-send-error", decoded.ID)
					}
					require.Equal(t, 1, probe.receives)
					require.Equal(t, 1, opens)
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedInitialSendOpening keeps successful idle streams asynchronous
// and leaves non-EOF send errors on the existing endpoint error path.
func TestGeneratedInitialSendOpening(t *testing.T) {
	for _, method := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
	}{
		{"bidirectional", (*genclient.Client).Exchange},
		{"client streaming", (*genclient.Client).Upload},
	} {
		for _, failSend := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fail-send=%t", method.name, failSend), func(t *testing.T) {
				ctx, cancel := context.WithCancel(catalogContext(t))
				defer cancel()
				var sendError error
				if failSend {
					sendError = statusWithDetail(t, codes.Unavailable, &goapb.ErrorResponse{
						Name: "send_failed", Id: "local-send-error", Msg: "cannot send initial request",
					})
				}
				var probe *initialSendProbe
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					if err := stream.RecvMsg(&emptypb.Empty{}); err != nil {
						return err
					}
					// An idle provider has no request to receive. Cancellation
					// ends this real RPC after the opening assertions.
					<-stream.Context().Done()
					return stream.Context().Err()
				}, grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, name string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
					stream, err := streamer(ctx, desc, conn, name, opts...)
					if err != nil {
						return nil, err
					}
					probe = &initialSendProbe{ClientStream: stream, sendError: sendError}
					return probe, nil
				}))
				result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
				if failSend {
					require.Nil(t, result)
					decoded := requireGenericCause(t, err, codes.Unavailable)
					require.Equal(t, "send_failed", decoded.Name)
					require.Same(t, sendError, errors.Unwrap(err))
				} else {
					require.NoError(t, err)
					require.NotNil(t, result)
				}
				require.NoError(t, ctx.Err())
				require.Equal(t, 1, probe.sends)
				require.Zero(t, probe.receives)
				require.Zero(t, probe.headers)
			})
		}
	}
}

func (s *initialSendProbe) SendMsg(message any) error {
	s.sends++
	if s.sendError != nil && !errors.Is(s.sendError, io.EOF) {
		return s.sendError
	}
	err := s.ClientStream.SendMsg(message)
	if s.sendError == nil {
		return err
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	// Header waits for the actual terminal status without decoding it. The
	// fixture then exposes EOF at SendMsg, leaving RecvMsg for generated code.
	if _, err := s.ClientStream.Header(); err != nil {
		return err
	}
	s.beforeEOF()
	return s.sendError
}

func (s *initialSendProbe) RecvMsg(message any) error {
	s.receives++
	return s.ClientStream.RecvMsg(message)
}

func (s *initialSendProbe) Header() (metadata.MD, error) {
	s.headers++
	return s.ClientStream.Header()
}
