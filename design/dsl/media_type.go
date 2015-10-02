package dsl

import (
	"mime"
	"strings"

	"bitbucket.org/pkg/inflect"
)
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
	if Design == nil {
		InitDesign()
	}
	if Design.MediaTypes == nil {
		Design.MediaTypes = make(map[string]*MediaTypeDefinition)
	}
	if topLevelDefinition(false) {
		identifier, ok := val.(string)
		if !ok {
			ReportError("media type identifier must be a string, got %#v", val)
			return nil
		}
		mediatype, _, err := mime.ParseMediaType(identifier)
		if err != nil {
			ReportError("invalid media type identifier %#v: %s",
				identifier, err)
		}
		elems := strings.Split(mediatype, ".")
		var prefix string
		if len(elems) > 1 {
			prefix = elems[len(elems)-2]
		}
		typeName := inflect.Camelize(prefix) + inflect.Camelize(elems[len(elems)-1]) + "Media"
		if _, ok := Design.MediaTypes[identifier]; ok {
			ReportError("media type %#v is defined twice", identifier)
			return nil
		}
		var d func()
		if len(dsl) > 0 {
			d = dsl[0]
		}
		mt := NewMediaTypeDefinition(typeName, identifier, d)
		Design.MediaTypes[identifier] = mt
		return mt
	} else if r, ok := resourceDefinition(false); ok {
		if m, ok := val.(*MediaTypeDefinition); ok {
			if m.UserTypeDefinition == nil {
				ReportError("invalid media type specification, media type is not initialized")
			} else {
				r.MediaType = m.Identifier
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
	return nil
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
// At the minimum a link has a name corresponding to one of the media type attribute names.
// A link may also define the view used to render the link content if different
// from "link".
// Examples:
//
// Link("vendor")
//
// Link("vendor", "view")
//
func Link(name string, view ...string) {
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
		if len(view) > 1 {
			ReportError("invalid syntax in Link definition for %#v, allowed syntax is Link(name) or Link(name, view)", name)
		}
		if len(view) > 0 {
			link.View = view[0]
		} else {
			link.View = "link"
		}
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
	mediatype, params, err := mime.ParseMediaType(id)
	if err != nil {
		ReportError("invalid media type identifier %#v: %s", id, err)
		return nil
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
		id = mime.FormatMediaType(mediatype, params)
	}
	typeName := m.TypeName + "Collection"
	mt := NewMediaTypeDefinition(typeName, id, func() {
		if mt, ok := mediaTypeDefinition(true); ok {
			tempMT := NewMediaTypeDefinition(m.TypeName, m.Identifier, m.DSL)
			if executeDSL(tempMT.DSL, tempMT) {
				mt.Views = tempMT.Views
				mt.Links = tempMT.Links
				mt.TypeName = typeName
				mt.AttributeDefinition = tempMT.AttributeDefinition
				mt.AttributeDefinition.Type = ArrayOf(mt.AttributeDefinition.Type)
			}
		}
	})
	if executeDSL(mt.DSL, mt) {
		Design.MediaTypes[id] = mt
	}
	return mt
}
