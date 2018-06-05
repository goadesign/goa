package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// Services holds the data computed from the design needed to generate the code
// of the services.
var Services = make(ServicesData)

var (
	// initTypeTmpl is the template used to render the code that initializes a
	// projected type or viewed result type or a result type.
	initTypeCodeTmpl = template.Must(template.New("initTypeCode").Funcs(template.FuncMap{"goify": codegen.Goify}).Parse(initTypeCodeT))
	// validateTypeCodeTmpl is the template used to render the code to
	// validate a projected type or a viewed result type.
	validateTypeCodeTmpl = template.Must(template.New("validateType").Funcs(template.FuncMap{"goify": codegen.Goify}).Parse(validateTypeT))
)

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
		// ViewsPkg is the name of the package containing the view types.
		ViewsPkg string
		// Methods lists the service interface methods.
		Methods []*MethodData
		// Schemes is the list of security schemes required by the
		// service methods.
		Schemes []*SchemeData
		// UserTypes lists the type definitions that the service
		// depends on.
		UserTypes []*UserTypeData
		// ErrorTypes lists the error type definitions that the service
		// depends on.
		ErrorTypes []*UserTypeData
		// Errors list the information required to generate error init
		// functions.
		ErrorInits []*ErrorInitData
		// ProjectedTypes lists the types which uses pointers for all fields to
		// define view specific validation logic.
		ProjectedTypes []*ProjectedTypeData
		// ViewedResultTypes lists all the viewed method result types.
		ViewedResultTypes []*ViewedResultTypeData
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
		// Temporary indicates whether the error is temporary.
		Temporary bool
		// Timeout indicates whether the error is due to timeouts.
		Timeout bool
		// Fault indicates whether the error is server-side fault.
		Fault bool
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
		// ViewedResult contains the data required to generate the code handling
		// views if any.
		ViewedResult *ViewedResultTypeData
		// ServerStream indicates that the service method receives a payload
		// stream or sends a result stream or both.
		ServerStream *StreamData
		// ClientStream indicates that the service method receives a result
		// stream or sends a payload result or both.
		ClientStream *StreamData
	}

	// StreamData is the data used to generate client and server interfaces that
	// a streaming endpoint implements. It is initialized if a method defines a
	// streaming payload or result or both.
	StreamData struct {
		// Interface is the name of the stream interface.
		Interface string
		// VarName is the name of the struct type that implements the stream
		// interface.
		VarName string
		// SendName is the type name sent through the stream.
		SendName string
		// SendRef is the reference to the type sent through the stream.
		SendRef string
		// RecvName is the type name received from the stream.
		RecvName string
		// RecvRef is the reference to the type received from the stream.
		RecvRef string
		// EndpointStruct is the name of the endpoint struct that holds a payload
		// reference (if any) and the endpoint server stream. It is set only if the
		// client sends a normal payload and server streams a result.
		EndpointStruct string
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

	// SchemeData describes a single security scheme.
	SchemeData struct {
		// Kind is the type of scheme, one of "Basic", "APIKey", "JWT"
		// or "OAuth2".
		Type string
		// SchemeName is the name of the scheme.
		SchemeName string
		// Name refers to a header or parameter name, based on In's
		// value.
		Name string
		// UsernameField is the name of the payload field that should be
		// initialized with the basic auth username if any.
		UsernameField string
		// UsernamePointer is true if the username field is a pointer.
		UsernamePointer bool
		// UsernameAttr is the name of the attribute that contains the
		// username.
		UsernameAttr string
		// UsernameRequired specifies whether the attribute that
		// contains the username is required.
		UsernameRequired bool
		// PasswordField is the name of the payload field that should be
		// initialized with the basic auth password if any.
		PasswordField string
		// PasswordPointer is true if the password field is a pointer.
		PasswordPointer bool
		// PasswordAttr is the name of the attribute that contains the
		// password.
		PasswordAttr string
		// PasswordRequired specifies whether the attribute that
		// contains the password is required.
		PasswordRequired bool
		// CredField contains the name of the payload field that should
		// be initialized with the API key, the JWT token or the OAuth2
		// access token.
		CredField string
		// CredPointer is true if the credential field is a pointer.
		CredPointer bool
		// CredRequired specifies if the key is a required attribute.
		CredRequired bool
		// KeyAttr is the name of the attribute that contains
		// the security tag (for APIKey, OAuth2, and JWT schemes).
		KeyAttr string
		// Scopes lists the scopes that apply to the scheme.
		Scopes []string
		// Flows describes the OAuth2 flows.
		Flows []*design.FlowExpr
		// In indicates the request element that holds the credential.
		In string
	}

	// ViewedResultTypeData contains the data used to generate a viewed result type
	// (i.e. a method result type with more than one view). The viewed result
	// type holds the projected type and a view based on which it creates the
	// projected type. It also contains the code to validate the viewed result
	// type and the functions to initialize a viewed result type from a result
	// type and vice versa.
	ViewedResultTypeData struct {
		// the viewed result type
		*UserTypeData
		// Views lists the views defined on the viewed result type.
		Views []*ViewData
		// Validate is the validation run on the viewed result type.
		Validate *ValidateData
		// Init is the constructor code to initialize a viewed result type from
		// a result type.
		Init *InitData
		// ResultInit is the constructor code to initialize a result type
		// from the viewed result type.
		ResultInit *InitData
		// FullName is the fully qualified name of the viewed result type.
		FullName string
		// FullRef is the complete reference to the viewed result type
		// (including views package name).
		FullRef string
		// IsCollection indicates whether the viewed result type is a collection.
		IsCollection bool
		// ViewName is the view name to use to render the result type. It is set
		// only if the result type has at most one view.
		ViewName string
		// ViewsPkg is the views package name.
		ViewsPkg string
	}

	// ViewData contains data about a result type view.
	ViewData struct {
		// Name is the view name.
		Name string
		// Description is the view description.
		Description string
	}

	// ProjectedTypeData contains the data used to generate a projected type for
	// the corresponding user type or result type in the service package. The
	// generated type uses pointers for all fields. It also contains the data
	// to generate view-based validation logic and transformation functions to
	// convert a projected type to its corresponding service type and vice versa.
	ProjectedTypeData struct {
		// the projected type
		*UserTypeData
		// Validations lists the validation functions to run on the projected type.
		// If the projected type corresponds to a result type then a validation
		// function for each view is generated. For user types, only one validation
		// function is generated.
		Validations []*ValidateData
		// Projections contains the code to create a projected type based on
		// views. If the projected type corresponds to a result type, then a
		// function for each view is generated.
		Projections []*InitData
		// TypeInits contains the code to convert a projected type to its
		// corresponding service type. If the projected type corresponds to a
		// result type, then a function for each view is generated.
		TypeInits []*InitData
		// ViewsPkg is the views package name.
		ViewsPkg string
	}

	// InitData contains the data to render a constructor.
	InitData struct {
		// Name is the name of the constructor function.
		Name string
		// Description is the function description.
		Description string
		// Args lists arguments to this function.
		Args []*InitArgData
		// ReturnTypeRef is the reference to the return type.
		ReturnTypeRef string
		// Code is the transformation code.
		Code string
		// Helpers contain the helpers used in the transformation code.
		Helpers []*codegen.TransformFunctionData
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Ref is the reference to the argument type.
		Ref string
	}

	// ValidateData contains data to render a validate function.
	ValidateData struct {
		// Name is the validation function name.
		Name string
		// Ref is the reference to the type on which the validation function
		// is defined.
		Ref string
		// Description is the description for the validation function.
		Description string
		// Validate is the validation code.
		Validate string
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
		viewspkg   string
		types      []*UserTypeData
		errTypes   []*UserTypeData
		errorInits []*ErrorInitData
		projTypes  []*ProjectedTypeData
		viewedRTs  []*ViewedResultTypeData
		seenErrors map[string]struct{}
		seen       map[string]struct{}
		seenProj   map[string]*ProjectedTypeData
		seenViewed map[string]*ViewedResultTypeData
	)
	{
		scope = codegen.NewNameScope()
		pkgName = scope.HashedUnique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
		viewspkg = pkgName + "views"
		seen = make(map[string]struct{})
		seenErrors = make(map[string]struct{})
		seenProj = make(map[string]*ProjectedTypeData)
		seenViewed = make(map[string]*ViewedResultTypeData)
		for _, e := range service.Methods {
			// Create user type for raw object payloads
			if _, ok := e.Payload.Type.(*design.Object); ok {
				e.Payload.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Payload),
					TypeName:      fmt.Sprintf("%sPayload", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Payload.Type.(design.UserType); ok {
				seen[ut.ID()] = struct{}{}
			}

			// Create user type for raw object results
			if _, ok := e.Result.Type.(*design.Object); ok {
				e.Result.Type = &design.UserTypeExpr{
					AttributeExpr: design.DupAtt(e.Result),
					TypeName:      fmt.Sprintf("%sResult", codegen.Goify(e.Name, true)),
				}
			}

			if ut, ok := e.Result.Type.(design.UserType); ok {
				seen[ut.ID()] = struct{}{}
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
			// collect inner user types
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
			// collect projected types
			if _, ok := m.Result.Type.(*design.ResultTypeExpr); ok {
				projected := design.DupAtt(m.Result)
				projTypes = append(projTypes, collectProjectedTypes(projected, m.Result, seenProj, scope, viewspkg)...)
			}
			for _, er := range m.Errors {
				recordError(er)
			}
		}
	}

	for _, t := range design.Root.Types {
		if svcs, ok := t.Attribute().Metadata["type:generate:force"]; ok {
			att := &design.AttributeExpr{Type: t}
			if len(svcs) > 0 {
				// Force generate type only in the specified services
				for _, svc := range svcs {
					if svc == service.Name {
						types = append(types, collectTypes(att, seen, scope)...)
						break
					}
				}
			} else {
				// Force generate type in all the services
				types = append(types, collectTypes(att, seen, scope)...)
			}
		}
	}

	var (
		methods []*MethodData
		schemes []*SchemeData
	)
	{
		methods = make([]*MethodData, len(service.Methods))
		for i, e := range service.Methods {
			m := buildMethodData(e, pkgName, scope)
			if rt, ok := e.Result.Type.(*design.ResultTypeExpr); ok {
				if vrt, ok := seenViewed[m.Result]; ok {
					m.ViewedResult = vrt
				} else {
					projected := seenProj[rt.ID()]
					vrt := buildViewedResultType(e.Result, projected.Type, scope, viewspkg)
					viewedRTs = append(viewedRTs, vrt)
					seenViewed[vrt.Name] = vrt
					m.ViewedResult = vrt
				}
			}
			methods[i] = m
			for _, r := range m.Requirements {
				for _, s := range r.Schemes {
					found := false
					for _, s2 := range schemes {
						if s.SchemeName == s2.SchemeName {
							found = true
							break
						}
					}
					if !found {
						schemes = append(schemes, s)
					}
				}
			}
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
		Name:              service.Name,
		Description:       desc,
		VarName:           codegen.Goify(service.Name, false),
		StructName:        codegen.Goify(service.Name, true),
		PkgName:           pkgName,
		ViewsPkg:          viewspkg,
		Methods:           methods,
		Schemes:           schemes,
		UserTypes:         types,
		ErrorTypes:        errTypes,
		ErrorInits:        errorInits,
		ProjectedTypes:    projTypes,
		ViewedResultTypes: viewedRTs,
		Scope:             scope,
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
		if _, ok := seen[dt.ID()]; ok {
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
		seen[dt.ID()] = struct{}{}
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
	_, fault := er.AttributeExpr.Metadata["goa:error:fault"]
	return &ErrorInitData{
		Name:        fmt.Sprintf("Make%s", codegen.Goify(er.Name, true)),
		Description: er.Description,
		ErrName:     er.Name,
		TypeName:    scope.GoTypeName(er.AttributeExpr),
		TypeRef:     scope.GoTypeRef(er.AttributeExpr),
		Temporary:   temporary,
		Timeout:     timeout,
		Fault:       fault,
	}
}

// buildMethodData creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func buildMethodData(m *design.MethodExpr, svcPkgName string, scope *codegen.NameScope) *MethodData {
	var (
		vname       string
		desc        string
		payloadName string
		payloadDef  string
		payloadRef  string
		payloadDesc string
		payloadEx   interface{}
		rname       string
		resultDef   string
		resultRef   string
		resultDesc  string
		resultEx    interface{}
		errors      []*ErrorInitData
		reqs        []*RequirementData
		schemes     []string
		svrStream   *StreamData
		cliStream   *StreamData
	)
	vname = codegen.Goify(m.Name, true)
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
		rname = scope.GoTypeName(m.Result)
		resultRef = scope.GoTypeRef(m.Result)
		if dt, ok := m.Result.Type.(design.UserType); ok {
			resultDef = scope.GoTypeDef(dt.Attribute(), true)
		}
		resultDesc = m.Result.Description
		if resultDesc == "" {
			resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
				rname, m.Service.Name, m.Name)
		}
		resultEx = m.Result.Example(design.Root.API.Random())
	}
	if len(m.Errors) > 0 {
		errors = make([]*ErrorInitData, len(m.Errors))
		for i, er := range m.Errors {
			errors[i] = buildErrorInitData(er, scope)
		}
	}
	if m.Stream != design.NoStreamKind {
		svrStream = &StreamData{
			Interface: vname + "ServerStream",
			VarName:   m.Name + "ServerStream",
		}
		cliStream = &StreamData{
			Interface: vname + "ClientStream",
			VarName:   m.Name + "ClientStream",
		}
		switch m.Stream {
		case design.ServerStreamKind:
			svrStream.SendName = rname
			svrStream.SendRef = resultRef
			cliStream.RecvName = rname
			cliStream.RecvRef = resultRef
			svrStream.EndpointStruct = vname + "EndpointInput"
		case design.ClientStreamKind:
			svrStream.RecvName = payloadName
			svrStream.RecvRef = payloadRef
			cliStream.SendName = payloadName
			cliStream.SendRef = payloadRef
		case design.BidirectionalStreamKind:
			svrStream.SendName = rname
			svrStream.SendRef = resultRef
			svrStream.RecvName = payloadName
			svrStream.RecvRef = payloadRef
			cliStream.SendName = payloadName
			cliStream.SendRef = payloadRef
			cliStream.RecvName = rname
			cliStream.RecvRef = resultRef
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
		VarName:      vname,
		Description:  desc,
		Payload:      payloadName,
		PayloadDef:   payloadDef,
		PayloadRef:   payloadRef,
		PayloadDesc:  payloadDesc,
		PayloadEx:    payloadEx,
		Result:       rname,
		ResultDef:    resultDef,
		ResultRef:    resultRef,
		ResultDesc:   resultDesc,
		ResultEx:     resultEx,
		Errors:       errors,
		Requirements: reqs,
		Schemes:      schemes,
		ServerStream: svrStream,
		ClientStream: cliStream,
	}
}

// buildSchemeData builds the scheme data for the given scheme and method expressions.
func buildSchemeData(s *design.SchemeExpr, m *design.MethodExpr) *SchemeData {
	if !design.IsObject(m.Payload.Type) {
		return nil
	}
	switch s.Kind {
	case design.BasicAuthKind:
		userAtt := design.TaggedAttribute(m.Payload, "security:username")
		user := codegen.Goify(userAtt, true)
		passAtt := design.TaggedAttribute(m.Payload, "security:password")
		pass := codegen.Goify(passAtt, true)
		return &SchemeData{
			Type:             s.Kind.String(),
			SchemeName:       s.SchemeName,
			UsernameAttr:     userAtt,
			UsernameField:    user,
			UsernamePointer:  m.Payload.IsPrimitivePointer(userAtt, true),
			UsernameRequired: m.Payload.IsRequired(userAtt),
			PasswordAttr:     passAtt,
			PasswordField:    pass,
			PasswordPointer:  m.Payload.IsPrimitivePointer(passAtt, true),
			PasswordRequired: m.Payload.IsRequired(passAtt),
		}
	case design.APIKeyKind:
		if keyAtt := design.TaggedAttribute(m.Payload, "security:apikey:"+s.SchemeName); keyAtt != "" {
			key := codegen.Goify(keyAtt, true)
			return &SchemeData{
				Type:         s.Kind.String(),
				Name:         s.Name,
				SchemeName:   s.SchemeName,
				CredField:    key,
				CredPointer:  m.Payload.IsPrimitivePointer(keyAtt, true),
				CredRequired: m.Payload.IsRequired(keyAtt),
				KeyAttr:      keyAtt,
				In:           s.In,
			}
		}
	case design.JWTKind:
		if keyAtt := design.TaggedAttribute(m.Payload, "security:token"); keyAtt != "" {
			key := codegen.Goify(keyAtt, true)
			var scopes []string
			if len(s.Scopes) > 0 {
				scopes = make([]string, len(s.Scopes))
				for i, s := range s.Scopes {
					scopes[i] = s.Name
				}
			}
			return &SchemeData{
				Type:         s.Kind.String(),
				Name:         s.Name,
				SchemeName:   s.SchemeName,
				CredField:    key,
				CredPointer:  m.Payload.IsPrimitivePointer(keyAtt, true),
				CredRequired: m.Payload.IsRequired(keyAtt),
				KeyAttr:      keyAtt,
				Scopes:       scopes,
				In:           s.In,
			}
		}
	case design.OAuth2Kind:
		if keyAtt := design.TaggedAttribute(m.Payload, "security:accesstoken"); keyAtt != "" {
			key := codegen.Goify(keyAtt, true)
			var scopes []string
			if len(s.Scopes) > 0 {
				scopes = make([]string, len(s.Scopes))
				for i, s := range s.Scopes {
					scopes[i] = s.Name
				}
			}
			return &SchemeData{
				Type:         s.Kind.String(),
				Name:         s.Name,
				SchemeName:   s.SchemeName,
				CredField:    key,
				CredPointer:  m.Payload.IsPrimitivePointer(keyAtt, true),
				CredRequired: m.Payload.IsRequired(keyAtt),
				KeyAttr:      keyAtt,
				Scopes:       scopes,
				Flows:        s.Flows,
				In:           s.In,
			}
		}
	}
	return nil
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

// collectProjectedTypes builds a projected type for every user type found
// when recursing through the attributes. It stores the projected types in
// data.
func collectProjectedTypes(projected, att *design.AttributeExpr, seen map[string]*ProjectedTypeData, scope *codegen.NameScope, viewspkg string) (data []*ProjectedTypeData) {
	collect := func(projected, att *design.AttributeExpr) []*ProjectedTypeData {
		return collectProjectedTypes(projected, att, seen, scope, viewspkg)
	}
	switch pt := projected.Type.(type) {
	case design.UserType:
		dt := att.Type.(design.UserType)
		if pd, ok := seen[dt.ID()]; ok {
			// a projected type is already created for this user type. We change the
			// attribute type to this seen projected type. The seen projected type
			// can be nil if the attribute type has a ciruclar type definition in
			// which case we don't change the attribute type until the projected type
			// is created during the recursion.
			if pd != nil {
				projected.Type = pd.Type
			}
			return
		}
		seen[dt.ID()] = nil
		pt.Rename(pt.Name() + "View")
		// We recurse before building the projected type so that user types within
		// a projected type is also converted to their respective projected types.
		types := collect(pt.Attribute(), dt.Attribute())
		pd := buildProjectedType(projected, att, scope, viewspkg)
		seen[dt.ID()] = pd
		// Change the attribute type to the newly created projected type.
		projected.Type = pd.Type
		data = append(data, pd)
		data = append(data, types...)
	case *design.Array:
		dt := att.Type.(*design.Array)
		data = append(data, collect(pt.ElemType, dt.ElemType)...)
	case *design.Map:
		dt := att.Type.(*design.Map)
		data = append(data, collect(pt.KeyType, dt.KeyType)...)
		data = append(data, collect(pt.ElemType, dt.ElemType)...)
	case *design.Object:
		dt := att.Type.(*design.Object)
		for _, n := range *pt {
			data = append(data, collect(n.Attribute, dt.Attribute(n.Name))...)
		}
	case design.Primitive:
		projected.ForcePointer = true
	}
	return
}

// buildProjectedType builds projected type for the given attribute of type
// user type.
func buildProjectedType(projected, att *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) *ProjectedTypeData {
	var (
		projections []*InitData
		typeInits   []*InitData
	)
	pt := projected.Type.(design.UserType)
	if _, isrt := pt.(*design.ResultTypeExpr); isrt {
		typeInits = buildTypeInits(projected, att, scope, viewspkg)
		projections = buildProjections(projected, att, scope, viewspkg)
	}
	varname := scope.GoTypeName(projected)
	return &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Name:        varname,
			Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
			VarName:     varname,
			Def:         scope.GoTypeDef(pt.Attribute(), true),
			Ref:         scope.GoTypeRef(projected),
			Type:        pt,
		},
		Projections: projections,
		TypeInits:   typeInits,
		Validations: buildValidations(projected, scope),
		ViewsPkg:    viewspkg,
	}
}

// buildViewedResultType builds a viewed result type from the given result type
// and projected type.
func buildViewedResultType(att *design.AttributeExpr, projected design.UserType, scope *codegen.NameScope, viewspkg string) *ViewedResultTypeData {
	rt := att.Type.(*design.ResultTypeExpr)
	var (
		views    []*ViewData
		resinit  *InitData
		init     *InitData
		validate *ValidateData
		data     map[string]interface{}
		buf      bytes.Buffer
		viewName string

		resvar  = scope.GoTypeName(att)
		resref  = scope.GoTypeRef(att)
		isarr   = design.IsArray(att.Type)
		vresref = scope.GoFullTypeRef(att, viewspkg)
	)
	if !rt.HasMultipleViews() {
		viewName = design.DefaultView
	}
	if v, ok := att.Metadata["view"]; ok && len(v) > 0 {
		viewName = v[0]
	}

	// collect result type views
	views = make([]*ViewData, 0, len(rt.Views))
	for _, view := range rt.Views {
		views = append(views, &ViewData{Name: view.Name, Description: view.Description})
	}

	// build validation data
	data = map[string]interface{}{
		"ArgVar":   "result",
		"Source":   "result",
		"Views":    views,
		"IsViewed": true,
	}
	if err := validateTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	name := "Validate"
	validate = &ValidateData{
		Name:        name,
		Description: fmt.Sprintf("%s runs the validations defined on the viewed result type %s.", name, resvar),
		Ref:         resref,
		Validate:    buf.String(),
	}
	buf.Reset()

	// build constructor to initialize viewed result type from result type
	data = map[string]interface{}{
		"ToViewed":      true,
		"ArgVar":        "res",
		"ReturnVar":     "vres",
		"Views":         views,
		"ReturnTypeRef": vresref,
		"IsCollection":  isarr,
		"TargetType":    scope.GoFullTypeName(att, viewspkg),
		"InitName":      "new" + scope.GoTypeName(&design.AttributeExpr{Type: projected}),
	}
	if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	name = "NewViewed" + resvar
	init = &InitData{
		Name:        name,
		Description: fmt.Sprintf("%s initializes viewed result type %s from result type %s using the given view.", name, resvar, resvar),
		Args: []*InitArgData{
			{Name: "res", Ref: scope.GoTypeRef(att)},
			{Name: "view", Ref: "string"},
		},
		ReturnTypeRef: vresref,
		Code:          buf.String(),
	}
	buf.Reset()

	// build constructor to initialize result type from viewed result type
	data = map[string]interface{}{
		"ToResult":      true,
		"ArgVar":        "vres",
		"ReturnVar":     "res",
		"Views":         views,
		"ReturnTypeRef": resref,
		"InitName":      "new" + scope.GoTypeName(att),
	}
	if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	name = "New" + resvar
	resinit = &InitData{
		Name:          name,
		Description:   fmt.Sprintf("%s initializes result type %s from viewed result type %s.", name, resvar, resvar),
		Args:          []*InitArgData{{Name: "vres", Ref: scope.GoFullTypeRef(att, viewspkg)}},
		ReturnTypeRef: resref,
		Code:          buf.String(),
	}
	buf.Reset()

	projected = wrapProjected(projected)
	return &ViewedResultTypeData{
		UserTypeData: &UserTypeData{
			Name:        resvar,
			Description: fmt.Sprintf("%s is the viewed result type that is projected based on a view.", resvar),
			VarName:     resvar,
			Def:         scope.GoTypeDef(projected.Attribute(), true),
			Ref:         resref,
			Type:        projected,
		},
		FullName:     scope.GoFullTypeName(att, viewspkg),
		FullRef:      vresref,
		ResultInit:   resinit,
		Init:         init,
		Views:        views,
		Validate:     validate,
		IsCollection: isarr,
		ViewName:     viewName,
		ViewsPkg:     viewspkg,
	}
}

// wrapProjected builds a viewed result type by wrapping the given projected
// in a result type with "projected" and "view" attributes.
func wrapProjected(projected design.UserType) design.UserType {
	rt := projected.(*design.ResultTypeExpr)
	pratt := &design.NamedAttributeExpr{
		Name:      "projected",
		Attribute: &design.AttributeExpr{Type: rt, Description: "Type to project"},
	}
	prview := &design.NamedAttributeExpr{
		Name:      "view",
		Attribute: &design.AttributeExpr{Type: design.String, Description: "View to render"},
	}
	return &design.ResultTypeExpr{
		UserTypeExpr: &design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{
				Type:       &design.Object{pratt, prview},
				Validation: &design.ValidationExpr{Required: []string{"projected", "view"}},
			},
			TypeName: rt.TypeName,
		},
		Identifier: rt.Identifier,
		Views:      rt.Views,
	}
}

// buildTypeInits builds the data to generate the constructor code to
// initialize a service result type from a projected type.
func buildTypeInits(projected, att *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) []*InitData {
	var (
		init []*InitData

		resvar = scope.GoTypeName(att)
	)
	prt := projected.Type.(*design.ResultTypeExpr)
	pobj := design.AsObject(projected.Type)
	parr := design.AsArray(projected.Type)
	if parr != nil {
		pobj = design.AsObject(parr.ElemType.Type)
	}
	init = make([]*InitData, 0, len(prt.Views))
	for _, view := range prt.Views {
		var typ design.DataType
		obj := &design.Object{}
		// Select only the attributes from the given view
		for _, n := range *view.Type.(*design.Object) {
			if attr := pobj.Attribute(n.Name); attr != nil {
				obj.Set(n.Name, attr)
			}
		}
		typ = obj
		if parr != nil {
			typ = &design.Array{ElemType: &design.AttributeExpr{
				Type: &design.ResultTypeExpr{
					UserTypeExpr: &design.UserTypeExpr{
						AttributeExpr: &design.AttributeExpr{Type: obj},
						TypeName:      scope.GoTypeName(parr.ElemType),
					},
				},
			}}
		}
		src := &design.AttributeExpr{
			Type: &design.ResultTypeExpr{
				UserTypeExpr: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: typ},
					TypeName:      scope.GoTypeName(projected),
				},
				Views:      prt.Views,
				Identifier: prt.Identifier,
			},
		}
		code, helpers := buildConstructorCode(src, att, "vres", "res", viewspkg, "", view.Name, scope, true)
		name := "new" + resvar
		if view.Name != design.DefaultView {
			name += codegen.Goify(view.Name, true)
		}
		init = append(init, &InitData{
			Name:          name,
			Description:   fmt.Sprintf("%s converts projected type %s to service type %s.", name, resvar, resvar),
			Args:          []*InitArgData{{Name: "vres", Ref: scope.GoFullTypeRef(projected, viewspkg)}},
			ReturnTypeRef: scope.GoTypeRef(att),
			Code:          code,
			Helpers:       helpers,
		})
	}
	return init
}

// buildProjections builds the data to generate the constructor code to
// project a result type to a projected type based on a view.
func buildProjections(projected, att *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) []*InitData {
	var projections []*InitData
	rt := att.Type.(*design.ResultTypeExpr)
	projections = make([]*InitData, 0, len(rt.Views))
	for _, view := range rt.Views {
		var (
			typ design.DataType

			tname = scope.GoTypeName(projected)
		)
		obj := &design.Object{}
		pobj := design.AsObject(projected.Type)
		parr := design.AsArray(projected.Type)
		if parr != nil {
			pobj = design.AsObject(parr.ElemType.Type)
		}
		// Select only the attributes from the given view
		for _, n := range *view.Type.(*design.Object) {
			if attr := pobj.Attribute(n.Name); attr != nil {
				obj.Set(n.Name, attr)
			}
		}
		typ = obj
		if parr != nil {
			typ = &design.Array{ElemType: &design.AttributeExpr{
				Type: &design.ResultTypeExpr{
					UserTypeExpr: &design.UserTypeExpr{
						AttributeExpr: &design.AttributeExpr{Type: obj},
						TypeName:      parr.ElemType.Type.Name(),
					},
				},
			}}
		}
		tgt := &design.AttributeExpr{
			Type: &design.ResultTypeExpr{
				UserTypeExpr: &design.UserTypeExpr{
					AttributeExpr: &design.AttributeExpr{Type: typ},
					TypeName:      projected.Type.Name(),
				},
				Views:      rt.Views,
				Identifier: rt.Identifier,
			},
		}
		code, helpers := buildConstructorCode(att, tgt, "res", "vres", "", viewspkg, view.Name, scope, false)
		name := "new" + tname
		if view.Name != design.DefaultView {
			name += codegen.Goify(view.Name, true)
		}
		projections = append(projections, &InitData{
			Name:          name,
			Description:   fmt.Sprintf("%s projects result type %s into projected type %s using the %q view.", name, scope.GoTypeName(att), tname, view.Name),
			Args:          []*InitArgData{{Name: "res", Ref: scope.GoTypeRef(att)}},
			ReturnTypeRef: scope.GoFullTypeRef(projected, viewspkg),
			Code:          code,
			Helpers:       helpers,
		})
	}
	return projections
}

// buildValidationData builds the data required to generate validations for the
// projected types.
func buildValidations(projected *design.AttributeExpr, scope *codegen.NameScope) []*ValidateData {
	var (
		validations []*ValidateData

		ref   = scope.GoTypeRef(projected)
		tname = scope.GoTypeName(projected)
	)
	if rt, isrt := projected.Type.(*design.ResultTypeExpr); isrt {
		// for result types we create a validation function containing view
		// specific validation logic for each view
		isarr := design.IsArray(projected.Type)
		for _, view := range rt.Views {
			var (
				data map[string]interface{}
				buf  bytes.Buffer
			)
			data = map[string]interface{}{
				"ArgVar":       "result",
				"Source":       "result",
				"IsCollection": isarr,
			}
			name := "Validate"
			if view.Name != "default" {
				name += codegen.Goify(view.Name, true)
			}
			if isarr {
				// dealing with an array type
				data["Source"] = "item"
				data["ValidateVar"] = name
			} else {
				o := &design.Object{}
				sobj := design.AsObject(projected.Type)
				fields := []map[string]interface{}{}
				// select only the attributes from the given view expression
				for _, n := range *view.Type.(*design.Object) {
					if attr := sobj.Attribute(n.Name); attr != nil {
						if rt2, ok := attr.Type.(*design.ResultTypeExpr); ok {
							// use explicitly specified view (if any) for the attribute,
							// otherwise use default
							vw := ""
							if v, ok := n.Attribute.Metadata["view"]; ok && len(v) > 0 && v[0] != design.DefaultView {
								vw = v[0]
							}
							fields = append(fields, map[string]interface{}{
								"Name":        n.Name,
								"ValidateVar": "Validate" + codegen.Goify(vw, true),
								"IsRequired":  rt2.Attribute().IsRequired(n.Name),
							})
						} else {
							o.Set(n.Name, attr)
						}
					}
				}
				data["Validate"] = codegen.RecursiveValidationCode(&design.AttributeExpr{Type: o, Validation: rt.Validation}, false, true, false, "result")
				data["Fields"] = fields
			}
			if err := validateTypeCodeTmpl.Execute(&buf, data); err != nil {
				panic(err) // bug
			}
			validations = append(validations, &ValidateData{
				Name:        name,
				Description: fmt.Sprintf("%s runs the validations defined on %s using the %q view.", name, tname, view.Name),
				Ref:         ref,
				Validate:    buf.String(),
			})
		}
	} else {
		// for a user type or a result type with single view, we generate only one validation
		// function containing the validation logic
		name := "Validate"
		validations = append(validations, &ValidateData{
			Name:        name,
			Description: fmt.Sprintf("%s runs the validations defined on %s.", name, tname),
			Ref:         ref,
			Validate:    codegen.RecursiveValidationCode(projected, false, true, false, "result"),
		})
	}
	return validations
}

// buildConstructorCode builds the transformation code to create a projected
// type from a service type and vice versa.
//
// src and tgt contains the projected type or a service type
//
// srcvar and tgtvar contains the variable name that holds the source and
// target data structures in the transformation code.
//
// srcpkg and tgtpkg contains the name of the package where the source and
// target types are defined.
//
// view parameter is used to generate the function name to call to initialize
// the items of a result type collection. This parameter is applicable only
// when the src/tgt is a result type collection.
//
// scope is used to compute the name of the user types when initializing fields
// that use them.
//
// unmarshal if true indicates that the code generated is to convert a
// projected type to a service type. If false, the code generated is to convert
// a service type to a projected type.
func buildConstructorCode(src, tgt *design.AttributeExpr, srcvar, tgtvar, srcpkg, tgtpkg string, view string, scope *codegen.NameScope, unmarshal bool) (string, []*codegen.TransformFunctionData) {
	var (
		code    string
		err     error
		helpers []*codegen.TransformFunctionData
	)
	rt := src.Type.(*design.ResultTypeExpr)
	arr := design.AsArray(tgt.Type)
	data := map[string]interface{}{
		"ArgVar":       srcvar,
		"ReturnVar":    tgtvar,
		"IsCollection": arr != nil,
		"TargetType":   scope.GoFullTypeName(tgt, tgtpkg),
	}
	if arr != nil {
		init := "new" + scope.GoTypeName(arr.ElemType)
		if view != "" && view != design.DefaultView {
			init += codegen.Goify(view, true)
		}
		data["InitName"] = init
	} else {
		// service type to projected type (or vice versa)
		trts := &design.Object{}
		t := design.DupAtt(tgt)
		tobj := design.AsObject(t.Type)
		for _, n := range *tobj {
			if _, ok := n.Attribute.Type.(*design.ResultTypeExpr); ok {
				trts.Set(n.Name, n.Attribute)
				tobj.Delete(n.Name)
			}
		}
		data["Source"] = srcvar
		data["Target"] = tgtvar
		// build code for target with no result types
		code, helpers, err = codegen.GoTypeTransform(src.Type, t.Type, srcvar, tgtvar, srcpkg, tgtpkg, unmarshal, scope)
		if err != nil {
			panic(err) // bug
		}
		data["Code"] = code
		if view != "" {
			data["InitName"] = tgtpkg + "." + scope.GoTypeName(src)
		}
		fields := make([]map[string]interface{}, 0, len(*trts))
		// iterate through the result types found in the target and add the
		// code to deal with result types
		for _, n := range *trts {
			finit := "new" + scope.GoTypeName(n.Attribute)
			if view != "" {
				v := ""
				if vatt := rt.View(view).AttributeExpr.Find(n.Name); vatt != nil {
					if attv, ok := vatt.Metadata["view"]; ok && len(attv) > 0 && attv[0] != design.DefaultView {
						// view is explicitly set for the result type on the attribute
						v = attv[0]
					}
				}
				finit += codegen.Goify(v, true)
			}
			fields = append(fields, map[string]interface{}{
				"VarName":   codegen.Goify(n.Name, true),
				"FieldInit": finit,
			})
		}
		data["Fields"] = fields
	}
	var buf bytes.Buffer
	if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), helpers
}

const (
	initTypeCodeT = `{{- if or .ToResult .ToViewed -}}
  var {{ .ReturnVar }} {{ .ReturnTypeRef }}
  switch {{ if .ToResult }}{{ .ArgVar }}.View{{ else }}view{{ end }} {
  {{- range .Views }}
  case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
    {{- if $.ToViewed }}
    p := {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }})
    {{ $.ReturnVar }} = {{ if not $.IsCollection }}&{{ end }}{{ $.TargetType }}{ p,  {{ printf "%q" .Name }} }
    {{- else }}
    {{ $.ReturnVar }} = {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }}.Projected)
    {{- end }}
  {{- end }}
  }
{{- else if .IsCollection -}}
{{ .ReturnVar }} := make({{ .TargetType }}, len({{ .ArgVar }}))
for i, n := range {{ .ArgVar }} {
  {{ .ReturnVar }}[i] = {{ .InitName }}(n)
}
{{- else -}}
{{ .Code }}
  {{- range .Fields }}
if {{ $.Source }}.{{ .VarName }} != nil {
  {{ $.Target }}.{{ .VarName }} = {{ .FieldInit }}({{ $.Source }}.{{ .VarName }})
}
  {{- end }}
{{- end }}
return {{ .ReturnVar }}`

	validateTypeT = `{{- if .IsViewed -}}
switch {{ .ArgVar }}.View {
	{{- range .Views }}
case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
	err = {{ $.ArgVar }}.Projected.Validate{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}()
	{{- end }}
default:
	err = goa.InvalidEnumValueError("view", {{ .Source }}.View, []interface{}{ {{ range .Views }}{{ printf "%q" .Name }}, {{ end }} })
}
{{- else -}}
	{{- if .IsCollection -}}
for _, {{ $.Source }} := range {{ $.ArgVar }} {
	if err2 := {{ $.Source }}.{{ .ValidateVar }}(); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
	{{- else -}}
	{{ .Validate }}
		{{- range .Fields -}}
			{{- if .IsRequired -}}
if {{ $.Source }}.{{ goify .Name true }} == nil {
	err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, {{ printf "%q" $.Source }}))
}
			{{- end }}
if {{ $.Source }}.{{ goify .Name true }} != nil {
	if err2 := {{ $.Source }}.{{ goify .Name true }}.{{ .ValidateVar }}(); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
		{{- end -}}
	{{- end -}}
{{- end -}}
`
)
