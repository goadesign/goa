// This file derives imports from the HTTP endpoints rendered into one
// generated file. Streaming-only files pass only their streaming endpoints.
package codegen

import (
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/expr"
)

type (
	// httpFileKind identifies one generated file whose sections have one fixed
	// import contract.
	httpFileKind uint8

	// httpGeneratedFile pairs one import contract with its generated file name.
	httpGeneratedFile struct {
		kind httpFileKind
		name string
	}

	// httpCLIFlag contains the input facts that decide one payload-builder
	// flag's conversion and omission checks.
	httpCLIFlag struct {
		attribute  *expr.AttributeExpr
		required   bool
		hasDefault bool
	}
)

const (
	httpTransportFile httpFileKind = iota + 1
	httpCodecFile
	httpTypesFile
	httpPathsFile
	httpPayloadBuilderFile
	httpWebSocketFile
	httpSSEFile
)

// httpFixedFileImports returns packages named directly by templates in one
// generated HTTP file. The same list is reserved before names are chosen and
// later written into the file header.
func httpFixedFileImports(service *expr.HTTPServiceExpr, client bool, kind httpFileKind) []*codegen.ImportSpec {
	var paths []string
	switch kind {
	case httpTransportFile:
		if client {
			paths = []string{"context", "net/http", codegen.GoaImport("").Path, codegen.GoaNamedImport("http", "goahttp").Path}
			if httpServiceUsesSSE(service) {
				paths = append(paths, "errors", "fmt", "mime")
			}
			if httpServiceSkipsRequestBody(service) || httpServiceSkipsResponseBody(service) {
				paths = append(paths, "io")
			}
			if httpServiceSkipsResponseBody(service) {
				paths = append(paths, "errors")
			}
			if serviceHasMultipartRequest(service) {
				paths = append(paths, "mime/multipart")
			}
			if len(httpWebSocketEndpoints(service)) > 0 {
				paths = append(paths, "github.com/gorilla/websocket", "time")
			}
		} else {
			paths = []string{"context", "net/http", codegen.GoaImport("").Path, codegen.GoaNamedImport("http", "goahttp").Path}
			if httpServiceHasMixedResults(service) {
				paths = append(paths, "strings")
			}
			if httpServiceSkipsResponseBody(service) {
				paths = append(paths, "bufio", "io")
			}
			if serviceHasMultipartRequest(service) {
				paths = append(paths, "mime/multipart")
			}
			if len(httpWebSocketEndpoints(service)) > 0 {
				paths = append(paths, "bufio", "github.com/gorilla/websocket")
			}
			if len(service.FileServers) > 0 {
				paths = append(paths, "fmt", "io", "path")
			}
		}
	case httpCodecFile:
		paths = []string{"net/http", codegen.GoaNamedImport("http", "goahttp").Path}
		if client || httpServerCodecUsesContext(service) {
			paths = append(paths, "context")
		}
		if client {
			paths = append(paths, "net/url")
			if httpClientCodecRestoresResponseBody(service) {
				paths = append(paths, "bytes")
			}
			if httpClientCodecUsesIO(service) {
				paths = append(paths, "io")
			}
		} else if httpServiceHasRequestBody(service) {
			paths = append(paths, "io")
		}
		if httpServiceUsesJSONRPC(service) {
			paths = append(paths, codegen.GoaImport("jsonrpc").Path)
			if !client {
				paths = append(paths, "bytes")
				if httpServiceDecodesJSONRPCParams(service) {
					paths = append(paths, "io")
				}
			}
		}
		if client && httpServiceGeneratesJSONRPCRequestID(service) {
			paths = append(paths, "github.com/google/uuid")
		}
		if httpCodecUsesGoa(service, client) {
			paths = append(paths, codegen.GoaImport("").Path)
		}
		if httpCodecUsesFmt(service) {
			paths = append(paths, "fmt")
		}
		if client && httpServiceSkipsRequestBody(service) {
			paths = append(paths, "os")
		}
		if httpCodecUsesErrors(service, client) {
			paths = append(paths, "errors")
		}
		if httpCodecUsesStrings(service, client) {
			paths = append(paths, "strings")
		}
		if httpCodecUsesStrconv(service) {
			paths = append(paths, "strconv")
		}
		if serviceHasMultipartRequest(service) {
			paths = append(paths, "mime/multipart")
		}
	case httpTypesFile:
		if httpServiceHasUnion(service) {
			paths = append(paths, "bytes", "encoding/json", "fmt", codegen.GoaImport("").Path)
		}
	case httpPathsFile:
		paths = httpPathImportPaths(service)
	case httpPayloadBuilderFile:
		return httpCLIPayloadBuilderFixedImports(service)
	case httpWebSocketFile:
		paths = []string{
			"context", "io", "net/http", "sync", "time", "github.com/gorilla/websocket",
			codegen.GoaImport("").Path, codegen.GoaNamedImport("http", "goahttp").Path,
		}
	case httpSSEFile:
		if client {
			paths = []string{
				"bytes", "context", "encoding/json", "errors", "fmt", "io", "net/http", "strconv", "strings", "sync",
				codegen.GoaNamedImport("http", "goahttp").Path,
			}
		} else {
			paths = []string{"context", "encoding/json", "fmt", "io", "net/http", "sync", "time"}
		}
		if httpServiceHasViewedSSE(service) {
			paths = append(paths, codegen.GoaImport("").Path)
		}
	}
	imports := make([]*codegen.ImportSpec, 0, len(paths))
	for _, importPath := range paths {
		switch importPath {
		case codegen.GoaImport("").Path:
			imports = append(imports, codegen.GoaImport(""))
		case codegen.GoaNamedImport("http", "goahttp").Path:
			imports = append(imports, codegen.GoaNamedImport("http", "goahttp"))
		default:
			imports = append(imports, codegen.SimpleImport(importPath))
		}
	}
	return imports
}

// httpCodecValidationImports returns the Go packages used by validation checks
// written directly into one client or server codec file.
func httpCodecValidationImports(service *expr.HTTPServiceExpr, client bool) []*codegen.ImportSpec {
	seen := make(map[string]struct{})
	var imports []*codegen.ImportSpec
	add := func(attribute *expr.AttributeExpr, policy codegen.GoLayoutPolicy) {
		for _, runtimeImport := range codegen.ValidationRuntimeImports(attribute, policy) {
			if _, ok := seen[runtimeImport.Path]; ok {
				continue
			}
			seen[runtimeImport.Path] = struct{}{}
			imports = append(imports, codegen.NewImport(runtimeImport.Name, runtimeImport.Path))
		}
	}
	addMapped := func(mapped *expr.MappedAttributeExpr) {
		if mapped == nil || mapped.IsEmpty() {
			return
		}
		if err := codegen.WalkMappedAttr(mapped, func(_ string, _ string, _ bool, attribute *expr.AttributeExpr) error {
			add(attribute, codegen.GoLayoutPolicy{})
			return nil
		}); err != nil {
			panic(err) // The callback above cannot return an error.
		}
	}
	for _, endpoint := range service.HTTPEndpoints {
		if client {
			for _, response := range endpoint.Responses {
				if response.Body != nil && response.Body.Type != expr.Empty {
					if _, named := response.Body.Type.(expr.UserType); !named {
						add(response.Body, codegen.GoLayoutPolicy{ArrayElementPointer: true})
					}
				}
				addMapped(response.Headers)
				addMapped(response.Cookies)
			}
			continue
		}
		if endpoint.Body != nil && endpoint.Body.Type != expr.Empty {
			if _, named := endpoint.Body.Type.(expr.UserType); !named {
				add(endpoint.Body, codegen.GoLayoutPolicy{
					Pointer:             !expr.IsPrimitive(endpoint.Body.Type),
					UnionPointer:        true,
					ArrayElementPointer: true,
					SumType:             true,
				})
			}
		}
		addMapped(endpoint.PathParams())
		addMapped(endpoint.QueryParams())
		addMapped(endpoint.Headers)
		addMapped(endpoint.Cookies)
		if policy := jsonRPCRequestIDPolicyFor(endpoint); policy != nil && policy.attribute != nil {
			add(policy.attribute.Attribute, codegen.GoLayoutPolicy{Pointer: policy.pointer})
		}
	}
	return imports
}

// httpCLIPayloadBuilderFixedImports returns the conversion and validation
// packages used by request flags written in one client payload-builder file.
func httpCLIPayloadBuilderFixedImports(service *expr.HTTPServiceExpr) []*codegen.ImportSpec {
	seen := make(map[string]struct{})
	var imports []*codegen.ImportSpec
	for _, endpoint := range service.HTTPEndpoints {
		if !needInit(endpoint.MethodExpr.Payload.Type) {
			continue
		}
		for _, flag := range httpCLIRequestFlags(endpoint) {
			validation := codegen.NeedsValidation(flag.attribute, codegen.GoLayoutPolicy{})
			for _, preference := range cli.FlagFieldImportPreferences(flag.attribute, validation, flag.required, flag.hasDefault) {
				if _, ok := seen[preference.Path]; ok {
					continue
				}
				seen[preference.Path] = struct{}{}
				imports = append(imports, preference)
			}
		}
	}
	return imports
}

// httpCLIRequestFlags returns the request values that become command-line
// flags for one method payload constructor.
func httpCLIRequestFlags(endpoint *expr.HTTPEndpointExpr) []httpCLIFlag {
	var flags []httpCLIFlag
	if endpoint.Body != nil && endpoint.Body.Type != expr.Empty {
		flag := httpCLIFlag{attribute: endpoint.Body, required: true, hasDefault: endpoint.Body.DefaultValue != nil}
		if origin := endpoint.Body.Meta["origin:attribute"]; len(origin) > 0 {
			flag.required = endpoint.MethodExpr.Payload.IsRequired(origin[0])
			flag.hasDefault = endpoint.MethodExpr.Payload.GetDefault(origin[0]) != nil
		}
		flags = append(flags, flag)
	}
	for _, mapping := range []struct {
		mapped       *expr.MappedAttributeExpr
		pathRequired bool
	}{
		{mapped: endpoint.PathParams(), pathRequired: true},
		{mapped: endpoint.QueryParams()},
		{mapped: endpoint.Headers},
		{mapped: endpoint.Cookies},
	} {
		mapped := mapping.mapped
		if mapped == nil || mapped.IsEmpty() {
			continue
		}
		_ = codegen.WalkMappedAttr(mapped, func(name, _ string, required bool, attribute *expr.AttributeExpr) error {
			hasDefault := requestElementDefault(endpoint.MethodExpr.Payload, name, attribute) != nil
			if mapping.pathRequired {
				required = true
				hasDefault = false
			}
			flags = append(flags, httpCLIFlag{
				attribute:  attribute,
				required:   required,
				hasDefault: hasDefault,
			})
			return nil
		})
	}
	if policy := jsonRPCRequestIDPolicyFor(endpoint); policy != nil && policy.attribute != nil {
		flags = append(flags, httpCLIFlag{
			attribute:  policy.attribute.Attribute,
			required:   policy.required,
			hasDefault: policy.defaultValue != nil,
		})
	}
	flags = append(flags, httpBasicAuthFlags(endpoint)...)
	return flags
}

// httpBasicAuthFlags returns the username and password fields added to the
// generated client command for a method that uses Basic authentication.
func httpBasicAuthFlags(endpoint *expr.HTTPEndpointExpr) []httpCLIFlag {
	if !httpEndpointUsesBasicAuth(endpoint) {
		return nil
	}
	payload := endpoint.MethodExpr.Payload
	flags := make([]httpCLIFlag, 0, 2)
	for _, tag := range []string{"security:username", "security:password"} {
		name := expr.TaggedAttribute(payload, tag)
		flags = append(flags, httpCLIFlag{
			attribute:  payload.Find(name),
			required:   payload.IsRequired(name),
			hasDefault: payload.GetDefault(name) != nil,
		})
	}
	return flags
}

// httpServerCodecUsesContext reports whether a normal HTTP response or error
// encoder receives a request context.
func httpServerCodecUsesContext(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.IsJSONRPC() || endpoint.UsesWebSocket() {
			continue
		}
		if endpoint.Redirect == nil || len(endpoint.HTTPErrors) > 0 {
			return true
		}
	}
	return false
}

// httpClientCodecRestoresResponseBody reports whether a shared response
// decoder can read and then replace an ordinary HTTP response body.
func httpClientCodecRestoresResponseBody(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if !endpoint.IsJSONRPC() || !endpoint.UsesSSE() {
			return true
		}
	}
	return false
}

// httpClientCodecUsesIO reports whether a response decoder, streamed request,
// or multipart request names an io type or helper.
func httpClientCodecUsesIO(service *expr.HTTPServiceExpr) bool {
	return httpClientCodecRestoresResponseBody(service) ||
		httpServiceSkipsRequestBody(service) || serviceHasMultipartRequest(service)
}

// httpGeneratedImportPlan returns generated service packages referenced by
// one HTTP file. Path files contain only standard Go values and need none.
func httpGeneratedImportPlan(service *expr.HTTPServiceExpr, client bool, kind httpFileKind, servicePackage, viewsPackage *codegen.ImportSpec) []*codegen.ImportSpec {
	if kind == httpPathsFile {
		return nil
	}
	if kind == httpTransportFile && client && !httpServiceSkipsResponseBody(service) && !serviceHasMultipartRequest(service) {
		return nil
	}
	if kind == httpCodecFile && !httpCodecUsesService(service) {
		return nil
	}
	if kind == httpTypesFile && !httpTypesUseService(service) {
		return nil
	}
	imports := []*codegen.ImportSpec{servicePackage}
	if viewsPackage == nil {
		return imports
	}
	switch kind {
	case httpCodecFile, httpTypesFile:
		return append(imports, viewsPackage)
	case httpWebSocketFile, httpSSEFile:
		if client {
			return append(imports, viewsPackage)
		}
	}
	return imports
}

// httpTypesUseService reports whether copied transport types or their
// constructors name a generated service type.
func httpTypesUseService(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		method := endpoint.MethodExpr
		if len(method.Errors) > 0 {
			return true
		}
		for _, attribute := range []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result, method.StreamingResult} {
			if attributeUsesServiceType(attribute, make(map[expr.UserType]struct{})) {
				return true
			}
		}
	}
	return false
}

// httpCodecUsesService reports whether one codec names a generated payload,
// result, stream wrapper, or designed error type.
func httpCodecUsesService(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		method := endpoint.MethodExpr
		for _, attribute := range []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result, method.StreamingResult} {
			if attributeUsesServiceType(attribute, make(map[expr.UserType]struct{})) {
				return true
			}
		}
		_, viewed := method.Result.Type.(*expr.ResultTypeExpr)
		if len(method.Errors) > 0 || viewed ||
			endpoint.SkipRequestBodyEncodeDecode || endpoint.SkipResponseBodyEncodeDecode {
			return true
		}
	}
	return false
}

// attributeUsesServiceType reports whether a generated reference reaches a
// user type declared in the service package.
func attributeUsesServiceType(attribute *expr.AttributeExpr, seen map[expr.UserType]struct{}) bool {
	if attribute == nil || attribute.Type == expr.Empty {
		return false
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return true
	case *expr.Array:
		return attributeUsesServiceType(actual.ElemType, seen)
	case *expr.Map:
		return attributeUsesServiceType(actual.KeyType, seen) || attributeUsesServiceType(actual.ElemType, seen)
	case *expr.Object:
		for _, named := range *actual {
			if attributeUsesServiceType(named.Attribute, seen) {
				return true
			}
		}
	}
	return false
}

// httpCLIParserFixedImports returns packages named directly by the shared CLI
// parser and HTTP endpoint-selection templates.
func httpCLIParserFixedImports(services ...*expr.HTTPServiceExpr) []*codegen.ImportSpec {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("flag"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("os"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
	}
	seen := make(map[string]struct{}, len(imports))
	for _, spec := range imports {
		seen[spec.Path] = struct{}{}
	}
	for _, service := range services {
		for _, endpoint := range service.HTTPEndpoints {
			payload := endpoint.MethodExpr.Payload
			if payload.Type == expr.Empty || needInit(payload.Type) {
				continue
			}
			for _, preference := range cli.FlagImportPreferences(payload, false) {
				if _, ok := seen[preference.Path]; ok {
					continue
				}
				seen[preference.Path] = struct{}{}
				imports = append(imports, preference)
			}
		}
	}
	return imports
}

// httpServiceHasPayloadBuilder reports whether the client CLI writes a
// constructor that combines request values into a service payload.
func httpServiceHasPayloadBuilder(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if needInit(endpoint.MethodExpr.Payload.Type) {
			return true
		}
	}
	return false
}

// httpServiceHasUnion reports whether a generated transport type file writes
// JSON methods for a union used by the selected side.
func httpServiceHasUnion(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		attributes := []*expr.AttributeExpr{
			endpoint.MethodExpr.Payload,
			endpoint.MethodExpr.StreamingPayload,
			endpoint.MethodExpr.Result,
			endpoint.MethodExpr.StreamingResult,
		}
		for _, attribute := range attributes {
			if attributeContainsUnion(attribute, make(map[expr.UserType]struct{})) {
				return true
			}
		}
	}
	return false
}

// attributeContainsUnion reports whether attribute reaches a union definition.
func attributeContainsUnion(attribute *expr.AttributeExpr, seen map[expr.UserType]struct{}) bool {
	if attribute == nil {
		return false
	}
	switch actual := attribute.Type.(type) {
	case *expr.Union:
		return true
	case expr.UserType:
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return attributeContainsUnion(actual.Attribute(), seen)
	case *expr.Object:
		for _, named := range *actual {
			if attributeContainsUnion(named.Attribute, seen) {
				return true
			}
		}
	case *expr.Array:
		return attributeContainsUnion(actual.ElemType, seen)
	case *expr.Map:
		return attributeContainsUnion(actual.KeyType, seen) || attributeContainsUnion(actual.ElemType, seen)
	}
	return false
}

// httpServiceUsesSSE reports whether the client transport checks an SSE
// response content type.
func httpServiceUsesSSE(service *expr.HTTPServiceExpr) bool {
	return len(httpSSEEndpoints(service)) > 0
}

// httpServiceHasViewedSSE reports whether an SSE stream emits code that checks
// or reconstructs a selected result view.
func httpServiceHasViewedSSE(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range httpSSEEndpoints(service) {
		if _, ok := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr); ok {
			return true
		}
	}
	return false
}

// httpServiceHasMixedResults reports whether one handler selects between a
// normal response and an SSE response from the Accept header.
func httpServiceHasMixedResults(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MethodExpr.HasMixedResults() {
			return true
		}
	}
	return false
}

// httpServiceSkipsRequestBody reports whether one client or server copies a
// request stream instead of encoding it.
func httpServiceSkipsRequestBody(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.SkipRequestBodyEncodeDecode {
			return true
		}
	}
	return false
}

// httpServiceSkipsResponseBody reports whether one client or server copies a
// response stream instead of decoding it.
func httpServiceSkipsResponseBody(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.SkipResponseBodyEncodeDecode {
			return true
		}
	}
	return false
}

// httpServiceHasRequestBody reports whether one server codec decodes a request
// body through an io.Reader.
func httpServiceHasRequestBody(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.Body != nil && endpoint.Body.Type != expr.Empty {
			return true
		}
	}
	return false
}

// httpServiceUsesJSONRPC reports whether shared codec templates write a
// JSON-RPC request or response envelope.
func httpServiceUsesJSONRPC(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.IsJSONRPC() {
			return true
		}
	}
	return false
}

// httpServiceDecodesJSONRPCParams reports whether a server gives JSON-RPC
// request parameters to a generated payload decoder.
func httpServiceDecodesJSONRPCParams(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.IsJSONRPC() && endpoint.MethodExpr.Payload.Type != expr.Empty {
			return true
		}
	}
	return false
}

// httpServiceGeneratesJSONRPCRequestID reports whether a client codec creates
// a UUID for a request whose design does not supply an ID.
func httpServiceGeneratesJSONRPCRequestID(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		policy := jsonRPCRequestIDPolicyFor(endpoint)
		if policy != nil && policy.generates() {
			return true
		}
	}
	return false
}

// httpCodecUsesGoa reports whether request, response, or validation code calls
// Goa's generated-error helpers.
func httpCodecUsesGoa(service *expr.HTTPServiceExpr, client bool) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if client {
			if endpoint.SSE != nil && endpoint.SSE.RequestIDField != "" {
				return true
			}
			if attributeHasNonRequiredValidation(endpoint.MethodExpr.Payload, make(map[expr.UserType]struct{})) {
				return true
			}
			if policy := jsonRPCRequestIDPolicyFor(endpoint); policy != nil && policy.attribute != nil &&
				attributeHasNonRequiredValidation(policy.attribute.Attribute, make(map[expr.UserType]struct{})) {
				return true
			}
			for _, response := range endpoint.Responses {
				if !response.Headers.IsEmpty() || !response.Cookies.IsEmpty() {
					return true
				}
			}
			continue
		}
		if endpoint.Body != nil && endpoint.Body.Type != expr.Empty ||
			!endpoint.Params.IsEmpty() || !endpoint.Headers.IsEmpty() || !endpoint.Cookies.IsEmpty() ||
			len(endpoint.HTTPErrors) > 0 {
			return true
		}
		for _, flag := range httpBasicAuthFlags(endpoint) {
			if flag.required {
				return true
			}
		}
	}
	return false
}

// attributeHasNonRequiredValidation reports whether request conversion writes
// an enum, format, pattern, range, or length check. Required-field checks are
// handled while the request values are assembled.
func attributeHasNonRequiredValidation(attribute *expr.AttributeExpr, seen map[expr.UserType]struct{}) bool {
	if attribute == nil {
		return false
	}
	if validation := expr.EffectiveValidation(attribute); validation != nil && !validation.HasRequiredOnly() {
		return true
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return attributeHasNonRequiredValidation(actual.Attribute(), seen)
	case *expr.Object:
		for _, named := range *actual {
			if attributeHasNonRequiredValidation(named.Attribute, seen) {
				return true
			}
		}
	case *expr.Array:
		return attributeHasNonRequiredValidation(actual.ElemType, seen)
	case *expr.Map:
		return attributeHasNonRequiredValidation(actual.KeyType, seen) || attributeHasNonRequiredValidation(actual.ElemType, seen)
	}
	return false
}

// httpCodecUsesFmt reports whether a generated conversion formats a value
// whose HTTP text form is not a direct primitive conversion.
func httpCodecUsesFmt(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MapQueryParams != nil {
			return true
		}
	}
	return false
}

// httpCodecUsesErrors reports whether a generated codec compares, joins, or
// unwraps errors.
func httpCodecUsesErrors(service *expr.HTTPServiceExpr, client bool) bool {
	if client {
		return httpClientCodecRestoresResponseBody(service)
	}
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.Body != nil && endpoint.Body.Type != expr.Empty ||
			len(endpoint.HTTPErrors) > 0 || !endpoint.Cookies.IsEmpty() {
			return true
		}
	}
	return false
}

// httpCodecUsesStrings reports whether a generated codec splits, joins, or
// searches request or response text.
func httpCodecUsesStrings(service *expr.HTTPServiceExpr, client bool) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MapQueryParams != nil || mappedAttributeHasArray(endpoint.PathParams()) || mappedAttributeHasArray(endpoint.QueryParams()) {
			return true
		}
		if client && httpEndpointUsesBearerAuthorizationHeader(endpoint) {
			return true
		}
		if !client && httpEndpointUsesHeaderSecurity(endpoint) {
			return true
		}
		if mappedAttributeHasArray(endpoint.Headers) || mappedAttributeHasArray(endpoint.Cookies) {
			return true
		}
		for _, response := range endpoint.Responses {
			if mappedAttributeHasArray(response.Headers) || mappedAttributeHasArray(response.Cookies) {
				return true
			}
		}
	}
	return false
}

// httpEndpointUsesBasicAuth reports whether the method reads a username and
// password from the HTTP Authorization header.
func httpEndpointUsesBasicAuth(endpoint *expr.HTTPEndpointExpr) bool {
	for _, requirement := range endpoint.Requirements {
		for _, scheme := range requirement.Schemes {
			if scheme.Kind == expr.BasicAuthKind {
				return true
			}
		}
	}
	return false
}

// httpEndpointUsesBearerAuthorizationHeader reports whether the client adds a
// Bearer prefix to the Authorization header when the supplied token has none.
func httpEndpointUsesBearerAuthorizationHeader(endpoint *expr.HTTPEndpointExpr) bool {
	for _, requirement := range endpoint.Requirements {
		for _, scheme := range requirement.Schemes {
			if scheme.In != "header" || scheme.Name != "Authorization" {
				continue
			}
			switch scheme.Kind {
			case expr.BearerKind, expr.JWTKind, expr.OAuth2Kind:
				return true
			}
		}
	}
	return false
}

// httpEndpointUsesHeaderSecurity reports whether the server removes an
// authentication prefix from a credential read from an HTTP header.
func httpEndpointUsesHeaderSecurity(endpoint *expr.HTTPEndpointExpr) bool {
	for _, requirement := range endpoint.Requirements {
		for _, scheme := range requirement.Schemes {
			if scheme.Kind != expr.BasicAuthKind && scheme.In == "header" {
				return true
			}
		}
	}
	return false
}

// httpCodecUsesStrconv reports whether a generated codec converts numeric or
// boolean values to or from HTTP text.
func httpCodecUsesStrconv(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MapQueryParams != nil {
			attribute := endpoint.MethodExpr.Payload
			if name := *endpoint.MapQueryParams; name != "" {
				attribute = attribute.Find(name)
			}
			if dataTypeHasTextConvertedValue(attribute.Type) {
				return true
			}
		}
		attributes := make([]*expr.MappedAttributeExpr, 0, 4+2*len(endpoint.Responses))
		attributes = append(attributes,
			endpoint.PathParams(), endpoint.QueryParams(), endpoint.Headers, endpoint.Cookies,
		)
		for _, response := range endpoint.Responses {
			attributes = append(attributes, response.Headers, response.Cookies)
		}
		for _, attribute := range attributes {
			if mappedAttributeHasTextConvertedValue(attribute) {
				return true
			}
		}
	}
	return false
}

// wireCatalogValidationImports returns the packages used by every validator
// emitted in one HTTP types file.
func wireCatalogValidationImports(catalog *wireTypeCatalog) []*codegen.ImportSpec {
	seen := make(map[string]struct{})
	var imports []*codegen.ImportSpec
	for _, record := range catalog.records {
		if !record.needsValidator {
			continue
		}
		imports = appendWireValidationImports(imports, seen, record.identity.attribute, record.identity.policy)
	}
	for _, root := range catalog.validationRoots {
		imports = appendWireValidationImports(imports, seen, root.attribute, root.policy)
	}
	return imports
}

// appendWireValidationImports adds each runtime package once for one validator
// emitted in the current HTTP types file.
func appendWireValidationImports(imports []*codegen.ImportSpec, seen map[string]struct{}, attribute *expr.AttributeExpr, policy wireTypePolicy) []*codegen.ImportSpec {
	for _, preference := range codegen.ValidationRuntimeImports(attribute, wireGoLayoutPolicy(policy)) {
		if _, ok := seen[preference.Path]; ok {
			continue
		}
		seen[preference.Path] = struct{}{}
		imports = append(imports, codegen.NewImport(preference.Name, preference.Path))
	}
	return imports
}

// httpPathImportPaths returns packages used by request path constructors.
func httpPathImportPaths(service *expr.HTTPServiceExpr) []string {
	seen := make(map[string]struct{})
	for _, endpoint := range service.HTTPEndpoints {
		pathParameters := endpoint.PathParams()
		if pathParameters.IsEmpty() {
			continue
		}
		seen["fmt"] = struct{}{}
		for _, named := range *expr.AsObject(pathParameters.Attribute().Type) {
			array, ok := underlyingDataType(named.Attribute.Type).(*expr.Array)
			if !ok {
				continue
			}
			seen["strings"] = struct{}{}
			switch underlyingDataType(array.ElemType.Type).Kind() {
			case expr.StringKind, expr.BytesKind, expr.AnyKind:
				seen["net/url"] = struct{}{}
			default:
				seen["strconv"] = struct{}{}
			}
		}
	}
	paths := make([]string, 0, len(seen))
	for importPath := range seen {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	return paths
}

// mappedAttributeHasArray reports whether one HTTP element stores an array.
func mappedAttributeHasArray(attribute *expr.MappedAttributeExpr) bool {
	if attribute == nil || attribute.IsEmpty() {
		return false
	}
	for _, named := range *expr.AsObject(attribute.Attribute().Type) {
		if _, ok := underlyingDataType(named.Attribute.Type).(*expr.Array); ok {
			return true
		}
	}
	return false
}

// mappedAttributeHasTextConvertedValue reports whether one HTTP element
// contains a numeric or boolean value converted through strconv.
func mappedAttributeHasTextConvertedValue(attribute *expr.MappedAttributeExpr) bool {
	if attribute == nil || attribute.IsEmpty() {
		return false
	}
	return dataTypeHasTextConvertedValue(attribute.Attribute().Type)
}

// dataTypeHasTextConvertedValue reports whether one mapped HTTP value contains
// a number or boolean that generated code converts with strconv.
func dataTypeHasTextConvertedValue(dataType expr.DataType) bool {
	dataType = underlyingDataType(dataType)
	switch actual := dataType.(type) {
	case *expr.Object:
		for _, field := range *actual {
			if dataTypeHasTextConvertedValue(field.Attribute.Type) {
				return true
			}
		}
	case *expr.Array:
		return dataTypeHasTextConvertedValue(actual.ElemType.Type)
	case *expr.Map:
		return dataTypeHasTextConvertedValue(actual.KeyType.Type) ||
			dataTypeHasTextConvertedValue(actual.ElemType.Type)
	default:
		switch dataType.Kind() {
		case expr.BooleanKind, expr.IntKind, expr.Int32Kind, expr.Int64Kind,
			expr.UIntKind, expr.UInt32Kind, expr.UInt64Kind, expr.Float32Kind, expr.Float64Kind:
			return true
		}
	}
	return false
}

// underlyingDataType returns the first non-user type in a named type chain.
func underlyingDataType(dataType expr.DataType) expr.DataType {
	for {
		userType, ok := dataType.(expr.UserType)
		if !ok {
			return dataType
		}
		dataType = userType.Attribute().Type
	}
}

// plannedFileHeader returns a source header with the imports recorded before
// generation names were frozen.
func plannedFileHeader(title, packageName, filePath string, services *ServicesData) *codegen.SectionTemplate {
	return codegen.Header(title, packageName, services.fileImports[filepathKey(filePath)])
}

// generatedFileOutputPackage returns the import path of the package that owns
// a generated file. Files below gen use the generated package as their base.
// Root files such as cmd/server/http.go use the module containing gen.
func generatedFileOutputPackage(services *ServicesData, filePath string) string {
	outputPath := strings.TrimPrefix(strings.ReplaceAll(filePath, "\\", "/"), "./")
	if strings.HasPrefix(outputPath, codegen.Gendir+"/") {
		outputPath = strings.TrimPrefix(outputPath, codegen.Gendir+"/")
		return path.Join(services.GenPkg(), path.Dir(outputPath))
	}
	return path.Join(path.Dir(services.GenPkg()), path.Dir(outputPath))
}

// serviceDataForOutput copies the package-name fields that a template writes
// so they match the imports selected by its actual output package.
func serviceDataForOutput(data *ServiceData, services *ServicesData, outputPackage string) *ServiceData {
	serviceCopy := *data.Service
	servicePath := path.Join(services.GenPkg(), serviceCopy.PathName)
	for filePath, imports := range services.fileImports {
		if generatedFileOutputPackage(services, filePath) != outputPackage {
			continue
		}
		for _, spec := range imports {
			if spec.Path == servicePath {
				serviceCopy.PkgName = spec.Name
				break
			}
		}
	}
	copy := *data
	copy.Service = &serviceCopy
	copy.Endpoints = make([]*EndpointData, len(data.Endpoints))
	for index, endpoint := range data.Endpoints {
		endpointCopy := *endpoint
		endpointCopy.ServicePkgName = serviceCopy.PkgName
		copy.Endpoints[index] = &endpointCopy
	}
	return &copy
}

// exampleServiceDataForOutput gives local variables the unique service package
// path chosen for this generation. Example files may use several services at
// once, and two service names can produce the same Go name.
func exampleServiceDataForOutput(data *ServiceData, services *ServicesData, outputPackage string) *ServiceData {
	copy := serviceDataForOutput(data, services, outputPackage)
	copy.Service.VarName = codegen.Goify(copy.Service.PathName, false)
	return copy
}

// serviceReferenceAttributes returns the named service attributes referenced
// by generated HTTP or JSON-RPC endpoint sections, including the nested result
// field selected as SSE event data.
func serviceReferenceAttributes(endpoints ...*expr.HTTPEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result, method.StreamingResult)
		if endpoint.SSE != nil && endpoint.SSE.DataField != "" {
			event := method.Result
			if method.HasMixedResults() {
				event = method.StreamingResult
			}
			if object := expr.AsObject(event.Type); object != nil {
				attributes = append(attributes, object.Attribute(endpoint.SSE.DataField))
			}
		}
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
	}
	return attributes
}

// wireCatalogImportAttributes returns the type shapes and generated function
// inputs written by one transport types file.
func wireCatalogImportAttributes(catalog *wireTypeCatalog) (definitions, references []*expr.AttributeExpr) {
	for _, record := range catalog.records {
		definitions = append(definitions, record.identity.attribute)
		if record.needsValidator || record.needsNestedCall || record.needsConstructor {
			references = append(references, record.identity.attribute)
		}
	}
	for _, root := range catalog.validationRoots {
		references = append(references, root.attribute)
	}
	return definitions, references
}

// httpCLIPayloadBuilderReferenceAttributes returns the values converted or
// validated by one generated HTTP payload-builder file.
func httpCLIPayloadBuilderReferenceAttributes(service *expr.HTTPServiceExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range service.HTTPEndpoints {
		if !needInit(endpoint.MethodExpr.Payload.Type) {
			continue
		}
		attributes = append(attributes,
			endpoint.MethodExpr.Payload,
			endpoint.Body,
			endpoint.PathParams().AttributeExpr,
			endpoint.QueryParams().AttributeExpr,
			endpoint.Headers.AttributeExpr,
			endpoint.Cookies.AttributeExpr,
		)
		for _, flag := range httpCLIRequestFlags(endpoint) {
			attributes = append(attributes, flag.attribute)
		}
	}
	return attributes
}

// httpTransportReferenceAttributes returns values named directly by raw-body
// client and server sections instead of by generated codecs.
func httpTransportReferenceAttributes(service *expr.HTTPServiceExpr) []*expr.AttributeExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.SkipRequestBodyEncodeDecode || endpoint.SkipResponseBodyEncodeDecode {
			endpoints = append(endpoints, endpoint)
		}
	}
	return serviceReferenceAttributes(endpoints...)
}

// filepathKey normalizes generated paths so file writers on every platform
// read the same planned import record.
func filepathKey(filePath string) string {
	return strings.ReplaceAll(filePath, "\\", "/")
}

// httpWebSocketEndpoints returns only the endpoints whose stream sections are
// rendered into WebSocket files.
func httpWebSocketEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesWebSocket() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// httpSSEEndpoints returns only the endpoints whose stream sections are
// rendered into Server-Sent Events files.
func httpSSEEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesSSE() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}
