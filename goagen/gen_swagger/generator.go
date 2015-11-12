package genswagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

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
// MountController mounts the Swagger spec controller under "/swagger.json".
func MountController(app *goa.Application) {
	logger := app.Logger.New("ctrl", "Swagger")
	logger.Info("mounting")
	app.Router.GET("/swagger.json", getSwagger)
	logger.Info("handler", "action", "Show", "route", "GET /swagger.json")
	logger.Info("mounted")
}

// getSwagger is the httprouter handle that returns the Swagger spec.
func getSwagger(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/schema+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(200)
	w.Write([]byte(spec))
}

// Generated spec
const spec = ` + "`" + `{{.spec}} ` + "`" + `
`
