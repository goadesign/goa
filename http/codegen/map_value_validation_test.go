// These tests render complete HTTP clients and servers whose bodies hold
// collections of objects with only required fields. The wire fields of such
// objects are pointers, so each element or value needs the object's nested
// validator even though the service type needs no check.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestHTTPCollectionOfRequiredObjectsValidation checks that request and
// response bodies call the nested validator for each array element and map
// value, and that the generated packages compile.
func TestHTTPCollectionOfRequiredObjectsValidation(t *testing.T) {
	tests := []struct {
		name string
		kind string
		body func(expr.DataType) expr.DataType
		file func(*Plan) []*codegen.File
		path string
		call string
	}{
		{
			name: "selected_response_array",
			kind: "selected_response",
			body: func(item expr.DataType) expr.DataType { return dsl.ArrayOf(item) },
			file: (*Plan).ClientFiles,
			path: "/client/encode_decode.go",
			call: `validateItemResponse(e, "body[*]")`,
		},
		{
			name: "selected_response_map",
			kind: "selected_response",
			body: func(item expr.DataType) expr.DataType { return dsl.MapOf(dsl.String, item) },
			file: (*Plan).ClientFiles,
			path: "/client/encode_decode.go",
			call: `validateItemResponse(v, "body[key]")`,
		},
		{
			name: "response_field_map",
			kind: "response",
			body: func(item expr.DataType) expr.DataType { return dsl.MapOf(dsl.String, item) },
			file: (*Plan).ClientTypeFiles,
			path: "/client/types.go",
			call: `validateItemResponseBody(v, "body.items[key]")`,
		},
		{
			name: "selected_request_map",
			kind: "selected_request",
			body: func(item expr.DataType) expr.DataType { return dsl.MapOf(dsl.String, item) },
			file: (*Plan).ServerFiles,
			path: "/server/encode_decode.go",
			call: `validateItemRequestBody(v, "body[key]")`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := requiredObjectCollectionRoot(t, test.kind, test.body)
			plan := linkedHTTPPlanForRoot(t, root)
			require.Contains(t, renderedHTTPFile(t, plannedHTTPFile(t, test.file(plan), test.path)), test.call)

			serviceFiles, err := service.Files(plan.servicePlan)
			require.NoError(t, err)
			files := slices.Clone(serviceFiles)
			files = append(files, plan.ClientFiles()...)
			files = append(files, plan.ClientTypeFiles()...)
			files = append(files, plan.ServerFiles()...)
			files = append(files, plan.ServerTypeFiles()...)
			files = append(files, plan.PathFiles()...)
			runGeneratedMixedSSECompile(t, files, "./gen/...")
		})
	}
}

// requiredObjectCollectionRoot places a collection of Item in a request or
// response, either as the whole body or as a field of the body.
func requiredObjectCollectionRoot(t *testing.T, kind string, body func(expr.DataType) expr.DataType) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		item := dsl.Type("Item", func() {
			dsl.Attribute("id", dsl.Int)
			dsl.Required("id")
		})
		fields := func() {
			dsl.Attribute("items", body(item))
			dsl.Required("items")
		}
		dsl.Service("Collections", func() {
			dsl.Method("Sync", func() {
				switch kind {
				case "selected_response", "response":
					dsl.Result(fields)
				case "selected_request":
					dsl.Payload(fields)
				default:
					t.Fatalf("unknown collection fixture %q", kind)
				}
				dsl.HTTP(func() {
					switch kind {
					case "selected_response":
						dsl.GET("/")
						dsl.Response(dsl.StatusOK, func() {
							dsl.Body("items")
						})
					case "response":
						dsl.GET("/")
					case "selected_request":
						dsl.POST("/")
						dsl.Body("items")
					}
				})
			})
		})
	})
}
