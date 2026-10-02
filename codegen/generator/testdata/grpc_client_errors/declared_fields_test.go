// These tests keep owned Goa fields, designed retries, dynamic custom names,
// and non-object declarations intact through the actual generated methods.
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
	genpb "generated.local/gen/grpc/catalog/pb"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

func TestGeneratedDeclaredOwnedGoaFields(t *testing.T) {
	for mask := range 8 {
		for _, empty := range []bool{false, true} {
			for _, joined := range []bool{false, true} {
				for _, wrapped := range []bool{false, true} {
					for _, method := range []string{"ReadErrors", "RetryErrors"} {
						t.Run(fmt.Sprintf("traits=%d/empty=%t/joined=%t/wrapped=%t/%s", mask, empty, joined, wrapped, method), func(t *testing.T) {
							var cause error = errors.New("retained failure")
							if joined {
								cause = errors.Join(status.Error(codes.Canceled, "independent stop"), errors.New("cleanup failed"))
							}
							owner := goa.NewServiceError(cause, "busy", mask&1 != 0, mask&2 != 0, mask&4 != 0)
							owner.ID, owner.Message = "owned-id", "complete Busy diagnostic"
							code := codes.Unavailable
							if empty {
								owner.Name, owner.ID, owner.Message = "", "", ""
								code = codes.Unknown
								switch {
								case owner.Timeout:
									code = codes.DeadlineExceeded
								case owner.Fault:
									code = codes.Internal
								case owner.Temporary:
									code = codes.Unavailable
								}
							}
							var input error = owner
							if wrapped {
								input = fmt.Errorf("catalog call: %w", input)
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
							ctx := catalogContext(t)
							err := callDeclaredFailure(t, ctx, client, method)
							require.NoError(t, ctx.Err())
							require.Equal(t, code, status.Code(original))
							require.Equal(t, input.Error(), status.Convert(original).Message())
							decoded := requireGenericCause(t, err, code)
							require.Equal(t, owner.Name, decoded.Name)
							require.Equal(t, owner.ID, decoded.ID)
							require.Equal(t, owner.Message, decoded.Message)
							require.Equal(t, owner.Timeout, decoded.Timeout)
							require.Equal(t, owner.Temporary, decoded.Temporary)
							require.Equal(t, owner.Fault, decoded.Fault)
							require.Same(t, original, errors.Unwrap(err))
							count := int32(1)
							if method == "RetryErrors" && (!empty || owner.Temporary) {
								count = 2
							}
							require.Equal(t, count, calls.Load())
						})
					}
				}
			}
		}
	}
}

func TestGeneratedDeclaredDynamicAndNonObject(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		missing := &gencatalog.Missing{Key: "book"}
		limited := &gencatalog.Rejection{Name: "limited", Reason: "catalog capacity"}
		locked := &gencatalog.Rejection{Name: "locked", Reason: "catalog edit locked"}
		notice := gencatalog.Notice("not Error text")
		join := func(left, right error) error {
			if reverse {
				return errors.Join(right, left)
			}
			return errors.Join(left, right)
		}
		for _, test := range []struct {
			name     string
			input    error
			code     codes.Code
			custom   error
			response proto.Message
		}{
			{"missing", missing, codes.NotFound, missing, &genpb.InspectErrorsMissingError{Key: proto.String("book")}},
			{"limited", limited, codes.ResourceExhausted, limited, &genpb.InspectErrorsLimitedError{Name: proto.String("limited"), Reason: proto.String("catalog capacity")}},
			{"locked", locked, codes.FailedPrecondition, locked, &genpb.InspectErrorsLockedError{Name: proto.String("locked"), Reason: proto.String("catalog edit locked")}},
			{"notice", notice, codes.AlreadyExists, nil, nil},
			{"different types", join(missing, limited), codes.Unknown, nil, nil},
			{"dynamic names", join(limited, locked), codes.Unknown, nil, nil},
			{"custom and status", join(missing, status.Error(codes.Canceled, "independent stop")), codes.Unknown, nil, nil},
			{"notice and custom", join(notice, missing), codes.Unknown, nil, nil},
		} {
			for _, wrapped := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/reverse=%t/wrapped=%t", test.name, reverse, wrapped), func(t *testing.T) {
					input := test.input
					if wrapped {
						input = fmt.Errorf("catalog call: %w", input)
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
					ctx := catalogContext(t)
					_, err := client.InspectErrors()(ctx, &gencatalog.Selection{Key: "book"})
					require.NoError(t, ctx.Err())
					require.Equal(t, test.code, status.Code(original))
					require.Equal(t, input.Error(), status.Convert(original).Message())
					require.Len(t, status.Convert(original).Proto().Details, 1)
					if test.custom != nil {
						require.Equal(t, test.custom, err)
						require.True(t, proto.Equal(test.response, goagrpc.DecodeError(original)))
						require.Nil(t, errors.Unwrap(err))
					} else {
						response, ok := goagrpc.DecodeError(original).(*goapb.ErrorResponse)
						require.True(t, ok)
						requireDeclaredResponse(t, response, input, nil)
						decoded := requireGenericCause(t, err, test.code)
						require.Equal(t, input.Error(), decoded.Message)
						require.Same(t, original, errors.Unwrap(err))
					}
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedDeclaredOwnedCauseAndHistory returns an owned declaration with
// a status cause, then checks a merged error's whole fields and original messages.
func TestGeneratedDeclaredOwnedCauseAndHistory(t *testing.T) {
	for _, joined := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			for _, method := range []string{"ReadErrors", "RetryErrors"} {
				t.Run(fmt.Sprintf("joined=%t/wrapped=%t/%s", joined, wrapped, method), func(t *testing.T) {
					var cause error = status.Error(codes.Canceled, "retained stop")
					if joined {
						cause = errors.Join(cause, status.Error(codes.DeadlineExceeded, "retained deadline"))
					}
					owner := goa.NewServiceError(cause, "busy", false, true, false)
					owner.ID = "busy-cause-id"
					var input error = owner
					if wrapped {
						input = fmt.Errorf("catalog call: %w", input)
					}
					var calls atomic.Int32
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						return nil, input
					})
					err := callDeclaredFailure(t, catalogContext(t), client, method)
					decoded := requireGenericCause(t, err, codes.Unavailable)
					require.Equal(t, owner.ID, decoded.ID)
					require.Equal(t, owner.Message, decoded.Message)
					expected := int32(1)
					if method == "RetryErrors" {
						expected = 2
					}
					require.Equal(t, expected, calls.Load())
				})
			}
		}
	}
	var left, right *goa.ServiceError
	require.ErrorAs(t, goa.MissingFieldError("key", "selection"), &left)
	require.ErrorAs(t, goa.InvalidEnumValueError("state", "bad", []any{"open"}), &right)
	merged := goa.MergeErrors(left, right)
	client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
		return nil, merged
	})
	err := callDeclaredFailure(t, catalogContext(t), client, "ReadErrors")
	decoded := requireGenericCause(t, err, codes.InvalidArgument)
	require.Equal(t, left.Name, decoded.Name)
	require.Equal(t, left.ID, decoded.ID)
	require.Equal(t, left.Message+"; "+right.Message, decoded.Message)
	require.Equal(t, left.Timeout && right.Timeout, decoded.Timeout)
	require.Equal(t, left.Temporary && right.Temporary, decoded.Temporary)
	require.Equal(t, left.Fault && right.Fault, decoded.Fault)
	response := goagrpc.DecodeError(errors.Unwrap(err)).(*goapb.ErrorResponse)
	require.Equal(t, []*goapb.ErrorField{
		{Name: left.Name, Msg: left.Message, Field: "key"},
		{Name: right.Name, Msg: right.Message, Field: "state"},
	}, response.History)
}
