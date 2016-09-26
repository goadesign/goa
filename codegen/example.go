package codegen

import (
	"sort"

	"github.com/goadesign/goa/design"
)

// GenerateExample produces a random value of the given type.
func GenerateExample(dt design.DataType, r *RandomGenerator, seen []string) interface{} {
}

var anyPrimitive = []Primitive{Boolean, Int32, Int64, UInt32, UInt64, Float32, Float64, String, Bytes}

// generatePrimitive produces a random value of the given primitive type.
func generatePrimitive(p design.Primitive, r *RandomGenerator, seen []string) interface{} {
	switch p {
	case Boolean:
		return r.Bool()
	case Int32, UInt32:
		return r.Int32()
	case Int64, UInt64:
		return r.Int64()
	case Float32:
		return r.Float32()
	case Float64:
		return r.Float64()
	case String, Bytes:
		return r.String()
	case Any:
		// to not make it too complicated, pick one of the primitive types
		return anyPrimitive[r.Int()%len(anyPrimitive)].GenerateExample(r, seen)
	default:
		panic("unknown primitive type") // bug
	}
}

// generateArray produces a random array value.
func generateArray(a *design.Array, r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = a.ElemType.Type.GenerateExample(r, seen)
	}
	return a.MakeSlice(res)
}

// generateObject returns a random value of the object.
func generateObject(o Object, r *RandomGenerator, seen []string) interface{} {
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

// generateMap returns a random map value.
func generateMap(m *design.Map, r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	pair := map[interface{}]interface{}{}
	for i := 0; i < count; i++ {
		pair[m.KeyType.Type.GenerateExample(r, seen)] = m.ElemType.Type.GenerateExample(r, seen)
	}
	return m.MakeMap(pair)
}
