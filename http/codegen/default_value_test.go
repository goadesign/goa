// This file checks that authored defaults become the exact flag text and JSON
// shape accepted by generated HTTP clients.
package codegen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestClientBodyDefaultUsesHTTPFieldNamesAndJSONBytes(t *testing.T) {
	object := expr.Object{
		&expr.NamedAttributeExpr{
			Name: "full:mapped_full",
			Attribute: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{
					"struct:tag:json":      {"full_name,omitempty"},
					"struct:tag:json:name": {"ignored_name"},
				},
			},
		},
		&expr.NamedAttributeExpr{
			Name: "bytes:mapped_bytes",
			Attribute: &expr.AttributeExpr{
				Type: expr.Bytes,
				Meta: expr.MetaExpr{"struct:tag:json:name": {"encoded_bytes"}},
			},
		},
		&expr.NamedAttributeExpr{
			Name:      "plain:mapped_plain",
			Attribute: &expr.AttributeExpr{Type: expr.String},
		},
		&expr.NamedAttributeExpr{
			Name: "skip:mapped_skip",
			Attribute: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{"struct:tag:json": {"-"}},
			},
		},
	}
	attribute := &expr.AttributeExpr{Type: &object}
	got := clientBodyDefault(attribute, map[string]any{
		"full":  "first",
		"bytes": "plain bytes",
		"plain": "third",
		"skip":  "hidden",
	})

	encoded, err := json.Marshal(got)
	require.NoError(t, err)
	require.JSONEq(t, `{"full_name":"first","encoded_bytes":"cGxhaW4gYnl0ZXM=","mapped_plain":"third"}`, string(encoded))
}

func TestClientBodyDefaultUsesRawTextForTopLevelBytes(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.Bytes}
	for _, test := range []struct {
		name  string
		value any
	}{
		{name: "authored string", value: "plain bytes"},
		{name: "authored byte slice", value: []byte("plain bytes")},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, "plain bytes", clientBodyDefault(attribute, test.value))
		})
	}
}

func TestClientBodyDefaultIsUsedOnlyByClientCLIPlanning(t *testing.T) {
	root := expr.RunDSL(t, func() {
		details := dsl.Type("DefaultDetails", func() {
			dsl.Attribute("content", dsl.Bytes, func() {
				dsl.Meta("struct:tag:json:name", "encoded_content")
			})
		})
		dsl.Service("defaults", func() {
			dsl.Method("object", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", details, func() {
						dsl.Default(map[string]any{"content": "plain bytes"})
					})
				})
				dsl.HTTP(func() {
					dsl.POST("/object")
					dsl.Body("value")
				})
			})
			dsl.Method("bytes", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.Bytes, func() {
						dsl.Default("plain bytes")
					})
				})
				dsl.HTTP(func() {
					dsl.POST("/bytes")
					dsl.Body("value")
				})
			})
		})
	})

	service := linkedHTTPPlanForRoot(t, root).services.Get("defaults")
	object := service.Endpoint("object").Payload.Request.PayloadInit
	require.Equal(t, map[string]any{"encoded_content": []byte("plain bytes")}, object.ClientArgs[0].DefaultValue)
	require.Equal(t, map[string]any{"content": "plain bytes"}, object.ServerArgs[0].DefaultValue)
	objectFlags, _ := buildFlags(service, service.Endpoint("object"))
	require.Equal(t, `{"encoded_content":"cGxhaW4gYnl0ZXM="}`, objectFlags[0].DefaultValue)
	bytes := service.Endpoint("bytes").Payload.Request.PayloadInit
	require.Equal(t, "plain bytes", bytes.ClientArgs[0].DefaultValue)
	require.Equal(t, "plain bytes", bytes.ServerArgs[0].DefaultValue)
	byteFlags, _ := buildFlags(service, service.Endpoint("bytes"))
	require.Equal(t, "plain bytes", byteFlags[0].DefaultValue)
}
