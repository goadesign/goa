package dsl

import (
	apidesign "goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

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
