// These tests author service-specific HTTP names and verify that inherited
// fields and nested declarations retain their service definitions.
package dsl_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestBodyFieldName(t *testing.T) {
	for _, test := range []struct {
		name    string
		mapping string
		second  string
		failure string
	}{
		{name: "valid inherited field", mapping: "display_name:displayName"},
		{name: "supported punctuation", mapping: "display_name: display-name.1 "},
		{name: "supported unicode", mapping: "display_name:表示名"},
		{name: "hidden tag", mapping: "display_name:-", failure: "JSON struct tag"},
		{name: "comma tag", mapping: "display_name:name,omitempty", failure: "JSON struct tag"},
		{name: "quote tag", mapping: `display_name:name"`, failure: "JSON struct tag"},
		{name: "backslash tag", mapping: `display_name:name\`, failure: "JSON struct tag"},
		{name: "unknown", mapping: "missing:displayName", failure: "missing"},
		{name: "malformed", mapping: "display_name", failure: "field:wireName"},
		{name: "empty", mapping: "display_name:", failure: "field:wireName"},
		{name: "collision", mapping: "display_name:other", failure: "collision"},
		{name: "duplicate", mapping: "display_name:displayName", second: "display_name:displayName", failure: "already"},
		{name: "conflict", mapping: "display_name:displayName", second: "display_name:label", failure: "already"},
		{name: "mapped sibling collision", mapping: "display_name:displayName", second: "other:displayName", failure: "collision"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var original, derived expr.UserType
			root, err := runDSL(t, func() {
				original = Type("Original", func() {
					Field(1, "display_name", String, func() {
						Meta("struct:tag:json", "StoredName", "string", "omitempty")
					})
					Field(2, "other", String)
					Required("display_name", "other")
				})
				derived = Type("Derived", func() { Extend(original) })
				Service("mapped", func() {
					HTTP(func() {
						BodyFieldName(derived, test.mapping)
						if test.second != "" {
							BodyFieldName(derived, test.second)
						}
					})
					Method("exchange", func() {
						Payload(derived)
						Result(derived)
						HTTP(func() { POST("/mapped") })
					})
				})
				Service("plain", func() {
					Method("exchange", func() {
						Payload(derived)
						Result(original)
						HTTP(func() { POST("/plain") })
					})
				})
			})
			if test.failure != "" {
				require.ErrorContains(t, err, test.failure)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, derived.Attribute().Find("display_name"))
			require.Nil(t, derived.Attribute().Find("display_name:displayName"))
			mapped := root.API.HTTP.Service("mapped").Endpoint("exchange")
			_, wire, _ := strings.Cut(test.mapping, ":")
			require.NotNil(t, mapped.Body.Find(test.mapping))
			require.Equal(t, []string{wire + ",string,omitempty"}, mapped.Body.Find(test.mapping).Meta["struct:tag:json"])
			require.Equal(t, []string{"StoredName", "string", "omitempty"}, derived.Attribute().Find("display_name").Meta["struct:tag:json"])
			require.NotNil(t, mapped.Responses[0].Body.Find(test.mapping))
			plain := root.API.HTTP.Service("plain").Endpoint("exchange")
			require.NotNil(t, plain.Body.Find("display_name"))
			require.NotNil(t, plain.Responses[0].Body.Find("display_name"))
		})
	}
	t.Run("unrelated method identity", func(t *testing.T) {
		identities := make([][]string, 0, 2)
		for _, enabled := range []bool{false, true} {
			root, err := runDSL(t, func() {
				selected := Type("Selected", func() { Attribute("display_name", String) })
				unrelated := Type("Unrelated", func() { Attribute("stable_name", String) })
				Service("example", func() {
					HTTP(func() {
						if enabled {
							BodyFieldName(selected, "display_name:displayName")
						}
					})
					Method("selected", func() {
						Payload(selected)
						HTTP(func() { POST("/selected") })
					})
					Method("unrelated", func() {
						Payload(unrelated)
						Result(unrelated)
						HTTP(func() { POST("/unrelated") })
					})
				})
			})
			require.NoError(t, err)
			endpoint := root.API.HTTP.Service("example").Endpoint("unrelated")
			values := make([]string, 0, 4)
			for _, body := range []*expr.AttributeExpr{endpoint.Body, endpoint.Responses[0].Body} {
				typ := body.Type.(expr.UserType)
				values = append(values, typ.ID(), expr.UserTypeExampleIdentity(typ).Seed())
				require.NotNil(t, body.Find("stable_name"))
			}
			identities = append(identities, values)
		}
		require.Equal(t, identities[0], identities[1])
	})

	for _, reversed := range []bool{false, true} {
		t.Run(fmt.Sprintf("shared child reversed %t", reversed), func(t *testing.T) {
			root, err := runDSL(t, func() {
				item := Type("SharedItem", func() { Attribute("display_name", String) })
				wrapper := Type("Wrapper", func() { Attribute("item", item) })
				envelope := Type("Envelope", func() {
					if reversed {
						Attribute("wrapped", wrapper)
						Attribute("first", item)
					} else {
						Attribute("first", item)
						Attribute("wrapped", wrapper)
					}
				})
				for _, name := range []string{"mapped", "plain"} {
					Service(name, func() {
						HTTP(func() {
							if name == "mapped" {
								BodyFieldName(item, "display_name:displayName")
							}
						})
						Method("exchange", func() {
							Payload(envelope)
							HTTP(func() { POST("/" + name) })
						})
					})
				}
			})
			require.NoError(t, err)
			mapped := root.API.HTTP.Service("mapped").Endpoint("exchange").Body.Find("wrapped")
			plain := root.API.HTTP.Service("plain").Endpoint("exchange").Body.Find("wrapped")
			require.NotEqual(t, plain.Type.(expr.UserType).ID(), mapped.Type.(expr.UserType).ID())
			require.NotNil(t, mapped.Find("item").Find("display_name:displayName"))
			require.NotNil(t, plain.Find("item").Find("display_name"))
		})
	}

	t.Run("flattened union discriminator", func(t *testing.T) {
		for _, excluded := range []bool{false, true} {
			_, err := runDSL(t, func() {
				branch := Type("Branch", func() { Attribute("kind", String) })
				payload := Type("Envelope", func() {
					Attribute("other", String)
					OneOf("choice", func() {
						Meta("oneof:json:flatten")
						Attribute("first", branch)
					})
				})
				Service("mapped", func() {
					HTTP(func() { BodyFieldName(branch, "kind:type") })
					Method("exchange", func() {
						Payload(payload)
						HTTP(func() {
							POST("/mapped")
							if excluded {
								Body("other")
							}
						})
					})
				})
			})
			if excluded {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "discriminator JSON name")
			}
		}
	})
}
