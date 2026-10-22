// These tests pass a native stop through two generated services. Each hop must
// preserve the stopped operation without describing it as a permanent fault.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gencatalog "generated.local/gen/catalog"
	goa "goa.design/goa/v3/pkg"
)

// TestGeneratedForwardedContextStop keeps a downstream cancellation or deadline
// intact when another generated service wraps and returns the received error.
func TestGeneratedForwardedContextStop(t *testing.T) {
	for _, method := range catalogMethods {
		for _, stop := range []struct {
			name string
			err  error
			code codes.Code
		}{
			{"canceled", context.Canceled, codes.Canceled},
			{"deadline", context.DeadlineExceeded, codes.DeadlineExceeded},
		} {
			t.Run(method.name+"/"+stop.name, func(t *testing.T) {
				var downstreamCalls, forwardingCalls atomic.Int32
				downstream := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					downstreamCalls.Add(1)
					return nil, stop.err
				})
				forwarding := newEncodingCatalogClient(t, func(ctx context.Context, payload any) (any, error) {
					forwardingCalls.Add(1)
					result, err := downstream.Read()(ctx, payload)
					if err != nil {
						return nil, fmt.Errorf("read downstream catalog: %w", err)
					}
					return result, nil
				})
				ctx := catalogContext(t)
				result, err := method.endpoint(forwarding)(ctx, &gencatalog.Selection{Key: "book"})
				require.Error(t, err)
				assert.Nil(t, result)
				assert.NoError(t, ctx.Err())
				assert.Equal(t, stop.code, status.Code(err))
				assert.Empty(t, status.Convert(err).Details())
				assert.ErrorContains(t, err, "read downstream catalog")
				var serviceError *goa.ServiceError
				assert.False(t, errors.As(err, &serviceError))
				assert.NotErrorIs(t, err, context.Canceled)
				assert.NotErrorIs(t, err, context.DeadlineExceeded)
				assert.EqualValues(t, 1, downstreamCalls.Load())
				assert.EqualValues(t, 1, forwardingCalls.Load())
			})
		}
	}
}
