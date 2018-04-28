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
		// ViewedTypes lists the type definitions used to render a result type
		// based on views.
		ViewedTypes []*ViewedTypeData
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
		// ViewedResult if non-nil indicates that the method result type has
		// more than one view.
		ViewedResult *ViewedTypeData
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

	// ViewedTypeData contains the data to generate types to deal with
	// view rendering and also the service constructor code to convert
	// a result type to its viewed counterpart (and vice versa). If ViewType
	// attribute is set, it denotes a result type having multiple views.
	// All the fields in the type struct are pointers except for the view
	// attribute in the viewed result type which is never a pointer.
	ViewedTypeData struct {
		// Result type for which the viewed type is generated.
		*UserTypeData
		// ViewType contains the data describing a viewed result type.
		ViewType *UserTypeData
		// FullRef is the complete reference to the viewed type
		// (including views package name).
		FullRef string
		// Views lists views defined on a result type
		Views []*ViewData
		// ToResult is the constructor code to convert a viewed result type
		// to a result type.
		ToResult *InitData
		// ToViewed is the constructor code to convert a result type to a
		// viewed result type.
		ToViewed *InitData
		// Validate is the validations run on the type which is not a
		// viewed type (i.e., ViewType is not set)
		Validate string
	}

	// ViewData contains data about a result type view.
	ViewData struct {
		// Name is the view name.
		Name string
		// Description is the view description.
		Description string
		// Conversion is the code to transform a result type to the view.
		Conversion *InitData
		// Validate is the validation code to run on a viewed result
		// type based on a view.
		Validate string
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
		scope       *codegen.NameScope
		pkgName     string
		viewsPkg    string
		types       []*UserTypeData
		errTypes    []*UserTypeData
		errorInits  []*ErrorInitData
		viewedTypes []*ViewedTypeData
		seenErrors  map[string]struct{}
		seen        map[string]struct{}
	)
	{
		scope = codegen.NewNameScope()
		pkgName = scope.Unique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
		viewsPkg = pkgName + "views"
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
		schemes []*SchemeData
	)
	{
		methods = make([]*MethodData, len(service.Methods))
		seen = make(map[string]struct{})
		for i, e := range service.Methods {
			m := buildMethodData(e, pkgName, scope)
			if isProjectable(e.Result.Type) {
				_, types := collectViewedTypes(e.Result, seen, scope, viewsPkg)
				viewedTypes = append(viewedTypes, types...)
				for _, t := range viewedTypes {
					if e.Result.Type.Name() == t.Name {
						m.ViewedResult = t
						break
					}
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
		Name:        service.Name,
		Description: desc,
		VarName:     codegen.Goify(service.Name, false),
		StructName:  codegen.Goify(service.Name, true),
		PkgName:     pkgName,
		ViewsPkg:    viewsPkg,
		Methods:     methods,
		Schemes:     schemes,
		UserTypes:   types,
		ErrorTypes:  errTypes,
		ErrorInits:  errorInits,
		ViewedTypes: viewedTypes,
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

// collectViewedTypes collects all the viewed types from the given result type.
// It returns the viewed result type and all the other viewed types found
// during the recursion.
func collectViewedTypes(at *design.AttributeExpr, seen map[string]struct{}, scope *codegen.NameScope, viewsPkg string) (viewed design.DataType, data []*ViewedTypeData) {
	if at == nil || at.Type == design.Empty {
		return
	}
	collect := func(at *design.AttributeExpr) (design.DataType, []*ViewedTypeData) {
		return collectViewedTypes(at, seen, scope, viewsPkg)
	}
	switch dt := at.Type.(type) {
	case design.UserType:
		vAtt := design.DupAtt(at)
		viewed = vAtt.Type
		if _, ok := seen[dt.Name()]; ok {
			return
		}
		seen[dt.Name()] = struct{}{}
		ut := viewed.(design.UserType)
		// Remove all validations from the viewed type so that we would end up
		// with a struct where all its attributes are pointers.
		ut.Attribute().Validation = &design.ValidationExpr{}
		_, types := collect(ut.Attribute())
		data = append(data, types...)
		data = append(data, buildViewedType(vAtt, at, scope, viewsPkg))
	case *design.Array:
		viewed = dt
		vt, types := collect(dt.ElemType)
		viewed.(*design.Array).ElemType.Type = vt
		data = append(data, types...)
	case *design.Map:
		viewed = dt
		vMap := viewed.(*design.Map)
		vt, types := collect(dt.KeyType)
		vMap.KeyType.Type = vt
		data = append(data, types...)
		vt, types = collect(dt.ElemType)
		vMap.ElemType.Type = vt
		data = append(data, types...)
	case *design.Object:
		viewed = dt
		vObj := viewed.(*design.Object)
		for _, n := range *dt {
			vt, types := collect(n.Attribute)
			att := design.DupAtt(n.Attribute)
			att.Type = vt
			data = append(data, types...)
			vObj.Set(n.Name, att)
		}
	default:
		viewed = dt
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

// isProjectable returns true if the given type is a result type with
// more than one view and not a collection.
func isProjectable(dt design.DataType) bool {
	if design.IsArray(dt) {
		return false
	}
	rt, ok := dt.(*design.ResultTypeExpr)
	return ok && len(rt.Views) > 1
}

// buildViewedType builds the type data to generate views and the constructor
// code to transform result type to its viewed counterpart (and vice versa).
func buildViewedType(vAtt, at *design.AttributeExpr, scope *codegen.NameScope, viewsPkg string) *ViewedTypeData {
	resultVar := scope.GoTypeName(at)
	ut := vAtt.Type.(design.UserType)
	dt := at.Type.(design.UserType)
	if !isProjectable(dt) {
		varName := scope.GoTypeName(vAtt)
		return &ViewedTypeData{
			UserTypeData: &UserTypeData{
				Name:        ut.Name(),
				Description: fmt.Sprintf("%s is the transformed type of %s type.", varName, resultVar),
				VarName:     varName,
				Def:         scope.GoTypeDef(ut.Attribute(), true),
				Ref:         scope.GoTypeRef(vAtt),
				Type:        ut,
			},
			Validate: codegen.RecursiveValidationCode(ut.Attribute(), true, false, false, "result"),
		}
	}
	vt := design.Dup(ut).(design.UserType)
	vt.Rename(ut.Name() + "View")
	vAtt.Type = vt
	ref := scope.GoTypeRef(vAtt)
	varName := codegen.Goify(vt.Name(), true)
	return &ViewedTypeData{
		ViewType: &UserTypeData{
			Name:        vt.Name(),
			Description: fmt.Sprintf("%s is the transformed type of %s type.", varName, resultVar),
			VarName:     varName,
			Def:         scope.GoTypeDef(ut.Attribute(), true),
			Ref:         ref,
			Type:        vt,
		},
		UserTypeData: &UserTypeData{
			Name:        dt.Name(),
			VarName:     resultVar,
			Description: fmt.Sprintf("%s is a result type with a view.", resultVar),
			Def:         fmt.Sprintf("struct {\n%s\n// View to render\nView string\n}\n", ref),
			Ref:         scope.GoTypeRef(at),
			Type:        dt,
		},
		FullRef:  scope.GoFullTypeRef(at, viewsPkg),
		Views:    buildViews(vAtt, scope),
		ToResult: convertResult(vAtt, at, "vRes", "res", viewsPkg, "", false, scope),
		ToViewed: convertResult(at, vAtt, "res", "v", "", viewsPkg, true, scope),
	}
}

// buildViews builds the view data for the views defined in the given
// result type.
func buildViews(vAtt *design.AttributeExpr, scope *codegen.NameScope) []*ViewData {
	var views []*ViewData
	rt, ok := vAtt.Type.(*design.ResultTypeExpr)
	if !ok {
		return views
	}
	vObj := design.AsObject(vAtt.Type)
	for _, view := range rt.Views {
		obj := &design.Object{}
		viewObj := view.Type.(*design.Object)
		for _, n := range *viewObj {
			obj.Set(n.Name, vObj.Attribute(n.Name))
		}
		att := &design.AttributeExpr{Type: obj}
		att.Validation = view.Parent.Validation
		views = append(views, &ViewData{
			Name:        view.Name,
			Description: view.Description,
			Conversion:  transformToView(vAtt, view, scope),
			Validate:    codegen.RecursiveValidationCode(att, false, true, false, "result"),
		})
	}
	return views
}

// convertResult converts the given result type to/from a viewed result type.
// If toViewed is true it converts the method result type to viewed
// result type.
func convertResult(src, tgt *design.AttributeExpr, srcVar, tgtVar, srcPkg, tgtPkg string, toViewed bool, scope *codegen.NameScope) *InitData {
	srcRT, ok := src.Type.(*design.ResultTypeExpr)
	if !ok {
		return nil
	}
	tgtRT, ok := tgt.Type.(*design.ResultTypeExpr)
	if !ok {
		return nil
	}
	// Collect all the non-result type attributes from the target and
	// generate the code using GoTypeTransform. For attributes that
	// are a result type, we well then add our generated converters
	// for initialization.
	nonRTObj := &design.Object{}
	nonRT := &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{Type: nonRTObj},
		TypeName:      tgtRT.TypeName,
	}
	if !toViewed {
		// Since we are converting a viewed result type to a method
		// result type we need to set the validations set on the method
		// result type.
		nonRT.Validation = tgtRT.Validation
	}
	srcObj := design.AsObject(src.Type)
	tgtObj := design.AsObject(tgt.Type)
	for _, n := range *srcObj {
		attr := tgtObj.Attribute(n.Name)
		if _, ok := attr.Type.(*design.ResultTypeExpr); !ok {
			nonRTObj.Set(n.Name, attr)
		}
	}
	code, helpers, err := codegen.GoTypeTransform(src.Type, nonRT, srcVar, tgtVar, srcPkg, tgtPkg, !toViewed, scope)
	if err != nil {
		panic(err) // bug
	}
	// Now iterate through src attributes of type result type and initialize them
	// in the tgt using the generated converters.
	for _, n := range *srcObj {
		if _, ok := n.Attribute.Type.(*design.ResultTypeExpr); !ok {
			continue
		}
		varN := codegen.Goify(n.Name, true)
		newVar := "New"
		if toViewed {
			newVar = "NewViewed"
		}
		code += fmt.Sprintf("\nif %s.%s != nil {\n%s.%s = %s%s(%s.%s)\n}\n", srcVar, varN, tgtVar, varN, newVar, scope.GoTypeName(n.Attribute), srcVar, varN)
	}
	var (
		varName   string
		desc      string
		args      []*InitArgData
		returnRef string
	)
	if toViewed {
		resultName := codegen.Goify(srcRT.TypeName, true)
		code += fmt.Sprintf("\nreturn &%s.%s{%s: %s}", tgtPkg, resultName, scope.GoTypeName(tgt), tgtVar)
		varName = "NewViewed" + resultName
		desc = fmt.Sprintf("%s converts result type %s to viewed result type %s.", varName, resultName, resultName)
		args = []*InitArgData{{Name: srcVar, Ref: scope.GoTypeRef(src)}}
		returnRef = scope.GoFullTypeRef(src, tgtPkg)
	} else {
		code += "\nreturn " + tgtVar
		resultName := codegen.Goify(tgtRT.TypeName, true)
		varName = "New" + resultName
		desc = fmt.Sprintf("%s converts viewed result type %s to result type %s.", varName, resultName, resultName)
		args = []*InitArgData{{Name: srcVar, Ref: scope.GoFullTypeRef(tgt, srcPkg)}}
		returnRef = scope.GoTypeRef(tgt)
	}
	return &InitData{
		VarName:     varName,
		Description: desc,
		Args:        args,
		ReturnRef:   returnRef,
		Code:        code,
		Helpers:     helpers,
	}
}

// transformToView transforms a viewed result type to another viewed result
// type with only the attributes in the given view.
func transformToView(vAtt *design.AttributeExpr, view *design.ViewExpr, scope *codegen.NameScope) *InitData {
	rt, ok := vAtt.Type.(*design.ResultTypeExpr)
	if !ok {
		return nil
	}
	var (
		viewN      string
		resultName string
		srcVar     string
		tgtVar     string
	)
	viewN = codegen.Goify(view.Name, true)
	resultName = codegen.Goify(view.Parent.TypeName, true)
	srcVar = "result"
	tgtVar = "t"
	// nonRT will contain all attributes that are non-result type.
	nonRTObj := &design.Object{}
	nonRT := &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{Type: nonRTObj},
		TypeName:      rt.TypeName,
	}

	viewObj := view.Type.(*design.Object)
	// First iteration through view attributes will collect only the
	// attributes that are not result types so that we can use
	// GoTypeTransform to generate the code to convert result type
	// to transformed result type (and vice versa) based on view.
	for _, n := range *viewObj {
		attr := vAtt.Find(n.Name)
		if _, ok := attr.Type.(*design.ResultTypeExpr); !ok {
			nonRTObj.Set(n.Name, attr)
		}
	}
	code, helpers, err := codegen.GoTypeTransform(vAtt.Type, nonRT, srcVar, tgtVar, "", "", false, scope)
	if err != nil {
		panic(err) // bug
	}
	// Second iteration through view attributes will append the
	// conversion code for result types based on view.
	for _, n := range *viewObj {
		attr := vAtt.Find(n.Name)
		if _, ok := attr.Type.(*design.ResultTypeExpr); !ok {
			continue
		}
		varN := codegen.Goify(n.Name, true)
		srcV := "result." + varN
		tgtV := "t." + varN
		useView := "Default"
		if attV, ok := n.Attribute.Metadata["view"]; ok {
			// if a view is explicitly set for the result type on the view attribute
			// use that view.
			useView = codegen.Goify(attV[0], true)
		}
		code += fmt.Sprintf("\nif %s != nil {\n%s = %s.As%s()\n}\n", srcV, tgtV, srcV, useView)
	}
	code += fmt.Sprintf("\nreturn &%s{\n%s: %s,\nView: %q,\n}", resultName, rt.TypeName, tgtVar, view.Name)
	varName := "As" + viewN
	ref := scope.GoTypeRef(&design.AttributeExpr{Type: view.Parent})
	return &InitData{
		VarName:     varName,
		Description: fmt.Sprintf("%s selects fields from the result type %s defined in the %s view.", varName, resultName, view.Name),
		Ref:         &InitArgData{Name: srcVar, Ref: ref},
		ReturnRef:   ref,
		Code:        code,
		Helpers:     helpers,
	}
}
