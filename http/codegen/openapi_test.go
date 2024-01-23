package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openapi "goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestOpenAPI(t *testing.T) {
	cases := map[string]struct {
		DSL     func()
		NilSpec bool
	}{
		"empty": {DSL: testdata.EmptyDSL, NilSpec: true},
		"valid": {DSL: testdata.SimpleDSL, NilSpec: false},
	}
	for k, c := range cases {
		// Reset global variables
		openapi.Definitions = make(map[string]*openapi.Schema)
		root := RunHTTPDSL(t, c.DSL)
		spec, err := OpenAPIFiles(root)
		require.NoError(t, err)
		assert.Equal(t, c.NilSpec, spec == nil, k)
	}
}

func TestOutputPath(t *testing.T) {
	// Reset global variables
	openapi.Definitions = make(map[string]*openapi.Schema)
	root := RunHTTPDSL(t, testdata.SimpleDSL)
	o, err := OpenAPIFiles(root)
	require.NoError(t, err)
	c := 4 // number of files we expect
	require.Len(t, o, c)
	assert.Equal(t, filepath.Join("gen", "http", "openapi.json"), o[0].Path)
	assert.Equal(t, filepath.Join("gen", "http", "openapi.yaml"), o[1].Path)
	assert.Equal(t, filepath.Join("gen", "http", "openapi3.json"), o[2].Path)
	assert.Equal(t, filepath.Join("gen", "http", "openapi3.yaml"), o[3].Path)
}
