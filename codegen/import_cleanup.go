package codegen

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type (
	// importInfo tracks information about an import and whether it's used.
	// Used by the optimized import cleanup algorithm.
	importInfo struct {
		spec      *ast.ImportSpec
		path      string
		localName string
		used      bool
	}
)

// buildImportMap creates a map of package names to import information.
// It handles standard imports, named imports, and excludes blank/dot imports.
func buildImportMap(file *ast.File) map[string]*importInfo {
	imports := make(map[string]*importInfo)

	for _, impDecl := range file.Imports {
		path := strings.Trim(impDecl.Path.Value, `"`)

		// Handle imports with explicit names
		if impDecl.Name != nil {
			switch impDecl.Name.Name {
			case "_", ".":
				// Blank imports (for side effects) and dot imports are always kept
				continue
			default:
				// Named import: use the alias as the local name
				imports[impDecl.Name.Name] = &importInfo{
					spec:      impDecl,
					path:      path,
					localName: impDecl.Name.Name,
					used:      false,
				}
			}
		} else {
			// Standard import: infer package name from path
			localName := inferPackageName(path)
			imports[localName] = &importInfo{
				spec:      impDecl,
				path:      path,
				localName: localName,
				used:      false,
			}
		}
	}

	return imports
}

// detectUsedImports performs a single AST walk to mark which imports are used.
// It looks for qualified identifiers (pkg.Name).
//
// Important: Do NOT treat identifiers that are part of import specs as usage,
// otherwise named imports will be falsely marked as used.
func detectUsedImports(file *ast.File, imports map[string]*importInfo) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			// Skip traversing into import specs to avoid counting alias idents
			// as usage.
			return false
		case *ast.SelectorExpr:
			// Handle qualified identifiers like fmt.Println, pkg.Type, etc.
			if ident, ok := x.X.(*ast.Ident); ok {
				if imp, exists := imports[ident.Name]; exists {
					imp.used = true
				}
			}
		}

		return true
	})
}

// removeUnusedImports deletes import specs that weren't marked as used.
func removeUnusedImports(fset *token.FileSet, file *ast.File, imports map[string]*importInfo) {
	for _, imp := range imports {
		if !imp.used {
			if imp.spec.Name != nil {
				astutil.DeleteNamedImport(fset, file, imp.spec.Name.Name, imp.path)
			} else {
				astutil.DeleteImport(fset, file, imp.path)
			}
		}
	}
}

// inferPackageName extracts the package name from an import path.
//
// Examples:
//   - "fmt" -> "fmt"
//   - "github.com/foo/bar" -> "bar"
//   - "gopkg.in/yaml.v2" -> "yaml"
func inferPackageName(path string) string {
	// Get the last component of the path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		path = path[idx+1:]
	}

	// Remove version suffixes like .v2, .v3, etc.
	if idx := strings.Index(path, ".v"); idx >= 0 {
		path = path[:idx]
	}

	// Remove other suffixes after dots (less common)
	if idx := strings.Index(path, "."); idx >= 0 {
		// Only if it looks like a version or special suffix
		suffix := path[idx:]
		if len(suffix) > 1 && (suffix[1] >= '0' && suffix[1] <= '9' || suffix[1] == 'v') {
			path = path[:idx]
		}
	}

	return path
}
