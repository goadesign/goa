// These tests check that form designs fail before generation when a request
// body cannot be represented by flat strings and repeated values.
package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
)

func TestFormRequestValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		payload func()
		http    func()
		error   string
	}{
		{
			"primitive", func() {
				Payload(String)
			}, func() {}, "nonempty object",
		},
		{"empty", func() {}, func() {}, "nonempty object"},
		{
			"nested object", func() {
				Payload(func() {
					Attribute("nested", func() {
						Attribute("name", String)
					})
				})
			}, func() {}, "primitive or an array",
		},
		{
			"nested array", func() {
				Payload(func() {
					Attribute("nested", ArrayOf(ArrayOf(String)))
				})
			}, func() {}, "primitive or an array",
		},
		{
			"map", func() {
				Payload(func() {
					Attribute("map", MapOf(String, String))
				})
			}, func() {}, "primitive or an array",
		},
		{
			"any", func() {
				Payload(func() {
					Attribute("value", Any)
				})
			}, func() {}, "primitive or an array",
		},
		{
			"custom type", func() {
				Payload(func() {
					Attribute("value", String, func() {
						Meta("struct:field:type", "Custom")
					})
				})
			}, func() {}, "custom Go field type",
		},
		{
			"multipart", func() {
				Payload(func() {
					Attribute("value", String)
				})
			}, MultipartRequest, "cannot be combined",
		},
		{
			"skip body", func() {
				Payload(func() {
					Attribute("value", String)
				})
			}, SkipRequestBodyEncodeDecode, "cannot be combined",
		},
		{
			"mapped only", func() {
				Payload(func() {
					Attribute("value", String)
				})
			}, func() {
				Header("value")
			}, "nonempty object",
		},
		{
			"valid", func() {
				Payload(func() {
					Attribute("value", String)
					Attribute("items", ArrayOf(Int))
				})
			}, func() {}, "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := runDSL(t, func() {
				Service("forms", func() {
					Method("send", func() {
						test.payload()
						HTTP(func() {
							POST("/form")
							FormRequest()
							test.http()
						})
					})
				})
			})
			if test.error == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.error)
			}
		})
	}
}

func TestFormRequestNamedValues(t *testing.T) {
	for _, test := range []struct {
		name   string
		custom bool
		error  string
	}{
		{"any alias", false, "primitive or an array"},
		{"custom alias", true, "custom Go field type"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := runDSL(t, func() {
				value := Type("Value", Any)
				if test.custom {
					value = Type("CustomValue", String, func() {
						Meta("struct:field:type", "Custom")
					})
				}
				Service("forms", func() {
					Method("send", func() {
						Payload(func() {
							Attribute("value", value)
						})
						HTTP(func() {
							POST("/form")
							FormRequest()
						})
					})
				})
			})
			require.ErrorContains(t, err, test.error)
		})
	}
}

func TestFormRequestJSONRPCRejected(t *testing.T) {
	_, err := runDSL(t, func() {
		Service("rpc", func() {
			JSONRPC(func() {
				POST("/rpc")
			})
			Method("send", func() {
				Payload(func() {
					Attribute("value", String)
				})
				JSONRPC(func() {
					FormRequest()
				})
			})
		})
	})
	require.ErrorContains(t, err, "FormRequest")
}
