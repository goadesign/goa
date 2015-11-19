package genswagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
)

// Generate is the generator entry point called by the meta generator.
func Generate(api *design.APIDefinition) ([]string, error) {
	app := kingpin.New("Swagger generator", "Swagger spec generator")
	codegen.RegisterFlags(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf(`invalid command line: %s. Command line was "%s"`,
			err, strings.Join(os.Args, " "))
	}
	s, err := New(api)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	swaggerDir := filepath.Join(codegen.OutputDir, "swagger")
	os.RemoveAll(swaggerDir)
	if err = os.MkdirAll(swaggerDir, 0755); err != nil {
		return nil, err
	}
	swaggerFile := filepath.Join(swaggerDir, "swagger.json")
	err = ioutil.WriteFile(swaggerFile, b, 0644)
	if err != nil {
		return nil, err
	}
	controllerFile := filepath.Join(swaggerDir, "swagger.go")
	tmpl, err := template.New("swagger").Parse(swaggerTmpl)
	if err != nil {
		panic(err.Error()) // bug
	}
	gg := codegen.NewGoGenerator(controllerFile)
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	gg.WriteHeader(fmt.Sprintf("%s Swagger Spec", api.Name), "swagger", imports)
	data := map[string]interface{}{
		"spec": string(b),
	}
	err = tmpl.Execute(gg, data)
	if err != nil {
		return nil, err
	}
	if err := gg.FormatCode(); err != nil {
		return nil, err
	}
	return []string{controllerFile, swaggerFile}, nil
}

const swaggerTmpl = `
// MountController mounts the swagger spec controller under "/swagger.json".
func MountController(service goa.Service) {
	service.Info("mount", "ctrl", "Swagger", "action", "Show", "route", "GET /swagger.json")
	h := goa.NewHTTPRouterHandle(service, "Swagger", "Show", getSwagger)
	service.HTTPHandler().(*httprouter.Router).Handle("GET", "/swagger.json", h)
}

// getSwagger is the httprouter handle that returns the Swagger spec.
// func getSwagger(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
func getSwagger(ctx *goa.Context) error {
	ctx.Header().Set("Content-Type", "application/swagger+json")
	ctx.Header().Set("Cache-Control", "public, max-age=3600")
	return ctx.Respond(200, []byte(spec))
}

// Generated spec
const spec = ` + "`" + `{{.spec}} ` + "`" + `
`
