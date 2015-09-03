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
	if _, ok := apiDefinition(false); ok {
		if identifier, ok := val.(string); ok {
			if _, ok := Design.MediaTypes[identifier]; ok {
				appendError(fmt.Errorf("media type %s is defined twice", identifier))
				return nil
			}
			mt = &MediaTypeDefinition{UserTypeDefinition: &UserTypeDefinition{Name: identifier}}
			if len(dsl) > 0 {
				if ok := executeDSL(dsl[0], mt); ok {
					Design.MediaTypes[identifier] = mt
				}
			}
		} else {
			appendError(fmt.Errorf("media type name must be a string, got %v", val))
		}
	} else if r, ok := resourceDefinition(false); ok {
		if m, ok := val.(*MediaTypeDefinition); ok {
			r.MediaType = m.Name
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			appendError(fmt.Errorf("media type must be a string or a *MediaTypeDefinition, got %v", val))
		}
	} else if r, ok := responseDefinition(true); ok {
		if m, ok := val.(*MediaTypeDefinition); ok {
			r.MediaType = m.Name
		} else if identifier, ok := val.(string); ok {
			r.MediaType = identifier
		} else {
			appendError(fmt.Errorf("media type must be a string or a *MediaTypeDefinition, got %v", val))
		}
	}
	return mt
}

// View adds a new view to the media type.
func View(name string, dsl func()) {
	if mt, ok := mediaTypeDefinition(true); ok {
		if _, ok = mt.Views[name]; ok {
			appendError(fmt.Errorf("multiple definitions for view %s in media type %s", name, mt.Name))
		}
		v := ViewDefinition{Name: name}
		if ok := executeDSL(dsl, &v); ok {
			mt.Views[name] = &v
		}
	}
}

// Attributes defines the media type attributes DSL.
func Attributes(dsl func()) {
	if mt, ok := mediaTypeDefinition(true); ok {
		executeDSL(dsl, &mt)
	}
}

// Links defines the media type links DSL.
func Links(dsl func()) {
	if mt, ok := mediaTypeDefinition(true); ok {
		executeDSL(dsl, &mt)
	}
}

// Link defines a media type link DSL.
func Link(name string, args ...interface{}) {
	if _, ok := mediaTypeDefinition(true); ok {
		Attribute(name, args...)
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
		Name:                m.UserTypeDefinition.Name,
		Description:         m.UserTypeDefinition.Description,
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
