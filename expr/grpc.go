package expr

type (
	// GRPCExpr contains the API level gRPC specific expressions.
	GRPCExpr struct {
		// Services contains the gRPC services created by the DSL.
		Services []*GRPCServiceExpr
		// Errors lists the error gRPC error responses defined globally.
		Errors []*GRPCErrorExpr
	}
)

// Service returns the service with the given name if any.
func (g *GRPCExpr) Service(name string) *GRPCServiceExpr {
	for _, res := range g.Services {
		if res.Name() == name {
			return res
		}
	}
	return nil
}

// ServiceFor creates a new or returns the existing service definition for
// the given service.
func (g *GRPCExpr) ServiceFor(s *ServiceExpr) *GRPCServiceExpr {
	if res := g.Service(s.Name); res != nil {
		return res
	}
	res := &GRPCServiceExpr{
		ServiceExpr: s,
	}
	g.Services = append(g.Services, res)
	return res
}

// EvalName returns the name printed in case of evaluation error.
func (g *GRPCExpr) EvalName() string {
	return "API GRPC"
}
