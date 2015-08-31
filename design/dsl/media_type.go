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
func MediaType(identifier string, dsl func()) *MediaTypeDefinition {
	mt := &MediaTypeDefinition{Identifier: identifier}
	if ok := executeDSL(dsl, mt); ok {
		Design.MediaTypes = append(Design.MediaTypes, mt)
	}

	return mt
}

// View adds a new view to the media type.
func View(name string, dsl func()) {
	if mt, ok := mediaTypeDefinition(); ok {
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
	if mt, ok := mediaTypeDefinition(); ok {
		executeDSL(dsl, &mt)
	}
}

// Links defines the media type links DSL.
func Links(dsl func()) {
	if mt, ok := mediaTypeDefinition(); ok {
		executeDSL(dsl, &mt)
	}
}

// Link defines a media type link DSL.
func Link(name string, args ...interface{}) {
	if _, ok := mediaTypeDefinition(); ok {
		Attribute(name, args...)
	}
}

// FIXME
//
// CollectionOf creates a collection media type from its element media type.
// A collection media type represents the content of responses that return a
// collection of resources such as "index" actions.
//
// func CollectionOf(m *MediaTypeDefinition) *MediaTypeDefinition {
// 	col := *m
// 	col.IsCollection = true
// 	return &col
// }
