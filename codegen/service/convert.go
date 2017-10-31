package service

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// ConvertData contains the info needed to render convert and create functions.
type ConvertData struct {
	// Name is the name of the function.
	Name string
	// ReceiverTypeRef is a reference to the receiver type.
	ReceiverTypeRef string
	// TypeRef is a reference to the external type.
	TypeRef string
	// TypeName is the name of the external type.
	TypeName string
	//  Code is the function code.
	Code string
}

// ConvertFile returns the file containing the conversion and creation functions
// if any.
func ConvertFile(root *design.RootExpr, service *design.ServiceExpr) (*codegen.File, error) {
	// Filter conversion and creation functions that are relevant for this
	// service
	svc := Services.Get(service.Name)
	var conversions, creations []*design.TypeMap
	for _, c := range root.Conversions {
		for _, m := range service.Methods {
			if ut, ok := m.Payload.Type.(design.UserType); ok {
				if ut.Name() == c.User.Name() {
					conversions = append(conversions, c)
					break
				}
			}
		}
		for _, m := range service.Methods {
			if ut, ok := m.Result.Type.(design.UserType); ok {
				if ut.Name() == c.User.Name() {
					conversions = append(conversions, c)
					break
				}
			}
		}
		for _, t := range svc.UserTypes {
			if c.User.Name() == t.Name {
				conversions = append(conversions, c)
				break
			}
		}
	}
	for _, c := range root.Creations {
		for _, m := range service.Methods {
			if ut, ok := m.Payload.Type.(design.UserType); ok {
				if ut.Name() == c.User.Name() {
					creations = append(creations, c)
					break
				}
			}
		}
		for _, m := range service.Methods {
			if ut, ok := m.Result.Type.(design.UserType); ok {
				if ut.Name() == c.User.Name() {
					creations = append(creations, c)
					break
				}
			}
		}
		for _, t := range svc.UserTypes {
			if c.User.Name() == t.Name {
				creations = append(creations, c)
				break
			}
		}
	}
	if len(conversions) == 0 && len(creations) == 0 {
		return nil, nil
	}

	// Retrieve external packages info
	ppm := make(map[string]struct{})
	for _, c := range conversions {
		pkg := reflect.TypeOf(c.External).PkgPath()
		ppm[pkg] = struct{}{}
	}
	for _, c := range creations {
		pkg := reflect.TypeOf(c.External).PkgPath()
		ppm[pkg] = struct{}{}
	}
	pkgs := make([]*codegen.ImportSpec, len(ppm))
	i := 0
	for pp := range ppm {
		pkgs[i] = &codegen.ImportSpec{Path: pp}
		i++
	}

	// Build header section
	pkgs = append(pkgs, &codegen.ImportSpec{Path: "context"})
	pkgs = append(pkgs, &codegen.ImportSpec{Path: "goa.design/goa"})
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "convert.go")
	sections := []*codegen.SectionTemplate{
		codegen.Header(service.Name+" service type conversion functions", svc.PkgName, pkgs),
	}

	var (
		names      = map[string]struct{}{}
		transFuncs []*codegen.TransformFunctionData
	)

	// Build conversion sections if any
	for _, c := range conversions {
		dt, err := designType(reflect.TypeOf(c.External))
		if err != nil {
			return nil, err
		}
		t := reflect.TypeOf(c.External)
		tgtPkg := t.String()
		tgtPkg = tgtPkg[:strings.Index(tgtPkg, ".")]
		code, tf, err := codegen.GoTypeTransform(c.User, dt, "t", "v", "", tgtPkg, false, svc.Scope)
		if err != nil {
			return nil, err
		}
		transFuncs = append(transFuncs, tf...)
		base := "ConvertTo" + t.Name()
		name := uniquify(base, names)
		ref := t.String()
		if design.IsObject(c.User) {
			ref = "*" + ref
		}
		data := ConvertData{
			Name:            name,
			ReceiverTypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: c.User}),
			TypeName:        t.Name(),
			TypeRef:         ref,
			Code:            code,
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "convert-to",
			Source: convertT,
			Data:   data,
		})
	}

	// Build creation sections if any
	for _, c := range creations {
		dt, err := designType(reflect.TypeOf(c.External))
		if err != nil {
			return nil, err
		}
		t := reflect.TypeOf(c.External)
		srcPkg := t.String()
		srcPkg = srcPkg[:strings.Index(srcPkg, ".")]
		code, tf, err := codegen.GoTypeTransform(dt, c.User, "v", "temp", srcPkg, "", false, svc.Scope)
		if err != nil {
			return nil, err
		}
		transFuncs = append(transFuncs, tf...)
		base := "CreateFrom" + t.Name()
		name := uniquify(base, names)
		ref := t.String()
		if design.IsObject(c.User) {
			ref = "*" + ref
		}
		data := ConvertData{
			Name:            name,
			ReceiverTypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: c.User}),
			TypeRef:         ref,
			Code:            code,
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "create-from",
			Source: createT,
			Data:   data,
		})
	}

	// Build transformation helper functions section if any.
	for _, tf := range transFuncs {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "convert-create-helper",
			Source: transformHelperT,
			Data:   tf,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}, nil
}

// uniquify checks if base is a key of taken and if not returns it. Otherwise
// uniquify appends integers to base starting at 2 and incremented by 1 each
// time a key already exists for the value. uniquify returns the unique value
// and updates taken with it.
func uniquify(base string, taken map[string]struct{}) string {
	name := base
	idx := 2
	_, ok := taken[name]
	for ok {
		name = base + strconv.Itoa(idx)
		idx++
		_, ok = taken[name]
	}
	taken[name] = struct{}{}
	return name
}

// designType returns a user type that represents the given external type.
// If val is a slice it must have at least one element.
// If val is a map it must have at least one key.
func designType(t reflect.Type, ctxs ...string) (design.DataType, error) {
	var ctx string
	if ctxs == nil {
		ctx = "<value>"
	} else {
		ctx = ctxs[0]
	}

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

	case reflect.Slice:
		e := t.Elem()
		if e.Kind() == reflect.Uint8 {
			return design.Bytes, nil
		}
		elem, err := designType(e, ctx+"[0]")
		if err != nil {
			return nil, err
		}
		return &design.Array{ElemType: &design.AttributeExpr{Type: elem}}, nil

	case reflect.Map:
		kt, err := designType(t.Key(), ctx+".key")
		if err != nil {
			return nil, err
		}
		vt, err := designType(t.Elem(), ctx+".value")
		if err != nil {
			return nil, err
		}
		return &design.Map{KeyType: &design.AttributeExpr{Type: kt}, ElemType: &design.AttributeExpr{Type: vt}}, nil

	case reflect.Struct:
		obj := make([]*design.NamedAttributeExpr, t.NumField())
		var required []string
		for i := 0; i < t.NumField(); i++ {
			f := t.FieldByIndex([]int{i})
			var fdt design.DataType
			var err error
			if f.Type.Kind() == reflect.Ptr {
				fdt, err = designType(f.Type.Elem(), ctx+"."+f.Name)
				if err != nil {
					return nil, err
				}
				if design.IsArray(fdt) {
					if ctx != "" {
						ctx = ": " + ctx + ": "
					}
					return nil, fmt.Errorf("%sfield of type pointer to slice are not supported, use slice instead", ctx)
				}
				if design.IsMap(fdt) {
					if ctx != "" {
						ctx = ": " + ctx
					}
					return nil, fmt.Errorf("%sfield of type pointer to map are not supported, use map instead", ctx)
				}
			} else {
				if isPrimitive(f.Type) {
					required = append(required, f.Name)
				}
				fdt, err = designType(f.Type, ctx+"."+f.Name)
				if err != nil {
					return nil, err
				}
			}
			obj[i] = &design.NamedAttributeExpr{
				Name:      f.Name,
				Attribute: &design.AttributeExpr{Type: fdt},
			}
		}
		o := design.Object(obj)
		ut := &design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{Type: &o},
			TypeName:      t.Name(),
		}
		if len(required) > 0 {
			ut.Validation = &design.ValidationExpr{Required: required}
		}
		return ut, nil

	case reflect.Ptr:
		dt, err := designType(t.Elem(), "(*"+ctx+")")
		if err != nil {
			return nil, err
		}
		if !design.IsObject(dt) {
			if ctx != "" {
				ctx = ctx + ": "
			}
			return nil, fmt.Errorf("%sonly pointer to struct can be converted", ctx)
		}
	}
	if ctx != "" {
		ctx = ctx + ": "
	}
	return nil, fmt.Errorf("%stype %s is not compatible with goa design types", ctx, t.Name())
}

// isPrimitive is true if the given kind matches a goa primitive type.
func isPrimitive(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		return true
	case reflect.Slice:
		e := t.Elem()
		if e.Kind() == reflect.Uint8 {
			return true
		}
		return false
	default:
		return false
	}
}

// compatible checks the user and external type definitions map recursively . It
// returns nil if they do, an error otherwise.
func compatible(from design.DataType, to reflect.Type, path ...string) error {
	// build contextual error message
	if path == nil {
		path = []string{"<value>"}
	}
	errpath := path[0]

	if design.IsArray(from) {
		if to.Kind() != reflect.Slice {
			return fmt.Errorf("types don't match: %s must be a slice", errpath)
		}
		return compatible(
			design.AsArray(from).ElemType.Type,
			to.Elem(),
			path[0]+"[0]",
		)
	}

	if design.IsMap(from) {
		if to.Kind() != reflect.Map {
			return fmt.Errorf("types don't match: %s is not a map", errpath)
		}
		if err := compatible(
			design.AsMap(from).ElemType.Type,
			to.Elem(),
			path[0]+".value",
		); err != nil {
			return err
		}
		return compatible(
			design.AsMap(from).KeyType.Type,
			to.Key(),
			path[0]+".key",
		)
	}

	if design.IsObject(from) {
		if to.Kind() != reflect.Struct {
			return fmt.Errorf("types don't match: %s is not a struct", errpath)
		}
		obj := design.AsObject(from)
		ma := design.NewMappedAttributeExpr(&design.AttributeExpr{Type: obj})
		for _, nat := range *obj {
			var (
				fname string
				ok    bool
				field reflect.StructField
			)
			{
				if ef, k := nat.Attribute.Metadata["struct.field.external"]; k {
					fname = ef[0]
					field, ok = to.FieldByName(ef[0])
				} else {
					ef := codegen.Goify(ma.ElemName(nat.Name), true)
					fname = ef
					field, ok = to.FieldByName(ef)
				}
			}
			if !ok {
				return fmt.Errorf("types don't match: could not find field %q of external type %s matching attribute %q of type %q",
					fname, to.Name(), nat.Name, from.Name())
			}
			err := compatible(
				nat.Attribute.Type,
				field.Type,
				path[0]+"."+fname,
			)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if isPrimitive(to) {
		dt, err := designType(to)
		if err != nil {
			return err
		}
		if dt == from {
			return nil
		}
	}

	return fmt.Errorf("types don't match: type of %s is %s but type of corresponding attribute is %s", errpath, to.Name(), from.Name())
}

// input: ConvertData
const convertT = `{{ printf "%s creates an instance of %s initialized from t." .Name .TypeName | comment }}
func (t {{ .ReceiverTypeRef }}) {{ .Name }}() {{ .TypeRef }} {
    {{ .Code }}
    return v
}
`

// input: ConvertData
const createT = `{{ printf "%s initializes t from the fields of v" .Name | comment }}
func (t {{ .ReceiverTypeRef }}) {{ .Name }}(v {{ .TypeRef }}) {
	{{ .Code }}
	*t = *temp
}
`

// input: TransformFunctionData
const transformHelperT = `{{ printf "%s builds a value of type %s from a value of type %s." .Name .ResultTypeRef .ParamTypeRef | comment }}
func {{ .Name }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
        {{ .Code }}
        return res
}
`
