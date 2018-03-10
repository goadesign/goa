package design

import "testing"

func TestIsPrimitive(t *testing.T) {
	var (
		primitiveUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		notPrimitiveUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: &Object{},
			},
		}
		primitiveResultType = &ResultTypeExpr{
			UserTypeExpr: primitiveUserType,
		}
		notPrimitiveResultType = &ResultTypeExpr{
			UserTypeExpr: notPrimitiveUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"boolean": {
			dt:       Boolean,
			expected: true,
		},
		"int": {
			dt:       Int,
			expected: true,
		},
		"int32": {
			dt:       Int32,
			expected: true,
		},
		"int64": {
			dt:       Int64,
			expected: true,
		},
		"uint": {
			dt:       UInt,
			expected: true,
		},
		"uint32": {
			dt:       UInt32,
			expected: true,
		},
		"uint64": {
			dt:       UInt64,
			expected: true,
		},
		"float32": {
			dt:       Float32,
			expected: true,
		},
		"float64": {
			dt:       Float64,
			expected: true,
		},
		"string": {
			dt:       String,
			expected: true,
		},
		"bytes": {
			dt:       Bytes,
			expected: true,
		},
		"any": {
			dt:       Any,
			expected: true,
		},
		"primitive user type": {
			dt:       primitiveUserType,
			expected: true,
		},
		"not primitive user type": {
			dt:       notPrimitiveUserType,
			expected: false,
		},
		"primitive result type": {
			dt:       primitiveResultType,
			expected: true,
		},
		"not primitive result type": {
			dt:       notPrimitiveResultType,
			expected: false,
		},
		"object": {
			dt:       &Object{},
			expected: false,
		},
		"array": {
			dt:       &Array{},
			expected: false,
		},
		"map": {
			dt:       &Map{},
			expected: false,
		},
	}

	for k, tc := range cases {
		if actual := IsPrimitive(tc.dt); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestPrimitiveIsCompatible(t *testing.T) {
	var (
		b    = bool(true)
		i    = int(1)
		i8   = int8(2)
		i16  = int16(3)
		i32  = int32(4)
		ui   = uint(5)
		ui8  = uint8(6)
		ui16 = uint16(7)
		ui32 = uint32(8)
		i64  = int64(9)
		ui64 = uint64(10)
		f32  = float32(10.1)
		f64  = float64(20.2)
		s    = string("string")
		bs   = []byte("bytes")
	)
	cases := map[string]struct {
		p        Primitive
		values   []interface{}
		expected bool
	}{
		"boolean compatible": {
			p:        Boolean,
			values:   []interface{}{b},
			expected: true,
		},
		"boolean not compatible": {
			p:        Boolean,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64, f32, f64, s, bs},
			expected: false,
		},
		"int compatible": {
			p:        Int,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32},
			expected: true,
		},
		"int not compatible": {
			p:        Int,
			values:   []interface{}{b, i64, ui64, f32, f64, s, bs},
			expected: false,
		},
		"int32 compatible": {
			p:        Int32,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32},
			expected: true,
		},
		"int32 not compatible": {
			p:        Int32,
			values:   []interface{}{b, i64, ui64, f32, f64, s, bs},
			expected: false,
		},
		"int64 compatible": {
			p:        Int64,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64},
			expected: true,
		},
		"int64 not compatible": {
			p:        Int64,
			values:   []interface{}{b, f32, f64, s, bs},
			expected: false,
		},
		"uint compatible": {
			p:        UInt,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32},
			expected: true,
		},
		"uint not compatible": {
			p:        UInt,
			values:   []interface{}{b, i64, ui64, f32, f64, s, bs},
			expected: false,
		},
		"uint32 compatible": {
			p:        UInt32,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32},
			expected: true,
		},
		"uint32 not compatible": {
			p:        UInt32,
			values:   []interface{}{b, i64, ui64, f32, f64, s, bs},
			expected: false,
		},
		"uint64 compatible": {
			p:        UInt64,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64},
			expected: true,
		},
		"uint64 not compatible": {
			p:        UInt64,
			values:   []interface{}{b, f32, f64, s, bs},
			expected: false,
		},
		"float32 compatible": {
			p:        Float32,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64, f32, f64},
			expected: true,
		},
		"float32 not compatible": {
			p:        Float32,
			values:   []interface{}{b, s, bs},
			expected: false,
		},
		"float64 compatible": {
			p:        Float64,
			values:   []interface{}{i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64, f32, f64},
			expected: true,
		},
		"float64 not compatible": {
			p:        Float64,
			values:   []interface{}{b, s, bs},
			expected: false,
		},
		"string compatible": {
			p:        String,
			values:   []interface{}{s},
			expected: true,
		},
		"string not compatible": {
			p:        String,
			values:   []interface{}{b, i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64, f32, f64, bs},
			expected: false,
		},
		"bytes compatible": {
			p:        Bytes,
			values:   []interface{}{s, bs},
			expected: true,
		},
		"bytes not compatible": {
			p:        Bytes,
			values:   []interface{}{b, i, i8, i16, i32, ui, ui8, ui16, ui32, i64, ui64, f32, f64},
			expected: false,
		},
	}

	for k, tc := range cases {
		for _, value := range tc.values {
			if actual := tc.p.IsCompatible(value); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}
