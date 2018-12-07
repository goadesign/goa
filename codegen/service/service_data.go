package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
		// ViewScope initialized with all the viewed types.
		ViewScope *codegen.NameScope
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
		// StreamingPayload is the name of the streaming payload type if any.
		StreamingPayload string
		// StreamingPayloadDef is the streaming payload type definition if any.
		StreamingPayloadDef string
		// StreamingPayloadRef is a reference to the streaming payload type if any.
		StreamingPayloadRef string
		// StreamingPayloadDesc is the streaming payload type description if any.
		StreamingPayloadDesc string
		// StreamingPayloadEx is an example of a valid streaming payload value.
		StreamingPayloadEx interface{}
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
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendTypeName is the type name sent through the stream.
		SendTypeName string
		// SendTypeRef is the reference to the type sent through the stream.
		SendTypeRef string
		// RecvName is the name of the receive function.
		RecvName string
		// RecvDesc is the description for the recv function.
		RecvDesc string
		// RecvTypeName is the type name received from the stream.
		RecvTypeName string
		// RecvTypeRef is the reference to the type received from the stream.
		RecvTypeRef string
		// MustClose indicates whether the stream should implement the Close()
		// function.
		MustClose bool
		// EndpointStruct is the name of the endpoint struct that holds a payload
		// reference (if any) and the endpoint server stream. It is set only if the
		// client sends a normal payload and server streams a result.
		EndpointStruct string
		// Kind is the kind of the stream (payload or result or bidirectional).
		Kind expr.StreamKind
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
		Type expr.UserType
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
		Flows []*expr.FlowExpr
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

// PayloadAnalyzer returns an attribute analyzer for service payload type.
// Service payload uses non-pointers to hold attributes having default
// values.
func PayloadAnalyzer(payload *expr.AttributeExpr, pkg string, scope *codegen.NameScope) codegen.AttributeAnalyzer {
	return TypeAnalyzer(payload, pkg, scope)
}

// ResultAnalyzer returns an attribute analyzer for service result type.
// Service result uses non-pointers to hold required attributes. Attributes
// having default values are stored in pointers to enable setting default
// values appropriately if none provided.
func ResultAnalyzer(result *expr.AttributeExpr, pkg string, scope *codegen.NameScope) codegen.AttributeAnalyzer {
	return codegen.NewAttributeAnalyzer(result, true, false, false, true, pkg, scope)
}

// ProjectedTypeAnalyzer returns an attribute analyzer for a projected type.
// Projected type uses pointers for all attributes (even the required ones)
// except for map and array type.
func ProjectedTypeAnalyzer(att *expr.AttributeExpr, pkg string, scope *codegen.NameScope) codegen.AttributeAnalyzer {
	return codegen.NewAttributeAnalyzer(att, true, true, true, true, pkg, scope)
}

// TypeAnalyzer returns an attribute analyzer for a service type.
// Service type uses non-pointers to hold attributes having default values.
func TypeAnalyzer(att *expr.AttributeExpr, pkg string, scope *codegen.NameScope) codegen.AttributeAnalyzer {
	return codegen.NewAttributeAnalyzer(att, true, false, false, true, pkg, scope)
}

// Get retrieves the data for the service with the given name computing it if
// needed. It returns nil if there is no service with the given name.
func (d ServicesData) Get(name string) *Data {
	if data, ok := d[name]; ok {
		return data
	}
	service := expr.Root.Service(name)
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
func (d ServicesData) analyze(service *expr.ServiceExpr) *Data {
	var (
		scope      *codegen.NameScope
		viewScope  *codegen.NameScope
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
		viewScope = codegen.NewNameScope()
		pkgName = scope.HashedUnique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
		viewspkg = pkgName + "views"
		seen = make(map[string]struct{})
		seenErrors = make(map[string]struct{})
		seenProj = make(map[string]*ProjectedTypeData)
		seenViewed = make(map[string]*ViewedResultTypeData)

		// A function to convert raw object type to user type.
		makeUserType := func(att *expr.AttributeExpr, name string) {
			if _, ok := att.Type.(*expr.Object); ok {
				att.Type = &expr.UserTypeExpr{
					AttributeExpr: expr.DupAtt(att),
					TypeName:      name,
				}
			}
			if ut, ok := att.Type.(expr.UserType); ok {
				seen[ut.ID()] = struct{}{}
			}
		}

		for _, e := range service.Methods {
			name := codegen.Goify(e.Name, true)
			// Create user type for raw object payloads
			makeUserType(e.Payload, name+"Payload")
			// Create user type for raw object streaming payloads
			makeUserType(e.StreamingPayload, name+"StreamingPayload")
			// Create user type for raw object results
			makeUserType(e.Result, name+"Result")
		}
		recordError := func(er *expr.ErrorExpr) {
			an := ResultAnalyzer(er.AttributeExpr, "", scope)
			errTypes = append(errTypes, collectTypes(an, seen)...)
			if er.Type == expr.ErrorResult {
				if _, ok := seenErrors[er.Name]; ok {
					return
				}
				seenErrors[er.Name] = struct{}{}
				errorInits = append(errorInits, buildErrorInitData(er, an))
			}
		}
		for _, er := range service.Errors {
			recordError(er)
		}

		// A function to collect inner user types from an attribute expression
		collectUserTypes := func(an codegen.AttributeAnalyzer) {
			if ut, ok := an.Attribute().Type.(expr.UserType); ok {
				an = an.Dup(ut.Attribute(), true)
			}
			types = append(types, collectTypes(an, seen)...)
		}
		for _, m := range service.Methods {
			// collect inner user types
			resultAn := ResultAnalyzer(m.Result, "", scope)
			collectUserTypes(PayloadAnalyzer(m.Payload, "", scope))
			collectUserTypes(PayloadAnalyzer(m.StreamingPayload, "", scope))
			collectUserTypes(resultAn)
			if _, ok := m.Result.Type.(*expr.ResultTypeExpr); ok {
				// collect projected types for the corresponding result type
				projectedAn := ProjectedTypeAnalyzer(expr.DupAtt(m.Result), viewspkg, viewScope)
				projTypes = append(projTypes, collectProjectedTypes(projectedAn, resultAn, viewspkg, seenProj)...)
			}
			for _, er := range m.Errors {
				recordError(er)
			}
		}
	}

	for _, t := range expr.Root.Types {
		if svcs, ok := t.Attribute().Meta["type:generate:force"]; ok {
			an := TypeAnalyzer(&expr.AttributeExpr{Type: t}, "", scope)
			if len(svcs) > 0 {
				// Force generate type only in the specified services
				for _, svc := range svcs {
					if svc == service.Name {
						types = append(types, collectTypes(an, seen)...)
						break
					}
				}
			} else {
				// Force generate type in all the services
				types = append(types, collectTypes(an, seen)...)
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
			m := buildMethodData(e, pkgName, service, scope)
			if rt, ok := e.Result.Type.(*expr.ResultTypeExpr); ok {
				if vrt, ok := seenViewed[m.Result]; ok {
					m.ViewedResult = vrt
				} else {
					projected := seenProj[rt.ID()]
					resultAn := ResultAnalyzer(e.Result, "", scope)
					projectedAn := ProjectedTypeAnalyzer(&expr.AttributeExpr{Type: projected.Type}, viewspkg, viewScope)
					vrt := buildViewedResultType(resultAn, projectedAn, viewspkg)
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
		ViewScope:         viewScope,
	}
	d[service.Name] = data

	return data
}

// collectTypes recurses through the attribute to gather all user types and
// records them in userTypes.
func collectTypes(an codegen.AttributeAnalyzer, seen map[string]struct{}) (data []*UserTypeData) {
	at := an.Attribute()
	if at == nil || at.Type == expr.Empty {
		return
	}
	collect := func(at *expr.AttributeExpr) []*UserTypeData { return collectTypes(an.Dup(at, true), seen) }
	switch dt := at.Type.(type) {
	case expr.UserType:
		if _, ok := seen[dt.ID()]; ok {
			return nil
		}
		data = append(data, &UserTypeData{
			Name:        dt.Name(),
			VarName:     an.Name(false),
			Description: dt.Attribute().Description,
			Def:         an.Dup(dt.Attribute(), true).Def(),
			Ref:         an.Ref(false),
			Type:        dt,
		})
		seen[dt.ID()] = struct{}{}
		data = append(data, collect(dt.Attribute())...)
	case *expr.Object:
		for _, nat := range *dt {
			data = append(data, collect(nat.Attribute)...)
		}
	case *expr.Array:
		data = append(data, collect(dt.ElemType)...)
	case *expr.Map:
		data = append(data, collect(dt.KeyType)...)
		data = append(data, collect(dt.ElemType)...)
	}
	return
}

// buildErrorInitData creates the data needed to generate code around endpoint error return values.
func buildErrorInitData(er *expr.ErrorExpr, an codegen.AttributeAnalyzer) *ErrorInitData {
	_, temporary := er.AttributeExpr.Meta["goa:error:temporary"]
	_, timeout := er.AttributeExpr.Meta["goa:error:timeout"]
	_, fault := er.AttributeExpr.Meta["goa:error:fault"]
	return &ErrorInitData{
		Name:        fmt.Sprintf("Make%s", codegen.Goify(er.Name, true)),
		Description: er.Description,
		ErrName:     er.Name,
		TypeName:    an.Name(false),
		TypeRef:     an.Ref(false),
		Temporary:   temporary,
		Timeout:     timeout,
		Fault:       fault,
	}
}

// buildMethodData creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func buildMethodData(m *expr.MethodExpr, svcPkgName string, service *expr.ServiceExpr, scope *codegen.NameScope) *MethodData {
	var (
		vname        string
		desc         string
		payloadName  string
		payloadDef   string
		payloadRef   string
		payloadDesc  string
		payloadEx    interface{}
		spayloadName string
		spayloadDef  string
		spayloadRef  string
		spayloadDesc string
		spayloadEx   interface{}
		rname        string
		resultDef    string
		resultRef    string
		resultDesc   string
		resultEx     interface{}
		errors       []*ErrorInitData
		reqs         []*RequirementData
		schemes      []string
		svrStream    *StreamData
		cliStream    *StreamData
	)
	vname = codegen.Goify(m.Name, true)
	desc = m.Description
	if desc == "" {
		desc = codegen.Goify(m.Name, true) + " implements " + m.Name + "."
	}
	if m.Payload.Type != expr.Empty {
		payloadAn := PayloadAnalyzer(m.Payload, "", scope)
		payloadName = payloadAn.Name(false)
		payloadRef = payloadAn.Ref(false)
		if dt, ok := m.Payload.Type.(expr.UserType); ok {
			payloadDef = payloadAn.Dup(dt.Attribute(), true).Def()
		}
		payloadDesc = m.Payload.Description
		if payloadDesc == "" {
			payloadDesc = fmt.Sprintf("%s is the payload type of the %s service %s method.",
				payloadName, m.Service.Name, m.Name)
		}
		payloadEx = m.Payload.Example(expr.Root.API.Random())
	}
	if m.StreamingPayload.Type != expr.Empty {
		spayloadAn := PayloadAnalyzer(m.StreamingPayload, "", scope)
		spayloadName = spayloadAn.Name(false)
		spayloadRef = spayloadAn.Ref(false)
		if dt, ok := m.StreamingPayload.Type.(expr.UserType); ok {
			spayloadDef = spayloadAn.Dup(dt.Attribute(), true).Def()
		}
		spayloadDesc = m.StreamingPayload.Description
		if spayloadDesc == "" {
			spayloadDesc = fmt.Sprintf("%s is the streaming payload type of the %s service %s method.",
				spayloadName, m.Service.Name, m.Name)
		}
		spayloadEx = m.StreamingPayload.Example(expr.Root.API.Random())
	}
	if m.Result.Type != expr.Empty {
		resultAn := ResultAnalyzer(m.Result, "", scope)
		rname = resultAn.Name(false)
		resultRef = resultAn.Ref(false)
		if dt, ok := m.Result.Type.(expr.UserType); ok {
			resultDef = resultAn.Dup(dt.Attribute(), true).Def()
		}
		resultDesc = m.Result.Description
		if resultDesc == "" {
			resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
				rname, m.Service.Name, m.Name)
		}
		resultEx = m.Result.Example(expr.Root.API.Random())
	}
	if len(m.Errors) > 0 {
		errors = make([]*ErrorInitData, len(m.Errors))
		for i, er := range m.Errors {
			resultAn := ResultAnalyzer(er.AttributeExpr, "", scope)
			errors[i] = buildErrorInitData(er, resultAn)
		}
	}
	if m.IsStreaming() {
		svrStream = &StreamData{
			Interface:      vname + "ServerStream",
			VarName:        m.Name + "ServerStream",
			EndpointStruct: vname + "EndpointInput",
			Kind:           m.Stream,
			SendName:       "Send",
			SendDesc:       fmt.Sprintf("Send streams instances of %q.", rname),
			SendTypeName:   rname,
			SendTypeRef:    resultRef,
			MustClose:      true,
		}
		cliStream = &StreamData{
			Interface:    vname + "ClientStream",
			VarName:      m.Name + "ClientStream",
			Kind:         m.Stream,
			RecvName:     "Recv",
			RecvDesc:     fmt.Sprintf("Recv reads instances of %q from the stream.", rname),
			RecvTypeName: rname,
			RecvTypeRef:  resultRef,
		}
		if m.Stream == expr.ClientStreamKind || m.Stream == expr.BidirectionalStreamKind {
			switch m.Stream {
			case expr.ClientStreamKind:
				if resultRef != "" {
					svrStream.SendName = "SendAndClose"
					svrStream.SendDesc = fmt.Sprintf("SendAndClose streams instances of %q and closes the stream.", rname)
					svrStream.MustClose = false
					cliStream.RecvName = "CloseAndRecv"
					cliStream.RecvDesc = fmt.Sprintf("CloseAndRecv stops sending messages to the stream and reads instances of %q from the stream.", rname)
				} else {
					cliStream.MustClose = true
				}
			case expr.BidirectionalStreamKind:
				cliStream.MustClose = true
			}
			svrStream.RecvName = "Recv"
			svrStream.RecvDesc = fmt.Sprintf("Recv reads instances of %q from the stream.", spayloadName)
			svrStream.RecvTypeName = spayloadName
			svrStream.RecvTypeRef = spayloadRef
			cliStream.SendName = "Send"
			cliStream.SendDesc = fmt.Sprintf("Send streams instances of %q.", spayloadName)
			cliStream.SendTypeName = spayloadName
			cliStream.SendTypeRef = spayloadRef
		}
	}
	for _, req := range m.Requirements {
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
		Name:                 m.Name,
		VarName:              vname,
		Description:          desc,
		Payload:              payloadName,
		PayloadDef:           payloadDef,
		PayloadRef:           payloadRef,
		PayloadDesc:          payloadDesc,
		PayloadEx:            payloadEx,
		StreamingPayload:     spayloadName,
		StreamingPayloadDef:  spayloadDef,
		StreamingPayloadRef:  spayloadRef,
		StreamingPayloadDesc: spayloadDesc,
		StreamingPayloadEx:   spayloadEx,
		Result:               rname,
		ResultDef:            resultDef,
		ResultRef:            resultRef,
		ResultDesc:           resultDesc,
		ResultEx:             resultEx,
		Errors:               errors,
		Requirements:         reqs,
		Schemes:              schemes,
		ServerStream:         svrStream,
		ClientStream:         cliStream,
	}
}

// buildSchemeData builds the scheme data for the given scheme and method expr.
func buildSchemeData(s *expr.SchemeExpr, m *expr.MethodExpr) *SchemeData {
	if !expr.IsObject(m.Payload.Type) {
		return nil
	}
	switch s.Kind {
	case expr.BasicAuthKind:
		userAtt := expr.TaggedAttribute(m.Payload, "security:username")
		user := codegen.Goify(userAtt, true)
		passAtt := expr.TaggedAttribute(m.Payload, "security:password")
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
	case expr.APIKeyKind:
		if keyAtt := expr.TaggedAttribute(m.Payload, "security:apikey:"+s.SchemeName); keyAtt != "" {
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
	case expr.JWTKind:
		if keyAtt := expr.TaggedAttribute(m.Payload, "security:token"); keyAtt != "" {
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
	case expr.OAuth2Kind:
		if keyAtt := expr.TaggedAttribute(m.Payload, "security:accesstoken"); keyAtt != "" {
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

// collectProjectedTypes builds a projected type for every user type found
// when recursing through the attributes. It stores the projected types in
// data.
func collectProjectedTypes(projectedAn, attAn codegen.AttributeAnalyzer, viewspkg string, seen map[string]*ProjectedTypeData) (data []*ProjectedTypeData) {
	collect := func(projected, att *expr.AttributeExpr) []*ProjectedTypeData {
		return collectProjectedTypes(projectedAn.Dup(projected, true), attAn.Dup(att, true), viewspkg, seen)
	}
	projected := projectedAn.Attribute()
	att := attAn.Attribute()
	switch pt := projected.Type.(type) {
	case expr.UserType:
		dt := att.Type.(expr.UserType)
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
		pd := buildProjectedType(projectedAn, attAn, viewspkg)
		seen[dt.ID()] = pd
		data = append(data, pd)
		data = append(data, types...)
	case *expr.Array:
		dt := att.Type.(*expr.Array)
		data = append(data, collect(pt.ElemType, dt.ElemType)...)
	case *expr.Map:
		dt := att.Type.(*expr.Map)
		data = append(data, collect(pt.KeyType, dt.KeyType)...)
		data = append(data, collect(pt.ElemType, dt.ElemType)...)
	case *expr.Object:
		dt := att.Type.(*expr.Object)
		for _, n := range *pt {
			data = append(data, collect(n.Attribute, dt.Attribute(n.Name))...)
		}
	}
	return
}

// buildProjectedType builds projected type for the given user type.
//
// projectedAn is the projected attribute analyzer for the given attribute
// analyzer att
//
// viewspkg is the name of the views package
//
func buildProjectedType(projectedAn, attAn codegen.AttributeAnalyzer, viewspkg string) *ProjectedTypeData {
	var (
		projections []*InitData
		typeInits   []*InitData
		validations []*ValidateData

		varname = projectedAn.Name(false)
		pt      = projectedAn.Attribute().Type.(expr.UserType)
	)
	{
		if _, isrt := pt.(*expr.ResultTypeExpr); isrt {
			typeInits = buildTypeInits(projectedAn, attAn)
			projections = buildProjections(projectedAn, attAn)
		}
		validations = buildValidations(projectedAn)
	}
	return &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Name:        varname,
			Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
			VarName:     varname,
			Def:         projectedAn.Dup(pt.Attribute(), true).Def(),
			Ref:         projectedAn.Ref(false),
			Type:        pt,
		},
		Projections: projections,
		TypeInits:   typeInits,
		Validations: validations,
		ViewsPkg:    viewspkg,
	}
}

// buildViewedResultType builds a viewed result type from the given result type
// and projected type.
//
// attAn is the result type analyzer
// projectedAn is the projected type analyzer
//
func buildViewedResultType(attAn, projectedAn codegen.AttributeAnalyzer, viewspkg string) *ViewedResultTypeData {
	var (
		att    = attAn.Attribute()
		vresAn = projectedAn.Dup(att, true)
		rt     = att.Type.(*expr.ResultTypeExpr)
		isarr  = expr.IsArray(att.Type)
	)
	vresAn.SetProperties(true, false, false, false)

	// collect result type views
	var (
		viewName string
		views    []*ViewData
	)
	{
		if !rt.HasMultipleViews() {
			viewName = expr.DefaultView
		}
		if v, ok := att.Meta["view"]; ok && len(v) > 0 {
			viewName = v[0]
		}
		views = make([]*ViewData, 0, len(rt.Views))
		for _, view := range rt.Views {
			views = append(views, &ViewData{Name: view.Name, Description: view.Description})
		}
	}

	// build validation data
	var validate *ValidateData
	{
		data := map[string]interface{}{
			"Projected": projectedAn.Name(false),
			"ArgVar":    "result",
			"Source":    "result",
			"Views":     views,
			"IsViewed":  true,
		}
		buf := &bytes.Buffer{}
		if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
			panic(err) // bug
		}
		name := "Validate" + vresAn.Name(false)
		validate = &ValidateData{
			Name:        name,
			Description: fmt.Sprintf("%s runs the validations defined on the viewed result type %s.", name, vresAn.Name(false)),
			Ref:         attAn.Ref(false),
			Validate:    buf.String(),
		}
	}

	// build constructor to initialize viewed result type from result type
	var init *InitData
	{
		data := map[string]interface{}{
			"ToViewed":      true,
			"ArgVar":        "res",
			"ReturnVar":     "vres",
			"Views":         views,
			"ReturnTypeRef": vresAn.Ref(true),
			"IsCollection":  isarr,
			"TargetType":    vresAn.Name(true),
			"InitName":      "new" + projectedAn.Name(false),
		}
		buf := &bytes.Buffer{}
		if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
			panic(err) // bug
		}
		name := "NewViewed" + vresAn.Name(false)
		init = &InitData{
			Name:        name,
			Description: fmt.Sprintf("%s initializes viewed result type %s from result type %s using the given view.", name, vresAn.Name(false), attAn.Name(false)),
			Args: []*InitArgData{
				{Name: "res", Ref: attAn.Ref(false)},
				{Name: "view", Ref: "string"},
			},
			ReturnTypeRef: vresAn.Ref(true),
			Code:          buf.String(),
		}
	}

	// build constructor to initialize result type from viewed result type
	var resinit *InitData
	{
		data := map[string]interface{}{
			"ToResult":      true,
			"ArgVar":        "vres",
			"ReturnVar":     "res",
			"Views":         views,
			"ReturnTypeRef": attAn.Ref(false),
			"InitName":      "new" + attAn.Name(false),
		}
		buf := &bytes.Buffer{}
		if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
			panic(err) // bug
		}
		name := "New" + attAn.Name(false)
		resinit = &InitData{
			Name:          name,
			Description:   fmt.Sprintf("%s initializes result type %s from viewed result type %s.", name, attAn.Name(false), vresAn.Name(false)),
			Args:          []*InitArgData{{Name: "vres", Ref: vresAn.Ref(true)}},
			ReturnTypeRef: attAn.Ref(false),
			Code:          buf.String(),
		}
	}

	projT := wrapProjected(projectedAn.Attribute().Type.(expr.UserType))
	resvar := attAn.Name(false)
	return &ViewedResultTypeData{
		UserTypeData: &UserTypeData{
			Name:        resvar,
			Description: fmt.Sprintf("%s is the viewed result type that is projected based on a view.", resvar),
			VarName:     resvar,
			Def:         vresAn.Dup(projT.Attribute(), true).Def(),
			Ref:         attAn.Ref(false),
			Type:        projT,
		},
		FullName:     vresAn.Name(true),
		FullRef:      vresAn.Ref(true),
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
func wrapProjected(projected expr.UserType) expr.UserType {
	rt := projected.(*expr.ResultTypeExpr)
	pratt := &expr.NamedAttributeExpr{
		Name:      "projected",
		Attribute: &expr.AttributeExpr{Type: rt, Description: "Type to project"},
	}
	prview := &expr.NamedAttributeExpr{
		Name:      "view",
		Attribute: &expr.AttributeExpr{Type: expr.String, Description: "View to render"},
	}
	return &expr.ResultTypeExpr{
		UserTypeExpr: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{
				Type:       &expr.Object{pratt, prview},
				Validation: &expr.ValidationExpr{Required: []string{"projected", "view"}},
			},
			TypeName: rt.TypeName,
		},
		Identifier: rt.Identifier,
		Views:      rt.Views,
	}
}

// buildTypeInits builds the data to generate the constructor code to
// initialize a result type from a projected type.
func buildTypeInits(projAn, attAn codegen.AttributeAnalyzer) []*InitData {
	projected := projAn.Attribute()
	prt := projected.Type.(*expr.ResultTypeExpr)
	pobj := expr.AsObject(projected.Type)
	parr := expr.AsArray(projected.Type)
	if parr != nil {
		// result type collection
		pobj = expr.AsObject(parr.ElemType.Type)
	}

	// For every view defined in the result type, build a constructor function
	// to create the result type from a projected type based on the view.
	var init []*InitData
	{
		init = make([]*InitData, 0, len(prt.Views))
		for _, view := range prt.Views {
			var (
				srcAn codegen.AttributeAnalyzer
				typ   expr.DataType

				obj = &expr.Object{}
			)
			{
				walkViewAttrs(pobj, view, func(name string, att, _ *expr.AttributeExpr) {
					obj.Set(name, att)
				})
				typ = obj
				if parr != nil {
					typ = &expr.Array{ElemType: &expr.AttributeExpr{
						Type: &expr.ResultTypeExpr{
							UserTypeExpr: &expr.UserTypeExpr{
								AttributeExpr: &expr.AttributeExpr{Type: obj},
								TypeName:      projAn.Dup(parr.ElemType, true).Name(false),
							},
						},
					}}
				}
				srcAn = projAn.Dup(&expr.AttributeExpr{
					Type: prt.Dup(&expr.AttributeExpr{Type: typ}),
				}, true)
			}

			var (
				name    string
				code    string
				helpers []*codegen.TransformFunctionData

				varn = attAn.Name(false)
			)
			{
				name = "new" + varn
				if view.Name != expr.DefaultView {
					name += codegen.Goify(view.Name, true)
				}
				code, helpers = buildConstructorCode(srcAn, attAn, "vres", "res", view.Name)
			}

			init = append(init, &InitData{
				Name:          name,
				Description:   fmt.Sprintf("%s converts projected type %s to service type %s.", name, varn, varn),
				Args:          []*InitArgData{{Name: "vres", Ref: projAn.Ref(true)}},
				ReturnTypeRef: attAn.Ref(false),
				Code:          code,
				Helpers:       helpers,
			})
		}
	}
	return init
}

// buildProjections builds the data to generate the constructor code to
// project a result type to a projected type based on a view.
func buildProjections(projAn, attAn codegen.AttributeAnalyzer) []*InitData {
	var (
		projections []*InitData

		projected = projAn.Attribute()
		rt        = attAn.Attribute().Type.(*expr.ResultTypeExpr)
		prt       = projected.Type.(*expr.ResultTypeExpr)
	)

	projections = make([]*InitData, 0, len(rt.Views))
	for _, view := range rt.Views {
		var (
			tgtAn codegen.AttributeAnalyzer
			typ   expr.DataType

			obj = &expr.Object{}
		)
		{
			pobj := expr.AsObject(projected.Type)
			parr := expr.AsArray(projected.Type)
			if parr != nil {
				// result type collection
				pobj = expr.AsObject(parr.ElemType.Type)
			}
			walkViewAttrs(pobj, view, func(name string, att, _ *expr.AttributeExpr) {
				obj.Set(name, att)
			})
			typ = obj
			if parr != nil {
				typ = &expr.Array{ElemType: &expr.AttributeExpr{
					Type: &expr.ResultTypeExpr{
						UserTypeExpr: &expr.UserTypeExpr{
							AttributeExpr: &expr.AttributeExpr{Type: obj},
							TypeName:      parr.ElemType.Type.Name(),
						},
					},
				}}
			}
			tgtAn = projAn.Dup(&expr.AttributeExpr{
				Type: prt.Dup(&expr.AttributeExpr{Type: typ}),
			}, true)
		}

		var (
			name    string
			code    string
			helpers []*codegen.TransformFunctionData

			tname = projAn.Name(false)
		)
		{
			name = "new" + tname
			if view.Name != expr.DefaultView {
				name += codegen.Goify(view.Name, true)
			}
			code, helpers = buildConstructorCode(attAn, tgtAn, "res", "vres", view.Name)
		}

		projections = append(projections, &InitData{
			Name:          name,
			Description:   fmt.Sprintf("%s projects result type %s into projected type %s using the %q view.", name, attAn.Name(false), tname, view.Name),
			Args:          []*InitArgData{{Name: "res", Ref: attAn.Ref(false)}},
			ReturnTypeRef: projAn.Ref(true),
			Code:          code,
			Helpers:       helpers,
		})
	}
	return projections
}

// buildValidationData builds the data required to generate validations for the
// projected types.
func buildValidations(projAn codegen.AttributeAnalyzer) []*ValidateData {
	var (
		validations []*ValidateData

		projected = projAn.Attribute()
		ut        = projected.Type.(expr.UserType)
	)
	if rt, isrt := ut.(*expr.ResultTypeExpr); isrt {
		// for result types we create a validation function containing view
		// specific validation logic for each view
		arr := expr.AsArray(projected.Type)
		for _, view := range rt.Views {
			data := map[string]interface{}{
				"Projected":    projAn.Name(false),
				"ArgVar":       "result",
				"Source":       "result",
				"IsCollection": arr != nil,
			}
			var (
				name string
				vn   string
			)
			{
				name = "Validate" + projAn.Name(false)
				if view.Name != "default" {
					vn = codegen.Goify(view.Name, true)
					name += vn
				}
			}

			if arr != nil {
				// dealing with an array type
				data["Source"] = "item"
				data["ValidateVar"] = "Validate" + projAn.Dup(arr.ElemType, true).Name(false) + vn
			} else {
				var (
					an     codegen.AttributeAnalyzer
					fields []map[string]interface{}

					o = &expr.Object{}
				)
				{
					walkViewAttrs(expr.AsObject(projected.Type), view, func(name string, attr, vatt *expr.AttributeExpr) {
						if rt, ok := attr.Type.(*expr.ResultTypeExpr); ok {
							// use explicitly specified view (if any) for the attribute,
							// otherwise use default
							vw := ""
							if v, ok := vatt.Meta["view"]; ok && len(v) > 0 && v[0] != expr.DefaultView {
								vw = v[0]
							}
							fields = append(fields, map[string]interface{}{
								"Name":        name,
								"ValidateVar": "Validate" + projAn.Dup(attr, true).Name(false) + codegen.Goify(vw, true),
								"IsRequired":  rt.Attribute().IsRequired(name),
							})
						} else {
							o.Set(name, attr)
						}
					})
					an = projAn.Dup(&expr.AttributeExpr{Type: o, Validation: rt.Validation}, true)
				}
				data["Validate"] = codegen.RecursiveValidationCode(an, "result")
				data["Fields"] = fields
			}

			buf := &bytes.Buffer{}
			if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
				panic(err) // bug
			}

			validations = append(validations, &ValidateData{
				Name:        name,
				Description: fmt.Sprintf("%s runs the validations defined on %s using the %q view.", name, projAn.Name(false), view.Name),
				Ref:         projAn.Ref(false),
				Validate:    buf.String(),
			})
		}
	} else {
		// for a user type or a result type with single view, we generate only one validation
		// function containing the validation logic
		name := "Validate" + projAn.Name(false)
		validations = append(validations, &ValidateData{
			Name:        name,
			Description: fmt.Sprintf("%s runs the validations defined on %s.", name, projAn.Name(false)),
			Ref:         projAn.Ref(false),
			Validate:    codegen.RecursiveValidationCode(projAn.Dup(ut.Attribute(), true), "result"),
		})
	}
	return validations
}

// buildConstructorCode builds the transformation code to create a projected
// type from a service type and vice versa.
//
// source and target contains the projected/service attribute analyzers
//
// sourceVar and targetVar contains the variable name that holds the source and
// target data structures in the transformation code.
//
// view is used to generate the constructor function name.
//
func buildConstructorCode(source, target codegen.AttributeAnalyzer, sourceVar, targetVar, view string) (string, []*codegen.TransformFunctionData) {
	var (
		helpers []*codegen.TransformFunctionData
		buf     bytes.Buffer
	)
	src := source.Attribute()
	tgt := target.Attribute()
	rt := src.Type.(*expr.ResultTypeExpr)
	arr := expr.AsArray(tgt.Type)

	data := map[string]interface{}{
		"ArgVar":       sourceVar,
		"ReturnVar":    targetVar,
		"IsCollection": arr != nil,
		"TargetType":   target.Name(true),
	}

	if arr != nil {
		// result type collection
		init := "new" + target.Dup(arr.ElemType, true).Name(false)
		if view != "" && view != expr.DefaultView {
			init += codegen.Goify(view, true)
		}
		data["InitName"] = init
		if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return buf.String(), helpers
	}

	// service type to projected type (or vice versa)
	targetRTs := &expr.Object{}
	tatt := expr.DupAtt(tgt)
	tobj := expr.AsObject(tatt.Type)
	for _, nat := range *tobj {
		if _, ok := nat.Attribute.Type.(*expr.ResultTypeExpr); ok {
			targetRTs.Set(nat.Name, nat.Attribute)
			tobj.Delete(nat.Name)
		}
	}
	data["Source"] = sourceVar
	data["Target"] = targetVar

	var (
		code string
		err  error
	)
	{

		// build code for target with no result types
		if code, helpers, err = codegen.GoTransform(source, target.Dup(tatt, true), sourceVar, targetVar, "transform"); err != nil {
			panic(err) // bug
		}
	}
	data["Code"] = code

	if view != "" {
		data["InitName"] = target.Dup(src, true).Name(true)
	}
	fields := make([]map[string]interface{}, 0, len(*targetRTs))
	// iterate through the result types found in the target and add the
	// code to initialize them
	for _, nat := range *targetRTs {
		finit := "new" + target.Dup(nat.Attribute, true).Name(false)
		if view != "" {
			v := ""
			if vatt := rt.View(view).AttributeExpr.Find(nat.Name); vatt != nil {
				if attv, ok := vatt.Meta["view"]; ok && len(attv) > 0 && attv[0] != expr.DefaultView {
					// view is explicitly set for the result type on the attribute
					v = attv[0]
				}
			}
			finit += codegen.Goify(v, true)
		}
		fields = append(fields, map[string]interface{}{
			"VarName":   codegen.Goify(nat.Name, true),
			"FieldInit": finit,
		})
	}
	data["Fields"] = fields

	if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), helpers
}

// walkViewAttrs iterates through the attributes in att that are found in the
// given view and executes the walker function.
func walkViewAttrs(obj *expr.Object, view *expr.ViewExpr, walker func(name string, attr, vatt *expr.AttributeExpr)) {
	for _, nat := range *expr.AsObject(view.Type) {
		if attr := obj.Attribute(nat.Name); attr != nil {
			walker(nat.Name, attr, nat.Attribute)
		}
	}
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
	err = Validate{{ $.Projected }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }}.Projected)
	{{- end }}
default:
	err = goa.InvalidEnumValueError("view", {{ .Source }}.View, []interface{}{ {{ range .Views }}{{ printf "%q" .Name }}, {{ end }} })
}
{{- else -}}
	{{- if .IsCollection -}}
for _, {{ $.Source }} := range {{ $.ArgVar }} {
	if err2 := {{ .ValidateVar }}({{ $.Source }}); err2 != nil {
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
	if err2 := {{ .ValidateVar }}({{ $.Source }}.{{ goify .Name true }}); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
		{{- end -}}
	{{- end -}}
{{- end -}}
`
)
