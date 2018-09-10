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
	ppm := make(map[string]string)
	for _, c := range conversions {
		pkg := reflect.TypeOf(c.External)
		p := pkg.PkgPath()
		alias := strings.Split(pkg.String(), ".")[0]
		ppm[p] = alias
	}
	for _, c := range creations {
		pkg := reflect.TypeOf(c.External)
		p := pkg.PkgPath()
		alias := strings.Split(pkg.String(), ".")[0]
		ppm[p] = alias
	}
	pkgs := make([]*codegen.ImportSpec, len(ppm))
	i := 0
	for pp, alias := range ppm {
		pkgs[i] = &codegen.ImportSpec{Name: alias, Path: pp}
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
		var dt design.DataType
		if err := buildDesignType(&dt, reflect.TypeOf(c.External), c.User); err != nil {
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
		var dt design.DataType
		if err := buildDesignType(&dt, reflect.TypeOf(c.External), c.User); err != nil {
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

type dtRec struct {
	path string
	seen map[string]design.DataType
}

func (r dtRec) append(p string) dtRec {
	r.path += p
	return r
}

// buildDesignType builds a user type that represents the given external type.
// ref is the user type the data type being built is converted to or created
// from. It's used to compute the non-generated type field names and can be nil
// if no matching attribute exists.
func buildDesignType(dt *design.DataType, t reflect.Type, ref design.DataType, recs ...dtRec) error {
	// check compatibility
	if ref != nil {
		if err := compatible(ref, t); err != nil {
			return fmt.Errorf("%q: %s", t.Name(), err)
		}
	}

	// handle recursive data structures
	var rec dtRec
	if recs != nil {
		rec = recs[0]
		if s, ok := rec.seen[t.Name()]; ok {
			*dt = s
			return nil
		}
	} else {
		rec.path = "<value>"
		rec.seen = make(map[string]design.DataType)
	}

	switch t.Kind() {
	case reflect.Bool:
		*dt = design.Boolean

	case reflect.Int:
		*dt = design.Int

	case reflect.Int32:
		*dt = design.Int32

	case reflect.Int64:
		*dt = design.Int64

	case reflect.Uint:
		*dt = design.UInt

	case reflect.Uint32:
		*dt = design.UInt32

	case reflect.Uint64:
		*dt = design.UInt64

	case reflect.Float32:
		*dt = design.Float32

	case reflect.Float64:
		*dt = design.Float64

	case reflect.String:
		*dt = design.String

	case reflect.Slice:
		e := t.Elem()
		if e.Kind() == reflect.Uint8 {
			*dt = design.Bytes
			return nil
		}
		var eref design.DataType
		if ref != nil {
			eref = design.AsArray(ref).ElemType.Type
		}
		var elem design.DataType
		if err := buildDesignType(&elem, e, eref, rec.append("[0]")); err != nil {
			return fmt.Errorf("%s", err)
		}
		*dt = &design.Array{ElemType: &design.AttributeExpr{Type: elem}}

	case reflect.Map:
		var kref, vref design.DataType
		if ref != nil {
			m := design.AsMap(ref)
			kref = m.KeyType.Type
			vref = m.ElemType.Type
		}
		var kt design.DataType
		if err := buildDesignType(&kt, t.Key(), kref, rec.append(".key")); err != nil {
			return fmt.Errorf("%s", err)
		}
		var vt design.DataType
		if err := buildDesignType(&vt, t.Elem(), vref, rec.append(".value")); err != nil {
			return fmt.Errorf("%s", err)
		}
		*dt = &design.Map{KeyType: &design.AttributeExpr{Type: kt}, ElemType: &design.AttributeExpr{Type: vt}}

	case reflect.Struct:
		var oref *design.Object
		if ref != nil {
			oref = design.AsObject(ref)
		}

		// Build list of fields that should not be ignored.
		var fields []reflect.StructField
		for i := 0; i < t.NumField(); i++ {
			f := t.FieldByIndex([]int{i})
			atn, _ := attributeName(oref, f.Name)
			if oref != nil {
				if at := oref.Attribute(atn); at != nil {
					if m := at.Metadata["struct.field.external"]; len(m) > 0 {
						if m[0] == "-" {
							continue
						}
					}
				}
			}
			fields = append(fields, f)
		}

		// Avoid infinite recursions
		obj := design.Object(make([]*design.NamedAttributeExpr, len(fields)))
		ut := &design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{
				Type:     &obj,
				Metadata: map[string][]string{"goa.external": nil},
			},
			TypeName: t.Name(),
		}
		*dt = ut
		rec.seen[t.Name()] = ut
		var required []string
		for i, f := range fields {
			recf := rec.append("." + f.Name)
			atn, fn := attributeName(oref, f.Name)
			var aref design.DataType
			if oref != nil {
				if at := oref.Attribute(atn); at != nil {
					aref = at.Type
				}
			}
			var fdt design.DataType
			if f.Type.Kind() == reflect.Ptr {
				if err := buildDesignType(&fdt, f.Type.Elem(), aref, recf); err != nil {
					return fmt.Errorf("%q.%s: %s", t.Name(), f.Name, err)
				}
				if design.IsArray(fdt) {
					return fmt.Errorf("%s: field of type pointer to slice are not supported, use slice instead", recf.path)
				}
				if design.IsMap(fdt) {
					return fmt.Errorf("%s: field of type pointer to map are not supported, use map instead", recf.path)
				}
			} else if f.Type.Kind() == reflect.Struct {
				return fmt.Errorf("%s: fields of type struct must use pointers", recf.path)
			} else {
				if isPrimitive(f.Type) {
					required = append(required, atn)
				}
				if err := buildDesignType(&fdt, f.Type, aref, rec.append("."+f.Name)); err != nil {
					return fmt.Errorf("%q.%s: %s", t.Name(), f.Name, err)
				}
			}
			name := atn
			if fn != "" {
				name = name + ":" + fn
			}
			obj[i] = &design.NamedAttributeExpr{
				Name: name,
				Attribute: &design.AttributeExpr{
					Type:     fdt,
					Metadata: map[string][]string{"goa.external": nil},
				},
			}
		}
		if len(required) > 0 {
			ut.Validation = &design.ValidationExpr{Required: required}
		}
		return nil

	case reflect.Ptr:
		rec.path = "*(" + rec.path + ")"
		if err := buildDesignType(dt, t.Elem(), ref, rec); err != nil {
			return err
		}
		if !design.IsObject(*dt) {
			return fmt.Errorf("%s: only pointer to struct can be converted", rec.path)
		}
	default:
		*dt = design.Any
	}
	return nil
}

// attributeName computes the name of the attribute for the given field name and
// object that must contain the matching attribute.
func attributeName(obj *design.Object, name string) (string, string) {
	if obj == nil {
		return name, ""
	}
	// first look for a "struct.field.external" metadata
	for _, nat := range *obj {
		if m := nat.Attribute.Metadata["struct.field.external"]; len(m) > 0 {
			if m[0] == name {
				return nat.Name, name
			}
		}
	}
	// next look for an exact match
	for _, nat := range *obj {
		if nat.Name == name {
			return name, ""
		}
	}
	// next try to lower case first letter
	ln := strings.ToLower(name[0:1]) + name[1:]
	for _, nat := range *obj {
		if nat.Name == ln {
			return ln, name
		}
	}
	// finally look for a snake case representation
	sn := codegen.SnakeCase(name)
	for _, nat := range *obj {
		if nat.Name == sn {
			return sn, name
		}
	}
	// no match, return field name
	return name, ""
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
	case reflect.Interface:
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

type compRec struct {
	path string
	seen map[string]struct{}
}

func (r compRec) append(p string) compRec {
	r.path += p
	return r
}

// compatible checks the user and external type definitions map recursively . It
// returns nil if they do, an error otherwise.
func compatible(from design.DataType, to reflect.Type, recs ...compRec) error {
	// deference if needed
	if to.Kind() == reflect.Ptr {
		return compatible(from, to.Elem(), recs...)
	}

	toName := to.Name()
	if toName == "" {
		toName = to.Kind().String()
	}

	// handle recursive data structures
	var rec compRec
	if recs != nil {
		rec = recs[0]
		if _, ok := rec.seen[from.Hash()+"-"+toName]; ok {
			return nil
		}
	} else {
		rec = compRec{path: "<value>", seen: make(map[string]struct{})}
	}
	rec.seen[from.Hash()+"-"+toName] = struct{}{}

	if design.IsArray(from) {
		if to.Kind() != reflect.Slice {
			return fmt.Errorf("types don't match: %s must be a slice", rec.path)
		}
		return compatible(
			design.AsArray(from).ElemType.Type,
			to.Elem(),
			rec.append("[0]"),
		)
	}

	if design.IsMap(from) {
		if to.Kind() != reflect.Map {
			return fmt.Errorf("types don't match: %s is not a map", rec.path)
		}
		if err := compatible(
			design.AsMap(from).ElemType.Type,
			to.Elem(),
			rec.append(".value"),
		); err != nil {
			return err
		}
		return compatible(
			design.AsMap(from).KeyType.Type,
			to.Key(),
			rec.append(".key"),
		)
	}

	if design.IsObject(from) {
		if to.Kind() != reflect.Struct {
			return fmt.Errorf("types don't match: %s is a %s, expected a struct", rec.path, toName)
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
					if fname == "-" {
						continue
					}
					field, ok = to.FieldByName(ef[0])
				} else {
					ef := codegen.Goify(ma.ElemName(nat.Name), true)
					fname = ef
					field, ok = to.FieldByName(ef)
				}
			}
			if !ok {
				return fmt.Errorf("types don't match: could not find field %q of external type %q matching attribute %q of type %q",
					fname, toName, nat.Name, from.Name())
			}
			err := compatible(
				nat.Attribute.Type,
				field.Type,
				rec.append("."+fname),
			)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if isPrimitive(to) {
		var dt design.DataType
		if err := buildDesignType(&dt, to, nil); err != nil {
			return err
		}
		if design.Equal(dt, from) {
			return nil
		}
	}

	return fmt.Errorf("types don't match: type of %s is %s but type of corresponding attribute is %s", rec.path, toName, from.Name())
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
