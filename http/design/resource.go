package design

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/dimfeld/httppath"
	"github.com/goadesign/goa/design"
)

type (
	// ResourceExpr describes a REST resource.
	// It defines both a media type and a set of actions that can be executed through HTTP
	// requests.
	ResourceExpr struct {
		// Resource name
		Name string
		// Schemes is the supported API URL schemes
		Schemes []string
		// Common URL prefix to all resource action HTTP requests
		BasePath string
		// Path and query string parameters that apply to all actions.
		Params *design.AttributeExpr
		// Name of parent resource if any
		ParentName string
		// Optional description
		Description string
		// Default media type, describes the resource attributes
		MediaType string
		// Default view name if default media type is MediaTypeDefinition
		DefaultViewName string
		// Exposed resource actions indexed by name
		Actions map[string]*ActionExpr
		// FileServers is the list of static asset serving endpoints
		FileServers []*FileServerExpr
		// Action with canonical resource path
		CanonicalActionName string
		// Map of response definitions that apply to all actions indexed by name.
		Responses map[string]*ResponseExpr
		// Request headers that apply to all actions.
		Headers *design.AttributeExpr
		// DSLFunc contains the DSL used to create this definition if any.
		DSLFunc func()
		// metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}

	// ResourceIterator is the type of functions given to IterateResources.
	ResourceIterator func(r *ResourceExpr) error
)

// NewResourceExpr creates a resource definition but does not
// execute the DSL.
func NewResourceExpr(name string, dsl func()) *ResourceExpr {
	return &ResourceExpr{
		Name:      name,
		MediaType: "text/plain",
		DSLFunc:   dsl,
	}
}

// EvalName returns the generic definition name used in error messages.
func (r *ResourceExpr) EvalName() string {
	if r.Name != "" {
		return fmt.Sprintf("resource %#v", r.Name)
	}
	return "unnamed resource"
}

// IterateActions calls the given iterator passing in each resource action sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateActions returns that
// error.
func (r *ResourceExpr) IterateActions(it ActionIterator) error {
	names := make([]string, len(r.Actions))
	i := 0
	for n := range r.Actions {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(r.Actions[n]); err != nil {
			return err
		}
	}
	return nil
}

// IterateFileServers calls the given iterator passing each resource file server sorted by file
// path. Iteration stops if an iterator returns an error and in this case IterateFileServers returns
// that error.
func (r *ResourceExpr) IterateFileServers(it FileServerIterator) error {
	sort.Sort(ByFilePath(r.FileServers))
	for _, f := range r.FileServers {
		if err := it(f); err != nil {
			return err
		}
	}
	return nil
}

// IterateHeaders calls the given iterator passing in each response sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateHeaders returns that
// error.
func (r *ResourceExpr) IterateHeaders(it HeaderIterator) error {
	return iterateHeaders(r.Headers, r.Headers.IsRequired, it)
}

// CanonicalAction returns the canonical action of the resource if any.
// The canonical action is used to compute hrefs to resources.
func (r *ResourceExpr) CanonicalAction() *ActionExpr {
	name := r.CanonicalActionName
	if name == "" {
		name = "show"
	}
	ca, _ := r.Actions[name]
	return ca
}

// URITemplate returns a URI template to this resource.
// The result is the empty string if the resource does not have a "show" action
// and does not define a different canonical action.
func (r *ResourceExpr) URITemplate() string {
	ca := r.CanonicalAction()
	if ca == nil || len(ca.Routes) == 0 {
		return ""
	}
	return ca.Routes[0].FullPath()
}

// FullPath computes the base path to the resource actions concatenating the API and parent resource
// base paths as needed.
func (r *ResourceExpr) FullPath() string {
	if strings.HasPrefix(r.BasePath, "//") {
		return httppath.Clean(r.BasePath)
	}
	var basePath string
	if p := r.Parent(); p != nil {
		if ca := p.CanonicalAction(); ca != nil {
			if routes := ca.Routes; len(routes) > 0 {
				// Note: all these tests should be true at code generation time
				// as DSL validation makes sure that parent resources have a
				// canonical path.
				basePath = path.Join(routes[0].FullPath())
			}
		}
	} else {
		basePath = Root.BasePath
	}
	return httppath.Clean(path.Join(basePath, r.BasePath))
}

// Parent returns the parent resource if any, nil otherwise.
func (r *ResourceExpr) Parent() *ResourceExpr {
	if r.ParentName != "" {
		if parent := Root.Resource(r.ParentName); parent != nil {
			return parent
		}
	}
	return nil
}

// Finalize is run post DSL execution. It merges response definitions, creates implicit action
// parameters, initializes querystring parameters, sets path parameters as non zero attributes
// and sets the fallbacks for security schemes.
func (r *ResourceExpr) Finalize() {
	r.IterateFileServers(func(f *FileServerExpr) error {
		f.Finalize()
		return nil
	})
	r.IterateActions(func(a *ActionExpr) error {
		a.Finalize()
		return nil
	})
}

// ByFilePath makes FileServerExpr sortable for code generators.
type ByFilePath []*FileServerExpr

func (b ByFilePath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByFilePath) Len() int           { return len(b) }
func (b ByFilePath) Less(i, j int) bool { return b[i].FilePath < b[j].FilePath }
