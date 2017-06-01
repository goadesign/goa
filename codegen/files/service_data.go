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
		// Name is the endpoint name.
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

	// UserTypeData describes a user type used by the service types.
	UserTypeData struct {
		// Name is the generated type name.
		Name string
		// Description is the type description if any.
		Description string
		// TypeDef is the type definition.
		TypeDef string
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
	return d.analyze(service)
}

// Method returns the service method data for the endpoint with the given
// name, nil if there isn't one.
func (s *ServiceData) Method(endpointName string) *ServiceMethodData {
	for _, m := range s.Methods {
		if m.Name == endpointName {
			return m
		}
	}
	return nil
}

// buildData creates the data necessary to render the code of the given service.
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
		for _, e := range service.Endpoints {
			// Create user type for raw object payloads
			if _, ok := e.Payload.Type.(design.Object); ok {
				e.Payload.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Payload),
					TypeName:      fmt.Sprintf("%sPayload", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Payload.Type.(design.UserType); ok {
				seen[ut.Name()] = struct{}{}
			}

			// Create user type for raw object results
			if _, ok := e.Result.Type.(design.Object); ok {
				e.Result.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Result),
					TypeName:      fmt.Sprintf("%sResult", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Result.Type.(design.UserType); ok {
				seen[ut.Name()] = struct{}{}
			}
		}
		for _, e := range service.Endpoints {
			patt := e.Payload
			if ut, ok := patt.Type.(design.UserType); ok {
				patt = ut.Attribute()
			}
			types = append(types, collectTypes(patt, seen, scope)...)
			ratt := e.Result
			if ut, ok := ratt.Type.(design.UserType); ok {
				ratt = ut.Attribute()
			}
			types = append(types, collectTypes(ratt, seen, scope)...)
		}
	}

	var (
		methods []*ServiceMethodData
	)
	{
		methods = make([]*ServiceMethodData, len(service.Endpoints))
		for i, e := range service.Endpoints {
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

// collectTypes recurses through the attribute to gather all user types and
// records them in userTypes.
func collectTypes(at *design.AttributeExpr, seen map[string]struct{}, scope *codegen.NameScope) (data []*UserTypeData) {
	if at == nil || at.Type == design.Empty {
		return
	}
	collect := func(at *design.AttributeExpr) []*UserTypeData { return collectTypes(at, seen, scope) }
	switch dt := at.Type.(type) {
	case design.UserType:
		if _, ok := seen[dt.Name()]; ok {
			return nil
		}
		data = append(data, &UserTypeData{
			Name:        scope.GoTypeName(dt),
			Description: dt.Attribute().Description,
			TypeDef:     scope.GoTypeDef(dt.Attribute()),
		})
		seen[dt.Name()] = struct{}{}
		data = append(data, collect(dt.Attribute())...)
	case design.Object:
		for _, catt := range dt {
			data = append(data, collect(catt)...)
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
func buildServiceMethodData(m *design.EndpointExpr, scope *codegen.NameScope) *ServiceMethodData {
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
				payloadDef = scope.GoTypeDef(dt.Attribute())
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
				resultDef = scope.GoTypeDef(dt.Attribute())
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
