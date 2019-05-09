package codegen

import (
	"fmt"
	"strconv"

	"goa.design/goa/v3/pkg"
)

// DesignVersion contains the major component of the version of Goa used to
// author the design - either 2 or 3. This value is initialized when the
// generated tool is invoked by retrieving the information passed on the
// command line by the goa tool.
var DesignVersion int = pkg.Major

type (
	// ImportSpec defines a generated import statement.
	ImportSpec struct {
		// Name of imported package if needed.
		Name string
		// Go import path of package.
		Path string
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
