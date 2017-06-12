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
		// TypeName is the name of the payload type.
		TypeName string
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
	for _, a := range r.Actions {
		if a.MethodExpr.Payload.Type == design.Empty {
			continue
		}
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		ut, ok := body.(design.UserType)
		if !ok {
			continue // nothing to generate
		}
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
			def = restgen.GoTypeDef(scope, ut.Attribute(), true)
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
			body := restgen.ResponseBodyType(r, resp, a.MethodExpr.Result, suffix)
			ut, ok := body.(design.UserType)
			if !ok {
				continue // nothing to generate
			}
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
				def = restgen.GoTypeDef(scope, ut.Attribute(), true)
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
	for _, a := range r.Actions {
		if a.MethodExpr.Payload.Type == design.Empty {
			continue // no payload
		}
		body := restgen.RequestBodyType(r, a, "ServerRequestBody")
		ut, ok := body.(design.UserType)
		if ok && ut == a.MethodExpr.Payload.Type {
			continue // no need for a constructor
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
		bodyFieldsNoDefault []*files.FieldData
		bodyFieldsDefault   []*files.FieldData
		params              []*ParamData
	)
	{
		svc := files.Services.Get(r.Name())
		typeName = svc.Method(a.Name()).Payload
		name = fmt.Sprintf("New%s", typeName)
		desc = fmt.Sprintf("%s instantiates and validates the %s service %s endpoint server request body.",
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
		restgen.WalkMappedAttr(all, func(name, elem string, required bool, a *design.AttributeExpr) error {
			pointer := ""
			if queryParams.IsPrimitivePointer(name) {
				pointer = "*"
			}
			param := &ParamData{
				Name:         elem,
				VarName:      codegen.Goify(elem, false),
				FieldName:    codegen.GoifyAtt(a, name, true),
				TypeRef:      pointer + svc.Scope.GoTypeRef(a.Type),
				DefaultValue: a.DefaultValue,
			}
			params = append(params, param)
			return nil
		})

		headerParams := a.MappedHeaders()
		restgen.WalkMappedAttr(headerParams, func(name, elem string, required bool, a *design.AttributeExpr) error {
			pointer := ""
			if headerParams.IsPrimitivePointer(name) {
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
		TypeName:            typeName,
		BodyTypeRef:         bodyRef,
		BodyFieldsNoDefault: bodyFieldsNoDefault,
		BodyFieldsDefault:   bodyFieldsDefault,
		Params:              params,
	}
}

const typeDeclT = `{{ comment .Description }}
type {{ .VarName }} {{ .TypeDef }}
`

const payloadInitT = `{{ comment .Description }}
func {{ .Name }}({{ if .BodyTypeRef }}body {{ .BodyTypeRef }}, {{ end }}
{{- range .Params }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) (*{{ .TypeName }}, error) {
	p := {{ .TypeName }}{
	{{- range .BodyFieldsNoDefault }}
		{{ .VarName }}: body.{{ .VarName }},
	{{- end }}
	{{- range .Params }}
		{{ .FieldName }}: {{ .VarName }},
	{{- end }}
	}
	{{- range .BodyFieldsDefault }}
	if body.{{ .Name }} != nil {
		p.{{ .FieldName }} = *body.{{ .VarName }}
	} else {
		p.{{ .FieldName }} = {{ .DefaultValue }}
	}
	{{- end }}
	return &p, nil
}
`

const validateT = `{{ comment .Description }}
func (body *{{ .VarName }}) Validate() (err error) {
	{{ .Validate }}
	return
}
`
