package service

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

var (
	fns          = template.FuncMap{"comment": codegen.Comment}
	serviceTmpl  = template.Must(template.New("service").Funcs(fns).Parse(serviceT))
	payloadTmpl  = template.Must(template.New("payload").Funcs(fns).Parse(payloadT))
	resultTmpl   = template.Must(template.New("result").Funcs(fns).Parse(resultT))
	userTypeTmpl = template.Must(template.New("userType").Funcs(fns).Parse(userTypeT))
	errorTmpl    = template.Must(template.New("error").Funcs(fns).Parse(errorT))
)

// Service returns the service file for the given service.
func Service(service *design.ServiceExpr) codegen.File {
	path := filepath.Join(codegen.KebabCase(service.Name), "service.go")
	sections := func(genPkg string) []*codegen.Section {

		header := codegen.Header(service.Name+" service", codegen.Goify(service.Name, false),
			[]*codegen.ImportSpec{
				{Path: "context"},
			})

		svc := Services.Get(service.Name)
		def := &codegen.Section{
			Template: serviceTmpl,
			Data:     svc,
		}

		secs := []*codegen.Section{header, def}
		seen := make(map[string]struct{})

		for _, m := range svc.Methods {
			if m.PayloadDef != "" {
				if _, ok := seen[m.Payload]; !ok {
					seen[m.Payload] = struct{}{}
					secs = append(secs, &codegen.Section{
						Template: payloadTmpl,
						Data:     m,
					})
				}
			}
			if m.ResultDef != "" {
				if _, ok := seen[m.Result]; !ok {
					seen[m.Result] = struct{}{}
					secs = append(secs, &codegen.Section{
						Template: resultTmpl,
						Data:     m,
					})
				}
			}
		}

		for _, ut := range svc.UserTypes {
			if _, ok := seen[ut.Name]; !ok {
				secs = append(secs, &codegen.Section{
					Template: userTypeTmpl,
					Data:     ut,
				})
			}
		}

		var newErrorTypes []*UserTypeData
		for _, et := range svc.ErrorTypes {
			if _, ok := seen[et.Name]; !ok {
				secs = append(secs, &codegen.Section{
					Template: userTypeTmpl,
					Data:     et,
				})
				newErrorTypes = append(newErrorTypes, et)
			}
		}

		for _, et := range newErrorTypes {
			secs = append(secs, &codegen.Section{
				Template: errorTmpl,
				Data:     et,
			})
		}
		return secs
	}

	return codegen.NewSource(path, sections)
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
