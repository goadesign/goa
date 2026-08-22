// This file analyzes evaluated service designs into immutable render data.
// Public type declarations and references come from the frozen generated
// package catalog; mutable scopes are used only for private helper names.
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
		// StreamDeclaration is the shared JSON-RPC stream record when emitted.
		StreamDeclaration *codegen.NameDeclaration
		// EventDeclaration is the shared JSON-RPC SSE event record when emitted.
		EventDeclaration *codegen.NameDeclaration
		// ServerInterceptorsDeclaration is the server interceptor interface record.
		ServerInterceptorsDeclaration *codegen.NameDeclaration
		// ClientInterceptorsDeclaration is the client interceptor interface record.
		ClientInterceptorsDeclaration *codegen.NameDeclaration
		// ExampleStructDeclaration is the starter implementation struct record.
		ExampleStructDeclaration *codegen.NameDeclaration
		// ExampleConstructorDeclaration is the starter constructor record.
		ExampleConstructorDeclaration *codegen.NameDeclaration
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
		// viewUnions lists the sum-type unions emitted by the views package.
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
		// EventDeclaration is the exact JSON-RPC SSE event record.
		EventDeclaration *codegen.NameDeclaration
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
		// VarName is the lexical implementation type name retained during service
		// planning for transport generators.
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
		// Declaration is the exact package-level constructor record retained while
		// the service was planned.
		Declaration *codegen.NameDeclaration
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
		// Declaration is the immutable generated-package record that owns this
		// type in a service, views, or relocated package.
		Declaration *codegen.TypeDeclaration
		// Name is the type name.
		Name string
		// VarName is the corresponding Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// ErrorName is the retained Go expression returned by GoaErrorName.
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
		// MapDeclaration is the exact package-level view map record for this type.
		MapDeclaration *codegen.NameDeclaration
		// ToProjected is the exact private constructor that applies this view.
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
		// Views lists the views defined on the projected type.
		Views []*ViewData
	}

	// InitData contains the data to render a constructor to initialize service
	// types from viewed result types and vice versa.
	InitData struct {
		// Declaration is the exact package-level constructor record retained while
		// the service was planned.
		Declaration *codegen.NameDeclaration
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
		// Declaration is the exact package-level function record retained while
		// the service was planned.
		Declaration *codegen.NameDeclaration
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

	// ValidationCallData binds one nested validation call to the exact function
	// declaration that owns the rendered name.
	ValidationCallData struct {
		// Declaration is the exact package-level validator function record.
		Declaration *codegen.NameDeclaration
		// View is the selected result-type view.
		View string
		// Default reports whether View is the default result-type view.
		Default bool
	}

	// validationFieldData describes a nested result field validated by a
	// projected parent validator.
	validationFieldData struct {
		Name       string
		Call       *ValidationCallData
		IsRequired bool
	}

	// constructorFieldData binds one nested result field to the exact retained
	// private constructor called by its parent conversion.
	constructorFieldData struct {
		VarName     string
		Declaration *codegen.NameDeclaration
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
)

// linkServicesData resolves the exact service facts retained before generation
// freeze into immutable render data.
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

// Example computes attribute's example below the explicit semantic owner.
func (d *ServicesData) Example(attribute *expr.AttributeExpr, owner expr.ExampleIdentity) any {
	return attribute.Example(d.examples.At(owner))
}

// FieldExample computes attribute's example using the same stable field
// identity as the corresponding field in parent. Named user types own their
// fields globally; anonymous parents keep the caller-supplied owner.
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

// ServiceImport returns the frozen import alias for name's generated service
// package. The returned value is a copy that callers may add to one file.
func (d *ServicesData) ServiceImport(name string) *codegen.ImportSpec {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(serviceFacts.packagePath)
	return &codegen.ImportSpec{Name: spec.Name, Path: spec.Path}
}

// ViewImport returns the frozen import alias for name's generated views
// package. The returned value is a copy that callers may add to one file.
func (d *ServicesData) ViewImport(name string) *codegen.ImportSpec {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	spec := d.aliases.spec(serviceFacts.viewsPath)
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
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	return newServiceResolver(d.generation, d.aliases, serviceFacts.service, outputPackage).
		withValidators(serviceFacts.validators)
}

// ViewAttributor returns the frozen projected and viewed result declaration
// resolver for name as referenced from outputPackage.
func (d *ServicesData) ViewAttributor(name, outputPackage string) codegen.Attributor {
	serviceFacts := d.facts.serviceByID[name]
	if serviceFacts == nil {
		panic(fmt.Sprintf("service %q is not part of the analyzed design root", name))
	}
	data := d.Services[name]
	return newViewResolver(d.generation, d.aliases, serviceFacts.service, data.viewDerived).
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
