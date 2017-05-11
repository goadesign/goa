package dsl

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Counter used to create unique media type names for identifier-less media
// types.
var mediaTypeCount int

// MediaType defines a media type used to describe an endpoint response.
//
// Media types have a unique identifier as described in RFC6838. The identifier
// defines the default value for the Content-Type header of HTTP responses.
//
// The media type expression includes a listing of all the response attributes.
// Views specify which of the attributes are actually rendered so that the same
// media type expression may represent multiple rendering of a given response.
//
// All media types have a view named "default". This view is used to render the
// media type in responses when no other view is specified. If the default view
// is not explicitly described in the DSL then one is created that lists all the
// media type attributes.
//
// MediaType is a top level DSL.
// MediaType accepts two arguments: the media type identifier and the defining
// DSL.
//
// Example:
//
//    var BottleMT = MediaType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleMedia")         // Override generated type name
//        ContentType("application/json") // Override Content-Type header
//
//        Attributes(func() {
//            Attribute("id", Integer, "ID of bottle")
//            Attribute("href", String, "API href of bottle")
//            Attribute("account", Account, "Owner account")
//            Attribute("origin", Origin, "Details on wine origin")
//            Required("id", "href")
//        })
//
//        View("default", func() {        // Explicitly define default view
//            Attribute("id")
//            Attribute("href")
//        })
//
//        View("extended", func() {       // Define "extended" view
//            Attribute("id")
//            Attribute("href")
//            Attribute("account")
//            Attribute("origin")
//        })
//     })
//
func MediaType(identifier string, fn func()) *design.MediaTypeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	// Validate Media Type
	identifier, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		eval.ReportError("invalid media type identifier %#v: %s",
			identifier, err)
		// We don't return so that other errors may be captured in this
		// one run.
		identifier = "text/plain"
	}
	canonicalID := design.CanonicalIdentifier(identifier)
	// Validate that media type identifier doesn't clash
	if m := design.Root.UserType(canonicalID); m != nil {
		eval.ReportError(
			"media type %#v with canonical identifier %#v is defined twice",
			identifier, canonicalID)
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
	mt := design.NewMediaTypeExpr(typeName, identifier, fn)
	design.Root.MediaTypes = append(design.Root.MediaTypes, mt)

	return mt
}

// TypeName makes it possible to set the Go struct name for a type or media type
// in the generated code. By default goagen uses the name (type) or identifier
// (media type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
func TypeName(name string) {
	switch expr := eval.Current().(type) {
	case *design.MediaTypeExpr:
		expr.TypeName = name
	case *design.UserTypeExpr:
		expr.TypeName = name
	default:
		eval.IncompatibleDSL()
	}
}

// ContentType sets the value of the Content-Type response header. By default
// the ID of the media type is used.
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

// View adds a new view to a media type. A view has a name and lists attributes
// that are rendered when the view is used to produce a response. The attribute
// names must appear in the media type expression. If an attribute is itself a
// media type then the view may specify which view to use when rendering the
// attribute using the View function in the View DSL. If not specified then the
// view named "default" is used. Examples:
//
//	View("default", func() {
//              // "id" and "name" must be media type attributes
//		Attribute("id")
//		Attribute("name")
//	})
//
//	View("extended", func() {
//		Attribute("id")
//		Attribute("name")
//		Attribute("origin", func() {
//			// Use view "extended" to render attribute "origin"
//			View("extended")
//		})
//	})
//
func View(name string, adsl ...func()) {
	switch expr := eval.Current().(type) {
	case *design.MediaTypeExpr:
		mt := expr
		if mt.View(name) != nil {
			eval.ReportError("multiple expressions for view %#v in media type %#v", name, mt.TypeName)
			return
		}
		at := &design.AttributeExpr{}
		ok := false
		if len(adsl) > 0 {
			ok = eval.Execute(adsl[0], at)
		} else if a, ok := mt.Type.(*design.Array); ok {
			// inherit view from collection element if present
			elem := a.ElemType
			if elem != nil {
				if pa, ok2 := elem.Type.(*design.MediaTypeExpr); ok2 {
					if v := pa.View(name); v != nil {
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
			mt.Views = append(mt.Views, view)
		}

	case *design.AttributeExpr:
		if expr.Metadata == nil {
			expr.Metadata = make(map[string][]string)
		}
		expr.Metadata["view"] = []string{name}

	default:
		eval.IncompatibleDSL()
	}
}

// CollectionOf creates a collection media type from its element media type. A
// collection media type represents the content of responses that return a
// collection of values such as listings. The expression accepts an optional DSL
// as second argument that allows specifying which view(s) of the original media
// type apply.
//
// The resulting media type identifier is built from the element media type by
// appending the media type parameter "type" with value "collection".
//
// CollectionOf takes the element media type as first argument and an optional
// DSL as second argument.
// CollectionOf may appear wherever MediaType can.
//
// Example:
//
//     var DivisionResult = MediaType("application/vnd.goa.divresult", func() {
//         Attributes(func() {
//             Attribute("value", Float64)
//         })
//         View("default", func() {
//             Attribute("value")
//         })
//     })
//
//     var MultiResults = CollectionOf(DivisionResult)
//
func CollectionOf(v interface{}, adsl ...func()) *design.MediaTypeExpr {
	var m *design.MediaTypeExpr
	var ok bool
	m, ok = v.(*design.MediaTypeExpr)
	if !ok {
		if id, ok := v.(string); ok {
			if dt := design.Root.UserType(design.CanonicalIdentifier(id)); dt != nil {
				if mt, ok := dt.(*design.MediaTypeExpr); ok {
					m = mt
				}
			}
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
		// Cannot compute collection type name before element media type
		// DSL has executed since the DSL may modify element type name
		// via the TypeName function.
		mt.TypeName = m.TypeName + "Collection"
		mt.AttributeExpr = &design.AttributeExpr{Type: ArrayOf(m)}
		if len(adsl) > 0 {
			eval.Execute(adsl[0], mt)
		}
		if mt.Views == nil {
			// If the adsl didn't create any views (or there is no
			// adsl at all) then inherit the views from the
			// collection element.
			mt.Views = make([]*design.ViewExpr, len(m.Views))
			for i, v := range m.Views {
				v := v
				mt.Views[i] = v
			}
		}
	})
	// Do not execute the adsl right away, will be done last to make sure
	// the element adsl has run first.
	design.Root.GeneratedTypes = append(design.Root.GeneratedTypes, mt)
	return mt
}

// Reference sets a type or media type reference. The value itself can be a type
// or a media type.  The reference type attributes define the default properties
// for attributes with the same name in the type using the reference.
//
// Reference may be used in Type or MediaType.
// Reference accepts a single argument: the type or media type containing the
// attributes that define the default properties of the attributes of the type
// or media type that uses Reference.
//
// Example:
//
//	var Bottle = Type("bottle", func() {
//		Attribute("name", String, func() {
//			MinLength(3)
//		})
//		Attribute("vintage", Int32, func() {
//			Minimum(1970)
//		})
//		Attribute("somethingelse", String)
//	})
//
//	var BottleMedia = MediaType("vnd.goa.bottle", func() {
//		Reference(Bottle)
//		Attributes(func() {
//			Attribute("id", UInt64, "ID is the bottle identifier")
//
//                      // The type and validation of "name" and "vintage" are
//                      // inherited from the Bottle type "name" and "vintage"
//                      // attributes.
//			Attribute("name")
//			Attribute("vintage")
//		})
//	})
//
func Reference(t design.DataType) {
	switch def := eval.Current().(type) {
	case *design.MediaTypeExpr:
		def.Reference = t
	case *design.AttributeExpr:
		def.Reference = t
	default:
		eval.IncompatibleDSL()
	}
}

// Attributes implements the media type Attributes DSL. See MediaType.
func Attributes(fn func()) {
	mt, ok := eval.Current().(*design.MediaTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	eval.Execute(fn, mt)
}

// buildView builds a view expression given an attribute and a corresponding
// media type.
func buildView(name string, mt *design.MediaTypeExpr, at *design.AttributeExpr) (*design.ViewExpr, error) {
	if at.Type == nil {
		return nil, fmt.Errorf("invalid view DSL")
	}
	o := design.AsObject(at.Type)
	if o == nil {
		return nil, fmt.Errorf("invalid view DSL")
	}
	mto := design.AsObject(mt.Type)
	if mto == nil {
		mto = design.AsObject(mt.Type.(*design.Array).ElemType.Type)
	}
	for n, cat := range o {
		if existing, ok := mto[n]; ok {
			dup := design.DupAtt(existing)
			if dup.Metadata == nil {
				dup.Metadata = make(map[string][]string)
			}
			dup.Metadata["view"] = cat.Metadata["view"]
			o[n] = dup
		} else if n != "links" {
			return nil, fmt.Errorf("unknown attribute %#v", n)
		}
	}
	return &design.ViewExpr{
		AttributeExpr: at,
		Name:          name,
		Parent:        mt,
	}, nil
}
