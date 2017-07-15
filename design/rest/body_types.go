package rest

import (
	"net/http"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

// RequestBodyType returns the type of the request body given an endpoint. If
// the DSL defines a body explicitly via the Body function then the
// corresponding type is used instead of the payload type. Otherwise the type is
// computed by removing the attributes of the method payload used to define
// headers and parameters.
func RequestBodyType(a *HTTPEndpointExpr) design.DataType {
	if a.Body != nil {
		return a.Body.Type
	}

	var (
		dt      = a.MethodExpr.Payload.Type
		headers = a.MappedHeaders()
		params  = a.AllParams()
		suffix  = "ServerRequestBody"
		name    = codegen.Goify(a.Name(), true) + suffix
	)

	bodyOnly := len(*design.AsObject(headers.Type)) == 0 &&
		len(*design.AsObject(params.Type)) == 0

	// 1. If Payload is not an object then check whether there are params or
	// headers defined and if so return empty type (payload encoded in
	// request params or headers) otherwise return payload type (payload
	// encoded in request body).
	if !design.IsObject(dt) {
		if bodyOnly {
			return renameType(dt, name, "RequestBody")
		}
		return design.Empty
	}

	// 2. Remove header and param attributes
	body := design.NewMappedAttributeExpr(a.MethodExpr.Payload)
	removeAttributes(body, headers)
	removeAttributes(body, params)

	// 3. Return empty type if no attribute left
	if len(*design.AsObject(body.Type)) == 0 {
		return design.Empty
	}

	// 4. Build computed user type
	ut := &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	appendSuffix(ut.Attribute().Type, "RequestBody")

	return ut
}

// ResponseBodyType returns the type of the response body given a response and
// the corresponding service attribute (either a result or an error attribute).
// and result attribute. If the DSL defines a body explicitly via the Body
// function then the corresponding type is used instead of the attribute type.
// Otherwise the type is computed by removing the attributes of the method
// payload used to define headers and parameters. Also if the response defines a
// view then the response result type is projected first. suffix is appended to
// the created type name if any.
func ResponseBodyType(a *HTTPEndpointExpr, resp *HTTPResponseExpr) design.DataType {
	result := a.MethodExpr.Result
	if result == nil || result.Type == design.Empty {
		return design.Empty
	}
	if resp.Body != nil {
		return resp.Body.Type
	}

	var suffix string
	if len(a.Responses) > 1 {
		suffix = http.StatusText(resp.StatusCode)
	}

	var (
		dt      = result.Type
		headers = resp.MappedHeaders()
		name    = codegen.Goify(a.Name(), true) + suffix + "ResponseBody"
	)

	// 1. If Result is not an object then check whether there are headers
	// defined and if so return empty type (result encoded in response
	// headers) otherwise return renamed result type (result encoded in
	// response body).
	if !design.IsObject(dt) {
		if len(*design.AsObject(resp.Headers().Type)) == 0 {
			return renameType(dt, name, "ResponseBody")
		}
		return design.Empty
	}

	// 2. Project if response type is result type and attribute has a view.
	rt, isrt := dt.(*design.ResultTypeExpr)
	if isrt {
		if v := result.Metadata["view"]; len(v) > 0 {
			p, err := new(design.Projector).Project(rt, v[0])
			if err != nil {
				panic(err) // bug
			}
			dt = p.ResultType
			result = design.DupAtt(result)
			result.Type = dt
		}
	}

	// 3. Remove header attributes
	body := design.NewMappedAttributeExpr(result)
	removeAttributes(body, headers)

	// 4. Return empty type if no attribute left
	if len(*design.AsObject(body.Type)) == 0 {
		return design.Empty
	}

	// 5. Build computed user type
	userType := &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	if isrt {
		views := make([]*design.ViewExpr, len(rt.Views))
		for i, v := range rt.Views {
			mv := design.NewMappedAttributeExpr(v.AttributeExpr)
			removeAttributes(mv, headers)
			nv := &design.ViewExpr{
				AttributeExpr: mv.Attribute(),
				Name:          v.Name,
			}
			views[i] = nv
		}
		nmt := &design.ResultTypeExpr{
			UserTypeExpr: userType,
			Identifier:   rt.Identifier,
			ContentType:  rt.ContentType,
			Views:        views,
		}
		for _, v := range views {
			v.Parent = nmt
		}
		return nmt
	}
	appendSuffix(userType.Attribute().Type, "ResponseBody")

	return userType
}

// ErrorResponseBodyType returns the type of the response body given a error. If
// the DSL defines a body explicitly via the Body function then the
// corresponding type is used instead of the attribute type. Otherwise the type
// is computed by removing the attributes of the error used to define headers
// and parameters. Also if the error response defines a view then the result
// type is projected first. suffix is appended to the created type name if any.
func ErrorResponseBodyType(r *HTTPServiceExpr, a *HTTPEndpointExpr, v *HTTPErrorExpr) design.DataType {
	result := v.ErrorExpr.AttributeExpr
	if result == nil || result.Type == design.Empty {
		return design.Empty
	}
	resp := v.Response
	if resp.Body != nil {
		return resp.Body.Type
	}

	var (
		dt      = result.Type
		headers = resp.MappedHeaders()
		suffix  = codegen.Goify(v.ErrorExpr.Name, true) + "ResponseBody"
		name    = codegen.Goify(a.Name(), true) + suffix
	)

	// 1. If Result is not an object then check whether there are headers
	// defined and if so return empty type (result encoded in response
	// headers) otherwise return renamed result type (result encoded in
	// response body).
	if !design.IsObject(dt) {
		if len(*design.AsObject(resp.Headers().Type)) == 0 {
			return renameType(dt, name, suffix)
		}
		return design.Empty
	}

	// 2. Project if errorResponse type is result type and attribute has a view.
	rt, isrt := dt.(*design.ResultTypeExpr)
	if isrt {
		if v := result.Metadata["view"]; len(v) > 0 {
			p, err := new(design.Projector).Project(rt, v[0])
			if err != nil {
				panic(err) // bug
			}
			dt = p.ResultType
			result = design.DupAtt(result)
			result.Type = dt
		}
	}

	// 3. Remove header attributes
	body := design.NewMappedAttributeExpr(result)
	removeAttributes(body, headers)

	// 4. Return empty type if no attribute left
	if len(*design.AsObject(body.Type)) == 0 {
		return design.Empty
	}

	// 5. Build computed user type
	userType := &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	if isrt {
		views := make([]*design.ViewExpr, len(rt.Views))
		for i, v := range rt.Views {
			mv := design.NewMappedAttributeExpr(v.AttributeExpr)
			removeAttributes(mv, headers)
			nv := &design.ViewExpr{
				AttributeExpr: mv.Attribute(),
				Name:          v.Name,
			}
			views[i] = nv
		}
		nmt := &design.ResultTypeExpr{
			UserTypeExpr: userType,
			Identifier:   rt.Identifier,
			ContentType:  rt.ContentType,
			Views:        views,
		}
		for _, v := range views {
			v.Parent = nmt
		}
		return nmt
	}
	appendSuffix(userType.Attribute().Type, suffix)

	return userType
}

func renameType(dt design.DataType, name, suffix string) design.DataType {
	switch actual := dt.(type) {
	case design.UserType:
		rt := design.Dup(dt)
		if urt, ok := rt.(*design.UserTypeExpr); ok {
			urt.TypeName = name
		} else {
			rt.(*design.ResultTypeExpr).TypeName = name
		}
		appendSuffix(actual.Attribute().Type, suffix)
		return rt
	case *design.Object:
		rt := design.Dup(dt)
		appendSuffix(rt, suffix)
		return rt
	case *design.Array:
		rt := design.Dup(dt)
		appendSuffix(rt, suffix)
		return rt
	case *design.Map:
		rt := design.Dup(dt)
		appendSuffix(rt, suffix)
		return rt
	}
	return dt
}

func appendSuffix(dt design.DataType, suffix string) {
	switch actual := dt.(type) {
	case design.UserType:
		if ut, ok := actual.(*design.UserTypeExpr); ok {
			ut.TypeName = ut.TypeName + suffix
		} else {
			rt := actual.(*design.ResultTypeExpr)
			rt.TypeName = rt.TypeName + suffix
		}
		appendSuffix(actual.Attribute().Type, suffix)
	case *design.Object:
		for _, nat := range *actual {
			appendSuffix(nat.Attribute.Type, suffix)
		}
	case *design.Array:
		appendSuffix(actual.ElemType.Type, suffix)
	case *design.Map:
		appendSuffix(actual.KeyType.Type, suffix)
		appendSuffix(actual.ElemType.Type, suffix)
	}
}

func removeAttributes(attr, sub *design.MappedAttributeExpr) {
	codegen.WalkMappedAttr(sub, func(name, _ string, _ bool, _ *design.AttributeExpr) error {
		attr.Delete(name)
		if attr.Validation != nil {
			attr.Validation.RemoveRequired(name)
		}
		return nil
	})
}
