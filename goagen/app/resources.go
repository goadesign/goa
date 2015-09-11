package app

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/code"
)

// ResourcesWriter generate code for a goa application resources.
// Resources are data structures initialized by the application handlers and passed to controller
// actions.
type ResourcesWriter struct {
	*code.Writer
	ResourceTmpl *template.Template
}

// NewResourcesWriter returns a contexts code writer.
// Resources provide the glue between the underlying request data and the user controller.
func NewResourcesWriter(filename string) (*ResourcesWriter, error) {
	cw, err := code.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	funcMap := cw.FuncMap
	funcMap["join"] = strings.Join
	funcMap["gotypedef"] = code.GoTypeDef
	resourceTmpl, err := template.New("resource").Funcs(cw.FuncMap).Parse(resourceT)
	if err != nil {
		return nil, err
	}
	w := ResourcesWriter{
		Writer:       cw,
		ResourceTmpl: resourceTmpl,
	}
	return &w, nil
}

// ResourceTemplateData contains the information required to generate the resource code.
type ResourceTemplateData struct {
	Name              string                     // Name of resource
	Identifier        string                     // Identifier of resource media type
	Description       string                     // Description of resource
	Type              *design.UserTypeDefinition // Type of resource media type
	CanonicalTemplate string                     // CanonicalFormat represents the resource canonical path in the form of a fmt.Sprintf format.
	CanonicalParams   []string                   // CanonicalParams is the list of parameter names that appear in the resource canonical path in order.
}

// Write writes the code for the context types to the writer.
func (w *ResourcesWriter) Write(data *ResourceTemplateData) error {
	if data.Type == nil {
		return fmt.Errorf("missing resource type definition for %s", data.Name)
	}
	return w.ResourceTmpl.Execute(w.Writer, data)
}

const (
	resourceT = `// {{.Description}}
// Media type: {{.Identifier}}
type {{.Name}} {{gotypedef .Type 0 true false}}
{{if .CanonicalTemplate}}
// {{.Name}}Href returns the resource href.
func {{.Name}}Href({{join .CanonicalParams ", "}} string) string {
	return fmt.Sprintf("{{.CanonicalTemplate}}", {{join .CanonicalParams ", "}})
}
{{end}}`
)
