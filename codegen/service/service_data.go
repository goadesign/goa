// This file builds the values passed to service templates. Package-level Go
// names are shared by every file in the package, while each file may choose
// additional private helper names.
package service

import (
	"fmt"
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
		examples   *expr.ExampleGenerator
		aliases    *importAliases
		packages   map[*codegen.GeneratedPackage]*generatedPackageData
		facts      *rootFacts
	}

	// Data contains the data used to render the code related to a single
	// service.
	Data struct {
		// ServiceDeclaration is the exact package-level service interface record.
		ServiceDeclaration *codegen.NameDeclaration
		// AutherDeclaration is the exact package-level authorization interface
		// record. It is nil when the service has no security schemes.
		AutherDeclaration *codegen.NameDeclaration
		// APINameDeclaration is the exact package-level API name constant record.
		APINameDeclaration *codegen.NameDeclaration
		// APIVersionDeclaration is the exact package-level API version constant record.
		APIVersionDeclaration *codegen.NameDeclaration
		// ServiceNameDeclaration is the exact package-level service name constant record.
		ServiceNameDeclaration *codegen.NameDeclaration
		// MethodNamesDeclaration is the exact package-level method names variable record.
		MethodNamesDeclaration *codegen.NameDeclaration
		// EndpointsDeclaration is the exact package-level endpoint collection record.
		EndpointsDeclaration *codegen.NameDeclaration
		// NewEndpointsDeclaration is the exact endpoint constructor record.
		NewEndpointsDeclaration *codegen.NameDeclaration
		// ClientDeclaration is the exact package-level client record.
		ClientDeclaration *codegen.NameDeclaration
		// NewClientDeclaration is the exact client constructor record.
		NewClientDeclaration *codegen.NameDeclaration
		// ServerInterceptorsDeclaration is the server interceptor interface record.
		ServerInterceptorsDeclaration *codegen.NameDeclaration
		// ClientInterceptorsDeclaration is the client interceptor interface record.
		ClientInterceptorsDeclaration *codegen.NameDeclaration
		// ExampleStructDeclaration is the starter implementation struct record.
		ExampleStructDeclaration *codegen.NameDeclaration
		// ExampleConstructorDeclaration is the starter constructor record.
		ExampleConstructorDeclaration *codegen.NameDeclaration
		// ExampleServerInterceptorsConstructorDeclaration creates the starter
		// server interceptor implementation. It is nil when the service has no
		// server interceptors.
		ExampleServerInterceptorsConstructorDeclaration *codegen.NameDeclaration
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
		// VarName is the local Go variable that holds the service implementation in
		// generated starter programs.
		VarName string
		// PathName is the service name as used in file and import paths.
		PathName string
		// PkgName is the name of the package containing the generated service
		// code.
		PkgName string
		// ViewsPkg is the final views package name kept for existing plugins. It
		// is empty when the service does not generate a views package.
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
		// unions lists the values that hold one selected branch for the service.
		unions []*UnionTypeData
		// viewUnions lists the values that hold one selected branch in the views package.
		viewUnions []*UnionTypeData
		// viewedResultTypes lists all the viewed method result types.
		viewedResultTypes []*ViewedResultTypeData
		// viewDerived binds the independently rebuilt view graph to declarations
		// reserved while the service was planned.
		viewDerived map[expr.UserType]codegen.DerivedTypeID
	}

	// MethodData describes a single service method.
	MethodData struct {
		// EndpointDeclaration is the exact package-level endpoint constructor.
		EndpointDeclaration *codegen.NameDeclaration
		// EndpointInputDeclaration is the exact streaming endpoint input record.
		EndpointInputDeclaration *codegen.NameDeclaration
		// ServerStreamDeclaration is the exact server stream interface record.
		ServerStreamDeclaration *codegen.NameDeclaration
		// ClientStreamDeclaration is the exact client stream interface record.
		ClientStreamDeclaration *codegen.NameDeclaration
		// RequestDeclaration is the exact JSON-RPC request data record.
		RequestDeclaration *codegen.NameDeclaration
		// ResponseDeclaration is the exact JSON-RPC response data record.
		ResponseDeclaration *codegen.NameDeclaration
		// ServerEndpointWrapperDeclaration is the exact server endpoint wrapper.
		ServerEndpointWrapperDeclaration *codegen.NameDeclaration
		// ClientEndpointWrapperDeclaration is the exact client endpoint wrapper.
		ClientEndpointWrapperDeclaration *codegen.NameDeclaration
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
		// PayloadDeclaration supplies the generated Go type name for a named payload.
		// It is nil for primitive payloads.
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
		// StreamingPayloadDeclaration supplies the generated Go type name for a
		// named streaming payload. It is nil for primitive payloads.
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
		// StreamingResultDeclaration supplies the generated Go type name for a named
		// streaming result. It is nil for primitive results.
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
		// ResultDeclaration supplies the generated Go type name for a named result.
		// It is nil for primitive results.
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
		// HasMixedResults indicates whether the method defines Result and
		// StreamingResult separately so HTTP can return one normal response or an
		// SSE stream.
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
		// VarName is the unexported Go type name used by transport packages for this
		// stream implementation.
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

	// ErrorInitData describes an error returned by a service method.
	ErrorInitData struct {
		// Declaration is the package-level constructor submitted while the service
		// was planned. It is nil for custom errors because the service package does
		// not generate constructors for them.
		Declaration *codegen.NameDeclaration
		// Name is a read-only copy of the final constructor name kept for existing
		// plugins. It is empty for custom errors because they have no generated
		// service constructor.
		//
		// Deprecated: Use Declaration.Name().
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
		// InfoDeclaration is the exact interceptor metadata record.
		InfoDeclaration *codegen.NameDeclaration
		// PayloadDeclaration is the exact payload accessor interface when emitted.
		PayloadDeclaration *codegen.NameDeclaration
		// ResultDeclaration is the exact result accessor interface when emitted.
		ResultDeclaration *codegen.NameDeclaration
		// StreamingPayloadDeclaration is the exact streaming payload accessor interface when emitted.
		StreamingPayloadDeclaration *codegen.NameDeclaration
		// StreamingResultDeclaration is the exact streaming result accessor interface when emitted.
		StreamingResultDeclaration *codegen.NameDeclaration
		// Name is the name of the interceptor used in the generated code.
		Name string
		// DesignName is the name of the interceptor as defined in the design.
		DesignName string
		// Description is the description of the interceptor from the design.
		Description string
		// Service is the service name returned to this interceptor.
		Service string
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
		// InfoDeclaration is the private type that returns this method's name and
		// provides its field access methods.
		InfoDeclaration *codegen.NameDeclaration
		// ServerUnaryInfoDeclaration is the private call information type used by
		// the server endpoint call.
		ServerUnaryInfoDeclaration *codegen.NameDeclaration
		// ClientUnaryInfoDeclaration is the private call information type used by
		// the client endpoint call.
		ClientUnaryInfoDeclaration *codegen.NameDeclaration
		// StreamingSendInfoDeclaration is the private call information type used
		// while a stream value is sent.
		StreamingSendInfoDeclaration *codegen.NameDeclaration
		// StreamingRecvInfoDeclaration is the private call information type used
		// while a stream value is received.
		StreamingRecvInfoDeclaration *codegen.NameDeclaration
		// PayloadAccessDeclaration is the exact private payload accessor struct.
		PayloadAccessDeclaration *codegen.NameDeclaration
		// ResultAccessDeclaration is the exact private result accessor struct.
		ResultAccessDeclaration *codegen.NameDeclaration
		// StreamingPayloadAccessDeclaration is the exact private streaming payload accessor struct.
		StreamingPayloadAccessDeclaration *codegen.NameDeclaration
		// StreamingResultAccessDeclaration is the exact private streaming result accessor struct.
		StreamingResultAccessDeclaration *codegen.NameDeclaration
		// ServerWrapperDeclaration is the exact server interceptor wrapper function.
		ServerWrapperDeclaration *codegen.NameDeclaration
		// ClientWrapperDeclaration is the exact client interceptor wrapper function.
		ClientWrapperDeclaration *codegen.NameDeclaration
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
		// InterfaceDeclaration is the exact stream interface wrapped by this record.
		InterfaceDeclaration *codegen.NameDeclaration
		// WrapperDeclaration is the exact private interceptor stream wrapper struct.
		WrapperDeclaration *codegen.NameDeclaration
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
		// Declaration supplies this type's generated Go name and output package.
		Declaration *codegen.TypeDeclaration
		// Name is the type name.
		Name string
		// VarName is the corresponding Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// ErrorDescription is the authored text returned by Error.
		ErrorDescription string
		// ErrorName is the Go expression returned by GoaErrorName during planning.
		ErrorName string
		// IsServiceError reports whether this is Goa's built-in service error.
		IsServiceError bool
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

	// UnionTypeData describes a generated value that holds exactly one branch.
	UnionTypeData struct {
		// TypeDeclaration supplies the generated union type name.
		TypeDeclaration *codegen.NameDeclaration
		// KindDeclaration supplies the generated type that records the selected branch.
		KindDeclaration *codegen.NameDeclaration
		// Name is the final union type name copied for existing plugins.
		//
		// Deprecated: Use TypeDeclaration.
		Name string
		// KindName is the final selected-branch type name copied for existing plugins.
		//
		// Deprecated: Use KindDeclaration.
		KindName string
		// Fields describes each union branch.
		Fields []*UnionFieldData
		// Loc defines the file and Go package of the union type if overridden via
		// Meta. When nil the type is generated in the default service file.
		Loc *codegen.Location
		// TypeKey is the field that records the selected branch in JSON (defaults to "type").
		TypeKey string
		// ValueKey is the value field name for JSON marshaling (defaults to "value").
		ValueKey string
	}

	// UnionFieldData describes a single branch of a union.
	UnionFieldData struct {
		// Name is the branch name as defined in the DSL.
		Name string
		// KindConst is the final branch constant name copied for existing plugins.
		//
		// Deprecated: Use KindDeclaration.
		KindConst string
		// Constructor is the final branch constructor name copied for existing plugins.
		//
		// Deprecated: Use ConstructorDeclaration.
		Constructor string
		// KindDeclaration supplies the generated constant name for this branch.
		KindDeclaration *codegen.NameDeclaration
		// ConstructorDeclaration supplies the generated constructor name for this branch.
		ConstructorDeclaration *codegen.NameDeclaration
		// FieldName is the branch name used by the public As and Set methods.
		FieldName string
		// StorageName is the private struct field that stores this branch value.
		StorageName string
		// FieldType is the Go type used in the union struct field and public API.
		FieldType string
		// Nilable is true when the Go branch value can be nil even though selecting
		// a non-nil Goa OneOf branch value is required.
		Nilable bool
		// EmitPrimitiveAlias is true when the branch uses a generated primitive alias
		// that must be declared in the same file as the union type.
		EmitPrimitiveAlias bool
		// PrimitiveAliasType is the underlying Go type used by the generated branch
		// alias (for example "string" or "float64").
		PrimitiveAliasType string
		// TypeTag is the JSON "type" value that selects this branch.
		TypeTag string
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
		// UsernameIsAlias is true if the username field uses a named string type.
		UsernameIsAlias bool
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
		// PasswordIsAlias is true if the password field uses a named string type.
		PasswordIsAlias bool
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
		// CredIsAlias is true if the credential field uses a named string type.
		CredIsAlias bool
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
		// MapDeclaration is the exact package-level view map record for this type.
		MapDeclaration *codegen.NameDeclaration
		// ToProjected is the private constructor that copies only this view's
		// fields from a service result.
		ToProjected *codegen.NameDeclaration
		// ToResult is the exact private constructor that removes this view.
		ToResult *codegen.NameDeclaration
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
		// ViewsPkg is the final views package name kept for existing plugins.
		ViewsPkg string
		// Views lists the views defined on the projected type.
		Views []*ViewData
	}

	// InitData contains the data to render a constructor to initialize service
	// types from viewed result types and vice versa.
	InitData struct {
		// Declaration is the package-level constructor submitted while the service
		// was planned.
		Declaration *codegen.NameDeclaration
		// Name is a read-only copy of the final constructor name kept for existing
		// plugins.
		//
		// Deprecated: Use Declaration.Name().
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
		// Declaration is the package-level validation function submitted while the
		// service was planned.
		Declaration *codegen.NameDeclaration
		// Name is a read-only copy of the final validation function name kept for
		// existing plugins.
		//
		// Deprecated: Use Declaration.Name().
		Name string
		// Ref is the reference to the type on which the validation function
		// is defined.
		Ref string
		// Description is the description for the validation function.
		Description string
		// Validate is the validation code.
		Validate string
		// Calls lists nested validator functions called by Validate.
		Calls []*ValidationCallData
	}

	// ValidationCallData records the exact generated function called to validate
	// one nested result value.
	ValidationCallData struct {
		// Declaration is the exact package-level validator function record.
		Declaration *codegen.NameDeclaration
		// View is the selected result-type view.
		View string
		// Default reports whether View is the default result-type view.
		Default bool
	}

	// validationFieldData describes a nested result field checked by the
	// validation function for its parent's selected view.
	validationFieldData struct {
		Name       string
		Call       *ValidationCallData
		IsRequired bool
	}

	// constructorFieldData associates one child result field with the private
	// constructor called by its parent conversion.
	constructorFieldData struct {
		VarName     string
		Declaration *codegen.NameDeclaration
	}

	// unionDataKey selects one Goa OneOf definition by its generated definition
	// key and Go package path.
	unionDataKey struct {
		packagePath string
		identity    codegen.UnionDeclarationID
	}

	// userTypeDataKey distinguishes in-memory design types and the generated Go
	// declaration selected for each one.
	userTypeDataKey struct {
		origin      expr.UserType
		declaration *codegen.TypeDeclaration
	}

	// projectedTypePair records one result type rebuilt with only a view's fields
	// and the exact source declaration used to find its DerivedTypeID.
	projectedTypePair struct {
		source             expr.UserType
		projected          expr.UserType
		sourceAttribute    *expr.AttributeExpr
		projectedAttribute *expr.AttributeExpr
	}
)

// linkServicesData builds service template data from the values copied during
// planning and the Go names chosen by Generation.Freeze.
func linkServicesData(facts *rootFacts, generation *codegen.Generation, aliases *importAliases) (*ServicesData, error) {
	root := facts.root
	data := &ServicesData{
		Root:       root,
		Services:   make(map[string]*Data),
		generation: generation,
		examples:   facts.examples,
		aliases:    aliases,
		packages:   make(map[*codegen.GeneratedPackage]*generatedPackageData),
		facts:      facts,
	}
	for _, service := range facts.services {
		analyzed, err := data.analyze(service)
		if err != nil {
			return nil, err
		}
		service.data = analyzed
		data.Services[service.name] = analyzed
	}
	data.registerPackageData()
	return data, nil
}

// Example computes an example for attribute. The supplied ExampleIdentity
// selects the repeatable sequence from which the values are drawn.
func (d *ServicesData) Example(attribute *expr.AttributeExpr, owner expr.ExampleIdentity) any {
	return attribute.Example(d.examples.At(owner))
}

// FieldExample computes attribute's example from the same repeatable sequence
// as the matching field in parent. Fields of a named user type use that type's
// ExampleIdentity; fields of an anonymous parent use the caller's value.
func (d *ServicesData) FieldExample(attribute, parent *expr.AttributeExpr, name string, owner expr.ExampleIdentity) any {
	if typ, ok := parent.Type.(expr.UserType); ok {
		owner = expr.UserTypeExampleIdentity(typ)
	}
	return attribute.Example(d.examples.At(owner).Member(name))
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

// ServiceImport returns the import path and Go name used by outputPackage for
// the generated package of service name. The returned value is a copy that
// callers may add to one file.
func (d *ServicesData) ServiceImport(outputPackage, name string) *codegen.ImportSpec {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(outputPackage, serviceFacts.packagePath)
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// ViewImport returns the import path and Go name used by outputPackage for the
// views package of service name. The returned value is a copy that callers may
// add to one file.
func (d *ServicesData) ViewImport(outputPackage, name string) *codegen.ImportSpec {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(outputPackage, serviceFacts.viewsPath)
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// PackageImport returns the import path and Go name used by outputPackage for
// importPath. The returned value is a copy that callers may add to one file.
func (d *ServicesData) PackageImport(outputPackage, importPath string) *codegen.ImportSpec {
	spec := d.aliases.spec(outputPackage, importPath)
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// ServiceAttributor returns a type writer for service name as referenced from
// outputPackage. It follows explicit generated package locations and uses the
// same import names as service templates.
func (d *ServicesData) ServiceAttributor(name, outputPackage string) codegen.Attributor {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	return newServiceResolver(
		d.generation,
		d.aliases,
		serviceFacts.name,
		serviceFacts.packagePath,
		outputPackage,
	).
		withValidators(serviceFacts.validators)
}

// ViewAttributor returns a type writer for service name's result views as
// referenced from outputPackage.
func (d *ServicesData) ViewAttributor(name, outputPackage string) codegen.Attributor {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	data := d.Services[name]
	return newViewResolver(
		d.generation,
		d.aliases,
		serviceFacts.name,
		serviceFacts.viewsPath,
		data.viewDerived,
	).
		withValidators(serviceFacts.validators).
		withOutputPackage(outputPackage)
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
		UsernameIsAlias:  s.UsernameIsAlias,
		UsernameAttr:     s.UsernameAttr,
		UsernameRequired: s.UsernameRequired,
		PasswordField:    s.PasswordField,
		PasswordPointer:  s.PasswordPointer,
		PasswordIsAlias:  s.PasswordIsAlias,
		PasswordAttr:     s.PasswordAttr,
		PasswordRequired: s.PasswordRequired,
		CredField:        s.CredField,
		CredPointer:      s.CredPointer,
		CredIsAlias:      s.CredIsAlias,
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
