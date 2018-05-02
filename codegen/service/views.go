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

		var helpers []*codegen.TransformFunctionData
		for _, t := range svc.ProjectedTypes {
			for _, v := range t.Views {
				helpers = codegen.AppendHelpers(helpers, v.Project.Helpers)
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "as-view",
					Source: initTypeT,
					Data:   v.Project,
				})
			}
		}

		// validations
		for _, t := range svc.ProjectedTypes {
			if t.Validate != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "validate-type",
					Source: validateT,
					Data:   t,
				})
			}
		}

		// transform helpers
		for _, h := range helpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "transform-helpers",
				Source: transformHelperT,
				Data:   h,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

const validateT = `{{ printf "Validate runs the validations defined on %s." .VarName | comment }}
func (result {{ .Ref }}) Validate() (err error) {
	{{ .Validate }}
  return
}
`
