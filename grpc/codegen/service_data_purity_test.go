package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

type (
	// endpointExprSnapshot captures the identity of the design attributes
	// that the analyze pass historically rewrote in place. analyze must
	// treat the design expression tree as read-only: both the attribute
	// pointers and their types must be unchanged after the services data is
	// computed.
	endpointExprSnapshot struct {
		request           *expr.AttributeExpr
		requestType       expr.DataType
		streamingRequest  *expr.AttributeExpr
		streamingReqType  expr.DataType
		responseMessage   *expr.AttributeExpr
		responseType      expr.DataType
		errorMessages     map[string]*expr.AttributeExpr
		errorMessageTypes map[string]expr.DataType
	}
)

func TestAnalyzeLeavesDesignExpressionsUnchanged(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"unary-rpcs", testdata.UnaryRPCsDSL},
		{"unary-rpc-with-errors", testdata.UnaryRPCWithErrorsDSL},
		{"bidirectional-streaming-rpc-with-payload", testdata.BidirectionalStreamingRPCWithPayloadDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			snaps := make(map[string]endpointExprSnapshot)
			for _, svc := range root.API.GRPC.Services {
				for _, e := range svc.GRPCEndpoints {
					snap := endpointExprSnapshot{
						request:           e.Request,
						requestType:       e.Request.Type,
						streamingRequest:  e.StreamingRequest,
						streamingReqType:  e.StreamingRequest.Type,
						responseMessage:   e.Response.Message,
						responseType:      e.Response.Message.Type,
						errorMessages:     make(map[string]*expr.AttributeExpr, len(e.GRPCErrors)),
						errorMessageTypes: make(map[string]expr.DataType, len(e.GRPCErrors)),
					}
					for _, er := range e.GRPCErrors {
						snap.errorMessages[er.Name] = er.Response.Message
						snap.errorMessageTypes[er.Name] = er.Response.Message.Type
					}
					snaps[svc.Name()+"."+e.Name()] = snap
				}
			}

			services := CreateGRPCServices(root)
			for _, svc := range root.API.GRPC.Services {
				require.NotNil(t, services.Get(svc.Name()))
			}

			for _, svc := range root.API.GRPC.Services {
				for _, e := range svc.GRPCEndpoints {
					key := svc.Name() + "." + e.Name()
					snap := snaps[key]
					assert.Same(t, snap.request, e.Request, "%s: request attribute replaced", key)
					assert.True(t, snap.requestType == e.Request.Type, "%s: request type replaced", key)
					assert.Same(t, snap.streamingRequest, e.StreamingRequest, "%s: streaming request attribute replaced", key)
					assert.True(t, snap.streamingReqType == e.StreamingRequest.Type, "%s: streaming request type replaced", key)
					assert.Same(t, snap.responseMessage, e.Response.Message, "%s: response message attribute replaced", key)
					assert.True(t, snap.responseType == e.Response.Message.Type, "%s: response message type replaced", key)
					for _, er := range e.GRPCErrors {
						assert.Same(t, snap.errorMessages[er.Name], er.Response.Message, "%s: error %q response message attribute replaced", key, er.Name)
						assert.True(t, snap.errorMessageTypes[er.Name] == er.Response.Message.Type, "%s: error %q response message type replaced", key, er.Name)
					}
				}
			}
		})
	}
}
