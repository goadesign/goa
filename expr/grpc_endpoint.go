package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// GRPCEndpointExpr describes a gRPC endpoint. It embeds a MethodExpr
	// and adds gRPC specific properties.
	GRPCEndpointExpr struct {
		eval.DSLFunc
		// MethodExpr is the underlying method expression.
		MethodExpr *MethodExpr
		// Service is the parent service.
		Service *GRPCServiceExpr
		// Request is the message passed to the gRPC method.
		Request *AttributeExpr
		// StreamingRequest is the message passed to the gRPC method through a
		// stream.
		StreamingRequest *AttributeExpr
		// Responses is the success gRPC response from the method.
		Response *GRPCResponseExpr
		// GRPCErrors is the list of all the possible error gRPC responses.
		GRPCErrors []*GRPCErrorExpr
		// Metadata is the metadata to be sent in a gRPC request.
		Metadata *MappedAttributeExpr
		// Requirements is the list of security requirements for the gRPC endpoint.
		Requirements []*SecurityExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator, see dsl.Meta.
		Meta MetaExpr
	}
)

// Name of gRPC endpoint
func (e *GRPCEndpointExpr) Name() string {
	return e.MethodExpr.Name
}

// Description of gRPC endpoint
func (e *GRPCEndpointExpr) Description() string {
	return e.MethodExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (e *GRPCEndpointExpr) EvalName() string {
	var prefix, suffix string
	if e.Name() != "" {
		suffix = fmt.Sprintf("gRPC endpoint %#v", e.Name())
	} else {
		suffix = "unnamed gRPC endpoint"
	}
	if e.Service != nil {
		prefix = e.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Prepare initializes the Request and Response if nil.
func (e *GRPCEndpointExpr) Prepare() {
	if e.Request == nil {
		e.Request = &AttributeExpr{Type: Empty}
	}
	if e.Request.Validation == nil {
		e.Request.Validation = &ValidationExpr{}
	}
	if e.StreamingRequest == nil {
		e.StreamingRequest = &AttributeExpr{Type: Empty}
	}
	if e.StreamingRequest.Validation == nil {
		e.StreamingRequest.Validation = &ValidationExpr{}
	}
	if e.Metadata == nil {
		e.Metadata = NewEmptyMappedAttributeExpr()
	}
	if e.Metadata.Validation == nil {
		e.Metadata.Validation = &ValidationExpr{}
	}

	// Make sure there's a default response if none define explicitly
	if e.Response == nil {
		e.Response = &GRPCResponseExpr{StatusCode: 0}
	}
	e.Response.Prepare()

	// Error -> ResponseError
	methodErrors := map[string]struct{}{}
	for _, v := range e.GRPCErrors {
		methodErrors[v.Name] = struct{}{}
	}
	for _, me := range e.MethodExpr.Errors {
		if _, ok := methodErrors[me.Name]; ok {
			continue
		}
		methodErrors[me.Name] = struct{}{}
		var found bool
		for _, v := range e.Service.GRPCErrors {
			if me.Name == v.Name {
				e.GRPCErrors = append(e.GRPCErrors, v.Dup())
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Lookup undefined GRPC errors in API.
		for _, v := range Root.API.GRPC.Errors {
			if me.Name == v.Name {
				e.GRPCErrors = append(e.GRPCErrors, v.Dup())
			}
		}
	}
	// Inherit GRPC errors from service if the error has not added.
	for _, se := range e.Service.ServiceExpr.Errors {
		if _, ok := methodErrors[se.Name]; ok {
			continue
		}
		var found bool
		for _, resp := range e.Service.GRPCErrors {
			if se.Name == resp.Name {
				found = true
				e.GRPCErrors = append(e.GRPCErrors, resp.Dup())
				break
			}
		}
		if !found {
			for _, ae := range Root.API.GRPC.Errors {
				if se.Name == ae.Name {
					e.GRPCErrors = append(e.GRPCErrors, ae.Dup())
					break
				}
			}
		}
	}

	// Prepare responses
	for _, er := range e.GRPCErrors {
		er.Response.Prepare()
	}
}

// Validate validates the endpoint expression by checking if the request
// and responses contains the "rpc:tag" in the meta. It also makes sure
// that there is only one response per status code.
func (e *GRPCEndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if e.Name() == "" {
		verr.Add(e, "Endpoint name cannot be empty")
	}

	// error if payload, result, and error type define attribute of Any type
	// which is unsupported.
	verr.Merge(e.hasAnyType(e.MethodExpr.Payload, "Payload"))
	verr.Merge(e.hasAnyType(e.MethodExpr.Result, "Result"))
	for _, er := range e.MethodExpr.Errors {
		verr.Merge(e.hasAnyType(er.AttributeExpr, fmt.Sprintf("Error %q", er.Name)))
	}

	var hasMessage, hasMetadata bool
	// Validate request
	if e.Request.Type != Empty {
		hasMessage = true
		verr.Merge(e.Request.Validate("gRPC request message", e))
		verr.Merge(validateMessage(e.Request, e.MethodExpr.Payload, e, true))
	}
	if !e.Metadata.IsEmpty() {
		hasMetadata = true
		verr.Merge(e.Metadata.Validate("gRPC request metadata", e))
		verr.Merge(validateMetadata(e.Metadata, e.MethodExpr.Payload, e, true))
	}

	if pobj := AsObject(e.MethodExpr.Payload.Type); pobj != nil {
		secAttrs := getSecurityAttributes(e.MethodExpr)
		switch {
		case hasMessage && hasMetadata:
			// ensure the attributes defined in message are not defined in metadata.
			msgObj := AsObject(e.Request.Type)
			metObj := AsObject(e.Metadata.Type)
			for _, msgnat := range *msgObj {
				for _, metnat := range *metObj {
					if metnat.Name == msgnat.Name {
						verr.Add(e, "Attribute %q defined in both request message and metadata. Define the attribute in either message or metadata.", metnat.Name)
						break
					}
				}
			}
		case !hasMessage && !hasMetadata:
			// no request message or metadata is defined. Ensure that the method
			// payload attributes have "rpc:tag" set (except for security attributes
			// as they are added to request metadata by default)
			msgFields := &Object{}
			if len(secAttrs) > 0 {
				// add attributes to msgFields from the payload that are not
				// security attributes
				var found bool
				for _, nat := range *pobj {
					found = false
					for _, n := range secAttrs {
						if n == nat.Name {
							found = true
							break
						}
					}
					if !found {
						msgFields.Set(nat.Name, nat.Attribute)
					}
				}
			} else {
				msgFields = pobj
			}
			if len(*msgFields) > 0 {
				verr.Merge(validateRPCTags(msgFields, e))
			}
		}
	} else {
		if hasMessage && hasMetadata {
			verr.Add(e, "Both request message and metadata are defined, but payload is not an object. Define either metadata or message or make payload an object type.")
		}
	}

	// Validate response
	verr.Merge(e.Response.Validate(e))

	// Validate errors
	for _, er := range e.GRPCErrors {
		verr.Merge(er.Validate())
	}
	return verr
}

// Finalize ensures the request and response attributes are initialized.
func (e *GRPCEndpointExpr) Finalize() {
	if pobj := AsObject(e.MethodExpr.Payload.Type); pobj != nil {
		// addToMetadata adds the given field to metadata. tName maps the attribute
		// name to the given transport name.
		addToMetadata := func(field string, tName string) {
			attr := pobj.Attribute(field)
			e.Metadata.Type.(*Object).Set(field, attr)
			if tName != "" {
				e.Metadata.Map(tName, field)
			}
			if e.MethodExpr.Payload.IsRequired(field) {
				e.Metadata.Validation.AddRequired(field)
			}
		}

		// Initialize any security attributes in request metadata unless it is
		// specified explicitly in the request message via the DSL.
		if reqLen := len(e.MethodExpr.Requirements); reqLen > 0 {
			e.Requirements = make([]*SecurityExpr, 0, reqLen)
			for _, req := range e.MethodExpr.Requirements {
				dupReq := DupRequirement(req)
				for _, sch := range dupReq.Schemes {
					var field string
					switch sch.Kind {
					case NoKind:
						continue
					case BasicAuthKind:
						field = TaggedAttribute(e.MethodExpr.Payload, "security:username")
						sch.Name, sch.In = findKey(e, field)
						if sch.Name == "" {
							addToMetadata(field, "")
						}
						field = TaggedAttribute(e.MethodExpr.Payload, "security:password")
						sch.Name, sch.In = findKey(e, field)
						if sch.Name == "" {
							addToMetadata(field, "")
						}
						continue
					case APIKeyKind:
						field = TaggedAttribute(e.MethodExpr.Payload, "security:apikey:"+sch.SchemeName)
					case JWTKind:
						field = TaggedAttribute(e.MethodExpr.Payload, "security:token")
					case OAuth2Kind:
						field = TaggedAttribute(e.MethodExpr.Payload, "security:accesstoken")
					}
					sch.Name, sch.In = findKey(e, field)
					if sch.Name == "" {
						sch.Name = "authorization"
						addToMetadata(field, sch.Name)
					}
				}
				e.Requirements = append(e.Requirements, dupReq)
			}
		}

		// If endpoint defines streaming payload, then add the attributes in method
		// payload type to request metadata.
		if e.MethodExpr.StreamingPayload.Type != Empty {
			for _, nat := range *pobj {
				addToMetadata(nat.Name, "")
			}
		}

		// msgObj contains only the attributes in the method payload that must
		// be added to the request message type after removing attributes
		// specified in the request metadata.
		msgObj := Dup(pobj).(*Object)
		for _, nat := range *AsObject(e.Metadata.Type) {
			// initialize metadata attribute from method payload
			initAttrFromDesign(nat.Attribute, pobj.Attribute(nat.Name))
			if e.MethodExpr.Payload.IsRequired(nat.Name) {
				e.Metadata.Validation.AddRequired(nat.Name)
			}
			// remove metadata attributes from the message attributes
			msgObj.Delete(nat.Name)
		}

		// add any message attributes to request message if not added already
		if len(*msgObj) > 0 {
			if e.Request.Type == Empty {
				e.Request.Type = &Object{}
			}
			reqObj := AsObject(e.Request.Type)
			for _, nat := range *msgObj {
				if reqObj.Attribute(nat.Name) == nil {
					reqObj.Set(nat.Name, nat.Attribute)
				}
				if e.MethodExpr.Payload.IsRequired(nat.Name) {
					e.Request.Validation.AddRequired(nat.Name)
				}
			}
		}
		for _, nat := range *AsObject(e.Request.Type) {
			// initialize message attribute
			patt := DupAtt(pobj.Attribute(nat.Name))
			initAttrFromDesign(nat.Attribute, patt)
			if nat.Attribute.Meta == nil {
				nat.Attribute.Meta = patt.Meta
			} else {
				nat.Attribute.Meta.Merge(patt.Meta)
			}
		}
	} else {
		// method payload is not an object type.
		if e.MethodExpr.StreamingPayload.Type != Empty {
			// endpoint defines streaming payload. So add the method payload to
			// request metadata under "goa-payload" field
			e.Metadata.Type.(*Object).Set("goa_payload", e.MethodExpr.Payload)
			e.Metadata.Validation.AddRequired("goa_payload")
		} else {
			initAttrFromDesign(e.Request, e.MethodExpr.Payload)
		}
	}

	// Finalize streaming payload type if defined
	if e.MethodExpr.StreamingPayload.Type != Empty {
		attr := e.MethodExpr.StreamingPayload
		// If streaming payload is a user type, use the underlying attribute
		// for the grpc streaming request type. This ensures we are consistent
		// with how message types are finalized for code generation.
		if ut, ok := attr.Type.(UserType); ok {
			attr = ut.Attribute()
		}
		initAttrFromDesign(e.StreamingRequest, attr)
		if msgObj := AsObject(e.StreamingRequest.Type); msgObj != nil {
			for _, nat := range *msgObj {
				if e.MethodExpr.StreamingPayload.IsRequired(nat.Name) {
					e.StreamingRequest.Validation.AddRequired(nat.Name)
				}
			}
		}
	}

	// Finalize response
	e.Response.Finalize(e, e.MethodExpr.Result)

	// Finalize errors
	for _, gerr := range e.GRPCErrors {
		gerr.Finalize(e)
	}
}

// validateMessage validates the gRPC message. It compares the given message
// with the service type (Payload or Result) and ensures all the attributes
// defined in the message type are found in the service type and the attributes
// are set with unique "rpc:tag" numbers.
//
// msgAtt is the Request/Response message attribute. validateMessage assumes
// that the msgAtt is not Empty.
// serviceAtt is the Payload/Result attribute.
// e is the endpoint expression.
// req if true indicates the Request message is being validated.
func validateMessage(msgAtt, serviceAtt *AttributeExpr, e *GRPCEndpointExpr, req bool) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	msgKind := "Response"
	serviceKind := "Result"
	if req {
		msgKind = "Request"
		serviceKind = "Payload"
	}
	if isEmpty(serviceAtt) {
		verr.Add(e, "%s message is defined but %s is not defined in method", msgKind, serviceKind)
		return verr
	}

	if !IsObject(serviceAtt.Type) {
		// service type (payload or result) is a primitive, array, or map
		// The message type must have at most one field and that field must be
		// of the same type as the service type.
		msgObj := AsObject(msgAtt.Type)
		if flen := len(*msgObj); flen != 1 {
			verr.Add(e, "%s is not an object type. %s message should have at most 1 field. Got %d.", serviceKind, msgKind, flen)
		} else {
			for _, f := range *msgObj {
				if f.Attribute.Type != serviceAtt.Type {
					verr.Add(e, "%s message field %q is %q type but the %s type is %q.", msgKind, f.Name, f.Attribute.Type.Name(), serviceKind, serviceAtt.Type.Name())
				}
			}
		}
	} else {
		// service type is an object. Verify the attributes defined in the
		// message are found in the service type.
		// msgFields will contain the attributes from the service type that has the
		// same name as the message attributes so that we can validate the
		// rpc:tag in the meta.
		msgFields := &Object{}
		for _, nat := range *AsObject(msgAtt.Type) {
			if a := serviceAtt.Find(nat.Name); a != nil {
				msgFields.Set(nat.Name, a)
				break
			}
			verr.Add(e, "%s message attribute %q is not found in %s", msgKind, nat.Name, serviceKind)
		}
		// validate rpc:tag in meta for the message fields
		verr.Merge(validateRPCTags(msgFields, e))
	}
	return verr
}

// validateRPCTags verifies whether every attribute in the object type has
// "rpc:tag" set in the meta and the tag numbers are unique.
func validateRPCTags(fields *Object, e *GRPCEndpointExpr) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	foundRPC := make(map[string]string)
	for _, nat := range *fields {
		if IsUnion(nat.Attribute.Type) {
			continue
		}
		if tag, ok := nat.Attribute.FieldTag(); !ok {
			verr.Add(e, "attribute %q does not have \"rpc:tag\" defined in the meta, use \"Field\" to define the attribute of a type used in a gRPC method", nat.Name)
		} else if a, ok := foundRPC[tag]; ok {
			verr.Add(e, "field number %s in attribute %q already exists for attribute %q", tag, nat.Name, a)
		} else {
			foundRPC[tag] = nat.Name
		}
	}
	return verr
}

// validateMetadata validates the gRPC metadata. It compares the given metadata
// with the service type (Payload or Result) and ensures all the attributes
// defined in the metadata type are found in the service type.
//
// metAtt is the Request/Response metadata attribute. validateMetadata assumes
// that the metAtt is not Empty.
// serviceAtt is the Payload/Result attribute.
// e is the endpoint expression.
// req if true indicates the Request metadata is being validated.
func validateMetadata(metAtt *MappedAttributeExpr, serviceAtt *AttributeExpr, e *GRPCEndpointExpr, req bool) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	metKind := "Response"
	serviceKind := "Result"
	if req {
		metKind = "Request"
		serviceKind = "Payload"
	}
	if isEmpty(serviceAtt) {
		verr.Add(e, "%s metadata is defined but %s is not defined in method", metKind, serviceKind)
		return verr
	}
	if IsObject(serviceAtt.Type) {
		// service type is an object type. Ensure the attributes defined in
		// the metadata are found in the service type.
		for _, nat := range *AsObject(metAtt.Type) {
			if a := serviceAtt.Find(nat.Name); a == nil {
				verr.Add(e, "%s metadata attribute %q is not found in %s", metKind, nat.Name, serviceKind)
			}
		}
	} else {
		verr.Add(e, "%s metadata is defined but method %s is not an object type", metKind, serviceKind)
	}
	return verr
}

// getSecurityAttributes returns the attributes that describes a security
// scheme from a method expression.
func getSecurityAttributes(m *MethodExpr) []string {
	secAttrs := []string{}
	for _, req := range m.Requirements {
		for _, sch := range req.Schemes {
			switch sch.Kind {
			case BasicAuthKind:
				if field := TaggedAttribute(m.Payload, "security:username"); field != "" {
					secAttrs = append(secAttrs, field)
				}
				if field := TaggedAttribute(m.Payload, "security:password"); field != "" {
					secAttrs = append(secAttrs, field)
				}
			case APIKeyKind:
				if field := TaggedAttribute(m.Payload, "security:apikey:"+sch.SchemeName); field != "" {
					secAttrs = append(secAttrs, field)
				}
			case JWTKind:
				if field := TaggedAttribute(m.Payload, "security:token"); field != "" {
					secAttrs = append(secAttrs, field)
				}
			case OAuth2Kind:
				if field := TaggedAttribute(m.Payload, "security:accesstoken"); field != "" {
					secAttrs = append(secAttrs, field)
				}
			}
		}
	}
	return secAttrs
}

// hasAnyType recurses through the given attribute and returns validation error
// if any attribute is of Any type.
func (e *GRPCEndpointExpr) hasAnyType(a *AttributeExpr, typ string, seen ...map[string]struct{}) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if a.Type == Any {
		verr.Add(e, "%s type is Any type which is not supported in gRPC", typ)
	}
	switch actual := a.Type.(type) {
	case UserType:
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		if _, ok := s[actual.ID()]; ok {
			return verr
		}
		s[actual.ID()] = struct{}{}
		verr.Merge(e.hasAnyType(actual.Attribute(), typ, seen...))
	case *Array:
		if IsPrimitive(actual.ElemType.Type) {
			if actual.ElemType.Type == Any {
				verr.Add(e, "Array element type is Any type which is not supported in gRPC")
			}
			return verr
		}
		verr.Merge(e.hasAnyType(actual.ElemType, typ, seen...))
	case *Map:
		if IsPrimitive(actual.KeyType.Type) {
			if actual.KeyType.Type == Any {
				verr.Add(e, "Map key type is Any type which is not supported in gRPC")
			}
		} else {
			verr.Merge(e.hasAnyType(actual.KeyType, typ, seen...))
		}
		if IsPrimitive(actual.ElemType.Type) {
			if actual.ElemType.Type == Any {
				verr.Add(e, "Map element type is Any type which is not supported in gRPC")
			}
			return verr
		}
		verr.Merge(e.hasAnyType(actual.ElemType, typ, seen...))
	case *Object:
		for _, nat := range *actual {
			if IsPrimitive(nat.Attribute.Type) {
				if nat.Attribute.Type == Any {
					verr.Add(e, "Attribute %q is Any type which is not supported in gRPC", nat.Name)
				}
				continue
			}
			verr.Merge(e.hasAnyType(nat.Attribute, typ, seen...))
		}
	}
	return verr
}
