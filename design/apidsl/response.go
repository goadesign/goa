package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Response implements the response definition DSL. Response takes the name of the response as
// first parameter. goa defines all the standard HTTP status name as global variables so they can be
// readily used as response names. The response body data type can be specified as second argument.
// If a type is specified it overrides any type defined by in the response media type. Response also
// accepts optional arguments that correspond to the arguments defined by the corresponding response
// template (the response template with the same name) if there is one, see ResponseTemplate.
//
// A response may also optionally use an anonymous function as last argument to specify the response
// status code, media type and headers overriding what the default response or response template
// specifies:
//
//        Response(OK, "text/plain")              // OK response template accepts one argument:
//                                                // the media type identifier
//
//        Response(OK, BottleMedia)               // or a media type defined in the design
//
//        Response(OK, "application/vnd.bottle")  // optionally referred to by identifier
//
//        Response(OK, func() {
//                Media("application/vnd.bottle") // Alternatively media type is set with Media
//        })
//
//        Response(OK, BottleMedia, func() {
//                Headers(func() {                // Headers list the response HTTP headers
//                        Header("X-Request-Id")  // Header syntax is identical to Attribute's
//                })
//        })
//
//        Response(OK, BottleMedia, func() {
//                Status(201)                     // Set response status (overrides template's)
//        })
//
//        Response("MyResponse", func() {         // Define custom response (using no template)
//                Description("This is my response")
//                Media(BottleMedia)
//                Headers(func() {
//                        Header("X-Request-Id", func() {
//                                Pattern("[a-f0-9]+")
//                        })
//                })
//                Status(200)
//        })
//
// goa defines a default response template for all the HTTP status code. The default template simply sets
// the status code. So if an action can return NotFound for example all it has to do is specify
// Response(NotFound) - there is no need to specify the status code as the default response already
// does it, in other words:
//
//	Response(NotFound)
//
// is equivalent to:
//
//	Response(NotFound, func() {
//		Status(404)
//	})
//
// goa also defines a default response template for the OK response which takes a single argument:
// the identifier of the media type used to render the response. The API DSL can define additional
// response templates or override the default OK response template using ResponseTemplate.
//
// The media type identifier specified in a response definition via the Media function can be
// "generic" such as "text/plain" or "application/json" or can correspond to the identifier of a
// media type defined in the API DSL. In this latter case goa uses the media type definition to
// generate helper response methods. These methods know how to render the views defined on the media
// type and run the validations defined in the media type during rendering.
func Response(name string, paramsAndDSL ...interface{}) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.ActionDefinition:
		if def.Responses == nil {
			def.Responses = make(map[string]*design.ResponseDefinition)
		}
		if _, ok := def.Responses[name]; ok {
			dslengine.ReportError("response %s is defined twice", name)
			return
		}
		if resp := executeResponseDSL(name, paramsAndDSL...); resp != nil {
			if resp.Status == 200 && resp.MediaType == "" {
				resp.MediaType = def.Parent.MediaType
				resp.ViewName = def.Parent.DefaultViewName
			}
			resp.Parent = def
			def.Responses[name] = resp
		}

	case *design.ResourceDefinition:
		if def.Responses == nil {
			def.Responses = make(map[string]*design.ResponseDefinition)
		}
		if _, ok := def.Responses[name]; ok {
			dslengine.ReportError("response %s is defined twice", name)
			return
		}
		if resp := executeResponseDSL(name, paramsAndDSL...); resp != nil {
			if resp.Status == 200 && resp.MediaType == "" {
				resp.MediaType = def.MediaType
				resp.ViewName = def.DefaultViewName
			}
			resp.Parent = def
			def.Responses[name] = resp
		}

	default:
		dslengine.IncompatibleDSL()
	}
}

// Status sets the Response status.
func Status(status int) {
	if r, ok := responseDefinition(); ok {
		r.Status = status
	}
}

func executeResponseDSL(name string, paramsAndDSL ...interface{}) *design.ResponseDefinition {
	var params []string
	var dsl func()
	var ok bool
	var dt design.DataType
	if len(paramsAndDSL) > 0 {
		d := paramsAndDSL[len(paramsAndDSL)-1]
		if dsl, ok = d.(func()); ok {
			paramsAndDSL = paramsAndDSL[:len(paramsAndDSL)-1]
		}
		if len(paramsAndDSL) > 0 {
			t := paramsAndDSL[0]
			if dt, ok = t.(design.DataType); ok {
				paramsAndDSL = paramsAndDSL[1:]
			}
		}
		params = make([]string, len(paramsAndDSL))
		for i, p := range paramsAndDSL {
			params[i], ok = p.(string)
			if !ok {
				dslengine.ReportError("invalid response template parameter %#v, must be a string", p)
				return nil
			}
		}
	}
	var resp *design.ResponseDefinition
	if len(params) > 0 {
		if tmpl, ok := design.Design.ResponseTemplates[name]; ok {
			resp = tmpl.Template(params...)
		} else if tmpl, ok := design.Design.DefaultResponseTemplates[name]; ok {
			resp = tmpl.Template(params...)
		} else {
			dslengine.ReportError("no response template named %#v", name)
			return nil
		}
	} else {
		if ar, ok := design.Design.Responses[name]; ok {
			resp = ar.Dup()
		} else if ar, ok := design.Design.DefaultResponses[name]; ok {
			resp = ar.Dup()
			resp.Standard = true
		} else {
			resp = &design.ResponseDefinition{Name: name}
		}
	}
	if dsl != nil {
		if !dslengine.Execute(dsl, resp) {
			return nil
		}
		resp.Standard = false
	}
	if dt != nil {
		if mt, ok := dt.(*design.MediaTypeDefinition); ok {
			resp.MediaType = mt.Identifier
		}
		resp.Type = dt
		resp.Standard = false
	}
	return resp
}
