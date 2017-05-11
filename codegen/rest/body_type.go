package restgen

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// RequestBodyType returns the type of the request body given an action, it also
// returns whether the type is a public type (service package type) or private
// (only used for transport). If the design specifies a body explicitly using
// the Body DSL then it is returned. Otherwise one is computed by removing the
// attributes of the endpoint payload used to define headers and parameters.
func RequestBodyType(action *rest.ActionExpr, name string) (design.DataType, bool) {
	if action.Body != nil {
		if action.Body.Type != design.Empty {
			return action.Body.Type, false
		}
	}

	dt := action.EndpointExpr.Payload.Type
	if !design.IsObject(dt) {
		return dt, true
	}

	// 1. Return user type if no modification needed
	if _, ok := dt.(design.UserType); ok {
		if headers := action.Headers(); len(design.AsObject(headers.Type)) == 0 {
			if params := action.AllParams(); len(design.AsObject(params.Type)) == 0 {
				return dt, true
			}
		}
	}

	// 2. Remove header and param attributes
	body := rest.NewMappedAttributeExpr(action.EndpointExpr.Payload)
	removeAttributes(body, action.MappedHeaders())
	removeAttributes(body, action.AllParams())

	// 3. Build computed user type
	return &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}, false
}

// ResponseBodyType returns the type of the response body for the given response
// and result. If result's Body is not nil then its type is returned. Otherwise
// one is computed by removing the attributes of the endpoint result used to
// define the response headers from the attributes of the response. If the
// response defines a view then the resulting attribute is the result of
// projecting the media type with that view.
func ResponseBodyType(result *design.AttributeExpr, response *rest.HTTPResponseExpr, name string) design.DataType {
	if response.Body != nil {
		if response.Body.Type != design.Empty {
			return response.Body.Type
		}
		return nil
	}
	if result == nil || result.Type == design.Empty {
		return nil
	}

	dt := result.Type
	if !design.IsObject(dt) {
		return dt
	}

	// 1. Project if response type is media type and attribute has a
	// view.
	mt, ismt := dt.(*design.MediaTypeExpr)
	if ismt {
		if v := result.Metadata["view"]; len(v) > 0 {
			p, err := new(design.Projector).Project(mt, v[0])
			if err == nil {
				dt = p.MediaType
				result = design.DupAtt(result)
				result.Type = dt
			}
		}
	}

	// 2. Return user type if no modification needed
	if _, ok := dt.(design.UserType); ok {
		if headers := response.Headers(); len(design.AsObject(headers.Type)) == 0 {
			return dt
		}
	}

	// 3. Remove header attributes
	body := rest.NewMappedAttributeExpr(result)
	headers := response.MappedHeaders()
	removeAttributes(body, headers)

	// 4. Build computed user type
	userType := &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	if ismt {
		views := make([]*design.ViewExpr, len(mt.Views))
		for i, v := range mt.Views {
			mv := rest.NewMappedAttributeExpr(v.AttributeExpr)
			removeAttributes(mv, headers)
			nv := &design.ViewExpr{
				AttributeExpr: mv.Attribute(),
				Name:          v.Name,
			}
			views[i] = nv
		}
		nmt := &design.MediaTypeExpr{
			UserTypeExpr: userType,
			Identifier:   mt.Identifier,
			ContentType:  mt.ContentType,
			Views:        views,
		}
		for _, v := range views {
			v.Parent = nmt
		}
		return nmt
	}
	return userType
}

func removeAttributes(attr, sub *rest.MappedAttributeExpr) {
	WalkMappedAttr(sub, func(name, _ string, _ bool, _ *design.AttributeExpr) error {
		attr.Delete(name)
		return nil
	})
}
