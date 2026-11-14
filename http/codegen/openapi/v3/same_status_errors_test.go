// This file verifies that an OpenAPI v3 response documents every error an
// endpoint maps to its status code instead of only the last declared one.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSameStatusErrorResponses(t *testing.T) {
	for _, version := range []openapi.Version{openapi.Version30, openapi.Version32} {
		t.Run(string(version), func(t *testing.T) {
			root := codegen.RunDSL(t, testdata.SameStatusErrorsDSL)
			responses := New(root, version).Paths["/builds"].Post.Responses

			conflict := responses["409"].Value
			require.NotNil(t, conflict.Description)
			assert.Equal(t,
				"build_open: Another build is open\n\n"+
					"conflict: Request conflicts with the current state\n\n"+
					"quota_exceeded: Build quota reached",
				*conflict.Description,
			)
			require.Len(t, conflict.Content, 2)
			typed := conflict.Content["application/json"]
			require.NotNil(t, typed)
			require.Len(t, typed.Schema.AnyOf, 2)
			assert.Equal(t, "#/components/schemas/BuildOpen", typed.Schema.AnyOf[0].Ref)
			assert.Equal(t, "#/components/schemas/QuotaExceeded", typed.Schema.AnyOf[1].Ref)
			assert.Nil(t, typed.Example)
			assert.ElementsMatch(t, []string{"build_open", "quota_exceeded"}, keys(typed.Examples))
			standard := conflict.Content["application/vnd.goa.error"]
			require.NotNil(t, standard)
			assert.Equal(t, "#/components/schemas/Error", standard.Schema.Ref)
			assert.Equal(t, "conflict", standard.Example.(map[string]any)["name"])
			require.Contains(t, conflict.Headers, "X-Retry-After")
			assert.False(t, conflict.Headers["X-Retry-After"].Value.Required, "header set by one of three errors")

			notFound := responses["404"].Value
			require.NotNil(t, notFound.Description)
			assert.Equal(t, "not_found: Project not found\n\ngone: Project deleted", *notFound.Description)
			require.Len(t, notFound.Content, 1)
			shared := notFound.Content["application/vnd.goa.error"]
			require.NotNil(t, shared)
			assert.Equal(t, "#/components/schemas/Error", shared.Schema.Ref, "identical bodies share one schema")
			assert.Empty(t, shared.Schema.AnyOf)
			require.Len(t, shared.Examples, 2)
			assert.Equal(t, "not_found", shared.Examples["not_found"].Value.Value.(map[string]any)["name"])
			assert.Equal(t, "gone", shared.Examples["gone"].Value.Value.(map[string]any)["name"])
		})
	}
}

// TestMergeErrorResponsesHeaderRequired verifies that a merged header is
// required only when every error sets it as required, without changing the
// headers of the individual error responses.
func TestMergeErrorResponsesHeaderRequired(t *testing.T) {
	const (
		required = "required"
		optional = "optional"
		missing  = "missing"
	)
	cases := []struct {
		Name     string
		Headers  []string
		Required bool
	}{
		{"required by every error", []string{required, required}, true},
		{"optional for the first error", []string{optional, required}, false},
		{"optional for the last error", []string{required, optional}, false},
		{"missing from the first error", []string{missing, required}, false},
		{"missing from the last error", []string{required, missing}, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ers := make([]*errorResponse, len(c.Headers))
			for i, h := range c.Headers {
				description := "error"
				response := &Response{Description: &description}
				if h != missing {
					response.Headers = map[string]*HeaderRef{"X-Id": {Value: &Header{Required: h == required}}}
				}
				ers[i] = &errorResponse{name: "error", response: response}
			}
			merged := mergeErrorResponses(ers)
			require.Contains(t, merged.Headers, "X-Id")
			assert.Equal(t, c.Required, merged.Headers["X-Id"].Value.Required)
			for i, h := range c.Headers {
				if h != missing {
					assert.Equal(t, h == required, ers[i].response.Headers["X-Id"].Value.Required, "error %d header changed", i)
				}
			}
		})
	}
}

// keys returns the keys of examples.
func keys(examples map[string]*ExampleRef) []string {
	res := make([]string, 0, len(examples))
	for key := range examples {
		res = append(res, key)
	}
	return res
}
