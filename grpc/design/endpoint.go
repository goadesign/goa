package design

type (
	// EndpointExpr defines a single endpoint.
	EndpointExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		*eval.DSLFunc
		// Name of endpoint.
		Name string
		// Description of endpoint for consumption by humans.
		Description string
		// Request payload type.
		Request *FieldExpr
		// Response payload type.
		Response *FieldExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
		Metadata map[string]string
		// Protobuf indicates the protobuf file and identifier that define a gRPC rpc.
		// This field is exclusive with Name, Request and Response.
		Protobuf *ProtobufExpr
	}

	// EndpointGroupExpr describes a set of related endpoints.
	EndpointGroupExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		*eval.DSLFunc
		// Name of endpoint group.
		Name string
		// Description of endpoint group for consumption by humans.
		Description string
		// Endpoints grouped together.
		Endpoints []*EndpointExpr
		// Protobuf indicates the protobuf file and identifier that define a gRPC service.
		// This field is exclusive with Name and Endpoints.
		Protobuf *ProtobufExpr
	}
)
