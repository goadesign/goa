package dsl

import "fmt"
import . "github.com/raphael/goa/design"

// MediaType defines a media type DSL.
//
// MediaType("application/vnd.goa.example.bottle", func() {
//	Description("A bottle of wine")
//	Attributes(func() {
//		Attribute("id", Integer, "ID of bottle")
//		Attribute("href", String, "API href of bottle")
//		Attribute("origin", Origin, "Details on wine origin")
//		Links(func() {
//			Link("origin")
//		})
//              Required("href")
//      })
//	View("default", func() {
//		Attribute("id")
//		Attribute("href")
//		Attribute("links")
//	})
// })
//
// MediaType also refers to a media type (by name or by reference):
//
// 	ResponseTemplate("NotFound", func() {
//		Status(404)
//		MediaType("application/json")
//	})
//
// This function returns the newly defined media type in the first mode, nil otherwise.
func MediaType(val interface{}, dsl ...func()) *MediaTypeDefinition {
	var mt *MediaTypeDefinition
	if Design.MediaTypes == nil {
		Design.MediaTypes = make(map[string]*MediaTypeDefinition)
	}
	if _, ok := apiDefinition(false); ok {
		if identifier, ok := val.(string); ok {
			if _, ok := Design.MediaTypes[identifier]; ok {
				ReportError("media type %#v is defined twice", identifier)
				return nil
			}
			mt = &MediaTypeDefinition{
				UserTypeDefinition: &UserTypeDefinition{
					AttributeDefinition: &AttributeDefinition{},
					TypeName:            identifier,
				},
			}
			if len(dsl) > 0 {
				if ok := executeDSL(dsl[0], mt); ok {
					Design.MediaTypes[identifier] = mt
				}
			}
		} else {
			ReportError("media type name must be a string, got %#v", val)
		}
	} else if r, ok := resourceDefinition(false); ok {
		if m, ok := val.(*MediaTypeDefinition); ok {
			if m.UserTypeDefinition == nil {
				ReportError("invalid media type specification, media type is not initialized")
			} else {
				r.MediaType = m.TypeName
			}
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			ReportError("media type must be a string or a *MediaTypeDefinition, got %#v", val)
		}
	} else if r, ok := responseDefinition(true); ok {
		if m, ok := val.(*MediaTypeDefinition); ok {
			if m.UserTypeDefinition == nil {
				ReportError("invalid media type specification, media type is not initialized")
			} else {
				r.MediaType = m.TypeName
			}
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			ReportError("media type must be a string or a *MediaTypeDefinition, got %#v", val)
		}
	}
	return mt
}

// View adds a new view to the media type.
// It takes the view name and the DSL defining it.
// View can also be used to specify the view used to render an attribute.
func View(name string, dsl ...func()) {
	if mt, ok := mediaTypeDefinition(false); ok {
		if mt.Views == nil {
			mt.Views = make(map[string]*ViewDefinition)
		} else {
			if _, ok = mt.Views[name]; ok {
				ReportError("multiple definitions for view %#v in media type %#v", name, mt.TypeName)
			}
		}
		at := &AttributeDefinition{}
		ok := true
		if len(dsl) > 0 {
			ok = executeDSL(dsl[0], at)
		}
		if ok {
			mt.Views[name] = &ViewDefinition{
				AttributeDefinition: at,
				Name:                name,
				Parent:              mt,
			}
		}
	} else if a, ok := attributeDefinition(true); ok {
		a.View = name
	}
}

// Attributes defines the media type attributes DSL.
func Attributes(dsl func()) {
	if mt, ok := mediaTypeDefinition(true); ok {
		executeDSL(dsl, mt)
	}
}

// Links defines the media type links DSL.
func Links(dsl func()) {
	if mt, ok := mediaTypeDefinition(true); ok {
		executeDSL(dsl, mt)
	}
}

// Link defines a media type link DSL.
// At the minimum a link has a name potentially corresponding to one of the
// media type attribute names.
// A link may also define the view used to render the link content if different
// from "link".
// Finally a link can also optionally define the media type used to render its
// content if not the one associated with the attribute of same name.
// Examples:
//
// Link("vendor")
//
// Link("vendor", "view")
//
// Link("vendor", LinkMediaType)
//
// Link("vendor", "view", LinkMediaType)
//
func Link(name string, args ...interface{}) {
	if mt, ok := mediaTypeDefinition(true); ok {
		if mt.Links == nil {
			mt.Links = make(map[string]*LinkDefinition)
		} else {
			if _, ok := mt.Links[name]; ok {
				ReportError("duplicate definition for link %#v", name)
				return
			}
		}
		link := &LinkDefinition{Name: name, Parent: mt}
		var view string
		var lmt *MediaTypeDefinition
		switch len(args) {
		case 0:
			view = "default"
		case 1:
			if v, ok := args[0].(string); ok {
				view = v
			} else {
				if lmt, ok = args[0].(*MediaTypeDefinition); ok {
					view = "default"
				} else {
					ReportError("invalid Link argument, must be string or *MediaTypeDefinition, got %#v", args[0])
					return
				}
			}
		case 2:
			if v, ok := args[0].(string); ok {
				view = v
			} else {
				ReportError("invalid Link argument in first position, must be string, got %#v", args[0])
				return
			}
			if lmt, ok = args[1].(*MediaTypeDefinition); !ok {
				ReportError("invalid Link argument in second position, must be *MediaTypeDefinition, got %#v", args[0])
				return
			}
		default:
			ReportError("invalid Link argument count, must be 0, 1 or 2, got %#v", len(args))
			return
		}
		link.View = view
		link.MediaType = lmt
		mt.Links[name] = link
	}
}

// ArrayOf creates an array from its element type.
func ArrayOf(t DataType) *Array {
	at := AttributeDefinition{Type: t}
	return &Array{ElemType: &at}
}

// CollectionOf creates a collection media type from its element media type.
// A collection media type represents the content of responses that return a
// collection of resources such as "index" actions.
func CollectionOf(m *MediaTypeDefinition) *MediaTypeDefinition {
	id := m.Identifier
	if id != "" {
		id += ";type=collection"
	}
	mat := m.UserTypeDefinition.AttributeDefinition
	at := AttributeDefinition{
		Type:        &Array{ElemType: mat},
		Description: fmt.Sprintf("Collection of %s", mat.Description),
	}
	ut := UserTypeDefinition{
		AttributeDefinition: &at,
		TypeName:            m.UserTypeDefinition.TypeName + "Collection",
	}
	col := MediaTypeDefinition{
		// A media type is a type
		UserTypeDefinition: &ut,
		Identifier:         id,
		Links:              m.Links,
		Views:              m.Views,
	}
	return &col
}
