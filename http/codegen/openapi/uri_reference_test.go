// This file checks that FormatURIReference reaches every generated OpenAPI
// document as the "uri-reference" format, wherever the attribute is rendered.
package openapi_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

func TestGeneratedURIReferenceFormat(t *testing.T) {
	root := codegen.RunDSL(t, uriReferenceDSL)
	v2, err := openapiv2.NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)
	v2Locations := []string{
		"/paths/~1links/post/parameters/0/format",
		"/paths/~1links/post/parameters/1/items/format",
		"/paths/~1links/post/responses/201/headers/Location/format",
		"/definitions/LinksCreateRequestBody/properties/href/format",
	}
	v3Locations := []string{
		"/paths/~1links/post/parameters/0/schema/format",
		"/paths/~1links/post/parameters/1/schema/items/format",
		"/paths/~1links/post/responses/201/headers/Location/schema/format",
		"/components/schemas/CreateRequestBody/properties/href/format",
	}
	documents := []struct {
		name      string
		value     any
		locations []string
	}{
		{"2.0", v2, v2Locations},
		{"3.0", openapiv3.New(root, openapi.Version30), v3Locations},
		{"3.2", openapiv3.New(root, openapi.Version32), v3Locations},
	}
	formats := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"yaml", yaml.Marshal, yaml.Unmarshal},
	}
	for _, document := range documents {
		for _, format := range formats {
			t.Run(document.name+"/"+format.name, func(t *testing.T) {
				data, err := format.marshal(document.value)
				require.NoError(t, err)
				var decoded map[string]any
				require.NoError(t, format.unmarshal(data, &decoded))
				for _, location := range document.locations {
					require.Equal(t, "uri-reference", valueAt(t, decoded, location), location)
				}
			})
		}
	}
}

func uriReferenceDSL() {
	dsl.API("test", func() {
		dsl.Server("test", func() {
			dsl.Host("localhost", func() {
				dsl.URI("http://localhost:80")
			})
		})
	})
	dsl.Service("links", func() {
		dsl.Method("create", func() {
			dsl.Payload(func() {
				dsl.Attribute("href", dsl.String, func() {
					dsl.Format(dsl.FormatURIReference)
				})
				dsl.Attribute("next", dsl.String, func() {
					dsl.Format(dsl.FormatURIReference)
				})
				dsl.Attribute("refs", dsl.ArrayOf(dsl.String, func() {
					dsl.Format(dsl.FormatURIReference)
				}))
				dsl.Required("href")
			})
			dsl.Result(func() {
				dsl.Attribute("location", dsl.String, func() {
					dsl.Format(dsl.FormatURIReference)
				})
				dsl.Required("location")
			})
			dsl.HTTP(func() {
				dsl.POST("/links")
				dsl.Param("next")
				dsl.Param("refs")
				dsl.Response(dsl.StatusCreated, func() {
					dsl.Header("location:Location")
					dsl.Body(dsl.Empty)
				})
			})
		})
	})
}

// valueAt returns the value at the RFC 6901 JSON pointer in a decoded document.
func valueAt(t *testing.T, document map[string]any, pointer string) any {
	t.Helper()
	var current any = document
	for token := range strings.SplitSeq(strings.TrimPrefix(pointer, "/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch node := current.(type) {
		case map[string]any:
			value, ok := node[token]
			require.True(t, ok, "missing %q in %s", token, pointer)
			current = value
		case []any:
			index, err := strconv.Atoi(token)
			require.NoError(t, err, pointer)
			require.Less(t, index, len(node), pointer)
			current = node[index]
		default:
			require.Failf(t, "unexpected node", "%s: %T at %q", pointer, current, token)
		}
	}
	return current
}
