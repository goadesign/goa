package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/raphael/goa/design"
)

var (
	// Errors contains the DSL execution errors if any.
	Errors MultiError

	// Global DSL evaluation stack
	ctxStack contextStack

	// generatedMediaTypes contains DSL definitions that were created by the design DSL and
	// need to be executed as a second pass.
	// An example of this are media types defined with CollectionOf: the element media type
	// must be defined first then the definition created by CollectionOf must execute.
	generatedMediaTypes map[string]*design.MediaTypeDefinition
)

type (
	// MultiError collects all DSL errors. It implements error.
	MultiError []*Error

	// Error represents an error that occurred while running the API DSL.
	// It contains the name of the file and line number of where the error
	// occurred as well as the original Go error.
	Error struct {
		GoError error
		File    string
		Line    int
	}

	// DSL evaluation contexts stack
	contextStack []design.DSLDefinition
)

// RunDSL runs all the registered top level DSLs and returns any error.
// This function is called by the client package init.
// goagen creates that function during code generation.
func RunDSL() error {
	if design.Design == nil {
		return nil
	}
	Errors = nil
	// First run the top level API DSL to initialize responses and
	// response templates needed by resources.
	executeDSL(design.Design.DSL, design.Design)
	// The all the versions
	for _, v := range design.Design.Versions {
		executeDSL(v.DSL, v)
	}
	// Then run the user type DSLs
	for _, t := range design.Design.Types {
		executeDSL(t.DSL, t.AttributeDefinition)
	}
	// Then the media type DSLs
	for _, mt := range design.Design.MediaTypes {
		executeDSL(mt.DSL, mt)
	}
	// And now that we have everything the resources.
	for _, r := range design.Design.Resources {
		executeDSL(r.DSL, r)
	}
	// Now execute any generated media type definitions.
	for _, mt := range generatedMediaTypes {
		canonicalID := design.CanonicalIdentifier(mt.Identifier)
		design.Design.MediaTypes[canonicalID] = mt
		executeDSL(mt.DSL, mt)
	}
	generatedMediaTypes = make(map[string]*design.MediaTypeDefinition)

	// Don't attempt to validate syntactically incorrect DSL
	if Errors != nil {
		return Errors
	}

	// Validate DSL
	if err := design.Design.Validate(); err != nil {
		return err
	}
	if Errors != nil {
		return Errors
	}

	// Second pass post-validation does final merges with defaults and base types.
	for _, t := range design.Design.Types {
		finalizeType(t)
	}
	for _, mt := range design.Design.MediaTypes {
		finalizeMediaType(mt)
	}
	for _, r := range design.Design.Resources {
		finalizeResource(r)
	}

	return nil
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) current() design.DSLDefinition {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Error returns the error message.
func (m MultiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (de *Error) Error() (res string) {
	if err := de.GoError; err != nil {
		res = fmt.Sprintf("[%s:%d] %s", de.File, de.Line, err.Error())
	}
	return
}

// executeDSL runs DSL in given evaluation context and returns true if successful.
// It appends to Errors in case of failure (and returns false).
func executeDSL(dsl func(), ctx design.DSLDefinition) bool {
	if dsl == nil {
		return true
	}
	initCount := len(Errors)
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(Errors) <= initCount
}

// finalizeMediaType merges any base type attribute into the media type attributes
func finalizeMediaType(mt *design.MediaTypeDefinition) {
	if mt.Reference != nil {
		if bat := mt.AttributeDefinition; bat != nil {
			mt.AttributeDefinition.Inherit(bat)
		}
	}
}

// finalizeType merges any base type attribute into the type attributes
func finalizeType(ut *design.UserTypeDefinition) {
	if ut.Reference != nil {
		if bat := ut.AttributeDefinition; bat != nil {
			ut.AttributeDefinition.Inherit(bat)
		}
	}
}

// finalizeResource makes the final pass at the resource DSL. This is needed so that the order
// of DSL function calls is irrelevant. For example a resource response may be defined after an
// action refers to it.
func finalizeResource(r *design.ResourceDefinition) {
	r.IterateActions(func(a *design.ActionDefinition) error {
		// 1. Merge response definitions
		for name, resp := range a.Responses {
			if pr, ok := a.Parent.Responses[name]; ok {
				resp.Merge(pr)
			}
			if ar, ok := design.Design.Responses[name]; ok {
				resp.Merge(ar)
			}
			if dr, ok := design.Design.DefaultResponses[name]; ok {
				resp.Merge(dr)
			}
		}
		// 2. Create implicit action parameters for path wildcards that dont' have one
		for _, r := range a.Routes {
			wcs := design.ExtractWildcards(r.FullPath())
			for _, wc := range wcs {
				found := false
				var o design.Object
				if all := a.AllParams(); all != nil {
					o = all.Type.ToObject()
				} else {
					o = design.Object{}
					a.Params = &design.AttributeDefinition{Type: o}
				}
				for n := range o {
					if n == wc {
						found = true
						break
					}
				}
				if !found {
					o[wc] = &design.AttributeDefinition{Type: design.String}
				}
			}
		}
		// 3. Compute QueryParams from Params
		if params := a.Params; params != nil {
			queryParams := params.Dup()
			for _, route := range a.Routes {
				pnames := route.Params()
				for _, pname := range pnames {
					delete(queryParams.Type.ToObject(), pname)
				}
			}
			// (note: we may end up with required attribute names that don't correspond
			// to actual attributes cos' we just deleted them but that's probably OK.)
			a.QueryParams = queryParams
		}
		return nil
	})
}

// incompatibleDSL should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDSL(dslFunc string) {
	elems := strings.Split(dslFunc, ".")
	ReportError("invalid use of %s", elems[len(elems)-1])
}

// invalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func invalidArgError(expected string, actual interface{}) {
	ReportError("cannot use %#v (type %s) as type %s",
		actual, reflect.TypeOf(actual), expected)
}

// ReportError records a DSL error for reporting post DSL execution.
func ReportError(fm string, vals ...interface{}) {
	var suffix string
	if cur := ctxStack.current(); cur != nil {
		if ctx := cur.Context(); ctx != "" {
			suffix = fmt.Sprintf(" in %s", ctx)
		}
	} else {
		suffix = " (top level)"
	}
	err := fmt.Errorf(fm+suffix, vals...)
	file, line := computeErrorLocation()
	Errors = append(Errors, &Error{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go".
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
func computeErrorLocation() (file string, line int) {
	depth := 2
	_, file, line, _ = runtime.Caller(depth)
	ok := strings.HasSuffix(file, "_test.go") // Be nice with tests
	if !ok {
		nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
		ok = !nok
	}
	for !ok {
		depth++
		_, file, line, _ = runtime.Caller(depth)
		ok = strings.HasSuffix(file, "_test.go")
		if !ok {
			nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
			ok = !nok
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return
	}
	f, err := filepath.Rel(wd, file)
	if err != nil {
		return
	}
	file = f
	return
}

// topLevelDefinition returns true if the currently evaluated DSL is a root
// DSL (i.e. is not being run in the context of another definition).
func topLevelDefinition(failItNotTopLevel bool) bool {
	top := ctxStack.current() == nil
	if failItNotTopLevel && !top {
		incompatibleDSL(caller())
	}
	return top
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition(failIfNotAction bool) (*design.ActionDefinition, bool) {
	a, ok := ctxStack.current().(*design.ActionDefinition)
	if !ok && failIfNotAction {
		incompatibleDSL(caller())
	}
	return a, ok
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition(failIfNotAPI bool) (*design.APIDefinition, bool) {
	a, ok := ctxStack.current().(*design.APIDefinition)
	if !ok && failIfNotAPI {
		incompatibleDSL(caller())
	}
	return a, ok
}

// versionDefinition returns true and current context if it is an APIVersionDefinition,
// nil and false otherwise.
func versionDefinition(failIfNotVersion bool) (*design.APIVersionDefinition, bool) {
	a, ok := ctxStack.current().(*design.APIVersionDefinition)
	if !ok && failIfNotVersion {
		incompatibleDSL(caller())
	}
	return a, ok
}

// contactDefinition returns true and current context if it is an ContactDefinition,
// nil and false otherwise.
func contactDefinition(failIfNotContact bool) (*design.ContactDefinition, bool) {
	a, ok := ctxStack.current().(*design.ContactDefinition)
	if !ok && failIfNotContact {
		incompatibleDSL(caller())
	}
	return a, ok
}

// licenseDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func licenseDefinition(failIfNotLicense bool) (*design.LicenseDefinition, bool) {
	l, ok := ctxStack.current().(*design.LicenseDefinition)
	if !ok && failIfNotLicense {
		incompatibleDSL(caller())
	}
	return l, ok
}

// docsDefinition returns true and current context if it is a DocsDefinition,
// nil and false otherwise.
func docsDefinition(failIfNotDocs bool) (*design.DocsDefinition, bool) {
	a, ok := ctxStack.current().(*design.DocsDefinition)
	if !ok && failIfNotDocs {
		incompatibleDSL(caller())
	}
	return a, ok
}

// mediaTypeDefinition returns true and current context if it is a MediaTypeDefinition,
// nil and false otherwise.
func mediaTypeDefinition(failIfNotMT bool) (*design.MediaTypeDefinition, bool) {
	m, ok := ctxStack.current().(*design.MediaTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDSL(caller())
	}
	return m, ok
}

// typeDefinition returns true and current context if it is a UserTypeDefinition,
// nil and false otherwise.
func typeDefinition(failIfNotMT bool) (*design.UserTypeDefinition, bool) {
	m, ok := ctxStack.current().(*design.UserTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDSL(caller())
	}
	return m, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition(failIfNotAttribute bool) (*design.AttributeDefinition, bool) {
	a, ok := ctxStack.current().(*design.AttributeDefinition)
	if !ok && failIfNotAttribute {
		incompatibleDSL(caller())
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition(failIfNotResource bool) (*design.ResourceDefinition, bool) {
	r, ok := ctxStack.current().(*design.ResourceDefinition)
	if !ok && failIfNotResource {
		incompatibleDSL(caller())
	}
	return r, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition(failIfNotResponse bool) (*design.ResponseDefinition, bool) {
	r, ok := ctxStack.current().(*design.ResponseDefinition)
	if !ok && failIfNotResponse {
		incompatibleDSL(caller())
	}
	return r, ok
}

// Name of calling function.
func caller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "<unknown>"
	}
	return runtime.FuncForPC(pc).Name()
}
