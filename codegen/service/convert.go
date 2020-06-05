package service

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// convertData contains the info needed to render convert and create functions.
type convertData struct {
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

func commonPath(sep byte, paths ...string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	// Note, we treat string as []byte, not []rune as is often
	// done in Go. (And sep as byte, not rune). This is because
	// most/all supported OS' treat paths as string of non-zero
	// bytes. A filename may be displayed as a sequence of Unicode
	// runes (typically encoded as UTF-8) but paths are
	// not required to be valid UTF-8 or in any normalized form
	// (e.g. "é" (U+00C9) and "é" (U+0065,U+0301) are different
	// file names.
	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle the case where the
	// common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" (in which case we
	// make it "//", but this will get fixed up to "/" bellow).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

// getPkgImport returns the correct import path of a package.
// It's needed because the "reflect" package provides the binary import path
// ("goa.design/goa/vendor/some/package") for vendored packages
// instead the source import path ("some/package")
func getPkgImport(pkg, cwd string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	gosrc := path.Join(filepath.ToSlash(gopath), "src")
	cwd = filepath.ToSlash(cwd)

	// check for go modules
	if !strings.HasPrefix(cwd, gosrc) {
		return pkg
	}

	pkgpath := path.Join(gosrc, pkg)
	parentpath := commonPath(os.PathSeparator, cwd, pkgpath)

	// check for external packages
	if parentpath == gosrc {
		return pkg
	}

	rootpkg := string(parentpath[len(gosrc)+1:])

	// check for vendored packages
	vendorPrefix := path.Join(rootpkg, "vendor")
	if strings.HasPrefix(pkg, vendorPrefix) {
		return string(pkg[len(vendorPrefix)+1:])
	}

	return pkg
}

func getExternalTypeInfo(external interface{}) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	pkg := reflect.TypeOf(external)
	pkgImport := getPkgImport(pkg.PkgPath(), cwd)
	alias := strings.Split(pkg.String(), ".")[0]
	return pkgImport, alias, nil
}

// ConvertFile returns the file containing the conversion and creation functions
// if any.
func ConvertFile(root *expr.RootExpr, service *expr.ServiceExpr) (*codegen.File, error) {
	// Filter conversion and creation functions that are relevant for this
	// service
	svc := Services.Get(service.Name)
	var conversions, creations []*expr.TypeMap
	for _, c := range root.Conversions {
		for _, m := range service.Methods {
			if ut, ok := m.Payload.Type.(expr.UserType); ok {
				if ut.Name() == c.User.Name() {
					conversions = append(conversions, c)
					break
				}
			}
		}
		for _, m := range service.Methods {
			if ut, ok := m.Result.Type.(expr.UserType); ok {
				if ut.Name() == c.User.Name() {
					conversions = append(conversions, c)
					break
				}
			}
		}
		for _, t := range svc.userTypes {
			if c.User.Name() == t.Name {
				conversions = append(conversions, c)
				break
			}
		}
	}
	for _, c := range root.Creations {
		for _, m := range service.Methods {
			if ut, ok := m.Payload.Type.(expr.UserType); ok {
				if ut.Name() == c.User.Name() {
					creations = append(creations, c)
					break
				}
			}
		}
		for _, m := range service.Methods {
			if ut, ok := m.Result.Type.(expr.UserType); ok {
				if ut.Name() == c.User.Name() {
					creations = append(creations, c)
					break
				}
			}
		}
		for _, t := range svc.userTypes {
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
		pkgImport, alias, err := getExternalTypeInfo(c.External)
		if err != nil {
			return nil, err
		}
		ppm[pkgImport] = alias
	}
	for _, c := range creations {
		pkgImport, alias, err := getExternalTypeInfo(c.External)
		if err != nil {
			return nil, err
		}
		ppm[pkgImport] = alias
	}
	pkgs := make([]*codegen.ImportSpec, len(ppm))
	i := 0
	for pp, alias := range ppm {
		pkgs[i] = &codegen.ImportSpec{Name: alias, Path: pp}
		i++
	}

	// Build header section
	pkgs = append(pkgs, &codegen.ImportSpec{Path: "context"})
	pkgs = append(pkgs, codegen.GoaImport(""))
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "convert.go")
	sections := []*codegen.SectionTemplate{
		codegen.Header(service.Name+" service type conversion functions", svc.PkgName, pkgs),
	}

	var (
		names = map[string]struct{}{}

		transFuncs []*codegen.TransformFunctionData
	)

	// Build conversion sections if any
	for _, c := range conversions {
		var dt expr.DataType
		if err := buildDesignType(&dt, reflect.TypeOf(c.External), c.User); err != nil {
			return nil, err
		}
		t := reflect.TypeOf(c.External)
		tgtPkg := t.String()
		tgtPkg = tgtPkg[:strings.Index(tgtPkg, ".")]
		srcCtx := typeContext("", svc.Scope)
		tgtCtx := codegen.NewAttributeContext(false, false, false, tgtPkg, codegen.NewNameScope())
		srcAtt := &expr.AttributeExpr{Type: c.User}
		code, tf, err := codegen.GoTransform(
			&expr.AttributeExpr{Type: c.User}, &expr.AttributeExpr{Type: dt},
			"t", "v", srcCtx, tgtCtx, "transform", true)
		if err != nil {
			return nil, err
		}
		transFuncs = codegen.AppendHelpers(transFuncs, tf)
		base := "ConvertTo" + t.Name()
		name := uniquify(base, names)
		ref := t.String()
		if expr.IsObject(c.User) {
			ref = "*" + ref
		}
		data := convertData{
			Name:            name,
			ReceiverTypeRef: svc.Scope.GoTypeRef(srcAtt),
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
		var dt expr.DataType
		if err := buildDesignType(&dt, reflect.TypeOf(c.External), c.User); err != nil {
			return nil, err
		}
		t := reflect.TypeOf(c.External)
		srcPkg := t.String()
		srcPkg = srcPkg[:strings.Index(srcPkg, ".")]
		srcCtx := codegen.NewAttributeContext(false, false, false, srcPkg, codegen.NewNameScope())
		tgtCtx := typeContext("", svc.Scope)
		tgtAtt := &expr.AttributeExpr{Type: c.User}
		code, tf, err := codegen.GoTransform(
			&expr.AttributeExpr{Type: dt}, tgtAtt,
			"v", "temp", srcCtx, tgtCtx, "transform", true)
		if err != nil {
			return nil, err
		}
		transFuncs = codegen.AppendHelpers(transFuncs, tf)
		base := "CreateFrom" + t.Name()
		name := uniquify(base, names)
		ref := t.String()
		if expr.IsObject(c.User) {
			ref = "*" + ref
		}
		data := convertData{
			Name:            name,
			ReceiverTypeRef: codegen.NewNameScope().GoTypeRef(tgtAtt),
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
	seen := make(map[string]struct{})
	for _, tf := range transFuncs {
		if _, ok := seen[tf.Name]; ok {
			continue
		}
		seen[tf.Name] = struct{}{}
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
	seen map[string]expr.DataType
}

func (r dtRec) append(p string) dtRec {
	r.path += p
	return r
}

// buildDesignType builds a user type that represents the given external type.
// ref is the user type the data type being built is converted to or created
// from. It's used to compute the non-generated type field names and can be nil
// if no matching attribute exists.
func buildDesignType(dt *expr.DataType, t reflect.Type, ref expr.DataType, recs ...dtRec) error {
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
		rec.seen = make(map[string]expr.DataType)
	}

	switch t.Kind() {
	case reflect.Bool:
		*dt = expr.Boolean

	case reflect.Int:
		*dt = expr.Int

	case reflect.Int32:
		*dt = expr.Int32

	case reflect.Int64:
		*dt = expr.Int64

	case reflect.Uint:
		*dt = expr.UInt

	case reflect.Uint32:
		*dt = expr.UInt32

	case reflect.Uint64:
		*dt = expr.UInt64

	case reflect.Float32:
		*dt = expr.Float32

	case reflect.Float64:
		*dt = expr.Float64

	case reflect.String:
		*dt = expr.String

	case reflect.Slice:
		e := t.Elem()
		if e.Kind() == reflect.Uint8 {
			*dt = expr.Bytes
			return nil
		}
		var eref expr.DataType
		if ref != nil {
			eref = expr.AsArray(ref).ElemType.Type
		}
		var elem expr.DataType
		if err := buildDesignType(&elem, e, eref, rec.append("[0]")); err != nil {
			return fmt.Errorf("%s", err)
		}
		*dt = &expr.Array{ElemType: &expr.AttributeExpr{Type: elem}}

	case reflect.Map:
		var kref, vref expr.DataType
		if ref != nil {
			m := expr.AsMap(ref)
			kref = m.KeyType.Type
			vref = m.ElemType.Type
		}
		var kt expr.DataType
		if err := buildDesignType(&kt, t.Key(), kref, rec.append(".key")); err != nil {
			return fmt.Errorf("%s", err)
		}
		var vt expr.DataType
		if err := buildDesignType(&vt, t.Elem(), vref, rec.append(".value")); err != nil {
			return fmt.Errorf("%s", err)
		}
		*dt = &expr.Map{KeyType: &expr.AttributeExpr{Type: kt}, ElemType: &expr.AttributeExpr{Type: vt}}

	case reflect.Struct:
		var oref *expr.Object
		if ref != nil {
			oref = expr.AsObject(ref)
		}

		// Build list of fields that should not be ignored.
		var fields []reflect.StructField
		for i := 0; i < t.NumField(); i++ {
			f := t.FieldByIndex([]int{i})
			atn, _ := attributeName(oref, f.Name)
			if oref != nil {
				if at := oref.Attribute(atn); at != nil {
					if m := at.Meta["struct.field.external"]; len(m) > 0 {
						if m[0] == "-" {
							continue
						}
					}
				}
			}
			fields = append(fields, f)
		}

		// Avoid infinite recursions
		obj := expr.Object(make([]*expr.NamedAttributeExpr, len(fields)))
		ut := &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{Type: &obj},
			TypeName:      t.Name(),
			UID:           t.PkgPath() + "#" + t.Name(),
		}
		*dt = ut
		rec.seen[t.Name()] = ut
		var required []string
		for i, f := range fields {
			recf := rec.append("." + f.Name)
			atn, fn := attributeName(oref, f.Name)
			var aref expr.DataType
			if oref != nil {
				if at := oref.Attribute(atn); at != nil {
					aref = at.Type
				}
			}
			var fdt expr.DataType
			if f.Type.Kind() == reflect.Ptr {
				if err := buildDesignType(&fdt, f.Type.Elem(), aref, recf); err != nil {
					return fmt.Errorf("%q.%s: %s", t.Name(), f.Name, err)
				}
				if expr.IsArray(fdt) {
					return fmt.Errorf("%s: field of type pointer to slice are not supported, use slice instead", rec.path)
				}
				if expr.IsMap(fdt) {
					return fmt.Errorf("%s: field of type pointer to map are not supported, use map instead", rec.path)
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
			obj[i] = &expr.NamedAttributeExpr{
				Name:      name,
				Attribute: &expr.AttributeExpr{Type: fdt},
			}
		}
		if len(required) > 0 {
			ut.Validation = &expr.ValidationExpr{Required: required}
		}
		return nil

	case reflect.Ptr:
		rec.path = "*(" + rec.path + ")"
		if err := buildDesignType(dt, t.Elem(), ref, rec); err != nil {
			return err
		}
		if !expr.IsObject(*dt) {
			return fmt.Errorf("%s: only pointer to struct can be converted", rec.path)
		}
	default:
		*dt = expr.Any
	}
	return nil
}

// attributeName computes the name of the attribute for the given field name and
// object that must contain the matching attribute.
func attributeName(obj *expr.Object, name string) (string, string) {
	if obj == nil {
		return name, ""
	}
	// first look for a "struct.field.external" meta
	for _, nat := range *obj {
		if m := nat.Attribute.Meta["struct.field.external"]; len(m) > 0 {
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
	// next look for a lower camel case without acronym
	lcn := codegen.CamelCase(name, false, false)
	for _, nat := range *obj {
		if nat.Name == lcn {
			return lcn, name
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
func compatible(from expr.DataType, to reflect.Type, recs ...compRec) error {
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

	if expr.IsArray(from) {
		if to.Kind() != reflect.Slice {
			return fmt.Errorf("types don't match: %s must be a slice", rec.path)
		}
		return compatible(
			expr.AsArray(from).ElemType.Type,
			to.Elem(),
			rec.append("[0]"),
		)
	}

	if expr.IsMap(from) {
		if to.Kind() != reflect.Map {
			return fmt.Errorf("types don't match: %s is not a map", rec.path)
		}
		if err := compatible(
			expr.AsMap(from).ElemType.Type,
			to.Elem(),
			rec.append(".value"),
		); err != nil {
			return err
		}
		return compatible(
			expr.AsMap(from).KeyType.Type,
			to.Key(),
			rec.append(".key"),
		)
	}

	if expr.IsObject(from) {
		if to.Kind() != reflect.Struct {
			return fmt.Errorf("types don't match: %s is a %s, expected a struct", rec.path, toName)
		}
		obj := expr.AsObject(from)
		ma := expr.NewMappedAttributeExpr(&expr.AttributeExpr{Type: obj})
		for _, nat := range *obj {
			var (
				fname string
				ok    bool
				field reflect.StructField
			)
			{
				if ef, k := nat.Attribute.Meta["struct.field.external"]; k {
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
		var dt expr.DataType
		if err := buildDesignType(&dt, to, nil); err != nil {
			return err
		}
		if expr.Equal(dt, from) {
			return nil
		}
	}

	return fmt.Errorf("types don't match: type of %s is %s but type of corresponding attribute is %s", rec.path, toName, from.Name())
}

// input: convertData
const convertT = `{{ printf "%s creates an instance of %s initialized from t." .Name .TypeName | comment }}
func (t {{ .ReceiverTypeRef }}) {{ .Name }}() {{ .TypeRef }} {
    {{ .Code }}
    return v
}
`

// input: convertData
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
