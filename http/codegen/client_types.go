package codegen

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/service"
	httpdesign "goa.design/goa.v2/http/design"
)

var (
	clientTypeInitTmpl = template.Must(
		template.New("clientTypeInit").Funcs(funcMap).Parse(clientTypeInitT),
	)
	clientBodyInitTmpl = template.Must(
		template.New("clientBodyInit").Funcs(funcMap).Parse(clientBodyInitT),
	)
)

// ClientTypeFiles returns the HTTP transport client types files.
func ClientTypeFiles(root *httpdesign.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.HTTPServices))
	seen := make(map[string]struct{})
	for i, svc := range root.HTTPServices {
		fw[i] = clientType(svc, seen)
	}
	return fw
}

// clientType return the file containing the type definitions used by the HTTP
// transport for the given service client. seen keeps track of the names of the
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
//   * Request and response body fields (if the body is a struct) always hold
//     pointers to allow for explicit validation.
//
//   * Request header, path and query string parameter variables hold pointers
//     when not required. Request header, body fields and param variables that
//     have default values are never required (enforced by DSL engine).
//
//   * The result struct fields (if a struct) hold pointers when not required
//     or have a default value (so generated code can set when null)
//
//   * Response header variables hold pointers when not required and have no
//     default value.
//
func clientType(svc *httpdesign.ServiceExpr, seen map[string]struct{}) codegen.File {
	var (
		path     string
		sections func(string) []*codegen.Section

		rdata = HTTPServices.Get(svc.Name())
	)
	path = filepath.Join("http", codegen.SnakeCase(svc.Name()), "client", "types.go")
	sections = func(genPkg string) []*codegen.Section {
		header := codegen.Header(svc.Name()+" HTTP client types", "client",
			[]*codegen.ImportSpec{
				{Path: "unicode/utf8"},
				{Path: genPkg + "/" + service.Services.Get(svc.Name()).PkgName},
				{Path: "goa.design/goa.v2", Name: "goa"},
			},
		)

		var (
			initData       []*InitData
			validatedTypes []*TypeData

			secs = []*codegen.Section{header}
		)

		// request body types
		for _, a := range svc.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			if data := adata.Payload.Request.ClientBody; data != nil {
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
		for _, a := range svc.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			for _, resp := range adata.Result.Responses {
				if data := resp.ClientBody; data != nil {
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
		}

		// error body types
		for _, a := range svc.HTTPEndpoints {
			adata := rdata.Endpoint(a.Name())
			for _, herr := range adata.Errors {
				if data := herr.Response.ClientBody; data != nil {
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
		}

		// body attribute types
		for _, data := range rdata.ClientBodyAttributeTypes {
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
				Template: clientBodyInitTmpl,
				Data:     init,
			})
		}

		for _, adata := range rdata.Endpoints {
			// response to method result (client)
			for _, resp := range adata.Result.Responses {
				if init := resp.ResultInit; init != nil {
					secs = append(secs, &codegen.Section{
						Template: clientTypeInitTmpl,
						Data:     init,
					})
				}
			}

			// error response to method result (client)
			for _, herr := range adata.Errors {
				if init := herr.Response.ResultInit; init != nil {
					secs = append(secs, &codegen.Section{
						Template: clientTypeInitTmpl,
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

// input: InitData
const clientBodyInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ClientCode }}
	return body
}
`

// input: InitData
const clientTypeInitT = `{{ comment .Description }}
func {{ .Name }}({{- range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{- if .ClientCode }}
		{{ .ClientCode }}
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
