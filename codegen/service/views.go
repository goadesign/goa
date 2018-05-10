package service

import (
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// ViewsFile returns the views file for the given service containing types
// to render result types with more than one view appropriately.
func ViewsFile(genpkg string, service *design.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	if len(svc.ProjectedTypes) == 0 {
		return nil
	}
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "views", "view.go")
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(service.Name+" views", "views",
			[]*codegen.ImportSpec{
				{Path: "goa.design/goa", Name: "goa"},
				{Path: "unicode/utf8"},
			})
		sections = []*codegen.SectionTemplate{header}

		// type definitions
		for _, t := range svc.ProjectedTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "projected-type",
				Source: userTypeT,
				Data:   t.UserTypeData,
			})
		}

		// viewed result type init
		for _, t := range svc.ProjectedTypes {
			if t.Init != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "projected-type",
					Source: typeInitT,
					Data:   t.Init,
				})
			}
		}

		// validations
		for _, t := range svc.ProjectedTypes {
			if t.MustValidate {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "validate-type",
					Source: validateT,
					Data:   t,
				})
			}
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: ProjectedTypeData
const validateT = `{{ printf "Validate runs the validations defined on %s." .VarName | comment }}
func (result {{ .Ref }}) Validate() (err error) {
{{- if .Views }}
	projected := result.Projected
	switch result.View {
	{{- range .Views }}
	{{- if ne .Name "default" }}
	case {{ printf "%q" .Name }}:
	{{- else }}
	default:
	{{- end }}
	{{ .Validate }}
	{{- end }}
	}
{{- else }}
	{{ .Validate }}
{{- end }}
  return
}
`
