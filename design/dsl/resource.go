package dsl

import "github.com/raphael/goa/design"

// Resource implements the resource definition DSL. There is one resource definition per resource
// exposed by the API. The resource DSL allows setting the resource default media type. This media
// type is used to render the response body of actions that return the OK response (unless the
// action overrides the default). The default media type also sets the properties of the request
// payload attributes with the same name. See DefaultMedia.
//
// The resource DSL also allows listing the supported resource collection and resource collection
// item actions. Each action corresponds to a specific API endpoint. See Action.
//
// The resource DSL can also specify a parent resource. Parent resources have two effects.
// First, they set the prefix of all resource action paths to the parent resource href. Note that
// actions can override the path using an absolute path (that is a path starting with "//").
// Second, goa uses the parent resource href coupled with the resource BasePath if any to build
// hrefs to the resource collection or resource collection items. By default goa uses the show
// action if present to compute a resource href (basically concatenating the parent resource href
// with the base path and show action path). The resource definition may specify a canonical action
// via CanonicalActionName to override that default. Here is an example of a resource definition:
//
//	Resource("bottle", func() {
//		Description("A wine bottle") // Resource description
//		DefaultMedia(BottleMedia)    // Resource default media type
//		BasePath("/bottles")         // Common resource action path prefix if not ""
//		Parent("account")            // Name of parent resource if any
//		CanonicalActionName("get")   // Name of action that returns canonical representation if not "show"
//		UseTrait("Authenticated")    // Included trait if any, can appear more than once
//		APIVersion("v1")	     // API version exposing this resource, can appear more than once.
//
//	 	Action("show", func() {      // Action definition, can appear more than once
//			// ... Action DSL
//		})
//	 })
func Resource(name string, dsl func()) *design.ResourceDefinition {
	if design.Design == nil {
		InitDesign()
	}
	if design.Design.Resources == nil {
		design.Design.Resources = make(map[string]*design.ResourceDefinition)
	}
	var resource *design.ResourceDefinition
	if topLevelDefinition(true) {
		if _, ok := design.Design.Resources[name]; ok {
			ReportError("resource %#v is defined twice", name)
			return nil
		}
		resource = design.NewResourceDefinition(name, dsl)
		design.Design.Resources[name] = resource
	}
	return resource
}

// DefaultMedia sets a resource default media type by identifier or by reference using a value
// returned by MediaType:
//
// 	var _ = Resource("bottle", func() {
// 		DefaultMedia(BottleMedia)
// 		// ...
// 	})
//
// 	var _ = Resource("region", func() {
// 		DefaultMedia("vnd.goa.region")
// 		// ...
// 	})
//
// The default media type is used to build OK response definitions when no specific media type is
// given in the Response function call. The default media type is also used to set the default
// properties of attributes listed in action payloads. So if a media type defines an attribute
// "name" with associated validations then simply calling Attribute("name") inside a request
// Payload defines the payload attribute with the same type and validations.
func DefaultMedia(val interface{}) {
	if r, ok := resourceDefinition(true); ok {
		if m, ok := val.(*design.MediaTypeDefinition); ok {
			if m.UserTypeDefinition == nil {
				ReportError("invalid media type specification, media type is not initialized")
			} else {
				r.MediaType = design.CanonicalIdentifier(m.Identifier)
				m.Resource = r
			}
		} else if identifier, ok := val.(string); ok {
			r.MediaType = design.CanonicalIdentifier(identifier)
		} else {
			ReportError("media type must be a string or a *design.MediaTypeDefinition, got %#v", val)
		}
	}
}

// Parent sets the resource parent. The parent resource is used to compute the path to the resource
// actions as well as resource collection item hrefs. See Resource.
func Parent(p string) {
	if r, ok := resourceDefinition(true); ok {
		r.ParentName = p
	}
}

// CanonicalActionName sets the name of the action used to compute the resource collection and
// resource collection items hrefs. See Resource.
func CanonicalActionName(a string) {
	if r, ok := resourceDefinition(true); ok {
		r.CanonicalActionName = a
	}
}

// APIVersion define the API version(s) that expose this resource.
func APIVersion(versions ...string) {
	if r, ok := resourceDefinition(false); ok {
		r.APIVersions = append(r.APIVersions, versions...)
	} else if m, ok := mediaTypeDefinition(false); ok {
		m.APIVersions = append(m.APIVersions, versions...)
	} else if t, ok := typeDefinition(true); ok {
		t.APIVersions = append(t.APIVersions, versions...)
	}
}
