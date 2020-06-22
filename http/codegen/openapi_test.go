package codegen

import (
	"path/filepath"
	"testing"

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
		if err != nil {
			t.Fatalf("OpenAPI failed with %s", err)
		}
		if spec == nil && !c.NilSpec {
			t.Errorf("%s: unexpected specs: got nil, expected non-nil", k)
		}
		if spec != nil && c.NilSpec {
			t.Errorf("%s: unexpected specs: got non-nil, expected nil", k)
		}
	}
}

func TestOutputPath(t *testing.T) {
	// Reset global variables
	openapi.Definitions = make(map[string]*openapi.Schema)
	root := RunHTTPDSL(t, testdata.SimpleDSL)
	o, err := OpenAPIFiles(root)
	if err != nil {
		t.Fatalf("OpenAPI failed with %s", err)
	}
	c := 4 // number of files we expect
	if len(o) != c {
		t.Fatalf("unexpected number of OpenAPI files %d instead of %d", len(o), c)
	}
	if o[0].Path != filepath.Join("gen", "http", "openapi.json") {
		t.Errorf("invalid output path %#v", o[0].Path)
	}
	if o[1].Path != filepath.Join("gen", "http", "openapi.yaml") {
		t.Errorf("invalid output path %#v", o[1].Path)
	}
	if o[2].Path != filepath.Join("gen", "http", "openapi_v3.json") {
		t.Errorf("invalid output path %#v", o[2].Path)
	}
	if o[3].Path != filepath.Join("gen", "http", "openapi_v3.yaml") {
		t.Errorf("invalid output path %#v", o[3].Path)
	}
}
