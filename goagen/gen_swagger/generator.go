package genswagger

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

// Generator is the swagger code generator.
type Generator struct {
	outDir string // Path to output directory
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var outDir string
	set := flag.NewFlagSet("app", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.String("design", "", "")
	set.Parse(os.Args[2:])

	g := &Generator{outDir: outDir}

	return g.Generate(design.Design)
}

// Generate produces the skeleton main.
func (g *Generator) Generate(api *design.APIDefinition) (_ []string, err error) {
	var genfiles []string

	cleanup := func() {
		for _, f := range genfiles {
			os.Remove(f)
		}
	}

	go utils.Catch(nil, cleanup)

	defer func() {
		if err != nil {
			cleanup()
		}
	}()

	swaggerDir := filepath.Join(g.outDir, "swagger")
	os.RemoveAll(swaggerDir)
	if err = os.MkdirAll(swaggerDir, 0755); err != nil {
		return nil, err
	}
	genfiles = append(genfiles, swaggerDir)
	s, err := New(api)
	if err != nil {
		return nil, err
	}

	// JSON
	rawJSON, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	swaggerFile := filepath.Join(swaggerDir, "swagger.json")
	if err := ioutil.WriteFile(swaggerFile, rawJSON, 0644); err != nil {
		return nil, err
	}
	genfiles = append(genfiles, swaggerFile)

	// YAML
	var yamlSource interface{}
	if err = json.Unmarshal(rawJSON, &yamlSource); err != nil {
		return nil, err
	}

	rawYAML, err := yaml.Marshal(yamlSource)
	if err != nil {
		return nil, err
	}
	swaggerFile = filepath.Join(swaggerDir, "swagger.yaml")
	if err := ioutil.WriteFile(swaggerFile, rawYAML, 0644); err != nil {
		return nil, err
	}
	genfiles = append(genfiles, swaggerFile)

	// Go endpoint
	controllerFile := filepath.Join(swaggerDir, "swagger.go")
	genfiles = append(genfiles, controllerFile)
	file, err := codegen.SourceFileFor(controllerFile)
	if err != nil {
		return nil, err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	file.WriteHeader(fmt.Sprintf("%s Swagger Spec", api.Name), "swagger", imports)
	err = file.ExecuteTemplate("swagger", swaggerT, nil, api)
	if err != nil {
		return nil, err
	}
	if err = file.FormatCode(); err != nil {
		return nil, err
	}

	return genfiles, nil
}

const swaggerT = `
// MountController mounts the swagger spec controller.
func MountController(service *goa.Service) {
	service.ServeFiles("/swagger.json", "swagger/swagger.json")
}
`
