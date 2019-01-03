package design

import . "goa.design/goa/dsl"

var _ = API("divider", func() {
	Title("Divider Service")
	Description("An example illustrating error handling in goa. See docs/ErrorHandling.md.")
})

var _ = Service("divider", func() {

	// The "div_by_zero" error is defined at the service level and
	// thus may be returned by both "divide" and "integer_divide".
	Error("div_by_zero", ErrorResult, "divizion by zero")

	// The "timeout" error is also defined at the service level.
	Error("timeout", ErrorResult, "operation timed out, retry later.", func() {
		// Timeout indicates an error due to a timeout.
		Timeout()
		// Temporary indicates that the request may be retried.
		Temporary()
	})

	HTTP(func() {
		// Use HTTP status code 400 Bad Request for "div_by_zero"
		// errors.
		Response("div_by_zero", StatusBadRequest)

		// Use HTTP status code 504 Gateway Timeout for "timeout"
		// errors.
		Response("timeout", StatusGatewayTimeout)
	})

	GRPC(func() {
		// Use gRPC status code "InvalidArgument" for "div_by_zero"
		// errors.
		Response("div_by_zero", CodeInvalidArgument)

		// Use gRPC status code "DeadlineExceeded" for "timeout"
		// errors.
		Response("timeout", CodeDeadlineExceeded)
	})

	Method("integer_divide", func() {
		Payload(IntOperands)
		Result(Int)

		// The "has_remainder" error is defined at the method
		// level and is thus specific to "integer_divide".
		Error("has_remainder", ErrorResult, "integer division has remainder")

		HTTP(func() {
			GET("/idiv/{a}/{b}")
			Response(StatusOK)
			Response("has_remainder", StatusExpectationFailed)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("has_remainder", CodeUnknown)
		})
	})

	Method("divide", func() {
		Payload(FloatOperands)
		Result(Float64)

		HTTP(func() {
			GET("/div/{a}/{b}")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})
})

var IntOperands = Type("IntOperands", func() {
	Field(1, "a", Int, "Left operand")
	Field(2, "b", Int, "Right operand")
	Required("a", "b")
})

var FloatOperands = Type("FloatOperands", func() {
	Field(1, "a", Float64, "Left operand")
	Field(2, "b", Float64, "Right operand")
	Required("a", "b")
})
