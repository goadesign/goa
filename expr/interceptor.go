package expr

import (
	"goa.design/goa/v3/eval"
)

type (
	// InterceptorExpr describes an interceptor definition in the design.
	// Interceptors are used to inject user code into the request/response processing pipeline.
	// There are four kinds of interceptors, in order of execution:
	//   * client-side payload: executes after the payload is encoded and before the request is sent to the server
	//   * server-side request: executes after the request is decoded and before the payload is sent to the service
	//   * server-side result: executes after the service returns a result and before the response is encoded
	//   * client-side response: executes after the response is decoded and before the result is sent to the client
	InterceptorExpr struct {
		// Name is the name of the interceptor
		Name string
		// Description is the optional description of the interceptor
		Description string
		// ReadPayload lists the payload attribute names read by the interceptor
		ReadPayload *AttributeExpr
		// WritePayload lists the payload attribute names written by the interceptor
		WritePayload *AttributeExpr
		// ReadResult lists the result attribute names read by the interceptor
		ReadResult *AttributeExpr
		// WriteResult lists the result attribute names written by the interceptor
		WriteResult *AttributeExpr
		// ReadStreamingPayload lists the streaming payload attribute names read by the interceptor
		ReadStreamingPayload *AttributeExpr
		// WriteStreamingPayload lists the streaming payload attribute names written by the interceptor
		WriteStreamingPayload *AttributeExpr
		// ReadStreamingResult lists the streaming result attribute names read by the interceptor
		ReadStreamingResult *AttributeExpr
		// WriteStreamingResult lists the streaming result attribute names written by the interceptor
		WriteStreamingResult *AttributeExpr
	}
)

// EvalName returns the generic expression name used in error messages.
func (i *InterceptorExpr) EvalName() string {
	return "interceptor " + i.Name
}

// validate validates the interceptor.
func (i *InterceptorExpr) validate(m *MethodExpr) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	if i.ReadPayload != nil || i.WritePayload != nil {
		payloadObj := AsObject(m.Payload.Type)
		if payloadObj == nil {
			verr.Add(m, "interceptor %q cannot be applied because the method payload is not an object", i.Name)
		}
		if i.ReadPayload != nil {
			i.validateAttributeAccess(m, "read payload", verr, payloadObj, i.ReadPayload)
		}
		if i.WritePayload != nil {
			i.validateAttributeAccess(m, "write payload", verr, payloadObj, i.WritePayload)
		}
	}

	if i.ReadResult != nil || i.WriteResult != nil {
		if m.IsResultStreaming() {
			verr.Add(m, "interceptor %q cannot be applied because the method result is streaming", i.Name)
		}
		resultObj := AsObject(m.Result.Type)
		if resultObj == nil {
			verr.Add(m, "interceptor %q cannot be applied because the method result is not an object", i.Name)
		}
		if i.ReadResult != nil {
			i.validateAttributeAccess(m, "read result", verr, resultObj, i.ReadResult)
		}
		if i.WriteResult != nil {
			i.validateAttributeAccess(m, "write result", verr, resultObj, i.WriteResult)
		}
	}

	if i.ReadStreamingPayload != nil || i.WriteStreamingPayload != nil {
		if !m.IsPayloadStreaming() {
			verr.Add(m, "interceptor %q cannot be applied because the method payload is not streaming", i.Name)
		}
		payloadObj := AsObject(m.StreamingPayload.Type)
		if payloadObj == nil {
			verr.Add(m, "interceptor %q cannot be applied because the method payload is not an object", i.Name)
		}
		if i.ReadStreamingPayload != nil {
			i.validateAttributeAccess(m, "read streaming payload", verr, payloadObj, i.ReadStreamingPayload)
		}
		if i.WriteStreamingPayload != nil {
			i.validateAttributeAccess(m, "write streaming payload", verr, payloadObj, i.WriteStreamingPayload)
		}
	}

	if i.ReadStreamingResult != nil || i.WriteStreamingResult != nil {
		if !m.IsResultStreaming() {
			verr.Add(m, "interceptor %q cannot be applied because the method result is not streaming", i.Name)
		}
		resultObj := AsObject(m.Result.Type)
		if resultObj == nil {
			verr.Add(m, "interceptor %q cannot be applied because the method result is not an object", i.Name)
		}
		if i.ReadStreamingResult != nil {
			i.validateAttributeAccess(m, "read streaming result", verr, resultObj, i.ReadStreamingResult)
		}
		if i.WriteStreamingResult != nil {
			i.validateAttributeAccess(m, "write streaming result", verr, resultObj, i.WriteStreamingResult)
		}
	}

	return verr
}

// validateAttributeAccess validates that all attributes in attr exist in obj
func (i *InterceptorExpr) validateAttributeAccess(m *MethodExpr, source string, verr *eval.ValidationErrors, obj *Object, attr *AttributeExpr) {
	attrObj := AsObject(attr.Type)
	if attrObj == nil {
		verr.Add(m, "interceptor %q %s attribute is not an object", i.Name, source)
		return
	}
	for _, att := range *attrObj {
		if obj.Attribute(att.Name) == nil {
			verr.Add(m, "interceptor %q cannot %s attribute %q: attribute does not exist", i.Name, source, att.Name)
		}
	}
}
