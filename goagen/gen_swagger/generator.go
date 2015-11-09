package genswagger

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
)

// Generate is the generator entry point called by the meta generator.
func Generate(api *design.APIDefinition) ([]string, error) {
	s, err := New(api)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	swaggerFile := filepath.Join(codegen.OutputDir, "swagger.json")
	err = ioutil.WriteFile(swaggerFile, b, 0644)
	if err != nil {
		return nil, err
	}
	return []string{swaggerFile}, nil
}
