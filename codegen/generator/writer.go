package generator

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/raphael/goa/codegen/code"
)

type (
	// Writer is the generator code writer.
	Writer struct {
		*code.Writer
		// Factory is the factory function used to create the generator.
		Factory string
		// DesignPackage contains the (Go package) path to the user Go design package.
		DesignPackage string
	}
)

// NewWriter creates a new generator writer.
// factory is the name of the factory function for the generator.
// filename if the path to the "main.go" file being created.
// designPackage is the Go package path to the user design package.
func NewWriter(factory, filename, designPackage string) (*Writer, error) {
	w, err := code.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	return &Writer{
		Writer:        w,
		Factory:       factory,
		DesignPackage: designPackage,
	}
}

// WriteMain generates the Go source code for the generator "main" function.
// factory is the name of the function used to instantiate the generator.
func (w *Writer) WriteMain(factory string) error {
	imports = []string{
		w.DesignPackage,
		goaPackagePath(),
	}
	w.WriteHeader("main", imports)
	tmpl, err := template.New(mainTmpl).Parse()
	if err != nil {
		panic(err)
	}
	context := map[string]string{
		"Factory":       w.Factory,
		"DesignPackage": w.DesignPackage,
	}
	return tmpl.Execute(w.Writer, context)
}

// goaPackagePath is an helper function that returns the current goa package.
// This is necessary to support the use of github.com/raphael/goa and
// gopkg.in/raphael/goa.vX.
func (w *Writer) goaPackagePath() string {
	t := reflect.TypeOf(w)
	elems := strings.Split(t.PkgPath(), "/")
	return strings.Join(elems[:3], "/")
}

const mainTmpl = `
func main() {
	gen := {{.Factory}}("{{.DesignPackage}}")
	gen.Main()
}`
