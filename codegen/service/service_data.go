// This file analyzes evaluated service designs into immutable render data.
// Public type declarations and references come from the frozen generated
// package catalog; mutable scopes are used only for private helper names.
package service

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

var (
	// initTypeTmpl is the template used to render the code that initializes a
	// projected type or viewed result type or a result type.
	initTypeCodeTmpl = template.Must(
		template.New("initTypeCode").
			Funcs(template.FuncMap{"goify": codegen.Goify}).
			Parse(serviceTemplates.Read(returnTypeInitT)),
	)

	// validateTypeCodeTmpl is the template used to render the code to
	// validate a projected type or a viewed result type.
	validateTypeCodeTmpl = template.Must(
		template.New("validateType").
			Funcs(template.FuncMap{"goify": codegen.Goify}).
			Parse(serviceTemplates.Read(typeValidateT)),
	)
)

type (
	// ServicesData encapsulates the data computed from the service designs.
	ServicesData struct {
		Root     *expr.RootExpr
		Services map[string]*Data

		generation *codegen.Generation
		aliases    *importAliases
		packages   map[string]*generatedPackageData
		rootTypes  *rootTypeSet
	}

	// Data contains the data used to render the code related to a single
	// service.
	Data struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// APIName is the name of the API the service belongs to.
		APIName string
		// APIVersion is the API version.
		APIVersion string
		// StructName is the service struct name.
		StructName string
		// VarName is the service variable name (first letter in lowercase).
		VarName string
		// PathName is the service name as used in file and import paths.
		PathName string
		// PkgName is the name of the package containing the generated service
		// code.
		PkgName string
		// ViewsPkg is the name of the package containing the projected and viewed
		// result types.
		ViewsPkg string
		// Methods lists the service interface methods.
		Methods []*MethodData
		// Schemes is the list of security schemes required by the service methods.
		Schemes SchemesData
		// ServerInterceptors contains the data needed to render the server-side
		// interceptors code.
		ServerInterceptors []*InterceptorData
		// ClientInterceptors contains the data needed to render the client-side
		// interceptors code.
		ClientInterceptors []*InterceptorData
		// Scope initialized with all the service types.
		Scope *codegen.NameScope
		// ViewScope initialized with all the viewed types.
		ViewScope *codegen.NameScope
		// ProtoImports lists the import specifications for the custom
		// proto types used by the service.
		ProtoImports []*codegen.ImportSpec
		// userTypes lists the type definitions that the service depends on.
		userTypes []*UserTypeData
		// errorTypes lists the error type definitions that the service depends on.
		errorTypes []*UserTypeData
		// errorInits list the information required to generate error init
		// functions.
		errorInits []*ErrorInitData
		// projectedTypes lists the types which uses pointers for all fields to
		// define view specific validation logic.
		projectedTypes []*ProjectedTypeData
		// unions lists the sum-type unions defined for the service.
		unions []*UnionTypeData
		// viewedResultTypes lists all the viewed method result types.
		viewedResultTypes []*ViewedResultTypeData
		// viewDerived binds the independently rebuilt view graph to declarations
		// reserved while the service was planned.
		viewDerived map[expr.UserType]codegen.DerivedTypeID
	}

	// MethodData describes a single service method.
	MethodData struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// VarName is the Go method name.
		VarName string
		// Idempotent reports whether replaying the exact invocation has the
		// same externally visible effect as invoking the method once.
		Idempotent bool
		// Payload is the name of the payload type if any,
		Payload string
		// PayloadLoc defines the file and Go package of the payload type
		// if overridden via Meta.
		PayloadLoc *codegen.Location
		// PayloadDef is the payload type definition if any.
		PayloadDef string
		// PayloadRef is a reference to the payload type if any,
		PayloadRef string
		// PayloadDeclaration is the immutable generated declaration for a named
		// payload type. It is nil for primitive payloads.
		PayloadDeclaration *codegen.TypeDeclaration
		// PayloadDesc is the payload type description if any.
		PayloadDesc string
		// PayloadEx is an example of a valid payload value.
		PayloadEx any
		// PayloadDefault is the default value of the payload if any.
		PayloadDefault any
		// StreamingPayload is the name of the streaming payload type if any.
		StreamingPayload string
		// StreamingPayloadDef is the streaming payload type definition if any.
		StreamingPayloadDef string
		// StreamingPayloadRef is a reference to the streaming payload type if any.
		StreamingPayloadRef string
		// StreamingPayloadDeclaration is the immutable generated declaration for
		// a named streaming payload type. It is nil for primitive payloads.
		StreamingPayloadDeclaration *codegen.TypeDeclaration
		// StreamingPayloadDesc is the streaming payload type description if any.
		StreamingPayloadDesc string
		// StreamingPayloadEx is an example of a valid streaming payload value.
		StreamingPayloadEx any
		// StreamingResult is the name of the streaming result type if any (when different from Result).
		StreamingResult string
		// StreamingResultDef is the streaming result type definition if any.
		StreamingResultDef string
		// StreamingResultRef is the reference to the streaming result type if any.
		StreamingResultRef string
		// StreamingResultDeclaration is the immutable generated declaration for a
		// named streaming result type. It is nil for primitive results.
		StreamingResultDeclaration *codegen.TypeDeclaration
		// StreamingResultDesc is the streaming result type description if any.
		StreamingResultDesc string
		// StreamingResultEx is an example of a valid streaming result value.
		StreamingResultEx any
		// Result is the name of the result type if any.
		Result string
		// ResultLoc defines the file and Go package of the result type
		// if overridden via Meta.
		ResultLoc *codegen.Location
		// ResultDef is the result type definition if any.
		ResultDef string
		// ResultRef is the reference to the result type if any.
		ResultRef string
		// ResultDeclaration is the immutable generated declaration for a named
		// result type. It is nil for primitive results.
		ResultDeclaration *codegen.TypeDeclaration
		// ResultDesc is the result type description if any.
		ResultDesc string
		// ResultEx is an example of a valid result value.
		ResultEx any
		// Errors list the possible errors defined in the design if any.
		Errors []*ErrorInitData
		// ErrorLocs lists the file and Go package of the error type
		// if overridden via Meta indexed by error name.
		ErrorLocs map[string]*codegen.Location
		// IsJSONRPC indicates if the endpoint is a JSON-RPC endpoint.
		IsJSONRPC bool
		// IsJSONRPCSSE indicates if the JSON-RPC endpoint uses SSE transport.
		IsJSONRPCSSE bool
		// IsJSONRPCWebSocket indicates if the JSON-RPC endpoint uses WebSocket transport.
		IsJSONRPCWebSocket bool
		// Requirements contains the security requirements for the
		// method.
		Requirements RequirementsData
		// Schemes contains the security schemes types used by the
		// method.
		Schemes SchemesData
		// ServerInterceptors list the server interceptors that apply to this
		// method.
		ServerInterceptors []string
		// ClientInterceptors list the client interceptors that apply to this
		// method.
		ClientInterceptors []string
		// ViewedResult contains the data required to generate the code handling
		// views if any.
		ViewedResult *ViewedResultTypeData
		// ServerStream indicates that the service method receives a payload
		// stream or sends a result stream or both.
		ServerStream *StreamData
		// ClientStream indicates that the service method receives a result
		// stream or sends a payload result or both.
		ClientStream *StreamData
		// StreamKind is the kind of the stream (payload or result or
		// bidirectional).
		StreamKind expr.StreamKind
		// HasMixedResults indicates whether the method defines both Result and
		// StreamingResult with different types, enabling content negotiation at
		// the transport layer (e.g. JSON vs SSE over HTTP).
		HasMixedResults bool
		// SkipRequestBodyEncodeDecode is true if the method payload includes
		// the raw HTTP request body reader.
		SkipRequestBodyEncodeDecode bool
		// SkipResponseBodyEncodeDecode is true if the method result includes
		// the raw HTTP response body reader.
		SkipResponseBodyEncodeDecode bool
		// RequestStruct is the name of the data structure containing the
		// payload and request body reader when SkipRequestBodyEncodeDecode is
		// used.
		RequestStruct string
		// ResponseStruct is the name of the data structure containing the
		// result and response body reader when SkipResponseBodyEncodeDecode is
		// used.
		ResponseStruct string
		// EndpointField is the unique field name used in the generated client
		// struct to store the goa.Endpoint for this method. It is computed with a
		// scope that includes method names to avoid field/method name collisions.
		EndpointField string
		// StreamEndpointField is the unique field name used in the generated client
		// struct to store the "streaming mode" goa.Endpoint for mixed results. The
		// transport endpoint forces server streaming (e.g. sets "Accept:
		// text/event-stream") and returns the client stream interface.
		//
		// It is only set when HasMixedResults is true.
		StreamEndpointField string
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
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendWithContextDesc is the description for the send function with context.
		SendWithContextDesc string
		// SendTypeName is the type name sent through the stream.
		SendTypeName string
		// SendTypeRef is the reference to the type sent through the stream.
		SendTypeRef string
		// SendAndCloseName is the name of the send and close function (SSE only).
		SendAndCloseName string
		// SendAndCloseDesc is the description for the send and close function.
		SendAndCloseDesc string
		// SendAndCloseWithContextName is the name of the send and close function with context.
		SendAndCloseWithContextName string
		// SendAndCloseWithContextDesc is the description for the send and close function with context.
		SendAndCloseWithContextDesc string
		// RecvName is the name of the receive function.
		RecvName string
		// RecvDesc is the description for the recv function.
		RecvDesc string
		// RecvWithContextName is the name of the receive function with context.
		RecvWithContextName string
		// RecvWithContextDesc is the description for the recv function with context.
		RecvWithContextDesc string
		// RecvTypeName is the type name received from the stream.
		RecvTypeName string
		// RecvTypeRef is the reference to the type received from the stream.
		RecvTypeRef string
		// MustClose indicates whether the stream should implement the Close()
		// function.
		MustClose bool
		// EndpointStruct is the name of the endpoint struct that holds a payload
		// reference (if any) and the endpoint server stream.
		EndpointStruct string
		// Kind is the kind of the stream (payload, result or bidirectional).
		Kind expr.StreamKind
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

	// InterceptorData contains the data required to render the service-level
	// interceptor code. interceptors.go.tpl
	InterceptorData struct {
		// Name is the name of the interceptor used in the generated code.
		Name string
		// DesignName is the name of the interceptor as defined in the design.
		DesignName string
		// Description is the description of the interceptor from the design.
		Description string
		// Methods
		Methods []*MethodInterceptorData
		// ReadPayload contains payload attributes that the interceptor can
		// read.
		ReadPayload []*AttributeData
		// WritePayload contains payload attributes that the interceptor can
		// write.
		WritePayload []*AttributeData
		// ReadResult contains result attributes that the interceptor can read.
		ReadResult []*AttributeData
		// WriteResult contains result attributes that the interceptor can
		// write.
		WriteResult []*AttributeData
		// ReadStreamingPayload contains streaming payload attributes that the interceptor can read.
		ReadStreamingPayload []*AttributeData
		// WriteStreamingPayload contains streaming payload attributes that the interceptor can write.
		WriteStreamingPayload []*AttributeData
		// ReadStreamingResult contains streaming result attributes that the interceptor can read.
		ReadStreamingResult []*AttributeData
		// WriteStreamingResult contains streaming result attributes that the interceptor can write.
		WriteStreamingResult []*AttributeData
		// HasPayloadAccess indicates that the interceptor info object has a
		// payload access interface.
		HasPayloadAccess bool
		// HasResultAccess indicates that the interceptor info object has a
		// result access interface.
		HasResultAccess bool
		// HasStreamingPayloadAccess indicates that the interceptor info object has a
		// streaming payload access interface.
		HasStreamingPayloadAccess bool
		// HasStreamingResultAccess indicates that the interceptor info object has a
		// streaming result access interface.
		HasStreamingResultAccess bool
	}

	// MethodInterceptorData contains the data required to render the
	// method-level interceptor code.
	MethodInterceptorData struct {
		// MethodName is the name of the method.
		MethodName string
		// PayloadAccess is the name of the payload access struct.
		PayloadAccess string
		// ResultAccess is the name of the result access struct.
		ResultAccess string
		// StreamingPayloadAccess is the name of the streaming payload access struct.
		StreamingPayloadAccess string
		// StreamingResultAccess is the name of the streaming result access struct.
		StreamingResultAccess string
		// PayloadRef is the reference to the method payload type.
		PayloadRef string
		// ResultRef is the reference to the method result type.
		ResultRef string
		// StreamingPayloadRef is the reference to the streaming payload type.
		StreamingPayloadRef string
		// StreamingResultRef is the reference to the streaming result type.
		StreamingResultRef string
		// ServerStream is the stream data if the endpoint defines a server stream.
		ServerStream *StreamInterceptorData
		// ClientStream is the stream data if the endpoint defines a client stream.
		ClientStream *StreamInterceptorData
	}

	// StreamInterceptorData is the stream data for an interceptor.
	StreamInterceptorData struct {
		// Interface is the name of the stream interface.
		Interface string
		// SendName is the name of the send function.
		SendName string
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendTypeRef is the reference to the type sent through the stream.
		SendTypeRef string
		// RecvName is the name of the recv function.
		RecvName string
		// RecvWithContextName is the name of the recv function with context.
		RecvWithContextName string
		// RecvTypeRef is the reference to the type received from the stream.
		RecvTypeRef string
		// MustClose indicates whether the stream should implement the Close()
		// function.
		MustClose bool
		// EndpointStruct is the name of the endpoint struct that holds a payload
		// reference (if any) and the endpoint server stream.
		EndpointStruct string
	}

	// AttributeData describes a single attribute.
	AttributeData struct {
		// Name is the name of the attribute.
		Name string
		// TypeRef is the reference to the attribute type.
		TypeRef string
		// Pointer is true if the attribute is a pointer.
		Pointer bool
	}

	// RequirementsData is the list of security requirements.
	RequirementsData []*RequirementData

	// SchemesData is the list of security schemes.
	SchemesData []*SchemeData

	// RequirementData lists the schemes and scopes defined by a single
	// security requirement.
	RequirementData struct {
		// Schemes list the requirement schemes.
		Schemes []*SchemeData
		// Scopes list the required scopes.
		Scopes []string
	}

	// UserTypeData contains the data describing a user-defined type.
	UserTypeData struct {
		// Declaration is the immutable generated-package record that owns this
		// type in a service, views, or relocated package.
		Declaration *codegen.TypeDeclaration
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
		// Loc defines the file and Go package of the type if overridden
		// via Meta.
		Loc *codegen.Location
		// Type is the underlying type.
		Type expr.UserType
	}

	// UnionTypeData describes a generated sum-type union for a service.
	UnionTypeData struct {
		// Declaration is the immutable generated-package record that owns this
		// union in a service, views, or relocated package.
		Declaration *codegen.UnionDeclaration
		// Name is the Go type name of the union struct.
		Name string
		// KindName is the Go type name of the discriminator kind.
		KindName string
		// Fields describes each union branch.
		Fields []*UnionFieldData
		// Loc defines the file and Go package of the union type if overridden via
		// Meta. When nil the type is generated in the default service file.
		Loc *codegen.Location
		// TypeKey is the discriminator field name for JSON marshaling (defaults to "type").
		TypeKey string
		// ValueKey is the value field name for JSON marshaling (defaults to "value").
		ValueKey string
	}

	// UnionFieldData describes a single branch of a union.
	UnionFieldData struct {
		// Name is the branch name as defined in the DSL.
		Name string
		// KindConst is the Go identifier for the kind constant of this branch.
		KindConst string
		// Constructor is the Go identifier for the branch constructor function.
		Constructor string
		// FieldName is the struct field name in the union.
		FieldName string
		// FieldType is the Go type used in the union struct field and public API.
		FieldType string
		// Nilable is true when the Go branch value can be nil even though the
		// canonical union value is required.
		Nilable bool
		// EmitPrimitiveAlias is true when the branch uses a generated primitive alias
		// that must be declared in the same file as the union type.
		EmitPrimitiveAlias bool
		// PrimitiveAliasType is the underlying Go type used by the generated branch
		// alias (for example "string" or "float64").
		PrimitiveAliasType string
		// TypeTag is the JSON "type" discriminator value for this branch.
		TypeTag string

		reference  *expr.AttributeExpr
		definition *expr.AttributeExpr
	}

	// SchemeData describes a single security scheme.
	SchemeData struct {
		// Kind is the type of scheme, one of "Basic", "APIKey",
		// "Bearer", "JWT" or "OAuth2".
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
		// be initialized with the API key, the bearer token, the JWT
		// token or the OAuth2 access token.
		CredField string
		// CredPointer is true if the credential field is a pointer.
		CredPointer bool
		// CredRequired specifies if the key is a required attribute.
		CredRequired bool
		// KeyAttr is the name of the attribute that contains
		// the security tag (for APIKey, Bearer, OAuth2, and JWT schemes).
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
		// Attributes is the list of attributes rendered in the view.
		Attributes []string
		// TypeVarName is the Go variable name of the type that defines the view.
		TypeVarName string
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
		// Views lists the views defined on the projected type.
		Views []*ViewData
	}

	// InitData contains the data to render a constructor to initialize service
	// types from viewed result types and vice versa.
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

	// ValidateData contains data to render a validate function to validate a
	// projected type or a viewed result type based on views.
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

	// unionDataKey identifies one emitted union definition in one generated Go
	// package without encoding either fact into a string sentinel.
	unionDataKey struct {
		packagePath string
		identity    codegen.UnionTypeID
	}

	// userTypeDataKey distinguishes exact in-memory declarations and the frozen
	// package declaration selected for each one.
	userTypeDataKey struct {
		origin      expr.UserType
		declaration *codegen.TypeDeclaration
	}

	// projectedTypePair binds one rebuilt view declaration to the exact source
	// declaration that gives it a stable DerivedTypeID.
	projectedTypePair struct {
		source             expr.UserType
		projected          expr.UserType
		sourceAttribute    *expr.AttributeExpr
		projectedAttribute *expr.AttributeExpr
	}

	// unionBranchLookup resolves one branch's complete frozen declaration family.
	unionBranchLookup func(*expr.NamedAttributeExpr) (*codegen.UnionBranchDeclaration, error)
)

// NewServicesData analyzes root using declarations frozen by generation.
// Call Plan for every participating root and freeze generation first.
func NewServicesData(root *expr.RootExpr, generation *codegen.Generation) (*ServicesData, error) {
	aliases, err := newImportAliases(root, generation)
	if err != nil {
		return nil, err
	}
	data := &ServicesData{
		Root:       root,
		Services:   make(map[string]*Data),
		generation: generation,
		aliases:    aliases,
		packages:   make(map[string]*generatedPackageData),
		rootTypes:  newRootTypeSet(root),
	}
	for _, service := range root.Services {
		generation.GeneratedPackage(servicePackagePath(generation.GenPkg(), service)).Scope()
		analyzed, err := data.analyze(service)
		if err != nil {
			return nil, err
		}
		data.Services[service.Name] = analyzed
	}
	return data, nil
}

// Get retrieves the analyzed data for the service with the given name. It
// returns nil if there is no service with the given name.
func (d *ServicesData) Get(name string) *Data {
	return d.Services[name]
}

// GenPkg returns the generated module import path shared by every declaration,
// import alias, and transport built from this service analysis.
func (d *ServicesData) GenPkg() string {
	return d.generation.GenPkg()
}

// ServiceImport returns the frozen import alias for name's generated service
// package. The returned value is a copy that callers may add to one file.
func (d *ServicesData) ServiceImport(name string) *codegen.ImportSpec {
	service := d.Root.Service(name)
	if service == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(servicePackagePath(d.generation.GenPkg(), service))
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// ViewImport returns the frozen import alias for name's generated views
// package. The returned value is a copy that callers may add to one file.
func (d *ServicesData) ViewImport(name string) *codegen.ImportSpec {
	service := d.Root.Service(name)
	if service == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(servicePackagePath(d.generation.GenPkg(), service) + "/views")
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// PackageImport returns the frozen import alias for importPath. The returned
// value is a copy that callers may add to one generated file.
func (d *ServicesData) PackageImport(importPath string) *codegen.ImportSpec {
	spec := d.aliases.spec(importPath)
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// ServiceAttributor returns the frozen service declaration resolver for name
// as referenced from outputPackage. The returned resolver follows explicit
// generated package locations and uses the same import aliases as service
// rendering.
func (d *ServicesData) ServiceAttributor(name, outputPackage string) codegen.Attributor {
	service := d.Root.Service(name)
	if service == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	return newServiceResolver(d.generation, d.aliases, service, outputPackage)
}

// ViewAttributor returns the frozen projected and viewed result declaration
// resolver for name as referenced from outputPackage.
func (d *ServicesData) ViewAttributor(name, outputPackage string) codegen.Attributor {
	service := d.Root.Service(name)
	if service == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	data := d.Services[name]
	return newViewResolver(d.generation, d.aliases, service, data.viewDerived).withOutputPackage(outputPackage)
}

// Method returns the service method data for the method with the given name,
// nil if there isn't one.
func (d *Data) Method(name string) *MethodData {
	for _, m := range d.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// Scheme returns the scheme data with the given scheme name.
func (r RequirementsData) Scheme(name string) *SchemeData {
	for _, req := range r {
		for _, s := range req.Schemes {
			if s.SchemeName == name {
				return s
			}
		}
	}
	return nil
}

// Dup creates a copy of the scheme data.
func (s *SchemeData) Dup() *SchemeData {
	return &SchemeData{
		Type:             s.Type,
		SchemeName:       s.SchemeName,
		Name:             s.Name,
		UsernameField:    s.UsernameField,
		UsernamePointer:  s.UsernamePointer,
		UsernameAttr:     s.UsernameAttr,
		UsernameRequired: s.UsernameRequired,
		PasswordField:    s.PasswordField,
		PasswordPointer:  s.PasswordPointer,
		PasswordAttr:     s.PasswordAttr,
		PasswordRequired: s.PasswordRequired,
		CredField:        s.CredField,
		CredPointer:      s.CredPointer,
		CredRequired:     s.CredRequired,
		KeyAttr:          s.KeyAttr,
		Scopes:           s.Scopes,
		Flows:            s.Flows,
		In:               s.In,
	}
}

// Append appends a scheme data to schemes only if it doesn't exist.
func (s SchemesData) Append(d *SchemeData) SchemesData {
	found := false
	for _, se := range s {
		if se.SchemeName == d.SchemeName {
			found = true
			break
		}
	}
	if found {
		return s
	}
	return append(s, d)
}

// DedupeByType returns a new SchemesData slice that is deduplicated by scheme
// type.
func (s SchemesData) DedupeByType() SchemesData {
	seen := make(map[string]struct{})
	uniqueSchemes := SchemesData{}
	for _, s := range s {
		if _, ok := seen[s.Type]; !ok {
			seen[s.Type] = struct{}{}
			uniqueSchemes = append(uniqueSchemes, s)
		}
	}

	return uniqueSchemes
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d *ServicesData) analyze(service *expr.ServiceExpr) (*Data, error) {
	var (
		types      []*UserTypeData
		errTypes   []*UserTypeData
		errorInits []*ErrorInitData
		projTypes  []*ProjectedTypeData
		viewedRTs  []*ViewedResultTypeData
	)
	servicePackage := d.generation.GeneratedPackage(servicePackagePath(d.generation.GenPkg(), service))
	scope := servicePackage.Scope().Fork()
	scope.Unique("Use")       // Reserve "Use" for Endpoints struct Use method.
	scope.Unique("websocket") // Reserve "websocket" to avoid collision with gorilla/websocket
	viewScope := d.generation.GeneratedPackage(
		servicePackagePath(d.generation.GenPkg(), service) + "/views",
	).Scope().Fork()
	pkgName := scope.HashedUnique(service, strings.ToLower(codegen.Goify(service.Name, false)), "svc")
	viewspkg := pkgName + "views"
	seenTypes := make(map[userTypeDataKey]struct{})
	seenErrors := make(map[string]struct{})
	seenProjected := make(map[expr.UserType]expr.UserType)
	seenProj := make(map[expr.UserType]*ProjectedTypeData)
	seenViewed := make(map[string]*ViewedResultTypeData)
	viewDerived := make(map[expr.UserType]codegen.DerivedTypeID)
	serviceResolver := newServiceResolver(
		d.generation,
		d.aliases,
		service,
		servicePackagePath(d.generation.GenPkg(), service),
	)

	// A function to collect user types from an error expression
	recordError := func(er *expr.ErrorExpr) error {
		collected, err := d.collectTypes(er.AttributeExpr, service, serviceResolver, seenTypes, nil)
		if err != nil {
			return err
		}
		errTypes = append(errTypes, collected...)
		if er.Type == expr.ErrorResult {
			if _, ok := seenErrors[er.Name]; ok {
				return nil
			}
			seenErrors[er.Name] = struct{}{}
			errorInits = append(errorInits, buildErrorInitData(er, serviceResolver))
		}
		return nil
	}
	for _, er := range service.Errors {
		if err := recordError(er); err != nil {
			return nil, err
		}
	}

	// A function to collect inner user types from an attribute expression
	collectUserTypes := func(att *expr.AttributeExpr) error {
		if att == nil {
			return nil
		}
		var loc *codegen.Location
		resolver := serviceResolver
		if ut, ok := att.Type.(expr.UserType); ok {
			loc = codegen.UserTypeLocation(ut)
			resolver = serviceResolver.Enter(att).(*declarationResolver)
			att = ut.Attribute()
		}
		collected, err := d.collectTypes(att, service, resolver, seenTypes, loc)
		if err != nil {
			return err
		}
		types = append(types, collected...)
		return nil
	}
	for _, m := range service.Methods {
		// collect inner user types
		if err := collectUserTypes(m.Payload); err != nil {
			return nil, err
		}
		if err := collectUserTypes(m.StreamingPayload); err != nil {
			return nil, err
		}
		if err := collectUserTypes(m.Result); err != nil {
			return nil, err
		}
		// Collect streaming result types if different from Result
		if m.HasMixedResults() {
			if err := collectUserTypes(m.StreamingResult); err != nil {
				return nil, err
			}
		}
		// Collect projected types
		if hasResultType(m.Result) {
			projected, result := projectedResultRoot(service, m)
			pairs := projectTypePairs(projected, result, seenProjected)
			removeMeta(projected)
			views := d.generation.GeneratedPackage(servicePackagePath(d.generation.GenPkg(), service) + "/views")
			for _, pair := range pairs {
				identity := codegen.NewProjectedTypeID(pair.source)
				viewDerived[pair.projected.Origin()] = identity
			}
			viewResolver := newViewResolver(d.generation, d.aliases, service, viewDerived)
			for _, pair := range pairs {
				identity := codegen.NewProjectedTypeID(pair.source)
				declaration, err := views.DerivedType(identity)
				if err != nil {
					return nil, err
				}
				projectedType := buildProjectedType(
					pair.projectedAttribute,
					pair.sourceAttribute,
					viewspkg,
					serviceResolver,
					viewResolver,
					declaration,
				)
				seenProj[pair.source.Origin()] = projectedType
				projTypes = append(projTypes, projectedType)
			}
		}
		for _, er := range m.Errors {
			if err := recordError(er); err != nil {
				return nil, err
			}
		}
	}

	// A function to record method user types so that forced types are not
	// collected twice. Raw object method types are wrapped into synthesized
	// user types by codegen.NormalizeRoot before any generator runs: analyze
	// reads the design and never mutates it, so a raw object here means the
	// root was not normalized.
	recordMethodType := func(m *expr.MethodExpr, att *expr.AttributeExpr) {
		if att == nil || att.Type == expr.Empty {
			return
		}
		if _, ok := att.Type.(*expr.Object); ok {
			panic(fmt.Sprintf(
				"service %q method %q declares a raw object type: codegen.NormalizeRoot must run after eval finalization and before the generators read the design",
				service.Name, m.Name)) // bug
		}
		if ut, ok := att.Type.(expr.UserType); ok {
			declaration := serviceResolver.Enter(att).(*declarationResolver).userType(
				serviceResolver.owner(att),
				ut,
			)
			seenTypes[userTypeDataKey{origin: ut.Origin(), declaration: declaration}] = struct{}{}
		}
	}

	for _, m := range service.Methods {
		recordMethodType(m, m.Payload)
		recordMethodType(m, m.StreamingPayload)
		recordMethodType(m, m.Result)
		if m.HasMixedResults() {
			recordMethodType(m, m.StreamingResult)
		}
	}

	// Add forced types
	for _, t := range d.Root.Types {
		svcs, ok := t.Attribute().Meta["type:generate:force"]
		if !ok {
			continue
		}
		att := &expr.AttributeExpr{Type: t}
		if len(svcs) > 0 {
			// Force generate type only in the specified services
			if slices.Contains(svcs, service.Name) {
				collected, err := d.collectTypes(att, service, serviceResolver, seenTypes, nil)
				if err != nil {
					return nil, err
				}
				types = append(types, collected...)
			}
			continue
		}
		// Force generate type in all the services
		collected, err := d.collectTypes(att, service, serviceResolver, seenTypes, nil)
		if err != nil {
			return nil, err
		}
		types = append(types, collected...)
	}

	var (
		methods []*MethodData
		schemes SchemesData
	)
	methods = make([]*MethodData, len(service.Methods))
	for i, e := range service.Methods {
		m, err := d.buildMethodData(e, scope, serviceResolver)
		if err != nil {
			return nil, err
		}
		methods[i] = m
		for _, s := range m.Schemes {
			schemes = schemes.Append(s)
		}
		rt, ok := e.Result.Type.(*expr.ResultTypeExpr)
		if !ok {
			continue
		}
		var view string
		if v, ok := e.Result.Meta.Last(expr.ViewMetaKey); ok {
			view = v
		}
		if vrt, ok := seenViewed[m.Result+"::"+view]; ok {
			m.ViewedResult = vrt
			continue
		}
		projected := seenProj[rt.Origin()]
		projAtt := &expr.AttributeExpr{Type: projected.Type}
		viewedDeclaration, err := d.generation.GeneratedPackage(
			servicePackagePath(d.generation.GenPkg(), service) + "/views",
		).DerivedType(codegen.NewViewedResultTypeID(rt))
		if err != nil {
			return nil, err
		}
		vrt := buildViewedResultType(
			e.Result,
			projAtt,
			viewspkg,
			serviceResolver,
			newViewResolver(d.generation, d.aliases, service, viewDerived),
			viewedDeclaration,
		)
		found := false
		for _, rt := range viewedRTs {
			if rt.Type.ID() == vrt.Type.ID() {
				found = true
				break
			}
		}
		if !found {
			viewedRTs = append(viewedRTs, vrt)
		}
		m.ViewedResult = vrt
		seenViewed[vrt.Name+"::"+view] = vrt
	}

	// Compute unique EndpointField names using the service-level scope, after
	// method names are set. This records field identifiers without changing
	// existing method names.
	for _, m := range methods {
		m.EndpointField = scope.Unique(m.VarName+"Endpoint", "")
		if m.HasMixedResults {
			m.StreamEndpointField = scope.Unique(m.VarName+"StreamEndpoint", "")
		}
	}

	// Collect union sum-type definitions for the service.
	unionByPackage := make(map[unionDataKey]*UnionTypeData)
	seen := make(map[expr.UserType]struct{})
	collectUnions := func(att *expr.AttributeExpr, loc *codegen.Location) error {
		return d.collectUnionTypes(att, service, serviceResolver, loc, unionByPackage, seen, false)
	}
	for _, t := range types {
		if err := collectUnions(&expr.AttributeExpr{Type: t.Type}, t.Loc); err != nil {
			return nil, err
		}
	}
	for _, t := range errTypes {
		if err := collectUnions(&expr.AttributeExpr{Type: t.Type}, t.Loc); err != nil {
			return nil, err
		}
	}
	for _, m := range service.Methods {
		if m.Payload != nil {
			if err := collectUnions(m.Payload, codegen.UserTypeLocation(m.Payload.Type)); err != nil {
				return nil, err
			}
		}
		if m.StreamingPayload != nil {
			if err := collectUnions(m.StreamingPayload, codegen.UserTypeLocation(m.StreamingPayload.Type)); err != nil {
				return nil, err
			}
		}
		if m.Result != nil {
			if err := collectUnions(m.Result, codegen.UserTypeLocation(m.Result.Type)); err != nil {
				return nil, err
			}
		}
		for _, e := range m.Errors {
			if err := collectUnions(e.AttributeExpr, codegen.UserTypeLocation(e.Type)); err != nil {
				return nil, err
			}
		}
	}
	unions := make([]*UnionTypeData, 0, len(unionByPackage))
	for _, u := range unionByPackage {
		unions = append(unions, u)
	}
	sort.Slice(unions, func(i, j int) bool {
		if unions[i].Name != unions[j].Name {
			return unions[i].Name < unions[j].Name
		}
		var left, right string
		if unions[i].Loc != nil {
			left = unions[i].Loc.RelImportPath
		}
		if unions[j].Loc != nil {
			right = unions[j].Loc.RelImportPath
		}
		return left < right
	})

	desc := service.Description
	if desc == "" {
		desc = fmt.Sprintf("Service is the %s service interface.", service.Name)
	}

	varName := codegen.Goify(service.Name, false)
	data := &Data{
		Name:               service.Name,
		Description:        desc,
		APIName:            d.Root.API.Name,
		APIVersion:         d.Root.API.Version,
		VarName:            varName,
		PathName:           codegen.SnakeCase(varName),
		StructName:         codegen.Goify(service.Name, true),
		PkgName:            pkgName,
		ViewsPkg:           viewspkg,
		Methods:            methods,
		Schemes:            schemes,
		ServerInterceptors: d.collectInterceptors(service, methods, serviceResolver, true),
		ClientInterceptors: d.collectInterceptors(service, methods, serviceResolver, false),
		Scope:              scope,
		ViewScope:          viewScope,
		errorTypes:         errTypes,
		errorInits:         errorInits,
		userTypes:          types,
		projectedTypes:     projTypes,
		viewedResultTypes:  viewedRTs,
		unions:             unions,
		viewDerived:        viewDerived,
	}
	if err := d.registerPackageData(service, data); err != nil {
		return nil, err
	}
	return data, nil
}

// collectInterceptors returns the set of interceptors defined on the given
// service including any interceptor defined on specific service methods or API.
func (d *ServicesData) collectInterceptors(svc *expr.ServiceExpr, methods []*MethodData, resolver *declarationResolver, server bool) []*InterceptorData {
	var ints []*expr.InterceptorExpr
	if server {
		ints = d.Root.API.ServerInterceptors
		ints = append(ints, svc.ServerInterceptors...)
		for _, m := range svc.Methods {
			ints = append(ints, m.ServerInterceptors...)
		}
	} else {
		ints = d.Root.API.ClientInterceptors
		ints = append(ints, svc.ClientInterceptors...)
		for _, m := range svc.Methods {
			ints = append(ints, m.ClientInterceptors...)
		}
	}
	// remove duplicate interceptors
	sort.Slice(ints, func(i, j int) bool {
		return ints[i].Name < ints[j].Name
	})
	for i := 1; i < len(ints); i++ {
		if ints[i-1].Name == ints[i].Name {
			ints = append(ints[:i], ints[i+1:]...)
			i--
		}
	}

	res := make([]*InterceptorData, 0, len(ints))
	for _, i := range ints {
		res = append(res, buildInterceptorData(svc, methods, i, resolver, server))
	}
	return res
}

// declarationContext configures transformations and validations to resolve
// every named service or view type through its planned package declaration.
func declarationContext(resolver codegen.Attributor, pointer bool) *codegen.AttributeContext {
	return &codegen.AttributeContext{
		Pointer:    pointer,
		UseDefault: true,
		Scope:      resolver,
	}
}

// collectTypes recurses through the attribute to gather all user types and
// binds relocated types to their frozen package declarations.
func (d *ServicesData) collectTypes(at *expr.AttributeExpr, service *expr.ServiceExpr, resolver *declarationResolver, seen map[userTypeDataKey]struct{}, loc *codegen.Location) (data []*UserTypeData, err error) {
	if at == nil || at.Type == expr.Empty {
		return nil, nil
	}
	collect := func(at *expr.AttributeExpr, loc *codegen.Location) error {
		collected, err := d.collectTypes(at, service, resolver, seen, loc)
		data = append(data, collected...)
		return err
	}
	switch dt := at.Type.(type) {
	case expr.UserType:
		typeLoc := codegen.UserTypeLocation(dt)
		if typeLoc == nil {
			typeLoc = loc
		}
		entered := resolver.Enter(at).(*declarationResolver)
		declaration := entered.userType(entered.currentPath, dt)
		key := userTypeDataKey{origin: dt.Origin(), declaration: declaration}
		if _, ok := seen[key]; ok {
			return nil, nil
		}
		definitionResolver := entered.inOutputPackage(entered.currentPath)
		data = append(data, &UserTypeData{
			Declaration: declaration,
			Name:        dt.Name(),
			VarName:     declaration.Name(),
			Description: dt.Attribute().Description,
			Def:         definitionResolver.Def(dt.Attribute(), false, true),
			Ref:         definitionResolver.Ref(at, ""),
			Loc:         typeLoc,
			Type:        dt,
		})
		seen[key] = struct{}{}
		collected, collectErr := d.collectTypes(dt.Attribute(), service, entered, seen, typeLoc)
		data = append(data, collected...)
		if collectErr != nil {
			return nil, collectErr
		}
	case *expr.Object:
		for _, nat := range *dt {
			if err := collect(nat.Attribute, loc); err != nil {
				return nil, err
			}
		}
	case *expr.Array:
		if err := collect(dt.ElemType, loc); err != nil {
			return nil, err
		}
	case *expr.Map:
		if err := collect(dt.KeyType, loc); err != nil {
			return nil, err
		}
		if err := collect(dt.ElemType, loc); err != nil {
			return nil, err
		}
	case *expr.Union:
		for _, nat := range dt.Values {
			if userType, ok := generatedUnionBranch(nat, d.rootTypes); ok && loc != nil {
				collected, collectErr := d.collectTypes(&expr.AttributeExpr{Type: userType}, service, resolver, seen, loc)
				data = append(data, collected...)
				if collectErr != nil {
					return nil, collectErr
				}
				continue
			}
			if err := collect(nat.Attribute, loc); err != nil {
				return nil, err
			}
		}
	}
	return data, nil
}

// collectUnionTypes traverses the attribute to gather all union sum-type
// definitions referenced by the service. It records each emitted definition by
// generated package so Extend can copy one union into multiple packages while
// duplicate uses within one package share a definition. When view is true the
// provided location is used for all nested user types so that unions are
// generated in the views package and refer to view-local types (preventing
// import cycles).
func (d *ServicesData) collectUnionTypes(att *expr.AttributeExpr, service *expr.ServiceExpr, resolver *declarationResolver, loc *codegen.Location, unions map[unionDataKey]*UnionTypeData, seen map[expr.UserType]struct{}, view bool) error {
	if att == nil || att.Type == expr.Empty {
		return nil
	}
	recurse := func(att *expr.AttributeExpr, loc *codegen.Location) error {
		return d.collectUnionTypes(att, service, resolver, loc, unions, seen, view)
	}
	switch dt := att.Type.(type) {
	case expr.UserType:
		if _, ok := seen[dt.Origin()]; ok {
			return nil
		}
		seen[dt.Origin()] = struct{}{}
		typeLoc := loc
		entered := resolver.Enter(att).(*declarationResolver)
		if !view {
			if ownLocation := codegen.UserTypeLocation(dt); ownLocation != nil {
				typeLoc = ownLocation
			}
		}
		return d.collectUnionTypes(dt.Attribute(), service, entered, typeLoc, unions, seen, view)
	case *expr.Object:
		for _, nat := range sortedNamedAttributes(*dt) {
			if err := recurse(nat.Attribute, loc); err != nil {
				return err
			}
		}
	case *expr.Array:
		return recurse(dt.ElemType, loc)
	case *expr.Map:
		if err := recurse(dt.KeyType, loc); err != nil {
			return err
		}
		return recurse(dt.ElemType, loc)
	case *expr.Union:
		packagePath := servicePackagePath(d.generation.GenPkg(), service)
		if view {
			packagePath += "/views"
		} else if loc != nil {
			packagePath = generatedPackagePath(d.generation.GenPkg(), service, loc)
		}
		key := unionDataKey{packagePath: packagePath, identity: codegen.NewUnionTypeID(dt)}
		if _, ok := unions[key]; !ok {
			generatedPackage := d.generation.GeneratedPackage(packagePath)
			declaration, err := generatedPackage.Union(dt)
			if err != nil {
				return err
			}
			branchLookup := func(branch *expr.NamedAttributeExpr) (*codegen.UnionBranchDeclaration, error) {
				return generatedPackage.UnionBranch(dt, branch.Name)
			}
			unionData, err := buildUnionTypeData(dt, declaration, resolver.inOutputPackage(packagePath), loc, view, branchLookup)
			if err != nil {
				return err
			}
			if view {
				unions[key] = unionData
			} else {
				owner := d.generatedPackage(service, loc)
				ownedUnion, ok := owner.unions[key.identity]
				if !ok {
					ownedUnion = unionData
					owner.unions[key.identity] = ownedUnion
				}
				unions[key] = ownedUnion
			}
		}
		for _, nat := range dt.Values {
			if err := recurse(nat.Attribute, loc); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildUnionTypeData creates the data needed to generate a sum-type union
// struct, its discriminator kind, and branch metadata. When view is true the
// union is generated in the views package: field types are computed using the
// view scope and are always emitted unqualified so they refer to the
// view-local projected types.
func buildUnionTypeData(u *expr.Union, declaration *codegen.UnionDeclaration, attributor codegen.Attributor, loc *codegen.Location, view bool, branchLookup unionBranchLookup) (*UnionTypeData, error) {
	fields := make([]*UnionFieldData, len(u.Values))
	for i, nat := range u.Values {
		fieldName := codegen.Goify(nat.Name, true)
		branchDeclaration, err := branchLookup(nat)
		if err != nil {
			return nil, err
		}
		fieldType := attributor.Enter(nat.Attribute).Ref(nat.Attribute, "")
		primitiveAliasType, hasPrimitiveAlias := primitiveAliasGoType(nat.Attribute.Type)
		_, isUserType := nat.Attribute.Type.(expr.UserType)
		emitPrimitiveAlias := hasPrimitiveAlias && !isUserType && attributor.Package(nat.Attribute) == ""
		var definition *expr.AttributeExpr
		if _, emitsAlias := branchDeclaration.Type(); emitsAlias {
			definition = nat.Attribute.Type.(expr.UserType).Attribute()
		}
		fields[i] = &UnionFieldData{
			Name:               nat.Name,
			KindConst:          branchDeclaration.KindConst(),
			Constructor:        branchDeclaration.Constructor(),
			FieldName:          fieldName,
			FieldType:          fieldType,
			Nilable:            codegen.IsNilable(nat.Attribute.Type),
			EmitPrimitiveAlias: emitPrimitiveAlias,
			PrimitiveAliasType: primitiveAliasType,
			TypeTag:            nat.Name,
			reference:          nat.Attribute,
			definition:         definition,
		}
	}

	return &UnionTypeData{
		Declaration: declaration,
		Name:        declaration.Name(),
		KindName:    declaration.KindName(),
		Fields:      fields,
		Loc:         loc,
		TypeKey:     u.GetTypeKey(),
		ValueKey:    u.GetValueKey(),
	}, nil
}

// sortedNamedAttributes returns object fields sorted by attribute name.
// Union naming uses NameScope uniqueness, so callers that discover unions while
// traversing objects must use a deterministic field order to avoid oscillating
// generated identifiers across runs.
func sortedNamedAttributes(attrs []*expr.NamedAttributeExpr) []*expr.NamedAttributeExpr {
	if len(attrs) < 2 {
		return attrs
	}
	sorted := slices.Clone(attrs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// primitiveAliasGoType resolves the native Go type for a primitive alias branch.
// It uses expr.IsPrimitive to enforce the type contract and then unwraps aliases.
func primitiveAliasGoType(dt expr.DataType) (string, bool) {
	if !expr.IsPrimitive(dt) {
		return "", false
	}
	for {
		ut, ok := dt.(expr.UserType)
		if !ok {
			return codegen.GoNativeTypeName(dt), true
		}
		dt = ut.Attribute().Type
	}
}

// serviceTypeData returns the frozen declaration name, owning-package
// definition, and service-package reference for one normalized method type.
func serviceTypeData(attribute *expr.AttributeExpr, resolver *declarationResolver) (string, string, string) {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return resolver.Name(attribute, "", false, true),
			"",
			resolver.Ref(attribute, "")
	}
	entered := resolver.Enter(attribute).(*declarationResolver)
	declaration := entered.userType(entered.currentPath, userType)
	definitionResolver := entered.inOutputPackage(entered.currentPath)
	return declaration.Name(),
		definitionResolver.Def(userType.Attribute(), false, true),
		resolver.Ref(attribute, "")
}

// serviceTypeDeclaration returns the frozen declaration for a named method
// type. Primitive method types do not own generated declarations.
func serviceTypeDeclaration(attribute *expr.AttributeExpr, resolver *declarationResolver) *codegen.TypeDeclaration {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return nil
	}
	entered := resolver.Enter(attribute).(*declarationResolver)
	return entered.userType(entered.currentPath, userType)
}

// buildErrorInitData creates the data needed to generate code around endpoint error return values.
func buildErrorInitData(er *expr.ErrorExpr, resolver *declarationResolver) *ErrorInitData {
	_, temporary := er.Meta["goa:error:temporary"]
	_, timeout := er.Meta["goa:error:timeout"]
	_, fault := er.Meta["goa:error:fault"]
	return &ErrorInitData{
		Name:        fmt.Sprintf("Make%s", codegen.Goify(er.Name, true)),
		Description: er.Description,
		ErrName:     er.Name,
		TypeName:    resolver.Name(er.AttributeExpr, "", false, true),
		TypeRef:     resolver.Ref(er.AttributeExpr, ""),
		Temporary:   temporary,
		Timeout:     timeout,
		Fault:       fault,
	}
}

// buildMethodData creates the data needed to render the given endpoint. It
// records the user types needed by the service definition in userTypes.
func (d *ServicesData) buildMethodData(m *expr.MethodExpr, scope *codegen.NameScope, resolver *declarationResolver) (*MethodData, error) {
	var (
		vname       string
		desc        string
		payloadName string
		payloadLoc  *codegen.Location
		payloadDef  string
		payloadRef  string
		payloadDesc string
		payloadEx   any
		rname       string
		resultLoc   *codegen.Location
		resultDef   string
		resultRef   string
		resultDesc  string
		resultEx    any
		errors      []*ErrorInitData
		errorLocs   map[string]*codegen.Location
		isJSONRPC   bool
		reqs        = make(RequirementsData, 0, len(m.Requirements))
		schemes     SchemesData
	)
	vname = scope.Unique(codegen.Goify(m.Name, true), "Endpoint")
	desc = m.Description
	if desc == "" {
		desc = codegen.Goify(m.Name, true) + " implements " + m.Name + "."
	}
	if m.Payload.Type != expr.Empty {
		payloadLoc = codegen.UserTypeLocation(m.Payload.Type)
		payloadName, payloadDef, payloadRef = serviceTypeData(m.Payload, resolver)
		payloadDesc = m.Payload.Description
		if payloadDesc == "" {
			payloadDesc = fmt.Sprintf("%s is the payload type of the %s service %s method.",
				payloadName, m.Service.Name, m.Name)
		}
		payloadEx = m.Payload.Example(d.Root.API.ExampleGenerator)
	}
	if m.Result.Type != expr.Empty {
		resultLoc = codegen.UserTypeLocation(m.Result.Type)
		rname, resultDef, resultRef = serviceTypeData(m.Result, resolver)
		resultDesc = m.Result.Description
		if resultDesc == "" {
			resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
				rname, m.Service.Name, m.Name)
		}
		resultEx = m.Result.Example(d.Root.API.ExampleGenerator)
	}
	if len(m.Errors) > 0 {
		errors = make([]*ErrorInitData, len(m.Errors))
		errorLocs = make(map[string]*codegen.Location, len(m.Errors))
		for i, er := range m.Errors {
			errors[i] = buildErrorInitData(er, resolver)
			errorLocs[er.Name] = codegen.UserTypeLocation(er.Type)
		}
	}

	_, isJSONRPC = m.Meta["jsonrpc"]

	// Check if this JSON-RPC method uses SSE or WebSocket
	var isJSONRPCSSE bool
	var isJSONRPCWebSocket bool
	if isJSONRPC && m.IsStreaming() {
		if httpJSONRPCSvc := d.Root.API.JSONRPC.HTTPExpr.Service(m.Service.Name); httpJSONRPCSvc != nil {
			for _, e := range httpJSONRPCSvc.HTTPEndpoints {
				if e.MethodExpr.Name == m.Name {
					if e.SSE != nil {
						isJSONRPCSSE = true
					} else {
						isJSONRPCWebSocket = true
					}
					break
				}
			}
		}
	}

	for _, req := range expr.EffectiveSecurityRequirements(m.Requirements) {
		var rs SchemesData
		for _, s := range req.Schemes {
			sch := BuildSchemeData(s, m)
			rs = rs.Append(sch)
			schemes = schemes.Append(sch)
		}
		reqs = append(reqs, &RequirementData{Schemes: rs, Scopes: req.Scopes})
	}

	// Unfortunately we can't completely isolate the service codegen from
	// the underlying transport when wanting to skip Goa's built-in decoding.
	skipRequestBodyEncodeDecode := false
	skipResponseBodyEncodeDecode := false
	var httpSvc *expr.HTTPServiceExpr
	for _, svc := range d.Root.API.HTTP.Services {
		if svc.Name() == m.Service.Name {
			httpSvc = svc
			break
		}
	}
	if httpSvc != nil {
		if httpMet := httpSvc.Endpoint(m.Name); httpMet != nil {
			skipRequestBodyEncodeDecode = httpMet.SkipRequestBodyEncodeDecode
			skipResponseBodyEncodeDecode = httpMet.SkipResponseBodyEncodeDecode
		}
	}

	data := &MethodData{
		Name:                         m.Name,
		VarName:                      vname,
		Description:                  desc,
		Idempotent:                   m.Idempotent,
		Payload:                      payloadName,
		PayloadLoc:                   payloadLoc,
		PayloadDef:                   payloadDef,
		PayloadRef:                   payloadRef,
		PayloadDeclaration:           serviceTypeDeclaration(m.Payload, resolver),
		PayloadDesc:                  payloadDesc,
		PayloadEx:                    payloadEx,
		PayloadDefault:               m.Payload.DefaultValue,
		Result:                       rname,
		ResultLoc:                    resultLoc,
		ResultDef:                    resultDef,
		ResultRef:                    resultRef,
		ResultDeclaration:            serviceTypeDeclaration(m.Result, resolver),
		ResultDesc:                   resultDesc,
		ResultEx:                     resultEx,
		Errors:                       errors,
		ErrorLocs:                    errorLocs,
		IsJSONRPC:                    isJSONRPC,
		IsJSONRPCSSE:                 isJSONRPCSSE,
		IsJSONRPCWebSocket:           isJSONRPCWebSocket,
		Requirements:                 reqs,
		Schemes:                      schemes,
		StreamKind:                   m.Stream,
		HasMixedResults:              m.HasMixedResults(),
		SkipRequestBodyEncodeDecode:  skipRequestBodyEncodeDecode,
		SkipResponseBodyEncodeDecode: skipResponseBodyEncodeDecode,
		RequestStruct:                vname + "RequestData",
		ResponseStruct:               vname + "ResponseData",
	}

	if err := d.initStreamData(data, m, vname, rname, resultRef, scope, resolver); err != nil {
		return nil, err
	}
	return data, nil
}

// initStreamData initializes the streaming payload data structures and methods.
func (d *ServicesData) initStreamData(data *MethodData, m *expr.MethodExpr, vname, rname, resultRef string, scope *codegen.NameScope, resolver *declarationResolver) error {
	if !m.IsStreaming() && !m.HasMixedResults() {
		return nil
	}
	var (
		spayloadName string
		spayloadRef  string
		spayloadDef  string
		spayloadDesc string
		spayloadEx   any
		srname       = rname     // streaming result name
		srref        = resultRef // streaming result ref
	)

	// If StreamingResult is different from Result, use it for streaming
	if m.HasMixedResults() && m.StreamingResult != nil && m.StreamingResult.Type != expr.Empty {
		srname, data.StreamingResultDef, srref = serviceTypeData(m.StreamingResult, resolver)
		data.StreamingResult = srname
		data.StreamingResultRef = srref
		data.StreamingResultDeclaration = serviceTypeDeclaration(m.StreamingResult, resolver)
		data.StreamingResultDesc = m.StreamingResult.Description
		if data.StreamingResultDesc == "" {
			data.StreamingResultDesc = fmt.Sprintf("%s is the streaming result type of the %s service %s method.",
				srname, m.Service.Name, m.Name)
		}
		data.StreamingResultEx = m.StreamingResult.Example(d.Root.API.ExampleGenerator)
	}

	if m.StreamingPayload != nil && m.StreamingPayload.Type != expr.Empty {
		spayloadName, spayloadDef, spayloadRef = serviceTypeData(m.StreamingPayload, resolver)
		data.StreamingPayloadDeclaration = serviceTypeDeclaration(m.StreamingPayload, resolver)
		spayloadDesc = m.StreamingPayload.Description
		if spayloadDesc == "" {
			spayloadDesc = fmt.Sprintf("%s is the streaming payload type of the %s service %s method.",
				spayloadName, m.Service.Name, m.Name)
		}
		spayloadEx = m.StreamingPayload.Example(d.Root.API.ExampleGenerator)
	}
	// For JSON-RPC WebSocket:
	// - Client streaming (no result streaming): no endpoint struct needed, just payload
	// - Bidirectional streaming: endpoint struct needed for both payload and stream
	endpointStruct := vname + "EndpointInput"
	if data.IsJSONRPC && m.IsStreaming() && !data.IsJSONRPCSSE && m.Stream == expr.ClientStreamKind {
		endpointStruct = ""
	}
	// For mixed results with SSE, treat as server streaming
	streamKind := m.Stream
	if m.HasMixedResults() && !m.IsStreaming() {
		// Mixed results with SSE should be treated as server streaming
		streamKind = expr.ServerStreamKind
	}
	svrStream := &StreamData{
		Interface:           vname + "ServerStream",
		VarName:             scope.Unique(codegen.Goify(m.Name, true), "ServerStream"),
		EndpointStruct:      endpointStruct,
		Kind:                streamKind,
		SendName:            "Send",
		SendDesc:            fmt.Sprintf("Send streams instances of %q.", srname),
		SendWithContextName: "SendWithContext",
		SendWithContextDesc: fmt.Sprintf("SendWithContext streams instances of %q with context.", srname),
		SendTypeName:        srname,
		SendTypeRef:         srref,
		MustClose:           true,
	}
	cliStream := &StreamData{
		Interface:           vname + "ClientStream",
		VarName:             scope.Unique(codegen.Goify(m.Name, true), "ClientStream"),
		Kind:                streamKind,
		RecvName:            "Recv",
		RecvDesc:            fmt.Sprintf("Recv reads instances of %q from the stream.", srname),
		RecvWithContextName: "RecvWithContext",
		RecvWithContextDesc: fmt.Sprintf("RecvWithContext reads instances of %q from the stream with context.", srname),
		RecvTypeName:        srname,
		RecvTypeRef:         srref,
	}
	// For SSE server streaming, we need both Send (for notifications) and SendAndClose (for final response)
	if data.IsJSONRPCSSE && m.Stream == expr.ServerStreamKind && resultRef != "" {
		svrStream.SendAndCloseName = "SendAndClose"
		svrStream.SendAndCloseDesc = fmt.Sprintf("SendAndClose sends a final response with %q and closes the stream.", srname)
		// For JSON-RPC SSE, methods take context directly; align names accordingly
		svrStream.SendWithContextName = "Send"
		svrStream.RecvWithContextName = "Recv"
		// Update Send description to clarify it's for notifications only
		svrStream.SendDesc = fmt.Sprintf("Send streams JSON-RPC notifications with %q. Notifications do not expect a response.", srname)
	}
	if streamKind == expr.ClientStreamKind || streamKind == expr.BidirectionalStreamKind {
		switch streamKind {
		case expr.ClientStreamKind:
			if srref != "" {
				svrStream.SendName = "SendAndClose"
				svrStream.SendDesc = fmt.Sprintf("SendAndClose streams instances of %q and closes the stream.", srname)
				svrStream.SendWithContextName = "SendAndCloseWithContext"
				svrStream.SendWithContextDesc = fmt.Sprintf("SendAndCloseWithContext streams instances of %q and closes the stream with context.", srname)
				svrStream.MustClose = false
				cliStream.RecvName = "CloseAndRecv"
				cliStream.RecvDesc = fmt.Sprintf("CloseAndRecv stops sending messages to the stream and reads instances of %q from the stream.", srname)
				cliStream.RecvWithContextName = "CloseAndRecvWithContext"
				cliStream.RecvWithContextDesc = fmt.Sprintf("CloseAndRecvWithContext stops sending messages to the stream and reads instances of %q from the stream with context.", srname)
			} else {
				cliStream.MustClose = true
			}
		case expr.BidirectionalStreamKind:
			cliStream.MustClose = true
		}
		svrStream.RecvName = "Recv"
		svrStream.RecvDesc = fmt.Sprintf("Recv reads instances of %q from the stream.", spayloadName)
		svrStream.RecvWithContextName = "RecvWithContext"
		svrStream.RecvWithContextDesc = fmt.Sprintf("RecvWithContext reads instances of %q from the stream with context.", spayloadName)
		svrStream.RecvTypeName = spayloadName
		svrStream.RecvTypeRef = spayloadRef
		cliStream.SendName = "Send"
		cliStream.SendDesc = fmt.Sprintf("Send streams instances of %q.", spayloadName)
		cliStream.SendWithContextName = "SendWithContext"
		cliStream.SendWithContextDesc = fmt.Sprintf("SendWithContext streams instances of %q with context.", spayloadName)
		cliStream.SendTypeName = spayloadName
		cliStream.SendTypeRef = spayloadRef
	}
	data.ClientStream = cliStream
	data.ServerStream = svrStream
	data.StreamingPayload = spayloadName
	data.StreamingPayloadDef = spayloadDef
	data.StreamingPayloadRef = spayloadRef
	data.StreamingPayloadDesc = spayloadDesc
	data.StreamingPayloadEx = spayloadEx
	return nil
}

// buildInterceptorData creates the data needed to generate interceptor code.
func buildInterceptorData(svc *expr.ServiceExpr, methods []*MethodData, i *expr.InterceptorExpr, resolver *declarationResolver, server bool) *InterceptorData {
	data := &InterceptorData{
		Name:        codegen.Goify(i.Name, true),
		DesignName:  i.Name,
		Description: i.Description,
	}
	if len(svc.Methods) == 0 {
		return data
	}
	attributesCollected := false
	for _, m := range svc.Methods {
		applies := false
		intExprs := m.ServerInterceptors
		if !server {
			intExprs = m.ClientInterceptors
		}
		for _, in := range intExprs {
			if in.Name == i.Name {
				if !attributesCollected {
					payload, result, streamingPayload := m.Payload, m.Result, m.StreamingPayload
					data.ReadPayload = collectAttributes(i.ReadPayload, payload, resolver)
					data.WritePayload = collectAttributes(i.WritePayload, payload, resolver)
					data.ReadResult = collectAttributes(i.ReadResult, result, resolver)
					data.WriteResult = collectAttributes(i.WriteResult, result, resolver)
					data.ReadStreamingPayload = collectAttributes(i.ReadStreamingPayload, streamingPayload, resolver)
					data.WriteStreamingPayload = collectAttributes(i.WriteStreamingPayload, streamingPayload, resolver)
					data.ReadStreamingResult = collectAttributes(i.ReadStreamingResult, result, resolver)
					data.WriteStreamingResult = collectAttributes(i.WriteStreamingResult, result, resolver)
					if len(data.ReadPayload) > 0 || len(data.WritePayload) > 0 {
						data.HasPayloadAccess = true
					}
					if len(data.ReadResult) > 0 || len(data.WriteResult) > 0 {
						data.HasResultAccess = true
					}
					if len(data.ReadStreamingPayload) > 0 || len(data.WriteStreamingPayload) > 0 {
						data.HasStreamingPayloadAccess = true
					}
					if len(data.ReadStreamingResult) > 0 || len(data.WriteStreamingResult) > 0 {
						data.HasStreamingResultAccess = true
					}
					attributesCollected = true
				}
				applies = true
				break
			}
		}
		if !applies {
			continue
		}
		var md *MethodData
		for _, mt := range methods {
			if m.Name == mt.Name {
				md = mt
				break
			}
		}
		data.Methods = append(data.Methods, buildInterceptorMethodData(i, md))
		if server {
			md.ServerInterceptors = append(md.ServerInterceptors, i.Name)
		} else {
			md.ClientInterceptors = append(md.ClientInterceptors, i.Name)
		}
	}
	return data
}

// buildInterceptorMethodData creates the data needed to generate interceptor
// method code.
func buildInterceptorMethodData(i *expr.InterceptorExpr, md *MethodData) *MethodInterceptorData {
	var serverStream, clientStream *StreamInterceptorData
	if md.ServerStream != nil {
		serverStream = &StreamInterceptorData{
			Interface:           md.ServerStream.Interface,
			SendName:            md.ServerStream.SendName,
			SendWithContextName: md.ServerStream.SendWithContextName,
			SendTypeRef:         md.ServerStream.SendTypeRef,
			RecvName:            md.ServerStream.RecvName,
			RecvWithContextName: md.ServerStream.RecvWithContextName,
			RecvTypeRef:         md.ServerStream.RecvTypeRef,
			MustClose:           md.ServerStream.MustClose,
			EndpointStruct:      md.ServerStream.EndpointStruct,
		}
	}
	if md.ClientStream != nil {
		clientStream = &StreamInterceptorData{
			Interface:           md.ClientStream.Interface,
			SendName:            md.ClientStream.SendName,
			SendWithContextName: md.ClientStream.SendWithContextName,
			SendTypeRef:         md.ClientStream.SendTypeRef,
			RecvName:            md.ClientStream.RecvName,
			RecvWithContextName: md.ClientStream.RecvWithContextName,
			RecvTypeRef:         md.ClientStream.RecvTypeRef,
			MustClose:           md.ClientStream.MustClose,
		}
	}
	var payloadAccess, resultAccess, streamingPayloadAccess, streamingResultAccess string
	if i.ReadPayload != nil || i.WritePayload != nil {
		payloadAccess = codegen.Goify(i.Name, false) + md.VarName + "Payload"
	}
	if i.ReadResult != nil || i.WriteResult != nil {
		resultAccess = codegen.Goify(i.Name, false) + md.VarName + "Result"
	}
	if i.ReadStreamingPayload != nil || i.WriteStreamingPayload != nil {
		streamingPayloadAccess = codegen.Goify(i.Name, false) + md.VarName + "StreamingPayload"
	}
	if i.ReadStreamingResult != nil || i.WriteStreamingResult != nil {
		streamingResultAccess = codegen.Goify(i.Name, false) + md.VarName + "StreamingResult"
	}
	return &MethodInterceptorData{
		MethodName:             md.VarName,
		PayloadAccess:          payloadAccess,
		ResultAccess:           resultAccess,
		PayloadRef:             md.PayloadRef,
		ResultRef:              md.ResultRef,
		StreamingPayloadAccess: streamingPayloadAccess,
		StreamingPayloadRef:    md.StreamingPayloadRef,
		StreamingResultAccess:  streamingResultAccess,
		StreamingResultRef:     md.ResultRef,
		ClientStream:           clientStream,
		ServerStream:           serverStream,
	}
}

// BuildSchemeData builds the scheme data for the given scheme and method expr.
func BuildSchemeData(s *expr.SchemeExpr, m *expr.MethodExpr) *SchemeData {
	if !expr.IsObject(m.Payload.Type) {
		return nil
	}
	if s.Kind == expr.BasicAuthKind {
		userAtt := expr.TaggedAttribute(m.Payload, "security:username")
		passAtt := expr.TaggedAttribute(m.Payload, "security:password")
		return &SchemeData{
			Type:             s.Kind.String(),
			SchemeName:       s.SchemeName,
			UsernameAttr:     userAtt,
			UsernameField:    codegen.Goify(userAtt, true),
			UsernamePointer:  m.Payload.IsPrimitivePointer(userAtt, true),
			UsernameRequired: m.Payload.IsRequired(userAtt),
			PasswordAttr:     passAtt,
			PasswordField:    codegen.Goify(passAtt, true),
			PasswordPointer:  m.Payload.IsPrimitivePointer(passAtt, true),
			PasswordRequired: m.Payload.IsRequired(passAtt),
			Scopes:           schemeScopes(s),
		}
	}
	// The remaining scheme kinds all carry a single credential attribute
	// identified by a kind-specific security tag on the method payload.
	var tag string
	switch s.Kind {
	case expr.APIKeyKind:
		tag = "security:apikey:" + s.SchemeName
	case expr.BearerKind:
		tag = "security:bearer"
	case expr.JWTKind:
		tag = "security:token"
	case expr.OAuth2Kind:
		tag = "security:accesstoken"
	default:
		return nil
	}
	keyAtt := expr.TaggedAttribute(m.Payload, tag)
	if keyAtt == "" {
		return nil
	}
	data := &SchemeData{
		Type:         s.Kind.String(),
		Name:         s.Name,
		SchemeName:   s.SchemeName,
		CredField:    codegen.Goify(keyAtt, true),
		CredPointer:  m.Payload.IsPrimitivePointer(keyAtt, true),
		CredRequired: m.Payload.IsRequired(keyAtt),
		KeyAttr:      keyAtt,
		Scopes:       schemeScopes(s),
		In:           s.In,
	}
	if s.Kind == expr.OAuth2Kind {
		data.Flows = s.Flows
	}
	return data
}

// schemeScopes returns the scope names defined by the scheme, nil when the
// scheme defines none.
func schemeScopes(s *expr.SchemeExpr) []string {
	if len(s.Scopes) == 0 {
		return nil
	}
	scopes := make([]string, len(s.Scopes))
	for i, sc := range s.Scopes {
		scopes[i] = sc.Name
	}
	return scopes
}

// collectAttributes resolves the interceptor fields selected from parent into
// the generated names, type references, and pointer behavior rendered by the
// interceptor templates.
func collectAttributes(attrNames, parent *expr.AttributeExpr, resolver codegen.Attributor) []*AttributeData {
	if attrNames == nil {
		return nil
	}
	obj := expr.AsObject(attrNames.Type)
	if obj == nil {
		return nil
	}
	data := make([]*AttributeData, len(*obj))
	parentResolver := resolver.Enter(parent)
	for i, nat := range *obj {
		parentAttr := parent.Find(nat.Name)
		if parentAttr == nil {
			// Attribute references are validated at design time so a miss
			// here would surface as a nil deref at template render time.
			panic(fmt.Sprintf("attribute %q not found in parent attribute", nat.Name)) // bug
		}
		data[i] = &AttributeData{
			Name:    codegen.Goify(nat.Name, true),
			TypeRef: parentResolver.Ref(parentAttr, parentResolver.Package(parentAttr)),
			Pointer: parent.IsPrimitivePointer(nat.Name, true),
		}
	}
	return data
}

// projectTypePairs rewrites a copied result graph into pointer-backed view
// types and returns each generated declaration with its exact source. The
// source Origin makes independently rebuilt plan and render graphs select the
// same package record.
func projectTypePairs(projected, source *expr.AttributeExpr, seen map[expr.UserType]expr.UserType) []*projectedTypePair {
	collect := func(projected, source *expr.AttributeExpr) []*projectedTypePair {
		return projectTypePairs(projected, source, seen)
	}
	switch projectedType := projected.Type.(type) {
	case expr.UserType:
		sourceType := source.Type.(expr.UserType)
		origin := sourceType.Origin()
		if existing, ok := seen[origin]; ok {
			if existing != nil {
				projected.Type = existing
			}
			return nil
		}
		seen[origin] = nil
		projectedType.Rename(projectedType.Name() + "View")
		nested := collect(projectedType.Attribute(), sourceType.Attribute())
		seen[origin] = projectedType
		return append([]*projectedTypePair{{
			source:             sourceType,
			projected:          projectedType,
			sourceAttribute:    source,
			projectedAttribute: projected,
		}}, nested...)
	case *expr.Array:
		return collect(projectedType.ElemType, source.Type.(*expr.Array).ElemType)
	case *expr.Map:
		sourceMap := source.Type.(*expr.Map)
		pairs := collect(projectedType.KeyType, sourceMap.KeyType)
		return append(pairs, collect(projectedType.ElemType, sourceMap.ElemType)...)
	case *expr.Object:
		sourceObject := source.Type.(*expr.Object)
		var pairs []*projectedTypePair
		for _, field := range *projectedType {
			pairs = append(pairs, collect(field.Attribute, sourceObject.Attribute(field.Name))...)
		}
		return pairs
	case *expr.Union:
		sourceUnion := source.Type.(*expr.Union)
		var pairs []*projectedTypePair
		for index, branch := range projectedType.Values {
			pairs = append(pairs, collect(branch.Attribute, sourceUnion.Values[index].Attribute)...)
		}
		return pairs
	default:
		return nil
	}
}

// projectedResultRoot returns the root attribute used to collect projected
// view types for m.Result. NormalizeRoot synthesizes user types for raw object
// method results before service analysis; projected view collection keeps the
// pre-normalization shape by traversing those synthetic wrappers' attributes
// directly instead of generating view-local types for the wrappers themselves.
func projectedResultRoot(service *expr.ServiceExpr, m *expr.MethodExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
	identity := codegen.NewMethodResultIdentity(service.Name, m.Name)
	if ut, ok := m.Result.Type.(*expr.UserTypeExpr); ok && identity.Matches(ut) {
		return expr.DupAtt(ut.Attribute()), ut.Attribute()
	}
	return expr.DupAtt(m.Result), m.Result
}

// hasResultType returns true if the given attribute has a result type recursively.
func hasResultType(att *expr.AttributeExpr, seens ...map[expr.UserType]struct{}) bool {
	if _, ok := att.Type.(*expr.ResultTypeExpr); ok {
		return true
	}
	var seen map[expr.UserType]struct{}
	if len(seens) > 0 {
		seen = seens[0]
	} else {
		seen = make(map[expr.UserType]struct{})
	}
	switch a := att.Type.(type) {
	case expr.UserType:
		origin := a.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return hasResultType(a.Attribute(), seen)
	case *expr.Array:
		return hasResultType(a.ElemType, seen)
	case *expr.Map:
		return hasResultType(a.KeyType, seen) || hasResultType(a.ElemType, seen)
	case *expr.Object:
		for _, nat := range *a {
			if hasResultType(nat.Attribute, seen) {
				return true
			}
		}
	case *expr.Union:
		for _, nat := range a.Values {
			if hasResultType(nat.Attribute, seen) {
				return true
			}
		}
	}
	return false
}

// buildProjectedType returns the render data for one pointer-backed view
// declaration and its conversions to the exact source service type.
func buildProjectedType(projected, att *expr.AttributeExpr, viewspkg string, serviceResolver, viewResolver *declarationResolver, declaration *codegen.TypeDeclaration) *ProjectedTypeData {
	var (
		projections []*InitData
		typeInits   []*InitData
		views       []*ViewData

		varname = declaration.Name()
		pt      = projected.Type.(expr.UserType)
	)
	if _, isrt := pt.(*expr.ResultTypeExpr); isrt {
		typeInits = buildViewConversions(projected, att, serviceResolver, viewResolver, true)
		projections = buildViewConversions(projected, att, serviceResolver, viewResolver, false)
		serviceName, _, _ := serviceTypeData(att, serviceResolver)
		views = buildViews(att.Type.(*expr.ResultTypeExpr), serviceName)
	}
	validations := buildValidations(projected, viewResolver)
	removeMeta(projected)
	return &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Declaration: declaration,
			Name:        varname,
			Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
			VarName:     varname,
			Def:         viewResolver.Def(pt.Attribute(), true, true),
			Ref:         viewResolver.Ref(projected, ""),
			Type:        pt,
		},
		Projections: projections,
		TypeInits:   typeInits,
		Validations: validations,
		ViewsPkg:    viewspkg,
		Views:       views,
	}
}

// buildViews builds the view data for all the views in the given result type.
func buildViews(rt *expr.ResultTypeExpr, typeName string) []*ViewData {
	views := make([]*ViewData, len(rt.Views))
	for i, view := range rt.Views {
		vatt := expr.AsObject(view.Type)
		attrs := make([]string, len(*vatt))
		for j, nat := range *vatt {
			attrs[j] = nat.Name
		}
		views[i] = &ViewData{
			Name:        view.Name,
			Description: view.Description,
			Attributes:  attrs,
			TypeVarName: typeName,
		}
	}
	return views
}

// buildViewedResultType builds a viewed result type from the given result type
// and projected type.
func buildViewedResultType(att, projected *expr.AttributeExpr, viewspkg string, serviceResolver, viewResolver *declarationResolver, declaration *codegen.TypeDeclaration) *ViewedResultTypeData {
	// collect result type views
	rt := att.Type.(*expr.ResultTypeExpr)
	isarr := expr.IsArray(att.Type)
	var viewName string
	if !rt.HasMultipleViews() {
		viewName = expr.DefaultView
	}
	if v, ok := att.Meta.Last(expr.ViewMetaKey); ok {
		viewName = v
	}
	projectedDeclaration := viewResolver.userType(viewResolver.currentPath, projected.Type.(expr.UserType))
	views := buildViews(rt, declaration.Name())

	// build validation data
	resvar, _, serviceRef := serviceTypeData(att, serviceResolver)
	projT := wrapProjected(projected.Type.(expr.UserType))
	wrapperResolver := viewResolver.bindDerived(projT, codegen.NewViewedResultTypeID(rt))
	resref := wrapperResolver.refDeclaration(declaration, att.Type)
	data := map[string]any{
		"Projected": projectedDeclaration.Name(),
		"ArgVar":    "result",
		"Source":    "result",
		"Views":     views,
		"IsViewed":  true,
	}
	buf := &bytes.Buffer{}
	if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	name := "Validate" + resvar
	validate := &ValidateData{
		Name:        name,
		Description: fmt.Sprintf("%s runs the validations defined on the viewed result type %s.", name, resvar),
		Ref:         resref,
		Validate:    buf.String(),
	}

	// build constructor to initialize viewed result type from result type
	serviceViewResolver := wrapperResolver.withOutputPackage(serviceResolver.outputPath)
	vresref := serviceViewResolver.refDeclaration(declaration, att.Type)
	data = map[string]any{
		"ToViewed":      true,
		"ArgVar":        "res",
		"ReturnVar":     "vres",
		"Views":         views,
		"ReturnTypeRef": vresref,
		"IsCollection":  isarr,
		"TargetType":    serviceViewResolver.Name(&expr.AttributeExpr{Type: projT}, "", false, true),
		"InitName":      "new" + projectedDeclaration.Name(),
	}
	buf = &bytes.Buffer{}
	if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	name = "NewViewed" + resvar
	init := &InitData{
		Name:        name,
		Description: fmt.Sprintf("%s initializes viewed result type %s from result type %s using the given view.", name, resvar, resvar),
		Args: []*InitArgData{
			{Name: "res", Ref: serviceRef},
			{Name: "view", Ref: "string"},
		},
		ReturnTypeRef: vresref,
		Code:          buf.String(),
	}

	// build constructor to initialize result type from viewed result type
	resref = serviceRef
	data = map[string]any{
		"ToResult":      true,
		"ArgVar":        "vres",
		"ReturnVar":     "res",
		"Views":         views,
		"ReturnTypeRef": resref,
		"InitName":      "new" + resvar,
	}
	buf = &bytes.Buffer{}
	if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	name = "New" + resvar
	resinit := &InitData{
		Name:          name,
		Description:   fmt.Sprintf("%s initializes result type %s from viewed result type %s.", name, resvar, resvar),
		Args:          []*InitArgData{{Name: "vres", Ref: vresref}},
		ReturnTypeRef: resref,
		Code:          buf.String(),
	}

	return &ViewedResultTypeData{
		UserTypeData: &UserTypeData{
			Declaration: declaration,
			Name:        resvar,
			Description: fmt.Sprintf("%s is the viewed result type that is projected based on a view.", resvar),
			VarName:     resvar,
			Def:         wrapperResolver.Def(projT.Attribute(), false, true),
			Ref:         resref,
			Type:        projT,
		},
		FullName:     serviceViewResolver.Name(&expr.AttributeExpr{Type: projT}, "", false, true),
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

// buildViewConversions builds the data to generate the constructor code that
// converts between a result type and its projected type, one constructor per
// view. When toResult is true the constructors initialize the result type from
// the projected type, otherwise they project the result type to the projected
// type based on the view.
func buildViewConversions(projected, att *expr.AttributeExpr, serviceResolver, viewResolver *declarationResolver, toResult bool) []*InitData {
	vrt := att.Type.(*expr.ResultTypeExpr)
	if toResult {
		vrt = projected.Type.(*expr.ResultTypeExpr)
	}
	pobj := expr.AsObject(projected.Type)
	parr := expr.AsArray(projected.Type)
	if parr != nil {
		// result type collection
		pobj = expr.AsObject(parr.ElemType.Type)
	}

	init := make([]*InitData, 0, len(vrt.Views))
	serviceName, _, serviceRef := serviceTypeData(att, serviceResolver)
	projectedType := projected.Type.(expr.UserType)
	projectedDeclaration := viewResolver.userType(viewResolver.currentPath, projectedType)
	serviceViewResolver := viewResolver.withOutputPackage(serviceResolver.outputPath)
	for _, view := range vrt.Views {
		var typ expr.DataType
		obj := &expr.Object{}
		walkViewAttrs(pobj, view, func(name string, att, _ *expr.AttributeExpr) {
			obj.Set(name, att)
		})
		typ = obj
		if parr != nil {
			ename := parr.ElemType.Type.Name()
			if toResult {
				ename = viewResolver.Name(parr.ElemType, "", false, true)
			}
			typ = &expr.Array{ElemType: &expr.AttributeExpr{
				Type: &expr.ResultTypeExpr{
					UserTypeExpr: &expr.UserTypeExpr{
						AttributeExpr: &expr.AttributeExpr{Type: obj},
						TypeName:      ename,
					},
				},
			}}
		}
		wname := projected.Type.Name()
		if toResult {
			wname = projectedDeclaration.Name()
		}
		// viewed is the projected type narrowed down to the view attributes.
		viewed := &expr.AttributeExpr{
			Type: &expr.ResultTypeExpr{
				UserTypeExpr: &expr.UserTypeExpr{
					AttributeExpr: &expr.AttributeExpr{Type: typ},
					TypeName:      wname,
				},
				Views:      vrt.Views,
				Identifier: vrt.Identifier,
			},
		}

		viewedType := viewed.Type.(expr.UserType)
		viewIdentity, ok := viewResolver.derived[projectedType.Origin()]
		if !ok {
			panic(fmt.Sprintf("projected type %q has no planned derived identity", projectedType.Name())) // bug
		}
		viewedResolver := serviceViewResolver.bindDerived(viewedType, viewIdentity)
		if projectedArray := expr.AsArray(projected.Type); projectedArray != nil {
			projectedElement := projectedArray.ElemType.Type.(expr.UserType)
			elementIdentity, ok := viewResolver.derived[projectedElement.Origin()]
			if !ok {
				panic(fmt.Sprintf("projected element type %q has no planned derived identity", projectedElement.Name())) // bug
			}
			viewedElement := expr.AsArray(viewed.Type).ElemType.Type.(expr.UserType)
			viewedResolver = viewedResolver.bindDerived(viewedElement, elementIdentity)
		}
		if toResult {
			srcCtx := declarationContext(viewedResolver, true)
			tgtCtx := declarationContext(serviceResolver, false)
			resvar := serviceName
			name := "new" + resvar
			if view.Name != expr.DefaultView {
				name += codegen.Goify(view.Name, true)
			}
			elementInit := ""
			if parr != nil {
				serviceElement := expr.AsArray(att.Type).ElemType
				serviceElementResolver := serviceResolver.Enter(serviceElement).(*declarationResolver)
				elementInit = serviceElementResolver.userType(
					serviceElementResolver.currentPath,
					serviceElement.Type.(expr.UserType),
				).Name()
			}
			code, helpers := buildConstructorCode(
				viewed,
				att,
				"vres",
				"res",
				srcCtx,
				tgtCtx,
				view.Name,
				elementInit,
				serviceResolver.declarationName,
			)
			init = append(init, &InitData{
				Name:          name,
				Description:   fmt.Sprintf("%s converts projected type %s to service type %s.", name, resvar, resvar),
				Args:          []*InitArgData{{Name: "vres", Ref: serviceViewResolver.Ref(projected, "")}},
				ReturnTypeRef: serviceRef,
				Code:          code,
				Helpers:       helpers,
			})
		} else {
			srcCtx := declarationContext(serviceResolver, false)
			tgtCtx := declarationContext(viewedResolver, true)
			tname := projectedDeclaration.Name()
			name := "new" + tname
			if view.Name != expr.DefaultView {
				name += codegen.Goify(view.Name, true)
			}
			elementInit := ""
			if parr != nil {
				projectedElement := parr.ElemType.Type.(expr.UserType)
				elementInit = viewResolver.userType(viewResolver.currentPath, projectedElement).Name()
			}
			code, helpers := buildConstructorCode(
				att,
				viewed,
				"res",
				"vres",
				srcCtx,
				tgtCtx,
				view.Name,
				elementInit,
				viewedResolver.declarationName,
			)
			init = append(init, &InitData{
				Name:          name,
				Description:   fmt.Sprintf("%s projects result type %s to projected type %s using the %q view.", name, serviceName, tname, view.Name),
				Args:          []*InitArgData{{Name: "res", Ref: serviceRef}},
				ReturnTypeRef: serviceViewResolver.Ref(projected, ""),
				Code:          code,
				Helpers:       helpers,
			})
		}
	}
	return init
}

// buildValidations builds the data required to generate validations for the
// projected types.
func buildValidations(projected *expr.AttributeExpr, resolver *declarationResolver) []*ValidateData {
	ut := projected.Type.(expr.UserType)
	tname := resolver.Name(projected, "", false, true)
	var validations []*ValidateData
	if rt, isrt := ut.(*expr.ResultTypeExpr); isrt {
		// for result types we create a validation function containing view
		// specific validation logic for each view
		arr := expr.AsArray(projected.Type)
		for _, view := range rt.Views {
			data := map[string]any{
				"Projected":    tname,
				"ArgVar":       "result",
				"Source":       "result",
				"IsCollection": arr != nil,
			}
			var vn string
			name := "Validate" + tname
			if view.Name != expr.DefaultView {
				vn = codegen.Goify(view.Name, true)
				name += vn
			}

			if arr != nil {
				// dealing with an array type
				data["Source"] = "item"
				data["ValidateVar"] = "Validate" + resolver.Name(arr.ElemType, "", false, true) + vn
			} else {
				var fields []map[string]any
				o := &expr.Object{}
				walkViewAttrs(expr.AsObject(projected.Type), view, func(name string, attr, vatt *expr.AttributeExpr) {
					if rt, ok := attr.Type.(*expr.ResultTypeExpr); ok {
						// use explicitly specified view (if any) for the attribute,
						// otherwise use default
						vw := ""
						if v, ok := vatt.Meta.Last(expr.ViewMetaKey); ok && v != expr.DefaultView {
							vw = v
						}
						fields = append(fields, map[string]any{
							"Name":        name,
							"ValidateVar": "Validate" + resolver.Name(attr, "", false, true) + codegen.Goify(vw, true),
							"IsRequired":  rt.Attribute().IsRequired(name),
						})
					} else {
						o.Set(name, attr)
					}
				})
				ctx := declarationContext(resolver, !expr.IsPrimitive(projected.Type))
				data["Validate"] = codegen.ValidationCode(&expr.AttributeExpr{Type: o, Validation: rt.Validation}, rt, ctx, true, false, true, "result")
				data["Fields"] = fields
			}

			buf := &bytes.Buffer{}
			if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
				panic(err) // bug
			}

			validations = append(validations, &ValidateData{
				Name:        name,
				Description: fmt.Sprintf("%s runs the validations defined on %s using the %q view.", name, tname, view.Name),
				Ref:         resolver.Ref(projected, ""),
				Validate:    buf.String(),
			})
		}
	} else {
		// for a user type or a result type with single view, we generate only one validation
		// function containing the validation logic
		name := "Validate" + tname
		ctx := declarationContext(resolver, !expr.IsPrimitive(projected.Type))
		validations = append(validations, &ValidateData{
			Name:        name,
			Description: fmt.Sprintf("%s runs the validations defined on %s.", name, tname),
			Ref:         resolver.Ref(projected, ""),
			Validate:    codegen.ValidationCode(ut.Attribute(), ut, ctx, true, expr.IsAlias(ut), true, "result"),
		})
	}
	return validations
}

// buildConstructorCode builds the transformation code to create a projected
// type from a service type and vice versa.
//
// source and target contains the projected/service contextual attributes
//
// sourceVar and targetVar contains the variable name that holds the source and
// target data structures in the transformation code.
//
// view is used to generate the constructor function name.
func buildConstructorCode(src, tgt *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *codegen.AttributeContext, view, elementInitName string, nestedInitName func(*expr.AttributeExpr) string) (string, []*codegen.TransformFunctionData) {
	var (
		helpers []*codegen.TransformFunctionData
		buf     bytes.Buffer
	)
	rt := src.Type.(*expr.ResultTypeExpr)
	arr := expr.AsArray(tgt.Type)

	data := map[string]any{
		"ArgVar":       sourceVar,
		"ReturnVar":    targetVar,
		"IsCollection": arr != nil,
		"TargetType":   targetCtx.Scope.Name(tgt, targetCtx.Pkg(tgt), targetCtx.Pointer, targetCtx.UseDefault),
	}

	if arr != nil {
		// result type collection
		init := "new" + elementInitName
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

	// build code for target with no result types
	code, helpers, err := codegen.GoTransform(src, tatt, sourceVar, targetVar, sourceCtx, targetCtx, "transform", true)
	if err != nil {
		panic(err) // bug
	}
	data["Code"] = code

	fields := make([]map[string]any, 0, len(*targetRTs))
	// iterate through the result types found in the target and add the
	// code to initialize them
	for _, nat := range *targetRTs {
		finit := "new" + nestedInitName(nat.Attribute)
		if view != "" {
			v := ""
			if vatt := rt.View(view).Find(nat.Name); vatt != nil {
				if attv, ok := vatt.Meta.Last(expr.ViewMetaKey); ok && attv != expr.DefaultView {
					// view is explicitly set for the result type on the attribute
					v = attv
				}
			}
			finit += codegen.Goify(v, true)
		}
		fields = append(fields, map[string]any{
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

// removeMeta removes the meta attributes from the given attribute. This is
// needed to make sure that any field name overriding is removed when
// generating protobuf types (as protogen itself won't honor these overrides).
func removeMeta(att *expr.AttributeExpr) {
	if err := codegen.Walk(att, func(a *expr.AttributeExpr) error {
		delete(a.Meta, "struct:pkg:path")
		return nil
	}); err != nil {
		panic(err) // bug
	}
}
