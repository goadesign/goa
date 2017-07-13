package rest

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design/rest"
)

var (
	funcMap = template.FuncMap{"comment": codegen.Comment}

	typeDeclTmpl = template.Must(
		template.New("typeDecl").Funcs(funcMap).Parse(typeDeclT),
	)
	typeInitTmpl = template.Must(
		template.New("typeInit").Funcs(funcMap).Parse(typeInitT),
	)
	bodyInitTmpl = template.Must(
		template.New("bodyInit").Funcs(funcMap).Parse(bodyInitT),
	)
	validateTmpl = template.Must(
		template.New("validate").Funcs(funcMap).Parse(validateT),
	)
)

// Types returns the HTTP transport type files.
func Types(root *rest.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.Resources))
	seen := make(map[string]struct{})
	for i, r := range root.Resources {
		fw[i] = Type(r, seen)
	}
	return fw
}

// Type return the file containing the type definitions used by the HTTP
// transport for the given resource (service). seen keeps track of the names of
// the types that have already been generated to prevent duplicate code
// generation.
//
// Below are the rules governing whether values are pointers or not. Note that
// the rules only applies to values that hold primitive types, values that hold
// slices, maps or objects always use pointers either implicitly - slices and
// maps - or explicitly - objects.
//
//   * The payload struct fields (if a struct) hold pointers when not required
//     and have no default value.
//
//   * Request body fields (if the body is a struct) always hold pointers to
//     allow for explicit validation.
//
//   * Request header, path and query string parameter variables hold pointers
//     when not required. Request header, body fields and param variables that
//     have default values are never required (enforced by DSL engine).
//
//   * The result struct fields (if a struct) hold pointers when not required
//     or have a default value (so generated code can set when null)
//
//   * Response body fields (if the body is a struct) and header variables hold
//     pointers when not required and have no default value.
//
func Type(r *rest.ResourceExpr, seen map[string]struct{}) codegen.File {
	var (
		path     string
		sections func(string) []*codegen.Section

		rdata = Resources.Get(r.Name())
	)
	path = filepath.Join(codegen.SnakeCase(r.Name()), "transport", "http_types.go")
	sections = func(genPkg string) []*codegen.Section {
		header := codegen.Header(r.Name()+" HTTP transport types", "transport",
			[]*codegen.ImportSpec{
				{Path: genPkg + "/" + files.Services.Get(r.Name()).PkgName},
				{Path: "goa.design/goa.v2", Name: "goa"},
			},
		)

		var (
			initData       []*InitData
			validatedTypes []*TypeData

			secs = []*codegen.Section{header}
		)

		// request body types
		for _, a := range r.Actions {
			adata := rdata.Action(a.Name())
			if data := adata.Payload.Request.Body; data != nil {
				if data.Def != "" {
					secs = append(secs, &codegen.Section{
						Template: typeDeclTmpl,
						Data:     data,
					})
				}
				if data.Init != nil {
					initData = append(initData, data.Init)
				}
				if data.ValidateDef != "" {
					validatedTypes = append(validatedTypes, data)
				}
			}
		}

		// response body types
		for _, a := range r.Actions {
			adata := rdata.Action(a.Name())
			for _, resp := range adata.Result.Responses {
				if data := resp.Body; data != nil {
					if data.Def != "" {
						secs = append(secs, &codegen.Section{
							Template: typeDeclTmpl,
							Data:     data,
						})
					}
					if data.Init != nil {
						initData = append(initData, data.Init)
					}
					if data.ValidateDef != "" {
						validatedTypes = append(validatedTypes, data)
					}
				}
			}
		}

		// error body types
		for _, a := range r.Actions {
			adata := rdata.Action(a.Name())
			for _, herr := range adata.Errors {
				if data := herr.Response.Body; data != nil {
					if data.Def != "" {
						secs = append(secs, &codegen.Section{
							Template: typeDeclTmpl,
							Data:     data,
						})
					}
					if data.Init != nil {
						initData = append(initData, data.Init)
					}
					if data.ValidateDef != "" {
						validatedTypes = append(validatedTypes, data)
					}
				}
			}
		}

		// body attribute types
		for _, data := range rdata.BodyAttributeTypes {
			if data.Def != "" {
				secs = append(secs, &codegen.Section{
					Template: typeDeclTmpl,
					Data:     data,
				})
			}
		}

		// body constructors
		for _, init := range initData {
			secs = append(secs, &codegen.Section{
				Template: bodyInitTmpl,
				Data:     init,
			})
		}

		for _, adata := range rdata.Actions {

			// request to method payload (server)
			if init := adata.Payload.Request.PayloadInit; init != nil {
				secs = append(secs, &codegen.Section{
					Template: typeInitTmpl,
					Data:     init,
				})
			}

			// response to method result (client)
			for _, resp := range adata.Result.Responses {
				if init := resp.ResultInit; init != nil {
					secs = append(secs, &codegen.Section{
						Template: typeInitTmpl,
						Data:     init,
					})
				}
			}
		}

		// validate methods
		for _, data := range validatedTypes {
			secs = append(secs, &codegen.Section{
				Template: validateTmpl,
				Data:     data,
			})
		}

		return secs
	}
	return codegen.NewSource(path, sections)
}

// input: TypeData
const typeDeclT = `{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
`

// input: InitData
const typeInitT = `{{ comment .Description }}
func {{ .Name }}({{- range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{- if .Code }}
		{{ .Code }}
		{{- if .ReturnTypeAttribute }}
		res := &{{ .ReturnTypeName }}{
			{{ .ReturnTypeAttribute }}: v,
		}
		{{- end }}
		{{- if .ReturnIsStruct }}
			{{- range .Args }}
				{{- if .FieldName -}}
			v.{{ .FieldName }} = {{ if .Pointer }}&{{ end }}{{ .Name }}
				{{ end }}
			{{- end }}
		{{- end }}
		return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
	{{- else }}
		{{- if .ReturnIsStruct }}
			return &{{ .ReturnTypeName }}{
			{{- range .Args }}
				{{- if .FieldName }}
				{{ .FieldName }}: {{ if .Pointer }}&{{ end }}{{ .Name }},
				{{- end }}
			{{- end }}
			}
		{{- end }}
	{{ end -}}
}
`

// input: InitData
const bodyInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
	return body
}
`

// input: TypeData
const validateT = `{{ comment .Description }}
func (body *{{ .VarName }}) Validate() (err error) {
	{{ .ValidateDef }}
	return
}
`
