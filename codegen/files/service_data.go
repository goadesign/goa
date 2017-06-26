package files

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

// Services holds the data computed from the design needed to generate the code
// of the services.
var Services = make(ServicesData)

type (
	// ServicesData encapsulates the data computed from the service designs.
	ServicesData map[string]*ServiceData

	// ServiceData contains the data used to render the code related to a
	// single service.
	ServiceData struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// VarName is the service struct name.
		VarName string
		// Methods lists the service interface methods.
		Methods []*ServiceMethodData
		// UserTypes lists the types definitions that the service
		// depends on.
		UserTypes []*UserTypeData
		// Scope initialized with all the service types.
		Scope *codegen.NameScope
	}

	// ServiceMethodData describes a single service method.
	ServiceMethodData struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// VarName is the Go method name.
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

	// UserTypeData contains the data describing a data type.
	UserTypeData struct {
		// Name is the type name.
		Name string
		// VarName is the corresponding Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// Fields list the object fields. Only present when describing a
		// user type.
		Fields []*FieldData
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
func (d ServicesData) Get(name string) *ServiceData {
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
func (s *ServiceData) Method(name string) *ServiceMethodData {
	for _, m := range s.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d ServicesData) analyze(service *design.ServiceExpr) *ServiceData {
	var (
		scope   *codegen.NameScope
		varName string
		types   []*UserTypeData
		seen    map[string]struct{}
	)
	{
		scope = codegen.NewNameScope()
		varName = codegen.Goify(service.Name, true)
		seen = make(map[string]struct{})
		// Reserve service, payload and result type names
		scope.Unique(service, varName)
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
		}
	}

	var (
		methods []*ServiceMethodData
	)
	{
		methods = make([]*ServiceMethodData, len(service.Methods))
		for i, e := range service.Methods {
			m := buildServiceMethodData(e, scope)
			methods[i] = m
		}
	}

	var (
		desc string
	)
	{
		desc = service.Description
		if desc == "" {
			desc = fmt.Sprintf("%s is the %s service interface.", varName, service.Name)
		}
	}

	data := &ServiceData{
		Name:        service.Name,
		Description: desc,
		VarName:     varName,
		Methods:     methods,
		UserTypes:   types,
		Scope:       scope,
	}
	d[service.Name] = data

	return data
}

// BuildFieldsData returns a ordered slice of field data representing the given
// user type attributes.
func BuildFieldsData(ut design.UserType, scope *codegen.NameScope) []*FieldData {
	obj := design.AsObject(ut.Attribute().Type)
	fields := make([]*FieldData, len(*obj))
	for i, nat := range *obj {
		fields[i] = &FieldData{
			Name:         nat.Name,
			VarName:      codegen.Goify(nat.Name, true),
			TypeRef:      scope.GoTypeRef(nat.Attribute.Type),
			Required:     ut.Attribute().IsRequired(nat.Name),
			DefaultValue: nat.Attribute.DefaultValue,
		}
	}

	return fields
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
		fields := BuildFieldsData(dt, scope)
		data = append(data, &UserTypeData{
			Name:        dt.Name(),
			VarName:     scope.GoTypeName(dt),
			Description: dt.Attribute().Description,
			Fields:      fields,
			Def:         scope.GoTypeDef(dt.Attribute(), req),
			Ref:         scope.GoTypeRef(dt),
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

// buildServiceMethodData creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func buildServiceMethodData(m *design.MethodExpr, scope *codegen.NameScope) *ServiceMethodData {
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
				payloadName = scope.GoTypeName(dt)
				payloadRef = "*" + payloadName
				payloadDef = scope.GoTypeDef(dt.Attribute(), true)
			default:
				payloadName = scope.GoTypeName(dt)
				payloadRef = payloadName
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
				resultName = scope.GoTypeName(dt)
				resultRef = "*" + resultName
				resultDef = scope.GoTypeDef(dt.Attribute(), false)
			default:
				resultName = scope.GoTypeName(dt)
				resultRef = resultName
			}
			resultDesc = m.Result.Description
			if resultDesc == "" {
				resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
					resultName, m.Service.Name, m.Name)
			}
		}
	}
	return &ServiceMethodData{
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
