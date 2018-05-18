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
	// initResultTypeTmpl is the template used to render the code that
	// initializes a result type or viewed result type.
	initResultTypeCodeTmpl = template.Must(template.New("initResultCodeType").Parse(initResultTypeCodeT))
	// validateResultTypeCodeTmpl is the template used to render the code to
	// validate a viewed result type.
	validateResultTypeCodeTmpl = template.Must(template.New("validateResultType").Parse(validateResultTypeT))
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
		// ViewedResult contains the data required to generated the code handling
		// multiple views if any.
		ViewedResult *ProjectedTypeData
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

	// ProjectedTypeData contains the data used to generate a user type that can
	// be projected based on a view. The generated type uses pointers for all
	// fields so that view specific validation logic may be implemented.
	// A projected type is generated for every user type found in a method
	// result type having multiple views. If the user type is a result
	// type, then a viewed result type is generated which holds the projected
	// type and a view attribute holding the view name. Finally, the generated
	// code also defines functions that convert the result types to and from the
	// corresponding viewed result type as well as project the viewed result
	// type based on a view.
	ProjectedTypeData struct {
		// This holds the projected type or the viewed result type.
		*UserTypeData
		// Validate is the validation code run on the projected type.
		Validate string
		// MustValidate indicates whether to generate the validation code for the
		// projected type.
		MustValidate bool

		// fields set only for a viewed result type

		// Views lists the views defined on the result type and the functions
		// to project a viewed result type.
		Views []*ProjectData
		// FullRef is the complete reference to the viewed result type
		// (including views package name).
		FullRef string
		// IsCollection indicates whether the viewed result type is a collection.
		IsCollection bool
		// ConvertToResult is the code to convert a viewed result type to a
		// result type.
		ConvertToResult *InitData
		// ViewsPkg is the views package name.
		ViewsPkg string
	}

	// ProjectData contains data about projecting a result type based on
	// a view.
	ProjectData struct {
		// Name is the view name.
		Name string
		// Description is the view description.
		Description string
		// Validate is the validations run on the projected type based on a view.
		Validate string
		// Project is the code to project a result type based on a view.
		Project *InitData
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
		ReturnRef string
		// ReturnTypeName is the name of the struct to be returned. This is used
		// in tandem with ReturnIsStruct.
		ReturnTypeName string
		// ReturnIsStruct is true if the return type is a struct.
		ReturnIsStruct bool
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
		// FieldName is the name of the data structure field that should
		// be initialized with the argument if any.
		FieldName string
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
		seenErrors map[string]struct{}
		seen       map[string]struct{}
		seenProj   map[string]*ProjectedTypeData
	)
	{
		scope = codegen.NewNameScope()
		pkgName = scope.Unique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
		viewspkg = pkgName + "views"
		seen = make(map[string]struct{})
		seenErrors = make(map[string]struct{})
		seenProj = make(map[string]*ProjectedTypeData)
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
			if rt, ok := m.Result.Type.(*design.ResultTypeExpr); ok && rt.HasMultipleViews() {
				projected := dupAttNoRequired(m.Result)
				projTypes = append(projTypes, collectProjectedTypes(projected, m.Result, seenProj, scope, viewspkg)...)
			}
			for _, er := range m.Errors {
				recordError(er)
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
			if rt, ok := e.Result.Type.(*design.ResultTypeExpr); ok && rt.HasMultipleViews() {
				m.ViewedResult = seenProj[rt.TypeName]
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
		Name:           service.Name,
		Description:    desc,
		VarName:        codegen.Goify(service.Name, false),
		StructName:     codegen.Goify(service.Name, true),
		PkgName:        pkgName,
		ViewsPkg:       viewspkg,
		Methods:        methods,
		Schemes:        schemes,
		UserTypes:      types,
		ErrorTypes:     errTypes,
		ErrorInits:     errorInits,
		ProjectedTypes: projTypes,
		Scope:          scope,
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

// dupAttNoRequired creates a copy of the given attribute expression and
// removes the required attributes from the validation. This method is
// recursive so that the required attribute validation is removed from every
// attribute expression underneath.
func dupAttNoRequired(a *design.AttributeExpr, seen ...map[string]struct{}) *design.AttributeExpr {
	a = design.DupAtt(a)
	switch actual := a.Type.(type) {
	case design.UserType:
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		if _, ok := s[actual.Name()]; ok {
			return a
		}
		s[actual.Name()] = struct{}{}
		if actual.Attribute().Validation != nil {
			actual.Attribute().Validation.Required = []string{}
		}
		actual.SetAttribute(dupAttNoRequired(actual.Attribute(), seen...))
	case *design.Array:
		actual.ElemType = dupAttNoRequired(actual.ElemType, seen...)
	case *design.Map:
		actual.KeyType = dupAttNoRequired(actual.KeyType, seen...)
		actual.ElemType = dupAttNoRequired(actual.ElemType, seen...)
	case *design.Object:
		for _, nat := range *actual {
			nat.Attribute = dupAttNoRequired(nat.Attribute, seen...)
		}
	}
	return a
}

// collectProjectedTypes collects all the projected types from the given
// result type and stores them in data.
func collectProjectedTypes(projected, att *design.AttributeExpr, seen map[string]*ProjectedTypeData, scope *codegen.NameScope, viewspkg string) (data []*ProjectedTypeData) {
	collect := func(projected, att *design.AttributeExpr) []*ProjectedTypeData {
		return collectProjectedTypes(projected, att, seen, scope, viewspkg)
	}
	switch pt := projected.Type.(type) {
	case design.UserType:
		dt := att.Type.(design.UserType)
		if _, ok := seen[dt.Name()]; ok {
			return
		}
		if pd := buildProjectedType(projected, att, scope, viewspkg); pd != nil {
			data = append(data, pd)
			seen[pd.Name] = pd
		}
		if vr := buildViewedResultType(att, projected, scope, viewspkg); vr != nil {
			data = append(data, vr)
			seen[vr.Name] = vr
			projected.Type = vr.Type
		}
		data = append(data, collect(pt.Attribute(), dt.Attribute())...)
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
	}
	return
}

// buildProjectedType builds projected type for the given attribute of type
// user type or result type.
func buildProjectedType(projected, att *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) *ProjectedTypeData {
	var (
		validate string
	)
	pt := projected.Type.(design.UserType)
	rt, isrt := pt.(*design.ResultTypeExpr)
	switch {
	case isrt && design.IsArray(pt):
		return nil
	case !isrt || (isrt && !rt.HasMultipleViews()):
		validate = codegen.RecursiveValidationCode(att, false, true, false, "result")
	case isrt && rt.HasMultipleViews():
		pt.Rename(pt.Name() + "View")
	}
	varname := scope.GoTypeName(projected)
	return &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Name:        pt.Name(),
			Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
			VarName:     varname,
			Def:         scope.GoTypeDef(pt.Attribute(), false),
			Ref:         scope.GoTypeRef(projected),
			Type:        pt,
		},
		Validate:     validate,
		MustValidate: validate != "",
	}
}

// buildViewedResultType builds a viewed result type from the given result type
// and projected type.
func buildViewedResultType(att, projected *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) *ProjectedTypeData {
	rt, isrt := att.Type.(*design.ResultTypeExpr)
	if !isrt || (isrt && !rt.HasMultipleViews()) {
		return nil
	}
	var (
		ut           *UserTypeData
		views        []*ProjectData
		isCollection bool
	)
	resvar := scope.GoTypeName(att)
	desc := fmt.Sprintf("%s is the viewed result type that is projected based on a view.", resvar)
	isCollection = design.IsArray(projected.Type)
	if isCollection {
		prt, _ := projected.Type.(*design.ResultTypeExpr)
		ut = &UserTypeData{
			Name:        resvar,
			Description: desc,
			VarName:     resvar,
			Def:         scope.GoTypeDef(rt.Attribute(), true),
			Ref:         scope.GoTypeRef(projected),
			Type:        prt,
		}
	} else {
		vrt := &design.ResultTypeExpr{
			UserTypeExpr: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{
					Type: &design.Object{
						&design.NamedAttributeExpr{Name: "projected", Attribute: &design.AttributeExpr{Type: projected.Type, Description: "Type to project"}},
						&design.NamedAttributeExpr{Name: "view", Attribute: &design.AttributeExpr{Type: design.String, Description: "View to render"}},
					},
					Validation: &design.ValidationExpr{Required: []string{"projected", "view"}},
				},
				TypeName: rt.TypeName,
			},
			Identifier: rt.Identifier,
			Views:      rt.Views,
		}
		ut = &UserTypeData{
			Name:        resvar,
			Description: desc,
			VarName:     resvar,
			Def:         scope.GoTypeDef(vrt.Attribute(), true),
			Ref:         scope.GoTypeRef(&design.AttributeExpr{Type: vrt}),
			Type:        vrt,
		}
	}
	views = buildProjectData(att, projected, scope, viewspkg)
	data := map[string]interface{}{
		"IsCollection": isCollection,
		"Views":        views,
	}
	var (
		buf      bytes.Buffer
		validate string
	)
	if err := validateResultTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	validate = buf.String()
	return &ProjectedTypeData{
		UserTypeData:    ut,
		Validate:        validate,
		Views:           views,
		FullRef:         scope.GoFullTypeRef(att, viewspkg),
		ConvertToResult: convertToResult(projected, att, viewspkg, scope),
		ViewsPkg:        viewspkg,
		MustValidate:    validate != "",
		IsCollection:    isCollection,
	}
}

// buildProjectData builds the data to generate the constructor code to
// project a result type to a viewed result type and the validation code to
// validate the viewed result type based on a view.
func buildProjectData(att, projected *design.AttributeExpr, scope *codegen.NameScope, viewspkg string) []*ProjectData {
	rt := att.Type.(*design.ResultTypeExpr)
	var (
		defaultv *ProjectData
		views    []*ProjectData
	)
	views = make([]*ProjectData, 0, len(rt.Views))
	vobj := design.AsObject(projected.Type)
	for _, view := range rt.Views {
		var validate string
		if !design.IsArray(projected.Type) {
			obj := &design.Object{}
			for _, n := range *view.Type.(*design.Object) {
				if attr := vobj.Attribute(n.Name); attr != nil {
					obj.Set(n.Name, attr)
				}
			}
			valattr := &design.AttributeExpr{Type: obj}
			if rt.Validation != nil {
				valattr.Validation = rt.Validation.Dup()
			}
			validate = codegen.RecursiveValidationCode(valattr, false, true, false, "result.Projected")
		}
		pd := &ProjectData{
			Name:        view.Name,
			Description: view.Description,
			Project:     projectToView(att, projected, view, scope, viewspkg),
			Validate:    validate,
		}
		if view.Name == "default" {
			// if default view, append the ProjectData for this view outside the
			// loop so that it is the last item in the list. It is easier this
			// way to default the validation logic to the "default" view if the
			// "view" attribute is not set in the viewed result type.
			defaultv = pd
			continue
		}
		views = append(views, pd)
	}
	views = append(views, defaultv)
	return views
}

// convertToResult converts the given viewed result type to result type.
func convertToResult(projected, res *design.AttributeExpr, viewspkg string, scope *codegen.NameScope) *InitData {
	code, helpers := buildConstructorCode(projected, res, "vres", "res", viewspkg, "", "", scope)
	rname := scope.GoTypeName(res)
	name := "New" + rname
	return &InitData{
		Name:        name,
		Description: fmt.Sprintf("%s converts viewed result type %s to result type %s.", name, rname, rname),
		Args:        []*InitArgData{{Name: "vres", Ref: scope.GoFullTypeRef(res, viewspkg)}},
		ReturnRef:   scope.GoTypeRef(res),
		Code:        code,
		Helpers:     helpers,
	}
}

// projectToView builds the constructor function to project the given
// result type to the given viewed result type based on a view.
func projectToView(res, vres *design.AttributeExpr, view *design.ViewExpr, scope *codegen.NameScope, viewspkg string) *InitData {
	obj := &design.Object{}
	vresobj := design.AsObject(vres.Type)
	varr := design.AsArray(vres.Type)
	if varr != nil {
		vresobj = design.AsObject(varr.ElemType.Type)
	}
	// Select only the attributes from the given view
	for _, n := range *view.Type.(*design.Object) {
		if attr := vresobj.Attribute(n.Name); attr != nil {
			// Add any specific view metadata from view attribute
			if v, ok := n.Attribute.Metadata["view"]; ok {
				if attr.Metadata == nil {
					attr.Metadata = design.MetadataExpr{}
				}
				attr.Metadata["view"] = v
			}
			obj.Set(n.Name, attr)
		}
	}
	var typ design.DataType
	typ = obj
	if varr != nil {
		typ = &design.Array{ElemType: &design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{Type: obj},
				TypeName:      scope.GoTypeName(varr.ElemType),
			}}}
	}
	tgt := &design.AttributeExpr{
		Type: &design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{Type: typ},
			TypeName:      scope.GoTypeName(vres),
		},
	}
	code, helpers := buildConstructorCode(res, tgt, "res", "vres", "", viewspkg, view.Name, scope)
	rname := scope.GoTypeName(res)
	name := "New" + rname + codegen.Goify(view.Name, true)
	return &InitData{
		Name:        name,
		Description: fmt.Sprintf("%s projects result type %s into viewed result type %s using the %s view.", name, rname, rname, view.Name),
		Args:        []*InitArgData{{Name: "res", Ref: scope.GoTypeRef(res)}},
		ReturnRef:   scope.GoFullTypeRef(res, viewspkg),
		Code:        code,
		Helpers:     helpers,
	}
}

func buildConstructorCode(src, tgt *design.AttributeExpr, srcvar, tgtvar, srcpkg, tgtpkg string, view string, scope *codegen.NameScope) (string, []*codegen.TransformFunctionData) {
	var (
		code    string
		err     error
		helpers []*codegen.TransformFunctionData
	)
	data := map[string]interface{}{
		"ArgVar":    srcvar,
		"ReturnVar": tgtvar,
		"View":      view,
	}
	// t contains only the attributes of type other than result type from tgt
	t := design.DupAtt(tgt)
	if arr := design.AsArray(t.Type); arr != nil {
		data["IsCollection"] = true
		data["TargetType"] = scope.GoFullTypeRef(t, tgtpkg)
		data["InitName"] = "New" + scope.GoTypeName(arr.ElemType) + codegen.Goify(view, true)
	} else {
		// trts contains only the attributes of type result type from tgt
		trts := &design.Object{}
		tobj := design.AsObject(t.Type)
		for _, n := range *tobj {
			if _, ok := n.Attribute.Type.(*design.ResultTypeExpr); ok {
				trts.Set(n.Name, n.Attribute)
				tobj.Delete(n.Name)
			}
		}
		if view == "" {
			srcvar += ".Projected"
		}
		data["Source"] = srcvar
		data["Target"] = tgtvar
		code, helpers, err = codegen.GoTypeTransform(src.Type, t.Type, srcvar, tgtvar, srcpkg, tgtpkg, view == "", scope)
		if err != nil {
			panic(err) // bug
		}
		data["Code"] = code
		if view != "" {
			data["InitName"] = tgtpkg + "." + scope.GoTypeName(src)
		}
		fields := make([]map[string]interface{}, 0, len(*trts))
		for _, n := range *trts {
			init := "New" + scope.GoTypeName(n.Attribute)
			if view != "" {
				v := "default"
				if attv, ok := n.Attribute.Metadata["view"]; ok {
					// view is explicitly set for the result type on the attribute
					v = attv[0]
				}
				init += codegen.Goify(v, true)
			}
			fields = append(fields, map[string]interface{}{
				"VarName":   codegen.Goify(n.Name, true),
				"FieldInit": init,
			})
		}
		data["Fields"] = fields
	}
	var buf bytes.Buffer
	if err := initResultTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), helpers
}

const (
	initResultTypeCodeT = `{{- if .IsCollection }}
	{{- .ReturnVar }} := make({{ .TargetType }}, len({{ .ArgVar }}))
for i, n := range {{ .ArgVar }} {
	{{ .ReturnVar }}[i] = {{ .InitName }}(n)
}
{{- else }}
	{{- .Code }}
	{{- range .Fields }}
if {{ $.Source }}.{{ .VarName }} != nil {
	{{ $.Target }}.{{ .VarName }} = {{ .FieldInit }}({{ $.Source }}.{{ .VarName }})
}
	{{- end }}
{{- end }}
{{- if and .View (not .IsCollection) }}
return &{{ .InitName }}{ {{ .ReturnVar }}, {{ printf "%q" .View }} }
{{- else }}
return {{ .ReturnVar }}
{{- end -}}
`

	validateResultTypeT = `{{- if .IsCollection -}}
for _, projected := range result {
	if err2 := projected.Validate(); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
{{- else -}}
switch result.View {
	{{- range .Views }}
	{{- if ne .Name "default" }}
	case {{ printf "%q" .Name }}:
	{{- else }}
	default:
	{{- end }}
	{{ .Validate }}
	{{- end }}
}
{{- end -}}
`
)
