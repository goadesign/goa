package codegen

import (
	"fmt"
	"path/filepath"
	"strconv"

	"goa.design/goa/v3/expr"
	goa "goa.design/goa/v3/pkg"
)

// DesignVersion contains the major component of the version of Goa used to
// author the design - either 2 or 3. This value is initialized when the
// generated tool is invoked by retrieving the information passed on the
// command line by the goa tool.
var DesignVersion = goa.Major

type (
	// ImportSpec defines a generated import statement.
	ImportSpec struct {
		// Name of imported package if needed.
		Name string
		// Go import path of package.
		Path string
	}

	// Location defines a file location and import details.
	Location struct {
		// FilePath is the path to the file.
		FilePath string
		// RelImportPath is the Go import path starting after the gen
		// folder.
		RelImportPath string
	}
)

// NewImport creates an import spec.
func NewImport(name, path string) *ImportSpec {
	return &ImportSpec{Name: name, Path: path}
}

// SimpleImport creates an import with no explicit path component.
func SimpleImport(path string) *ImportSpec {
	return &ImportSpec{Path: path}
}

// GoaImport creates an import for a Goa package.
func GoaImport(rel string) *ImportSpec {
	name := ""
	if rel == "" {
		name = "goa"
		rel = "pkg"
	}
	return GoaNamedImport(rel, name)
}

// GoaNamedImport creates an import for a Goa package with the given name.
func GoaNamedImport(rel, name string) *ImportSpec {
	root := "goa.design/goa"
	if DesignVersion > 2 {
		root += "/v" + strconv.Itoa(DesignVersion)
	}
	if rel != "" {
		rel = "/" + rel
	}
	return &ImportSpec{Name: name, Path: root + rel}
}

// Code returns the Go import statement for the ImportSpec.
func (s *ImportSpec) Code() string {
	if len(s.Name) > 0 {
		return fmt.Sprintf(`%s "%s"`, s.Name, s.Path)
	}
	return fmt.Sprintf(`"%s"`, s.Path)
}

// UserTypeLocation returns the location of the user type if set via the
// attr:pkg:path metadata, nil otherwise..
func UserTypeLocation(dt expr.DataType) *Location {
	ut, ok := dt.(expr.UserType)
	if !ok {
		return nil
	}
	p, ok := ut.Attribute().Meta.Last("struct:pkg:path")
	if !ok || p == "" {
		return nil
	}
	return &Location{
		FilePath:      filepath.Join(filepath.FromSlash(p), SnakeCase(ut.Name())+".go"),
		RelImportPath: p,
	}
}

// PackageName returns the package name of the given location.
func (loc *Location) PackageName() string {
	if loc == nil {
		return ""
	}
	return Goify(filepath.Base(loc.RelImportPath), false)
}

// GetMetaType retrieves the type and package defined by the struct:field:type
// metadata if any.
func GetMetaType(att *expr.AttributeExpr) (typeName string, importS *ImportSpec) {
	if att == nil {
		return
	}
	if args, ok := att.Meta["struct:field:type"]; ok {
		if len(args) > 0 {
			typeName = args[0]
		}
		if len(args) > 1 {
			importS = &ImportSpec{Path: args[1]}
		}
		if len(args) > 2 {
			importS.Name = args[2]
		}
	}
	return
}

// GetMetaTypeImports parses the attribute for all user defined imports
func GetMetaTypeImports(att *expr.AttributeExpr) []*ImportSpec {
	return safelyGetMetaTypeImports(att, nil)
}

// safelyGetMetaTypeImports parses attributes while keeping track of previous usertypes to avoid infinite recursion
func safelyGetMetaTypeImports(att *expr.AttributeExpr, seen map[string]struct{}) []*ImportSpec {
	if att == nil {
		return nil
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	uniqueImports := make(map[ImportSpec]struct{})
	imports := make([]*ImportSpec, 0)

	switch t := att.Type.(type) {
	case expr.UserType:
		if _, wasSeen := seen[t.ID()]; wasSeen {
			return imports
		}
		seen[t.ID()] = struct{}{}
		for _, im := range safelyGetMetaTypeImports(t.Attribute(), seen) {
			if im != nil {
				uniqueImports[*im] = struct{}{}
			}
		}
	case *expr.Array:
		for _, im := range safelyGetMetaTypeImports(t.ElemType, seen) {
			if im != nil {
				uniqueImports[*im] = struct{}{}
			}
		}
	case *expr.Map:
		for _, im := range safelyGetMetaTypeImports(t.ElemType, seen) {
			if im != nil {
				uniqueImports[*im] = struct{}{}
			}
		}
		for _, im := range safelyGetMetaTypeImports(t.KeyType, seen) {
			if im != nil {
				uniqueImports[*im] = struct{}{}
			}
		}
	case *expr.Object:
		for _, na := range *t {
			for _, im := range safelyGetMetaTypeImports(na.Attribute, seen) {
				if im != nil {
					uniqueImports[*im] = struct{}{}
				}
			}
		}
	}
	_, im := GetMetaType(att)
	if im != nil {
		uniqueImports[*im] = struct{}{}
	}
	for imp := range uniqueImports {
		// Copy loop variable into body so next iteration doesn't overwrite its address https://stackoverflow.com/questions/27610039/golang-appending-leaves-only-last-element
		copy := imp
		imports = append(imports, &copy)
	}
	return imports
}

// AddServiceMetaTypeImports adds meta type imports for each method of the service expr
func AddServiceMetaTypeImports(header *SectionTemplate, svc *expr.ServiceExpr) {
	for _, m := range svc.Methods {
		AddImport(header, GetMetaTypeImports(m.Payload)...)
		AddImport(header, GetMetaTypeImports(m.StreamingPayload)...)
		AddImport(header, GetMetaTypeImports(m.Result)...)
	}
}
