// This file verifies that a Swagger response documents every error an endpoint
// maps to its status code instead of only the last declared one.
package openapiv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSameStatusErrorResponses(t *testing.T) {
	root := codegen.RunDSL(t, testdata.SameStatusErrorsDSL)
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)
	responses := spec.Paths["/builds"].(*Path).Post.Responses

	conflict := responses["409"]
	assert.Equal(t,
		"build_open: Another build is open\n\n"+
			"conflict: Request conflicts with the current state\n\n"+
			"quota_exceeded: Build quota reached",
		conflict.Description,
	)
	require.NotNil(t, conflict.Schema)
	assert.Equal(t, "#/definitions/BuildOpen", conflict.Schema.Ref, "first declared error schema")
	assert.Empty(t, conflict.Schema.AnyOf, "Swagger 2.0 schemas cannot use anyOf")
	assert.Contains(t, conflict.Headers, "X-Retry-After")
	require.Contains(t, conflict.Examples, "application/vnd.goa.error")
	assert.Equal(t, "conflict", conflict.Examples["application/vnd.goa.error"].(map[string]any)["name"])

	notFound := responses["404"]
	assert.Equal(t, "not_found: Project not found\n\ngone: Project deleted", notFound.Description)
	require.NotNil(t, notFound.Schema)
	assert.Equal(t, "#/definitions/BuildsStartNotFoundResponseBody", notFound.Schema.Ref, "first declared error schema")
	assert.Empty(t, notFound.Schema.AnyOf, "Swagger 2.0 schemas cannot use anyOf")
	require.Contains(t, notFound.Examples, "application/vnd.goa.error")
	assert.Equal(t, "not_found", notFound.Examples["application/vnd.goa.error"].(map[string]any)["name"], "first declared error example")
}
