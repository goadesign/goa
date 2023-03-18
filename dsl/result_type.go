package dsl

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Counter used to create unique result type names for identifier-less result
// types.
var resultTypeCount int

// ResultType defines a result type used to describe a method response.
//
// Result types have a unique identifier as described in RFC 6838. Result types
// may also define a type name used to override the default Go type name
// generated from the identifier.
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
// Note: it is not required to use a ResultType to describe the type of a method
// result, Type can also be used and is preferred if there is no need to define
// multiple views.
//
// ResultType is a top level DSL.
//
// ResultType accepts two or three arguments: the result type identifier, an
// optional type name and the defining DSL.
//
// Example:
//
//	var BottleMT = ResultType("application/vnd.goa.example.bottle", "BottleResult", func() {
//	    Description("A bottle of wine")
//
//	    Attributes(func() {
//	        Attribute("id", Int, "ID of bottle")
//	        Attribute("href", String, "API href of bottle")
//	        Attribute("account", Account, "Owner account")
//	        Attribute("origin", Origin, "Details on wine origin")
//	        Required("id", "href")
//	    })
//
//	    View("default", func() {        // Explicitly define default view
//	        Attribute("id")
//	        Attribute("href")
//	    })
//
//	    View("extended", func() {       // Define "extended" view
//	        Attribute("id")
//	        Attribute("href")
//	        Attribute("account")
//	        Attribute("origin")
//	    })
//	 })
func ResultType(identifier string, args ...any) *expr.ResultTypeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	var (
		typeName string
		fn       func()
	)
	{
		var err error
		identifier, typeName, err = mediaTypeToResultType(identifier)
		if err != nil {
			eval.ReportError("invalid result type identifier %#v: %s", identifier, err)
			// We don't return so that other errors may be captured
		}
		if len(args) > 0 {
			switch a := args[0].(type) {
			case func():
				fn = a
			case string:
				typeName = a
			default:
				eval.InvalidArgError("function or string", args[0])
			}
			if len(args) > 1 {
				if fn != nil {
					eval.ReportError("DSL function must be last argument")
				}
				if f, ok := args[1].(func()); ok {
					fn = f
				} else {
					eval.InvalidArgError("function", args[1])
				}
				if len(args) > 2 {
					eval.ReportError("too many arguments")
				}
			}
		}
	}
	canonicalID := expr.CanonicalIdentifier(identifier)
	// Validate that result type identifier doesn't clash
	for _, rt := range expr.Root.ResultTypes {
		if re := rt.(*expr.ResultTypeExpr); re.Identifier == canonicalID {
			eval.ReportError(
				"result type %#v with canonical identifier %#v is defined twice",
				identifier, canonicalID)
			return nil
		}
	}
	// Now save the type in the API result types map
	mt := expr.NewResultTypeExpr(typeName, identifier, fn)
	expr.Root.ResultTypes = append(expr.Root.ResultTypes, mt)

	return mt
}

// TypeName makes it possible to set the Go struct name for a type or result
// type in the generated code. By default goa uses the name (type) or identifier
// (result type) given in the DSL and computes a valid Go identifier from it.
// This function makes it possible to override that and provide a custom name.
// name must be a valid Go identifier.
//
// TypeName must appear in a Type or ResultType expression.
func TypeName(name string) {
	switch e := eval.Current().(type) {
	case expr.UserType:
		e.Rename(name)
	case *expr.AttributeExpr:
		e.AddMeta("struct:type:name", name)
	default:
		eval.IncompatibleDSL()
	}
}

// View has two usages:
//
// - when used inside a ResultType DSL function it defines a view for the result
// type. A view lists a subset of the result type attributes that are used when
// marshalling responses.
//
// - when used inside a Result DSL function it defines the view used to marshal
// the result type returned by the method.
//
// Note that the view used to render a response can also be set dynamically by
// the method code in which case the result function should not specify a view
// in the design.  The attribute names listed in a view must be identical to
// existing attributes in the result type on which the view is defined. If an
// attribute is itself a result type then the view may specify which view to use
// when marshaling the attribute using the View function recursively, see
// example below. All result types must have a view called "default" which is
// the view used to marshal results when no specific view is specified.
//
// View must appear in a ResultType or a Result expression.
//
// View accepts two arguments for the first usage: the view name and its
// defining DSL.  View accepts a single argument for the second usage: the view
// name used to render the result.
//
// Examples:
//
//	// MyResultType defines 2 views.
//	var MyResultType = ResultType("application/vnd.goa.my", func() {
//	    Attributes(func() {
//	        Attribute("id", String)
//	        Attribute("name", String)
//	        Attribute("origin", OriginResult)
//	    })
//
//	    View("default", func() {
//	        // "id" and "name" must be result type attributes
//	        Attribute("id")
//	        Attribute("name")
//	    })
//
//	    View("extended", func() {
//	        Attribute("id")
//	        Attribute("name")
//	        Attribute("origin", func() {
//	            // Use view "extended" to render attribute "origin"
//	            View("extended")
//	        })
//	    })
//	})
//
//	// MyMethod uses the extended view of MyResultType to marshal the
//	// response.
//	var _ = Service("MyService", func() {
//	    Method("MyMethod", func() {
//	        Result(MyResultType, func() { View("extended") })
//	        GRPC(func() {})
//	    })
//	})
func View(name string, adsl ...func()) {
	switch e := eval.Current().(type) {
	case *expr.ResultTypeExpr:
		if e.View(name) != nil {
			eval.ReportError("multiple expressions for view %#v in result type %#v", name, e.TypeName)
			return
		}
		at := &expr.AttributeExpr{}
		ok := false
		var a *expr.Array
		if len(adsl) > 0 {
			ok = eval.Execute(adsl[0], at)
		} else if a, ok = e.Type.(*expr.Array); ok {
			// inherit view from collection element if present
			if elem := a.ElemType; elem != nil {
				if pa, ok2 := elem.Type.(*expr.ResultTypeExpr); ok2 {
					print(pa)
					print(pa.Views)
					if v := pa.View(name); v != nil {
						at = v.AttributeExpr
					} else {
						eval.ReportError("unknown view %#v", name)
						return
					}
				}
			}
		}
		if ok {
			view, err := buildView(name, e, at)
			if err != nil {
				eval.ReportError(err.Error())
				return
			}
			e.Views = append(e.Views, view)
		}

	case *expr.AttributeExpr:
		e.AddMeta("view", name)

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
//	var DivisionResult = ResultType("application/vnd.goa.divresult", func() {
//	    Attributes(func() {
//	        Attribute("value", Float64)
//	        Attribute("remainder", Int)
//	    })
//	    View("default", func() {
//	        Attribute("value")
//	        Attribute("remainder")
//	    })
//	    View("tiny", func() {
//	        Attribute("value")
//	    })
//	})
//
//	var MultiResults = CollectionOf(DivisionResult)
//
//	var TinyMultiResults = CollectionOf(DivisionResult, func() {
//	    View("tiny")  // use "tiny" view to render the collection elements
//	})
//
//	var MultiResultsExample = CollectionOf(DivisionResult, func() {
//	    Attributes(func() {
//	        Example("DivisionResult Collection Examples", func() {
//	            Value([]Val{
//	                {
//	                    "value":     4.167,
//	                    "remainder": 0,
//	                },
//	                {
//	                    "value":     3.0,
//	                    "remainder": 0,
//	                },
//	            })
//	        })
//	    })
//	})
func CollectionOf(v any, adsl ...func()) *expr.ResultTypeExpr {
	var m *expr.ResultTypeExpr
	var ok bool
	m, ok = v.(*expr.ResultTypeExpr)
	if !ok {
		if id, ok := v.(string); ok {
			// Check if a result type exists with the given type name
			if dt := expr.Root.UserType(id); dt != nil {
				if mt, ok := dt.(*expr.ResultTypeExpr); ok {
					m = mt
				}
			} else {
				// Check if a result type exists with the given identifier
				id, typeName, err := mediaTypeToResultType(id)
				if dt := expr.Root.UserType(typeName); dt != nil {
					if mt, ok := dt.(*expr.ResultTypeExpr); ok {
						m = mt
					}
				}
				if err != nil {
					eval.ReportError("invalid result type identifier %#v in CollectionOf: %s", id, err)
				}
			}
		}
	}
	if m == nil {
		eval.ReportError("invalid CollectionOf argument: not a result type and not a known result type identifier")
		// don't return nil to avoid panics, the error will get reported at the end
		return expr.NewResultTypeExpr("InvalidCollection", "text/plain", nil)
	}
	id := m.Identifier
	rtype, params, err := mime.ParseMediaType(id)
	if err != nil {
		eval.ReportError("invalid result type identifier %#v: %s", id, err)
		// don't return nil to avoid panics, the error will get reported at the end
		return expr.NewResultTypeExpr("InvalidCollection", "text/plain", nil)
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
	canonical := expr.CanonicalIdentifier(id)
	if mt := expr.Root.GeneratedResultType(canonical); mt != nil {
		// Already have a type for this collection, reuse it.
		return mt
	}
	mt := expr.NewResultTypeExpr("", id, func() {
		rt, ok := eval.Current().(*expr.ResultTypeExpr)
		if !ok {
			eval.IncompatibleDSL()
			return
		}
		// Cannot compute collection type name before element result type
		// DSL has executed since the DSL may modify element type name
		// via the TypeName function.
		rt.TypeName = m.TypeName + "Collection"
		rt.AttributeExpr = &expr.AttributeExpr{Type: ArrayOf(m)}
		if len(adsl) > 0 {
			eval.Execute(adsl[0], rt)
		}
		if rt.Views == nil {
			// If the DSL didn't create any view (or there is no DSL
			// at all) then inherit the views from the collection
			// element.
			rt.Views = make([]*expr.ViewExpr, len(m.Views))
			for i, v := range m.Views {
				v := v
				rt.Views[i] = v
			}
		}
	})
	// do not execute the DSL right away, will be done last to make sure
	// the element DSL has run first.
	*expr.Root.GeneratedTypes = append(*expr.Root.GeneratedTypes, mt)
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
func Reference(t expr.DataType) {
	if !expr.IsObject(t) {
		eval.ReportError("argument of Reference must be an object, got %s", t.Name())
		return
	}
	switch def := eval.Current().(type) {
	case *expr.ResultTypeExpr:
		def.References = append(def.References, t)
	case *expr.AttributeExpr:
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
//	var CreateBottlePayload = Type("CreateBottlePayload", func() {
//	   Attribute("name", String, func() {
//	      MinLength(3)
//	   })
//	   Attribute("vintage", Int32, func() {
//	      Minimum(1970)
//	   })
//	})
//
//	var UpdateBottlePayload = Type("UpatePayload", func() {
//	    Attribute("id", String, "ID of bottle to update")
//	    Extend(CreateBottlePayload) // Adds attributes "name" and "vintage"
//	})
func Extend(t expr.DataType) {
	if !expr.IsObject(t) {
		eval.ReportError("argument of Extend must be an object, got %s", t.Name())
		return
	}
	switch def := eval.Current().(type) {
	case *expr.ResultTypeExpr:
		def.Bases = append(def.Bases, t)
	case *expr.AttributeExpr:
		def.Bases = append(def.Bases, t)
	default:
		eval.IncompatibleDSL()
	}
}

// Attributes implements the result type Attributes DSL. See ResultType.
func Attributes(fn func()) {
	mt, ok := eval.Current().(*expr.ResultTypeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	eval.Execute(fn, mt.AttributeExpr)
}

// mediaTypeToResultType returns the formatted identifier and the result type
// name from the given identifier string. If the given identifier is invalid it
// returns text/plain as the identifier and an error.
func mediaTypeToResultType(identifier string) (string, string, error) {
	identifier, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		identifier = "text/plain"
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
		elems[i] = expr.Title(e)
	}
	typeName := strings.Join(elems, "")
	if typeName == "" {
		resultTypeCount++
		typeName = fmt.Sprintf("ResultType%d", resultTypeCount)
	}
	return identifier, typeName, err
}

// buildView builds a view expression given an attribute and a corresponding
// result type. The attribute must be an object listing the child attributes
// that make up the view.
func buildView(name string, mt *expr.ResultTypeExpr, at *expr.AttributeExpr) (*expr.ViewExpr, error) {
	if at.Type == nil {
		return nil, fmt.Errorf("invalid view DSL")
	}
	o := expr.AsObject(at.Type)
	if o == nil {
		return nil, fmt.Errorf("invalid view DSL")
	}
	for _, nat := range *o {
		n := nat.Name
		cat := nat.Attribute
		if existing := mt.Find(n); existing != nil {
			dup := expr.DupAtt(existing)
			if _, ok := cat.Meta["view"]; ok {
				dup.AddMeta("view", cat.Meta["view"]...)
			}
			o.Set(n, dup)
		} else if n != "links" {
			return nil, fmt.Errorf("unknown attribute %#v", n)
		}
	}
	return &expr.ViewExpr{
		AttributeExpr: at,
		Name:          name,
		Parent:        mt,
	}, nil
}
