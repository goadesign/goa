package design

import (
	"fmt"
	"testing"

	"goa.design/goa/eval"
)

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
	cases := map[string]struct {
		attName    string
		useDefault bool
		expected   bool
	}{
		"primitive pointer": {
			attName:    "foo",
			useDefault: false,
			expected:   true,
		},
		"no attribute": {
			attName:    "zoo",
			useDefault: false,
			expected:   false,
		},
		"not primitive": {
			attName:    "bar",
			useDefault: false,
			expected:   false,
		},
		"primitive but bytes": {
			attName:    "baz",
			useDefault: false,
			expected:   false,
		},
		"primitive but any": {
			attName:    "qux",
			useDefault: false,
			expected:   false,
		},
		"primitive but required": {
			attName:    "quux",
			useDefault: false,
			expected:   false,
		},
		"primitive but default value": {
			attName:    "corge",
			useDefault: true,
			expected:   false,
		},
	}

	attribute := AttributeExpr{
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
	for k, tc := range cases {
		if actual := attribute.IsPrimitivePointer(tc.attName, tc.useDefault); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
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
