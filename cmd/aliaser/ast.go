package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

// ExportedFunc contains the details needed to alias an exported function.
type ExportedFunc struct {
	// Name is the name of the function.
	Name string
	// Comment is the function header comment.
	Comment string
	// Declaration is the function signature.
	Declaration string
	// Return is true if the function returns a value.
	Return bool
	// Call contains the code that calls the function.
	Call string
}

// PackageDecl represents a package import declaration.
type PackageDecl struct {
	// Name is the local package if not the default, empty string otherwise.
	Name string
	// ImportPath is the package import path.
	ImportPath string
}

// LocalName returns the package local name.
func (p PackageDecl) LocalName() string {
	if p.Name != "" {
		return p.Name
	}
	elems := strings.Split(p.ImportPath, "/")
	return elems[len(elems)-1]
}

// ParseFuncs parses the Go package at the given path and returns the list of
// exported functions indexed by name as well as the list of dependent packages
// indexed by local name.
func ParseFuncs(pkgPath string) (map[string]*ExportedFunc, map[string]*PackageDecl, error) {
	var (
		fset *token.FileSet
		p    *ast.Package
	)
	{
		fset = token.NewFileSet()
		f, err := parser.ParseDir(fset, pkgPath, noTestFilter, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		if len(f) > 1 {
			return nil, nil, fmt.Errorf("found multiple package declarations in '%s'", pkgPath)
		}
		for _, pp := range f {
			p = pp
		}
	}

	var (
		funcs  = make(map[string]*ExportedFunc)
		imprts = make(map[string]*PackageDecl)
		used   = make(map[string]struct{})
	)
	for _, file := range p.Files {
		if strings.HasSuffix(file.Name.String(), "_test") {
			continue
		}
		for _, p := range file.Imports {
			path := strings.Trim(p.Path.Value, `"`)
			var name string
			if p.Name != nil {
				name = p.Name.String()
			}
			decl, ok := imprts[path]
			if ok {
				if decl.Name != name {
					return nil, nil, fmt.Errorf("package %q is imported using different names in different files (%q and %q), packages must be imported using the same local name in all files", path, name, imprts[path].Name)
				}
			} else {
				decl := &PackageDecl{ImportPath: path, Name: name}
				imprts[decl.LocalName()] = decl
			}
		}

		for _, decl := range file.Decls {
			fdecl, ok := decl.(*ast.FuncDecl)
			if !ok || !fdecl.Name.IsExported() || fdecl.Recv != nil {
				continue
			}
			ef, err := newExportedFunc(fset, fdecl)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create exported function %s", fdecl.Name)
			}
			funcs[fdecl.Name.String()] = ef

			// If result if of the form *package.Type then record the
			// value of package.
			if fdecl.Type.Results != nil {
				for _, field := range fdecl.Type.Results.List {
					if s, ok := field.Type.(*ast.StarExpr); ok {
						if t, ok := s.X.(*ast.SelectorExpr); ok {
							if i, ok := t.X.(*ast.Ident); ok {
								used[i.Name] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	for n := range imprts {
		if _, ok := used[n]; !ok {
			delete(imprts, n)
		}
	}

	return funcs, imprts, nil
}

// noTestFilter returns true if the name of the given file finished by "_test.go"
func noTestFilter(f os.FileInfo) bool {
	return !strings.HasSuffix(f.Name(), "_test.go")
}

// newExportedFunc creates a ExportedFunc from a parsed function.
func newExportedFunc(fset *token.FileSet, decl *ast.FuncDecl) (*ExportedFunc, error) {
	var (
		com, dcl, call string
		ret            bool
		err            error
	)
	{
		if decl.Doc == nil {
			fmt.Printf("WARN: %s - Missing comment\n", decl.Name.String())
		} else {
			if com, err = text(fset, decl.Doc.Pos(), decl.Doc.End()); err != nil {
				return nil, err
			}
		}
		if dcl, err = text(fset, decl.Type.Pos(), decl.Type.End()); err != nil {
			return nil, err
		}
		ret = decl.Type.Results != nil
		call = decl.Name.String() + "("
		var params []string
		for _, p := range decl.Type.Params.List {
			_, isEllipsis := p.Type.(*ast.Ellipsis)
			for _, n := range p.Names {
				t := n.String()
				if isEllipsis {
					t += "..."
				}
				params = append(params, t)
			}
		}
		call += strings.Join(params, ", ")
		call += ")"
	}

	return &ExportedFunc{
		Name:        decl.Name.String(),
		Comment:     com,
		Declaration: dcl,
		Return:      ret,
		Call:        call,
	}, nil
}

// text extracts the text contained betwee start and end in the fset file set.
func text(fset *token.FileSet, start, end token.Pos) (string, error) {
	var (
		f           = fset.File(start)
		startOffset = f.Offset(start)
		endOffset   = f.Offset(end)
	)
	// Let OS do the caching
	byts, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %s", f.Name(), err)
	}
	return string(byts[startOffset:endOffset]), nil
}
