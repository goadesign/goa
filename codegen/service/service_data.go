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

	// ProjectedTypeData contains the data to generate a user type that can
	// be projected based on a view. The generated type uses pointers for all
	// fields so that view specific validation logic may be implemented.
	// A projected type will be generated for every user type found in a
	// method result type having multiple views. If the user type is a result
	// type, then a viewed result type is generated which embeds the projected
	// type with a view attribute holding the view name. Finally, the generated
	// code also defines functions that convert the result types to and from the
	// corresponding viewed result type as well as project the viewed result
	// type based on a view.
	ProjectedTypeData struct {
		*UserTypeData
		// Required is the list of required attributes for the projected type
		// used in generating the validation code.
		Required []string
		// Validate is the validation code run on the projected type.
		Validate string
		// FullRef is the complete reference to the viewed result type
		// (including views package name).
		FullRef string
		// Views lists the views defined on a result type and the functions
		// to project a viewed result type.
		Views []*ProjectData
		// ConvertToResult is the code to convert a viewed result type to a
		// result type.
		ConvertToResult *InitData
		// ConvertToViewed is the code to convert a result type to a viewed
		// result type.
		ConvertToViewed *InitData
	}

	// ProjectData contains data about projecting a viewed result type based on
	// a view.
	ProjectData struct {
		// Name is the view name.
		Name string
		// Description is the view description.
		Description string
		// Project is the code to project a viewed result type based on a view.
		Project *InitData
	}

	// InitData contains the data to generate a constructor function
	// to initialize a type from another type.
	InitData struct {
		// VarName is the name of the constructor function.
		VarName string
		// Description is the function description.
		Description string
		// Args lists arguments to this function.
		Args []*InitArgData
		// Ref is the reference to the source type.
		Ref *InitArgData
		// ReturnRef is the reference to the return type.
		ReturnRef string
		// Code is the transformation code.
		Code string
		// Helpers contain the helpers used in the transformation code.
		Helpers []*codegen.TransformFunctionData
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Ref is the reference to the argument.
		Ref string
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
			if isProjectable(m.Result.Type) {
				_, pts := collectProjectedTypes(m.Result, seenProj, scope, viewspkg)
				projTypes = append(projTypes, pts...)
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
			if isProjectable(e.Result.Type) {
				rt := e.Result.Type.(*design.ResultTypeExpr)
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

// collectProjectedTypes collects all the projected types from the given
// result type and stores them in data.
func collectProjectedTypes(att *design.AttributeExpr, seen map[string]*ProjectedTypeData, scope *codegen.NameScope, viewspkg string) (projected *design.AttributeExpr, data []*ProjectedTypeData) {
	collect := func(att *design.AttributeExpr) (*design.AttributeExpr, []*ProjectedTypeData) {
		return collectProjectedTypes(att, seen, scope, viewspkg)
	}
	var types []*ProjectedTypeData
	projected = att
	switch dt := projected.Type.(type) {
	case design.UserType:
		if _, ok := seen[dt.Name()]; ok {
			return
		}
		seen[dt.Name()] = nil
		projected = design.DupAtt(att)
		pt := projected.Type.(design.UserType)
		var required []string
		if v := pt.Attribute().Validation; v != nil {
			// Remove all required attribute validations from the viewed type
			// so that we would end up with a struct where all its attributes
			// are pointers.
			required = v.Required
			v.Required = []string{}
		}
		_, types = collect(pt.Attribute())
		data = append(data, buildProjectedTypes(projected, att, required, seen, scope, viewspkg)...)
		data = append(data, types...)
	case *design.Array:
		dt.ElemType, types = collect(dt.ElemType)
		data = append(data, types...)
	case *design.Map:
		dt.KeyType, types = collect(dt.KeyType)
		data = append(data, types...)
		dt.ElemType, types = collect(dt.ElemType)
		data = append(data, types...)
	case *design.Object:
		for _, n := range *dt {
			n.Attribute, types = collect(n.Attribute)
			data = append(data, types...)
		}
	}
	return
}

// isProjectable returns true if the given type is a result type with
// more than one view and not a collection.
func isProjectable(dt design.DataType) bool {
	if design.IsArray(dt) {
		return false
	}
	rt, ok := dt.(*design.ResultTypeExpr)
	return ok && len(rt.Views) > 1
}

// buildProjectedTypes builds projected type for the given attribute of type
// user type. If the attribute is a result type with multiple views, then it
// also generates a viewed result type which implements functions to convert
// the result type to and from a viewed result type.
func buildProjectedTypes(projected, att *design.AttributeExpr, required []string, seen map[string]*ProjectedTypeData, scope *codegen.NameScope, viewspkg string) []*ProjectedTypeData {
	var types []*ProjectedTypeData
	ptrt := projected.Type.(design.UserType)
	if !isProjectable(ptrt) {
		// ptrt is a user type or a result type with a single view.
		// This type must still be generated in the viewspkg.
		varname := scope.GoTypeName(projected)
		valattr := design.DupAtt(ptrt.Attribute())
		if valattr.Validation != nil {
			valattr.Validation.Required = required
		}
		p := &ProjectedTypeData{
			UserTypeData: &UserTypeData{
				Name:        ptrt.Name(),
				Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
				VarName:     varname,
				Def:         scope.GoTypeDef(ptrt.Attribute(), false),
				Ref:         scope.GoTypeRef(projected),
				Type:        ptrt,
			},
			Required: required,
			Validate: codegen.RecursiveValidationCode(valattr, false, true, false, "result"),
		}
		types = append(types, p)
		seen[ptrt.Name()] = p
		return types
	}
	// ptrt is a result type with multiple views. Generate two types in the
	// 1. a type with "View" suffix that uses pointers for all attributes
	// 2. a type that embeds the View type and has a "view" attribute
	// (viewed result type)
	vatt := design.DupAtt(projected)
	vt := vatt.Type.(design.UserType)
	vt.Rename(ptrt.Name() + "View")
	ptvar := scope.GoTypeName(vatt)
	ptref := scope.GoTypeRef(vatt)
	types = append(types, &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Name:        vt.Name(),
			Description: fmt.Sprintf("%s is a type which is projected based on a view.", ptvar),
			VarName:     ptvar,
			Def:         scope.GoTypeDef(vt.Attribute(), false),
			Ref:         ptref,
			Type:        vt,
		},
	})
	resvar := scope.GoTypeName(projected)
	views, validate := buildProjectData(vatt, required, scope)
	vr := &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Name:        ptrt.Name(),
			Description: fmt.Sprintf("%s is the viewed result type that projects %s based on a view.", resvar, ptvar),
			VarName:     resvar,
			Def:         fmt.Sprintf("struct {\n%s\n// View to render\nView string\n}\n", ptref),
			Ref:         scope.GoTypeRef(projected),
			Type:        ptrt,
		},
		Required:        required,
		Validate:        validate,
		FullRef:         scope.GoFullTypeRef(projected, viewspkg),
		Views:           views,
		ConvertToResult: convertResult(vatt, att, "vres", "res", viewspkg, "", false, scope),
		ConvertToViewed: convertResult(att, vatt, "res", "v", "", viewspkg, true, scope),
	}
	types = append(types, vr)
	seen[ptrt.Name()] = vr
	return types
}

// buildProjectData builds the project data for the viewed result type.
// It also returns the validation code to validate the viewed result type
// based on a view.
func buildProjectData(vatt *design.AttributeExpr, required []string, scope *codegen.NameScope) ([]*ProjectData, string) {
	rt, ok := vatt.Type.(*design.ResultTypeExpr)
	if !ok {
		return nil, ""
	}
	var (
		views    []*ProjectData
		validate string
	)
	vObj := design.AsObject(vatt.Type)
	vtgt := "result"
	validate = fmt.Sprintf("switch %s.View {", vtgt)
	for _, view := range rt.Views {
		obj := &design.Object{}
		viewObj := view.Type.(*design.Object)
		for _, n := range *viewObj {
			obj.Set(n.Name, vObj.Attribute(n.Name))
		}
		att := &design.AttributeExpr{Type: obj}
		if rt.Validation != nil {
			att.Validation = rt.Validation.Dup()
			att.Validation.Required = required
		}
		validate += fmt.Sprintf("\ncase %q:\n%s", view.Name, codegen.RecursiveValidationCode(att, false, true, false, vtgt))
		views = append(views, &ProjectData{
			Name:        view.Name,
			Description: view.Description,
			Project:     projectToView(vatt, view, scope),
		})
	}
	validate += "\n}"
	return views, validate
}

// convertResult converts the given result type to/from a viewed result type.
// If toViewed is true it converts the method result type to viewed
// result type.
func convertResult(src, tgt *design.AttributeExpr, srcvar, tgtvar, srcpkg, tgtpkg string, toViewed bool, scope *codegen.NameScope) *InitData {
	srcrt := src.Type.(*design.ResultTypeExpr)
	tgtrt := tgt.Type.(*design.ResultTypeExpr)
	// Collect all the non-result type attributes from the target and
	// generate the code using GoTypeTransform. For attributes that
	// are a result type, we well then add our generated converters
	// for initialization.
	obj := &design.Object{}
	nonrt := &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{Type: obj},
		TypeName:      tgtrt.TypeName,
	}
	sobj := design.AsObject(src.Type)
	tobj := design.AsObject(tgt.Type)
	for _, n := range *sobj {
		tatt := tobj.Attribute(n.Name)
		if _, ok := tatt.Type.(*design.ResultTypeExpr); !ok {
			obj.Set(n.Name, tatt)
		}
	}
	if !toViewed {
		// Since we are converting a viewed result type to a method
		// result type we need to set the validations set on the method
		// result type.
		nonrt.Validation = tgtrt.Validation
	}
	code, helpers, err := codegen.GoTypeTransform(src.Type, nonrt, srcvar, tgtvar, srcpkg, tgtpkg, !toViewed, scope)
	if err != nil {
		panic(err) // bug
	}
	// Now iterate through src attributes of type result type and initialize them
	// in the tgt using the generated converters.
	for _, n := range *sobj {
		if _, ok := n.Attribute.Type.(*design.ResultTypeExpr); !ok {
			continue
		}
		tatt := tobj.Attribute(n.Name)
		varn := codegen.Goify(n.Name, true)
		init := "New"
		if toViewed {
			init = "NewViewed"
		}
		init += scope.GoTypeName(tatt)
		code += fmt.Sprintf("\nif %s.%s != nil {\n%s.%s = %s(%s.%s)\n}\n", srcvar, varn, tgtvar, varn, init, srcvar, varn)
	}
	var (
		vname     string
		desc      string
		args      []*InitArgData
		returnRef string
	)
	if toViewed {
		rname := codegen.Goify(srcrt.TypeName, true)
		code += fmt.Sprintf("\nreturn &%s.%s{%s: %s}", tgtpkg, rname, scope.GoTypeName(tgt), tgtvar)
		vname = "NewViewed" + rname
		desc = fmt.Sprintf("%s converts result type %s to viewed result type %s.", vname, rname, rname)
		args = []*InitArgData{{Name: srcvar, Ref: scope.GoTypeRef(src)}}
		returnRef = scope.GoFullTypeRef(src, tgtpkg)
	} else {
		code += "\nreturn " + tgtvar
		rname := codegen.Goify(tgtrt.TypeName, true)
		vname = "New" + rname
		desc = fmt.Sprintf("%s converts viewed result type %s to result type %s.", vname, rname, rname)
		args = []*InitArgData{{Name: srcvar, Ref: scope.GoFullTypeRef(tgt, srcpkg)}}
		returnRef = scope.GoTypeRef(tgt)
	}
	return &InitData{
		VarName:     vname,
		Description: desc,
		Args:        args,
		ReturnRef:   returnRef,
		Code:        code,
		Helpers:     helpers,
	}
}

// projectToView projects a viewed result type with only the attributes in the
// given view.
func projectToView(vatt *design.AttributeExpr, view *design.ViewExpr, scope *codegen.NameScope) *InitData {
	rt, ok := vatt.Type.(*design.ResultTypeExpr)
	if !ok {
		return nil
	}
	var (
		rname  string
		srcvar string
		tgtvar string
	)
	rname = codegen.Goify(view.Parent.TypeName, true)
	srcvar = "result"
	tgtvar = "t"
	// nonrt will contain all attributes that are non-result type.
	obj := &design.Object{}
	nonrt := &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{Type: obj},
		TypeName:      rt.TypeName,
	}

	viewObj := view.Type.(*design.Object)
	// First iteration through view attributes will collect only the
	// attributes that are not result types so that we can use
	// GoTypeTransform to generate the code to convert result type
	// to transformed result type (and vice versa) based on view.
	for _, n := range *viewObj {
		attr := vatt.Find(n.Name)
		if _, ok := attr.Type.(*design.ResultTypeExpr); !ok {
			obj.Set(n.Name, attr)
		}
	}
	code, helpers, err := codegen.GoTypeTransform(vatt.Type, nonrt, srcvar, tgtvar, "", "", false, scope)
	if err != nil {
		panic(err) // bug
	}
	// Second iteration through view attributes will append the
	// conversion code for result types based on view.
	for _, n := range *viewObj {
		attr := vatt.Find(n.Name)
		if _, ok := attr.Type.(*design.ResultTypeExpr); !ok {
			continue
		}
		varn := codegen.Goify(n.Name, true)
		srcv := "result." + varn
		tgtv := "t." + varn
		useView := "Default"
		if attV, ok := n.Attribute.Metadata["view"]; ok {
			// if a view is explicitly set for the result type on the view attribute
			// use that view.
			useView = codegen.Goify(attV[0], true)
		}
		code += fmt.Sprintf("\nif %s != nil {\n%s = %s.As%s()\n}\n", srcv, tgtv, srcv, useView)
	}
	code += fmt.Sprintf("\nreturn &%s{\n%s: %s,\nView: %q,\n}", rname, rt.TypeName, tgtvar, view.Name)
	vname := "As" + codegen.Goify(view.Name, true)
	ref := scope.GoTypeRef(&design.AttributeExpr{Type: view.Parent})
	return &InitData{
		VarName:     vname,
		Description: fmt.Sprintf("%s projects viewed result type %s using the %s view.", vname, rname, view.Name),
		Ref:         &InitArgData{Name: srcvar, Ref: ref},
		ReturnRef:   ref,
		Code:        code,
		Helpers:     helpers,
	}
}
