package http

import (
	"errors"
	"testing"
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
