package genswagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/utils"
)

// Generate is the generator entry point called by the meta generator.
func Generate(api *design.APIDefinition) (_ []string, err error) {
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

	app := kingpin.New("Swagger generator", "Swagger spec generator")
	codegen.RegisterFlags(app)
	_, err = app.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf(`invalid command line: %s. Command line was "%s"`,
			err, strings.Join(os.Args, " "))
	}
	s, err := New(api)
	if err != nil {
		return
	}
	b, err := json.Marshal(s)
	if err != nil {
		return
	}
	swaggerDir := filepath.Join(codegen.OutputDir, "swagger")
	os.RemoveAll(swaggerDir)
	if err = os.MkdirAll(swaggerDir, 0755); err != nil {
		return
	}
	genfiles = append(genfiles, swaggerDir)
	swaggerFile := filepath.Join(swaggerDir, "swagger.json")
	err = ioutil.WriteFile(swaggerFile, b, 0644)
	if err != nil {
		return
	}
	genfiles = append(genfiles, swaggerFile)
	controllerFile := filepath.Join(swaggerDir, "swagger.go")
	genfiles = append(genfiles, controllerFile)
	gg := codegen.NewGoGenerator(controllerFile)
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/julienschmidt/httprouter"),
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	gg.WriteHeader(fmt.Sprintf("%s Swagger Spec", api.Name), "swagger", imports)
	gg.Write([]byte(swagger))
	if err = gg.FormatCode(); err != nil {
		return
	}

	return genfiles, nil
}

const swagger = `
// MountController mounts the swagger spec controller under "/swagger.json".
func MountController(service goa.Service) {
	service.ServeFiles("/swagger.json", "swagger/swagger.json")
}
`
