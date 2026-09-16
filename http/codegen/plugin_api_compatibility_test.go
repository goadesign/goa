// This file checks the public HTTP plugin fields kept for existing plugins.
package codegen

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

var (
	_ func(string, *ServicesData) []*codegen.File                      = ClientFiles
	_ func(string, *ServicesData) []*codegen.File                      = ClientCLIFiles
	_ func(string, *ServicesData) []*codegen.File                      = ServerFiles
	_ func(string, *ServicesData) []*codegen.File                      = ServerTypeFiles
	_ func(string, *ServicesData) []*codegen.File                      = ClientTypeFiles
	_ func(*ServicesData) []*codegen.File                              = PathFiles
	_ func(string, *expr.HTTPServiceExpr, *ServicesData) *codegen.File = ClientEncodeDecodeFile
	_ func(string, *expr.HTTPServiceExpr, *ServicesData) *codegen.File = ServerEncodeDecodeFile
	_ func(string, *expr.HTTPServiceExpr, *ServicesData) *codegen.File = WebsocketClientFile
)

// TestReleasedHTTPNamesMatchDeclarations checks all 23 released string fields,
// including optional fields that are empty when no declaration exists.
func TestReleasedHTTPNamesMatchDeclarations(t *testing.T) {
	plan := linkedHTTPPlanForRoot(t, releasedHTTPNamesRoot(t))
	service := plan.services.Get("Names")
	assertReleasedName(t, service.ServerStruct, service.ServerStructDeclaration)
	assertReleasedName(t, service.MountPointStruct, service.MountPointStructDeclaration)
	assertReleasedName(t, service.ServerInit, service.ServerInitDeclaration)
	assertReleasedName(t, service.MountServer, service.MountServerDeclaration)
	assertReleasedName(t, service.ClientStruct, service.ClientStructDeclaration)

	for _, endpoint := range service.Endpoints {
		assertReleasedName(t, endpoint.MountHandler, endpoint.MountHandlerDeclaration)
		assertReleasedName(t, endpoint.HandlerInit, endpoint.HandlerInitDeclaration)
		assertReleasedName(t, endpoint.RequestDecoder, endpoint.RequestDecoderDeclaration)
		assertReleasedName(t, endpoint.ResponseEncoder, endpoint.ResponseEncoderDeclaration)
		assertReleasedName(t, endpoint.ErrorEncoder, endpoint.ErrorEncoderDeclaration)
		assertReleasedName(t, endpoint.ClientStruct, endpoint.ClientStructDeclaration)
		assertReleasedName(t, endpoint.RequestEncoder, endpoint.RequestEncoderDeclaration)
		assertReleasedName(t, endpoint.ResponseDecoder, endpoint.ResponseDecoderDeclaration)
		assertReleasedName(t, endpoint.BuildStreamPayload, endpoint.BuildStreamPayloadDeclaration)
	}

	multipart := service.Endpoint("Multipart")
	for _, data := range []*MultipartData{multipart.MultipartRequestDecoder, multipart.MultipartRequestEncoder} {
		require.NotNil(t, data)
		assertReleasedName(t, data.FuncName, data.FuncDeclaration)
		assertReleasedName(t, data.InitName, data.InitDeclaration)
	}
	stream := service.Endpoint("Watch").SSE
	require.NotNil(t, stream)
	assertReleasedName(t, stream.StructName, stream.StructDeclaration)
	require.NotEmpty(t, service.FileServers)
	assertReleasedName(t, service.FileServers[0].MountHandler, service.FileServers[0].MountHandlerDeclaration)
	empty := service.Endpoint("Empty")
	require.Nil(t, empty.RequestDecoderDeclaration)
	require.Empty(t, empty.RequestDecoder)
	socket := service.Endpoint("Socket").ServerWebSocket
	require.NotNil(t, socket)
	assertReleasedName(t, socket.VarName, socket.VarDeclaration)
	assertReleasedName(t, service.Endpoint("Complete").RequestInit.Name, service.Endpoint("Complete").RequestInit.Declaration)
	typeData := releasedTypeData(t, service, func(data *TypeData) bool { return data.NestedValidatorDeclaration != nil })
	assertReleasedName(t, typeData.VarName, typeData.Declaration)
	assertReleasedName(t, typeData.ValidatorName, typeData.ValidatorDeclaration)
	assertReleasedName(t, typeData.NestedValidatorName, typeData.NestedValidatorDeclaration)
}

// TestReleasedHTTPFileFunctionsUsePlannedPackage checks that public helpers
// render the retained HTTP plan and reject a different generated package.
func TestReleasedHTTPFileFunctionsUsePlannedPackage(t *testing.T) {
	plan := linkedHTTPPlanForRoot(t, expr.RunDSL(t, testdata.MultiSimpleDSL))
	genpkg := plan.services.GenPkg()
	for _, files := range []struct {
		name     string
		released func(string, *ServicesData) []*codegen.File
		planned  func(*ServicesData) []*codegen.File
	}{
		{name: "client", released: ClientFiles, planned: clientFiles},
		{name: "client CLI", released: ClientCLIFiles, planned: clientCLIFiles},
		{name: "server", released: ServerFiles, planned: serverFiles},
		{name: "server types", released: ServerTypeFiles, planned: serverTypeFiles},
		{name: "client types", released: ClientTypeFiles, planned: clientTypeFiles},
	} {
		t.Run(files.name, func(t *testing.T) {
			require.Len(t, files.released(genpkg, plan.services), len(files.planned(plan.services)))
			require.PanicsWithValue(
				t,
				`HTTP generation package "other.local/gen" does not match planned package "generated.local/gen"`,
				func() {
					files.released("other.local/gen", plan.services)
				},
			)
		})
	}
	require.Len(t, PathFiles(plan.services), len(pathFiles(plan.services)))
	service := plan.root.API.HTTP.Services[0]
	require.Equal(t, clientEncodeDecodeFile(service, plan.services).Path, ClientEncodeDecodeFile(genpkg, service, plan.services).Path)
	require.Equal(t, serverEncodeDecodeFile(service, plan.services).Path, ServerEncodeDecodeFile(genpkg, service, plan.services).Path)

	websocket := linkedHTTPPlanForRoot(t, releasedHTTPNamesRoot(t))
	websocketService := websocket.root.API.HTTP.Service("Names")
	require.Equal(
		t,
		websocketClientFile(websocketService, websocket.services).Path,
		WebsocketClientFile(websocket.services.GenPkg(), websocketService, websocket.services).Path,
	)
}

// TestReleasedHTTPNameUsesCollisionSuffix checks that a compatibility field
// copies the final package name instead of rebuilding an unsuffixed name.
func TestReleasedHTTPNameUsesCollisionSuffix(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("read-data", func() { dsl.HTTP(func() { dsl.GET("/first") }) })
			dsl.Method("read_data", func() { dsl.HTTP(func() { dsl.GET("/second") }) })
		})
	})
	plan := linkedHTTPPlanForRoot(t, root)
	endpoint := plan.services.Get("Calc").Endpoint("read_data")
	require.NotEqual(t, "MountReadDataHandler", endpoint.MountHandlerDeclaration.Name())
	require.Contains(t, endpoint.MountHandlerDeclaration.Name(), "MountReadData")
	require.Equal(t, endpoint.MountHandlerDeclaration.Name(), endpoint.MountHandler)
}

// TestReleasedSSEDataFieldTypeMatchesPlannedValue verifies existing plugins
// can still read the final type of an explicitly mapped SSE data field.
func TestReleasedSSEDataFieldTypeMatchesPlannedValue(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Events", func() {
			dsl.Method("Watch", func() {
				dsl.StreamingResult(func() {
					dsl.Attribute("value", dsl.String)
				})
				dsl.HTTP(func() {
					dsl.GET("/watch")
					dsl.ServerSentEvents("value")
				})
			})
		})
	})
	stream := linkedHTTPPlanForRoot(t, root).services.Get("Events").Endpoint("Watch").SSE
	require.NotNil(t, stream)
	require.NotEmpty(t, stream.DataField)
	require.Equal(t, stream.Data.TypeRef, stream.DataFieldTypeRef)
}

// TestReleasedHTTPNameFieldsRemainSourceCompatible checks old keyed literals
// and template reads still compile.
func TestReleasedHTTPNameFieldsRemainSourceCompatible(t *testing.T) {
	data := struct {
		Endpoint   *EndpointData
		Service    *ServiceData
		FileServer *FileServerData
		Multipart  *MultipartData
		Stream     *SSEData
		WebSocket  *WebSocketData
		Init       *InitData
		Type       *TypeData
	}{
		&EndpointData{MountHandler: "MountReadHandler"},
		&ServiceData{ServerStruct: "Server"},
		&FileServerData{MountHandler: "MountAssetJSON"},
		&MultipartData{FuncName: "DecoderFunc", InitName: "NewDecoder"},
		&SSEData{StructName: "ReadServerStream"},
		&WebSocketData{VarName: "SocketServerStream"},
		&InitData{Name: "NewBody"},
		&TypeData{VarName: "Body", ValidatorName: "ValidateBody", NestedValidatorName: "validateBodyAt"},
	}
	tmpl := template.Must(template.New("released-fields").Parse(
		`{{.Endpoint.MountHandler}} {{.Service.ServerStruct}} {{.FileServer.MountHandler}} {{.Multipart.FuncName}} {{.Stream.StructName}} {{.WebSocket.VarName}} {{.Init.Name}} {{.Type.VarName}} {{.Type.ValidatorName}} {{.Type.NestedValidatorName}}`,
	))
	var rendered bytes.Buffer
	require.NoError(t, tmpl.Execute(&rendered, data))
	require.Equal(t, "MountReadHandler Server MountAssetJSON DecoderFunc ReadServerStream SocketServerStream NewBody Body ValidateBody validateBodyAt", rendered.String())
}
