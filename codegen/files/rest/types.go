package rest

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/codegen/rest"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

type (
	// PayloadInitData contains the data needed to render the
	// payload constructor.
	PayloadInitData struct {
		// Name is the function name.
		Name string
		// Description is the function description.
		Description string
		// FullTypeName is the qualified (including the package name)
		// name of the payload type.
		FullTypeName string
		// BodyTypeRef is a reference to the body type
		BodyTypeRef string
		// BodyFieldsNoDefault contain the list of body struct fields
		// that correspond to attributes with no default value.
		BodyFieldsNoDefault []*files.FieldData
		// BodyFieldsDefault contain the list of body struct fields
		// that correspond to attributes with default value.
		BodyFieldsDefault []*files.FieldData
		// Params is the list of constructor parameters other than body.
		Params []*ParamData
		// Attribute contains the payload attribute.
		Attribute *design.AttributeExpr
	}

	// ValidateData contains the data needed to render the body types
	// Validate methods.
	ValidateData struct {
		// Description is the type description.
		Description string
		// VarName is the Go type name.
		VarName string
		// Validate contains the Go validation code.
		Validate string
	}
)

var (
	funcMap = template.FuncMap{"comment": codegen.Comment}

	typeDeclTmpl = template.Must(
		template.New("typeDecl").Funcs(funcMap).Parse(typeDeclT),
	)
	payloadInitTmpl = template.Must(
		template.New("payloadInit").Funcs(funcMap).Parse(payloadInitT),
	)
	validateTmpl = template.Must(
		template.New("validate").Funcs(funcMap).Parse(validateT),
	)
)

// Types returns the HTTP transport type files.
func Types(root *rest.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.Resources))
	for i, r := range root.Resources {
		fw[i] = Type(r)
	}
	return fw
}

// Type return the file containing the type definitions used by the HTTP
// transport.
//
// Below are the rules governing whether values are pointers or not. Note that
// the rules only applies to values that hold primitive types, values that hold
// slices, maps or objects always use pointers either implicitly - slices and
// maps - or explicitly - objects.
//
//   * Request body fields (if the body is a struct) always hold pointers to
//     allow for explicit validation.
//
//   * Request header, path and query string parameter variables hold pointers
//     when not required. Request header, body fields and param variables that
//     have default values are never required (enforced by DSL engine).
//
//   * The payload and result struct fields (if a struct) hold pointers when not
//     required *and* have no default value.
//
//   * Response body fields (if the body is a struct) and header variables hold
//     pointers when not required and have no default value.
//
func Type(r *rest.ResourceExpr) codegen.File {
	path := filepath.Join(codegen.SnakeCase(r.Name()), "transport", "http_types.go")
	sections := func(genPkg string) []*codegen.Section {
		types := requestBodyTypes(r)
		types = append(types, responseBodyTypes(r)...)
		inits := payloadInits(r)

		header := codegen.Header(r.Name()+" HTTP transport types", "transport",
			[]*codegen.ImportSpec{
				{Path: genPkg + "/" + files.Services.Get(r.Name()).PkgName},
				{Path: "goa.design/goa.v2", Name: "goa"},
			},
		)

		secs := []*codegen.Section{header}
		for _, typ := range types {
			secs = append(secs, &codegen.Section{
				Template: typeDeclTmpl,
				Data:     typ,
			})
		}
		for _, ini := range inits {
			secs = append(secs, &codegen.Section{
				Template: payloadInitTmpl,
				Data:     ini,
			})
		}
		for _, typ := range types {
			if typ.ValidateDef == "" {
				continue
			}
			secs = append(secs, &codegen.Section{
				Template: validateTmpl,
				Data:     typ,
			})
		}
		return secs
	}
	return codegen.NewSource(path, sections)
}

func requestBodyTypes(r *rest.ResourceExpr) []*TypeData {
	var types []*TypeData
	scope := files.Services.Get(r.Name()).Scope
	seen := make(map[string]struct{})
	for _, a := range r.Actions {
		if a.MethodExpr.Payload.Type == design.Empty {
			continue
		}
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		if body == nil || body == design.Empty {
			continue
		}
		ut, ok := body.(design.UserType)
		if !ok {
			continue // nothing to generate
		}
		if _, ok := seen[ut.Name()]; ok {
			continue // already in the list
		}
		seen[ut.Name()] = struct{}{}
		var (
			name        string
			desc        string
			def         string
			validate    string
			validateRef string
		)
		{
			name = scope.GoTypeName(ut)
			desc = ut.Attribute().Description
			if desc == "" {
				desc = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint request body.", name, r.Name(), a.Name())
			}
			def = restgen.GoTypeDef(scope, ut.Attribute(), true, true)
			validate = codegen.RecursiveValidationCode(ut.Attribute(), true, false, "body") //
			if validate != "" {
				validateRef = "err = goa.MergeErrors(err, body.Validate())"
			}
		}
		types = append(types, &TypeData{
			Name:        ut.Name(),
			VarName:     name,
			Description: desc,
			Def:         def,
			Ref:         scope.GoTypeRef(ut),
			ValidateDef: validate,
			ValidateRef: validateRef,
		})
	}
	return types
}

func responseBodyTypes(r *rest.ResourceExpr) []*TypeData {
	var types []*TypeData
	scope := files.Services.Get(r.Name()).Scope
	seen := make(map[string]struct{})
	for _, a := range r.Actions {
		for _, resp := range a.Responses {
			var suffix string
			if len(a.Responses) > 1 {
				suffix = http.StatusText(resp.StatusCode)
			}
			body := restgen.ResponseBodyType(r, resp, a.MethodExpr.Result, suffix)
			if body == nil || body == design.Empty {
				continue
			}
			ut, ok := body.(design.UserType)
			if !ok {
				continue // nothing to generate
			}
			if _, ok := seen[ut.Name()]; ok {
				continue // already in the list
			}
			seen[ut.Name()] = struct{}{}
			var (
				desc        string
				name        string
				def         string
				validate    string
				validateRef string
			)
			{
				name = scope.GoTypeName(ut)
				desc = ut.Attribute().Description
				if desc == "" {
					desc = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s response body.", name, r.Name(), a.Name(), http.StatusText(resp.StatusCode))
				}
				def = restgen.GoTypeDef(scope, ut.Attribute(), false, false)
				validate = codegen.RecursiveValidationCode(ut.Attribute(), true, true, "body")
				if validate != "" {
					validateRef = "err = goa.MergeErrors(err, body.Validate())"
				}
			}
			types = append(types, &TypeData{
				Name:        ut.Name(),
				VarName:     name,
				Description: desc,
				Def:         def,
				Ref:         scope.GoTypeRef(ut),
				ValidateDef: validate,
				ValidateRef: validateRef,
			})
		}
	}
	return types
}

func payloadInits(r *rest.ResourceExpr) []*PayloadInitData {
	var data []*PayloadInitData
	seen := make(map[string]struct{})
	for _, a := range r.Actions {
		if a.MethodExpr.Payload.Type == design.Empty {
			continue // no payload
		}
		ut, ok := a.MethodExpr.Payload.Type.(design.UserType)
		if !ok {
			continue
		}
		if _, ok := seen[ut.Name()]; ok {
			continue // already in the list
		}
		seen[ut.Name()] = struct{}{}
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		but, _ := body.(design.UserType)
		data = append(data, payloadInitData(r, a, but))
	}
	return data
}

// payloadInitData computes the data needed to generate the payload constructor.
// a must be backed by an endpoint that has a payload type that requires a
// constructor (i.e. a user type).
func payloadInitData(r *rest.ResourceExpr, a *rest.ActionExpr, body design.UserType) *PayloadInitData {
	var (
		name                string
		desc                string
		fullTypeName        string
		bodyRef             string
		bodyFieldsNoDefault []*files.FieldData
		bodyFieldsDefault   []*files.FieldData
		params              []*ParamData
	)
	{
		svc := files.Services.Get(r.Name())
		fullTypeName = svc.PkgName + "." + svc.Method(a.Name()).Payload
		name = fmt.Sprintf("New%s", svc.Method(a.Name()).Payload)
		desc = fmt.Sprintf("%s instantiates and validates the %s service %s endpoint payload.",
			name,
			r.Name(),
			a.Name())
		if body != design.Empty {
			bodyRef = svc.Scope.GoTypeRef(body)
		}

		bfields := rest.NewMappedAttributeExpr(body.Attribute())
		restgen.WalkMappedAttr(bfields, func(name, elem string, required bool, a *design.AttributeExpr) error {
			field := &files.FieldData{
				Name:         name,
				VarName:      codegen.GoifyAtt(a, name, true),
				TypeRef:      svc.Scope.GoTypeRef(a.Type),
				Required:     required,
				DefaultValue: a.DefaultValue,
			}
			if a.DefaultValue != nil {
				bodyFieldsDefault = append(bodyFieldsDefault, field)
				return nil
			}
			bodyFieldsNoDefault = append(bodyFieldsNoDefault, field)
			return nil
		})

		queryParams := a.QueryParams()
		all := a.AllParams()
		restgen.WalkMappedAttr(all, func(name, elem string, required bool, att *design.AttributeExpr) error {
			pointer := ""
			if queryParams.IsPrimitivePointer(name, true) {
				pointer = "*"
			}
			// Use payload attribute so it works with path params
			// that may not be required but whose corresponding
			// payload attribute may be.
			req := a.MethodExpr.Payload.IsRequired(name)
			param := &ParamData{
				Name:         elem,
				FieldName:    codegen.GoifyAtt(att, name, true),
				VarName:      codegen.Goify(elem, false),
				Required:     req,
				Pointer:      pointer != "",
				TypeRef:      pointer + svc.Scope.GoTypeRef(att.Type),
				DefaultValue: att.DefaultValue,
			}
			params = append(params, param)
			return nil
		})

		headerParams := a.MappedHeaders()
		restgen.WalkMappedAttr(headerParams, func(name, elem string, required bool, a *design.AttributeExpr) error {
			pointer := ""
			if headerParams.IsPrimitivePointer(name, true) {
				pointer = "*"
			}
			param := &ParamData{
				VarName:      codegen.Goify(elem, false),
				FieldName:    codegen.GoifyAtt(a, name, true),
				TypeRef:      pointer + svc.Scope.GoTypeRef(a.Type),
				DefaultValue: a.DefaultValue,
			}
			params = append(params, param)
			return nil
		})
	}
	return &PayloadInitData{
		Name:                name,
		Description:         desc,
		FullTypeName:        fullTypeName,
		BodyTypeRef:         bodyRef,
		BodyFieldsNoDefault: bodyFieldsNoDefault,
		BodyFieldsDefault:   bodyFieldsDefault,
		Params:              params,
		Attribute:           a.MethodExpr.Payload,
	}
}

const typeDeclT = `{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
`

const payloadInitT = `{{ comment .Description }}
func {{ .Name }}({{ if .BodyTypeRef }}body {{ .BodyTypeRef }}, {{ end }}
{{- range .Params }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) *{{ .FullTypeName }} {
	p := {{ .FullTypeName }}{
	{{- range .BodyFieldsNoDefault }}
		{{ .VarName }}: {{ if .Required }}*{{ end }}body.{{ .VarName }},
	{{- end }}
	{{- range .Params }}
		{{ .FieldName }}: {{ if and (not .Required) (not .Pointer) ($.Attribute.IsPrimitivePointer .Name true) }}{{/* i.e. path params */}}&{{ end }}{{ .VarName }},
	{{- end }}
	}
	{{- range .BodyFieldsDefault }}
	if body.{{ .Name }} != nil {
		p.{{ .FieldName }} = *body.{{ .VarName }}
	} else {
		p.{{ .FieldName }} = {{ .DefaultValue }}
	}
	{{- end }}
	return &p
}
`

const validateT = `{{ comment .Description }}
func (body *{{ .VarName }}) Validate() (err error) {
	{{ .ValidateDef }}
	return
}
`
