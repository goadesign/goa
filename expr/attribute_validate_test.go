// This file tests validation expressions and rules on attributes.
package expr

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/eval"
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

		fieldNotExistType      = &Object{}
		notAResultType         = Boolean
		viewNotDefinedTypeName = "ViewNotDefinedType"

		errAttributeTypeNil      = fmt.Errorf("attribute type is nil")
		errRequiredFieldNotExist = fmt.Errorf(`%srequired field %q does not exist in type %s`, normalizedCtx, "foo", fieldNotExistType.Name())
		errViewButNotAResultType = fmt.Errorf("%s uses view %q but %q is not a result type", normalizedCtx, metadata["view"][0], notAResultType.Name())
		errTypeNotDefineView     = fmt.Errorf("%s: type %q does not define view %q", normalizedCtx, viewNotDefinedTypeName, "foo")
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
					TypeName: viewNotDefinedTypeName,
					AttributeExpr: &AttributeExpr{
						Type: Boolean,
					},
				},
				Views: []*ViewExpr{
					{Name: "bar"},
				},
			},
			metadata: metadata,
			expected: &eval.ValidationErrors{Errors: []error{errTypeNotDefineView}},
		},
		"custom package in parent type": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &UserTypeExpr{
									AttributeExpr: &AttributeExpr{
										Type: &Object{
											&NamedAttributeExpr{
												Name: "Description",
												Attribute: &AttributeExpr{
													Type: String,
												},
											},
										},
									},
									TypeName: "SecondType",
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"custom packages in child type": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &UserTypeExpr{
									AttributeExpr: &AttributeExpr{
										Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
										Type: &Object{
											&NamedAttributeExpr{
												Name: "Description",
												Attribute: &AttributeExpr{
													Type: String,
												},
											},
										},
									},
									TypeName: "SecondType",
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"matching custom packages between sub-type and parent": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &UserTypeExpr{
									AttributeExpr: &AttributeExpr{
										Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
										Type: &Object{
											&NamedAttributeExpr{
												Name: "Description",
												Attribute: &AttributeExpr{
													Type: String,
												},
											},
										},
									},
									TypeName: "SecondType",
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"conflicting custom packages between sub-type and parent": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &UserTypeExpr{
									AttributeExpr: &AttributeExpr{
										Meta: MetaExpr{"struct:pkg:path": []string{"types2"}},
										Type: &Object{
											&NamedAttributeExpr{
												Name: "Description",
												Attribute: &AttributeExpr{
													Type: String,
												},
											},
										},
									},
									TypeName: "SecondType",
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"conflicting custom packages between sub-type in array and parent": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &Array{
									ElemType: &AttributeExpr{
										Type: &UserTypeExpr{
											AttributeExpr: &AttributeExpr{
												Meta: MetaExpr{"struct:pkg:path": []string{"types2"}},
												Type: &Object{
													&NamedAttributeExpr{
														Name: "Description",
														Attribute: &AttributeExpr{
															Type: String,
														},
													},
												},
											},
											TypeName: "SecondType",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"conflicting custom packages between sub-type in map key and parent": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &Map{
									KeyType: &AttributeExpr{
										Type: &UserTypeExpr{
											AttributeExpr: &AttributeExpr{
												Meta: MetaExpr{"struct:pkg:path": []string{"types2"}},
												Type: &Object{
													&NamedAttributeExpr{
														Name: "Description",
														Attribute: &AttributeExpr{
															Type: String,
														},
													},
												},
											},
											TypeName: "SecondType",
										},
									},
									ElemType: &AttributeExpr{
										Type: String,
									},
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
		},
		"conflicting custom packages between sub-type in map element and parent": {
			typ: &UserTypeExpr{
				TypeName: "FirstType",
				AttributeExpr: &AttributeExpr{
					Meta: MetaExpr{"struct:pkg:path": []string{"types"}},
					Type: &Object{
						&NamedAttributeExpr{
							Name: "thing",
							Attribute: &AttributeExpr{
								Type: &Map{
									KeyType: &AttributeExpr{
										Type: String,
									},
									ElemType: &AttributeExpr{
										Type: &UserTypeExpr{
											AttributeExpr: &AttributeExpr{
												Meta: MetaExpr{"struct:pkg:path": []string{"types2"}},
												Type: &Object{
													&NamedAttributeExpr{
														Name: "Description",
														Attribute: &AttributeExpr{
															Type: String,
														},
													},
												},
											},
											TypeName: "SecondType",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &eval.ValidationErrors{Errors: []error{}},
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
				if len(actual.Errors) > 0 {
					t.Errorf("%#v", actual.Errors[0])
				}
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

func TestAttributeExprValidationValidate(t *testing.T) {
	var (
		max       = 1.0
		exclMax   = 2.0
		min       = 3.0
		ExclMin   = 4.0
		MaxLength = 5
		MinLength = 6
		parent    = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{Type: String},
			TypeName:      "Parent",
		}
	)
	cases := map[string]struct {
		min, max, exclMin, exclMax *float64
		minLen, maxLen             *int
		expected                   string
	}{
		"min and max":         {min: &min, max: &max, expected: "attribute: minimum is greater than maximum"},
		"min and exclMax":     {min: &min, exclMax: &exclMax, expected: "attribute: minimum is greater than or equal to exclusive maximum"},
		"exclMin and max":     {exclMin: &ExclMin, max: &max, expected: "attribute: exclusive minimum is greater than or equal to maximum"},
		"exclMin and exclMax": {exclMin: &ExclMin, exclMax: &exclMax, expected: "attribute: exclusive minimum is greater than exclusive maximum"},
		"max and exclMax":     {max: &max, exclMax: &exclMax, expected: "attribute: both maximum and exclusive maximum are defined"},
		"min and exclMin":     {min: &min, exclMin: &ExclMin, expected: "attribute: both minimum and exclusive minimum are defined"},
		"minLen and maxLen":   {minLen: &MinLength, maxLen: &MaxLength, expected: "attribute: min length is greater than max length"},
	}
	for k, tc := range cases {
		validation := &ValidationExpr{
			Minimum:          tc.min,
			Maximum:          tc.max,
			ExclusiveMinimum: tc.exclMin,
			ExclusiveMaximum: tc.exclMax,
			MinLength:        tc.minLen,
			MaxLength:        tc.maxLen,
		}
		if actual := validation.Validate("", parent); actual.Error() != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual.Error(), tc.expected)
		}
	}
}

func TestAttributeExprValidateChecksEachCall(t *testing.T) {
	parent := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "Parent",
	}
	attribute := &AttributeExpr{
		Type:       &Object{},
		Validation: &ValidationExpr{Required: []string{"missing"}},
	}

	first := attribute.Validate("payload", parent)
	second := attribute.Validate("payload", parent)

	if first == nil {
		t.Error("first check returned no errors, expected 1")
	} else if len(first.Errors) != 1 {
		t.Errorf("first check returned %d errors, expected 1", len(first.Errors))
	}
	if second == nil {
		t.Error("second check returned no errors, expected 1")
	} else if len(second.Errors) != 1 {
		t.Errorf("second check returned %d errors, expected 1", len(second.Errors))
	}
}

func TestAttributeExprValidateChecksSharedTypeOncePerCall(t *testing.T) {
	minLength, maxLength := 2, 1
	shared := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: &Object{
			&NamedAttributeExpr{
				Name: "value",
				Attribute: &AttributeExpr{
					Type: String,
					Validation: &ValidationExpr{
						MinLength: &minLength,
						MaxLength: &maxLength,
					},
				},
			},
		}},
		TypeName: "Shared",
	}
	attribute := &AttributeExpr{Type: &Object{
		&NamedAttributeExpr{
			Name:      "first",
			Attribute: &AttributeExpr{Type: shared},
		},
		&NamedAttributeExpr{
			Name:      "second",
			Attribute: &AttributeExpr{Type: shared},
		},
	}}
	parent := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "Parent",
	}

	checks := []*eval.ValidationErrors{
		attribute.Validate("payload", parent),
		attribute.Validate("payload", parent),
	}
	for i, result := range checks {
		if result == nil {
			t.Errorf("check %d returned no errors, expected 1", i+1)
			continue
		}
		if len(result.Errors) != 1 {
			t.Errorf("check %d returned %d errors, expected 1", i+1, len(result.Errors))
			continue
		}
		if got, want := result.Errors[0].Error(), "field value - min length is greater than max length"; got != want {
			t.Errorf("check %d returned %q, expected %q", i+1, got, want)
		}
	}
}
