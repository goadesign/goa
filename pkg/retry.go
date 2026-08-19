// This file implements the retry behavior used by generated clients. The
// transport packages preserve whether a failed call can be attempted again;
// this package owns when a typed service invocation is replayed.
package goa

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

const retryBackoff = 100 * time.Millisecond

// RetryEndpoint retries endpoint once when the returned error describes a
// temporary failure. Generated clients use temporaryErrorNames for designed
// errors whose retry trait is known from the service design. The original
// context and request value are reused, so callers retain their deadline and
// cancellation behavior.
func RetryEndpoint(endpoint Endpoint, temporaryErrorNames ...string) Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		response, err := endpoint(ctx, request)
		if err == nil || ctx.Err() != nil || !isRetryableError(err, temporaryErrorNames) {
			return response, err
		}

		timer := time.NewTimer(retryDelay())
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return endpoint(ctx, request)
		}
	}
}

// isRetryableError recognizes transport errors, Goa's default service error,
// and generated error values named as temporary by the method design.
func isRetryableError(err error, temporaryErrorNames []string) bool {
	var serviceError *ServiceError
	if errors.As(err, &serviceError) && serviceError.Temporary {
		return true
	}

	var retryable interface {
		Retryable() bool
	}
	if errors.As(err, &retryable) && retryable.Retryable() {
		return true
	}

	var namer GoaErrorNamer
	if !errors.As(err, &namer) {
		return false
	}
	name := namer.GoaErrorName()
	for _, temporaryName := range temporaryErrorNames {
		if name == temporaryName {
			return true
		}
	}
	return false
}

// retryDelay spreads concurrent retries over the interval from half to one
// and a half times the base delay.
func retryDelay() time.Duration {
	return retryBackoff/2 + time.Duration(rand.Int64N(int64(retryBackoff)))
}
