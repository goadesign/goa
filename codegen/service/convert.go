// This file generates ConvertTo and CreateFrom functions for service types
// mapped to external Go structs. Service-side names come from the completed
// package records, including nested types placed in packages by design
// metadata.
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

// convertFiles formats the package-owned external conversion files aggregated
// from every linked root plan. It does not inspect design mappings or allocate
// generated names.
func convertFiles(conversions []*externalConversionFileFacts) []*codegen.File {
	files := make([]*codegen.File, len(conversions))
	for index, retained := range conversions {
		sections := []*codegen.SectionTemplate{
			codegen.Header(
				"External type conversion functions",
				codegen.Goify(path.Base(retained.owner.ImportPath()), false),
				retained.imports.Imports(),
			),
		}
		for _, operation := range retained.operations {
			name, source := "convert-to", serviceTemplates.Read(convertT)
			if operation.direction == externalCreateFrom {
				name, source = "create-from", serviceTemplates.Read(createT)
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   name,
				Source: source,
				Data:   operation.data,
			})
		}
		for _, operation := range retained.operations {
			for _, helper := range operation.helpers {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "convert-create-helper",
					Source: serviceTemplates.Read(transformHelperT),
					Data:   helper,
				})
			}
		}
		files[index] = &codegen.File{
			Path:             filepath.Join(retained.owner.OutputDirectory(), "convert.go"),
			SectionTemplates: sections,
		}
	}
	return files
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

	rootpkg := parentpath[len(gosrc)+1:]

	// check for vendored packages
	vendorPrefix := path.Join(rootpkg, "vendor")
	if strings.HasPrefix(pkg, vendorPrefix) {
		return pkg[len(vendorPrefix)+1:]
	}

	return pkg
}

// getExternalReflectTypeInfo returns the source import path and authored
// package qualifier for one named reflected type.
func getExternalReflectTypeInfo(pkg reflect.Type) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	pkgImport := getPkgImport(pkg.PkgPath(), cwd)
	alias := strings.Split(pkg.String(), ".")[0]
	return pkgImport, alias, nil
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
	path  string
	seen  map[reflect.Type]expr.DataType
	named map[expr.UserType]reflect.Type
}

func appendPath(r dtRec, p string) dtRec {
	r.path += p
	return r
}

// buildExternalDesignType returns the reflected design graph and the exact Go
// type behind every named node that may require a package-qualified reference.
func buildExternalDesignType(t reflect.Type, ref expr.DataType) (expr.DataType, map[expr.UserType]reflect.Type, error) {
	named := make(map[expr.UserType]reflect.Type)
	rec := dtRec{
		path:  "<value>",
		seen:  make(map[reflect.Type]expr.DataType),
		named: named,
	}
	var dataType expr.DataType
	if err := buildDesignType(&dataType, t, ref, rec); err != nil {
		return nil, nil, err
	}
	return dataType, named, nil
}

// buildDesignType builds a user type that represents the given external type.
// ref is the user type the data type being built is converted to or created
// from. It's used to compute the non-generated type field names and can be nil
// if no matching attribute exists.
func buildDesignType(dt *expr.DataType, t reflect.Type, ref expr.DataType, recs ...dtRec) error {
	// check compatibility
	if ref != nil {
		if err := compatible(ref, t); err != nil {
			return fmt.Errorf("%q: %w", t.Name(), err)
		}
	}

	// handle recursive data structures
	var rec dtRec
	if recs != nil {
		rec = recs[0]
		if s, ok := rec.seen[t]; ok {
			*dt = s
			return nil
		}
	} else {
		rec.path = "<value>"
		rec.seen = make(map[reflect.Type]expr.DataType)
		rec.named = make(map[expr.UserType]reflect.Type)
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
		if err := buildDesignType(&elem, e, eref, appendPath(rec, "[0]")); err != nil {
			return fmt.Errorf("%w", err)
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
		if err := buildDesignType(&kt, t.Key(), kref, appendPath(rec, ".key")); err != nil {
			return fmt.Errorf("%w", err)
		}
		var vt expr.DataType
		if err := buildDesignType(&vt, t.Elem(), vref, appendPath(rec, ".value")); err != nil {
			return fmt.Errorf("%w", err)
		}
		*dt = &expr.Map{KeyType: &expr.AttributeExpr{Type: kt}, ElemType: &expr.AttributeExpr{Type: vt}}

	case reflect.Struct:
		var oref *expr.Object
		if ref != nil {
			oref = expr.AsObject(ref)
		}

		// Keep only fields represented by the matching design object. External
		// structs may contain additional fields, but generated transforms neither
		// read nor write them and therefore must not reserve their package imports.
		var fields []reflect.StructField
		for i := 0; i < t.NumField(); i++ {
			f := t.FieldByIndex([]int{i})
			atn, _ := attributeName(oref, f.Name)
			if oref != nil {
				at := oref.Attribute(atn)
				if at == nil {
					continue
				}
				if m := at.Meta["struct:field:external"]; len(m) > 0 {
					if m[0] == "-" {
						continue
					}
				}
				if m := at.Meta["struct.field.external"]; len(m) > 0 { // Deprecated syntax. Only present for backward compatibility.
					if m[0] == "-" {
						continue
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
		rec.seen[t] = ut
		rec.named[ut] = t
		var required []string
		for i, f := range fields {
			recf := appendPath(rec, "."+f.Name)
			atn, fn := attributeName(oref, f.Name)
			var aref expr.DataType
			if oref != nil {
				if at := oref.Attribute(atn); at != nil {
					aref = at.Type
				}
			}
			var fdt expr.DataType
			switch f.Type.Kind() {
			case reflect.Pointer:
				if err := buildDesignType(&fdt, f.Type.Elem(), aref, recf); err != nil {
					return fmt.Errorf("%q.%s: %w", t.Name(), f.Name, err)
				}
				if expr.IsArray(fdt) {
					return fmt.Errorf("%s: field of type pointer to slice are not supported, use slice instead", rec.path)
				}
				if expr.IsMap(fdt) {
					return fmt.Errorf("%s: field of type pointer to map are not supported, use map instead", rec.path)
				}
			case reflect.Struct:
				return fmt.Errorf("%s: fields of type struct must use pointers", recf.path)
			default:
				if isPrimitive(f.Type) {
					required = append(required, atn)
				}
				if err := buildDesignType(&fdt, f.Type, aref, appendPath(rec, "."+f.Name)); err != nil {
					return fmt.Errorf("%q.%s: %w", t.Name(), f.Name, err)
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

	case reflect.Pointer:
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
	// first look for a "struct:field:external" meta
	for _, nat := range *obj {
		if m := nat.Attribute.Meta["struct:field:external"]; len(m) > 0 {
			if m[0] == name {
				return nat.Name, name
			}
		}
	}
	for _, nat := range *obj { // Deprecated syntax. Only present for backward compatibility.
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

func appendCompPath(r compRec, p string) compRec {
	r.path += p
	return r
}

// compatible checks the user and external type definitions map recursively . It
// returns nil if they do, an error otherwise.
func compatible(from expr.DataType, to reflect.Type, recs ...compRec) error {
	// deference if needed
	if to.Kind() == reflect.Pointer {
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
			appendCompPath(rec, "[0]"),
		)
	}

	if expr.IsMap(from) {
		if to.Kind() != reflect.Map {
			return fmt.Errorf("types don't match: %s is not a map", rec.path)
		}
		if err := compatible(
			expr.AsMap(from).ElemType.Type,
			to.Elem(),
			appendCompPath(rec, ".value"),
		); err != nil {
			return err
		}
		return compatible(
			expr.AsMap(from).KeyType.Type,
			to.Key(),
			appendCompPath(rec, ".key"),
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
			if ef, k := nat.Attribute.Meta["struct:field:external"]; k {
				fname = ef[0]
				if fname == "-" {
					continue
				}
				field, ok = to.FieldByName(ef[0])
			} else if ef, k := nat.Attribute.Meta["struct.field.external"]; k { // Deprecated syntax. Only present for backward compatibility.
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
			if !ok {
				return fmt.Errorf("types don't match: could not find field %q of external type %q matching attribute %q of type %q",
					fname, toName, nat.Name, from.Name())
			}
			err := compatible(
				nat.Attribute.Type,
				field.Type,
				appendCompPath(rec, "."+fname),
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
