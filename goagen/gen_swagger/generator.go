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
	tmpl, err := template.New("swagger").Parse(swaggerTmpl)
	if err != nil {
		panic(err.Error()) // bug
	}
	genfiles = append(genfiles, controllerFile)
	gg := codegen.NewGoGenerator(controllerFile)
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/julienschmidt/httprouter"),
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	gg.WriteHeader(fmt.Sprintf("%s Swagger Spec", api.Name), "swagger", imports)
	data := map[string]interface{}{
		"spec": string(b),
	}
	if err = tmpl.Execute(gg, data); err != nil {
		return
	}
	if err = gg.FormatCode(); err != nil {
		return
	}

	return genfiles, nil
}

const swaggerTmpl = `
// MountController mounts the swagger spec controller under "/swagger.json".
func MountController(service goa.Service) {
	ctrl := service.NewController("Swagger")
	service.Info("mount", "ctrl", "Swagger", "action", "Show", "route", "GET /swagger.json")
	h := ctrl.NewHTTPRouterHandle("Show", getSwagger)
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
