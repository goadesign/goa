package dsl

import (
	"fmt"
	"mime"
	"strings"

	apidesign "github.com/goadesign/goa/design"
	apidsl "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rest/design"
)

// Counter used to create unique media type names for identifier-less media types.
var mediaTypeCount int

// MediaType describes the representation of a resource used in a response body.
//
// Media types are defined with a unique identifier as described in RFC6838. The identifier also
// defines the default value for the Content-Type header of responses. The ContentType function
// allows overridding the default as shown in the example below.
//
// The media type definition includes a listing of all the potential attributes that can appear in
// the body. Views specify which of the attributes are actually rendered so that the same media type
// definition may represent multiple rendering of a given resource representation.
//
// All media types must define a view named "default". This view is used to render the media type in
// response bodies when no other view is specified.
//
// A media type definition may also define links to other media types. This is done by first
// defining an attribute for the linked-to media type and then referring to that attribute in the
// Links function. Views may elect to render the attribute, the links or both. Linked-to media types
// are rendered using the "link" view by default and must thus define that view.
//
// MediaType accepts two arguments: the media type identifier and the defining DSL.
//
// MediaType is a top level DSL.
//
// Example:
//
//    var BottleMT = MediaType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleMedia")         // Override default generated name
//        ContentType("application/json") // Override default Content-Type header
//        Attributes(func() {
//            Attribute("id", Integer, "ID of bottle")
//            Attribute("href", String, "API href of bottle")
//            Attribute("account", Account, "Owner account")
//            Attribute("origin", Origin, "Details on wine origin")
//            Links(func() {
//                Link("account")         // Define link to Account media type
//                Link("origin", "tiny")  // View used to render link
//            })
//            Required("id", "href")
//        })
//        View("default", func() {        // Define default view
//            Attribute("id")
//            Attribute("href")
//            Attribute("links")          // Render links
//        })
//        View("extended", func() {       // Define "extended" view
//            Attribute("id")
//            Attribute("href")
//            Attribute("account")        // Render account inline
//            Attribute("origin")         // Render origin inline
//            Attribute("links")          // Render links
//        })
//     })
//
func MediaType(identifier string, adsl func()) *design.MediaTypeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	// Validate Media Type
	identifier, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		eval.ReportError("invalid media type identifier %#v: %s",
			identifier, err)
		// We don't return so that other errors may be
		// captured in this one run.
		identifier = "text/plain"
	}
	canonicalID := design.CanonicalIdentifier(identifier)
	// Validate that media type identifier doesn't clash
	if m := design.Root.MediaType(canonicalID); m != nil {
		eval.ReportError("media type %#v with canonical identifier %#v is defined twice", identifier, canonicalID)
		return nil
	}
	identifier = mime.FormatMediaType(identifier, params)
	lastPart := identifier
	lastPartIndex := strings.LastIndex(identifier, "/")
	if lastPartIndex > -1 {
		lastPart = identifier[lastPartIndex+1:]
	}
	plusIndex := strings.Index(lastPart, "+")
	if plusIndex > 0 {
		lastPart = lastPart[:plusIndex]
	}
	lastPart = strings.TrimPrefix(lastPart, "vnd.")
	elems := strings.Split(lastPart, ".")
	for i, e := range elems {
		elems[i] = strings.Title(e)
	}
	typeName := strings.Join(elems, "")
	if typeName == "" {
		mediaTypeCount++
		typeName = fmt.Sprintf("MediaType%d", mediaTypeCount)
	}
	// Now save the type in the API media types map
	mt := design.NewMediaTypeExpr(typeName, identifier, adsl)
	design.Root.MediaTypes = append(design.Root.MediaTypes, mt)
	return mt
}

// Media defines the media type used to render the response body.
//
// Media may appear in a Response or a ResponseTemplate DSL.
//
// Media accepts one or two arguments. The first argument is the media type or its identifier. The
// second optional argument is the name of the view used to render the response if always the same.
// When using an identifier to specify the media type the value may or may not match a media type
// defined in the design. If it does then the effect is the same as using the media type instance.
//
// Examples:
//
//	Response("OK", func() {
//		Status(200)
//		Media(BottleMedia)
//	})
//
// Is equivalent to:
//
//	Response("OK", func() {
//		Status(200)
//		Media("application/vnd.bottle")
//	})
//
// (assuming the BottleMedia media type identifier is "application/vnd.bottle")
// This is valid, the generated code uses the media type identifier to set the response Content-Type
// header:
//
//	Response("NotFound", func() {
//		Status(400)
//		Media("application/json")
//	})
//
func Media(val interface{}, viewName ...string) {
	if r, ok := eval.Current().(*design.ResponseExpr); ok {
		if m, ok := val.(*design.MediaTypeExpr); ok {
			if m != nil {
				r.MediaType = m.Identifier
			}
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			eval.ReportError("media type must be a string or a pointer to MediaTypeExpr, got %#v", val)
		}
		if len(viewName) == 1 {
			r.ViewName = viewName[0]
		} else if len(viewName) > 1 {
			eval.ReportError("too many arguments given to DefaultMedia")
		}
	} else {
		eval.IncompatibleDSL()
	}
}

// Reference sets a type or media type reference. The value itself can be a type or a media type.
// The reference type attributes define the default properties for attributes with the same name in
// the type using the reference. So for example if a type is defined as such:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name", func() {
//			MinLength(3)
//		})
//		Attribute("vintage", Integer, func() {
//			Minimum(1970)
//		})
//		Attribute("somethingelse")
//	})
//
// Declaring the following media type:
//
//	var BottleMedia = MediaType("vnd.goa.bottle", func() {
//		Reference(Bottle)
//		Attributes(func() {
//			Attribute("id", Integer)
//			Attribute("name")
//			Attribute("vintage")
//		})
//	})
//
// defines the "name" and "vintage" attributes with the same type and validations as defined in
// the Bottle type.
func Reference(t apidesign.DataType) {
	switch def := eval.Current().(type) {
	case *design.MediaTypeExpr:
		def.Reference = t
	case *apidesign.AttributeExpr:
		def.Reference = t
	default:
		eval.IncompatibleDSL()
	}
}

// TypeName makes it possible to set the Go struct name for a type or media type in the generated
// code. By default goagen uses the name (type) or identifier (media type) given in the adsl and
// computes a valid Go identifier from it. This function makes it possible to override that and
// provide a custom name. name must be a valid Go identifier.
func TypeName(name string) {
	switch def := eval.Current().(type) {
	case *design.MediaTypeExpr:
		def.TypeName = name
	case *apidesign.UserTypeExpr:
		def.TypeName = name
	default:
		eval.IncompatibleDSL()
	}
}

// ContentType sets the value of the Content-Type response header. By default the ID of the media
// type is used.
//
//    ContentType("application/json")
//
func ContentType(typ string) {
	if mt, ok := eval.Current().(*design.MediaTypeExpr); ok {
		mt.ContentType = typ
	} else {
		eval.IncompatibleDSL()
	}
}

// View adds a new view to a media type. A view has a name and lists attributes that are
// rendered when the view is used to produce a response. The attribute names must appear in the
// media type definition. If an attribute is itself a media type then the view may specify which
// view to use when rendering the attribute using the View function in the View adsl. If not
// specified then the view named "default" is used. Examples:
//
//	View("default", func() {
//		Attribute("id")		// "id" and "name" must be media type attributes
//		Attribute("name")
//	})
//
//	View("extended", func() {
//		Attribute("id")
//		Attribute("name")
//		Attribute("origin", func() {
//			View("extended")	// Use view "extended" to render attribute "origin"
//		})
//	})
func View(name string, adsl ...func()) {
	switch def := eval.Current().(type) {
	case *design.MediaTypeExpr:
		mt := def

		if mt.Views == nil {
			mt.Views = make(map[string]*design.ViewExpr)
		} else {
			if _, ok := mt.Views[name]; ok {
				eval.ReportError("multiple definitions for view %#v in media type %#v", name, mt.TypeName)
				return
			}
		}
		at := &apidesign.AttributeExpr{}
		ok := false
		if len(adsl) > 0 {
			ok = eval.Execute(adsl[0], at)
		} else if a, ok := mt.Type.(*apidesign.Array); ok {
			// inherit view from collection element if present
			elem := a.ElemType
			if elem != nil {
				if pa, ok2 := elem.Type.(*design.MediaTypeExpr); ok2 {
					if v, ok2 := pa.Views[name]; ok2 {
						at = v.AttributeExpr
						ok = true
					} else {
						eval.ReportError("unknown view %#v", name)
						return
					}
				}
			}
		}
		if ok {
			view, err := buildView(name, mt, at)
			if err != nil {
				eval.ReportError(err.Error())
				return
			}
			mt.Views[name] = view
		}

	case *apidesign.AttributeExpr:
		def.Metadata["view"] = []string{name}

	default:
		eval.IncompatibleDSL()
	}
}

// buildView builds a view definition given an attribute and a corresponding media type.
func buildView(name string, mt *design.MediaTypeExpr, at *apidesign.AttributeExpr) (*design.ViewExpr, error) {
	if at.Type == nil {
		return nil, fmt.Errorf("invalid view DSL")
	}
	o, ok := at.Type.(apidesign.Object)
	if !ok {
		return nil, fmt.Errorf("invalid view DSL")
	}
	if o != nil {
		mto, ok := mt.Type.(apidesign.Object)
		if !ok {
			mto = mt.Type.(*apidesign.Array).ElemType.Type.(apidesign.Object)
		}
		for n, cat := range o {
			if existing, ok := mto[n]; ok {
				dup := apidesign.DupAtt(existing)
				dup.Metadata["view"] = cat.Metadata["view"]
				o[n] = dup
			} else if n != "links" {
				return nil, fmt.Errorf("unknown attribute %#v", n)
			}
		}
	}
	return &design.ViewExpr{
		AttributeExpr: at,
		Name:          name,
		Parent:        mt,
	}, nil
}

// Attributes implements the media type attributes adsl. See MediaType.
func Attributes(adsl func()) {
	mt, ok := eval.Current().(*design.MediaTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	eval.Execute(adsl, mt)
}

// Links implements the media type links adsl. See MediaType.
func Links(adsl func()) {
	mt, ok := eval.Current().(*design.MediaTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	eval.Execute(adsl, mt)
}

// Link adds a link to a media type. At the minimum a link has a name corresponding to one of the
// media type attribute names. A link may also define the view used to render the linked-to
// attribute. The default view used to render links is "link". Examples:
//
//	Link("origin")		// Use the "link" view of the "origin" attribute
//	Link("account", "tiny")	// Use the "tiny" view of the "account" attribute
func Link(name string, view ...string) {
	mt, ok := eval.Current().(*design.MediaTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if mt.Links == nil {
		mt.Links = make(map[string]*design.LinkExpr)
	} else {
		if _, ok := mt.Links[name]; ok {
			eval.ReportError("duplicate definition for link %#v", name)
			return
		}
	}
	link := &design.LinkExpr{Name: name, Parent: mt}
	if len(view) > 1 {
		eval.ReportError("invalid syntax in Link definition for %#v, allowed syntax is Link(name) or Link(name, view)", name)
	}
	if len(view) > 0 {
		link.View = view[0]
	} else {
		link.View = "link"
	}
	mt.Links[name] = link
}

// CollectionOf creates a collection media type from its element media type. A collection media
// type represents the content of responses that return a collection of resources such as "list"
// actions. This function can be called from any place where a media type can be used.
// The resulting media type identifier is built from the element media type by appending the media
// type parameter "type" with value "collection".
func CollectionOf(v interface{}, adsl ...func()) *design.MediaTypeExpr {
	var m *design.MediaTypeExpr
	var ok bool
	m, ok = v.(*design.MediaTypeExpr)
	if !ok {
		if id, ok := v.(string); ok {
			m = design.Root.MediaType(id)
		}
	}
	if m == nil {
		eval.ReportError("invalid CollectionOf argument: not a media type and not a known media type identifier")
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewMediaTypeExpr("InvalidCollection", "text/plain", nil)
	}
	id := m.Identifier
	mediatype, params, err := mime.ParseMediaType(id)
	if err != nil {
		eval.ReportError("invalid media type identifier %#v: %s", id, err)
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewMediaTypeExpr("InvalidCollection", "text/plain", nil)
	}
	hasType := false
	for param := range params {
		if param == "type" {
			hasType = true
			break
		}
	}
	if !hasType {
		params["type"] = "collection"
	}
	id = mime.FormatMediaType(mediatype, params)
	canonical := design.CanonicalIdentifier(id)
	if mt := design.Root.GeneratedMediaType(canonical); mt != nil {
		// Already have a type for this collection, reuse it.
		return mt
	}
	mt := design.NewMediaTypeExpr("", id, func() {
		mt, ok := eval.Current().(*design.MediaTypeExpr)
		if !ok {
			eval.IncompatibleDSL()
			return
		}
		// Cannot compute collection type name before element media type DSL has executed
		// since the DSL may modify element type name via the TypeName function.
		mt.TypeName = m.TypeName + "Collection"
		mt.AttributeExpr = &apidesign.AttributeExpr{Type: apidsl.ArrayOf(m)}
		if len(adsl) > 0 {
			eval.Execute(adsl[0], mt)
		}
		if mt.Views == nil {
			// If the adsl didn't create any views (or there is no adsl at all)
			// then inherit the views from the collection element.
			mt.Views = make(map[string]*design.ViewExpr)
			for n, v := range m.Views {
				mt.Views[n] = v
			}
		}
	})
	// Do not execute the adsl right away, will be done last to make sure the element adsl has run
	// first.
	design.Root.GeneratedMediaTypes = append(design.Root.GeneratedMediaTypes, mt)
	return mt
}
