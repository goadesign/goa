package expr

import (
	"fmt"
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

	// 1. If Payload is a union type, then the request body is a struct with
	// two fields: the union type and its value.
	if IsUnion(payload.Type) {
		return unionToObject(payload, name, suffix, a.Service.Name())
	}

	// 2. If Payload is not an object then check whether there are
	// params, cookies or headers defined and if so return empty type
	// (payload encoded in request params or headers) otherwise return
	// payload type (payload encoded in request body).
	if !IsObject(payload.Type) {
		if bodyOnly {
			payload = DupAtt(payload)
			RemovePkgPath(payload)
			renameType(payload, name, suffix)
			return payload
		}
		return &AttributeExpr{Type: Empty}
	}

	// 3. Remove header, param and cookies attributes
	body := NewMappedAttributeExpr(payload)
	RemovePkgPath(body.AttributeExpr)
	extendBodyAttribute(body)
	removeAttributes(body, headers)
	removeAttributes(body, cookies)
	removeAttributes(body, params)
	if a.MapQueryParams != nil && *a.MapQueryParams != "" {
		removeAttribute(body, *a.MapQueryParams)
	}
	for att := range defaultRequestHeaderAttributes(a) {
		removeAttribute(body, att)
	}

	// 4. Return empty type if no attribute left
	if len(*AsObject(body.Type)) == 0 {
		return &AttributeExpr{Type: Empty}
	}

	// 5. Build computed user type
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
	if IsUnion(att.Type) {
		return unionToObject(att, e.Name(), "StreamingBody", e.Service.Name())
	}
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

	// 1. Handle the case where the body is set explicitly in the design.
	// We need to create a type with an endpoint specific response body type
	// name to handle the case where the same type is used by multiple
	// methods with potentially different result types.
	if resp.Body != nil {
		if IsUnion(resp.Body.Type) {
			return unionToObject(resp.Body, name, suffix, svc.Name())
		}
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

	// 2. Map unions to objects.
	if IsUnion(attr.Type) {
		return unionToObject(attr, name, suffix, svc.Name())
	}

	// 3. If attribute is not an object then check whether there are headers or
	// cookies defined and if so return empty type (attr encoded in response
	// header or cookie) otherwise return renamed attr type (attr encoded in
	// response body).
	if !IsObject(attr.Type) {
		if resp.Headers.IsEmpty() && resp.Cookies.IsEmpty() {
			attr = DupAtt(attr)
			RemovePkgPath(attr)
			renameType(attr, name, "Response") // Do not use ResponseBody as it could clash with name of element
			return attr
		}
		return &AttributeExpr{Type: Empty}
	}
	body := NewMappedAttributeExpr(attr)
	RemovePkgPath(body.AttributeExpr)
	extendBodyAttribute(body)

	// 4. Remove header and cookie attributes
	removeAttributes(body, resp.Headers)
	removeAttributes(body, resp.Cookies)

	// 5. Return empty type if no attribute left
	if len(*AsObject(body.Type)) == 0 {
		return &AttributeExpr{Type: Empty}
	}

	// 6. Build computed user type
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

// unionToObject returns an object adequate to serialize the given union type.
func unionToObject(att *AttributeExpr, name, suffix, svcName string) *AttributeExpr {
	values := AsUnion(att.Type).Values
	names := make([]interface{}, len(values))
	vals := make([]string, len(values))
	for i, nat := range values {
		names[i] = nat.Attribute.Type.Name()
		vals[i] = fmt.Sprintf("- %q", nat.Attribute.Type.Name())
	}
	obj := Object([]*NamedAttributeExpr{{
		"Type", &AttributeExpr{
			Type:        String,
			Description: "Union type name, one of:\n" + strings.Join(vals, "\n"),
			Validation:  &ValidationExpr{Values: names},
		}}, {
		"Value", &AttributeExpr{
			Type:         String,
			Description:  "JSON formatted union value",
			UserExamples: []*ExampleExpr{{Value: `"JSON"`}},
		}},
	})
	uatt := &AttributeExpr{
		Type:       &obj,
		Validation: &ValidationExpr{Required: []string{"Type", "Value"}},
	}
	ut := &UserTypeExpr{
		AttributeExpr: uatt,
		TypeName:      name,
		UID:           concat(svcName, "#", name),
	}
	wrapper := &AttributeExpr{Type: ut, Description: att.Description}
	renameType(wrapper, name, suffix)
	return wrapper
}

// concat concatenates the given strings with "smart(?) casing".
// The concatenation algorithm is:
//
// 1) If the first string contains underscores and starts with a lower case,
// the rest of the strings are converted to lower case and concatenated with
// underscores.
// e.g. concat("my_endpoint", "Request", "BODY") => "my_endpoint_request_body"
//
// 2) If the first string contains underscores and starts with a upper case,
// the rest of the strings are converted to title case and concatenated with
// underscores.
// e.g. concat("My_endpoint", "response", "body") => "My_endpoint_Response_Body"
//
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
			name += "_" + Title(strs[i])
		}
	default:
		name = Title(name)
		for i := 1; i < len(strs); i++ {
			name += Title(strs[i])
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

// RemovePkgPath traverses the given data type and removes the "struct:pkg:path"
// metadata from all the user type attributes.
func RemovePkgPath(attr *AttributeExpr) {
	walk(attr.Type, func(ut UserType) {
		delete(ut.Attribute().Meta, "struct:pkg:path")
	})
	for _, pt := range attr.Bases {
		if dt, ok := pt.(UserType); ok {
			RemovePkgPath(dt.Attribute())
		}
	}
}

// appendSuffix recursively traverses the given data type and appends the given
// suffix to all the user type names.
func appendSuffix(dt DataType, suffix string) {
	walk(dt, func(ut UserType) {
		ut.Rename(ut.Name() + suffix)
	})
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

// extendedBodyAttribute returns an attribute describing the HTTP
// request/response body type by merging any Bases and References to the parent
// attribute. This must be invoked during validation or to determine the actual
// body type by removing any headers/params/cookies.
func extendBodyAttribute(body *MappedAttributeExpr) {
	att := body.AttributeExpr
	if isEmpty(att) {
		return
	}
	for _, ref := range att.References {
		ru, ok := ref.(UserType)
		if !ok {
			continue
		}
		att.Inherit(ru.Attribute())
	}
	// unset references so that they don't get added back to the body type during
	// finalize
	att.References = nil
	for _, base := range att.Bases {
		ru, ok := base.(UserType)
		if !ok {
			continue
		}
		att.Merge(ru.Attribute())
	}
	// unset bases so that they don't get added back to the body type during
	// finalize
	att.Bases = nil
}

// walk traverses the given data type and invokes the given function for each
// user type it finds including dt itself.
func walk(dt DataType, do func(UserType)) {
	walkrec(dt, do, make(map[string]struct{}))
}

func walkrec(dt DataType, do func(UserType), seen map[string]struct{}) {
	switch dt := dt.(type) {
	case UserType:
		if _, ok := seen[dt.ID()]; ok {
			return
		}
		do(dt)
		seen[dt.ID()] = struct{}{}
		walkrec(dt.Attribute().Type, do, seen)
	case *Object:
		for _, nat := range *dt {
			walkrec(nat.Attribute.Type, do, seen)
		}
	case *Array:
		walkrec(dt.ElemType.Type, do, seen)
	case *Map:
		walkrec(dt.KeyType.Type, do, seen)
		walkrec(dt.ElemType.Type, do, seen)
	case *Union:
		for _, nat := range dt.Values {
			walkrec(nat.Attribute.Type, do, seen)
		}
	}
}
