package dsl

import "goa.design/goa/v3/expr"

const (
	// Boolean is the type for a JSON boolean.
	Boolean = expr.Boolean

	// Int is the type for a signed integer.
	Int = expr.Int

	// Int32 is the type for a signed 32-bit integer.
	Int32 = expr.Int32

	// Int64 is the type for a signed 64-bit integer.
	Int64 = expr.Int64

	// UInt is the type for an unsigned integer.
	UInt = expr.UInt

	// UInt32 is the type for an unsigned 32-bit integer.
	UInt32 = expr.UInt32

	// UInt64 is the type for an unsigned 64-bit integer.
	UInt64 = expr.UInt64

	// Float32 is the type for a 32-bit floating number.
	Float32 = expr.Float32

	// Float64 is the type for a 64-bit floating number.
	Float64 = expr.Float64

	// String is the type for a JSON string.
	String = expr.String

	// Bytes is the type for binary data.
	Bytes = expr.Bytes

	// Any is the type for an arbitrary JSON value (interface{} in Go).
	Any = expr.Any
)

// Empty represents empty values.
var Empty = expr.Empty
