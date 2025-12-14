package grpc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

// TestNewErrorResponseHistory tests that history is correctly included for merged errors
func TestNewErrorResponseHistory(t *testing.T) {
	// Test simple error - should not have history
	simpleErr := goa.MissingFieldError("username", "body")
	resp := NewErrorResponse(simpleErr)
	assert.Nil(t, resp.History)

	// Test merged error - should have history
	mergedErr := goa.MergeErrors(
		goa.MissingFieldError("username", "body"),
		goa.InvalidFormatError("data", "{invalid}", goa.FormatJSON, fmt.Errorf("invalid JSON")),
	)
	mergedResp := NewErrorResponse(mergedErr)
	assert.NotNil(t, mergedResp.History)
	assert.Len(t, mergedResp.History, 2)
	assert.Equal(t, goa.MissingField, mergedResp.History[0].Name)
	assert.Equal(t, "username", mergedResp.History[0].Field)
	assert.Equal(t, goa.InvalidFormat, mergedResp.History[1].Name)
	assert.Equal(t, "data", mergedResp.History[1].Field)
}

// TestEncodeErrorStatusCodes tests that validation errors get mapped to InvalidArgument
func TestEncodeErrorStatusCodes(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "missing_field",
			err:          goa.MissingFieldError("username", "body"),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid_format",
			err:          goa.InvalidFormatError("data", "{invalid}", goa.FormatJSON, fmt.Errorf("invalid JSON")),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "decode_payload",
			err:          &goa.ServiceError{Name: "decode_payload", Message: "failed to decode"},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "timeout",
			err:          &goa.ServiceError{Name: "timeout", Message: "timed out", Timeout: true},
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name:         "fault",
			err:          goa.Fault("internal error"),
			expectedCode: codes.Internal,
		},
		{
			name:         "temporary",
			err:          &goa.ServiceError{Name: "unavailable", Message: "service unavailable", Temporary: true},
			expectedCode: codes.Unavailable,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			encoded := EncodeError(c.err)
			st, ok := status.FromError(encoded)
			assert.True(t, ok)
			assert.Equal(t, c.expectedCode, st.Code())

			// Check that details are included
			details := st.Details()
			assert.Len(t, details, 1)
			_, ok = details[0].(*goapb.ErrorResponse)
			assert.True(t, ok)
		})
	}
}
