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
		// StructName is the service struct name.
		StructName string
		// VarName is the service variable name (first letter in
		// lowercase).
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
		// Schemes is the unique security schemes for the service.
		Schemes []*SchemeData
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
		// Temporary indicates whether the error is temporary.
		Temporary bool
		// Timeout indicates whether the error is due to timeouts.
		Timeout bool
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
		// Requirements contains the security requirements for the
		// method.
		Requirements []*RequirementData
		// Schemes contains the security schemes types used by the
		// method.
		Schemes []string
	}

	// RequirementData lists the schemes and scopes defined by a single
	// security requirement.
	RequirementData struct {
		// Schemes list the requirement schemes.
		Schemes []*SchemeData
		// Scopes list the required scopes.
		Scopes []string
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

	// SchemeData describes a single security scheme.
	SchemeData struct {
		// Kind is the type of scheme, one of "Basic", "APIKey", "JWT"
		// or "OAuth2".
		Type string
		// Name is the name of the scheme.
		Name string
		// UsernameField is the name of the payload field that should be
		// initialized with the basic auth username if any.
		UsernameField string
		// UsernamePointer is true if the username field is a pointer.
		UsernamePointer bool
		// UsernameAttr is the attribute name containing the username.
		UsernameAttr string
		// PasswordField is the name of the payload field that should be
		// initialized with the basic auth password if any.
		PasswordField string
		// PasswordPointer is true if the password field is a pointer.
		PasswordPointer bool
		// PasswordAttr is the attribute name containing the password.
		PasswordAttr string
		// CredField contains the name of the payload field that should
		// be initialized with the API key, the JWT token or the OAuth2
		// access token.
		CredField string
		// CredPointer is true if the credential field is a pointer.
		CredPointer bool
		// KeyAttr is the nane of the attribute that contains the
		// security tag (for APIKey, OAuth2, and JWT schemes).
		KeyAttr string
		// Scopes lists the scopes that apply to the scheme.
		Scopes []string
		// Flows describes the OAuth2 flows.
		Flows []*design.FlowExpr
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
		pkgName    string
		types      []*UserTypeData
		errTypes   []*UserTypeData
		errorInits []*ErrorInitData
		seenErrors map[string]struct{}
		seen       map[string]struct{}
	)
	{
		scope = codegen.NewNameScope()
		pkgName = scope.Unique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
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
		recordError := func(er *design.ErrorExpr) {
			errTypes = append(errTypes, collectTypes(er.AttributeExpr, seen, scope)...)
			if er.Type == design.ErrorResult {
				if _, ok := seenErrors[er.Name]; ok {
					return
				}
				seenErrors[er.Name] = struct{}{}
				errorInits = append(errorInits, buildErrorInitData(er, scope))
			}
		}
		for _, er := range service.Errors {
			recordError(er)
		}
		for _, m := range service.Methods {
			patt := m.Payload
			if ut, ok := patt.Type.(design.UserType); ok {
				patt = ut.Attribute()
			}
			types = append(types, collectTypes(patt, seen, scope)...)
			ratt := m.Result
			if ut, ok := ratt.Type.(design.UserType); ok {
				ratt = ut.Attribute()
			}
			types = append(types, collectTypes(ratt, seen, scope)...)
			for _, er := range m.Errors {
				recordError(er)
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
		VarName:     codegen.Goify(service.Name, false),
		StructName:  codegen.Goify(service.Name, true),
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
			Name:        dt.Name(),
			VarName:     scope.GoTypeName(at),
			Description: dt.Attribute().Description,
			Def:         scope.GoTypeDef(dt.Attribute(), true),
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
	_, temporary := er.AttributeExpr.Metadata["goa:error:temporary"]
	_, timeout := er.AttributeExpr.Metadata["goa:error:timeout"]
	return &ErrorInitData{
		Name:        fmt.Sprintf("Make%s", codegen.Goify(er.Name, true)),
		Description: er.Description,
		ErrName:     er.Name,
		TypeName:    scope.GoTypeName(er.AttributeExpr),
		TypeRef:     scope.GoTypeRef(er.AttributeExpr),
		Temporary:   temporary,
		Timeout:     timeout,
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
		reqs        []*RequirementData
		schemes     []string
	)
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
	for _, req := range requirements(m) {
		var rs []*SchemeData
		for _, s := range req.Schemes {
			rs = append(rs, buildSchemeData(s, m))
			found := false
			for _, es := range schemes {
				if es == s.Kind.String() {
					found = true
					break
				}
			}
			if !found {
				schemes = append(schemes, s.Kind.String())
			}
		}
		reqs = append(reqs, &RequirementData{Schemes: rs, Scopes: req.Scopes})
	}

	return &MethodData{
		Name:         m.Name,
		VarName:      varName,
		Description:  desc,
		Payload:      payloadName,
		PayloadDef:   payloadDef,
		PayloadRef:   payloadRef,
		PayloadDesc:  payloadDesc,
		PayloadEx:    payloadEx,
		Result:       resultName,
		ResultDef:    resultDef,
		ResultRef:    resultRef,
		ResultDesc:   resultDesc,
		ResultEx:     resultEx,
		Errors:       errors,
		Requirements: reqs,
		Schemes:      schemes,
	}
}

// buildSchemeData builds the scheme data for the given scheme and method expressions.
func buildSchemeData(s *design.SchemeExpr, m *design.MethodExpr) *SchemeData {
	if !design.IsObject(m.Payload.Type) {
		return nil
	}
	switch s.Kind {
	case design.BasicAuthKind:
		userAtt, user := taggedAttribute(m.Payload, "security:username")
		passAtt, pass := taggedAttribute(m.Payload, "security:password")
		return &SchemeData{
			Type:            s.Kind.String(),
			Name:            s.SchemeName,
			UsernameAttr:    userAtt,
			UsernameField:   user,
			UsernamePointer: m.Payload.IsPrimitivePointer(userAtt, true),
			PasswordAttr:    passAtt,
			PasswordField:   pass,
			PasswordPointer: m.Payload.IsPrimitivePointer(passAtt, true),
		}
	case design.APIKeyKind:
		if keyAtt, key := taggedAttribute(m.Payload, "security:apikey:"+s.SchemeName); key != "" {
			return &SchemeData{
				Type:        s.Kind.String(),
				Name:        s.SchemeName,
				CredField:   key,
				CredPointer: m.Payload.IsPrimitivePointer(keyAtt, true),
				KeyAttr:     keyAtt,
			}
		}
	case design.JWTKind:
		if keyAtt, key := taggedAttribute(m.Payload, "security:token"); key != "" {
			var scopes []string
			if len(s.Scopes) > 0 {
				scopes = make([]string, len(s.Scopes))
				for i, s := range s.Scopes {
					scopes[i] = s.Name
				}
			}
			return &SchemeData{
				Type:        s.Kind.String(),
				Name:        s.SchemeName,
				CredField:   key,
				CredPointer: m.Payload.IsPrimitivePointer(keyAtt, true),
				KeyAttr:     keyAtt,
				Scopes:      scopes,
			}
		}
	case design.OAuth2Kind:
		if keyAtt, key := taggedAttribute(m.Payload, "security:accesstoken"); key != "" {
			var scopes []string
			if len(s.Scopes) > 0 {
				scopes = make([]string, len(s.Scopes))
				for i, s := range s.Scopes {
					scopes[i] = s.Name
				}
			}
			return &SchemeData{
				Type:        s.Kind.String(),
				Name:        s.SchemeName,
				CredField:   key,
				CredPointer: m.Payload.IsPrimitivePointer(keyAtt, true),
				KeyAttr:     keyAtt,
				Scopes:      scopes,
				Flows:       s.Flows,
			}
		}
	}
	return nil
}

// taggedAttribute returns the name and corresponding field name of child
// attribute of p with the given tag if p is an object.
func taggedAttribute(a *design.AttributeExpr, tag string) (string, string) {
	obj := design.AsObject(a.Type)
	if obj == nil {
		return "", ""
	}
	for _, at := range *obj {
		if _, ok := at.Attribute.Metadata[tag]; ok {
			return at.Name, codegen.Goify(at.Name, true)
		}
	}
	return "", ""
}

// requirements returns the security requirements for the given method.
func requirements(m *design.MethodExpr) []*design.SecurityExpr {
	for _, r := range m.Requirements {
		// Handle special case of no security
		for _, s := range r.Schemes {
			if s.Kind == design.NoKind {
				return nil
			}
		}
	}
	if len(m.Requirements) > 0 {
		return m.Requirements
	}
	if len(m.Service.Requirements) > 0 {
		return m.Service.Requirements
	}
	return design.Root.API.Requirements
}
