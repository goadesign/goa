package service

import (
	"fmt"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// Services holds the data computed from the design needed to generate the code
// of the services.
var Services = make(ServicesData)

type (
	// ServicesData encapsulates the data computed from the service designs.
	ServicesData map[string]*Data

	// Data contains the data used to render the code related to a
	// single service.
	Data struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// VarName is the service struct name.
		VarName string
		// PkgName is the name of the package containing the generated
		// service code.
		PkgName string
		// Methods lists the service interface methods.
		Methods []*MethodData
		// UserTypes lists the types definitions that the service
		// depends on.
		UserTypes []*UserTypeData
		// ErrorTypes lists the error types definitions that the service
		// depends on.
		ErrorTypes []*UserTypeData
		// Errors list the information required to generate error init
		// functions.
		ErrorInits []*ErrorInitData
		// Scope initialized with all the service types.
		Scope *codegen.NameScope
	}

	// ErrorInitData describes an error returned by a service method of type
	// ErrorResult.
	ErrorInitData struct {
		// Name is the name of the init function.
		Name string
		// Description is the error description.
		Description string
		// ErrName is the name of the error.
		ErrName string
		// TypeName is the error struct type name.
		TypeName string
		// TypeRef is the reference to the error type.
		TypeRef string
	}

	// MethodData describes a single service method.
	MethodData struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// VarName is the Go method name.
		VarName string
		// Payload is the name of the payload type if any,
		Payload string
		// PayloadDef is the payload type definition if any.
		PayloadDef string
		// PayloadRef is a reference to the payload type if any,
		PayloadRef string
		// PayloadDesc is the payload type description if any.
		PayloadDesc string
		// PayloadEx is an example of a valid payload value.
		PayloadEx interface{}
		// Result is the name of the result type if any.
		Result string
		// ResultDef is the result type definition if any.
		ResultDef string
		// ResultRef is the reference to the result type if any.
		ResultRef string
		// ResultDesc is the result type description if any.
		ResultDesc string
		// ResultEx is an example of a valid result value.
		ResultEx interface{}
		// Errors list the possible errors defined in the design if any.
		Errors []*ErrorInitData
	}

	// UserTypeData contains the data describing a data type.
	UserTypeData struct {
		// Name is the type name.
		Name string
		// VarName is the corresponding Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// Def is the type definition Go code.
		Def string
		// Ref is the reference to the type.
		Ref string
		// Type is the underlying type.
		Type design.UserType
	}

	// FieldData contains the data needed to render a single field.
	FieldData struct {
		// Name is the name of the attribute.
		Name string
		// VarName is the name of the Go type field.
		VarName string
		// TypeRef is the reference to the field type.
		TypeRef string
		// Required is true if the field is required.
		Required bool
		// DefaultValue is the payload attribute default value if any.
		DefaultValue interface{}
	}
)

// Get retrieves the data for the service with the given name computing it if
// needed. It returns nil if there is no service with the given name.
func (d ServicesData) Get(name string) *Data {
	if data, ok := d[name]; ok {
		return data
	}
	service := design.Root.Service(name)
	if service == nil {
		return nil
	}
	d[name] = d.analyze(service)
	return d[name]
}

// Method returns the service method data for the method with the given name,
// nil if there isn't one.
func (s *Data) Method(name string) *MethodData {
	for _, m := range s.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d ServicesData) analyze(service *design.ServiceExpr) *Data {
	var (
		scope      *codegen.NameScope
		varName    string
		pkgName    string
		types      []*UserTypeData
		errTypes   []*UserTypeData
		errorInits []*ErrorInitData
		seenErrors map[string]struct{}
		seen       map[string]struct{}
	)
	{
		scope = codegen.NewNameScope()
		varName = codegen.Goify(service.Name, true)
		pkgName = strings.ToLower(codegen.Goify(service.Name, false))
		seen = make(map[string]struct{})
		seenErrors = make(map[string]struct{})
		for _, e := range service.Methods {
			// Create user type for raw object payloads
			if _, ok := e.Payload.Type.(*design.Object); ok {
				e.Payload.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Payload),
					TypeName:      fmt.Sprintf("%sPayload", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Payload.Type.(design.UserType); ok {
				seen[ut.Name()] = struct{}{}
			}

			// Create user type for raw object results
			if _, ok := e.Result.Type.(*design.Object); ok {
				e.Result.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Result),
					TypeName:      fmt.Sprintf("%sResult", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Result.Type.(design.UserType); ok {
				seen[ut.Name()] = struct{}{}
			}
		}
		for _, e := range service.Methods {
			patt := e.Payload
			if ut, ok := patt.Type.(design.UserType); ok {
				patt = ut.Attribute()
			}
			types = append(types, collectTypes(patt, seen, scope, true)...)
			ratt := e.Result
			if ut, ok := ratt.Type.(design.UserType); ok {
				ratt = ut.Attribute()
			}
			types = append(types, collectTypes(ratt, seen, scope, false)...)
			for _, er := range e.Errors {
				errTypes = append(errTypes, collectTypes(er.AttributeExpr, seen, scope, false)...)
				if er.Type == design.ErrorResult {
					if _, ok := seenErrors[er.Name]; ok {
						continue
					}
					seenErrors[er.Name] = struct{}{}
					errorInits = append(errorInits, buildErrorInitData(er, scope))
				}
			}
		}
	}

	var (
		methods []*MethodData
	)
	{
		methods = make([]*MethodData, len(service.Methods))
		for i, e := range service.Methods {
			m := buildMethodData(e, pkgName, scope)
			methods[i] = m
		}
	}

	var (
		desc string
	)
	{
		desc = service.Description
		if desc == "" {
			desc = fmt.Sprintf("Service is the %s service interface.", service.Name)
		}
	}

	data := &Data{
		Name:        service.Name,
		Description: desc,
		VarName:     varName,
		PkgName:     pkgName,
		Methods:     methods,
		UserTypes:   types,
		ErrorTypes:  errTypes,
		ErrorInits:  errorInits,
		Scope:       scope,
	}
	d[service.Name] = data

	return data
}

// collectTypes recurses through the attribute to gather all user types and
// records them in userTypes.
func collectTypes(at *design.AttributeExpr, seen map[string]struct{}, scope *codegen.NameScope, req bool) (data []*UserTypeData) {
	if at == nil || at.Type == design.Empty {
		return
	}
	collect := func(at *design.AttributeExpr) []*UserTypeData { return collectTypes(at, seen, scope, req) }
	switch dt := at.Type.(type) {
	case design.UserType:
		if _, ok := seen[dt.Name()]; ok {
			return nil
		}
		data = append(data, &UserTypeData{
			Name:        dt.Name(),
			VarName:     scope.GoTypeName(at),
			Description: dt.Attribute().Description,
			Def:         scope.GoTypeDef(dt.Attribute(), req),
			Ref:         scope.GoTypeRef(at),
			Type:        dt,
		})
		seen[dt.Name()] = struct{}{}
		data = append(data, collect(dt.Attribute())...)
	case *design.Object:
		for _, nat := range *dt {
			data = append(data, collect(nat.Attribute)...)
		}
	case *design.Array:
		data = append(data, collect(dt.ElemType)...)
	case *design.Map:
		data = append(data, collect(dt.KeyType)...)
		data = append(data, collect(dt.ElemType)...)
	}
	return
}

// buildErrorInitData creates the data needed to generate code around endpoint error return values.
func buildErrorInitData(er *design.ErrorExpr, scope *codegen.NameScope) *ErrorInitData {
	return &ErrorInitData{
		Name:        fmt.Sprintf("New%s", codegen.Goify(er.Name, true)),
		Description: er.Description,
		ErrName:     er.Name,
		TypeName:    scope.GoTypeName(er.AttributeExpr),
		TypeRef:     scope.GoTypeRef(er.AttributeExpr),
	}
}

// buildMethodData creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func buildMethodData(m *design.MethodExpr, svcPkgName string, scope *codegen.NameScope) *MethodData {
	var (
		varName     string
		desc        string
		payloadName string
		payloadDef  string
		payloadRef  string
		payloadDesc string
		payloadEx   interface{}
		resultName  string
		resultDef   string
		resultRef   string
		resultDesc  string
		resultEx    interface{}
		errors      []*ErrorInitData
	)
	{
		varName = codegen.Goify(m.Name, true)
		desc = m.Description
		if desc == "" {
			desc = codegen.Goify(m.Name, true) + " implements " + m.Name + "."
		}
		if m.Payload.Type != design.Empty {
			payloadName = scope.GoTypeName(m.Payload)
			payloadRef = scope.GoTypeRef(m.Payload)
			if dt, ok := m.Payload.Type.(design.UserType); ok {
				payloadDef = scope.GoTypeDef(dt.Attribute(), true)
			}
			payloadDesc = m.Payload.Description
			if payloadDesc == "" {
				payloadDesc = fmt.Sprintf("%s is the payload type of the %s service %s method.",
					payloadName, m.Service.Name, m.Name)
			}
			payloadEx = m.Payload.Example(design.Root.API.Random())
		}
		if m.Result.Type != design.Empty {
			resultName = scope.GoTypeName(m.Result)
			resultRef = scope.GoTypeRef(m.Result)
			if dt, ok := m.Result.Type.(design.UserType); ok {
				resultDef = scope.GoTypeDef(dt.Attribute(), true)
			}
			resultDesc = m.Result.Description
			if resultDesc == "" {
				resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
					resultName, m.Service.Name, m.Name)
			}
			resultEx = m.Result.Example(design.Root.API.Random())
		}
		if len(m.Errors) > 0 {
			errors = make([]*ErrorInitData, len(m.Errors))
			for i, er := range m.Errors {
				errors[i] = buildErrorInitData(er, scope)
			}
		}
	}
	return &MethodData{
		Name:        m.Name,
		VarName:     varName,
		Description: desc,
		Payload:     payloadName,
		PayloadDef:  payloadDef,
		PayloadRef:  payloadRef,
		PayloadDesc: payloadDesc,
		PayloadEx:   payloadEx,
		Result:      resultName,
		ResultDef:   resultDef,
		ResultRef:   resultRef,
		ResultDesc:  resultDesc,
		ResultEx:    resultEx,
		Errors:      errors,
	}
}
