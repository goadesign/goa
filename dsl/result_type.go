package dsl

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa/design"
	"goa.design/goa/eval"
)

// Counter used to create unique result type names for identifier-less result
// types.
var resultTypeCount int

// ResultType defines a result type used to describe a method response.
//
// Result types have a unique identifier as described in RFC 6838. The
// identifier defines the default value for the Content-Type header of HTTP
// responses.
//
// The result type expression includes a listing of all the response attributes.
// Views specify which of the attributes are actually rendered so that the same
// result type expression may represent multiple rendering of a given response.
//
// All result types have a view named "default". This view is used to render the
// result type in responses when no other view is specified. If the default view
// is not explicitly described in the DSL then one is created that lists all the
// result type attributes.
//
// ResultType is a top level DSL.
//
// ResultType accepts two arguments: the result type identifier and the defining
// DSL.
//
// Example:
//
//    var BottleMT = ResultType("application/vnd.goa.example.bottle", func() {
//        Description("A bottle of wine")
//        TypeName("BottleResult")         // Override generated type name
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
func ResultType(identifier string, fn func()) *design.ResultTypeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	// Validate Result Type
	identifier, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		eval.ReportError("invalid result type identifier %#v: %s",
			identifier, err)
		// We don't return so that other errors may be captured in this
		// one run.
		identifier = "text/plain"
	}
	canonicalID := design.CanonicalIdentifier(identifier)
	// Validate that result type identifier doesn't clash
	for _, rt := range design.Root.ResultTypes {
		if re := rt.(*design.ResultTypeExpr); re.Identifier == canonicalID {
			eval.ReportError(
				"result type %#v with canonical identifier %#v is defined twice",
				identifier, canonicalID)
			return nil
		}
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
		resultTypeCount++
		typeName = fmt.Sprintf("ResultType%d", resultTypeCount)
	}
	// Now save the type in the API result types map
	mt := design.NewResultTypeExpr(typeName, identifier, fn)
	design.Root.ResultTypes = append(design.Root.ResultTypes, mt)

	return mt
}

// TypeName makes it possible to set the Go struct name for a type or result
// type in the generated code. By default goa uses the name (type) or identifier
// (result type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
func TypeName(name string) {
	switch expr := eval.Current().(type) {
	case design.UserType:
		expr.Rename(name)
	case *design.AttributeExpr:
		if expr.Metadata == nil {
			expr.Metadata = make(design.MetadataExpr)
		}
		expr.Metadata["struct:type:name"] = []string{name}
	default:
		eval.IncompatibleDSL()
	}
}

// ContentType sets the value of the Content-Type response header. By default
// the ID of the result type is used.
//
// ContentType may appear in a ResultType expression.
// ContentType accepts one argument: the mime type as defined by RFC 6838.
//
//    var _ = ResultType("application/vnd.myapp.mytype") {
//        ContentType("application/json")
//    }
//
func ContentType(typ string) {
	if mt, ok := eval.Current().(*design.ResultTypeExpr); ok {
		mt.ContentType = typ
	} else {
		eval.IncompatibleDSL()
	}
}

// View adds a new view to a result type. A view has a name and lists attributes
// that are rendered when the view is used to produce a response. The attribute
// names must appear in the result type expression. If an attribute is itself a
// result type then the view may specify which view to use when rendering the
// attribute using the View function in the View DSL. If not specified then the
// view named "default" is used.
//
// View must appear in a ResultType expression.
//
// View accepts two arguments: the view name and its defining DSL.
//
// Examples:
//
//	View("default", func() {
//		// "id" and "name" must be result type attributes
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
	case *design.ResultTypeExpr:
		mt := expr
		if mt.View(name) != nil {
			eval.ReportError("multiple expressions for view %#v in result type %#v", name, mt.TypeName)
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
				if pa, ok2 := elem.Type.(*design.ResultTypeExpr); ok2 {
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

// CollectionOf creates a collection result type from its element result type. A
// collection result type represents the content of responses that return a
// collection of values such as listings. The expression accepts an optional DSL
// as second argument that allows specifying which view(s) of the original result
// type apply.
//
// The resulting result type identifier is built from the element result type by
// appending the result type parameter "type" with value "collection".
//
// CollectionOf must appear wherever ResultType can.
//
// CollectionOf takes the element result type as first argument and an optional
// DSL as second argument.
//
// Example:
//
//     var DivisionResult = ResultType("application/vnd.goa.divresult", func() {
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
func CollectionOf(v interface{}, adsl ...func()) *design.ResultTypeExpr {
	var m *design.ResultTypeExpr
	var ok bool
	m, ok = v.(*design.ResultTypeExpr)
	if !ok {
		if id, ok := v.(string); ok {
			if dt := design.Root.UserType(design.CanonicalIdentifier(id)); dt != nil {
				if mt, ok := dt.(*design.ResultTypeExpr); ok {
					m = mt
				}
			}
		}
	}
	if m == nil {
		eval.ReportError("invalid CollectionOf argument: not a result type and not a known result type identifier")
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewResultTypeExpr("InvalidCollection", "text/plain", nil)
	}
	id := m.Identifier
	rtype, params, err := mime.ParseMediaType(id)
	if err != nil {
		eval.ReportError("invalid result type identifier %#v: %s", id, err)
		// don't return nil to avoid panics, the error will get reported at the end
		return design.NewResultTypeExpr("InvalidCollection", "text/plain", nil)
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
	id = mime.FormatMediaType(rtype, params)
	canonical := design.CanonicalIdentifier(id)
	if mt := design.Root.GeneratedResultType(canonical); mt != nil {
		// Already have a type for this collection, reuse it.
		return mt
	}
	mt := design.NewResultTypeExpr("", id, func() {
		rt, ok := eval.Current().(*design.ResultTypeExpr)
		if !ok {
			eval.IncompatibleDSL()
			return
		}
		// Cannot compute collection type name before element result type
		// DSL has executed since the DSL may modify element type name
		// via the TypeName function.
		rt.TypeName = m.TypeName + "Collection"
		rt.AttributeExpr = &design.AttributeExpr{Type: ArrayOf(m)}
		if len(adsl) > 0 {
			eval.Execute(adsl[0], rt)
		}
		if rt.Views == nil {
			// If the DSL didn't create any view (or there is no DSL
			// at all) then inherit the views from the collection
			// element.
			rt.Views = make([]*design.ViewExpr, len(m.Views))
			for i, v := range m.Views {
				v := v
				rt.Views[i] = v
			}
		}
	})
	// do not execute the DSL right away, will be done last to make sure
	// the element DSL has run first.
	*design.Root.GeneratedTypes = append(*design.Root.GeneratedTypes, mt)
	return mt
}

// Reference sets a type or result type reference. The value itself can be a
// type or a result type. The reference type attributes define the default
// properties for attributes with the same name in the type using the reference.
//
// Reference may be used in Type or ResultType, it may appear multiple times in
// which case attributes are looked up in each reference in order of appearance
// in the DSL.
//
// Reference accepts a single argument: the type or result type containing the
// attributes that define the default properties of the attributes of the type
// or result type that uses Reference.
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
//	var BottleResult = ResultType("vnd.goa.bottle", func() {
//		Reference(Bottle)
//		Attributes(func() {
//			Attribute("id", UInt64, "ID is the bottle identifier")
//
//			// The type and validation of "name" and "vintage" are
//			// inherited from the Bottle type "name" and "vintage"
//			// attributes.
//			Attribute("name")
//			Attribute("vintage")
//		})
//	})
//
func Reference(t design.DataType) {
	if !design.IsObject(t) {
		eval.ReportError("argument of Reference must be an object, got %s", t.Name())
		return
	}
	switch def := eval.Current().(type) {
	case *design.ResultTypeExpr:
		def.References = append(def.References, t)
	case *design.AttributeExpr:
		def.References = append(def.References, t)
	default:
		eval.IncompatibleDSL()
	}
}

// Extend adds the parameter type attributes to the type using Extend. The
// parameter type must be an object.
//
// Extend may be used in Type or ResultType. Extend accepts a single argument:
// the type or result type containing the attributes to be copied.
//
// Example:
//
//    var CreateBottlePayload = Type("CreateBottlePayload", func() {
//       Attribute("name", String, func() {
//          MinLength(3)
//       })
//       Attribute("vintage", Int32, func() {
//          Minimum(1970)
//       })
//    })
//
//    var UpdateBottlePayload = Type("UpatePayload", func() {
//        Atribute("id", String, "ID of bottle to update")
//        Extend(CreateBottlePayload) // Adds attributes "name" and "vintage"
//    })
//
func Extend(t design.DataType) {
	if !design.IsObject(t) {
		eval.ReportError("argument of Extend must be an object, got %s", t.Name())
		return
	}
	switch def := eval.Current().(type) {
	case *design.ResultTypeExpr:
		def.Bases = append(def.Bases, t)
	case *design.AttributeExpr:
		def.Bases = append(def.Bases, t)
	default:
		eval.IncompatibleDSL()
	}
}

// Attributes implements the result type Attributes DSL. See ResultType.
func Attributes(fn func()) {
	mt, ok := eval.Current().(*design.ResultTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	eval.Execute(fn, mt)
}

// buildView builds a view expression given an attribute and a corresponding
// result type.
func buildView(name string, mt *design.ResultTypeExpr, at *design.AttributeExpr) (*design.ViewExpr, error) {
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
	for _, nat := range *o {
		n := nat.Name
		cat := nat.Attribute
		if existing := mt.Find(n); existing != nil {
			dup := design.DupAtt(existing)
			if dup.Metadata == nil {
				dup.Metadata = make(map[string][]string)
			}
			if len(cat.Metadata["view"]) > 0 {
				dup.Metadata["view"] = cat.Metadata["view"]
			}
			o.Set(n, dup)
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
