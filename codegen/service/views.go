package service

import (
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

type viewedType struct {
	// Name is the type name.
	Name string
	// Views is the view data for all views defined in the type.
	Views []*ViewData
}

// ViewsFile returns the views file for the given service which contains
// logic to render result types using the defined views.
func ViewsFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	if len(svc.projectedTypes) == 0 {
		return nil
	}
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(svc.VarName), "views", "view.go")
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
		for _, t := range svc.viewedResultTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "viewed-result-type",
				Source: userTypeT,
				Data:   t.UserTypeData,
			})
		}
		for _, t := range svc.projectedTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "projected-type",
				Source: userTypeT,
				Data:   t.UserTypeData,
			})
		}

		// generate a map for result types with view name as key and the fields
		// rendered in the view as value.
		var (
			rtdata []*viewedType
			seen   = make(map[string]struct{})
		)
		{
			for _, t := range svc.viewedResultTypes {
				name := t.Views[0].TypeVarName
				if _, ok := seen[name]; !ok {
					rtdata = append(rtdata, &viewedType{Name: name, Views: t.Views})
					seen[name] = struct{}{}
				}
			}
			for _, t := range svc.projectedTypes {
				if len(t.Views) == 0 {
					continue
				}
				name := t.Views[0].TypeVarName
				if _, ok := seen[name]; !ok {
					rtdata = append(rtdata, &viewedType{Name: name, Views: t.Views})
					seen[name] = struct{}{}
				}
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "viewed-type-map",
			Source: viewedMapT,
			Data: map[string]interface{}{
				"ViewedTypes": rtdata,
			},
		})

		// validations
		for _, t := range svc.viewedResultTypes {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "validate-viewed-result-type",
				Source: validateT,
				Data:   t.Validate,
			})
		}
		for _, t := range svc.projectedTypes {
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

// input: map[string]interface{}{"ViewedTypes": []*viewedType}
const viewedMapT = `var (
{{- range .ViewedTypes }}
	{{ printf "%sMap is a map of attribute names in result type %s indexed by view name." .Name .Name | comment }}
	{{ .Name }}Map = map[string][]string{
	{{- range .Views }}
		"{{ .Name }}": []string{
			{{- range $n := .Attributes }}
				"{{ $n }}",
			{{- end }}
		},
	{{- end }}
	}
{{- end }}
)
`
