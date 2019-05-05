package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// GRPCResponseExpr defines a gRPC response including its status code, result
	// type, and metadata.
	GRPCResponseExpr struct {
		// gRPC status code
		StatusCode int
		// Response description
		Description string
		// Response Message if any
		Message *AttributeExpr
		// Parent expression, one of EndpointExpr, ServiceExpr or
		// RootExpr.
		Parent eval.Expression
		// Headers is the header metadata to be sent in the gRPC response.
		Headers *MappedAttributeExpr
		// Trailers is the trailer metadata to be sent in the gRPC response.
		Trailers *MappedAttributeExpr
		// Meta is a list of key/value pairs.
		Meta MetaExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (r *GRPCResponseExpr) EvalName() string {
	var suffix string
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.EvalName())
	}
	return "gRPC response" + suffix
}

// Prepare makes sure the response message and metadata are initialized.
func (r *GRPCResponseExpr) Prepare() {
	if r.Message == nil {
		r.Message = &AttributeExpr{Type: Empty}
	}
	if r.Message.Validation == nil {
		r.Message.Validation = &ValidationExpr{}
	}
	if r.Headers == nil {
		r.Headers = NewEmptyMappedAttributeExpr()
	}
	if r.Headers.Validation == nil {
		r.Headers.Validation = &ValidationExpr{}
	}
	if r.Trailers == nil {
		r.Trailers = NewEmptyMappedAttributeExpr()
	}
	if r.Trailers.Validation == nil {
		r.Trailers.Validation = &ValidationExpr{}
	}
}

// Validate checks that the response definition is consistent: its status is set
// and the result type definition if any is valid.
func (r *GRPCResponseExpr) Validate(e *GRPCEndpointExpr) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	var hasMessage, hasHeaders, hasTrailers bool
	if r.Message.Type != Empty {
		hasMessage = true
		verr.Merge(r.Message.Validate("gRPC response message", r))
		verr.Merge(validateMessage(r.Message, e.MethodExpr.Result, e, false))
	}
	if !r.Headers.IsEmpty() {
		hasHeaders = true
		verr.Merge(r.Headers.Validate("gRPC response header metadata", r))
		verr.Merge(validateMetadata(r.Headers, e.MethodExpr.Result, e, false))
	}
	if !r.Trailers.IsEmpty() {
		hasTrailers = true
		verr.Merge(r.Trailers.Validate("gRPC response trailer metadata", r))
		verr.Merge(validateMetadata(r.Trailers, e.MethodExpr.Result, e, false))
	}

	if robj := AsObject(e.MethodExpr.Result.Type); robj != nil {
		switch {
		case hasMessage && hasHeaders:
			// ensure the attributes defined in message are not defined in
			// header metadata.
			metObj := AsObject(r.Headers.Type)
			for _, nat := range *AsObject(r.Message.Type) {
				if metObj.Attribute(nat.Name) != nil {
					verr.Add(e, "Attribute %q defined in both response message and header metadata. Define the attribute in either message or header metadata.", nat.Name)
				}
			}
		case hasMessage && hasTrailers:
			// ensure the attributes defined in message are not defined in
			// trailer metadata.
			metObj := AsObject(r.Trailers.Type)
			for _, nat := range *AsObject(r.Message.Type) {
				if metObj.Attribute(nat.Name) != nil {
					verr.Add(e, "Attribute %q defined in both response message and trailer metadata. Define the attribute in either message or trailer metadata.", nat.Name)
				}
			}
		case hasHeaders && hasTrailers:
			// ensure the attributes defined in header metadata are not defined in
			// trailer metadata
			hdrObj := AsObject(r.Headers.Type)
			for _, nat := range *AsObject(r.Trailers.Type) {
				if hdrObj.Attribute(nat.Name) != nil {
					verr.Add(e, "Attribute %q defined in both response header and trailer metadata. Define the attribute in either header or trailer metadata.", nat.Name)
				}
			}
		case !hasMessage && !hasHeaders && !hasTrailers:
			// no response message or metadata is defined. Ensure that the method
			// result attributes have "rpc:tag" set
			validateRPCTags(robj, e)
		}
	} else {
		switch {
		case hasMessage && hasHeaders:
			verr.Add(e, "Both response message and header metadata are defined, but result is not an object. Define either header metadata or message or make result an object type.")
		case hasMessage && hasTrailers:
			verr.Add(e, "Both response message and trailer metadata are defined, but result is not an object. Define either trailer metadata or message or make result an object type.")
		case hasHeaders && hasTrailers:
			verr.Add(e, "Both response header and trailer metadata are defined, but result is not an object. Define either trailer or header metadata or make result an object type.")
		}
	}

	return verr
}

// Finalize ensures that the response message type is set. If Message DSL is
// used to set the response message then the message type is set by mapping
// the attributes to the method Result expression. If no response message set
// explicitly, the message is set from the method Result expression.
func (r *GRPCResponseExpr) Finalize(a *GRPCEndpointExpr, svcAtt *AttributeExpr) {
	r.Parent = a

	if svcObj := AsObject(svcAtt.Type); svcObj != nil {
		// msgObj contains only the attributes in the method result that must
		// be added to the response message type after removing attributes
		// specified in the response metadata.
		msgObj := Dup(svcObj).(*Object)
		// Initialize response header metadata if present
		for _, nat := range *AsObject(r.Headers.Type) {
			// initialize metadata attribute from method result
			initAttrFromDesign(nat.Attribute, svcObj.Attribute(nat.Name))
			if svcAtt.IsRequired(nat.Name) {
				r.Headers.Validation.AddRequired(nat.Name)
			}
			// remove metadata attributes from the message attributes
			msgObj.Delete(nat.Name)
		}
		// Initialize response trailer metadata if present
		for _, nat := range *AsObject(r.Trailers.Type) {
			// initialize metadata attribute from method result
			initAttrFromDesign(nat.Attribute, svcObj.Attribute(nat.Name))
			if svcAtt.IsRequired(nat.Name) {
				r.Trailers.Validation.AddRequired(nat.Name)
			}
			// remove metadata attributes from the message attributes
			msgObj.Delete(nat.Name)
		}
		// add any message attributes to response message if not added already
		if len(*msgObj) > 0 {
			if r.Message.Type == Empty {
				r.Message.Type = &Object{}
			}
			resObj := AsObject(r.Message.Type)
			for _, nat := range *msgObj {
				if resObj.Attribute(nat.Name) == nil {
					resObj.Set(nat.Name, nat.Attribute)
				}
				if svcAtt.IsRequired(nat.Name) {
					r.Message.Validation.AddRequired(nat.Name)
				}
			}
		}
		for _, nat := range *AsObject(r.Message.Type) {
			// initialize message attribute from method result
			svcAtt := DupAtt(svcObj.Attribute(nat.Name))
			initAttrFromDesign(nat.Attribute, svcAtt)
			if nat.Attribute.Meta == nil {
				nat.Attribute.Meta = svcAtt.Meta
			} else {
				nat.Attribute.Meta.Merge(svcAtt.Meta)
			}
		}
	} else {
		// method result is not an object type. Initialize response header or
		// trailer metadata if defined or else initialize response message.
		if !r.Headers.IsEmpty() {
			initAttrFromDesign(r.Headers.AttributeExpr, svcAtt)
		} else if !r.Trailers.IsEmpty() {
			initAttrFromDesign(r.Trailers.AttributeExpr, svcAtt)
		} else {
			initAttrFromDesign(r.Message, svcAtt)
		}
	}

	// Set zero value for optional attributes in messages and metadata if not set
	// already
	if IsObject(r.Message.Type) {
		setZero(r.Message)
	}
}

// Dup creates a copy of the response expression.
func (r *GRPCResponseExpr) Dup() *GRPCResponseExpr {
	return &GRPCResponseExpr{
		StatusCode:  r.StatusCode,
		Description: r.Description,
		Parent:      r.Parent,
		Meta:        r.Meta,
		Message:     DupAtt(r.Message),
		Headers:     NewMappedAttributeExpr(r.Headers.Attribute()),
		Trailers:    NewMappedAttributeExpr(r.Trailers.Attribute()),
	}
}
