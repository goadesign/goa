package rest

import (
	"fmt"
	"strings"

	"goa.design/goa.v2/design"
)

type (
	// FileServerExpr defines an endpoint that servers static assets.
	FileServerExpr struct {
		// Resource is the parent resource.
		Resource *ResourceExpr
		// Description for docs
		Description string
		// Docs points to the service external documentation
		Docs *design.DocsExpr
		// FilePath is the file path to the static asset(s)
		FilePath string
		// RequestPath is the HTTP path that servers the assets.
		RequestPath string
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}

	// FileServerWalker is the type of functions given to WalkFileServers.
	FileServerWalker func(*FileServerExpr) error
)

// EvalName returns the generic definition name used in error messages.
func (f *FileServerExpr) EvalName() string {
	suffix := fmt.Sprintf("file server %s", f.FilePath)
	var prefix string
	if f.Resource != nil {
		prefix = f.Resource.EvalName() + " "
	}
	return prefix + suffix
}

// Finalize normalizes the request path.
func (f *FileServerExpr) Finalize() {
	// Make sure request path starts with a "/" so codegen can rely on it.
	if !strings.HasPrefix(f.RequestPath, "/") {
		f.RequestPath = "/" + f.RequestPath
	}
}

// IsDir returns true if the file server serves a directory, false otherwise.
func (f *FileServerExpr) IsDir() bool {
	return WildcardRegex.MatchString(f.RequestPath)
}
