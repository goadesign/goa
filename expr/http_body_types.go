package expr

import (
	"net/http"
	"strings"
	"unicode"
)

// defaultRequestHeaderAttributes returns a map keyed by the names of the
// payload attributes that should come from the request HTTP headers by default.
// This includes mapping done for certain authorization schemes (basic auth,
// JWT, OAuth). The corresponding boolean value indicates whether the value maps
// directly to a payload attribute (true) or whether the value is used to
// compute the payload attribute (false). The only case where the value is
// computed by the generated code at this point is for basic authorization (the
// single "Authorization" header is used to compute both the username and
// password attributes).
func defaultRequestHeaderAttributes(e *HTTPEndpointExpr) map[string]bool {
	if len(e.MethodExpr.Requirements) == 0 {
		return nil
	}
	headers := make(map[string]bool)
	for _, req := range e.MethodExpr.Requirements {
		for _, sch := range req.Schemes {
			var field string
			switch sch.Kind {
			case NoKind:
				continue
			case BasicAuthKind:
				user := TaggedAttribute(e.MethodExpr.Payload, "security:username")
				if name, _ := findKey(e, user); name == "" {
					// No explicit mapping, use HTTP header by default
					headers[user] = false
				}
				pass := TaggedAttribute(e.MethodExpr.Payload, "security:password")
				if name, _ := findKey(e, pass); name == "" {
					// No explicit mapping, use HTTP header by default
					headers[pass] = false
				}
				continue
			case APIKeyKind:
				field = TaggedAttribute(e.MethodExpr.Payload, "security:apikey:"+sch.SchemeName)
			case JWTKind:
				field = TaggedAttribute(e.MethodExpr.Payload, "security:token")
			case OAuth2Kind:
				field = TaggedAttribute(e.MethodExpr.Payload, "security:accesstoken")
			}
			if name, _ := findKey(e, field); name == "" {
				// No explicit mapping, use HTTP header by default
				headers[field] = true
			}
		}
	}
	return headers
}

// httpRequestBody returns an attribute describing the HTTP request body of the
// given endpoint. If the DSL defines a body explicitly via the Body function
// then the corresponding attribute is used. Otherwise the attribute is computed
// by removing the attributes of the method payload used to define headers and
// parameters.
func httpRequestBody(a *HTTPEndpointExpr) *AttributeExpr {
	const suffix = "RequestBody"
	var (
		name = concat(a.Name(), "Request", "Body")
	)
	if a.Body != nil {
		a.Body = DupAtt(a.Body)
		renameType(a.Body, name, suffix)
		return a.Body
	}

	var (
		payload  = a.MethodExpr.Payload
		headers  = a.Headers
		cookies  = a.Cookies
		params   = a.Params
		bodyOnly = headers.IsEmpty() && params.IsEmpty() && cookies.IsEmpty() && a.MapQueryParams == nil
	)

	// 1. If Payload is not an object then check whether there are params,
	// cookies or headers defined and if so return empty type (payload encoded
	// in request params or headers) otherwise return payload type (payload
	// encoded in request body).
	if !IsObject(payload.Type) {
		if bodyOnly {
			payload = DupAtt(payload)
			renameType(payload, name, suffix)
			return payload
		}
		return &AttributeExpr{Type: Empty}
	}

	// 2. Remove header, param and cookies attributes
	body := NewMappedAttributeExpr(payload)
	removeAttributes(body, headers)
	removeAttributes(body, cookies)
	removeAttributes(body, params)
	if a.MapQueryParams != nil && *a.MapQueryParams != "" {
		removeAttribute(body, *a.MapQueryParams)
	}
	for att := range defaultRequestHeaderAttributes(a) {
		removeAttribute(body, att)
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
		UID:           a.Service.Name() + "#" + a.Name(),
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
		TypeName:      concat(e.Name(), "Streaming", "Body"),
		UID:           concat(e.Service.Name(), e.Name(), "Streaming", "Body"),
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
// to define cookies and headers.
func httpResponseBody(a *HTTPEndpointExpr, resp *HTTPResponseExpr) *AttributeExpr {
	var name, suffix string
	if len(a.Responses) > 1 {
		suffix = http.StatusText(resp.StatusCode)
	}
	name = a.Name() + suffix
	return buildHTTPResponseBody(name, a.MethodExpr.Result, resp, a.Service)
}

// httpErrorResponseBody returns an attribute describing the response body of a
// given error. If the DSL defines a body explicitly via the Body function then
// the corresponding attribute is returned. Otherwise the attribute is computed
// by removing the attributes of the error used to define cookies, headers and
// parameters.
func httpErrorResponseBody(e *HTTPEndpointExpr, v *HTTPErrorExpr) *AttributeExpr {
	name := e.Name() + "_" + v.ErrorExpr.Name
	return buildHTTPResponseBody(name, v.ErrorExpr.AttributeExpr, v.Response, e.Service)
}

func buildHTTPResponseBody(name string, attr *AttributeExpr, resp *HTTPResponseExpr, svc *HTTPServiceExpr) *AttributeExpr {
	const suffix = "ResponseBody"
	name = concat(name, "Response", "Body")
	if attr == nil || attr.Type == Empty {
		return &AttributeExpr{Type: Empty}
	}

	// 0. Handle the case where the body is set explicitely in the design.
	// We need to create a type with an endpoint specific response body type
	// name to handle the case where the same type is used by multiple
	// methods with potentially different result types.
	if resp.Body != nil {
		if !IsObject(resp.Body.Type) {
			return resp.Body
		}
		if len(*AsObject(resp.Body.Type)) == 0 {
			return &AttributeExpr{Type: Empty}
		}
		att := DupAtt(resp.Body)
		renameType(att, name, suffix)
		return att
	}

	// 1. If attribute is not an object then check whether there are headers or
	// cookies defined and if so return empty type (attr encoded in response
	// header or cookie) otherwise return renamed attr type (attr encoded in
	// response body).
	if !IsObject(attr.Type) {
		if resp.Headers.IsEmpty() && resp.Cookies.IsEmpty() {
			attr = DupAtt(attr)
			renameType(attr, name, "Response") // Do not use ResponseBody as it could clash with name of element
			return attr
		}
		return &AttributeExpr{Type: Empty}
	}
	body := NewMappedAttributeExpr(attr)

	// 2. Remove header and cookie attributes
	removeAttributes(body, resp.Headers)
	removeAttributes(body, resp.Cookies)

	// 3. Return empty type if no attribute left
	if len(*AsObject(body.Type)) == 0 {
		return &AttributeExpr{Type: Empty}
	}

	// 4. Build computed user type
	userType := &UserTypeExpr{
		AttributeExpr: body.Attribute(),
		TypeName:      name,
		UID:           concat(svc.Name(), "#", name),
	}

	// Remember original type name for example to generate friendly OpenAPI
	// specs.
	if t, ok := attr.Type.(UserType); ok {
		userType.AttributeExpr.AddMeta("name:original", t.Name())
	}

	appendSuffix(userType.Attribute().Type, suffix)
	rt, isrt := attr.Type.(*ResultTypeExpr)
	if !isrt {
		return &AttributeExpr{
			Type:       userType,
			Validation: userType.Validation,
			Meta:       attr.Meta,
		}
	}
	views := make([]*ViewExpr, len(rt.Views))
	for i, v := range rt.Views {
		mv := NewMappedAttributeExpr(v.AttributeExpr)
		removeAttributes(mv, resp.Headers)
		removeAttributes(mv, resp.Cookies)
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
	return &AttributeExpr{
		Type:       nmt,
		Validation: userType.Validation,
		Meta:       attr.Meta,
	}
}

// buildBodyTypeName concatenates the given strings to generate the
// endpoint's body type name.
//
// The concatenation algorithm is:
// 1) If the first string contains underscores and starts with a lower case,
// the rest of the strings are converted to lower case and concatenated with
// underscores.
// e.g. concat("my_endpoint", "Request", "BODY") => "my_endpoint_request_body"
// 2) If the first string contains underscores and starts with a upper case,
// the rest of the strings are converted to title case and concatenated with
// underscores.
// e.g. concat("My_endpoint", "response", "body") => "My_endpoint_Response_Body"
// 3) If the first string is a single word or camelcased, the rest of the
// strings are concatenated to form a valid upper camelcase.
// e.g. concat("myEndpoint", "streaming", "Body") => "MyEndpointStreamingBody"
//
func concat(strs ...string) string {
	if len(strs) == 1 {
		return strs[0]
	}

	// hasUnderscore returns true if the string has at least one underscore.
	hasUnderscore := func(str string) bool {
		for i := 0; i < len(str); i++ {
			if rune(str[i]) == '_' {
				return true
			}
		}
		return false
	}

	// isLower returns true if the first letter in the screen is lower-case.
	isLower := func(str string) bool {
		return unicode.IsLower(rune(str[0]))
	}

	name := strs[0]
	switch {
	case isLower(name) && hasUnderscore(name):
		for i := 1; i < len(strs); i++ {
			name += "_" + strings.ToLower(strs[i])
		}
	case !isLower(name) && hasUnderscore(name):
		for i := 1; i < len(strs); i++ {
			name += "_" + strings.Title(strs[i])
		}
	default:
		name = strings.Title(name)
		for i := 1; i < len(strs); i++ {
			name += strings.Title(strs[i])
		}
	}
	return name
}

func renameType(att *AttributeExpr, name, suffix string) {
	rt := att.Type
	switch rtt := rt.(type) {
	case UserType:
		rtt.Rename(name)
		appendSuffix(rtt.Attribute().Type, suffix)
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

// extendedHTTPRequestBody returns an attribute describing the HTTP request body.
// This is used only in the validation phase to figure out the request body when
// method Payload extends or references other types.
func extendedHTTPRequestBody(a *HTTPEndpointExpr) *AttributeExpr {
	const suffix = "RequestBody"
	var (
		name = concat(a.Name(), "Request", "Body")
	)
	if a.Body != nil {
		a.Body = DupAtt(a.Body)
		renameType(a.Body, name, suffix)
		return a.Body
	}

	var (
		payload  = a.MethodExpr.Payload
		headers  = a.Headers
		params   = a.Params
		bodyOnly = headers.IsEmpty() && params.IsEmpty() && a.MapQueryParams == nil
	)

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

	// Merge extended and referenced types
	payload = DupAtt(payload)
	for _, ref := range payload.References {
		ru, ok := ref.(UserType)
		if !ok {
			continue
		}
		payload.Inherit(ru.Attribute())
	}
	for _, base := range payload.Bases {
		ru, ok := base.(UserType)
		if !ok {
			continue
		}
		payload.Merge(ru.Attribute())
	}

	// 2. Remove header and param attributes
	body := NewMappedAttributeExpr(payload)
	removeAttributes(body, headers)
	removeAttributes(body, params)
	if a.MapQueryParams != nil && *a.MapQueryParams != "" {
		removeAttribute(body, *a.MapQueryParams)
	}
	for att := range defaultRequestHeaderAttributes(a) {
		removeAttribute(body, att)
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
		UID:           a.Service.Name() + "#" + a.Name(),
	}
	appendSuffix(ut.Attribute().Type, suffix)

	return &AttributeExpr{
		Type:         ut,
		Validation:   att.Validation,
		UserExamples: att.UserExamples,
	}
}
