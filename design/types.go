// Package design defines types which are used to describe the data structures used by both the
// request and response messages used by services.
//
// There are primitive types corresponding to scalar values (bool, string, integers and numbers),
// array types which represent a collection of items, map types which represent maps of key/value
// pairs and object types describing data structures with fields.
//
// The package also defines user types which are named types and media types which describe HTTP
// media types.
package design

import (
	"reflect"
	"sort"
)

type (
	// A Kind defines the conceptual type that a DataType represents.
	Kind uint

	// DataType is the common interface to all types.
	DataType interface {
		// Kind of data type, one of the Kind enum.
		Kind() Kind
		// Name returns the type name.
		Name() string
		// IsPrimitive returns true if the underlying type is one of the primitive types.
		IsPrimitive() bool
		// IsArray returns true if the underlying type is an array, a user type which is an
		// array or a media type whose type is an array.
		IsArray() bool
		// IsMap returns true if the underlying type is a hash map, a user type which is a
		// hash map or a media type whose type is a hash map.
		IsMap() bool
		// IsObject returns true if the underlying type is an object, a user type which is
		// an object or a media type whose type is an object.
		IsObject() bool
		// ToObject returns the underlying object if any (i.e. if IsObject returns true),
		// nil otherwise.
		ToObject() Object
		// ToArray returns the underlying array if any (i.e. if IsArray returns true), nil
		// otherwise.
		ToArray() *Array
		// ToMap returns the underlying map if any (i.e. if IsMap returns true), nil
		// otherwise.
		ToMap() *Map
		// IsCompatible checks whether val has a Go type that is compatible with the data
		// type.
		IsCompatible(val interface{}) bool
	}

	// DataStructure is the interface implemented by all data structure types.
	// That is field expressions, user types and media types.
	DataStructure interface {
		// Expr returns the data structure expression.
		Expr() *AttributeExpr
		// Walk traverses the data structure recursively and calls the given function once
		// on each field starting with the field returned by Expr.
		// User type and media type fields are traversed once even for recursive
		// expressions to avoid infinite recursion.
		// Walk stops and returns the error if the function returns a non-nil error.
		Walk(func(*AttributeExpr) error) error
	}

	// Primitive is the type for null, boolean, integer, number, string, and time.
	Primitive Kind

	// Array is the type used to describe field arrays or repeated fields.
	Array struct {
		ElemType *AttributeExpr
	}

	// Map is the type used to describe maps of fields.
	Map struct {
		KeyType  *AttributeExpr
		ElemType *AttributeExpr
	}

	// Object is the type used to describe composite data structures.
	Object map[string]*AttributeExpr

	// UserTypeExpr is the struct used to describe user defined types.
	UserTypeExpr struct {
		// A user type expression is a field expression.
		*AttributeExpr
		// Name of type
		TypeName string
	}

	// ArrayVal is the type used to set the default value for arrays.
	ArrayVal []interface{}

	// MapVal is the type used to set the default value for maps.
	MapVal map[interface{}]interface{}
)

const (
	// BooleanKind represents a boolean.
	BooleanKind Kind = iota + 1
	// Int32Kind represents a 32-bit integer.
	Int32Kind
	// Int64Kind represents a 64-bit integer.
	Int64Kind
	// Float32Kind represents a 32-bit floating number.
	Float32Kind
	// Float64Kind represents a 64-bit floating number.
	Float64Kind
	// StringKind represents a JSON string.
	StringKind
	// ArrayKind represents a JSON array.
	ArrayKind
	// ObjectKind represents a JSON object.
	ObjectKind
	// MapKind represents a JSON object where the keys are not known in advance.
	MapKind
	// UserTypeKind represents a user type.
	UserTypeKind
	// MediaTypeKind represents a media type.
	MediaTypeKind
	// AnyKind represents a unknown type.
	AnyKind
)

const (
	// Boolean is the type for a JSON boolean.
	Boolean = Primitive(BooleanKind)

	// Int32 is the type for a 32-bit integer.
	Int32 = Primitive(Int32Kind)

	// Int64 is the type for a 64-bit integer.
	Int64 = Primitive(Int64Kind)

	// Float32 is the type for a 32-bit floating number.
	Float32 = Primitive(Float32Kind)

	// Float64 is the type for a 64-bit floating number.
	Float64 = Primitive(Float64Kind)

	// String is the type for a JSON string.
	String = Primitive(StringKind)

	// Any is the type for an arbitrary JSON value (interface{} in Go).
	Any = Primitive(AnyKind)
)

// DataType implementation

// Kind implements DataKind.
func (p Primitive) Kind() Kind { return Kind(p) }

// Name returns the type name appropriate for logging.
func (p Primitive) Name() string {
	switch p {
	case Boolean:
		return "boolean"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case String:
		return "string"
	case Any:
		return "any"
	default:
		panic("unknown primitive type") // bug
	}
}

// IsPrimitive returns true.
func (p Primitive) IsPrimitive() bool { return true }

// IsObject returns false.
func (p Primitive) IsObject() bool { return false }

// IsArray returns false.
func (p Primitive) IsArray() bool { return false }

// IsMap returns false.
func (p Primitive) IsMap() bool { return false }

// ToObject returns nil.
func (p Primitive) ToObject() Object { return nil }

// ToArray returns nil.
func (p Primitive) ToArray() *Array { return nil }

// ToMap returns nil.
func (p Primitive) ToMap() *Map { return nil }

// IsCompatible returns true if val is compatible with p.
func (p Primitive) IsCompatible(val interface{}) bool {
	if p == Any {
		return true
	}
	switch val.(type) {
	case bool:
		return p == Boolean
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return p == Int32 || p == Int64 || p == Float32 || p == Float64
	case float32, float64:
		return p == Float32 || p == Float64
	case string:
		return p == String
	}
	return false
}

var anyPrimitive = []Primitive{Boolean, Int32, Int64, Float32, Float64, String}

// GenerateExample returns an instance of the given data type.
func (p Primitive) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	switch p {
	case Boolean:
		return r.Bool()
	case Int32:
		return r.Int32()
	case Int64:
		return r.Int64()
	case Float32:
		return r.Float32()
	case Float64:
		return r.Float64()
	case String:
		return r.String()
	case Any:
		// to not make it too complicated, pick one of the primitive types
		return anyPrimitive[r.Int()%len(anyPrimitive)].GenerateExample(r, seen)
	default:
		panic("unknown primitive type") // bug
	}
}

// Kind implements DataKind.
func (a *Array) Kind() Kind { return ArrayKind }

// Name returns the type name.
func (a *Array) Name() string {
	return "array"
}

// IsPrimitive returns false.
func (a *Array) IsPrimitive() bool { return false }

// IsObject returns false.
func (a *Array) IsObject() bool { return false }

// IsArray returns true.
func (a *Array) IsArray() bool { return true }

// IsMap returns false.
func (a *Array) IsMap() bool { return false }

// ToObject returns nil.
func (a *Array) ToObject() Object { return nil }

// ToArray returns a.
func (a *Array) ToArray() *Array { return a }

// ToMap returns nil.
func (a *Array) ToMap() *Map { return nil }

// IsCompatible returns true if val is compatible with p.
func (a *Array) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	if k != reflect.Array && k != reflect.Slice {
		return false
	}
	v := reflect.ValueOf(val)
	for i := 0; i < v.Len(); i++ {
		compat := (a.ElemType.Type != nil) && a.ElemType.Type.IsCompatible(v.Index(i).Interface())
		if !compat {
			return false
		}
	}
	return true
}

// GenerateExample produces a random array value.
func (a *Array) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = a.ElemType.Type.GenerateExample(r, seen)
	}
	return a.MakeSlice(res)
}

// MakeSlice examines the key type from the Array and create a slice with builtin type if possible.
// The idea is to avoid generating []interface{} and produce more known types.
func (a *Array) MakeSlice(s []interface{}) interface{} {
	slice := reflect.MakeSlice(toReflectType(a), 0, len(s))
	for _, item := range s {
		slice = reflect.Append(slice, reflect.ValueOf(item))
	}
	return slice.Interface()
}

// ToSlice converts an ArrayVal into a slice.
func (a ArrayVal) ToSlice() []interface{} {
	arr := make([]interface{}, len(a))
	for i, elem := range a {
		switch actual := elem.(type) {
		case ArrayVal:
			arr[i] = actual.ToSlice()
		case MapVal:
			arr[i] = actual.ToMap()
		default:
			arr[i] = actual
		}
	}
	return arr
}

// Kind implements DataKind.
func (o Object) Kind() Kind { return ObjectKind }

// Name returns the type name.
func (o Object) Name() string { return "object" }

// IsPrimitive returns false.
func (o Object) IsPrimitive() bool { return false }

// IsObject returns true.
func (o Object) IsObject() bool { return true }

// IsArray returns false.
func (o Object) IsArray() bool { return false }

// IsMap returns false.
func (o Object) IsMap() bool { return false }

// ToObject returns the underlying object.
func (o Object) ToObject() Object { return o }

// ToArray returns nil.
func (o Object) ToArray() *Array { return nil }

// ToMap returns nil.
func (o Object) ToMap() *Map { return nil }

// Merge copies other's fields into o overridding any pre-existing field with the same name.
func (o Object) Merge(other Object) {
	for n, att := range other {
		o[n] = DupAtt(att)
	}
}

// IsCompatible returns true if o describes the (Go) type of val.
func (o Object) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	return k == reflect.Map || k == reflect.Struct
}

// GenerateExample returns a random value of the object.
func (o Object) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	// ensure fixed ordering
	keys := make([]string, 0, len(o))
	for n := range o {
		keys = append(keys, n)
	}
	sort.Strings(keys)

	res := make(map[string]interface{})
	for _, n := range keys {
		att := o[n]
		res[n] = att.Type.GenerateExample(r, seen)
	}
	return res
}

// Kind implements DataKind.
func (h *Map) Kind() Kind { return MapKind }

// Name returns the type name.
func (h *Map) Name() string { return "hash" }

// IsPrimitive returns false.
func (h *Map) IsPrimitive() bool { return false }

// IsObject returns false.
func (h *Map) IsObject() bool { return false }

// IsArray returns false.
func (h *Map) IsArray() bool { return false }

// IsMap returns true.
func (h *Map) IsMap() bool { return true }

// ToObject returns nil.
func (h *Map) ToObject() Object { return nil }

// ToArray returns nil.
func (h *Map) ToArray() *Array { return nil }

// ToMap returns the underlying hash map.
func (h *Map) ToMap() *Map { return h }

// IsCompatible returns true if o describes the (Go) type of val.
func (h *Map) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	if k != reflect.Map {
		return false
	}
	v := reflect.ValueOf(val)
	for _, key := range v.MapKeys() {
		keyCompat := h.KeyType.Type == nil || h.KeyType.Type.IsCompatible(key.Interface())
		elemCompat := h.ElemType.Type == nil || h.ElemType.Type.IsCompatible(v.MapIndex(key).Interface())
		if !keyCompat || !elemCompat {
			return false
		}
	}
	return true
}

// GenerateExample returns a random hash value.
func (h *Map) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	pair := map[interface{}]interface{}{}
	for i := 0; i < count; i++ {
		pair[h.KeyType.Type.GenerateExample(r, seen)] = h.ElemType.Type.GenerateExample(r, seen)
	}
	return h.MakeMap(pair)
}

// MakeMap examines the key type from a Map and create a map with builtin type if possible.
// The idea is to avoid generating map[interface{}]interface{}, which cannot be handled by json.Marshal.
func (h *Map) MakeMap(m map[interface{}]interface{}) interface{} {
	hash := reflect.MakeMap(toReflectType(h))
	for key, value := range m {
		hash.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
	return hash.Interface()
}

// ToMap converts a MapVal to a map.
func (h MapVal) ToMap() map[interface{}]interface{} {
	mp := make(map[interface{}]interface{}, len(h))
	for k, v := range h {
		switch actual := v.(type) {
		case ArrayVal:
			mp[k] = actual.ToSlice()
		case MapVal:
			mp[k] = actual.ToMap()
		default:
			mp[k] = actual
		}
	}
	return mp
}

// AttributeIterator is the type of the function given to IterateAttributes.
type AttributeIterator func(string, *AttributeExpr) error

// IterateAttributes calls the given iterator passing in each field sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateObject returns that
// error.
func (o Object) IterateAttributes(it AttributeIterator) error {
	names := make([]string, len(o))
	i := 0
	for n := range o {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(n, o[n]); err != nil {
			return err
		}
	}
	return nil
}

// toReflectType converts the DataType to reflect.Type.
func toReflectType(dtype DataType) reflect.Type {
	switch dtype.Kind() {
	case BooleanKind:
		return reflect.TypeOf(true)
	case Int32Kind:
		return reflect.TypeOf(int32(0))
	case Int64Kind:
		return reflect.TypeOf(int64(0))
	case Float32Kind:
		return reflect.TypeOf(float32(0))
	case Float64Kind:
		return reflect.TypeOf(float64(0))
	case StringKind:
		return reflect.TypeOf("")
	case ObjectKind, UserTypeKind, MediaTypeKind:
		return reflect.TypeOf(map[string]interface{}{})
	case ArrayKind:
		return reflect.SliceOf(toReflectType(dtype.ToArray().ElemType.Type))
	case MapKind:
		hash := dtype.ToMap()
		// avoid complication: not allow object as the hash key
		var ktype reflect.Type
		if !hash.KeyType.Type.IsObject() {
			ktype = toReflectType(hash.KeyType.Type)
		} else {
			ktype = reflect.TypeOf([]interface{}{}).Elem()
		}
		return reflect.MapOf(ktype, toReflectType(hash.ElemType.Type))
	default:
		return reflect.TypeOf([]interface{}{}).Elem()
	}
}
