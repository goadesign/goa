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
	// TypeData contains the data needed to render a type definition.
	TypeData struct {
		// VarName is the type name.
		VarName string
		// Description is the type human description.
		Description string
		// TypeDef is the type definition Go code.
		TypeDef string
		// Validate contains the validation code.
		Validate string
	}

	// PayloadInitData contains the data needed to render the
	// payload constructor.
	PayloadInitData struct {
		// Name is the function name.
		Name string
		// Description is the function description.
		Description string
		// TypeName is the name of the payload type.
		TypeName string
		// BodyTypeRef is a reference to the body type
		BodyTypeRef string
		// BodyFieldsNoDefault contain the list of body struct fields
		// that correspond to attributes with no default value.
		BodyFieldsNoDefault []*FieldData
		// BodyFieldsDefault contain the list of body struct fields
		// that correspond to attributes with default value.
		BodyFieldsDefault []*FieldData
		// ParamsNoDefault is the list of constructor parameters other
		// than body that correspond to attributes with no default value.
		ParamsNoDefault []*ParamData
		// ParamsNoDefault is the list of constructor parameters other
		// than body that correspond to attributes with a default value.
		ParamsDefault []*ParamData
	}

	// ParamData contains the data needed to render a single parameter.
	ParamData struct {
		// VarName is the name of the variable holding the param value.
		VarName string
		// FieldName is the name of the type field to be initialized
		// with the param value.
		FieldName string
		// TypeRef is a reference to the parameter type.
		TypeRef string
		// DefaultValue is the parameter attribute default value if any.
		DefaultValue interface{}
	}

	// FieldData contains the data needed to render a single field.
	FieldData struct {
		// FieldName is the name of the payload / body field.
		FieldName string
		// Required is true if the field is required.
		Required bool
		// DefaultValue is the payload attribute default value if any.
		DefaultValue interface{}
	}
)

var (
	typeDeclTmpl    = template.Must(template.New("typeDecl").Funcs(template.FuncMap{"comment": codegen.Comment}).Parse(typeDeclT))
	payloadInitTmpl = template.Must(template.New("payloadInit").Funcs(template.FuncMap{"comment": codegen.Comment}).Parse(payloadInitT))
	validateTmpl    = template.Must(template.New("validate").Funcs(template.FuncMap{"comment": codegen.Comment}).Parse(validateT))
)

// MarshalTypes return the file containing the type definitions used by the HTTP
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
func MarshalTypes(r *rest.ResourceExpr) codegen.File {
	path := filepath.Join(codegen.KebabCase(r.Name()), "transport", "http", "types.go")
	sections := func(genPkg string) []*codegen.Section {
		types := requestBodyTypes(r)
		types = append(types, responseBodyTypes(r)...)
		inits := payloadInits(r)

		var secs []*codegen.Section
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
			if v := typ.Validate; v != "" {
				secs = append(secs, &codegen.Section{
					Template: validateTmpl,
					Data:     typ,
				})
			}
		}
		return secs
	}
	return codegen.NewSource(path, sections)
}

func requestBodyTypes(r *rest.ResourceExpr) []*TypeData {
	var types []*TypeData
	scope := files.Services.Get(r.Name()).Scope
	for _, a := range r.Actions {
		if a.EndpointExpr.Payload.Type == design.Empty {
			continue
		}
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		ut, ok := body.(design.UserType)
		if !ok {
			continue // nothing to generate
		}
		var (
			name     string
			desc     string
			def      string
			validate string
		)
		{
			name = scope.GoTypeName(ut)
			desc = ut.Attribute().Description
			if desc == "" {
				desc = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint request body.", name, r.Name(), a.Name())
			}
			def = restgen.GoTypeDef(scope, rest.NewMappedAttributeExpr(ut.Attribute()), true)
			validate = codegen.RecursiveValidationCode(ut.Attribute(), false, false, "body")
		}
		types = append(types, &TypeData{
			VarName:     name,
			Description: desc,
			TypeDef:     def,
			Validate:    validate,
		})
	}
	return types
}

func responseBodyTypes(r *rest.ResourceExpr) []*TypeData {
	var types []*TypeData
	scope := files.Services.Get(r.Name()).Scope
	for _, a := range r.Actions {
		for _, resp := range a.Responses {
			var suffix string
			if len(a.Responses) > 1 {
				suffix = http.StatusText(resp.StatusCode)
			}
			body := restgen.ResponseBodyType(r, resp, a.EndpointExpr.Result, suffix)
			ut, ok := body.(design.UserType)
			if !ok {
				continue // nothing to generate
			}
			var (
				desc string
				name string
				def  string
			)
			{
				name = scope.GoTypeName(ut)
				desc = ut.Attribute().Description
				if desc == "" {
					desc = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s response body.", name, r.Name(), a.Name(), http.StatusText(resp.StatusCode))
				}
				def = restgen.GoTypeDef(scope, rest.NewMappedAttributeExpr(ut.Attribute()), true)
			}
			types = append(types, &TypeData{
				VarName:     name,
				Description: desc,
				TypeDef:     def,
			})
		}
	}
	return types
}

func payloadInits(r *rest.ResourceExpr) []*PayloadInitData {
	var data []*PayloadInitData
	for _, a := range r.Actions {
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		ut, ok := body.(design.UserType)
		if !ok {
			continue // nothing to generate
		}
		data = append(data, payloadInitData(r, a, ut))
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
		typeName            string
		bodyRef             string
		bodyFieldsNoDefault []*FieldData
		bodyFieldsDefault   []*FieldData
		paramsNoDefault     []*ParamData
		paramsDefault       []*ParamData
	)
	{
		svc := files.Services.Get(r.Name())
		typeName = svc.Method(a.Name()).Payload
		name = fmt.Sprintf("New%s", typeName)
		desc = fmt.Sprintf("%s instantiates and validates the %s service %s endpoint server request body.",
			name,
			r.Name(),
			a.Name())
		bodyRef = svc.Scope.GoTypeRef(body)

		bfields := rest.NewMappedAttributeExpr(body.Attribute())
		restgen.WalkMappedAttr(bfields, func(name, elem string, required bool, a *design.AttributeExpr) error {
			field := &FieldData{
				FieldName:    codegen.GoifyAtt(a, name, true),
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

		all := a.AllParams()
		restgen.WalkMappedAttr(all, func(name, elem string, required bool, a *design.AttributeExpr) error {
			param := &ParamData{
				VarName:      codegen.Goify(elem, false),
				FieldName:    codegen.GoifyAtt(a, name, true),
				TypeRef:      svc.Scope.GoTypeRef(a.Type),
				DefaultValue: a.DefaultValue,
			}
			if a.DefaultValue != nil {
				paramsDefault = append(paramsDefault, param)
				return nil
			}
			paramsNoDefault = append(paramsNoDefault, param)
			return nil
		})
	}
	return &PayloadInitData{
		Name:                name,
		Description:         desc,
		TypeName:            typeName,
		BodyTypeRef:         bodyRef,
		BodyFieldsNoDefault: bodyFieldsNoDefault,
		BodyFieldsDefault:   bodyFieldsDefault,
		ParamsNoDefault:     paramsNoDefault,
		ParamsDefault:       paramsDefault,
	}
}

const typeDeclT = `
// {{ comment .Description }}
{{ .TypeName }} {{ .TypeDef }}
`

const payloadInitT = `{{ comment .Description }}
func {{ .Name }}({{ if .BodyTypeRef }}body {{ .BodyTypeRef }}, {{ end }}{{ range .Params }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) (*{{ .TypeName }}, error) {
	p := {{ .TypeName }}{
	{{- range .BodyFieldsNoDefault }}
		{{ .FieldName }}: {{ if .Required }}*{{ end }}{{ .FieldName }},
	{{ end -}}
	{{- range .ParamsNoDefault }}
		{{ .FieldName }}: {{ if .Required }}*{{ end }}{{ .VarName }},
	{{ end -}}
	}
	{{- range .BodyFieldsDefault }}
	if body.{{ .FieldName }} != nil {
		p.{{ .FieldName }} = *body.{{ .FieldName }}
	} else {
		p.{{ .FieldName }} = {{ .DefaultValue }}
	}
	{{ end -}}
	{{- range .ParamsDefault }}
	if {{ .VarName }} != nil {
		p.{{ .FieldName }} = *{{ .VarName }}
	} else {
		p.{{ .FieldName }} = {{ .DefaultValue }}
	}
	{{ end -}}
	return &p, nil
}
`

const validateT = `
// {{ comment .Description }}
func (body *{{ .TypeName }}) Validate() (err error) {
	{{ .Validate }}
	return
}
`
