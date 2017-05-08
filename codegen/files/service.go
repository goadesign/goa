package files

import (
	"fmt"
	"path/filepath"
	"sort"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// serviceData contains the data necessary to render the service
	// template.
	serviceData struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// VarName is the service struct name.
		VarName string
		// Methods lists the service interface methods.
		Methods []*serviceMethod
		// UserTypes lists the types definitions that the service depends on.
		UserTypes []*userType
	}

	// serviceMethod describes a single service method.
	serviceMethod struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// VarName is the method struct name.
		VarName string
		// Payload is the name of the payload type if any,
		Payload string
		// PayloadRef is a reference to the payload type if any,
		PayloadRef string
		// PayloadDesc is the payload type description if any.
		PayloadDesc string
		// PayloadDef is the payload type definition if any.
		PayloadDef string
		// Result is the name of the result type if any.
		Result string
		// ResultDesc is the result type description if any.
		ResultDesc string
		// ResultDef is the result type definition if any.
		ResultDef string
		// ResultRef is the reference to the result type if any.
		ResultRef string
	}

	// userType describes a user type used by the service types.
	userType struct {
		// VarName is the generated type name.
		VarName string
		// Description is the type description if any.
		Description string
		// TypeDef is the type definition.
		TypeDef string
	}
)

var (
	// serviceTmpl is the template used to render the body of the service file.
	serviceTmpl = template.Must(template.New("service").Parse(serviceT))

	// ServiceScope is the naming scope used to render the service types.
	ServiceScope = codegen.NewNameScope()
)

// Service returns the service file for the given service.
func Service(service *design.ServiceExpr) codegen.File {
	path := filepath.Join("service", service.Name+".go")
	sections := func(genPkg string) []*codegen.Section {
		var (
			header, body *codegen.Section
			userTypes    = make(map[string]design.UserType)
		)
		{
			header = codegen.Header(service.Name+" service", "service",
				[]*codegen.ImportSpec{
					{Path: "context"},
					{Path: "goa.design/goa.v2"},
				})
			body = &codegen.Section{
				Template: serviceTmpl,
				Data:     buildServiceData(service, userTypes),
			}
		}

		return []*codegen.Section{header, body}
	}

	return codegen.NewSource(path, sections)
}

// buildServiceData creates the data necessary to render the code of the given
// service. It records the user types needed by the service definition in
// userTypes.
func buildServiceData(service *design.ServiceExpr, userTypes map[string]design.UserType) *serviceData {
	varName := codegen.Goify(service.Name, true)
	ServiceScope.Unique(service, varName) // reserve it
	methods := make([]*serviceMethod, len(service.Endpoints))
	for i, e := range service.Endpoints {
		methods[i] = buildServiceMethod(e, userTypes)
	}
	names := make([]string, len(userTypes))
	i := 0
	for n := range userTypes {
		names[i] = n
		i++
	}
	sort.Strings(names)
	types := make([]*userType, len(userTypes))
	for i, n := range names {
		ut := userTypes[n]
		types[i] = &userType{
			VarName:     ServiceScope.Unique(ut, codegen.Goify(n, true)),
			Description: ut.Attribute().Description,
			TypeDef:     codegen.GoTypeDef(ut.Attribute(), false),
		}
	}
	desc := service.Description
	if desc == "" {
		desc = fmt.Sprintf("%s is the %s service interface.", varName, service.Name)
	}
	return &serviceData{
		Name:        service.Name,
		Description: desc,
		VarName:     varName,
		Methods:     methods,
		UserTypes:   types,
	}
}

// buildServiceMethod creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func buildServiceMethod(m *design.EndpointExpr, userTypes map[string]design.UserType) *serviceMethod {
	var walkTypes func(*design.AttributeExpr) error
	walkTypes = func(at *design.AttributeExpr) error {
		if ut, ok := at.Type.(design.UserType); ok {
			if _, ok := userTypes[ut.Name()]; ok {
				return nil
			}
			userTypes[ut.Name()] = ut
			codegen.Walk(ut.Attribute(), walkTypes)
		} else if design.IsObject(at.Type) {
			for _, catt := range design.AsObject(at.Type) {
				walkTypes(catt)
			}
		}
		return nil
	}

	var (
		varName     string
		desc        string
		payloadName string
		payloadDesc string
		payloadDef  string
		payloadRef  string
		resultName  string
		resultDesc  string
		resultDef   string
		resultRef   string
	)
	{
		varName = codegen.Goify(m.Name, true)
		desc = m.Description
		if desc == "" {
			desc = codegen.Goify(m.Name, true) + " implements " + m.Name + "."
		}
		if m.Payload != nil && m.Payload.Type != design.Empty {
			switch dt := m.Payload.Type.(type) {
			case design.UserType:
				payloadName = ServiceScope.Unique(dt, codegen.GoType(dt, false), "Payload")
				payloadRef = "*" + payloadName
				payloadDef = codegen.GoTypeDef(dt.Attribute(), false)
				walkTypes(dt.Attribute())
			case design.Object:
				payloadName = fmt.Sprintf("%s%sPayload", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				payloadName = ServiceScope.Unique(dt, payloadName, "")
				payloadRef = "*" + payloadName
				payloadDef = codegen.GoTypeDef(m.Payload, false)
				walkTypes(m.Payload)
			case *design.Array:
				payloadName = fmt.Sprintf("%s%sPayload", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				payloadName = ServiceScope.Unique(dt, payloadName, "")
				payloadRef = payloadName
				payloadDef = codegen.GoTypeDef(m.Payload, false)
				walkTypes(dt.ElemType)
			case *design.Map:
				payloadName = fmt.Sprintf("%s%sPayload", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				payloadName = ServiceScope.Unique(dt, payloadName, "")
				payloadRef = payloadName
				payloadDef = codegen.GoTypeDef(m.Payload, false)
				walkTypes(dt.KeyType)
				walkTypes(dt.ElemType)
			default:
				payloadName = codegen.GoNativeType(m.Payload.Type)
			}
			payloadDesc = m.Payload.Description
			if payloadDesc == "" {
				payloadDesc = fmt.Sprintf("%s is the payload type of the %s service %s method.",
					payloadName, m.Service.Name, m.Name)
			}
		}
		if m.Result != nil && m.Result.Type != design.Empty {
			switch dt := m.Result.Type.(type) {
			case design.UserType:
				resultName = ServiceScope.Unique(dt, codegen.GoType(dt, false), "Result")
				resultRef = "*" + resultName
				resultDef = codegen.GoTypeDef(dt.Attribute(), false)
				walkTypes(dt.Attribute())
			case design.Object:
				resultName = fmt.Sprintf("%s%sResult", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				resultName = ServiceScope.Unique(dt, resultName, "")
				resultRef = "*" + resultName
				resultDef = codegen.GoTypeDef(m.Result, false)
				walkTypes(m.Result)
			case *design.Array:
				resultName = fmt.Sprintf("%s%sResult", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				resultName = ServiceScope.Unique(dt, resultName, "")
				resultRef = resultName
				resultDef = codegen.GoTypeDef(m.Result, false)
				walkTypes(dt.ElemType)
			case *design.Map:
				resultName = fmt.Sprintf("%s%sResult", codegen.Goify(m.Service.Name, true), codegen.Goify(m.Name, true))
				resultName = ServiceScope.Unique(dt, resultName, "")
				resultRef = resultName
				resultDef = codegen.GoTypeDef(m.Result, false)
				walkTypes(dt.KeyType)
				walkTypes(dt.ElemType)
			default:
				resultName = codegen.GoNativeType(m.Result.Type)
			}
			resultDesc = m.Result.Description
			if resultDesc == "" {
				resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
					resultName, m.Service.Name, m.Name)
			}
		}
	}
	return &serviceMethod{
		Name:        m.Name,
		VarName:     varName,
		Description: desc,
		Payload:     payloadName,
		PayloadDesc: payloadDesc,
		PayloadDef:  payloadDef,
		PayloadRef:  payloadRef,
		Result:      resultName,
		ResultDesc:  resultDesc,
		ResultDef:   resultDef,
		ResultRef:   resultRef,
	}
}

// serviceT is the template used to write an service definition.
const serviceT = `
{{- define "interface" }}
	// {{ .Description }}
	{{ .VarName }} interface {
{{- range .Methods }}
		// {{ .Description }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) {{ if .Result }}({{ .ResultRef }}, error){{ else }}error{{ end }}
{{- end }}
	}
{{end -}}

{{ define "payloads" -}}
{{ range .Methods -}}
{{ if .PayloadDef }}
	// {{ .PayloadDesc }}
	{{ .Payload }} {{ .PayloadDef }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ define "results" -}}
{{ range .Methods -}}
{{ if .ResultDef }}
	// {{ .ResultDesc }}
	{{ .Result }} {{ .ResultDef }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ define "types" -}}
{{ range .UserTypes }}
{{- if .Description -}}
	// {{ .Description }}
{{- end }}
	{{ .VarName }} {{ .TypeDef }}
{{ end -}}
{{ end -}}

type (
{{- template "interface" . -}}
{{- template "payloads" . -}}
{{- template "results" . -}}
{{- template "types" . -}}
)
`
