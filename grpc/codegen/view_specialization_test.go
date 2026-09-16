// This file checks that generated gRPC code keeps design-selected views in
// source and reads transport metadata only when the caller selects the view.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestUnaryViewedResultSpecialization(t *testing.T) {
	t.Run("missing or conflicting headers keep design selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.MessageResultTypeWithExplicitViewDSL)
		services := CreateGRPCServices(root)
		clientFiles := clientFiles(services)
		serverFiles := serverFiles(services)
		require.Len(t, clientFiles, 2)
		require.Len(t, serverFiles, 2)

		decoder := codegen.SectionsCode(t, clientFiles[1].Section("response-decoder"))
		assert.NotContains(t, decoder, `hdr.Get("goa-view")`)
		assert.Contains(t, decoder, `View: "tiny"`)
		assert.NotContains(t, decoder, "switch")
		encoder := codegen.SectionsCode(t, serverFiles[1].Section("response-encoder"))
		assert.Contains(t, encoder, `Append("goa-view", "tiny")`)
		assert.NotContains(t, encoder, `Append("goa-view", vres.View)`)
		testutil.AssertGo(t, "testdata/golden/viewed_result_fixed_response_encoder.go.golden", encoder)
		response := services.Get("ServiceMessageResultTypeWithExplicitView").Endpoints[0].Response
		require.Len(t, response.ClientConverts, 1)
		require.Equal(t, "tiny", response.ClientConverts[0].View)
		require.Same(t, response.ClientConverts[0].Convert, response.ClientConvert)
	})

	t.Run("caller selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.MessageResultTypeWithViewsDSL)
		services := CreateGRPCServices(root)
		clientFiles := clientFiles(services)
		serverFiles := serverFiles(services)
		require.Len(t, clientFiles, 2)
		require.Len(t, serverFiles, 2)

		decoder := codegen.SectionsCode(t, clientFiles[1].Section("response-decoder"))
		assert.Contains(t, decoder, `hdr.Get("goa-view")`)
		assert.Contains(t, decoder, "View: view")
		assert.Contains(t, decoder, "switch view")
		assert.Contains(t, decoder, "NewMethodMessageResultTypeWithViewsResultTiny(message)")
		assert.Contains(t, decoder, "NewMethodMessageResultTypeWithViewsResult(message)")
		encoder := codegen.SectionsCode(t, serverFiles[1].Section("response-encoder"))
		assert.Contains(t, encoder, `Append("goa-view", vres.View)`)
		assert.Contains(t, encoder, `return nil, goa.InvalidEnumValueError("view", vres.View`)
		testutil.AssertGo(t, "testdata/golden/viewed_result_dynamic_response_encoder.go.golden", encoder)
		response := services.Get("ServiceMessageResultTypeWithViews").Endpoints[0].Response
		require.Len(t, response.ServerConverts, 2)
		require.Equal(t, "default", response.ServerConverts[1].View)
		require.Same(t, response.ServerConverts[1].Convert, response.ServerConvert)
		require.Len(t, response.ClientConverts, 2)
		require.Equal(t, "default", response.ClientConverts[1].View)
		require.Same(t, response.ClientConverts[1].Convert, response.ClientConvert)
		tinyValidation := response.ClientConverts[0].Convert.Validation
		fullValidation := response.ClientConverts[1].Convert.Validation
		require.NotNil(t, tinyValidation)
		require.NotNil(t, fullValidation)
		require.NotSame(t, tinyValidation, fullValidation)
		require.Equal(t, "ValidateMethodMessageResultTypeWithViewsResponseTiny", tinyValidation.Declaration.Name())
		require.Equal(t, "ValidateMethodMessageResultTypeWithViewsResponse", fullValidation.Declaration.Name())
		assert.Contains(t, tinyValidation.Def, `MissingFieldError("IntField", "message")`)
		assert.NotContains(t, tinyValidation.Def, `MissingFieldError("StringField", "message")`)
		assert.Contains(t, fullValidation.Def, `MissingFieldError("StringField", "message")`)
		require.Less(t,
			strings.Index(decoder, tinyValidation.Declaration.Name()+"(message)"),
			strings.Index(decoder, "NewMethodMessageResultTypeWithViewsResultTiny(message)"),
		)
		require.Less(t,
			strings.Index(decoder, fullValidation.Declaration.Name()+"(message)"),
			strings.Index(decoder, "NewMethodMessageResultTypeWithViewsResult(message)"),
		)
	})
}

func TestStreamingViewedResultSpecialization(t *testing.T) {
	t.Run("design selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ServerStreamingResultCollectionWithExplicitViewDSL)
		services := CreateGRPCServices(root)
		serverFiles := serverFiles(services)
		clientFiles := clientFiles(services)
		require.Len(t, serverFiles, 2)
		require.Len(t, clientFiles, 2)

		serverStruct := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-struct-type"))
		clientStruct := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-struct-type"))
		assert.NotContains(t, serverStruct, "\n\tview")
		assert.NotContains(t, clientStruct, "\n\tview")
		stream := services.Get("ServiceServerStreamingResultTypeCollectionWithExplicitView").Endpoints[0].ClientStream
		require.Len(t, stream.RecvConverts, 1)
		require.Equal(t, "tiny", stream.RecvConverts[0].View)
		require.Same(t, stream.RecvConverts[0].Convert, stream.RecvConvert)
	})

	t.Run("caller selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ServerStreamingResultWithViewsDSL)
		services := CreateGRPCServices(root)
		serverFiles := serverFiles(services)
		clientFiles := clientFiles(services)
		require.Len(t, serverFiles, 2)
		require.Len(t, clientFiles, 2)

		serverStruct := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-struct-type"))
		clientStruct := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-struct-type"))
		assert.Contains(t, serverStruct, "\n\tview")
		assert.Contains(t, clientStruct, "\n\tview")
		assert.Contains(t, serverStruct, "sentView string")
		assert.Contains(t, clientStruct, "viewSet bool")
		send := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-send"))
		assert.Contains(t, send, `if view == "" {`)
		assert.Contains(t, send, `view = "default"`)
		assert.Contains(t, send, `if s.sentView != "" && view != s.sentView`)
		assert.Contains(t, send, `SetHeader(metadata.Pairs("goa-view", view))`)
		assert.Contains(t, send, `return goa.InvalidEnumValueError("view", view`)
		require.Less(t,
			strings.Index(send, `InvalidEnumValueError("view", view`),
			strings.Index(send, `SetHeader(metadata.Pairs("goa-view", view))`),
		)
		testutil.AssertGo(t, "testdata/golden/viewed_result_dynamic_stream_send.go.golden", send)
		recv := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-recv"))
		assert.Contains(t, recv, `s.stream.Header()`)
		assert.Contains(t, recv, `goa.MissingFieldError("goa-view", "metadata")`)
		assert.Contains(t, recv, "switch s.view")
		assert.Contains(t, recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeViewTiny(v)")
		assert.Contains(t, recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeView(v)")
		stream := services.Get("ServiceServerStreamingUserTypeRPC").Endpoints[0].ServerStream
		require.Len(t, stream.SendConverts, 2)
		require.Equal(t, "default", stream.SendConverts[1].View)
		require.Same(t, stream.SendConverts[1].Convert, stream.SendConvert)
		clientStream := services.Get("ServiceServerStreamingUserTypeRPC").Endpoints[0].ClientStream
		require.Len(t, clientStream.RecvConverts, 2)
		require.Equal(t, "default", clientStream.RecvConverts[1].View)
		require.Same(t, clientStream.RecvConverts[1].Convert, clientStream.RecvConvert)
		tinyValidation := clientStream.RecvConverts[0].Convert.Validation
		fullValidation := clientStream.RecvConverts[1].Convert.Validation
		require.NotNil(t, tinyValidation)
		require.NotNil(t, fullValidation)
		require.NotSame(t, tinyValidation, fullValidation)
		require.Equal(t, "ValidateMethodServerStreamingUserTypeRPCResponseTiny", tinyValidation.Declaration.Name())
		require.Equal(t, "ValidateMethodServerStreamingUserTypeRPCResponse", fullValidation.Declaration.Name())
		assert.Contains(t, tinyValidation.Def, `MissingFieldError("IntField", "message")`)
		assert.NotContains(t, tinyValidation.Def, `MissingFieldError("DoubleField", "message")`)
		assert.Contains(t, fullValidation.Def, `MissingFieldError("DoubleField", "message")`)
		require.Less(t,
			strings.Index(recv, tinyValidation.Declaration.Name()+"(v)"),
			strings.Index(recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeViewTiny(v)"),
		)
		require.Less(t,
			strings.Index(recv, fullValidation.Declaration.Name()+"(v)"),
			strings.Index(recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeView(v)"),
		)
	})
}

// TestViewedResponseValidationNamesIgnoreViewOrder checks that each generated
// name describes its selected view regardless of DSL declaration order.
func TestViewedResponseValidationNamesIgnoreViewOrder(t *testing.T) {
	namesFor := func(views []string) map[string]string {
		root := RunGRPCDSL(t, func() {
			result := d.ResultType("application/vnd.validation-order", func() {
				d.TypeName("ValidationOrder")
				d.Attributes(func() {
					d.Field(1, "first", d.String)
					d.Field(2, "second", d.String)
					d.Required("first", "second")
				})
				for _, view := range views {
					d.View(view, func() {
						if view == "tiny" {
							d.Attribute("first")
						} else {
							d.Attribute("second")
						}
					})
				}
			})
			d.Service("ValidationOrderService", func() {
				d.Method("Show", func() {
					d.Result(result)
					d.GRPC(func() {})
				})
			})
		})
		response := CreateGRPCServices(root).Get("ValidationOrderService").Endpoints[0].Response
		names := make(map[string]string, len(response.ClientConverts))
		for _, conversion := range response.ClientConverts {
			names[conversion.View] = conversion.Convert.Validation.Declaration.Name()
		}
		return names
	}

	forward := namesFor([]string{"tiny", "compact"})
	reverse := namesFor([]string{"compact", "tiny"})
	require.Equal(t, forward, reverse)
	require.Equal(t, "ValidateShowResponseTiny", forward["tiny"])
	require.Equal(t, "ValidateShowResponseCompact", forward["compact"])
	require.Equal(t, "ValidateShowResponse", forward[expr.DefaultView])
}

// TestGRPCProtobufValidationForRecursiveView checks that selected and omitted
// recursive fields are each visited once while their validation rules are kept
// or removed.
func TestGRPCProtobufValidationForRecursiveView(t *testing.T) {
	responseType := &expr.UserTypeExpr{TypeName: "Response", UID: "response"}
	responseValue := &expr.AttributeExpr{Type: expr.String}
	responseChild := &expr.AttributeExpr{Type: responseType}
	responseType.AttributeExpr = &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "value", Attribute: responseValue},
			{Name: "child", Attribute: responseChild},
		},
		Validation: &expr.ValidationExpr{Required: []string{"value", "child"}},
	}
	selectedType := &expr.UserTypeExpr{TypeName: "Selected", UID: "selected"}
	selectedChild := &expr.AttributeExpr{Type: selectedType}
	selectedType.AttributeExpr = &expr.AttributeExpr{
		Type:       &expr.Object{{Name: "child", Attribute: selectedChild}},
		Validation: &expr.ValidationExpr{Required: []string{"child"}},
	}

	validation := grpcProtobufValidationForView(
		&expr.AttributeExpr{Type: responseType},
		&expr.AttributeExpr{Type: selectedType},
	)
	validationObject := validation.Type.(expr.UserType).Attribute()
	require.Equal(t, []string{"child"}, validationObject.Validation.Required)
	require.Nil(t, expr.AsObject(validationObject.Type).Attribute("value").Validation)
}

// TestGRPCProtobufValidationForSharedViewField checks that omitting one field
// does not remove checks from a selected sibling with the same nested type.
func TestGRPCProtobufValidationForSharedViewField(t *testing.T) {
	nested := &expr.UserTypeExpr{
		TypeName: "Nested",
		UID:      "nested",
		AttributeExpr: &expr.AttributeExpr{
			Type:       &expr.Object{{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}}},
			Validation: &expr.ValidationExpr{Required: []string{"value"}},
		},
	}
	response := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "selected", Attribute: &expr.AttributeExpr{Type: nested}},
			{Name: "omitted", Attribute: &expr.AttributeExpr{Type: nested}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"selected", "omitted"}},
	}
	selected := &expr.AttributeExpr{
		Type:       &expr.Object{{Name: "selected", Attribute: &expr.AttributeExpr{Type: nested}}},
		Validation: &expr.ValidationExpr{Required: []string{"selected"}},
	}

	validation := grpcProtobufValidationForView(response, selected)
	fields := expr.AsObject(validation.Type)
	selectedType := fields.Attribute("selected").Type.(expr.UserType)
	omittedType := fields.Attribute("omitted").Type.(expr.UserType)
	require.NotNil(t, selectedType.Attribute().Validation)
	require.Equal(t, []string{"value"}, selectedType.Attribute().Validation.Required)
	require.Nil(t, omittedType.Attribute().Validation)
}
