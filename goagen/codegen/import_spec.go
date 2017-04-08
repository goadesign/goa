package codegen

import (
	"fmt"

	"github.com/goadesign/goa/design"
)

// ImportSpec defines a generated import statement.
type ImportSpec struct {
	Name string
	Path string
}

// NewImport creates an import spec.
func NewImport(name, path string) *ImportSpec {
	return &ImportSpec{Name: name, Path: path}
}

// SimpleImport creates an import with no explicit path component.
func SimpleImport(path string) *ImportSpec {
	return &ImportSpec{Path: path}
}

// Code returns the Go import statement for the ImportSpec.
func (s *ImportSpec) Code() string {
	if len(s.Name) > 0 {
		return fmt.Sprintf(`%s "%s"`, s.Name, s.Path)
	}
	return fmt.Sprintf(`"%s"`, s.Path)
}

// AttributeImports constructs a new ImportsSpec slice from an existing slice and adds in imports specified in
// struct:field:type Metadata tags.
func AttributeImports(att *design.AttributeDefinition, imports []*ImportSpec, seen []*design.AttributeDefinition) []*ImportSpec {

	for _, a := range seen {
		if att == a {
			return imports
		}
	}
	seen = append(seen, att)

	if tname, ok := att.Metadata["struct:field:type"]; ok {
		if len(tname) > 1 {
			tagImp := SimpleImport(tname[1])
			impSlice := []*ImportSpec{tagImp}
			imports = appendImports(imports, impSlice)
		}
	}

	switch t := att.Type.(type) {
	case *design.UserTypeDefinition:
		return appendImports(imports, AttributeImports(t.AttributeDefinition, imports, seen))
	case *design.MediaTypeDefinition:
		return appendImports(imports, AttributeImports(t.AttributeDefinition, imports, seen))
	case design.Object:
		t.IterateAttributes(func(n string, t *design.AttributeDefinition) error {
			imports = appendImports(imports, AttributeImports(t, imports, seen))
			return nil
		})
		return imports
	case *design.Array:
		return appendImports(imports, AttributeImports(t.ElemType, imports, seen))
	case *design.Hash:
		imports = appendImports(imports, AttributeImports(t.KeyType, imports, seen))
		return appendImports(imports, AttributeImports(t.ElemType, imports, seen))
	}

	return imports
}

// appendImports appends two ImportSpec slices and preserves uniqueness
func appendImports(i, a []*ImportSpec) []*ImportSpec {
	for _, v := range a {
		contains := false
		for _, att := range i {
			if att.Path == v.Path {
				contains = true
				break
			}
		}
		if contains != true {
			i = append(i, v)
		}
	}
	return i
}
