// This file builds one HTTP design that exercises every public name kept for
// existing plugins and provides small assertions shared by compatibility tests.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// releasedHTTPNamesRoot returns a service with ordinary, multipart, streaming,
// empty, and raw-body endpoints plus a file server.
func releasedHTTPNamesRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		child := dsl.Type("Child", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.Pattern("value")
			})
			dsl.Required("value")
		})
		dsl.Service("Names", func() {
			dsl.Method("Complete", func() {
				dsl.Payload(func() {
					dsl.Attribute("child", child)
					dsl.Required("child")
				})
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/complete")
				})
			})
			dsl.Method("Multipart", func() {
				dsl.Payload(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/multipart")
					dsl.MultipartRequest()
				})
			})
			dsl.Method("Watch", func() {
				dsl.StreamingResult(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/watch")
					dsl.ServerSentEvents()
				})
			})
			dsl.Method("Socket", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/socket")
				})
			})
			dsl.Method("Raw", func() {
				dsl.HTTP(func() {
					dsl.POST("/raw")
					dsl.SkipRequestBodyEncodeDecode()
				})
			})
			dsl.Method("Empty", func() {
				dsl.HTTP(func() {
					dsl.GET("/empty")
				})
			})
			dsl.Files("/asset.json", "asset.json")
		})
	})
}

// assertReleasedName checks that one public compatibility string contains the
// final name selected by its declaration.
func assertReleasedName(t *testing.T, name string, declaration *codegen.NameDeclaration) {
	t.Helper()
	if declaration == nil {
		require.Empty(t, name)
		return
	}
	require.Equal(t, declaration.Name(), name)
}

// releasedTypeData returns the first planned HTTP body type accepted by match.
func releasedTypeData(t *testing.T, service *ServiceData, match func(*TypeData) bool) *TypeData {
	t.Helper()
	candidates := append([]*TypeData(nil), service.ServerBodyAttributeTypes...)
	candidates = append(candidates, service.ClientBodyAttributeTypes...)
	for _, endpoint := range service.Endpoints {
		if endpoint.Payload != nil && endpoint.Payload.Request != nil {
			candidates = append(candidates, endpoint.Payload.Request.ServerBody, endpoint.Payload.Request.ClientBody)
		}
		if endpoint.Result != nil {
			for _, response := range endpoint.Result.Responses {
				candidates = append(candidates, response.ServerBody...)
				candidates = append(candidates, response.ClientBody)
			}
		}
	}
	for _, candidate := range candidates {
		if candidate != nil && match(candidate) {
			return candidate
		}
	}
	require.FailNow(t, "planned HTTP body type was not found")
	return nil
}
