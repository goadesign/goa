package http

import (
	"errors"
	"io"
	nethttp "net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientError_Unwrap(t *testing.T) {
	sentinelError := errors.New("this is na error")
	alternateSentinelError := errors.New("another error")

	tests := []struct {
		name             string
		err              error
		checkedSentinel  error
		expectedCausedBy bool
	}{
		{
			name: "caused by sentinel",
			err: ErrRequestError(
				"AService",
				"Something went wrong",
				sentinelError,
			),
			checkedSentinel:  sentinelError,
			expectedCausedBy: true,
		},
		{
			name: "null cause hypothesis",
			err: ErrRequestError(
				"AService",
				"Something went wrong",
				sentinelError,
			),
			checkedSentinel:  alternateSentinelError,
			expectedCausedBy: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				isCausedBy := errors.Is(tt.err, tt.checkedSentinel)

				if isCausedBy != tt.expectedCausedBy {
					if tt.expectedCausedBy {
						t.Errorf("got error %#v, should be caused by %#v", tt.err, tt.checkedSentinel)
					} else {
						t.Errorf("got error %#v, must NOT be caused by %#v", tt.err, tt.checkedSentinel)
					}
				}
			},
		)
	}
}

func TestClientErrorRetryable(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "request EOF",
			err:       ErrRequestError("service", "method", io.EOF),
			retryable: true,
		},
		{
			name:      "truncated response",
			err:       ErrDecodingError("service", "method", io.ErrUnexpectedEOF),
			retryable: true,
		},
		{
			name:      "bad gateway",
			err:       ErrInvalidResponse("service", "method", nethttp.StatusBadGateway, ""),
			retryable: true,
		},
		{
			name:      "bad request",
			err:       ErrInvalidResponse("service", "method", nethttp.StatusBadRequest, ""),
			retryable: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var clientErr *ClientError
			if !assert.ErrorAs(t, tc.err, &clientErr) {
				return
			}
			assert.Equal(t, tc.retryable, clientErr.Retryable())
		})
	}
}
