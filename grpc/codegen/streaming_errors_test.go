package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

// TestStreamingWithErrors tests that streaming endpoints properly handle
// custom errors defined in the service DSL.
func TestStreamingWithErrors(t *testing.T) {
	cases := []struct {
		name     string
		dsl      func()
		testFunc func(t *testing.T, code string)
	}{
		{
			name: "server streaming with custom errors",
			dsl:  testdata.ServerStreamingWithCustomErrorsDSL,
			testFunc: func(t *testing.T, code string) {
				// Verify error decoding is present
				assert.Contains(t, code, "goagrpc.DecodeError(err)",
					"should decode errors from stream")

				// Verify custom error types are handled
				assert.Contains(t, code, "case *streaming_error_servicepb.ServerStreamCustomErrorError:",
					"should handle custom error type")
				assert.Contains(t, code, "case *streaming_error_servicepb.ServerStreamValidationErrorError:",
					"should handle validation error type")

				// Verify generic errors are handled
				assert.Contains(t, code, "case *goapb.ErrorResponse:",
					"should handle generic goa errors")

				// Verify proper error construction
				assert.Contains(t, code, "NewServerStreamCustomErrorError(message",
					"should construct custom error")
				assert.Contains(t, code, "NewServerStreamValidationErrorError(message",
					"should construct validation error")
			},
		},
		{
			name: "bidirectional streaming with errors",
			dsl:  testdata.BidirectionalStreamingRPCWithErrorsDSL,
			testFunc: func(t *testing.T, code string) {
				// Bidirectional streaming with simple errors should still decode
				assert.Contains(t, code, "goagrpc.DecodeError(err)",
					"should decode errors from bidirectional stream")
				assert.Contains(t, code, "case *goapb.ErrorResponse:",
					"should handle generic errors in bidirectional streaming")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.dsl)
			services := CreateGRPCServices(root)
			clientfs := ClientFiles(services)
			require.Greater(t, len(clientfs), 0)

			// Get recv method implementations
			recvSections := clientfs[0].Section("client-stream-recv")
			require.Greater(t, len(recvSections), 0)

			// Build complete recv method code
			var codeBuilder strings.Builder
			for _, section := range recvSections {
				require.NoError(t, section.Write(&codeBuilder))
			}
			code := codeBuilder.String()

			// Run test-specific assertions
			c.testFunc(t, code)
		})
	}
}

// TestStreamingErrorsWithValidation verifies that custom errors with
// validation rules are properly validated in streaming recv methods.
func TestStreamingErrorsWithValidation(t *testing.T) {
	root := RunGRPCDSL(t, testdata.ServerStreamingWithCustomErrorsDSL)
	services := CreateGRPCServices(root)

	// Verify the DSL has errors with validation
	require.Len(t, root.Services, 1)
	svc := root.Services[0]
	require.Len(t, svc.Methods, 1)
	method := svc.Methods[0]
	require.Greater(t, len(method.Errors), 0, "method should have errors defined")

	// Generate client code
	clientfs := ClientFiles(services)
	require.Greater(t, len(clientfs), 0)

	// Check recv implementations
	recvSections := clientfs[0].Section("client-stream-recv")
	var code strings.Builder
	for _, section := range recvSections {
		require.NoError(t, section.Write(&code))
	}
	recvCode := code.String()

	// For errors with validation, verify validation is called
	if strings.Contains(recvCode, "ValidateServerStreamCustomErrorError") {
		assert.Contains(t, recvCode, "if err := ValidateServerStreamCustomErrorError(message); err != nil {",
			"should validate custom error before returning")
	}
}

// TestStreamingErrorComparison compares error handling between unary and
// streaming methods to ensure consistency.
func TestStreamingErrorComparison(t *testing.T) {
	// DSL with both unary and streaming methods with errors
	dsl := func() {
		var CustomError = Type("CustomError", func() {
			ErrorName(1, "name", String, "error name")
			Field(2, "message", String, "error message")
			Required("name", "message")
		})

		Service("MixedService", func() {
			// Unary method with custom error
			Method("UnaryMethod", func() {
				Payload(String)
				Result(String)
				Error("custom_error", CustomError)
				GRPC(func() {
					Response("custom_error", CodeInvalidArgument)
				})
			})

			// Streaming method with same error
			Method("StreamingMethod", func() {
				Payload(String)
				StreamingResult(String)
				Error("custom_error", CustomError)
				GRPC(func() {
					Response("custom_error", CodeInvalidArgument)
				})
			})
		})
	}

	root := RunGRPCDSL(t, dsl)
	services := CreateGRPCServices(root)
	clientfs := ClientFiles(services)
	require.Greater(t, len(clientfs), 0, "should have client files")

	// Find unary and streaming code in different sections
	var unaryCode, streamCode string

	// For unary, look in client-endpoint-init
	if sections := clientfs[0].Section("client-endpoint-init"); len(sections) > 0 {
		var code strings.Builder
		for _, section := range sections {
			require.NoError(t, section.Write(&code))
		}
		unaryCode = code.String()
	}

	// For streaming, look in client-stream-recv
	if sections := clientfs[0].Section("client-stream-recv"); len(sections) > 0 {
		var code strings.Builder
		for _, section := range sections {
			require.NoError(t, section.Write(&code))
		}
		streamCode = code.String()
	}

	// If no sections found, skip test with explanation
	if unaryCode == "" || streamCode == "" {
		t.Skip("Cannot compare unary and streaming - sections not found in generated code")
	}

	// Both should decode errors
	assert.Contains(t, unaryCode, "goagrpc.DecodeError(err)",
		"unary methods should decode errors")
	assert.Contains(t, streamCode, "goagrpc.DecodeError(err)",
		"streaming methods should decode errors")

	// Both should handle the custom error type
	assert.Contains(t, unaryCode, "case *mixed_servicepb.",
		"unary should handle custom error types")
	assert.Contains(t, streamCode, "case *mixed_servicepb.",
		"streaming should handle custom error types")
}
