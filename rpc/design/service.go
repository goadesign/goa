package design

import "github.com/goadesign/goa/eval"

type (
	// ServiceExpr describes a set of related endpoints.
	ServiceExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of endpoint group.
		Name string
		// Description of endpoint group for consumption by humans.
		Description string
		// Endpoints is the list of service endpoints.
		Endpoints []*EndpointExpr
		// Protobuf indicates the protobuf file and identifier that define a gRPC service.
		// This field is exclusive with Name and Endpoints.
		Protobuf *ProtobufExpr
	}
)

// EvalName returns the generic expression name used in error messages.
func (s *ServiceExpr) EvalName() string {
	if s.Name == "" {
		return "unnamed service"
	}
	return "service " + s.Name
}
