// These tests send synthetic protobuf replies through generated gRPC clients.
// The protobuf test server can omit required fields that generated service
// values always encode, so the client must validate the received response.
package clienterrors_test

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goagrpc "goa.design/goa/v3/grpc"
	goa "goa.design/goa/v3/pkg"
)

type (
	// catalogMethod identifies one generated endpoint and its existing retry rule.
	catalogMethod struct {
		name       string
		endpoint   func(*genclient.Client) goa.Endpoint
		idempotent bool
	}

	// catalogReply supplies one server response, including its transport metadata.
	catalogReply struct {
		message proto.Message
		header  metadata.MD
		trailer metadata.MD
		err     error
	}
)

var catalogMethods = []catalogMethod{
	{"ordinary", (*genclient.Client).Read, false},
	{"ordinary errors", (*genclient.Client).ReadErrors, false},
	{"idempotent", (*genclient.Client).Retry, true},
	{"idempotent errors", (*genclient.Client).RetryErrors, true},
}

// TestGeneratedResponseValidation returns precise validation errors for absent, empty, and
// unknown response values.
func TestGeneratedResponseValidation(t *testing.T) {
	for _, method := range catalogMethods {
		for _, test := range []struct {
			name    string
			state   *string
			errName string
		}{
			{"absent", nil, goa.MissingField},
			{"empty", proto.String(""), goa.InvalidEnumValue},
			{"unknown", proto.String("unknown"), goa.InvalidEnumValue},
			{"valid", proto.String("open"), ""},
		} {
			t.Run(method.name+"/"+test.name, func(t *testing.T) {
				var calls atomic.Int32
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					calls.Add(1)
					return sendCatalogReply(stream, catalogReply{
						message: &genpb.ReadResponse{State: test.state},
					})
				})
				result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				if test.errName == "" {
					require.NoError(t, err)
					entry, ok := result.(*gencatalog.Entry)
					require.True(t, ok)
					require.Equal(t, "open", entry.State)
				} else {
					require.Nil(t, result)
					serviceError := requireValidation(t, err, test.errName)
					if test.state == nil {
						require.NotNil(t, serviceError.Field)
						require.Equal(t, "state", *serviceError.Field)
					}
				}
				require.EqualValues(t, 1, calls.Load())
			})
		}
	}
}

// TestGeneratedMergedResponseValidation preserves both original errors when response
// validation finds two invalid fields.
func TestGeneratedMergedResponseValidation(t *testing.T) {
	for _, method := range catalogMethods {
		t.Run(method.name, func(t *testing.T) {
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				return sendCatalogReply(stream, catalogReply{
					message: &genpb.ReadResponse{State: proto.String("unknown"), Title: proto.String("")},
				})
			})
			result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
			require.Nil(t, result)
			validation := requireValidation(t, err, goa.InvalidEnumValue)
			history := validation.History()
			require.Len(t, history, 2)
			require.Equal(t, goa.InvalidEnumValue, history[0].Name)
			require.Equal(t, goa.InvalidLength, history[1].Name)
			require.NotEqual(t, history[0].ID, history[1].ID)
		})
	}
}

// TestGeneratedMetadataValidation checks header and trailer errors before returning a
// service result.
func TestGeneratedMetadataValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		header  metadata.MD
		trailer metadata.MD
		errName string
		history int
	}{
		{"valid", metadata.Pairs("count", "2"), metadata.Pairs("flags", "true", "flags", "false"), "", 0},
		{"missing header", nil, metadata.Pairs("flags", "true"), goa.MissingField, 1},
		{"missing trailer", metadata.Pairs("count", "2"), nil, goa.MissingField, 1},
		{"invalid header", metadata.Pairs("count", "bad"), metadata.Pairs("flags", "true"), goa.InvalidFieldType, 1},
		{"invalid trailer element", metadata.Pairs("count", "2"), metadata.Pairs("flags", "bad"), goa.InvalidFieldType, 1},
		{"merged errors", metadata.Pairs("count", "bad"), metadata.Pairs("flags", "bad"), goa.InvalidFieldType, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				return sendCatalogReply(stream, catalogReply{
					message: &genpb.ReadResponse{State: proto.String("open")},
					header:  test.header, trailer: test.trailer,
				})
			})
			result, err := client.Metadata()(catalogContext(t), nil)
			if test.errName == "" {
				require.NoError(t, err)
				entry, ok := result.(*gencatalog.MetadataEntry)
				require.True(t, ok)
				require.Equal(t, 2, entry.Count)
				require.Equal(t, []bool{true, false}, entry.Flags)
			} else {
				require.Nil(t, result)
				serviceError := requireValidation(t, err, test.errName)
				require.Len(t, serviceError.History(), test.history)
				for _, original := range serviceError.History() {
					require.NotEmpty(t, original.ID)
					require.Equal(t, test.errName, original.Name)
					require.NotNil(t, original.Field)
				}
			}
		})
	}
}

// TestGeneratedViews validates the selected view and permits fields omitted by that view.
func TestGeneratedViews(t *testing.T) {
	for _, test := range []struct {
		name    string
		view    string
		state   *string
		title   *string
		errName string
	}{
		{"compact", "compact", proto.String("open"), nil, ""},
		{"default", "default", proto.String("open"), proto.String("Book"), ""},
		{"missing compact state", "compact", nil, nil, goa.MissingField},
		{"missing title", "default", proto.String("open"), nil, goa.MissingField},
		{"unknown view", "unknown", proto.String("open"), proto.String("Book"), goa.InvalidEnumValue},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				return sendCatalogReply(stream, catalogReply{
					message: &genpb.ReadResponse{State: test.state, Title: test.title},
					header:  metadata.Pairs("goa-view", test.view),
				})
			})
			result, err := client.View()(catalogContext(t), nil)
			if test.errName == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				require.Nil(t, result)
				requireValidation(t, err, test.errName)
			}
		})
	}
}

// TestGeneratedWrongPayloadDoesNotCallRPC rejects the wrong Go payload type without calling
// the server.
func TestGeneratedWrongPayloadDoesNotCallRPC(t *testing.T) {
	for _, method := range catalogMethods {
		t.Run(method.name, func(t *testing.T) {
			var calls atomic.Int32
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				calls.Add(1)
				return sendCatalogReply(stream, catalogReply{message: &genpb.ReadResponse{State: proto.String("open")}})
			})
			result, err := method.endpoint(client)(catalogContext(t), "not a selection")
			require.Nil(t, result)
			var clientError *goagrpc.ClientError
			require.ErrorAs(t, err, &clientError)
			require.Equal(t, "invalid_type", clientError.Name)
			require.False(t, clientError.Fault)
			require.EqualValues(t, 0, calls.Load())
		})
	}
}

// TestInvokerPreservesLocalErrors keeps decoder type errors and application callback errors
// unchanged.
func TestInvokerPreservesLocalErrors(t *testing.T) {
	t.Run("wrong response Go type", func(t *testing.T) {
		invoker := goagrpc.NewInvoker(
			func(context.Context, any, ...grpc.CallOption) (any, error) {
				return "not protobuf", nil
			}, nil, genclient.DecodeReadResponse)
		result, err := invoker.Invoke(catalogContext(t), nil)
		require.Nil(t, result)
		var clientError *goagrpc.ClientError
		require.ErrorAs(t, err, &clientError)
		require.Equal(t, "invalid_type", clientError.Name)
	})
	for _, stage := range []string{"encoder", "decoder"} {
		t.Run(stage, func(t *testing.T) {
			original := errors.New("application callback failed")
			calls := 0
			remote := func(context.Context, any, ...grpc.CallOption) (any, error) {
				calls++
				return &genpb.ReadResponse{State: proto.String("open")}, nil
			}
			var encoder goagrpc.RequestEncoder
			var decoder goagrpc.ResponseDecoder
			if stage == "encoder" {
				encoder = func(context.Context, any, *metadata.MD) (any, error) {
					return nil, original
				}
			} else {
				decoder = func(context.Context, any, metadata.MD, metadata.MD) (any, error) {
					return nil, original
				}
			}
			result, err := goagrpc.NewInvoker(remote, encoder, decoder).Invoke(catalogContext(t), nil)
			require.Nil(t, result)
			require.Same(t, original, err)
			if stage == "encoder" {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}

// newCatalogClient connects the generated client to a bounded local server.
// The handler controls protobuf replies; all client conversion code is generated.
func newCatalogClient(t *testing.T, handler grpc.StreamHandler, options ...grpc.DialOption) *genclient.Client {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnknownServiceHandler(handler))
	done := make(chan error, 1)
	go func() {
		done <- server.Serve(listener)
	}()
	options = append(options,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	conn, err := grpc.NewClient("passthrough:///catalog", options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close())
		server.Stop()
		select {
		case err := <-done:
			if err != nil {
				require.ErrorIs(t, err, grpc.ErrServerStopped)
			}
		case <-time.After(5 * time.Second):
			t.Error("local gRPC server did not stop")
		}
	})
	return genclient.NewClient(conn)
}

// sendCatalogReply consumes a request and sends the specified wire response.
// Error responses deliberately do not pass through a generated service encoder.
func sendCatalogReply(stream grpc.ServerStream, reply catalogReply) error {
	if err := stream.RecvMsg(&emptypb.Empty{}); err != nil {
		return err
	}
	if len(reply.header) > 0 {
		if err := stream.SendHeader(reply.header); err != nil {
			return err
		}
	}
	stream.SetTrailer(reply.trailer)
	if reply.err != nil {
		return reply.err
	}
	return stream.SendMsg(reply.message)
}

// requireValidation checks the public validation error without parsing its text.
func requireValidation(t *testing.T, err error, name string) *goa.ServiceError {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotEmpty(t, serviceError.ID)
	require.False(t, serviceError.Temporary)
	require.False(t, serviceError.Timeout)
	require.False(t, serviceError.Fault)
	return serviceError
}

// catalogContext bounds each local RPC, including retries, to five seconds.
// Test cleanup cancels any unfinished calls.
func catalogContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
