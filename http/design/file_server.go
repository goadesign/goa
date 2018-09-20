package design

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/design"
)

type (
	// FileServerExpr defines an endpoint that servers static assets.
	FileServerExpr struct {
		// Service is the parent service.
		Service *ServiceExpr
		// Description for docs
		Description string
		// Docs points to the service external documentation
		Docs *design.DocsExpr
		// FilePath is the file path to the static asset(s)
		FilePath string
		// RequestPaths is the list of HTTP paths that serve the assets.
		RequestPaths []string
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (f *FileServerExpr) EvalName() string {
	suffix := fmt.Sprintf("file server %s", f.FilePath)
	var prefix string
	if f.Service != nil {
		prefix = f.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Finalize normalizes the request path.
func (f *FileServerExpr) Finalize() {
	current := f.RequestPaths[0]
	paths := f.Service.Paths
	if len(paths) == 0 {
		paths = []string{"/"}
	}
	f.RequestPaths = make([]string, len(paths))
	for i, sp := range paths {
		p := path.Join(sp, current)
		// Make sure request path starts with a "/" so codegen can rely on it.
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		f.RequestPaths[i] = p
	}
}

// IsDir returns true if the file server serves a directory, false otherwise.
func (f *FileServerExpr) IsDir() bool {
	return design.WildcardRegex.MatchString(f.RequestPaths[0])
}
