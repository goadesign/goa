package dsl

import (
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rest/design"
)

// Resource describes a set of related endpoints, if implementing a REST API then it describes a
// single REST resource.
//
// The resource DSL allows listing the supported resource actions. Each action corresponds to a
// single API endpoint. See Action.
//
// The resource DSL also allows setting the resource default media type. This media
// type is used to render the response body of actions that return the OK response (unless the
// action overrides the default). The default media type also sets the properties of the request
// payload attributes with the same name. See DefaultMedia.
//
// The resource DSL can also specify a parent resource. Defining a parent resources has two effects.
// First, parent resources set the prefix of all resource action paths to the parent resource href.
// Note that actions can override the path using an absolute path (that is a path starting with
// "//").  Second, goa uses the parent resource href coupled with the resource BasePath if any to
// build request paths.
//
// By default goa uses the show action if present to compute a resource href (basically
// concatenating the parent resource href with the base path and show action path). The resource
// definition may specify a canonical action via CanonicalActionName to override that default.
//
// Resource is a top level DSL.
//
// Resource accepts two arguments: the name of the resource and its defining API.
//
// Example:
//
//     var _ = Resource("bottle", func() {
//         Description("A wine bottle")    // Resource description
//         DefaultMedia(BottleMedia)       // Resource default media type if any
//         BasePath("/bottles")            // Common resource action path prefix if any
//         Parent("account")               // Name of parent resource if any
//         CanonicalActionName("get")      // Name of action that used to compute
//                                         // href if not "show"
//         UseTrait("Authenticated")       // Included trait, can appear more than once
//
//         Response(Unauthorized, ErrorMedia) // Common responses to all actions
//         Response(BadRequest, ErrorMedia)   // can appear more than once
//
//         Action("create", func() {       // Action definition
//             // ... Action DSL
//         })
//
//         Action("get", func() {
//             // ... Action DSL
//         })
//     })
//
func Resource(name string, dsl func()) *design.ResourceExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if res := design.Root.Resource(name); res != nil {
		eval.ReportError("resource %#v is defined twice", name)
		return nil
	}
	resource := design.NewResourceExpr(name, dsl)
	design.Root.Resources = append(design.Root.Resources, resource)
	return resource
}
