package design

import (
	"net/http"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// RequestBody returns an attribute describing the request body of the given
// endpoint. If the DSL defines a body explicitly via the Body function then the
// corresponding attribute is used. Otherwise the attribute is computed by
// removing the attributes of the method payload used to define headers and
// parameters.
func RequestBody(a *EndpointExpr) *design.AttributeExpr {
	if a.Body != nil {
		return a.Body
	}

	var (
		payload   = a.MethodExpr.Payload
		headers   = a.Headers
		params    = a.Params
		suffix    = "RequestBody"
		name      = codegen.Goify(a.Name(), true) + suffix
		userField string
		passField string
	)
	{
		obj := design.AsObject(payload.Type)
		if obj != nil {
			for _, at := range *obj {
				if _, ok := at.Attribute.Meta["security:username"]; ok {
					userField = at.Name
				}
				if _, ok := at.Attribute.Meta["security:password"]; ok {
					passField = at.Name
				}
				if userField != "" && passField != "" {
					break
				}
			}
		}
	}

	bodyOnly := headers.IsEmpty() && params.IsEmpty() && a.MapQueryParams == nil

	// 1. If Payload is not an object then check whether there are params or
	// headers defined and if so return empty type (payload encoded in
	// request params or headers) otherwise return payload type (payload
	// encoded in request body).
	if !design.IsObject(payload.Type) {
		if bodyOnly {
			payload = design.DupAtt(payload)
			renameType(payload, name, "RequestBody")
			return payload
		}
		return &design.AttributeExpr{Type: design.Empty}
	}

	// 2. Remove header and param attributes
	body := design.NewMappedAttributeExpr(payload)
	removeAttributes(body, headers)
	removeAttributes(body, params)
	if a.MapQueryParams != nil && *a.MapQueryParams != "" {
		removeAttribute(body, *a.MapQueryParams)
	}
	if userField != "" {
		removeAttribute(body, userField)
	}
	if passField != "" {
		removeAttribute(body, passField)
	}

	// 3. Return empty type if no attribute left
	if len(*design.AsObject(body.Type)) == 0 {
		return &design.AttributeExpr{Type: design.Empty}
	}

	// 4. Build computed user type
	att := body.Attribute()
	ut := &design.UserTypeExpr{
		AttributeExpr: att,
		TypeName:      name,
	}
	appendSuffix(ut.Attribute().Type, "RequestBody")

	return &design.AttributeExpr{
		Type:         ut,
		Validation:   att.Validation,
		UserExamples: att.UserExamples,
	}
}

// StreamingBody TODO
func StreamingBody(e *EndpointExpr) *design.AttributeExpr {
	if !e.MethodExpr.IsStreaming() || e.MethodExpr.Stream == design.ServerStreamKind {
		return nil
	}
	att := e.MethodExpr.StreamingPayload
	if !design.IsObject(att.Type) {
		return design.DupAtt(att)
	}
	suffix := "StreamingBody"
	ut := &design.UserTypeExpr{
		AttributeExpr: design.DupAtt(att),
		TypeName:      codegen.Goify(e.Name(), true) + suffix,
	}
	appendSuffix(ut.Attribute().Type, suffix)

	return &design.AttributeExpr{
		Type:         ut,
		Validation:   att.Validation,
		UserExamples: att.UserExamples,
	}
}

// ResponseBody returns an attribute representing the response body for the
// given endpoint and response. If the DSL defines a body explicitly via the
// Body function then the corresponding attribute is used. Otherwise the
// attribute is computed by removing the attributes of the method payload used
// to define headers.
func ResponseBody(a *EndpointExpr, resp *HTTPResponseExpr) *design.AttributeExpr {
	var name, suffix string
	if len(a.Responses) > 1 {
		suffix = codegen.Goify(http.StatusText(resp.StatusCode), true)
	}
	name = codegen.Goify(a.Name(), true) + suffix
	return buildResponseBody(name, a.MethodExpr.Result, resp)
}

// ErrorResponseBody returns an attribute describing the response body of a
// given error. If the DSL defines a body explicitly via the Body function then
// the corresponding attribute is returned. Otherwise the attribute is computed
// by removing the attributes of the error used to define headers and
// parameters.
func ErrorResponseBody(a *EndpointExpr, v *ErrorExpr) *design.AttributeExpr {
	var name, suffix string
	suffix = codegen.Goify(v.ErrorExpr.Name, true)
	name = codegen.Goify(a.Name(), true) + suffix
	return buildResponseBody(name, v.ErrorExpr.AttributeExpr, v.Response)
}

func buildResponseBody(name string, attr *design.AttributeExpr, resp *HTTPResponseExpr) *design.AttributeExpr {
	name += "ResponseBody"
	if attr == nil || attr.Type == design.Empty {
		return &design.AttributeExpr{Type: design.Empty}
	}
	if resp.Body != nil {
		return resp.Body
	}

	// 1. If attribute is not an object then check whether there are headers
	// defined and if so return empty type (attr encoded in response
	// headers) otherwise return renamed attr type (attr encoded in
	// response body).
	if !design.IsObject(attr.Type) {
		if resp.Headers.IsEmpty() {
			attr = design.DupAtt(attr)
			renameType(attr, name, "ResponseBody")
			return attr
		}
		return &design.AttributeExpr{Type: design.Empty}
	}

	// 2. Remove header attributes
	body := design.NewMappedAttributeExpr(attr)
	removeAttributes(body, resp.Headers)

	// 3. Return empty type if no attribute left
	if len(*design.AsObject(body.Type)) == 0 {
		return &design.AttributeExpr{Type: design.Empty}
	}

	// 4. Build computed user type
	userType := &design.UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	appendSuffix(userType.Attribute().Type, "ResponseBody")
	rt, isrt := attr.Type.(*design.ResultTypeExpr)
	if !isrt {
		return &design.AttributeExpr{Type: userType, Validation: userType.Validation}
	}
	views := make([]*design.ViewExpr, len(rt.Views))
	for i, v := range rt.Views {
		mv := design.NewMappedAttributeExpr(v.AttributeExpr)
		removeAttributes(mv, resp.Headers)
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
	return &design.AttributeExpr{Type: nmt, Validation: userType.Validation}
}

func renameType(att *design.AttributeExpr, name, suffix string) {
	rt := att.Type
	switch rt.(type) {
	case design.UserType:
		rt.(design.UserType).Rename(name)
		appendSuffix(rt.(design.UserType).Attribute().Type, suffix)
	case *design.Object:
		appendSuffix(rt, suffix)
	case *design.Array:
		appendSuffix(rt, suffix)
	case *design.Map:
		appendSuffix(rt, suffix)
	}
}

func appendSuffix(dt design.DataType, suffix string, seen ...map[string]struct{}) {
	var s map[string]struct{}
	if len(seen) > 0 {
		s = seen[0]
	} else {
		s = make(map[string]struct{})
		seen = append(seen, s)
	}
	switch actual := dt.(type) {
	case design.UserType:
		if _, ok := s[actual.ID()]; ok {
			return
		}
		actual.Rename(actual.Name() + suffix)
		s[actual.ID()] = struct{}{}
		appendSuffix(actual.Attribute().Type, suffix, seen...)
	case *design.Object:
		for _, nat := range *actual {
			appendSuffix(nat.Attribute.Type, suffix, seen...)
		}
	case *design.Array:
		appendSuffix(actual.ElemType.Type, suffix, seen...)
	case *design.Map:
		appendSuffix(actual.KeyType.Type, suffix, seen...)
		appendSuffix(actual.ElemType.Type, suffix, seen...)
	}
}

func removeAttributes(attr, sub *design.MappedAttributeExpr) {
	codegen.WalkMappedAttr(sub, func(name, _ string, _ bool, _ *design.AttributeExpr) error {
		removeAttribute(attr, name)
		return nil
	})
}

func removeAttribute(attr *design.MappedAttributeExpr, name string) {
	attr.Delete(name)
	if attr.Validation != nil {
		attr.Validation.RemoveRequired(name)
	}
	for _, ex := range attr.UserExamples {
		if m, ok := ex.Value.(map[string]interface{}); ok {
			delete(m, name)
		}
	}
}
