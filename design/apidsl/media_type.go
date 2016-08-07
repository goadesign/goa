package apidsl

import (
	"fmt"
	"mime"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Counter used to create unique media type names for identifier-less media types.
var mediaTypeCount int

// MediaType implements the media type definition DSL. A media type definition describes the
// representation of a resource used in a response body.
//
// Media types are defined with a unique identifier as defined by RFC6838. The identifier also
// defines the default value for the Content-Type header of responses. The ContentType DSL allows
// overridding the default as shown in the example below.
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
// Links DSL. Views may then elect to render one or the other or both. Links are rendered using the
// special "link" view. Media types that are linked to must define that view. Here is an example
// showing all the possible media type sub-definitions:
//
//    MediaType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleMedia")         // Override default generated name
//        ContentType("application/json") // Override default Content-Type header value
//        Attributes(func() {
//            Attribute("id", Integer, "ID of bottle")
//            Attribute("href", String, "API href of bottle")
//            Attribute("account", Account, "Owner account")
//            Attribute("origin", Origin, "Details on wine origin")
//            Links(func() {
//                Link("account")         // Defines link to Account media type
//                Link("origin", "tiny")  // Set view used to render link if not "link"
//            })
//            Required("id", "href")
//        })
//        View("default", func() {
//            Attribute("id")
//            Attribute("href")
//            Attribute("links")          // Renders links
//        })
//        View("extended", func() {
//            Attribute("id")
//            Attribute("href")
//            Attribute("account")        // Renders account inline
//            Attribute("origin")         // Renders origin inline
//            Attribute("links")          // Renders links
//        })
//     })
//
// This function returns the media type definition so it can be referred to throughout the apidsl.
func MediaType(identifier string, apidsl func()) *design.MediaTypeDefinition {
	if design.Design.MediaTypes == nil {
		design.Design.MediaTypes = make(map[string]*design.MediaTypeDefinition)
	}

	if !dslengine.IsTopLevelDefinition() {
		dslengine.IncompatibleDSL()
		return nil
	}

	// Validate Media Type
	identifier, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		dslengine.ReportError("invalid media type identifier %#v: %s",
			identifier, err)
		// We don't return so that other errors may be
		// captured in this one run.
		identifier = "text/plain"
	}
	canonicalID := design.CanonicalIdentifier(identifier)
	// Validate that media type identifier doesn't clash
	if _, ok := design.Design.MediaTypes[canonicalID]; ok {
		dslengine.ReportError("media type %#v with canonical identifier %#v is defined twice", identifier, canonicalID)
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
	mt := design.NewMediaTypeDefinition(typeName, identifier, apidsl)
	design.Design.MediaTypes[canonicalID] = mt
	return mt
}

// Media sets a response media type by name or by reference using a value returned by MediaType:
//
//	Response("NotFound", func() {
//		Status(404)
//		Media("application/json")
//	})
//
// If Media uses a media type defined in the design then it may optionally specify a view name:
//
//	Response("OK", func() {
//		Status(200)
//		Media(BottleMedia, "tiny")
//	})
//
// Specifying a media type is useful for responses that always return the same view.
//
// Media can be used inside Response or ResponseTemplate.
func Media(val interface{}, viewName ...string) {
	if r, ok := responseDefinition(); ok {
		if m, ok := val.(*design.MediaTypeDefinition); ok {
			if m != nil {
				r.MediaType = m.Identifier
			}
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			dslengine.ReportError("media type must be a string or a pointer to MediaTypeDefinition, got %#v", val)
		}
		if len(viewName) == 1 {
			r.ViewName = viewName[0]
		} else if len(viewName) > 1 {
			dslengine.ReportError("too many arguments given to DefaultMedia")
		}
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
func Reference(t design.DataType) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.MediaTypeDefinition:
		def.Reference = t
	case *design.AttributeDefinition:
		def.Reference = t
	default:
		dslengine.IncompatibleDSL()
	}
}

// TypeName makes it possible to set the Go struct name for a type or media type in the generated
// code. By default goagen uses the name (type) or identifier (media type) given in the apidsl and
// computes a valid Go identifier from it. This function makes it possible to override that and
// provide a custom name. name must be a valid Go identifier.
func TypeName(name string) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.MediaTypeDefinition:
		def.TypeName = name
	case *design.UserTypeDefinition:
		def.TypeName = name
	default:
		dslengine.IncompatibleDSL()
	}
}

// ContentType sets the value of the Content-Type response header. By default the ID of the media
// type is used.
//
//    ContentType("application/json")
//
func ContentType(typ string) {
	if mt, ok := mediaTypeDefinition(); ok {
		mt.ContentType = typ
	}
}

// View adds a new view to a media type. A view has a name and lists attributes that are
// rendered when the view is used to produce a response. The attribute names must appear in the
// media type definition. If an attribute is itself a media type then the view may specify which
// view to use when rendering the attribute using the View function in the View apidsl. If not
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
func View(name string, apidsl ...func()) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.MediaTypeDefinition:
		mt := def

		if !mt.Type.IsObject() && !mt.Type.IsArray() {
			dslengine.ReportError("cannot define view on non object and non collection media types")
			return
		}
		if mt.Views == nil {
			mt.Views = make(map[string]*design.ViewDefinition)
		} else {
			if _, ok := mt.Views[name]; ok {
				dslengine.ReportError("multiple definitions for view %#v in media type %#v", name, mt.TypeName)
				return
			}
		}
		at := &design.AttributeDefinition{}
		ok := false
		if len(apidsl) > 0 {
			ok = dslengine.Execute(apidsl[0], at)
		} else if mt.Type.IsArray() {
			// inherit view from collection element if present
			elem := mt.Type.ToArray().ElemType
			if elem != nil {
				if pa, ok2 := elem.Type.(*design.MediaTypeDefinition); ok2 {
					if v, ok2 := pa.Views[name]; ok2 {
						at = v.AttributeDefinition
						ok = true
					} else {
						dslengine.ReportError("unknown view %#v", name)
						return
					}
				}
			}
		}
		if ok {
			view, err := buildView(name, mt, at)
			if err != nil {
				dslengine.ReportError(err.Error())
				return
			}
			mt.Views[name] = view
		}

	case *design.AttributeDefinition:
		def.View = name

	default:
		dslengine.IncompatibleDSL()
	}
}

// buildView builds a view definition given an attribute and a corresponding media type.
func buildView(name string, mt *design.MediaTypeDefinition, at *design.AttributeDefinition) (*design.ViewDefinition, error) {
	if at.Type == nil || !at.Type.IsObject() {
		return nil, fmt.Errorf("invalid view DSL")
	}
	o := at.Type.ToObject()
	if o != nil {
		mto := mt.Type.ToObject()
		if mto == nil {
			mto = mt.Type.ToArray().ElemType.Type.ToObject()
		}
		for n, cat := range o {
			if existing, ok := mto[n]; ok {
				dup := design.DupAtt(existing)
				dup.View = cat.View
				o[n] = dup
			} else if n != "links" {
				return nil, fmt.Errorf("unknown attribute %#v", n)
			}
		}
	}
	return &design.ViewDefinition{
		AttributeDefinition: at,
		Name:                name,
		Parent:              mt,
	}, nil
}

// Attributes implements the media type attributes apidsl. See MediaType.
func Attributes(apidsl func()) {
	if mt, ok := mediaTypeDefinition(); ok {
		dslengine.Execute(apidsl, mt)
	}
}

// Links implements the media type links apidsl. See MediaType.
func Links(apidsl func()) {
	if mt, ok := mediaTypeDefinition(); ok {
		dslengine.Execute(apidsl, mt)
	}
}

// Link adds a link to a media type. At the minimum a link has a name corresponding to one of the
// media type attribute names. A link may also define the view used to render the linked-to
// attribute. The default view used to render links is "link". Examples:
//
//	Link("origin")		// Use the "link" view of the "origin" attribute
//	Link("account", "tiny")	// Use the "tiny" view of the "account" attribute
func Link(name string, view ...string) {
	if mt, ok := mediaTypeDefinition(); ok {
		if mt.Links == nil {
			mt.Links = make(map[string]*design.LinkDefinition)
		} else {
			if _, ok := mt.Links[name]; ok {
				dslengine.ReportError("duplicate definition for link %#v", name)
				return
			}
		}
		link := &design.LinkDefinition{Name: name, Parent: mt}
		if len(view) > 1 {
			dslengine.ReportError("invalid syntax in Link definition for %#v, allowed syntax is Link(name) or Link(name, view)", name)
		}
		if len(view) > 0 {
			link.View = view[0]
		} else {
			link.View = "link"
		}
		mt.Links[name] = link
	}
}

// CollectionOf creates a collection media type from its element media type. A collection media
// type represents the content of responses that return a collection of resources such as "list"
// actions. This function can be called from any place where a media type can be used.
// The resulting media type identifier is built from the element media type by appending the media
// type parameter "type" with value "collection".
func CollectionOf(v interface{}, apidsl ...func()) *design.MediaTypeDefinition {
	var m *design.MediaTypeDefinition
	var ok bool
	m, ok = v.(*design.MediaTypeDefinition)
	if !ok {
		if id, ok := v.(string); ok {
			m = design.Design.MediaTypes[design.CanonicalIdentifier(id)]
		}
	}
	if m == nil {
		dslengine.ReportError("invalid CollectionOf argument: not a media type and not a known media type identifier")
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewMediaTypeDefinition("InvalidCollection", "text/plain", nil)
	}
	id := m.Identifier
	mediatype, params, err := mime.ParseMediaType(id)
	if err != nil {
		dslengine.ReportError("invalid media type identifier %#v: %s", id, err)
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewMediaTypeDefinition("InvalidCollection", "text/plain", nil)
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
	if mt, ok := design.GeneratedMediaTypes[canonical]; ok {
		// Already have a type for this collection, reuse it.
		return mt
	}
	mt := design.NewMediaTypeDefinition("", id, func() {
		if mt, ok := mediaTypeDefinition(); ok {
			// Cannot compute collection type name before element media type DSL has executed
			// since the DSL may modify element type name via the TypeName function.
			mt.TypeName = m.TypeName + "Collection"
			mt.AttributeDefinition = &design.AttributeDefinition{Type: ArrayOf(m)}
			if len(apidsl) > 0 {
				dslengine.Execute(apidsl[0], mt)
			}
			if mt.Views == nil {
				// If the apidsl didn't create any views (or there is no apidsl at all)
				// then inherit the views from the collection element.
				mt.Views = make(map[string]*design.ViewDefinition)
				for n, v := range m.Views {
					mt.Views[n] = v
				}
			}
		}
	})
	// Do not execute the apidsl right away, will be done last to make sure the element apidsl has run
	// first.
	design.GeneratedMediaTypes[canonical] = mt
	return mt
}
