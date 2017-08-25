package service

import (
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// File returns the service file for the given service.
func File(service *design.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, codegen.KebabCase(service.Name), "service.go")
	header := codegen.Header(
		service.Name+" service",
		codegen.Goify(service.Name, false),
		[]*codegen.ImportSpec{
			{Path: "context"},
		})
	svc := Services.Get(service.Name)
	def := &codegen.SectionTemplate{Name: "service", Source: serviceT, Data: svc}

	sections := []*codegen.SectionTemplate{header, def}
	seen := make(map[string]struct{})

	for _, m := range svc.Methods {
		if m.PayloadDef != "" {
			if _, ok := seen[m.Payload]; !ok {
				seen[m.Payload] = struct{}{}
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "service-payload",
					Source: payloadT,
					Data:   m,
				})
			}
		}
		if m.ResultDef != "" {
			if _, ok := seen[m.Result]; !ok {
				seen[m.Result] = struct{}{}
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "result-payload",
					Source: resultT,
					Data:   m,
				})
			}
		}
	}

	for _, ut := range svc.UserTypes {
		if _, ok := seen[ut.Name]; !ok {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "service-user-type",
				Source: userTypeT,
				Data:   ut,
			})
		}
	}

	var newErrorTypes []*UserTypeData
	for _, et := range svc.ErrorTypes {
		if _, ok := seen[et.Name]; !ok {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "error-user-type",
				Source: userTypeT,
				Data:   et,
			})
			newErrorTypes = append(newErrorTypes, et)
		}
	}

	for _, et := range newErrorTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "service-error",
			Source: errorT,
			Data:   et,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// serviceT is the template used to write an service definition.
const serviceT = `
{{ comment .Description }}
type Service interface {
{{- range .Methods }}
	// {{ .Description }}
	{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) {{ if .Result }}({{ .ResultRef }}, error){{ else }}error{{ end }}
{{- end }}
}
`

const payloadT = `{{ comment .PayloadDesc }}
type {{ .Payload }} {{ .PayloadDef }}
`

const resultT = `{{ comment .ResultDesc }}
type {{ .Result }} {{ .ResultDef }}
`

const userTypeT = `{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
`

const errorT = `// Error returns {{ printf "%q" .Name }}.
func (e {{ .Ref }}) Error() string {
	return {{ printf "%q" .Name }}
}
`
