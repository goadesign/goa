package controller

import "github.com/goadesign/goa/design"

type ControllerData struct {
	// Name of controller
	Name string
	// Actions are the controller endpoints.
	Actions []*ActionData
}

type ActionData struct {
	// Name of action
	Name string
	// Request data type. Contains payload, params and headers for HTTP.
	RequestType *design.UserTypeExpr
	// Responses list th successful responses mapped from the return value.
	// In most cases there's only one in which case the action return type is the response
	// type. If there's more than one then the action return type is interface.
	// The endpoints code matches the response type against the possible responses to infer the
	// status code and map headers.
	Responses []*design.UserTypeExpr
}

const tmpl = `func {{ .Name }}(context.Context, request {{ GoTypeRef .RequestType }}) ({{ .ResponseType }}, error) {
	return {{ .Example }}, nil
}
`
