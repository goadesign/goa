package codegen

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

type (
	// httpEndpointExprSnapshot captures the identity of the design
	// attributes that the analyze pass historically rewrote in place.
	// analyze must treat the design expression tree as read-only: the
	// attribute pointers, their types and their marshal tag meta must be
	// unchanged after the services data is computed.
	httpEndpointExprSnapshot struct {
		body              *expr.AttributeExpr
		bodyType          expr.DataType
		streamingBody     *expr.AttributeExpr
		streamingBodyType expr.DataType
		responseBodies    []*expr.AttributeExpr
		responseBodyTypes []expr.DataType
		errorBodies       map[string]*expr.AttributeExpr
		errorBodyTypes    map[string]expr.DataType
	}
)

func TestAnalyzeLeavesDesignExpressionsUnchanged(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"payload-body-user", testdata.PayloadBodyUserDSL},
		{"aliased-payload-and-result", testdata.AliasTypeDSL},
		{"result-body-multiple-views", testdata.ResultBodyMultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"primitive-error", testdata.PrimitiveErrorResponseDSL},
		{"service-error", testdata.ServiceErrorResponseDSL},
		{"websocket-streaming-payload", testdata.StreamingPayloadDSL},
		{"websocket-bidirectional", testdata.BidirectionalStreamingDSL},
		{"sse", testdata.SSEObjectDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			snaps := make(map[string]httpEndpointExprSnapshot)
			tags := make(map[*expr.AttributeExpr][]string)
			for _, svc := range root.API.HTTP.Services {
				for _, e := range svc.HTTPEndpoints {
					snap := httpEndpointExprSnapshot{
						body:           e.Body,
						bodyType:       e.Body.Type,
						errorBodies:    make(map[string]*expr.AttributeExpr, len(e.HTTPErrors)),
						errorBodyTypes: make(map[string]expr.DataType, len(e.HTTPErrors)),
					}
					if e.StreamingBody != nil {
						snap.streamingBody = e.StreamingBody
						snap.streamingBodyType = e.StreamingBody.Type
					}
					for _, resp := range e.Responses {
						snap.responseBodies = append(snap.responseBodies, resp.Body)
						snap.responseBodyTypes = append(snap.responseBodyTypes, resp.Body.Type)
					}
					for _, er := range e.HTTPErrors {
						snap.errorBodies[er.Name] = er.Response.Body
						snap.errorBodyTypes[er.Name] = er.Response.Body.Type
					}
					snaps[svc.Name()+"."+e.Name()] = snap
					collectMarshalTagMeta(e.Body, tags, make(map[string]struct{}))
					collectMarshalTagMeta(e.StreamingBody, tags, make(map[string]struct{}))
					for _, resp := range e.Responses {
						collectMarshalTagMeta(resp.Body, tags, make(map[string]struct{}))
					}
					for _, er := range e.HTTPErrors {
						collectMarshalTagMeta(er.Response.Body, tags, make(map[string]struct{}))
					}
				}
			}

			plan := linkedHTTPPlanForRoot(t, root)
			for _, svc := range root.API.HTTP.Services {
				require.NotNil(t, plan.services.Get(svc.Name()))
			}

			for _, svc := range root.API.HTTP.Services {
				for _, e := range svc.HTTPEndpoints {
					key := svc.Name() + "." + e.Name()
					snap := snaps[key]
					assert.Same(t, snap.body, e.Body, "%s: body attribute replaced", key)
					assert.True(t, snap.bodyType == e.Body.Type, "%s: body type replaced", key)
					if snap.streamingBody != nil {
						assert.Same(t, snap.streamingBody, e.StreamingBody, "%s: streaming body attribute replaced", key)
						assert.True(t, snap.streamingBodyType == e.StreamingBody.Type, "%s: streaming body type replaced", key)
					}
					require.Len(t, e.Responses, len(snap.responseBodies), "%s: responses changed", key)
					for i, resp := range e.Responses {
						assert.Same(t, snap.responseBodies[i], resp.Body, "%s: response #%d body attribute replaced", key, i)
						assert.True(t, snap.responseBodyTypes[i] == resp.Body.Type, "%s: response #%d body type replaced", key, i)
					}
					for _, er := range e.HTTPErrors {
						assert.Same(t, snap.errorBodies[er.Name], er.Response.Body, "%s: error %q response body attribute replaced", key, er.Name)
						assert.True(t, snap.errorBodyTypes[er.Name] == er.Response.Body.Type, "%s: error %q response body type replaced", key, er.Name)
					}
				}
			}

			// addMarshalTags must only annotate the detached shaped bodies,
			// never the design attributes.
			after := make(map[*expr.AttributeExpr][]string)
			for _, svc := range root.API.HTTP.Services {
				for _, e := range svc.HTTPEndpoints {
					collectMarshalTagMeta(e.Body, after, make(map[string]struct{}))
					collectMarshalTagMeta(e.StreamingBody, after, make(map[string]struct{}))
					for _, resp := range e.Responses {
						collectMarshalTagMeta(resp.Body, after, make(map[string]struct{}))
					}
					for _, er := range e.HTTPErrors {
						collectMarshalTagMeta(er.Response.Body, after, make(map[string]struct{}))
					}
				}
			}
			assert.Equal(t, tags, after, "struct:tag meta written into design attributes")
		})
	}
}

// collectMarshalTagMeta records the struct:tag:* meta keys carried by the
// attributes reachable from att so the test can assert that analysis does not
// add marshal tags to the design expression tree.
func collectMarshalTagMeta(att *expr.AttributeExpr, tags map[*expr.AttributeExpr][]string, seen map[string]struct{}) {
	if att == nil {
		return
	}
	var keys []string
	for k := range att.Meta {
		if strings.HasPrefix(k, "struct:tag:") {
			keys = append(keys, k)
		}
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		tags[att] = keys
	}
	switch dt := att.Type.(type) {
	case expr.UserType:
		if _, ok := seen[dt.ID()]; ok {
			return
		}
		seen[dt.ID()] = struct{}{}
		collectMarshalTagMeta(dt.Attribute(), tags, seen)
	case *expr.Object:
		for _, nat := range *dt {
			collectMarshalTagMeta(nat.Attribute, tags, seen)
		}
	case *expr.Array:
		collectMarshalTagMeta(dt.ElemType, tags, seen)
	case *expr.Map:
		collectMarshalTagMeta(dt.KeyType, tags, seen)
		collectMarshalTagMeta(dt.ElemType, tags, seen)
	case *expr.Union:
		for _, nat := range dt.Values {
			collectMarshalTagMeta(nat.Attribute, tags, seen)
		}
	}
}
