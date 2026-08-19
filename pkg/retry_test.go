package goa_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	goa "goa.design/goa/v3/pkg"
)

type (
	retryableTestError struct{}
	namedTestError     struct{}
)

func (retryableTestError) Error() string {
	return "retryable transport error"
}

func (retryableTestError) Retryable() bool {
	return true
}

func (namedTestError) Error() string {
	return "temporary designed error"
}

func (namedTestError) GoaErrorName() string {
	return "temporary"
}

func TestRetryEndpoint(t *testing.T) {
	cases := []struct {
		name                string
		err                 error
		temporaryErrorNames []string
		wantAttempts        int
	}{
		{
			name:         "temporary service error",
			err:          goa.TemporaryError("temporary", "try again"),
			wantAttempts: 2,
		},
		{
			name:         "retryable transport error",
			err:          retryableTestError{},
			wantAttempts: 2,
		},
		{
			name:                "temporary designed error",
			err:                 namedTestError{},
			temporaryErrorNames: []string{"temporary"},
			wantAttempts:        2,
		},
		{
			name:         "permanent error",
			err:          errors.New("permanent"),
			wantAttempts: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := &struct{}{}
			attempts := 0
			endpoint := goa.RetryEndpoint(func(_ context.Context, got any) (any, error) {
				assert.Same(t, request, got)
				attempts++
				if attempts == 1 {
					return nil, tc.err
				}
				return "ok", nil
			}, tc.temporaryErrorNames...)

			result, err := endpoint(context.Background(), request)

			assert.Equal(t, tc.wantAttempts, attempts)
			if tc.wantAttempts == 1 {
				assert.ErrorIs(t, err, tc.err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "ok", result)
			}
		})
	}
}

func TestRetryEndpointCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	retryErr := retryableTestError{}
	endpoint := goa.RetryEndpoint(func(context.Context, any) (any, error) {
		attempts++
		return nil, retryErr
	})

	_, err := endpoint(ctx, nil)

	assert.ErrorIs(t, err, retryErr)
	assert.Equal(t, 1, attempts)
}
