package codegen

import (
	"path/filepath"

	"goa.design/goa/codegen"
	httpdesign "goa.design/goa/http/design"
)

// ClientTypeFiles returns the HTTP transport client types files.
func ClientTypeFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.HTTPServices))
	seen := make(map[string]struct{})
	for i, svc := range root.HTTPServices {
		fw[i] = clientType(genpkg, svc, seen)
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
func clientType(genpkg string, svc *httpdesign.ServiceExpr, seen map[string]struct{}) *codegen.File {
	var (
		path  string
		rdata = HTTPServices.Get(svc.Name())
	)
	path = filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "types.go")
	sd := HTTPServices.Get(svc.Name())
	header := codegen.Header(svc.Name()+" HTTP client types", "client",
		[]*codegen.ImportSpec{
			{Path: "unicode/utf8"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: sd.Service.PkgName},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/" + "views", Name: sd.Service.ViewsPkg},
			{Path: "goa.design/goa", Name: "goa"},
		},
	)

	var (
		initData       []*InitData
		validatedTypes []*TypeData

		sections = []*codegen.SectionTemplate{header}
	)

	// request body types
	for _, a := range svc.HTTPEndpoints {
		adata := rdata.Endpoint(a.Name())
		if data := adata.Payload.Request.ClientBody; data != nil {
			if data.Def != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-request-body",
					Source: typeDeclT,
					Data:   data,
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
			if data := resp.ClientBody; data != nil && adata.Method.ViewedResult == nil {
				if data.Def != "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-response-body",
						Source: typeDeclT,
						Data:   data,
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
		for _, gerr := range adata.Errors {
			for _, herr := range gerr.Errors {
				if data := herr.Response.ClientBody; data != nil {
					if data.Def != "" {
						sections = append(sections, &codegen.SectionTemplate{
							Name:   "client-error-body",
							Source: typeDeclT,
							Data:   data,
						})
					}
					if data.ValidateDef != "" {
						validatedTypes = append(validatedTypes, data)
					}
				}
			}
		}
	}

	for _, data := range rdata.ClientBodyAttributeTypes {
		if data.Def != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-body-attributes",
				Source: typeDeclT,
				Data:   data,
			})
		}

		if data.ValidateDef != "" {
			validatedTypes = append(validatedTypes, data)
		}
	}

	// body constructors
	for _, init := range initData {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-body-init",
			Source: clientBodyInitT,
			Data:   init,
		})
	}

	for _, adata := range rdata.Endpoints {
		// response to method result (client)
		for _, resp := range adata.Result.Responses {
			if init := resp.ResultInit; init != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-result-init",
					Source: clientTypeInitT,
					Data:   init,
				})
			}
		}

		// error response to method result (client)
		for _, gerr := range adata.Errors {
			for _, herr := range gerr.Errors {
				if init := herr.Response.ResultInit; init != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-error-result-init",
						Source: clientTypeInitT,
						Data:   init,
					})
				}
			}
		}
	}

	// body attribute types
	// validate methods
	for _, data := range validatedTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-validate",
			Source: validateT,
			Data:   data,
		})
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: InitData
const clientBodyInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .ClientArgs }}{{ .Name }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ClientCode }}
	return body
}
`

// input: InitData
const clientTypeInitT = `{{ comment .Description }}
func {{ .Name }}({{- range .ClientArgs }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{- if .ClientCode }}
		{{ .ClientCode }}
		{{- if .ReturnTypeAttribute }}
		res := &{{ .ReturnTypeName }}{
			{{ .ReturnTypeAttribute }}: v,
		}
		{{- end }}
		{{- if .ReturnIsStruct }}
			{{- range .ClientArgs }}
				{{- if .FieldName }}
			v.{{ .FieldName }} = {{ if .Pointer }}&{{ end }}{{ .Name }}
				{{- end }}
			{{- end }}
		{{- end }}
		return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
	{{- else }}
		{{- if .ReturnIsStruct }}
			return &{{ .ReturnTypeName }}{
			{{- range .ClientArgs }}
				{{- if .FieldName }}
				{{ .FieldName }}: {{ if .Pointer }}&{{ end }}{{ .Name }},
				{{- end }}
			{{- end }}
			}
		{{- end }}
	{{ end -}}
}
`

const toResultT = `{{ comment .Description }}
func {{ .VarName }}(vRes {{ .Ref }}) {{ .ReturnRef }} {
{{ .Code }}
	return res
}
`
