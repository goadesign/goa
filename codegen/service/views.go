package service

import (
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// ViewsFile returns the views file for the given service to render result
// types (if any) using the defined views.
func ViewsFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
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
		for _, t := range svc.ViewedResultTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "viewed-result-type",
				Source: userTypeT,
				Data:   t.UserTypeData,
			})
		}
		for _, t := range svc.ProjectedTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "projected-type",
				Source: userTypeT,
				Data:   t.UserTypeData,
			})
		}

		// validations
		for _, t := range svc.ViewedResultTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "validate-viewed-result-type",
				Source: validateT,
				Data:   t.Validate,
			})
		}
		for _, t := range svc.ProjectedTypes {
			for _, v := range t.Validations {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "validate-projected-type",
					Source: validateT,
					Data:   v,
				})
			}
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: ValidateData
const validateT = `{{ comment .Description }}
func {{ .Name }}(result {{ .Ref }}) (err error) {
	{{ .Validate }}
  return
}
`
