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
	cases := []struct {
		Name  string
		DSL   func()
		Paths []string
		Err   string
	}{{
		Name: "default",
		DSL:  testdata.SimpleDSL,
		Paths: []string{
			filepath.Join("gen", "http", "openapi.json"),
			filepath.Join("gen", "http", "openapi.yaml"),
			filepath.Join("gen", "http", "openapi3.json"),
			filepath.Join("gen", "http", "openapi3.yaml"),
			filepath.Join("gen", "http", "openapi3.2.json"),
			filepath.Join("gen", "http", "openapi3.2.yaml"),
		},
	}, {
		Name: "versions subset",
		DSL:  testdata.OpenAPIVersionsSubsetDSL,
		Paths: []string{
			filepath.Join("gen", "http", "openapi3.2.json"),
			filepath.Join("gen", "http", "openapi3.2.yaml"),
		},
	}, {
		Name: "path override",
		DSL:  testdata.OpenAPIPathOverrideDSL,
		Paths: []string{
			filepath.Join("gen", "http", "openapi.json"),
			filepath.Join("gen", "http", "openapi.yaml"),
			filepath.Join("gen", "http", "openapi3.json"),
			filepath.Join("gen", "http", "openapi3.yaml"),
			filepath.Join("gen", "docs", "openapi.json"),
			filepath.Join("gen", "docs", "openapi.yaml"),
		},
	}, {
		Name: "invalid version",
		DSL:  testdata.OpenAPIInvalidVersionDSL,
		Err:  `invalid value "4.0" for meta "openapi:versions": valid values are "2.0", "3.0" and "3.2"`,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := RunHTTPDSL(t, c.DSL)
			o, err := OpenAPIFiles(root)
			if c.Err != "" {
				require.EqualError(t, err, c.Err)
				return
			}
			require.NoError(t, err)
			require.Len(t, o, len(c.Paths))
			for i, p := range c.Paths {
				assert.Equal(t, p, o[i].Path)
			}
		})
	}
}
