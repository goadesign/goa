// This file checks that authored defaults become the exact flag text and JSON
// shape accepted by generated HTTP clients.
package codegen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type customJSONDefault int64

// String differs from JSON so a switch to scalar CLI formatting is observable.
func (v customJSONDefault) String() string {
	return fmt.Sprintf("text:%d", int64(v))
}

// MarshalJSON makes preservation of the authored value's methods observable.
func (v customJSONDefault) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("wire:%d", int64(v)))
}

func TestCustomDefaultPreservesConsumerSerialization(t *testing.T) {
	for _, test := range []struct {
		name      string
		primitive expr.Primitive
		value     any
		target    string
		flag      string
		nested    string
		schema    string
		yaml      string
	}{
		{"raw string", expr.String, json.RawMessage("true"), "json.RawMessage",
			"true", `{"value":true}`, `{"default":true}`, "default:\n    - 116\n    - 114\n    - 117\n    - 101\n"},
		{"raw bytes", expr.Bytes, json.RawMessage("true"), "json.RawMessage",
			`"true"`, `{"value":"dHJ1ZQ=="}`, `{"default":true}`, "default:\n    - 116\n    - 114\n    - 117\n    - 101\n"},
		{"named JSON and Stringer", expr.Int64, customJSONDefault(7), "OtherNumber",
			`"wire:7"`, `{"value":"wire:7"}`, `{"default":"wire:7"}`, "default: 7\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var declared expr.UserType
			expr.RunDSL(t, func() {
				declared = dsl.Type("CustomDefault", test.primitive, func() {
					if test.target == "json.RawMessage" {
						dsl.Meta("struct:field:type", test.target, "encoding/json")
					} else {
						dsl.Meta("struct:field:type", test.target)
					}
					dsl.Default(test.value)
				})
			})
			attribute := declared.Attribute()
			require.Equal(t, test.value, attribute.DefaultValue)
			require.Equal(t, reflect.TypeOf(test.value), reflect.TypeOf(attribute.DefaultValue))
			body := clientBodyDefault(attribute, attribute.DefaultValue)
			plan := cli.NewFlagPlan(attribute, test.target, test.target, nil)
			flag := cli.NewFlagDataForPlan("defaults", "show", "value", plan, "", false, nil, body)
			require.Equal(t, test.flag, flag.DefaultValue)
			object := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "value", Attribute: attribute},
			}}
			nested := clientBodyDefault(object, map[string]any{"value": attribute.DefaultValue})
			encoded, err := json.Marshal(nested)
			require.NoError(t, err)
			require.JSONEq(t, test.nested, string(encoded))
			schema := &openapi.Schema{DefaultValue: openapi.ToStringMap(attribute.DefaultValue)}
			encoded, err = json.Marshal(schema)
			require.NoError(t, err)
			require.JSONEq(t, test.schema, string(encoded))
			encoded, err = yaml.Marshal(schema)
			require.NoError(t, err)
			require.Equal(t, test.yaml, string(encoded))
		})
	}
}

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
