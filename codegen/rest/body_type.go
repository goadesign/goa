package rest

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// RequestBodyType returns the type of the request body given an action. If the
// design specifies a body explicitly using the Body DSL then it is returned.
// Otherwise one is computed by removing the attributes of the endpoint payload
// used to define headers and parameters.
func RequestBodyType(action *rest.ActionExpr) *design.AttributeExpr {
	if action.Body != nil {
		return action.Body
	}
	if design.IsObject(action.EndpointExpr.Payload.Type) {
		body := rest.NewMappedAttributeExpr(action.EndpointExpr.Payload)
		removeAttributes(body, action.MappedHeaders())
		removeAttributes(body, action.AllParams())
		return body.Attribute()
	}
	return action.EndpointExpr.Payload
}

// ResponseBodyType returns the type of the response body for the given response
// and action. If the design specifies a body explicitly using the Body DSL then
// it is returned. Otherwise one is computed by removing the attributes of the
// endpoint result used to define the response headers.
// If the response defines a view then the resulting attribute is the result of
// projecting the media type with that view.
func ResponseBodyType(action *rest.ActionExpr, response *rest.HTTPResponseExpr) *design.AttributeExpr {
	if response.Body != nil {
		return response.Body
	}
	res := action.EndpointExpr.Result
	if design.IsObject(res.Type) {
		mt, ok := res.Type.(*design.MediaTypeExpr)
		if ok {
			if v := res.Metadata["view"]; len(v) > 0 {
				p, err := new(design.Projector).Project(mt, v[0])
				if err == nil {
					res = p.MediaType.AttributeExpr
				}
			}
		}
		body := rest.NewMappedAttributeExpr(res)
		removeAttributes(body, response.MappedHeaders())
		return body.Attribute()
	}
	return action.EndpointExpr.Result
}

func removeAttributes(attr, sub *rest.MappedAttributeExpr) {
	WalkMappedAttr(sub, func(name, _ string, _ bool, _ *design.AttributeExpr) error {
		attr.Delete(name)
		return nil
	})
}
