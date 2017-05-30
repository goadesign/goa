package restgen

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// RequestBodyType returns the type of the request body given an action. If the
// DSL defines a body explicitly via the Body function then the corresponding
// type is used instead of the payload type. Otherwise the type is computed by
// removing the attributes of the endpoint payload used to define headers and
// parameters.
func RequestBodyType(r *rest.ResourceExpr, a *rest.ActionExpr, suffix string) design.DataType {
	if a.Body != nil {
		return a.Body.Type
	}

	var (
		dt      = a.EndpointExpr.Payload.Type
		headers = a.MappedHeaders()
		params  = a.AllParams()
	)

	bodyOnly := len(design.AsObject(headers.Type)) == 0 &&
		len(design.AsObject(params.Type)) == 0

	// 1. If Payload is not an object then check whether there are params or
	// headers defined and if so return empty type (payload encoded in
	// request params or headers) otherwise return payload type (payload
	// encoded in request body).
	if !design.IsObject(dt) {
		if bodyOnly {
			return dt
		}
		return design.Empty
	}

	// 2. Return user type if no modification needed
	if _, ok := dt.(design.UserType); ok {
		if len(design.AsObject(headers.Type)) == 0 && len(design.AsObject(params.Type)) == 0 {
			return dt
		}
	}

	// 3. Remove header and param attributes
	body := rest.NewMappedAttributeExpr(a.EndpointExpr.Payload)
	removeAttributes(body, headers)
	removeAttributes(body, params)

	// 4. Return empty type if no attribute left
	if len(design.AsObject(body.Type)) == 0 {
		return design.Empty
	}

	// 5. Build computed user type
	name := codegen.Goify(a.Name(), true) + suffix
	return &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
}

// ResponseBodyType returns the type of the response body given a response and
// the corresponding service attribute (either a result or an error attribute).
// and result attribute. If the DSL defines a body explicitly via the Body
// function then the corresponding type is used instead of the attribute type.
// Otherwise the type is computed by removing the attributes of the endpoint
// payload used to define headers and parameters. Also if the response defines a
// view then the response media type is projected first. suffix is appended to
// the created type name if any.
func ResponseBodyType(r *rest.ResourceExpr, resp *rest.HTTPResponseExpr, result *design.AttributeExpr, suffix string) design.DataType {
	if result == nil || result.Type == design.Empty {
		return design.Empty
	}
	if resp.Body != nil {
		return resp.Body.Type
	}

	var (
		dt      = result.Type
		headers = resp.MappedHeaders()
	)

	// 1. If Result is not an object then check whether there are headers
	// defined and if so return empty type (result encoded in response
	// headers) otherwise return result type (result encoded in response
	// body).
	if !design.IsObject(dt) {
		if len(design.AsObject(resp.Headers().Type)) == 0 {
			return dt
		}
		return design.Empty
	}

	// 2. Project if response type is media type and attribute has a
	// view.
	mt, ismt := dt.(*design.MediaTypeExpr)
	if ismt {
		if v := result.Metadata["view"]; len(v) > 0 {
			p, err := new(design.Projector).Project(mt, v[0])
			if err != nil {
				panic(err) // bug
			}
			dt = p.MediaType
			result = design.DupAtt(result)
			result.Type = dt
		}
	}

	// 3. Return user type if no modification needed
	if _, ok := dt.(design.UserType); ok {
		if headers := resp.Headers(); len(design.AsObject(headers.Type)) == 0 {
			return dt
		}
	}

	// 4. Remove header attributes
	body := rest.NewMappedAttributeExpr(result)
	removeAttributes(body, headers)

	// 5. Return empty type if no attribute left
	if len(design.AsObject(body.Type)) == 0 {
		return design.Empty
	}

	// 6. Build computed user type
	action := resp.Parent.(*rest.ActionExpr)
	name := codegen.Goify(action.Name(), true) + suffix + "ResponseBody"
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
