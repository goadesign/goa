package genswagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
)

// Generator is the swagger code generator.
type Generator struct{}

// Generate is the generator entry point called by the meta generator.
func Generate(roots []interface{}) (files []string, err error) {
	api := roots[0].(*design.APIDefinition)
	g := new(Generator)
	root := &cobra.Command{
		Use:   "goagen",
		Short: "Swagger generator",
		Long:  "Swagger generator",
		Run:   func(*cobra.Command, []string) { files, err = g.Generate(api) },
	}
	codegen.RegisterFlags(root)
	NewCommand().RegisterFlags(root)
	root.Execute()
	return
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

	swaggerDir := filepath.Join(codegen.OutputDir, "swagger")
	os.RemoveAll(swaggerDir)
	if err = os.MkdirAll(swaggerDir, 0755); err != nil {
		return nil, err
	}
	genfiles = append(genfiles, swaggerDir)

	err = api.IterateVersions(func(ver *design.APIVersionDefinition) error {
		s, err := New(api, ver)
		if err != nil {
			return err
		}
		b, err := json.Marshal(s)
		if err != nil {
			return err
		}
		verDir := swaggerDir
		if ver.Version != "" {
			verDir = filepath.Join(swaggerDir, ver.Version)
			if err = os.MkdirAll(verDir, 0755); err != nil {
				return err
			}
			genfiles = append(genfiles, verDir)
		}
		swaggerFile := filepath.Join(verDir, "swagger.json")
		err = ioutil.WriteFile(swaggerFile, b, 0644)
		if err != nil {
			return err
		}
		genfiles = append(genfiles, swaggerFile)
		return nil
	})
	if err != nil {
		return nil, err
	}
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
	var versions []*design.APIVersionDefinition
	api.IterateVersions(func(ver *design.APIVersionDefinition) error {
		versions = append(versions, ver)
		return nil
	})
	err = file.ExecuteTemplate("swagger", swaggerT, nil, map[string]interface{}{"Versions": versions})
	if err != nil {
		return nil, err
	}
	if err = file.FormatCode(); err != nil {
		return nil, err
	}

	return genfiles, nil
}

const swaggerT = `
// MountController mounts the swagger spec controllers (one per API version) under "/swagger.json".
func MountController(service *goa.Service) {
{{range .Versions}}	service{{if .Version}}.Version("{{.Version}}"){{end}}{{/*
*/}}.ServeFiles("/swagger.json", "swagger/{{if .Version}}{{.Version}}/{{end}}swagger.json")
{{end}}
}
`
