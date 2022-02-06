package expr

import "testing"

func TestAsObject(t *testing.T) {
	var (
		object = &Object{
			&NamedAttributeExpr{
				Name: "foo",
				Attribute: &AttributeExpr{
					Type: String,
				},
			},
		}
		objectUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: object,
			},
		}
		notObjectUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		objectResultType = &ResultTypeExpr{
			UserTypeExpr: objectUserType,
		}
		notObjectResultType = &ResultTypeExpr{
			UserTypeExpr: notObjectUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected *Object
	}{
		"object user type": {
			dt:       objectUserType,
			expected: object,
		},
		"not object user type": {
			dt:       notObjectUserType,
			expected: nil,
		},
		"object result type": {
			dt:       objectResultType,
			expected: object,
		},
		"not object result type": {
			dt:       notObjectResultType,
			expected: nil,
		},
		"object": {
			dt:       object,
			expected: object,
		},
		"not object": {
			dt:       Boolean,
			expected: nil,
		},
	}

	for k, tc := range cases {
		if actual := AsObject(tc.dt); actual != tc.expected {
			if Equal(actual, tc.expected) != true {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestAsArray(t *testing.T) {
	var (
		array = &Array{
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		arrayUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: array,
			},
		}
		notArrayUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		arrayResultType = &ResultTypeExpr{
			UserTypeExpr: arrayUserType,
		}
		notArrayResultType = &ResultTypeExpr{
			UserTypeExpr: notArrayUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected *Array
	}{
		"array user type": {
			dt:       arrayUserType,
			expected: array,
		},
		"not array user type": {
			dt:       notArrayUserType,
			expected: nil,
		},
		"array result type": {
			dt:       arrayResultType,
			expected: array,
		},
		"not array result type": {
			dt:       notArrayResultType,
			expected: nil,
		},
		"array": {
			dt:       array,
			expected: array,
		},
		"not array": {
			dt:       Boolean,
			expected: nil,
		},
	}

	for k, tc := range cases {
		if actual := AsArray(tc.dt); actual != tc.expected {
			if Equal(actual, tc.expected) != true {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestAsMap(t *testing.T) {
	var (
		mapIntString = &Map{
			KeyType: &AttributeExpr{
				Type: Int,
			},
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		mapUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: mapIntString,
			},
		}
		notMapUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		mapResultType = &ResultTypeExpr{
			UserTypeExpr: mapUserType,
		}
		notMapResultType = &ResultTypeExpr{
			UserTypeExpr: notMapUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected *Map
	}{
		"map user type": {
			dt:       mapUserType,
			expected: mapIntString,
		},
		"not map user type": {
			dt:       notMapUserType,
			expected: nil,
		},
		"map result type": {
			dt:       mapResultType,
			expected: mapIntString,
		},
		"not map result type": {
			dt:       notMapResultType,
			expected: nil,
		},
		"map": {
			dt:       mapIntString,
			expected: mapIntString,
		},
		"not map": {
			dt:       Boolean,
			expected: nil,
		},
	}

	for k, tc := range cases {
		if actual := AsMap(tc.dt); actual != tc.expected {
			if Equal(actual, tc.expected) != true {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestIsObject(t *testing.T) {
	var (
		object = &Object{
			&NamedAttributeExpr{
				Name: "foo",
				Attribute: &AttributeExpr{
					Type: String,
				},
			},
		}
		objectUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: object,
			},
		}
		notObjectUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		objectResultType = &ResultTypeExpr{
			UserTypeExpr: objectUserType,
		}
		notObjectResultType = &ResultTypeExpr{
			UserTypeExpr: notObjectUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"object user type": {
			dt:       objectUserType,
			expected: true,
		},
		"not object user type": {
			dt:       notObjectUserType,
			expected: false,
		},
		"object result type": {
			dt:       objectResultType,
			expected: true,
		},
		"not object result type": {
			dt:       notObjectResultType,
			expected: false,
		},
		"object": {
			dt:       object,
			expected: true,
		},
		"not object": {
			dt:       Boolean,
			expected: false,
		},
	}

	for k, tc := range cases {
		if actual := IsObject(tc.dt); actual != tc.expected {
			if actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestIsArray(t *testing.T) {
	var (
		array = &Array{
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		arrayUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: array,
			},
		}
		notArrayUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		arrayResultType = &ResultTypeExpr{
			UserTypeExpr: arrayUserType,
		}
		notArrayResultType = &ResultTypeExpr{
			UserTypeExpr: notArrayUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"array user type": {
			dt:       arrayUserType,
			expected: true,
		},
		"not array user type": {
			dt:       notArrayUserType,
			expected: false,
		},
		"array result type": {
			dt:       arrayResultType,
			expected: true,
		},
		"not array result type": {
			dt:       notArrayResultType,
			expected: false,
		},
		"array": {
			dt:       array,
			expected: true,
		},
		"not array": {
			dt:       Boolean,
			expected: false,
		},
	}

	for k, tc := range cases {
		if actual := IsArray(tc.dt); actual != tc.expected {
			if actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestIsMap(t *testing.T) {
	var (
		mapIntString = &Map{
			KeyType: &AttributeExpr{
				Type: Int,
			},
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		mapUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: mapIntString,
			},
		}
		notMapUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		mapResultType = &ResultTypeExpr{
			UserTypeExpr: mapUserType,
		}
		notMapResultType = &ResultTypeExpr{
			UserTypeExpr: notMapUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"map user type": {
			dt:       mapUserType,
			expected: true,
		},
		"not map user type": {
			dt:       notMapUserType,
			expected: false,
		},
		"map result type": {
			dt:       mapResultType,
			expected: true,
		},
		"not map result type": {
			dt:       notMapResultType,
			expected: false,
		},
		"map": {
			dt:       mapIntString,
			expected: true,
		},
		"not map": {
			dt:       Boolean,
			expected: false,
		},
	}

	for k, tc := range cases {
		if actual := IsMap(tc.dt); actual != tc.expected {
			if actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

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

func TestIsAlias(t *testing.T) {
	var (
		aliasUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: String,
			},
		}
		notAliasUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: &Object{},
			},
		}
		aliasResultType = &ResultTypeExpr{
			UserTypeExpr: aliasUserType,
		}
		notAliasResultType = &ResultTypeExpr{
			UserTypeExpr: notAliasUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"alias user type": {
			dt:       aliasUserType,
			expected: true,
		},
		"not alias user type": {
			dt:       notAliasUserType,
			expected: false,
		},
		"alias result type": {
			dt:       aliasResultType,
			expected: true,
		},
		"not alias result type": {
			dt:       notAliasResultType,
			expected: false,
		},
	}
	for k, tc := range cases {
		if actual := IsAlias(tc.dt); tc.expected != actual {
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
		ss   = []string{"foo", "bar"}
		is   = []int{1, 2}
	)
	cases := map[string]struct {
		p        Primitive
		values   []interface{}
		expected bool
	}{
		"any": {
			p:        Any,
			values:   []interface{}{b},
			expected: true,
		},
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
		"not supported types": {
			p:        Boolean,
			values:   []interface{}{ss, is},
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

func TestArrayIsCompatible(t *testing.T) {
	var (
		b  = true
		i  = 1
		ia = [2]int{1, 2}
		is = []int{3, 4}
	)
	cases := map[string]struct {
		typ      DataType
		values   []interface{}
		expected bool
	}{
		"compatible": {
			typ:      Int,
			values:   []interface{}{ia, is},
			expected: true,
		},
		"not array and slice": {
			typ:      String,
			values:   []interface{}{b, i},
			expected: false,
		},
		"array but not compatible": {
			typ:      String,
			values:   []interface{}{ia},
			expected: false,
		},
		"slice but not compatible": {
			typ:      String,
			values:   []interface{}{is},
			expected: false,
		},
	}

	for k, tc := range cases {
		array := Array{
			ElemType: &AttributeExpr{
				Type: tc.typ,
			},
		}
		for _, value := range tc.values {
			if actual := array.IsCompatible(value); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestObjectRename(t *testing.T) {
	cases := map[string]struct {
		old, new string
		expected []string
	}{
		"renamed": {
			old:      "foo",
			new:      "qux",
			expected: []string{"qux", "bar"},
		},
		"unmatched": {
			old:      "baz",
			new:      "qux",
			expected: []string{"foo", "bar"},
		},
	}

	for k, tc := range cases {
		object := &Object{
			&NamedAttributeExpr{
				Name: "foo",
				Attribute: &AttributeExpr{
					Type: String,
				},
			},
			&NamedAttributeExpr{
				Name: "bar",
				Attribute: &AttributeExpr{
					Type: String,
				},
			},
		}
		object.Rename(tc.old, tc.new)
		for _, s := range tc.expected {
			if att := object.Attribute(s); att == nil {
				t.Errorf("%s: %s not found", k, s)
			}
		}
	}
}

func TestObjectIsCompatible(t *testing.T) {
	var (
		b = true
		i = 1
		s = struct {
			Foo string
		}{
			Foo: "foo",
		}
		m = map[int]string{}
	)
	cases := map[string]struct {
		values   []interface{}
		expected bool
	}{
		"compatible": {
			values:   []interface{}{s, m},
			expected: true,
		},
		"not comatible": {
			values:   []interface{}{b, i},
			expected: false,
		},
	}

	object := Object{}
	for k, tc := range cases {
		for _, value := range tc.values {
			if actual := object.IsCompatible(value); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestMapIsCompatible(t *testing.T) {
	var (
		b   = true
		i   = 1
		ism = map[int]string{
			1: "foo",
		}
		ssm = map[string]string{
			"bar": "bar",
		}
		iim = map[int]int{
			2: 2,
		}
	)
	cases := map[string]struct {
		values   []interface{}
		expected bool
	}{
		"compatible": {
			values:   []interface{}{ism},
			expected: true,
		},
		"not comatible": {
			values:   []interface{}{b, i},
			expected: false,
		},
		"map but not comatible": {
			values:   []interface{}{ssm, iim},
			expected: false,
		},
	}

	m := Map{
		KeyType: &AttributeExpr{
			Type: Int,
		},
		ElemType: &AttributeExpr{
			Type: String,
		},
	}
	for k, tc := range cases {
		for _, value := range tc.values {
			if actual := m.IsCompatible(value); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}

func TestQualifiedTypeName(t *testing.T) {
	var (
		array = &Array{
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		mapStringString = &Map{
			KeyType: &AttributeExpr{
				Type: String,
			},
			ElemType: &AttributeExpr{
				Type: String,
			},
		}
		mapStringArray = &Map{
			KeyType: &AttributeExpr{
				Type: String,
			},
			ElemType: &AttributeExpr{
				Type: array,
			},
		}
		mapStringMap = &Map{
			KeyType: &AttributeExpr{
				Type: String,
			},
			ElemType: &AttributeExpr{
				Type: mapStringString,
			},
		}
	)
	cases := map[string]struct {
		t        DataType
		expected string
	}{
		"boolean": {
			t:        Boolean,
			expected: "boolean",
		},
		"int": {
			t:        Int,
			expected: "int",
		},
		"int32": {
			t:        Int32,
			expected: "int32",
		},
		"int64": {
			t:        Int64,
			expected: "int64",
		},
		"uint": {
			t:        UInt,
			expected: "uint",
		},
		"uint32": {
			t:        UInt32,
			expected: "uint32",
		},
		"uint64": {
			t:        UInt64,
			expected: "uint64",
		},
		"float32": {
			t:        Float32,
			expected: "float32",
		},
		"float64": {
			t:        Float64,
			expected: "float64",
		},
		"string": {
			t:        String,
			expected: "string",
		},
		"bytes": {
			t:        Bytes,
			expected: "bytes",
		},
		"any": {
			t:        Any,
			expected: "any",
		},
		"user type": {
			t: &UserTypeExpr{
				TypeName: "userType",
			},
			expected: "userType",
		},
		"result type": {
			t: &ResultTypeExpr{
				UserTypeExpr: &UserTypeExpr{
					TypeName: "resultType",
				},
			},
			expected: "resultType",
		},
		"object": {
			t:        &Object{},
			expected: "object",
		},
		"array": {
			t:        array,
			expected: "array<string>",
		},
		"map": {
			t:        mapStringString,
			expected: "map<string, string>",
		},
		"map contains array": {
			t:        mapStringArray,
			expected: "map<string, array<string>>",
		},
		"map contains map": {
			t:        mapStringMap,
			expected: "map<string, map<string, string>>",
		},
	}

	for k, tc := range cases {
		if actual := QualifiedTypeName(tc.t); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
