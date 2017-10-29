package conversion

import (
	"fmt"
	"reflect"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// DesignType returns a user type that represents the given external type.
// If val is a slice it must have at least one element.
// If val is a map it must have at least one key.
func DesignType(val interface{}) (design.DataType, error) {
	t := reflect.TypeOf(val)
	v := reflect.ValueOf(val)

	switch t.Kind() {
	case reflect.Bool:
		return design.Boolean, nil
	case reflect.Int:
		return design.Int, nil
	case reflect.Int32:
		return design.Int32, nil
	case reflect.Int64:
		return design.Int64, nil
	case reflect.Uint:
		return design.UInt, nil
	case reflect.Uint32:
		return design.UInt32, nil
	case reflect.Uint64:
		return design.UInt64, nil
	case reflect.Float32:
		return design.Float32, nil
	case reflect.Float64:
		return design.Float64, nil
	case reflect.String:
		return design.String, nil
	case reflect.Array:
		e := v.Index(0)
		if e.Kind() == reflect.Uint8 {
			return design.Bytes, nil
		}
		elem, err := DesignType(e.Interface())
		if err != nil {
			return nil, err
		}
		return &design.Array{ElemType: &design.AttributeExpr{Type: elem}}, nil
	case reflect.Map:
		key := v.MapKeys()[0]
		kt, err := DesignType(key.Interface())
		if err != nil {
			return nil, err
		}
		vt, err := DesignType(v.MapIndex(key))
		if err != nil {
			return nil, err
		}
		return &design.Map{KeyType: &design.AttributeExpr{Type: kt}, ElemType: &design.AttributeExpr{Type: vt}}, nil
	case reflect.Struct:
	default:
		return nil, fmt.Errorf("type %T is not compatible with goa design types", val)
	}
}

// Compatible checks the user and external type definitions map recursively . It
// returns nil if they do, an error otherwise.
func Compatible(from design.DataType, to interface{}, path ...string) error {
	// build contextual error message
	if path == nil {
		path = []string{"<value>"}
	}
	errpath := path[0]

	if design.IsArray(from) {
		if reflect.TypeOf(to).Kind() != reflect.Slice {
			return fmt.Errorf("types don't match: %s must be a slice", errpath)
		}
		v := reflect.ValueOf(to)
		if v.Len() != 1 {
			return fmt.Errorf("slice %s must contain exactly one item", errpath)
		}
		return Compatible(
			design.AsArray(from).ElemType.Type,
			v.Index(0).Interface(),
			path[0]+"[0]",
		)
	}

	if design.IsMap(from) {
		if reflect.TypeOf(to).Kind() != reflect.Map {
			return fmt.Errorf("types don't match: %s is not a map", errpath)
		}
		v := reflect.ValueOf(to)
		if v.Len() != 1 {
			return fmt.Errorf("map %s must contain exactly one key", errpath)
		}
		if err := Compatible(
			design.AsMap(from).ElemType.Type,
			v.MapIndex(v.MapKeys()[0]).Interface(),
			path[0]+".value",
		); err != nil {
			return err
		}
		return Compatible(
			design.AsMap(from).KeyType.Type,
			v.MapKeys()[0].Interface(),
			path[0]+".key",
		)
	}

	if design.IsObject(from) {
		if reflect.TypeOf(to).Kind() != reflect.Struct {
			return fmt.Errorf("types don't match: %s is not a struct", errpath)
		}
		obj := design.AsObject(from)
		ma := design.NewMappedAttributeExpr(&design.AttributeExpr{Type: obj})
		v := reflect.ValueOf(to)
		t := reflect.TypeOf(to)
		for _, nat := range *obj {
			var (
				fname string
				ok    bool
			)
			{
				if ef, k := nat.Attribute.Metadata["struct.field.external"]; k {
					fname = ef[0]
					_, ok = t.FieldByName(ef[0])
				} else {
					ef := codegen.Goify(ma.ElemName(nat.Name), true)
					if f, k := t.FieldByName(ef); k {
						fname = f.Name
						ok = true
					}
				}
			}
			if !ok {
				return fmt.Errorf("types don't match: could not find field %q of external type %T matching attribute %q of type %q",
					fname, to, nat.Name, from.Name())
			}
			err := Compatible(
				nat.Attribute.Type,
				v.FieldByName(fname).Interface(),
				path[0]+"."+fname,
			)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Primitive
	if !from.IsCompatible(to) {
		return fmt.Errorf("types don't match: type of %s is %T but type of corresponding attribute is %s", errpath, to, from.Name())
	}

	return nil
}
