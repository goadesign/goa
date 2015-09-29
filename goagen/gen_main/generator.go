package genmain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"bitbucket.org/pkg/inflect"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Generator is the application code generator.
type Generator int

// Generate is the generator entry point called by the meta generator.
func Generate(api *design.APIDefinition) ([]string, error) {
	g, err := NewGenerator()
	if err != nil {
		return nil, err
	}
	return g.Generate(api)
}

// NewGenerator returns the application code generator.
func NewGenerator() (*Generator, error) {
	app := kingpin.New("Main generator", "application main generator")
	goagen.RegisterFlags(app)
	NewCommand().RegisterFlags(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf(`invalid command line: %s. Command line was "%s"`,
			err, strings.Join(os.Args, " "))
	}
	return new(Generator), nil
}

// Generate produces the skeleton main.
func (g *Generator) Generate(api *design.APIDefinition) ([]string, error) {
	var genfiles []string
	mainFile := filepath.Join(goagen.OutputDir, "main.go")
	_, err := os.Stat(mainFile)
	if err != nil || goagen.Force {
		tmpl, err := template.New("main").Funcs(template.FuncMap{"tempvar": tempvar}).Parse(mainTmpl)
		if err != nil {
			panic(err.Error()) // bug
		}
		g := goagen.NewGoGenerator(mainFile)
		title := fmt.Sprintf("%s: Application Main", api.Name)
		imports := []*goagen.ImportSpec{
			goagen.SimpleImport("github.com/raphael/goa"),
			goagen.NewImport("log", "gopkg.in/inconshreveable/log15.v2"),
		}
		g.WriteHeader(title, "main", imports)
		data := map[string]interface{}{
			"Name":      AppName,
			"Resources": api.Resources,
		}
		err = tmpl.Execute(g, data)
		if err != nil {
			return nil, err
		}
		if err := g.FormatCode(); err != nil {
			return nil, err
		}
		genfiles = []string{"main.go"}
	}
	tmpl, err := template.New("ctrl").Parse(ctrlTmpl)
	if err != nil {
		panic(err.Error()) // bug
	}
	imports := []*goagen.ImportSpec{goagen.SimpleImport("github.com/raphael/goa")}
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		filename := filepath.Join(goagen.OutputDir, r.FormatName(true, false)) + ".go"
		_, err := os.Stat(filename)
		if err != nil || goagen.Force {
			resGen := goagen.NewGoGenerator(filename)
			resGen.WriteHeader(fmt.Sprintf("Resource %s Controller", r.Name), "main", imports)
			err := r.IterateActions(func(a *design.ActionDefinition) error {
				name := inflect.Camelize(a.Name) + inflect.Camelize(a.Parent.Name)
				resp, ok := a.Responses["OK"]
				var mt *design.MediaTypeDefinition
				if ok {
					mt = api.MediaTypes[resp.MediaType]
				}
				data := map[string]interface{}{
					"Description": a.Description,
					"Name":        name,
					"OKResp":      resp,
					"MediaType":   mt,
				}
				err := tmpl.Execute(resGen, data)
				if err != nil {
					panic(err.Error())
				}
				return err
			})
			if err != nil {
				return err
			}
			if err := resGen.FormatCode(); err != nil {
				return err
			}
			genfiles = append(genfiles, filename)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return genfiles, nil
}

// tempCount is the counter used to create unique temporary variable names.
var tempCount int

// tempvar generates a unique temp var name.
func tempvar() string {
	tempCount++
	return fmt.Sprintf("c%d", tempCount)
}

const mainTmpl = `
func main() {
	// Setup logger
	goa.Log.SetHandler(log.StdoutHandler)

	// Create application
	app := goa.New("{{.Name}}")

{{range $name, $res := .Resources}}	// Create "{{$res.FormatName true true}}" resource controller
	{{$tmp := tempvar}}{{$tmp}} := goa.NewController("{{$res.FormatName true true}}")

	// Register the resource action handlers
	{{$tmp}}.SetHandlers(goa.Handlers{
{{range .Actions}}		"{{.FormatName true}}":   {{.FormatName false}}{{.Parent.FormatName false true}},
{{end}}	})

	// Mount controller onto application
	app.Mount({{$tmp}})
{{end}}
	// Run application, listen on port 8080
	app.Run(":8080")
}
`
const ctrlTmpl = `// {{.Description}}
func {{.Name}}(c *app.{{.Name}}Context) error {
	return {{if .OKResp}}c.{{.OKResp.FormatName false}}({{if .MediaType}}&{{.MediaType.TypeName}}{}{{end}}){{else}}nil{{end}}
}
`
