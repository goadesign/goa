package codegen

import (
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

// GRPCServices holds the data computed from the design needed to generate the
// transport code of the services.
var GRPCServices = make(ServicesData)

type (
	// ServicesData encapsulates the data computed from the expr.
	ServicesData map[string]*ServiceData

	// ServiceData contains the data used to render the code related to a
	// single service.
	ServiceData struct {
		// Service contains the related service data.
		Service *service.Data
		// PkgName is the name of the generated package in *.pb.go.
		PkgName string
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// Endpoints describes the gRPC service endpoints.
		Endpoints []*EndpointData
		// Messages describes the message data for this service.
		Messages []*service.UserTypeData
		// ServerStruct is the name of the gRPC server struct.
		ServerStruct string
		// ClientStruct is the name of the gRPC client struct,
		ClientStruct string
		// ServerInit is the name of the constructor of the server struct.
		ServerInit string
		// ClientInit is the name of the constructor of the client struct.
		ClientInit string
		// ServerInterface is the name of the gRPC server interface implemented
		// by the service.
		ServerInterface string
		// ClientInterface is the name of the gRPC client interface implemented
		// by the service.
		ClientInterface string
		// ClientInterfaceInit is the name of the client constructor function in
		// the generated pb.go package.
		ClientInterfaceInit string
		// TransformHelpers is the list of transform functions required by the
		// constructors.
		TransformHelpers []*codegen.TransformFunctionData
		// Validations is the validation logic for gRPC messages.
		Validations []*ValidationData
		// Scope is the name scope for protocol buffers
		Scope *codegen.NameScope
	}

	// EndpointData contains the data used to render the code related to
	// gRPC endpoint.
	EndpointData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// PkgName is the name of the generated package in *.pb.go.
		PkgName string
		// ServicePkgName is the name of the service package name.
		ServicePkgName string
		// Method is the data for the underlying method expression.
		Method *service.MethodData
		// PayloadRef is the fully qualified reference to the method payload.
		PayloadRef string
		// ResultRef is the fully qualified reference to the method result.
		ResultRef string
		// ViewedResultRef is the fully qualified reference to the viewed result.
		ViewedResultRef string
		// Request is the gRPC request data.
		Request *RequestData
		// Response is the gRPC response data.
		Response *ResponseData
		// MetadataSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request metadata.
		MetadataSchemes []*service.SchemeData
		// MessageSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request message.
		MessageSchemes []*service.SchemeData
		// Errors describes the method gRPC errors.
		Errors []*ErrorData

		// server side

		// ServerStruct is the name of the gRPC server struct.
		ServerStruct string
		// ServerInterface is the name of the gRPC server interface implemented
		// by the service.
		ServerInterface string
		// ServerStream is the server stream data.
		ServerStream *StreamData

		// client side

		// ClientStruct is the name of the gRPC client struct,
		ClientStruct string
		// ClientInterface is the name of the gRPC client interface implemented
		// by the service.
		ClientInterface string
		// ClientStream is the client stream data.
		ClientStream *StreamData
	}

	// MetadataData describes a gRPC metadata field.
	MetadataData struct {
		// Name is the name of the metadata key.
		Name string
		// AttributeName is the name of the corresponding attribute.
		AttributeName string
		// Description is the metadata description.
		Description string
		// FieldName is the name of the struct field that holds the
		// metadata value if any, empty string otherwise.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the metadata value.
		VarName string
		// TypeName is the name of the type.
		TypeName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Required is true if the metadata is required.
		Required bool
		// Pointer is true if and only the metadata variable is a pointer.
		Pointer bool
		// StringSlice is true if the metadata value type is array of strings.
		StringSlice bool
		// Slice is true if the metadata value type is an array.
		Slice bool
		// MapStringSlice is true if the metadata value type is a map of string
		// slice.
		MapStringSlice bool
		// Map is true if the metadata value type is a map.
		Map bool
		// Type describes the datatype of the variable value. Mainly
		// used for conversion.
		Type expr.DataType
		// Validate contains the validation code if any.
		Validate string
		// DefaultValue contains the default value if any.
		DefaultValue interface{}
		// Example is an example value.
		Example interface{}
	}

	// ErrorData contains the error information required to generate the
	// transport decode (client) and encode (server) code.
	ErrorData struct {
		// StatusCode is the response gRPC status code.
		StatusCode string
		// Name is the error name.
		Name string
		// Ref is a reference to the error type.
		Ref string
		// Response is the error response data.
		Response *ResponseData
	}

	// RequestData describes a gRPC request.
	RequestData struct {
		// Description is the request description.
		Description string
		// Message is the gRPC request message.
		Message *service.UserTypeData
		// Metadata is the request metadata.
		Metadata []*MetadataData
		// ServerConvert is the request data with constructor function to
		// initialize the method payload type from the generated payload type in
		// *.pb.go.
		ServerConvert *ConvertData
		// ClientConvert is the request data with constructor function to
		// initialize the generated payload type in *.pb.go from the
		// method payload.
		ClientConvert *ConvertData
		// CLIArgs is the list of arguments for the command-line client.
		// This is set only for the client side.
		CLIArgs []*InitArgData
	}

	// ResponseData describes a gRPC success or error response.
	ResponseData struct {
		// StatusCode is the return code of the response.
		StatusCode string
		// Description is the response description.
		Description string
		// Message is the gRPC response message.
		Message *service.UserTypeData
		// Headers is the response header metadata.
		Headers []*MetadataData
		// Trailers is the response trailer metadata.
		Trailers []*MetadataData
		// ServerConvert is the type data with constructor function to
		// initialize the generated response type in *.pb.go from the
		// method result type or the projected result type.
		ServerConvert *ConvertData
		// ClientConvert is the type data with constructor function to
		// initialize the method result type or the projected result type
		// from the generated response type in *.pb.go.
		ClientConvert *ConvertData
	}

	// ConvertData contains the data to convert source type to a target type.
	// For request type, it contains data to transform gRPC request type to the
	// corresponding payload type (server) and vice versa (client).
	// For response type, it contains data to transform gRPC response type to the
	// corresponding result type (client) and vice versa (server).
	ConvertData struct {
		// SrcName is the fully qualified name of the source type.
		SrcName string
		// SrcRef is the fully qualified reference to the source type.
		SrcRef string
		// TgtName is the fully qualified name of the target type.
		TgtName string
		// TgtRef is the fully qualified reference to the target type.
		TgtRef string
		// Init contains the data required to render the constructor if any
		// to transform the source type to a target type.
		Init       *InitData
		Validation *ValidationData
	}

	// ValidationData contains the data necessary to render the validation
	// function.
	ValidationData struct {
		// Name is the validation function name.
		Name string
		// Def is the validation function definition.
		Def string
		// VarName is the name of the argument.
		ArgName string
		// SrcName is the fully qualified name of the type being validated.
		SrcName string
		// SrcRef is the fully qualified reference to the type being validated.
		SrcRef string
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Name is the constructor function name.
		Name string
		// Description is the function description.
		Description string
		// Args is the list of constructor arguments.
		Args []*InitArgData
		// ReturnVarName is the name of the variable to be returned.
		ReturnVarName string
		// ReturnTypeRef is the qualified (including the package name)
		// reference to the return type.
		ReturnTypeRef string
		// ReturnIsStruct is true if the return type is a struct.
		ReturnIsStruct bool
		// Code is the transformation code.
		Code string
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Description is the argument description.
		Description string
		// Reference to the argument, e.g. "&body".
		Ref string
		// FieldName is the name of the data structure field that should
		// be initialized with the argument if any.
		FieldName string
		// TypeName is the argument type name.
		TypeName string
		// TypeRef is the argument type reference.
		TypeRef string
		// Pointer is true if a pointer to the arg should be used.
		Pointer bool
		// Required is true if the arg is required to build the payload.
		Required bool
		// DefaultValue is the default value of the arg.
		DefaultValue interface{}
		// Validate contains the validation code for the argument
		// value if any.
		Validate string
		// Example is a example value
		Example interface{}
	}

	// StreamData contains data to render the stream struct type that implements
	// the service stream interface.
	StreamData struct {
		// VarName is the name of the struct type.
		VarName string
		// Type is the stream type (client or server).
		Type string
		// ServiceInterface is the service interface that the struct implements.
		ServiceInterface string
		// Interface is the stream interface in *.pb.go stored in the struct.
		Interface string
		// Endpoint is the streaming endpoint data.
		Endpoint *EndpointData
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendRef is the fully	qualified reference to the type sent across the
		// stream.
		SendRef string
		// SendConvert is the type sent through the stream. It contains the
		// constructor to convert the service send type to the type expected by
		// the gRPC send type (in *.pb.go)
		SendConvert *ConvertData
		// RecvConvert is the type received through the stream. It contains the
		// constructor to convert the gRPC type (in *.pb.go) to the service receive
		// type.
		RecvConvert *ConvertData
		// RecvName is the name of the receive function.
		RecvName string
		// RecvDesc is the description for the recv function.
		RecvDesc string
		// RecvRef is the fully	qualified reference to the type received from the
		// stream.
		RecvRef string
		// MustClose indicates whether to generate the Close() function
		// for the stream.
		MustClose bool
	}
)

const (
	// pbPkgName is the directory name where the .proto file is generated and
	// compiled.
	pbPkgName = "pb"
)

// Get retrieves the transport data for the service with the given name
// computing it if needed. It returns nil if there is no service with the given
// name.
func (d ServicesData) Get(name string) *ServiceData {
	if data, ok := d[name]; ok {
		return data
	}
	service := expr.Root.API.GRPC.Service(name)
	if service == nil {
		return nil
	}
	d[name] = d.analyze(service)
	return d[name]
}

// Endpoint returns the service method transport data for the endpoint with the
// given name, nil if there isn't one.
func (sd *ServiceData) Endpoint(name string) *EndpointData {
	for _, ed := range sd.Endpoints {
		if ed.Method.Name == name {
			return ed
		}
	}
	return nil
}

// HasUnaryEndpoint returns true if the service has at least one unary endpoint.
func (sd *ServiceData) HasUnaryEndpoint() bool {
	for _, ed := range sd.Endpoints {
		if ed.ServerStream == nil {
			return true
		}
	}
	return false
}

// HasStreamingEndpoint returns true if the service has at least one streaming
// endpoint.
func (sd *ServiceData) HasStreamingEndpoint() bool {
	for _, ed := range sd.Endpoints {
		if ed.ServerStream != nil {
			return true
		}
	}
	return false
}

// ValidationFor returns the validation data (if any) for the given message.
func (sd *ServiceData) ValidationFor(name string) *ValidationData {
	for _, v := range sd.Validations {
		if v.SrcName == name {
			return v
		}
	}
	return nil
}

// analyze creates the data necessary to render the code of the given service.
func (d ServicesData) analyze(gs *expr.GRPCServiceExpr) *ServiceData {
	var (
		sd      *ServiceData
		seen    map[string]struct{}
		svcVarN string

		svc   = service.Services.Get(gs.Name())
		scope = codegen.NewNameScope()
	)
	{
		svcVarN = scope.HashedUnique(gs.ServiceExpr, codegen.Goify(svc.Name, true))
		sd = &ServiceData{
			Service:             svc,
			Name:                svcVarN,
			Description:         svc.Description,
			PkgName:             pbPkgName,
			ServerStruct:        "Server",
			ClientStruct:        "Client",
			ServerInit:          "New",
			ClientInit:          "NewClient",
			ServerInterface:     svcVarN + "Server",
			ClientInterface:     svcVarN + "Client",
			ClientInterfaceInit: fmt.Sprintf("%s.New%sClient", pbPkgName, svcVarN),
			Scope:               scope,
		}
		seen = make(map[string]struct{})
	}
	for _, e := range gs.GRPCEndpoints {
		en := protoBufify(e.Name(), true)

		// convert request and response types to protocol buffer message types
		makeProtoBufMessage(e.Request, en+"Request", sd.Scope)
		if e.MethodExpr.StreamingPayload.Type != expr.Empty {
			makeProtoBufMessage(e.StreamingRequest, en+"StreamingRequest", sd.Scope)
		}
		makeProtoBufMessage(e.Response.Message, en+"Response", sd.Scope)

		// collect all the nested messages and return the top-level message
		collect := func(att *expr.AttributeExpr) *service.UserTypeData {
			msgs := collectMessages(att, sd, seen)
			sd.Messages = append(sd.Messages, msgs...)
			return msgs[0]
		}

		var (
			payloadRef      string
			resultRef       string
			viewedResultRef string
			errors          []*ErrorData

			md = svc.Method(e.Name())
		)
		{
			if e.MethodExpr.Payload.Type != expr.Empty {
				payloadRef = svc.Scope.GoFullTypeRef(e.MethodExpr.Payload, svc.PkgName)
			}
			if e.MethodExpr.Result.Type != expr.Empty {
				resultRef = svc.Scope.GoFullTypeRef(e.MethodExpr.Result, svc.PkgName)
			}
			if md.ViewedResult != nil {
				viewedResultRef = md.ViewedResult.FullRef
			}
			errors = buildErrorsData(e, sd)
		}

		// build request data
		var (
			request *RequestData
			reqMD   []*MetadataData

			payload = service.TypeContext(e.MethodExpr.Payload, svc.PkgName, svc.Scope)
			req     = protoBufContext(e.Request, sd.PkgName, sd.Scope)
		)
		{
			reqMD = extractMetadata(e.Metadata, payload, svc.Scope)
			request = &RequestData{
				Description:   e.Request.Description,
				Metadata:      reqMD,
				ServerConvert: buildRequestConvertData(req, payload, reqMD, e, sd, true),
				ClientConvert: buildRequestConvertData(req, payload, reqMD, e, sd, false),
			}
			if obj := expr.AsObject(e.Request.Type); len(*obj) > 0 {
				// add the request message as the first argument to the CLI
				request.CLIArgs = append(request.CLIArgs, &InitArgData{
					Name:     "message",
					Ref:      "message",
					TypeName: protoBufGoFullTypeName(e.Request, sd.PkgName, sd.Scope),
					TypeRef:  protoBufGoFullTypeRef(e.Request, sd.PkgName, sd.Scope),
					Example:  e.Request.Example(expr.Root.API.Random()),
				})
			}
			// pass the metadata as arguments to client CLI args
			for _, m := range reqMD {
				request.CLIArgs = append(request.CLIArgs, &InitArgData{
					Name:      m.VarName,
					Ref:       m.VarName,
					FieldName: m.FieldName,
					TypeName:  m.TypeName,
					TypeRef:   m.TypeRef,
					Pointer:   m.Pointer,
					Required:  m.Required,
					Validate:  m.Validate,
					Example:   m.Example,
				})
			}
			if e.StreamingRequest.Type != expr.Empty {
				request.Message = collect(e.StreamingRequest)
			} else {
				request.Message = collect(e.Request)
			}
		}

		// build response data
		var (
			response *ResponseData
			hdrs     []*MetadataData
			trlrs    []*MetadataData

			resp   = protoBufContext(e.Response.Message, sd.PkgName, sd.Scope)
			result = resultContext(e, sd)
		)
		{
			hdrs = extractMetadata(e.Response.Headers, result, svc.Scope)
			trlrs = extractMetadata(e.Response.Trailers, result, svc.Scope)
			response = &ResponseData{
				StatusCode:    statusCodeToGRPCConst(e.Response.StatusCode),
				Description:   e.Response.Description,
				Headers:       hdrs,
				Trailers:      trlrs,
				ServerConvert: buildResponseConvertData(resp, result, hdrs, trlrs, e, sd, true),
				ClientConvert: buildResponseConvertData(resp, result, hdrs, trlrs, e, sd, false),
			}
			// If the endpoint is a streaming endpoint, no message is returned
			// by gRPC. Hence, no need to set response message.
			if e.Response.Message.Type != expr.Empty || !e.MethodExpr.IsStreaming() {
				response.Message = collect(e.Response.Message)
			}
		}

		// gather security requirements
		var (
			msgSch []*service.SchemeData
			metSch []*service.SchemeData
		)
		{
			for _, req := range e.Requirements {
				for _, sch := range req.Schemes {
					s := service.Scheme(md.Requirements, sch.SchemeName).Dup()
					s.In = sch.In
					switch s.In {
					case "message":
						msgSch = service.AppendScheme(msgSch, s)
					default:
						metSch = service.AppendScheme(metSch, s)
					}
				}
			}
		}
		ed := &EndpointData{
			ServiceName:     svc.Name,
			PkgName:         sd.PkgName,
			ServicePkgName:  svc.PkgName,
			Method:          md,
			PayloadRef:      payloadRef,
			ResultRef:       resultRef,
			ViewedResultRef: viewedResultRef,
			Request:         request,
			Response:        response,
			MessageSchemes:  msgSch,
			MetadataSchemes: metSch,
			Errors:          errors,
			ServerStruct:    sd.ServerStruct,
			ServerInterface: sd.ServerInterface,
			ClientStruct:    sd.ClientStruct,
			ClientInterface: sd.ClientInterface,
		}
		sd.Endpoints = append(sd.Endpoints, ed)
		if e.MethodExpr.IsStreaming() {
			ed.ServerStream = buildStreamData(e, sd, true)
			ed.ClientStream = buildStreamData(e, sd, false)
		}
	}
	return sd
}

// collectMessages recurses through the attribute to gather all the messages.
func collectMessages(at *expr.AttributeExpr, sd *ServiceData, seen map[string]struct{}) (data []*service.UserTypeData) {
	if at == nil {
		return
	}
	collect := func(at *expr.AttributeExpr) []*service.UserTypeData {
		return collectMessages(at, sd, seen)
	}
	switch dt := at.Type.(type) {
	case expr.UserType:
		if _, ok := seen[dt.Name()]; ok {
			return nil
		}
		name := protoBufMessageName(at, sd.Scope)
		ref := protoBufGoFullTypeRef(at, sd.PkgName, sd.Scope)
		att := dt.Attribute()
		if rt, ok := dt.(*expr.ResultTypeExpr); ok {
			if a := unwrapAttr(expr.DupAtt(rt.Attribute())); expr.IsArray(a.Type) {
				// result type collection
				att = &expr.AttributeExpr{Type: expr.AsObject(rt)}
			}
		}
		data = append(data, &service.UserTypeData{
			Name:        dt.Name(),
			VarName:     name,
			Description: dt.Attribute().Description,
			Def:         protoBufMessageDef(att, sd.Scope),
			Ref:         ref,
			Type:        dt,
		})
		ca := protoBufContext(at, "", sd.Scope)
		if vDef := codegen.RecursiveValidationCode(ca, "message"); vDef != "" {
			sd.Validations = append(sd.Validations, &ValidationData{
				Name:    "Validate" + name,
				Def:     vDef,
				ArgName: "message",
				SrcName: name,
				SrcRef:  ref,
			})
		}
		seen[dt.Name()] = struct{}{}
		data = append(data, collect(att)...)
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

// buildRequestConvertData builds the convert data for the server and client
// requests.
//	* server side - converts generated gRPC request type in *.pb.go and the
//									gRPC metadata to method payload type.
//	* client side - converts method payload type to generated gRPC request
//									type in *.pb.go.
//
// svr param indicates that the convert data is generated for server side.
func buildRequestConvertData(request, payload *codegen.ContextualAttribute, md []*MetadataData, e *expr.GRPCEndpointExpr, sd *ServiceData, svr bool) *ConvertData {
	// Server-side: No need to build convert data if payload is not an object
	// type and request message is empty.
	if (svr && isEmpty(e.Request.Type) && !expr.IsObject(e.MethodExpr.Payload.Type)) ||
		// Client-side: No need to build convert data if streaming payload since
		// all attributes in method payload is encoded into request metadata.
		(!svr && e.MethodExpr.IsPayloadStreaming()) {
		return nil
	}

	var (
		svc = sd.Service
	)

	if svr {
		// server side
		var (
			data        *InitData
			validations *ValidationData
		)
		{
			data = buildInitData(request, payload, "message", "payload", false, sd)
			validations = sd.ValidationFor(protoBufMessageName(request.Attribute.Expr(), sd.Scope))
			data.Description = fmt.Sprintf("%s builds the payload of the %q endpoint of the %q service from the gRPC request type.", data.Name, e.Name(), svc.Name)
			for _, m := range md {
				// pass the metadata as arguments to payload constructor in server
				data.Args = append(data.Args, &InitArgData{
					Name:      m.VarName,
					Ref:       m.VarName,
					FieldName: m.FieldName,
					TypeName:  m.TypeName,
					TypeRef:   m.TypeRef,
					Pointer:   m.Pointer,
					Required:  m.Required,
					Validate:  m.Validate,
					Example:   m.Example,
				})
			}
		}
		return &ConvertData{
			SrcName:    request.Attribute.Name(),
			SrcRef:     request.Attribute.Ref(),
			TgtName:    payload.Attribute.Name(),
			TgtRef:     payload.Attribute.Ref(),
			Init:       data,
			Validation: validations,
		}
	}

	// client side

	var (
		data *InitData
	)
	{
		data = buildInitData(payload, request, "payload", "message", true, sd)
		data.Description = fmt.Sprintf("%s builds the gRPC request type from the payload of the %q endpoint of the %q service.", data.Name, e.Name(), svc.Name)
	}
	return &ConvertData{
		SrcName: payload.Attribute.Name(),
		SrcRef:  payload.Attribute.Ref(),
		TgtName: request.Attribute.Name(),
		TgtRef:  request.Attribute.Ref(),
		Init:    data,
	}
}

// buildResponseConvertData builds the convert data for the server and client
// responses.
//	* server side - converts method result type to generated gRPC response type
//									in *.pb.go
//	* client side - converts generated gRPC response type in *.pb.go and
//									response metadata to method result type.
//
// svr param indicates that the convert data is generated for server side.
func buildResponseConvertData(response, result *codegen.ContextualAttribute, hdrs, trlrs []*MetadataData, e *expr.GRPCEndpointExpr, sd *ServiceData, svr bool) *ConvertData {
	if e.MethodExpr.IsStreaming() || (!svr && isEmpty(e.MethodExpr.Result.Type)) {
		return nil
	}

	var (
		svc = sd.Service
	)

	if svr {
		// server side

		var (
			data *InitData
		)
		{
			data = buildInitData(result, response, "result", "message", true, sd)
			data.Description = fmt.Sprintf("%s builds the gRPC response type from the result of the %q endpoint of the %q service.", data.Name, e.Name(), svc.Name)
		}
		return &ConvertData{
			SrcName: result.Attribute.Name(),
			SrcRef:  result.Attribute.Ref(),
			TgtName: response.Attribute.Name(),
			TgtRef:  response.Attribute.Ref(),
			Init:    data,
		}
	}

	// client side

	var (
		data *InitData
	)
	{
		data = buildInitData(response, result, "message", "result", false, sd)
		data.Description = fmt.Sprintf("%s builds the result type of the %q endpoint of the %q service from the gRPC response type.", data.Name, e.Name(), svc.Name)
		for _, m := range hdrs {
			// pass the headers as arguments to result constructor in client
			data.Args = append(data.Args, &InitArgData{
				Name:      m.VarName,
				Ref:       m.VarName,
				FieldName: m.FieldName,
				TypeName:  m.TypeName,
				TypeRef:   m.TypeRef,
				Pointer:   m.Pointer,
				Required:  m.Required,
				Validate:  m.Validate,
				Example:   m.Example,
			})
		}
		for _, m := range trlrs {
			// pass the trailers as arguments to result constructor in client
			data.Args = append(data.Args, &InitArgData{
				Name:      m.VarName,
				Ref:       m.VarName,
				FieldName: m.FieldName,
				TypeName:  m.TypeName,
				TypeRef:   m.TypeRef,
				Pointer:   m.Pointer,
				Required:  m.Required,
				Validate:  m.Validate,
				Example:   m.Example,
			})
		}
	}
	return &ConvertData{
		SrcName:    response.Attribute.Name(),
		SrcRef:     response.Attribute.Ref(),
		TgtName:    result.Attribute.Name(),
		TgtRef:     result.Attribute.Ref(),
		Init:       data,
		Validation: sd.ValidationFor(protoBufMessageName(response.Attribute.Expr(), sd.Scope)),
	}
}

// buildInitData builds the transformation code to convert source to target.
//
// source, target are the source and target contextual attributesr used in the
// transformation
//
// sourceVar, targetVar are the source and target variable names used in the
// transformation
//
// proto if true indicates the target type is a protocol buffer type
//
// sd is the ServiceData
//
func buildInitData(source, target *codegen.ContextualAttribute, sourceVar, targetVar string, proto bool, sd *ServiceData) *InitData {
	var (
		name     string
		isStruct bool
		code     string
		helpers  []*codegen.TransformFunctionData
		args     []*InitArgData
		err      error
	)
	{
		isStruct = expr.IsObject(target.Attribute.Expr().Type)
		n := sd.Scope.GoTypeName(target.Attribute.Expr())
		if !isStruct {
			// If target is array, map, or primitive the name will be suffixed with
			// the definition (e.g int, []string, map[int]string) which is incorrect.
			n = sd.Scope.GoTypeName(source.Attribute.Expr())
		}
		name = "New" + n
		code, helpers, err = protoBufTransform(source, target, sourceVar, targetVar, proto)
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
			return nil
		}
		sd.TransformHelpers = codegen.AppendHelpers(sd.TransformHelpers, helpers)
		if (!proto && !isEmpty(source.Attribute.Expr().Type)) || (proto && !isEmpty(target.Attribute.Expr().Type)) {
			args = []*InitArgData{
				&InitArgData{
					Name:     sourceVar,
					Ref:      sourceVar,
					TypeName: source.Attribute.Name(),
					TypeRef:  source.Attribute.Ref(),
					Example:  source.Attribute.Expr().Example(expr.Root.API.Random()),
				},
			}
		}
	}
	return &InitData{
		Name:           name,
		ReturnVarName:  targetVar,
		ReturnTypeRef:  target.Attribute.Ref(),
		ReturnIsStruct: isStruct,
		Code:           code,
		Args:           args,
	}
}

// buildErrorsData builds the error data for all the error responses in the
// endpoint expression. The response message for each error response are
// inferred from the method's error expression if not specified explicitly.
func buildErrorsData(e *expr.GRPCEndpointExpr, sd *ServiceData) []*ErrorData {
	var (
		errors []*ErrorData

		svc = sd.Service
	)
	errors = make([]*ErrorData, 0, len(e.GRPCErrors))
	for _, v := range e.GRPCErrors {
		var responseData *ResponseData
		{
			responseData = &ResponseData{
				StatusCode:  statusCodeToGRPCConst(v.Response.StatusCode),
				Description: v.Response.Description,
			}
		}
		errors = append(errors, &ErrorData{
			Name:     v.Name,
			Ref:      svc.Scope.GoFullTypeRef(v.ErrorExpr.AttributeExpr, svc.PkgName),
			Response: responseData,
		})
	}
	return errors
}

// buildStreamData builds the StreamData for the server and client streams.
//
// svr param indicates that the stream data is built for the server.
func buildStreamData(e *expr.GRPCEndpointExpr, sd *ServiceData, svr bool) *StreamData {
	var (
		varn      string
		intName   string
		svcInt    string
		sendName  string
		sendDesc  string
		sendRef   string
		sendType  *ConvertData
		recvName  string
		recvDesc  string
		recvRef   string
		recvType  *ConvertData
		mustClose bool
		typ       string

		svc      = sd.Service
		ed       = sd.Endpoint(e.Name())
		md       = ed.Method
		spayload = service.TypeContext(e.MethodExpr.StreamingPayload, svc.PkgName, svc.Scope)
		result   = resultContext(e, sd)
		request  = protoBufContext(e.StreamingRequest, sd.PkgName, sd.Scope)
		response = protoBufContext(e.Response.Message, sd.PkgName, sd.Scope)
	)
	{
		resVar := "result"
		if md.ViewedResult != nil {
			resVar = "vresult"
		}
		if svr {
			typ = "server"
			varn = md.ServerStream.VarName
			intName = fmt.Sprintf("%s.%s_%sServer", sd.PkgName, svc.StructName, md.VarName)
			svcInt = fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface)
			if e.MethodExpr.Result.Type != expr.Empty {
				sendName = md.ServerStream.SendName
				sendRef = ed.ResultRef
				sendType = &ConvertData{
					SrcName: result.Attribute.Name(),
					SrcRef:  result.Attribute.Ref(),
					TgtName: response.Attribute.Name(),
					TgtRef:  response.Attribute.Ref(),
					Init:    buildInitData(result, response, resVar, "v", true, sd),
				}
			}
			if e.MethodExpr.StreamingPayload.Type != expr.Empty {
				recvName = md.ServerStream.RecvName
				recvRef = spayload.Attribute.Ref()
				recvType = &ConvertData{
					SrcName:    request.Attribute.Name(),
					SrcRef:     request.Attribute.Ref(),
					TgtName:    spayload.Attribute.Name(),
					TgtRef:     spayload.Attribute.Ref(),
					Init:       buildInitData(request, spayload, "v", "spayload", false, sd),
					Validation: sd.ValidationFor(protoBufMessageName(request.Attribute.Expr(), sd.Scope)),
				}
			}
			mustClose = md.ServerStream.MustClose
		} else {
			typ = "client"
			varn = md.ClientStream.VarName
			intName = fmt.Sprintf("%s.%s_%sClient", sd.PkgName, svc.StructName, md.VarName)
			svcInt = fmt.Sprintf("%s.%s", svc.PkgName, md.ClientStream.Interface)
			if e.MethodExpr.StreamingPayload.Type != expr.Empty {
				sendName = md.ClientStream.SendName
				sendRef = spayload.Attribute.Ref()
				sendType = &ConvertData{
					SrcName: spayload.Attribute.Name(),
					SrcRef:  spayload.Attribute.Ref(),
					TgtName: request.Attribute.Name(),
					TgtRef:  request.Attribute.Ref(),
					Init:    buildInitData(spayload, request, "spayload", "v", true, sd),
				}
			}
			if e.MethodExpr.Result.Type != expr.Empty {
				recvName = md.ClientStream.RecvName
				recvRef = ed.ResultRef
				recvType = &ConvertData{
					SrcName:    response.Attribute.Name(),
					SrcRef:     response.Attribute.Ref(),
					TgtName:    result.Attribute.Name(),
					TgtRef:     result.Attribute.Ref(),
					Init:       buildInitData(response, result, "v", resVar, false, sd),
					Validation: sd.ValidationFor(protoBufMessageName(response.Attribute.Expr(), sd.Scope)),
				}
			}
			mustClose = md.ClientStream.MustClose
		}
		if sendType != nil {
			sendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint gRPC stream.", sendName, sendType.TgtName, md.Name)
		}
		if recvType != nil {
			recvDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint gRPC stream.", recvName, recvType.SrcName, md.Name)
		}
	}
	return &StreamData{
		VarName:          varn,
		Type:             typ,
		Interface:        intName,
		ServiceInterface: svcInt,
		Endpoint:         ed,
		SendName:         sendName,
		SendDesc:         sendDesc,
		SendRef:          sendRef,
		SendConvert:      sendType,
		RecvName:         recvName,
		RecvDesc:         recvDesc,
		RecvRef:          recvRef,
		RecvConvert:      recvType,
		MustClose:        mustClose,
	}
}

// extractMetadata collects the request/response metadata from the given
// metadata attribute and service type (payload/result).
func extractMetadata(a *expr.MappedAttributeExpr, service *codegen.ContextualAttribute, scope *codegen.NameScope) []*MetadataData {
	var metadata []*MetadataData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *expr.AttributeExpr) error {
		var (
			varn      string
			fieldName string
			pointer   bool

			arr = expr.AsArray(c.Type)
			mp  = expr.AsMap(c.Type)
			ca  = service.Dup(c, required)
		)
		{
			varn = scope.Unique(codegen.Goify(name, false))
			pointer = ca.IsPointer() && expr.IsPrimitive(c.Type)
			fieldName = codegen.Goify(name, true)
			if !expr.IsObject(service.Attribute.Expr().Type) {
				fieldName = ""
			}
		}
		metadata = append(metadata, &MetadataData{
			Name:          elem,
			AttributeName: name,
			Description:   c.Description,
			FieldName:     fieldName,
			VarName:       varn,
			Required:      required,
			Type:          c.Type,
			TypeName:      scope.GoTypeName(c),
			TypeRef:       scope.GoTypeRef(c),
			Pointer:       pointer,
			Slice:         arr != nil,
			StringSlice:   arr != nil && arr.ElemType.Type.Kind() == expr.StringKind,
			Map:           mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == expr.StringKind &&
				mp.ElemType.Type.Kind() == expr.ArrayKind &&
				expr.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == expr.StringKind,
			Validate:     codegen.RecursiveValidationCode(ca, varn),
			DefaultValue: c.DefaultValue,
			Example:      c.Example(expr.Root.API.Random()),
		})
		return nil
	})
	return metadata
}

// resultContext returns the contextual attributer for the result of the given
// endpoint.
func resultContext(e *expr.GRPCEndpointExpr, sd *ServiceData) *codegen.ContextualAttribute {
	svc := sd.Service
	md := svc.Method(e.Name())
	if md.ViewedResult != nil {
		vresAtt := expr.AsObject(md.ViewedResult.Type).Attribute("projected")
		return service.ProjectedTypeContext(vresAtt, svc.ViewsPkg, svc.ViewScope)
	}
	return service.TypeContext(e.MethodExpr.Result, svc.PkgName, svc.Scope)
}

// isEmpty returns true if given type is empty.
func isEmpty(dt expr.DataType) bool {
	if dt == expr.Empty {
		return true
	}
	if o := expr.AsObject(dt); o != nil && len(*o) == 0 {
		return true
	}
	return false
}

// input: InitData
const typeInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
{{- if .ReturnIsStruct }}
	{{- range .Args }}
		{{- if .FieldName }}
			{{ $.ReturnVarName }}.{{ .FieldName }} = {{ if .Pointer }}&{{ end }}{{ .Name }}
		{{- end }}
	{{- end }}
{{- end }}
	return {{ .ReturnVarName }}
}
`

// input: ValidationData
const validateT = `{{ printf "%s runs the validations defined on %s." .Name .SrcName | comment }}
func {{ .Name }}({{ .ArgName }} {{ .SrcRef }}) (err error) {
	{{ .Def }}
	return
}
`

// streamSendT renders the function implementing the Send method in
// stream interface.
// input: StreamData
const streamSendT = `{{ comment .SendDesc }}
func (s *{{ .VarName }}) {{ .SendName }}(res {{ .SendRef }}) error {
{{- if and .Endpoint.Method.ViewedResult (eq .Type "server") }}
	{{- if .Endpoint.Method.ViewedResult.ViewName }}
		vres := {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(res, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
	{{- else }}
		vres := {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(res, s.view)
	{{- end }}
{{- end }}
	v := {{ .SendConvert.Init.Name }}({{ if and .Endpoint.Method.ViewedResult (eq .Type "server") }}vres.Projected{{ else }}res{{ end }})
	return s.stream.{{ .SendName }}(v)
}
`

// streamRecvT renders the function implementing the Recv method in
// stream interface.
// input: StreamData
const streamRecvT = `{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvRef }}, error) {
	var res {{ .RecvRef }}
	v, err := s.stream.{{ .RecvName }}()
	if err != nil {
		return res, err
	}
{{- if and .Endpoint.Method.ViewedResult (eq .Type "client") }}
	proj := {{ .RecvConvert.Init.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }})
	vres := {{ if not .Endpoint.Method.ViewedResult.IsCollection }}&{{ end }}{{ .Endpoint.Method.ViewedResult.FullName }}{Projected: proj, View: {{ if .Endpoint.Method.ViewedResult.ViewName }}"{{ .Endpoint.Method.ViewedResult.ViewName }}"{{ else }}s.view{{ end }} }
	return {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.ResultInit.Name }}(vres), nil
{{- else }}
	return {{ .RecvConvert.Init.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
{{- end }}
}
`

// streamCloseT renders the function implementing the Close method in
// stream interface.
// input: StreamData
const streamCloseT = `
func (s *{{ .VarName }}) Close() error {
	{{ comment "nothing to do here" }}
	return nil
}
`

// streamSetViewT renders the function implementing the SetView method in
// server stream interface.
// input: StreamData
const streamSetViewT = `{{ printf "SetView sets the view." | comment }}
func (s *{{ .VarName }}) SetView(view string) {
	s.view = view
}
`
