package expr

import (
	"net/http"
)

// httpRequestBody returns an attribute describing the HTTP request body of the
// given endpoint. If the DSL defines a body explicitly via the Body function
// then the corresponding attribute is used. Otherwise the attribute is computed
// by removing the attributes of the method payload used to define headers and
// parameters.
func httpRequestBody(a *HTTPEndpointExpr) *AttributeExpr {
	if a.Body != nil {
		return a.Body
	}

	const suffix = "RequestBody"
	var (
		payload   = a.MethodExpr.Payload
		headers   = a.Headers
		params    = a.Params
		name      = a.Name() + suffix
		userField string
		passField string
	)
	{
		obj := AsObject(payload.Type)
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
	if !IsObject(payload.Type) {
		if bodyOnly {
			payload = DupAtt(payload)
			renameType(payload, name, suffix)
			return payload
		}
		return &AttributeExpr{Type: Empty}
	}

	// 2. Remove header and param attributes
	body := NewMappedAttributeExpr(payload)
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
	if len(*AsObject(body.Type)) == 0 {
		return &AttributeExpr{Type: Empty}
	}

	// 4. Build computed user type
	att := body.Attribute()
	ut := &UserTypeExpr{
		AttributeExpr: att,
		TypeName:      name,
	}
	appendSuffix(ut.Attribute().Type, suffix)

	return &AttributeExpr{
		Type:         ut,
		Validation:   att.Validation,
		UserExamples: att.UserExamples,
	}
}

// httpStreamingBody returns an attribute representing the structs being
// streamed via websocket.
func httpStreamingBody(e *HTTPEndpointExpr) *AttributeExpr {
	if !e.MethodExpr.IsStreaming() || e.MethodExpr.Stream == ServerStreamKind {
		return nil
	}
	att := e.MethodExpr.StreamingPayload
	if !IsObject(att.Type) {
		return DupAtt(att)
	}
	const suffix = "StreamingBody"
	ut := &UserTypeExpr{
		AttributeExpr: DupAtt(att),
		TypeName:      e.Name() + suffix,
	}
	appendSuffix(ut.Attribute().Type, suffix)

	return &AttributeExpr{
		Type:         ut,
		Validation:   att.Validation,
		UserExamples: att.UserExamples,
	}
}

// httpResponseBody returns an attribute representing the HTTP response body for
// the given endpoint and response. If the DSL defines a body explicitly via the
// Body function then the corresponding attribute is used. Otherwise the
// attribute is computed by removing the attributes of the method payload used
// to define headers.
func httpResponseBody(a *HTTPEndpointExpr, resp *HTTPResponseExpr) *AttributeExpr {
	var name, suffix string
	if len(a.Responses) > 1 {
		suffix = http.StatusText(resp.StatusCode)
	}
	name = a.Name() + suffix
	return buildHTTPResponseBody(name, a.MethodExpr.Result, resp)
}

// httpErrorResponseBody returns an attribute describing the response body of a
// given error. If the DSL defines a body explicitly via the Body function then
// the corresponding attribute is returned. Otherwise the attribute is computed
// by removing the attributes of the error used to define headers and
// parameters.
func httpErrorResponseBody(a *HTTPEndpointExpr, v *HTTPErrorExpr) *AttributeExpr {
	name := a.Name() + "_" + v.ErrorExpr.Name
	return buildHTTPResponseBody(name, v.ErrorExpr.AttributeExpr, v.Response)
}

func buildHTTPResponseBody(name string, attr *AttributeExpr, resp *HTTPResponseExpr) *AttributeExpr {
	const suffix = "ResponseBody"
	name += suffix
	if attr == nil || attr.Type == Empty {
		return &AttributeExpr{Type: Empty}
	}
	if resp.Body != nil {
		return resp.Body
	}

	// 1. If attribute is not an object then check whether there are headers
	// defined and if so return empty type (attr encoded in response
	// headers) otherwise return renamed attr type (attr encoded in
	// response body).
	if !IsObject(attr.Type) {
		if resp.Headers.IsEmpty() {
			attr = DupAtt(attr)
			renameType(attr, name, suffix)
			return attr
		}
		return &AttributeExpr{Type: Empty}
	}

	// 2. Remove header attributes
	body := NewMappedAttributeExpr(attr)
	removeAttributes(body, resp.Headers)

	// 3. Return empty type if no attribute left
	if len(*AsObject(body.Type)) == 0 {
		return &AttributeExpr{Type: Empty}
	}

	// 4. Build computed user type
	userType := &UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
	}
	appendSuffix(userType.Attribute().Type, suffix)
	rt, isrt := attr.Type.(*ResultTypeExpr)
	if !isrt {
		return &AttributeExpr{Type: userType, Validation: userType.Validation}
	}
	views := make([]*ViewExpr, len(rt.Views))
	for i, v := range rt.Views {
		mv := NewMappedAttributeExpr(v.AttributeExpr)
		removeAttributes(mv, resp.Headers)
		nv := &ViewExpr{
			AttributeExpr: mv.Attribute(),
			Name:          v.Name,
		}
		views[i] = nv
	}
	nmt := &ResultTypeExpr{
		UserTypeExpr: userType,
		Identifier:   rt.Identifier,
		ContentType:  rt.ContentType,
		Views:        views,
	}
	for _, v := range views {
		v.Parent = nmt
	}
	return &AttributeExpr{Type: nmt, Validation: userType.Validation}
}

func renameType(att *AttributeExpr, name, suffix string) {
	rt := att.Type
	switch rt.(type) {
	case UserType:
		rt.(UserType).Rename(name)
		appendSuffix(rt.(UserType).Attribute().Type, suffix)
	case *Object:
		appendSuffix(rt, suffix)
	case *Array:
		appendSuffix(rt, suffix)
	case *Map:
		appendSuffix(rt, suffix)
	}
}

func appendSuffix(dt DataType, suffix string, seen ...map[string]struct{}) {
	var s map[string]struct{}
	if len(seen) > 0 {
		s = seen[0]
	} else {
		s = make(map[string]struct{})
		seen = append(seen, s)
	}
	switch actual := dt.(type) {
	case UserType:
		if _, ok := s[actual.ID()]; ok {
			return
		}
		actual.Rename(actual.Name() + suffix)
		s[actual.ID()] = struct{}{}
		appendSuffix(actual.Attribute().Type, suffix, seen...)
	case *Object:
		for _, nat := range *actual {
			appendSuffix(nat.Attribute.Type, suffix, seen...)
		}
	case *Array:
		appendSuffix(actual.ElemType.Type, suffix, seen...)
	case *Map:
		appendSuffix(actual.KeyType.Type, suffix, seen...)
		appendSuffix(actual.ElemType.Type, suffix, seen...)
	}
}

func removeAttributes(attr, sub *MappedAttributeExpr) {
	o := AsObject(sub.Type)
	for _, nat := range *o {
		removeAttribute(attr, nat.Name)
	}
}

func removeAttribute(attr *MappedAttributeExpr, name string) {
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
