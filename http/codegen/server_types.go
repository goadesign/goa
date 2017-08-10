package codegen

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/service"
	httpdesign "goa.design/goa.v2/http/design"
)

var (
	funcMap = template.FuncMap{"comment": codegen.Comment}

	typeDeclTmpl = template.Must(
		template.New("typeDecl").Funcs(funcMap).Parse(typeDeclT),
	)
	serverTypeInitTmpl = template.Must(
		template.New("serverTypeInit").Funcs(funcMap).Parse(serverTypeInitT),
	)
	serverBodyInitTmpl = template.Must(
		template.New("serverBodyInit").Funcs(funcMap).Parse(serverBodyInitT),
	)
	validateTmpl = template.Must(
		template.New("validate").Funcs(funcMap).Parse(validateT),
	)
)

// ServerTypeFiles returns the HTTP transport type files.
func ServerTypeFiles(root *httpdesign.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.HTTPServices))
	seen := make(map[string]struct{})
	for i, r := range root.HTTPServices {
		fw[i] = serverType(r, seen)
	}
	return fw
}

// serverType return the file containing the type definitions used by the HTTP
// transport for the given service server. seen keeps track of the names of the
// types that have already been generated to prevent duplicate code generation.
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
func serverType(r *httpdesign.ServiceExpr, seen map[string]struct{}) codegen.File {
	var (
		path     string
		sections func(string) []*codegen.Section

		rdata = HTTPServices.Get(r.Name())
	)
	path = filepath.Join("http", codegen.SnakeCase(r.Name()), "server", "types.go")
	sections = func(genPkg string) []*codegen.Section {
		header := codegen.Header(r.Name()+" HTTP server types", "server",
			[]*codegen.ImportSpec{
				{Path: "unicode/utf8"},
				{Path: genPkg + "/" + service.Services.Get(r.Name()).PkgName},
				{Path: "goa.design/goa.v2", Name: "goa"},
			},
		)

		var (
			initData       []*InitData
			validatedTypes []*TypeData

			secs = []*codegen.Section{header}
		)

		// request body types
		for _, a := range r.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			if data := adata.Payload.Request.ServerBody; data != nil {
				if data.Def != "" {
					secs = append(secs, &codegen.Section{
						Template: typeDeclTmpl,
						Data:     data,
					})
				}
				if data.ValidateDef != "" {
					validatedTypes = append(validatedTypes, data)
				}
			}
		}

		// response body types
		for _, a := range r.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			for _, resp := range adata.Result.Responses {
				if data := resp.ServerBody; data != nil {
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
		for _, a := range r.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			for _, herr := range adata.Errors {
				if data := herr.Response.ServerBody; data != nil {
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
		for _, data := range rdata.ServerBodyAttributeTypes {
			if data.Def != "" {
				secs = append(secs, &codegen.Section{
					Template: typeDeclTmpl,
					Data:     data,
				})
			}

			if data.ValidateDef != "" {
				validatedTypes = append(validatedTypes, data)
			}
		}

		// body constructors
		for _, init := range initData {
			secs = append(secs, &codegen.Section{
				Template: serverBodyInitTmpl,
				Data:     init,
			})
		}

		for _, adata := range rdata.Endpoints {
			// request to method payload
			if init := adata.Payload.Request.PayloadInit; init != nil {
				secs = append(secs, &codegen.Section{
					Template: serverTypeInitTmpl,
					Data:     init,
				})
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
const serverTypeInitT = `{{ comment .Description }}
func {{ .Name }}({{- range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{- if .ServerCode }}
		{{ .ServerCode }}
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
const serverBodyInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ServerCode }}
	return body
}
`

// input: TypeData
const validateT = `{{ printf "Validate runs the validations defined on %s" .Name | comment }}
func (body *{{ .VarName }}) Validate() (err error) {
	{{ .ValidateDef }}
	return
}
`
