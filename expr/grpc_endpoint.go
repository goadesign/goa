// This file prepares, validates, and finalizes the gRPC transport contract for
// one service method.
package expr

import (
	"fmt"
	"slices"

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

const (
	// streamCompatMetaKey is the meta key that makes generated servers also
	// accept clients speaking the legacy stream protocol which predates the
	// typed stream envelope.
	streamCompatMetaKey = "grpc:stream:compat"
	// streamCompatLegacy is the only supported value of the stream
	// compatibility meta and selects the legacy metadata-based protocol.
	streamCompatLegacy = "v1"
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

// LegacyStreamCompat reports whether the generated server must also accept
// clients that speak the legacy stream protocol which carries the one-shot
// method payload in gRPC request metadata instead of a typed initial stream
// frame. It is enabled by setting Meta("grpc:stream:compat", "v1") on the
// method, the service or the API.
func (e *GRPCEndpointExpr) LegacyStreamCompat() bool {
	v, ok := e.streamCompatValue()
	return ok && v == streamCompatLegacy
}

// Validate validates the endpoint expression by checking if the request
// and responses contains the "rpc:tag" in the meta. It also makes sure
// that there is only one response per status code.
func (e *GRPCEndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if e.Name() == "" {
		verr.Add(e, "Endpoint name cannot be empty")
	}
	verr.Merge(e.validateStreamCompat())

	seenUnions := make(map[*Union]struct{})
	seenAttrs := make(map[*AttributeExpr]struct{})
	validateGRPCUnionShapes(e.MethodExpr.Payload, e.MethodExpr, verr, seenUnions, seenAttrs)
	validateGRPCUnionShapes(e.MethodExpr.StreamingPayload, e.MethodExpr, verr, seenUnions, seenAttrs)
	validateGRPCUnionShapes(e.MethodExpr.Result, e.MethodExpr, verr, seenUnions, seenAttrs)
	validateGRPCUnionShapes(e.MethodExpr.StreamingResult, e.MethodExpr, verr, seenUnions, seenAttrs)
	for _, er := range e.MethodExpr.Errors {
		validateGRPCUnionShapes(er.AttributeExpr, e.MethodExpr, verr, seenUnions, seenAttrs)
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
					found = slices.Contains(secAttrs, nat.Name)
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
	} else if hasMessage && hasMetadata {
		verr.Add(e, "Both request message and metadata are defined, but payload is not an object. Define either metadata or message or make payload an object type.")
	}

	// Validate response
	verr.Merge(e.Response.Validate(e))

	// Validate errors
	for _, er := range e.GRPCErrors {
		verr.Merge(er.Validate())
		// Custom object error types are rendered as protobuf messages so
		// their fields must define field numbers, mirroring the payload and
		// result checks above. Default ErrorResult errors travel in the gRPC
		// status and need no tags.
		if ee := e.MethodExpr.Error(er.Name); ee != nil && ee.Type != ErrorResult && IsObject(ee.Type) {
			verr.Merge(validateRPCTags(AsObject(ee.Type), e))
		}
	}
	verr.Merge(e.validateErrorMappings())
	return verr
}

// validateErrorMappings ensures inherited gRPC response policy describes the
// same concrete error value returned by the endpoint method.
func (e *GRPCEndpointExpr) validateErrorMappings() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	for _, mapping := range e.GRPCErrors {
		mapped, owner := mapping.mappedError()
		method := e.MethodExpr.Error(mapping.Name)
		if mapped == nil || method == nil || equivalentErrorAttributes(mapped.AttributeExpr, method.AttributeExpr) {
			continue
		}
		verr.Add(
			mapping.Response,
			`gRPC error mapping %q inherited from the %s uses error type %q, but method %q of service %q uses %q; both definitions must define the same error attribute (type, validations, defaults, and struct metadata)`,
			mapping.Name,
			owner,
			mapped.Type.Name(),
			e.MethodExpr.Name,
			e.MethodExpr.Service.Name,
			method.Type.Name(),
		)
	}
	return verr
}

func validateGRPCUnionShapes(att *AttributeExpr, parent eval.Expression, verr *eval.ValidationErrors, seenUnions map[*Union]struct{}, seenAttrs map[*AttributeExpr]struct{}) {
	if att == nil || att.Type == nil {
		return
	}
	if _, ok := seenAttrs[att]; ok {
		return
	}
	seenAttrs[att] = struct{}{}

	if u := AsUnion(att.Type); u != nil {
		if _, ok := seenUnions[u]; ok {
			return
		}
		seenUnions[u] = struct{}{}
		for _, ut := range u.Values {
			switch {
			case IsArray(ut.Attribute.Type):
				verr.Add(parent, "union type %s has array elements, not supported by gRPC", u.Name())
			case IsMap(ut.Attribute.Type):
				verr.Add(parent, "union type %s has map elements, not supported by gRPC", u.Name())
			}
			validateGRPCUnionShapes(ut.Attribute, parent, verr, seenUnions, seenAttrs)
		}
		return
	}

	if o := AsObject(att.Type); o != nil {
		for _, nat := range *o {
			validateGRPCUnionShapes(nat.Attribute, parent, verr, seenUnions, seenAttrs)
		}
		return
	}

	if ar := AsArray(att.Type); ar != nil {
		validateGRPCUnionShapes(ar.ElemType, parent, verr, seenUnions, seenAttrs)
		return
	}

	if m := AsMap(att.Type); m != nil {
		validateGRPCUnionShapes(m.KeyType, parent, verr, seenUnions, seenAttrs)
		validateGRPCUnionShapes(m.ElemType, parent, verr, seenUnions, seenAttrs)
		return
	}
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
		requirements := EffectiveSecurityRequirements(e.MethodExpr.Requirements)
		if reqLen := len(requirements); reqLen > 0 {
			e.Requirements = make([]*SecurityExpr, 0, reqLen)
			for _, req := range requirements {
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
					case BearerKind:
						field = TaggedAttribute(e.MethodExpr.Payload, "security:bearer")
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
		if ut, ok := e.MethodExpr.Payload.Type.(UserType); ok {
			// propagate the user set protobuf struct name from the user type to
			// the request message.
			if proto, ok := ut.Attribute().Meta.Last("struct:name:proto"); ok {
				if e.Request.Meta == nil {
					e.Request.Meta = ut.Attribute().Meta
				} else {
					e.Request.Meta["struct:name:proto"] = []string{proto}
				}
			}
		}
	} else {
		// method payload is not an object type.
		initAttrFromDesign(e.Request, e.MethodExpr.Payload)
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
				continue
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
			} else if !isMetadataEncodable(a.Type) {
				verr.Add(e, "%s metadata attribute %q must be a primitive or an array of primitives, got %s", metKind, nat.Name, a.Type.Name())
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
	var secAttrs []string

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
			case BearerKind:
				if field := TaggedAttribute(m.Payload, "security:bearer"); field != "" {
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

// streamCompatValue returns the value of the stream compatibility meta by
// looking up the endpoint, method, service and API expressions in that order.
func (e *GRPCEndpointExpr) streamCompatValue() (string, bool) {
	if v, ok := e.Meta.Last(streamCompatMetaKey); ok {
		return v, true
	}
	if v, ok := e.MethodExpr.Meta.Last(streamCompatMetaKey); ok {
		return v, true
	}
	if v, ok := e.Service.ServiceExpr.Meta.Last(streamCompatMetaKey); ok {
		return v, true
	}
	if v, ok := Root.API.Meta.Last(streamCompatMetaKey); ok {
		return v, true
	}
	return "", false
}

// validateStreamCompat validates the stream compatibility meta if set. The
// legacy stream protocol carries the one-shot method payload in gRPC metadata
// which can only encode primitive values and arrays of primitive values.
func (e *GRPCEndpointExpr) validateStreamCompat() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	value, ok := e.streamCompatValue()
	if !ok {
		return verr
	}
	if value != streamCompatLegacy {
		verr.Add(e, "invalid %q meta value %q: only %q is supported", streamCompatMetaKey, value, streamCompatLegacy)
		return verr
	}
	if !e.MethodExpr.IsPayloadStreaming() || e.MethodExpr.Payload.Type == Empty {
		// The meta only affects methods that combine a one-shot payload with
		// a streaming payload. Only report a missing effect when the meta is
		// set on the endpoint or method directly; service and API level metas
		// legitimately apply to a subset of the methods they cover.
		_, endpointLevel := e.Meta.Last(streamCompatMetaKey)
		_, methodLevel := e.MethodExpr.Meta.Last(streamCompatMetaKey)
		if endpointLevel || methodLevel {
			verr.Add(e, "%q meta requires the method to define both Payload and StreamingPayload", streamCompatMetaKey)
		}
		return verr
	}
	if obj := AsObject(e.MethodExpr.Payload.Type); obj != nil {
		metObj := AsObject(e.Metadata.Type)
		for _, nat := range *obj {
			if metObj.Attribute(nat.Name) != nil {
				// Attributes explicitly mapped to metadata travel in metadata
				// under both protocols and are validated separately.
				continue
			}
			if !isMetadataEncodable(nat.Attribute.Type) {
				verr.Add(e, "attribute %q of the method payload must be a primitive or an array of primitives to satisfy the %q meta", nat.Name, streamCompatMetaKey)
			}
		}
	} else if !isMetadataEncodable(e.MethodExpr.Payload.Type) {
		verr.Add(e, "the method payload must be a primitive, an array of primitives or an object to satisfy the %q meta", streamCompatMetaKey)
	}
	return verr
}

// isMetadataEncodable reports whether values of the given type can be carried
// in gRPC metadata, that is whether they can be encoded to and decoded from
// header strings.
func isMetadataEncodable(dt DataType) bool {
	if IsPrimitive(dt) {
		return true
	}
	arr := AsArray(dt)
	return arr != nil && IsPrimitive(arr.ElemType.Type)
}
