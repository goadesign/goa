package expr

import (
	"fmt"
	"path"
	"strings"
)

type (
	// HTTPFileServerExpr defines an endpoint that serves static assets
	// through HTTP.
	HTTPFileServerExpr struct {
		// Service is the parent service.
		Service *HTTPServiceExpr
		// Description for docs
		Description string
		// Docs points to the service external documentation
		Docs *DocsExpr
		// FilePath is the file path to the static asset(s)
		FilePath string
		// RequestPaths is the list of HTTP paths that serve the assets.
		RequestPaths []string
		// Meta is a list of key/value pairs
		Meta MetaExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (f *HTTPFileServerExpr) EvalName() string {
	suffix := fmt.Sprintf("file server %s", f.FilePath)
	var prefix string
	if f.Service != nil {
		prefix = f.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Finalize normalizes the request path.
func (f *HTTPFileServerExpr) Finalize() {
	current := f.RequestPaths[0]
	paths := f.Service.Paths
	if len(paths) == 0 {
		paths = []string{"/"}
	}
	f.RequestPaths = make([]string, len(paths))
	for i, sp := range paths {
		p := path.Join(Root.API.HTTP.Path, sp, current)
		// Make sure request path starts with a "/" so codegen can rely on it.
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		f.RequestPaths[i] = p
	}
}

// IsDir returns true if the file server serves a directory, false otherwise.
func (f *HTTPFileServerExpr) IsDir() bool {
	return HTTPWildcardRegex.MatchString(f.RequestPaths[0])
}
