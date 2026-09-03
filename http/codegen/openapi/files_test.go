// This file verifies the final JSON and YAML text written for OpenAPI files.
package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestToYAMLQuotesDateShapedStrings(t *testing.T) {
	encoded := toYAML(map[string]any{"example": "3117-95-58"})
	require.Contains(t, encoded, `example: "3117-95-58"`)

	var decoded map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(encoded), &decoded))
	require.Equal(t, "3117-95-58", decoded["example"])
}
