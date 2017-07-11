package rest

import (
	"fmt"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/codegen/rest"
	"goa.design/goa.v2/design"
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

		scope = files.Services.Get(r.Name()).Scope
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
			bodyAttributeTypes []*TypeData
			validatedTypes     []*TypeData

			secs = []*codegen.Section{header}
		)

		// request body types
		for _, a := range r.Actions {
			body := restgen.RequestBodyType(r, a)
			if data := bodyTypeData(body, r, a, true, scope, seen); data != nil {
				secs = append(secs, &codegen.Section{
					Template: typeDeclTmpl,
					Data:     data,
				})
				if data.ValidateDef != "" {
					validatedTypes = append(validatedTypes, data)
				}
			}
			collectUserTypes(body, func(ut design.UserType) {
				bodyAttributeTypes = append(bodyAttributeTypes, attributeTypeData(ut, true, scope, seen))
			})
		}

		// response body types
		for _, a := range r.Actions {
			for _, resp := range a.Responses {
				body := restgen.ResponseBodyType(r, a, resp)
				if data := bodyTypeData(body, r, a, false, scope, seen); data != nil {
					secs = append(secs, &codegen.Section{
						Template: typeDeclTmpl,
						Data:     data,
					})
					if data.ValidateDef != "" {
						validatedTypes = append(validatedTypes, data)
					}
				}
				collectUserTypes(body, func(ut design.UserType) {
					bodyAttributeTypes = append(bodyAttributeTypes, attributeTypeData(ut, false, scope, seen))
				})
			}
		}

		// error body types
		for _, a := range r.Actions {
			for _, herr := range a.HTTPErrors {
				body := restgen.ResponseBodyType(r, a, herr.Response)
				if data := bodyTypeData(body, r, a, false, scope, seen); data != nil {
					secs = append(secs, &codegen.Section{
						Template: typeDeclTmpl,
						Data:     data,
					})
					if data.ValidateDef != "" {
						validatedTypes = append(validatedTypes, data)
					}
				}
				collectUserTypes(body, func(ut design.UserType) {
					bodyAttributeTypes = append(bodyAttributeTypes, attributeTypeData(ut, false, scope, seen))
				})
			}
		}

		// body attribute types
		for _, typ := range bodyAttributeTypes {
			secs = append(secs, &codegen.Section{
				Template: typeDeclTmpl,
				Data:     typ,
			})
		}

		rdata := Resources.Get(r.Name())
		for _, a := range rdata.Actions {

			// method payload to request body (client)
			if body := a.Payload.Request.Body; body != nil && body.Init != nil {
				secs = append(secs, &codegen.Section{
					Template: bodyInitTmpl,
					Data:     body.Init,
				})
			}

			// request to method payload (server)
			if init := a.Payload.Request.PayloadInit; init != nil {
				secs = append(secs, &codegen.Section{
					Template: typeInitTmpl,
					Data:     init,
				})
			}

			// method result to response body (server)
			for _, resp := range a.Result.Responses {
				if body := resp.Body; body != nil && body.Init != nil {
					secs = append(secs, &codegen.Section{
						Template: bodyInitTmpl,
						Data:     body.Init,
					})
				}
			}

			// method error to response body (server)
			for _, aerr := range a.Errors {
				if body := aerr.Response.Body; body != nil && body.Init != nil {
					secs = append(secs, &codegen.Section{
						Template: bodyInitTmpl,
						Data:     body.Init,
					})
				}
			}

			// response to method result (client)
			for _, resp := range a.Result.Responses {
				if init := resp.ResultInit; init != nil {
					secs = append(secs, &codegen.Section{
						Template: typeInitTmpl,
						Data:     init,
					})
				}
			}
		}

		// validate methods
		for _, typ := range validatedTypes {
			secs = append(secs, &codegen.Section{
				Template: validateTmpl,
				Data:     typ,
			})
		}

		return secs
	}
	return codegen.NewSource(path, sections)
}

func bodyTypeData(dt design.DataType, r *rest.ResourceExpr, a *rest.ActionExpr, isRequest bool, scope *codegen.NameScope, seen map[string]struct{}) *TypeData {
	if dt == nil || dt == design.Empty {
		return nil
	}
	ut, ok := dt.(design.UserType)
	if !ok {
		return nil // nothing to generate
	}
	att := &design.AttributeExpr{Type: ut}
	var (
		name        string
		desc        string
		def         string
		validate    string
		validateRef string
	)
	{
		name = scope.GoTypeName(att)
		if _, ok := seen[name]; ok {
			return nil
		}
		seen[name] = struct{}{}
		desc = ut.Attribute().Description
		if desc == "" {
			ctx := "request"
			if !isRequest {
				ctx = "response"
			}
			desc = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s body.", name, r.Name(), a.Name(), ctx)
		}
		def = restgen.GoTypeDef(scope, ut.Attribute(), isRequest, false)
		validate = codegen.RecursiveValidationCode(ut.Attribute(), true, isRequest, "body") //
		if validate != "" {
			validateRef = "err = goa.MergeErrors(err, body.Validate())"
		}
	}
	return &TypeData{
		Name:        ut.Name(),
		VarName:     name,
		Description: desc,
		Def:         def,
		Ref:         scope.GoTypeRef(att),
		ValidateDef: validate,
		ValidateRef: validateRef,
	}
}

func attributeTypeData(ut design.UserType, req bool, scope *codegen.NameScope, seen map[string]struct{}) *TypeData {
	att := &design.AttributeExpr{Type: ut}
	var (
		name        string
		desc        string
		def         string
		validate    string
		validateRef string
	)
	{
		name = scope.GoTypeName(att)
		if _, ok := seen[name]; ok {
			return nil
		}
		seen[name] = struct{}{}
		desc = ut.Attribute().Description
		if desc == "" {
			ctx := "request"
			if !req {
				ctx = "response"
			}
			desc = name + " is used to define fields on " + ctx + " body types."
		}
		def = restgen.GoTypeDef(scope, ut.Attribute(), req, false)
		validate = codegen.RecursiveValidationCode(ut.Attribute(), true, req, "v") //
		if validate != "" {
			validateRef = "err = goa.MergeErrors(err, v.Validate())"
		}
	}
	return &TypeData{
		Name:        ut.Name(),
		VarName:     name,
		Description: desc,
		Def:         def,
		Ref:         scope.GoTypeRef(att),
		ValidateDef: validate,
		ValidateRef: validateRef,
	}
}

// collectUserTypes traverses the given data type recursively and calls back the
// given function for each attribute using a user type.
func collectUserTypes(dt design.DataType, cb func(design.UserType)) {
	switch actual := dt.(type) {
	case *design.Object:
		for _, nat := range *actual {
			collectUserTypes(nat.Attribute.Type, cb)
		}
	case *design.Array:
		collectUserTypes(actual.ElemType.Type, cb)
	case *design.Map:
		collectUserTypes(actual.KeyType.Type, cb)
		collectUserTypes(actual.ElemType.Type, cb)
	case design.UserType:
		cb(actual)
	}
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
func {{ .Name }}(res {{ .TypeRef }}) {{ .BodyTypeRef }} {
	var body {{ .BodyTypeRef }}
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
