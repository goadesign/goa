package expr

import (
	"fmt"
	"testing"

	"goa.design/goa/eval"
)

func TestTaggedAttribute(t *testing.T) {
	cases := map[string]struct {
		a        *AttributeExpr
		expected string
	}{
		"tagged attribute": {
			a: &AttributeExpr{
				Type: &Object{
					&NamedAttributeExpr{
						Name: "Foo",
						Attribute: &AttributeExpr{
							Meta: MetaExpr{
								"foo": []string{"foo"},
							},
						},
					},
				},
			},
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
						Name:      "foo",
						Attribute: &AttributeExpr{},
					},
				},
			},
			expected: "",
		},
	}

	for k, tc := range cases {
		if actual := TaggedAttribute(tc.a, "foo"); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAttributeExprValidate(t *testing.T) {
	var (
		ctx           = "ctx"
		normalizedCtx = ctx + " - "

		validation = &ValidationExpr{
			Required: []string{"foo"},
		}

		metadata = MetaExpr{
			"view": []string{"foo"},
		}

		errAttributeTypeNil      = fmt.Errorf("attribute type is nil")
		errRequiredFieldNotExist = fmt.Errorf(`%srequired field %q does not exist`, normalizedCtx, "foo")
		errViewButNotAResultType = fmt.Errorf("%sdefines a view %v but is not a result type", normalizedCtx, metadata["view"])
		errTypeNotDefineViewe    = fmt.Errorf("%stype does not define view %q", normalizedCtx, "foo")
	)
	cases := map[string]struct {
		typ        DataType
		validation *ValidationExpr
		metadata   MetaExpr
		expected   *eval.ValidationErrors
	}{
		"no error": {
			typ:      Boolean,
			expected: &eval.ValidationErrors{},
		},
		"attribute type is nil": {
			typ:      nil,
			expected: &eval.ValidationErrors{Errors: []error{errAttributeTypeNil}},
		},
		"attribute type is nil in the object": {
			typ: &Object{
				&NamedAttributeExpr{
					Name: "foo",
					Attribute: &AttributeExpr{
						Type: nil,
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{errAttributeTypeNil}},
		},
		"attribute type is nil in the array": {
			typ: &Array{
				ElemType: &AttributeExpr{
					Type: nil,
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{errAttributeTypeNil}},
		},
		"required field does not exist": {
			typ: &Object{
				&NamedAttributeExpr{
					Name: "bar",
					Attribute: &AttributeExpr{
						Type: Boolean,
					},
				},
			},
			validation: validation,
			expected:   &eval.ValidationErrors{Errors: []error{errRequiredFieldNotExist}},
		},
		"required field does not exist in the object": {
			typ: &Object{
				&NamedAttributeExpr{
					Name: "bar",
					Attribute: &AttributeExpr{
						Type: &Object{
							&NamedAttributeExpr{
								Name: "baz",
								Attribute: &AttributeExpr{
									Type: Boolean,
								},
							},
						},
					},
				},
			},
			validation: validation,
			expected:   &eval.ValidationErrors{Errors: []error{errRequiredFieldNotExist}},
		},
		"required field does not exist in the array": {
			typ: &Object{
				&NamedAttributeExpr{
					Name: "bar",
					Attribute: &AttributeExpr{
						Type: &Array{
							ElemType: &AttributeExpr{
								Type: Boolean,
							},
						},
					},
				},
			},
			validation: validation,
			expected:   &eval.ValidationErrors{Errors: []error{errRequiredFieldNotExist}},
		},
		"required field exists in extended attribute": {
			typ: &UserTypeExpr{
				TypeName: "Extended2Attr",
				AttributeExpr: &AttributeExpr{
					Type: &Object{
						&NamedAttributeExpr{
							Name: "bar",
							Attribute: &AttributeExpr{
								Type: &Array{
									ElemType: &AttributeExpr{
										Type: Boolean,
									},
								},
							},
						},
					},
					Bases: []DataType{
						&UserTypeExpr{
							TypeName: "Extended1Attr",
							AttributeExpr: &AttributeExpr{
								Type: &Object{
									&NamedAttributeExpr{
										Name: "foobar",
										Attribute: &AttributeExpr{
											Type: &Array{
												ElemType: &AttributeExpr{
													Type: Boolean,
												},
											},
										},
									},
								},
								Bases: []DataType{
									&UserTypeExpr{
										TypeName: "Attr",
										AttributeExpr: &AttributeExpr{
											Type: &Object{
												&NamedAttributeExpr{
													Name: "foo",
													Attribute: &AttributeExpr{
														Type: &Array{
															ElemType: &AttributeExpr{
																Type: Boolean,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			validation: validation,
			expected:   &eval.ValidationErrors{Errors: []error{}},
		},
		"defines a view but is not a result type": {
			typ:      Boolean,
			metadata: metadata,
			expected: &eval.ValidationErrors{Errors: []error{errViewButNotAResultType}},
		},
		"type does not define view": {
			typ: &ResultTypeExpr{
				UserTypeExpr: &UserTypeExpr{
					AttributeExpr: &AttributeExpr{
						Type: Boolean,
					},
				},
				Views: []*ViewExpr{
					{Name: "bar"},
				},
			},
			metadata: metadata,
			expected: &eval.ValidationErrors{Errors: []error{errTypeNotDefineViewe}},
		},
	}

	for k, tc := range cases {
		attribute := AttributeExpr{
			Type:       tc.typ,
			Validation: tc.validation,
			Meta:       tc.metadata,
		}
		if actual := attribute.Validate(ctx, nil); tc.expected != actual {
			if len(tc.expected.Errors) != len(actual.Errors) {
				t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
				t.Errorf("%#v", actual.Errors[0])
			} else {
				for i, err := range actual.Errors {
					if err.Error() != tc.expected.Errors[i].Error() {
						t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
					}
				}
			}
		}
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
		values    = []interface{}{"foo"}
		pattern   = "^foo$"
		minimum   = 1.1
		maximum   = 2.2
		minLength = 2
		maxLength = 3
	)
	cases := map[string]struct {
		values    []interface{}
		format    ValidationFormat
		pattern   string
		minimum   *float64
		maximum   *float64
		minLength *int
		maxLength *int
		expected  bool
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
		"minimum is not nil": {
			minimum:  &minimum,
			expected: false,
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
			values:    values,
			format:    FormatDate,
			pattern:   pattern,
			minimum:   &minimum,
			maximum:   &maximum,
			minLength: &minLength,
			maxLength: &maxLength,
			expected:  false,
		},
	}

	for k, tc := range cases {
		validation := &ValidationExpr{
			Values:    tc.values,
			Format:    tc.format,
			Pattern:   tc.pattern,
			Minimum:   tc.minimum,
			Maximum:   tc.maximum,
			MinLength: tc.minLength,
			MaxLength: tc.maxLength,
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
