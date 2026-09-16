// This file copies the values used to write generated files. A caller may
// change the copy without changing the HTTP files saved for the same service.
package codegen

import (
	"fmt"
	"reflect"

	"goa.design/goa/v3/codegen"
)

var immutableRenderPointers = map[reflect.Type]struct{}{
	reflect.TypeFor[*codegen.NameDeclaration]():        {},
	reflect.TypeFor[*codegen.TypeDeclaration]():        {},
	reflect.TypeFor[*codegen.UnionDeclaration]():       {},
	reflect.TypeFor[*codegen.UnionBranchDeclaration](): {},
	reflect.TypeFor[*codegen.Location]():               {},
	reflect.TypeFor[*codegen.GoTypePlan]():             {},
	reflect.TypeFor[*wireTypeRecord]():                 {},
	reflect.TypeFor[*wireUnionRecord]():                {},
}

// cloneRenderData copies maps, slices, pointers, and values stored in an
// interface. Generated name and type records are shared because they cannot be
// changed after their names are assigned.
func cloneRenderData(data any) any {
	if data == nil {
		return nil
	}
	return cloneRenderValue(reflect.ValueOf(data), make(map[clonePointer]reflect.Value)).Interface()
}

type clonePointer struct {
	typeOf  reflect.Type
	pointer uintptr
}

// cloneRenderValue remembers pointers it has already copied. This preserves
// repeated references and lets it copy values that refer back to themselves.
func cloneRenderValue(source reflect.Value, seen map[clonePointer]reflect.Value) reflect.Value {
	if !source.IsValid() {
		return source
	}
	if source.Type() == reflect.TypeFor[TypeData]() {
		data := source.Interface().(TypeData)
		copy := data
		copy.Init = copyInitData(data.Init)
		copy.Example = cloneRenderData(data.Example)
		return reflect.ValueOf(copy)
	}
	switch source.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.String:
		return source
	case reflect.Func:
		panic(fmt.Sprintf("HTTP template data contains function value of type %s", source.Type()))
	case reflect.Interface:
		if source.IsNil() {
			return reflect.Zero(source.Type())
		}
		copy := reflect.New(source.Type()).Elem()
		copy.Set(cloneRenderValue(source.Elem(), seen))
		return copy
	case reflect.Pointer:
		if source.IsNil() {
			return reflect.Zero(source.Type())
		}
		if _, ok := immutableRenderPointers[source.Type()]; ok {
			return source
		}
		key := clonePointer{source.Type(), uintptr(source.UnsafePointer())}
		if copy, ok := seen[key]; ok {
			return copy
		}
		copy := reflect.New(source.Type().Elem())
		seen[key] = copy
		copy.Elem().Set(cloneRenderValue(source.Elem(), seen))
		return copy
	case reflect.Slice:
		if source.IsNil() {
			return reflect.Zero(source.Type())
		}
		key := clonePointer{source.Type(), source.Pointer()}
		if copy, ok := seen[key]; ok {
			return copy
		}
		copy := reflect.MakeSlice(source.Type(), source.Len(), source.Len())
		seen[key] = copy
		for index := 0; index < source.Len(); index++ {
			copy.Index(index).Set(cloneRenderValue(source.Index(index), seen))
		}
		return copy
	case reflect.Map:
		if source.IsNil() {
			return reflect.Zero(source.Type())
		}
		key := clonePointer{source.Type(), uintptr(source.UnsafePointer())}
		if copy, ok := seen[key]; ok {
			return copy
		}
		copy := reflect.MakeMapWithSize(source.Type(), source.Len())
		seen[key] = copy
		iterator := source.MapRange()
		for iterator.Next() {
			copy.SetMapIndex(cloneRenderValue(iterator.Key(), seen), cloneRenderValue(iterator.Value(), seen))
		}
		return copy
	case reflect.Struct:
		copy := reflect.New(source.Type()).Elem()
		for index := 0; index < source.NumField(); index++ {
			field := source.Type().Field(index)
			if field.PkgPath != "" {
				panic(fmt.Sprintf("HTTP template data contains private field %s.%s", source.Type(), field.Name))
			}
			copy.Field(index).Set(cloneRenderValue(source.Field(index), seen))
		}
		return copy
	case reflect.Array:
		copy := reflect.New(source.Type()).Elem()
		for index := 0; index < source.Len(); index++ {
			copy.Index(index).Set(cloneRenderValue(source.Index(index), seen))
		}
		return copy
	default:
		panic(fmt.Sprintf("HTTP template data contains unsupported %s value", source.Kind()))
	}
}
