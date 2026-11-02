package expr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaggedAttribute(t *testing.T) {
	tagged := &AttributeExpr{
		Type: &Object{
			&NamedAttributeExpr{
				Name: "Foo",
				Attribute: &AttributeExpr{
					Meta: MetaExpr{"foo": []string{"foo"}},
				},
			},
		},
	}
	cases := map[string]struct {
		a        *AttributeExpr
		expected string
	}{
		"tagged attribute": {
			a:        tagged,
			expected: "Foo",
		},
		"not object": {
			a: &AttributeExpr{
				Type: Boolean,
			},
			expected: "",
		},
		"no meta": {
			a: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "Foo",
						Attribute: &AttributeExpr{},
					},
				},
			},
			expected: "",
		},
		"extended": {
			a: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "bar",
						Attribute: &AttributeExpr{Type: String},
					},
				},
				Bases: []DataType{
					&UserTypeExpr{
						TypeName:      "Extended",
						AttributeExpr: tagged,
					},
				},
			},
			expected: "Foo",
		},
		"recursively extended": {
			a: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "bar",
						Attribute: &AttributeExpr{Type: String},
					},
				},
				Bases: []DataType{
					&UserTypeExpr{
						TypeName: "Extended",
						AttributeExpr: &AttributeExpr{
							Type: &Object{
								&NamedAttributeExpr{
									Name: "foobar",
									Attribute: &AttributeExpr{
										Meta: MetaExpr{"foobar": []string{"foobar"}},
									},
								},
							},
							Bases: []DataType{
								&UserTypeExpr{
									TypeName:      "AnotherExtended",
									AttributeExpr: tagged,
								},
							},
						},
					},
				},
			},
			expected: "Foo",
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := TaggedAttribute(tc.a, "foo"); tc.expected != actual {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}

func TestAttributeExprAllRequired(t *testing.T) {
	cases := map[string]struct {
		typ      DataType
		expected []string
	}{
		"some required": {
			typ: &UserTypeExpr{
				AttributeExpr: &AttributeExpr{
					Validation: &ValidationExpr{
						Required: []string{"foo", "bar"},
					},
				},
			},
			expected: []string{"foo", "bar"},
		},
		"no required": {
			typ:      Boolean,
			expected: nil,
		},
	}

	for k, tc := range cases {
		attribute := AttributeExpr{
			Type: tc.typ,
		}
		if actual := attribute.AllRequired(); len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of all required values to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v != tc.expected[i] {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}

func TestAttributeExprIsRequired(t *testing.T) {
	cases := map[string]struct {
		attName  string
		expected bool
	}{
		"required": {
			attName:  "foo",
			expected: true,
		},
		"not required": {
			attName:  "bar",
			expected: false,
		},
	}

	for k, tc := range cases {
		attribute := AttributeExpr{
			Type: &UserTypeExpr{
				AttributeExpr: &AttributeExpr{
					Validation: &ValidationExpr{
						Required: []string{"foo"},
					},
				},
			},
		}
		if actual := attribute.IsRequired(tc.attName); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprIsRequiredNoDefault(t *testing.T) {
	cases := map[string]struct {
		attName  string
		expected bool
	}{
		"required and no default value": {
			attName:  "foo",
			expected: true,
		},
		"required and default value": {
			attName:  "bar",
			expected: false,
		},
		"not required": {
			attName:  "baz",
			expected: false,
		},
	}

	attribute := AttributeExpr{
		Type: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "foo",
						Attribute: &AttributeExpr{},
					},
					&NamedAttributeExpr{
						Name: "bar",
						Attribute: &AttributeExpr{
							DefaultValue: 1,
						},
					},
					&NamedAttributeExpr{
						Name:      "baz",
						Attribute: &AttributeExpr{},
					},
				},
				Validation: &ValidationExpr{
					Required: []string{"foo", "bar"},
				},
			},
		},
	}
	for k, tc := range cases {
		if actual := attribute.IsRequiredNoDefault(tc.attName); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprIsPrimitivePointer(t *testing.T) {
	var (
		attributeBoolean = AttributeExpr{
			Type: Boolean,
		}
		attributeObject = AttributeExpr{
			Type: &Object{
				&NamedAttributeExpr{
					Name: "foo",
					Attribute: &AttributeExpr{
						Type: String,
					},
				},
				&NamedAttributeExpr{
					Name: "bar",
					Attribute: &AttributeExpr{
						Type: &Array{
							ElemType: &AttributeExpr{
								Type: String,
							},
						},
					},
				},
				&NamedAttributeExpr{
					Name: "baz",
					Attribute: &AttributeExpr{
						Type: Bytes,
					},
				},
				&NamedAttributeExpr{
					Name: "qux",
					Attribute: &AttributeExpr{
						Type: Any,
					},
				},
				&NamedAttributeExpr{
					Name: "quux",
					Attribute: &AttributeExpr{
						Type: String,
					},
				},
				&NamedAttributeExpr{
					Name: "corge",
					Attribute: &AttributeExpr{
						Type:         String,
						DefaultValue: "default",
					},
				},
			},
			Validation: &ValidationExpr{
				Required: []string{"quux"},
			},
		}
	)
	cases := map[string]struct {
		attribute  AttributeExpr
		attName    string
		useDefault bool
		expected   bool
	}{
		"primitive pointer": {
			attribute:  attributeObject,
			attName:    "foo",
			useDefault: false,
			expected:   true,
		},
		"no attribute": {
			attribute:  attributeObject,
			attName:    "zoo",
			useDefault: false,
			expected:   false,
		},
		"not primitive": {
			attribute:  attributeObject,
			attName:    "bar",
			useDefault: false,
			expected:   false,
		},
		"primitive but bytes": {
			attribute:  attributeObject,
			attName:    "baz",
			useDefault: false,
			expected:   false,
		},
		"primitive but any": {
			attribute:  attributeObject,
			attName:    "qux",
			useDefault: false,
			expected:   false,
		},
		"primitive but required": {
			attribute:  attributeObject,
			attName:    "quux",
			useDefault: false,
			expected:   false,
		},
		"primitive but default value": {
			attribute:  attributeObject,
			attName:    "corge",
			useDefault: true,
			expected:   false,
		},
		"non object": {
			attribute:  attributeBoolean,
			attName:    "",    // should have panicked!
			useDefault: false, // should have panicked!
			expected:   false, // should have panicked!
		},
		// Test that non-primitive types (like maps, user types) return false early
		// This verifies the fix that adds an early return check for !IsPrimitive(att.Type)
		"non-primitive-early-return": {
			attribute: AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name: "userType",
						Attribute: &AttributeExpr{
							Type: &UserTypeExpr{
								AttributeExpr: &AttributeExpr{
									Type: &Object{
										&NamedAttributeExpr{
											Name:      "name",
											Attribute: &AttributeExpr{Type: String},
										},
									},
								},
							},
						},
					},
					&NamedAttributeExpr{
						Name: "mapType",
						Attribute: &AttributeExpr{
							Type: &Map{
								KeyType:  &AttributeExpr{Type: String},
								ElemType: &AttributeExpr{Type: String},
							},
						},
					},
				},
			},
			attName:    "userType",
			useDefault: false,
			expected:   false, // Non-primitive types should return false
		},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "non object" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			if actual := tc.attribute.IsPrimitivePointer(tc.attName, tc.useDefault); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}

func TestAttributeExprHasTag(t *testing.T) {
	var (
		tag = "view"
	)
	cases := map[string]struct {
		attribute *AttributeExpr
		tag       string
		expected  bool
	}{
		"has tag": {
			attribute: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name: "foo",
						Attribute: &AttributeExpr{
							Meta: MetaExpr{
								tag: []string{"default"},
							},
						},
					},
				},
			},
			tag:      tag,
			expected: true,
		},
		"attribute expr is nil": {
			attribute: nil,
			tag:       tag,
			expected:  false,
		},
		"not object": {
			attribute: &AttributeExpr{
				Type: String,
			},
			tag:      tag,
			expected: false,
		},
		"object but has no tag": {
			attribute: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "foo",
						Attribute: &AttributeExpr{},
					},
				},
			},
			tag: tag,
		},
	}

	for k, tc := range cases {
		if actual := tc.attribute.HasTag(tc.tag); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprHasTagPrefix(t *testing.T) {
	var (
		tag    = "security:apikey:api_key"
		prefix = "security:apikey"
	)
	cases := map[string]struct {
		attribute *AttributeExpr
		prefix    string
		expected  bool
	}{
		"has tag prefix": {
			attribute: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name: "foo",
						Attribute: &AttributeExpr{
							Meta: MetaExpr{
								tag: []string{"key"},
							},
						},
					},
				},
			},
			prefix:   prefix,
			expected: true,
		},
		"attribute expr is nil": {
			attribute: nil,
			prefix:    prefix,
			expected:  false,
		},
		"not object": {
			attribute: &AttributeExpr{
				Type: String,
			},
			prefix:   prefix,
			expected: false,
		},
		"object but has no tag": {
			attribute: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name:      "foo",
						Attribute: &AttributeExpr{},
					},
				},
			},
			prefix: prefix,
		},
	}

	for k, tc := range cases {
		if actual := tc.attribute.HasTagPrefix(tc.prefix); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprHasDefaultValue(t *testing.T) {
	var (
		object = &Object{
			&NamedAttributeExpr{
				Name: "foo",
				Attribute: &AttributeExpr{
					DefaultValue: 1,
				},
			},
			&NamedAttributeExpr{
				Name:      "bar",
				Attribute: &AttributeExpr{},
			},
		}
	)
	cases := map[string]struct {
		attName  string
		typ      DataType
		expected bool
	}{
		"has default value": {
			attName:  "foo",
			typ:      object,
			expected: true,
		},
		"no default value": {
			attName:  "bar",
			typ:      object,
			expected: false,
		},
		"not object": {
			typ:      Boolean,
			expected: false,
		},
	}

	for k, tc := range cases {
		attribute := AttributeExpr{
			Type: tc.typ,
		}
		if actual := attribute.HasDefaultValue(tc.attName); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestValidationExprHasRequiredOnly(t *testing.T) {
	var (
		values           = []any{"foo"}
		pattern          = "^foo$"
		exclusiveMinimum = 1.1
		minimum          = 1.1
		exclusiveMaximum = 2.2
		maximum          = 2.2
		minLength        = 2
		maxLength        = 3
	)
	cases := map[string]struct {
		values           []any
		format           ValidationFormat
		pattern          string
		exclusiveMinimum *float64
		minimum          *float64
		exclusiveMaximum *float64
		maximum          *float64
		minLength        *int
		maxLength        *int
		expected         bool
	}{
		"has required only": {
			expected: true,
		},
		"values is not nil": {
			values:   values,
			expected: false,
		},
		"format is not empty": {
			format:   FormatDate,
			expected: false,
		},
		"pattern is not empty": {
			pattern:  pattern,
			expected: false,
		},
		"exclusiveMinimum is not nil": {
			exclusiveMinimum: &exclusiveMinimum,
			expected:         false,
		},
		"minimum is not nil": {
			minimum:  &minimum,
			expected: false,
		},
		"exclusiveMaximum is not nil": {
			exclusiveMaximum: &exclusiveMaximum,
			expected:         false,
		},
		"maximum is not nil": {
			maximum:  &maximum,
			expected: false,
		},
		"min length is not nil": {
			minLength: &minLength,
			expected:  false,
		},
		"max length is not nil": {
			maxLength: &maxLength,
			expected:  false,
		},
		"complex validation": {
			values:           values,
			format:           FormatDate,
			pattern:          pattern,
			exclusiveMinimum: &exclusiveMinimum,
			minimum:          &minimum,
			exclusiveMaximum: &exclusiveMaximum,
			maximum:          &maximum,
			minLength:        &minLength,
			maxLength:        &maxLength,
			expected:         false,
		},
	}

	for k, tc := range cases {
		validation := &ValidationExpr{
			Values:           tc.values,
			Format:           tc.format,
			Pattern:          tc.pattern,
			ExclusiveMinimum: tc.exclusiveMinimum,
			Minimum:          tc.minimum,
			ExclusiveMaximum: tc.exclusiveMaximum,
			Maximum:          tc.maximum,
			MinLength:        tc.minLength,
			MaxLength:        tc.maxLength,
		}
		if actual := validation.HasRequiredOnly(); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprEvalName(t *testing.T) {
	cases := map[string]struct {
		expected string
	}{
		"testcase": {expected: "attribute"},
	}
	for key, testcase := range cases {
		attribute := AttributeExpr{}
		if actual := attribute.EvalName(); actual != testcase.expected {
			t.Errorf("%s: got %#v, expected %#v", key, actual, testcase.expected)
		}
	}
}

// TestAttributeExprFinalizeMultiLevelInheritanceWithBackEdges verifies that
// Finalize merges transitive base types (RootBase -> Ancestor -> Base ->
// Derived) into a child attribute on Node even when Ancestor has a value field
// referencing Node directly or through an array, map, or union.
func TestAttributeExprFinalizeMultiLevelInheritanceWithBackEdges(t *testing.T) {
	shapes := []string{"object", "array", "map", "union"}
	orders := []bool{false, true} // false: self first, true: nested first
	levels := []bool{false, true} // false: 3 levels (Ancestor -> Base -> Derived), true: 4 levels (RootBase -> Ancestor -> Base -> Derived)

	for _, shape := range shapes {
		for _, nestedFirst := range orders {
			for _, hasRootBase := range levels {
				name := fmt.Sprintf("shape=%s/nestedFirst=%v/hasRootBase=%v", shape, nestedFirst, hasRootBase)
				t.Run(name, func(t *testing.T) {
					var rootBase *UserTypeExpr
					if hasRootBase {
						rootBase = &UserTypeExpr{
							TypeName: "RootBase",
							AttributeExpr: &AttributeExpr{
								Type: &Object{
									{Name: "root_field", Attribute: &AttributeExpr{Type: String}},
								},
							},
						}
					}

					node := &UserTypeExpr{
						TypeName:      "Node",
						AttributeExpr: &AttributeExpr{Type: &Object{}},
					}

					var backEdgeType DataType
					switch shape {
					case "object":
						backEdgeType = node
					case "array":
						backEdgeType = &Array{ElemType: &AttributeExpr{Type: node}}
					case "map":
						backEdgeType = &Map{KeyType: &AttributeExpr{Type: String}, ElemType: &AttributeExpr{Type: node}}
					case "union":
						backEdgeType = &Union{
							TypeName: "BackChoice",
							Values: []*NamedAttributeExpr{
								{Name: "node", Attribute: &AttributeExpr{Type: node}},
							},
						}
					}

					minLen := 2
					ancestorAtt := &AttributeExpr{
						Type: &Object{
							{Name: "query", Attribute: &AttributeExpr{
								Type:       String,
								Validation: &ValidationExpr{MinLength: &minLen},
							}},
							{Name: "back", Attribute: &AttributeExpr{Type: backEdgeType}},
						},
						Validation: &ValidationExpr{Required: []string{"query"}},
					}
					if hasRootBase {
						ancestorAtt.Bases = []DataType{rootBase}
					}

					ancestor := &UserTypeExpr{
						TypeName:      "Ancestor",
						AttributeExpr: ancestorAtt,
					}

					base := &UserTypeExpr{
						TypeName: "Base",
						AttributeExpr: &AttributeExpr{
							Type:  &Object{},
							Bases: []DataType{ancestor},
						},
					}

					derived := &UserTypeExpr{
						TypeName: "Derived",
						AttributeExpr: &AttributeExpr{
							Type:  &Object{},
							Bases: []DataType{base},
						},
					}

					nestedAtt := &AttributeExpr{
						Type:  &Object{},
						Bases: []DataType{derived},
					}
					selfField := &NamedAttributeExpr{Name: "self", Attribute: &AttributeExpr{Type: node}}
					nestedField := &NamedAttributeExpr{Name: "nested", Attribute: nestedAtt}

					nodeObj := AsObject(node.Type)
					if nestedFirst {
						*nodeObj = []*NamedAttributeExpr{nestedField, selfField}
					} else {
						*nodeObj = []*NamedAttributeExpr{selfField, nestedField}
					}

					// Finalize in declaration order: (RootBase), Ancestor, Base, Derived, Node
					if hasRootBase {
						rootBase.Finalize()
					}
					ancestor.Finalize()
					base.Finalize()
					derived.Finalize()
					node.Finalize()

					// Verify nested in Node retained ancestor properties and validations
					nestedObj := AsObject(nestedAtt.Type)
					require.NotNil(t, nestedObj)
					queryAtt := nestedObj.Attribute("query")
					require.NotNil(t, queryAtt, "nested must retain query attribute from Ancestor")
					assert.Equal(t, String, queryAtt.Type)
					require.NotNil(t, queryAtt.Validation)
					require.NotNil(t, queryAtt.Validation.MinLength)
					assert.Equal(t, 2, *queryAtt.Validation.MinLength)
					assert.True(t, nestedAtt.IsRequired("query"), "nested must retain required validation for query")

					if hasRootBase {
						rootField := nestedObj.Attribute("root_field")
						require.NotNil(t, rootField, "nested must retain root_field from RootBase")
						assert.Equal(t, String, rootField.Type)
					}
				})
			}
		}
	}
}

// TestAttributeExprFinalizeReferenceAndExtendPreservesBaseDefault verifies that
// combining Reference and Extend at top-level or nested object attributes
// applies Reference before Extend without mutating shared base defaults, and
// keeps Required validations on Extend without propagating Base Required lists
// into unfinalized Reference targets.
func TestAttributeExprFinalizeReferenceAndExtendPreservesBaseDefault(t *testing.T) {
	requiredAncestor := &UserTypeExpr{
		TypeName: "RequiredAncestor",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "status", Attribute: &AttributeExpr{
					Type:         String,
					DefaultValue: "active",
				}},
			},
			Validation: &ValidationExpr{Required: []string{"status"}},
		},
	}

	sharedBase := &UserTypeExpr{
		TypeName: "SharedBase",
		AttributeExpr: &AttributeExpr{
			Type:  &Object{},
			Bases: []DataType{requiredAncestor},
		},
	}

	referencing := &UserTypeExpr{
		TypeName: "ReferencingType",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "status", Attribute: &AttributeExpr{
					Type:         String,
					DefaultValue: "custom",
				}},
			},
			References: []DataType{sharedBase},
		},
	}

	extending := &UserTypeExpr{
		TypeName: "ExtendingType",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "name", Attribute: &AttributeExpr{
					Type: String,
				}},
			},
			Bases: []DataType{sharedBase},
		},
	}

	// Repeated Finalize calls across referencing, extending, and sharedBase
	for i := 0; i < 3; i++ {
		referencing.Finalize()
		extending.Finalize()
		sharedBase.Finalize()
	}

	baseStatus := AsObject(sharedBase.Type).Attribute("status")
	require.NotNil(t, baseStatus)
	assert.Equal(t, "active", baseStatus.DefaultValue, "shared base default value must not be contaminated")

	refStatus := AsObject(referencing.Type).Attribute("status")
	require.NotNil(t, refStatus)
	assert.Equal(t, "custom", refStatus.DefaultValue, "referencing attribute must keep its authored default value")
	assert.False(t, referencing.IsRequired("status"), "referencing attribute must not inherit Required from unfinalized Extend base")

	extStatus := AsObject(extending.Type).Attribute("status")
	require.NotNil(t, extStatus)
	assert.Equal(t, "active", extStatus.DefaultValue, "extending attribute must inherit default value from base")
	assert.True(t, extending.IsRequired("status"), "extending attribute must inherit Required from base")

	nestedBase := &UserTypeExpr{
		TypeName: "NestedBase",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "query", Attribute: &AttributeExpr{Type: String}},
			},
		},
	}
	nestedParent := &UserTypeExpr{
		TypeName: "NestedParent",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "settings", Attribute: &AttributeExpr{
					Type: &Object{
						{Name: "options", Attribute: &AttributeExpr{
							Type: &Object{
								{Name: "query", Attribute: &AttributeExpr{
									Type:         String,
									DefaultValue: "reference",
								}},
							},
						}},
					},
				}},
			},
		},
	}
	options := &AttributeExpr{
		Type: &Object{
			{Name: "query", Attribute: &AttributeExpr{Type: String}},
		},
		Bases: []DataType{nestedBase},
	}
	child := &AttributeExpr{
		Type: &Object{
			{Name: "settings", Attribute: &AttributeExpr{
				Type: &Object{
					{Name: "options", Attribute: options},
				},
			}},
		},
		References: []DataType{nestedParent},
	}
	child.Finalize()
	assert.Nil(t, nestedBase.Find("query").DefaultValue, "nested Reference must not contaminate shared Base default")
	assert.Nil(t, options.Find("query").DefaultValue, "nested Extend must override outer Reference default")
}
