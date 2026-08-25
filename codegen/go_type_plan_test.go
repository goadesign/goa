// This file verifies that Go type planning retains every expression-derived
// layout decision before generated package names freeze.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestGoTypePlanRetainsNestedOwners verifies that a binding changes the owner
// inherited by declarations nested beneath the bound occurrence.
func TestGoTypePlanRetainsNestedOwners(t *testing.T) {
	const (
		rootOwner  = "generated.local/gen/service"
		unionOwner = "generated.local/gen/unions"
	)
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	branch := goTypeTestUserType("ChoiceText", expr.String)
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: branch}},
		},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "choice", Attribute: &expr.AttributeExpr{Type: union}},
	}}
	unionDeclaration := declareGoTypeTestUnion(t, generation, unionOwner, union)
	branchDeclaration := declareGoTypeTestUserType(t, generation, unionOwner, branch)
	binder := func(request GoTypeBindingRequest) (GoTypeBinding, error) {
		switch request.Attribute.Type {
		case union:
			require.Equal(t, rootOwner, request.InheritedOwner)
			return GoTypeBinding{Owner: unionOwner, Union: unionDeclaration}, nil
		case branch:
			require.Equal(t, unionOwner, request.InheritedOwner)
			return GoTypeBinding{Owner: request.InheritedOwner, Type: branchDeclaration}, nil
		default:
			return GoTypeBinding{}, fmt.Errorf("unexpected binding for %T", request.Attribute.Type)
		}
	}

	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  rootOwner,
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
		Bind:   binder,
	})
	require.NoError(t, err)
	choice := plan.Fields()[0]
	require.Equal(t, GoUnion, choice.Kind())
	require.Equal(t, unionOwner, choice.Owner())
	require.Same(t, unionDeclaration, choice.UnionDeclaration())
	require.Len(t, choice.Branches(), 1)
	require.Equal(t, unionOwner, choice.Branches()[0].Owner())
	require.Same(t, branchDeclaration, choice.Branches()[0].TypeDeclaration())
	require.Equal(t, []GoTypeImport{{Path: unionOwner}}, plan.ImportPreferences())

	require.NoError(t, generation.Freeze())
	formatter := plan.Link(rootOwner, goTypeTestQualifier)
	require.Equal(t, "struct {\n\tChoice unions.Choice\n}", formatter.Def())
	require.Equal(t, "unions", formatter.Enter(choice).Package())
}

// TestGoTypePlanUnionReferenceOwnsOnlyItsDeclarationImport verifies that a
// file referring to a named union does not import packages used only by the
// separate file that defines the union branches.
func TestGoTypePlanUnionReferenceOwnsOnlyItsDeclarationImport(t *testing.T) {
	const (
		outputOwner = "generated.local/gen/service"
		unionOwner  = "generated.local/gen/unions"
		branchOwner = "example.com/branch"
	)
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "value", Attribute: &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{
					"struct:field:type": {"branch.Value", branchOwner, "branch"},
				},
			}},
		},
	}
	declaration := declareGoTypeTestUnion(t, generation, unionOwner, union)
	plan, err := PlanGoType(&expr.AttributeExpr{Type: union}, GoTypePlanOptions{
		Owner:  outputOwner,
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			union: {Owner: unionOwner, Union: declaration},
		}),
	})
	require.NoError(t, err)
	require.Equal(t, []GoTypeImport{{Path: unionOwner}}, plan.ImportPreferences())

	require.NoError(t, generation.Freeze())
	linked := plan.Link(outputOwner, func(importPath string) string {
		require.Equal(t, unionOwner, importPath)
		return "unions"
	})
	require.Equal(t, "unions.Choice", linked.Name())
	require.Equal(t, []GoTypeImport{{Name: "unions", Path: unionOwner}}, linked.Imports())
}

// TestGoTypePlanRetainsFieldMetadata verifies field names, comments, tags, and
// custom primitive import identity are copied during planning.
func TestGoTypePlanRetainsFieldMetadata(t *testing.T) {
	field := &expr.AttributeExpr{
		Type:        expr.String,
		Description: "stored payload bytes",
		Meta: expr.MetaExpr{
			"struct:field:name":    {"PayloadID"},
			"struct:field:type":    {"json.RawMessage", "encoding/json", "json"},
			"struct:tag:json:name": {"payload_id"},
			"struct:tag:xml":       {"payload"},
		},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "payload", Attribute: field},
	}}

	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	plannedField := plan.Fields()[0]
	require.Equal(t, "PayloadID", plannedField.FieldName(true))
	require.Equal(t, "payloadID", plannedField.FieldName(false))
	require.Equal(t, "stored payload bytes", plannedField.Description())
	require.Equal(t, " `json:\"payload_id,omitempty\" xml:\"payload\"`", plannedField.Tag())
	goImport, ok := plannedField.Import()
	require.True(t, ok)
	require.Equal(t, GoTypeImport{Name: "json", Path: "encoding/json"}, goImport)
	require.Equal(t, []GoTypeImport{{Name: "json", Path: "encoding/json"}}, plan.ImportPreferences())

	formatter := plan.Link("generated.local/gen/service", func(importPath string) string {
		require.Equal(t, "encoding/json", importPath)
		return "json2"
	})
	require.Equal(t, "PayloadID", formatter.Enter(plannedField).Field(true))
	require.Equal(t, "struct {\n\t// stored payload bytes\n\tPayloadID *json2.RawMessage `json:\"payload_id,omitempty\" xml:\"payload\"`\n}", formatter.Def())
}

// TestGoTypePlanRebindsCustomTypeQualifier verifies that linking changes only
// the imported package identifier and preserves the complete authored Go type.
func TestGoTypePlanRebindsCustomTypeQualifier(t *testing.T) {
	const importPath = "example.com/wire"
	tests := []struct {
		name   string
		custom string
		want   string
	}{
		{name: "pointer", custom: "*wire.Value", want: "*wire2.Value"},
		{name: "slice", custom: "[]wire.Value", want: "[]wire2.Value"},
		{name: "nested pointer slice", custom: "[]*wire.Value", want: "[]*wire2.Value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{
				Type: expr.String,
				Meta: expr.MetaExpr{
					"struct:field:type": {test.custom, importPath, "wire"},
				},
			}
			plan, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  "generated.local/gen/service",
				Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
			})
			require.NoError(t, err)

			linked := plan.Link("generated.local/gen/service", func(path string) string {
				require.Equal(t, importPath, path)
				return "wire2"
			})
			require.Equal(t, test.want, linked.Name())
		})
	}
}

// TestGoTypePlanRetainsServiceErrorImport verifies that the built-in service
// error type uses the frozen alias selected for Goa's runtime package.
func TestGoTypePlanRetainsServiceErrorImport(t *testing.T) {
	const (
		owner   = "generated.local/gen/service"
		goaPath = "goa.design/goa/v3/pkg"
	)
	plan, err := PlanGoType(&expr.AttributeExpr{Type: expr.ErrorResult}, GoTypePlanOptions{
		Owner:  owner,
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	require.Equal(t, []GoTypeImport{{Name: "goa", Path: goaPath}}, plan.ImportPreferences())

	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage(owner)
	require.NoError(t, err)
	require.NoError(t, pkg.RequireImport(NewImport("goa", "example.com/fixed/goa")))
	for _, preference := range plan.ImportPreferences() {
		require.NoError(t, pkg.DeclareImport(NewImport(preference.Name, preference.Path)))
	}
	require.NoError(t, generation.Freeze())
	require.Equal(t, "goa2", pkg.ImportName(goaPath))

	linked := plan.Link(owner, pkg.ImportName)
	require.Equal(t, "goa2.ServiceError", linked.Name())
	require.Equal(t, "*goa2.ServiceError", linked.Ref())
	require.Equal(t, []GoTypeImport{{Name: "goa2", Path: goaPath}}, linked.Imports())
}

// TestGoTypePlanRetainsPointerAndDefaultPolicy verifies field indirection is a
// planning decision rather than a formatting-time expression query.
func TestGoTypePlanRetainsPointerAndDefaultPolicy(t *testing.T) {
	withDefault := &expr.AttributeExpr{Type: expr.String, DefaultValue: "ready"}
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "required", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "optional", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			{Name: "defaulted", Attribute: withDefault},
			{Name: "bytes", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
			{Name: "nested", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"required"}},
	}

	tests := []struct {
		name    string
		pointer bool
		want    []bool
		def     string
	}{
		{
			name: "Goa service policy",
			want: []bool{false, true, false, false, true},
			def:  "struct {\n\tRequired string\n\tOptional *int\n\tDefaulted string\n\tBytes []byte\n\tNested *struct {\n}\n}",
		},
		{
			name:    "forced primitive pointers",
			pointer: true,
			want:    []bool{true, true, true, false, true},
			def:     "struct {\n\tRequired *string\n\tOptional *int\n\tDefaulted *string\n\tBytes []byte\n\tNested *struct {\n}\n}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := GoLayoutPolicy{
				Pointer:        test.pointer,
				IgnoreRequired: true,
				UseDefault:     true,
				UnionPointer:   true,
				SumType:        true,
			}
			plan, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  "generated.local/gen/service",
				Policy: policy,
			})
			require.NoError(t, err)
			require.Equal(t, policy, plan.Policy())
			fields := plan.Fields()
			for index, want := range test.want {
				require.Equal(t, want, fields[index].IsPointer(), fields[index].FieldName(true))
			}
			require.Equal(t, test.def, plan.Link(plan.Owner(), goTypeTestQualifier).Def())
		})
	}
}

// TestGoTypePlanReportsReferencePointers verifies that callers can use the
// planned pointer choice without parsing a formatted Go type.
func TestGoTypePlanReportsReferencePointers(t *testing.T) {
	const owner = "example.com/gen/service"
	generation, err := NewGeneration("example.com/gen", nil)
	require.NoError(t, err)
	object := goTypeTestUserType("Message", &expr.Object{})
	alias := goTypeTestUserType("Label", expr.String)
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "label", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}
	binder := goTypeTestBinder(map[expr.DataType]GoTypeBinding{
		object: {Owner: owner, Type: declareGoTypeTestUserType(t, generation, owner, object)},
		alias:  {Owner: owner, Type: declareGoTypeTestUserType(t, generation, owner, alias)},
		union:  {Owner: owner, Union: declareGoTypeTestUnion(t, generation, owner, union)},
	})
	tests := []struct {
		name      string
		attribute *expr.AttributeExpr
		pointer   bool
	}{
		{
			name:      "object",
			attribute: &expr.AttributeExpr{Type: object},
			pointer:   true,
		},
		{
			name:      "anonymous object",
			attribute: &expr.AttributeExpr{Type: &expr.Object{}},
			pointer:   false,
		},
		{
			name:      "union",
			attribute: &expr.AttributeExpr{Type: union},
			pointer:   true,
		},
		{
			name:      "primitive alias",
			attribute: &expr.AttributeExpr{Type: alias},
			pointer:   false,
		},
		{
			name:      "array",
			attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
			pointer:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanGoType(test.attribute, GoTypePlanOptions{
				Owner: owner,
				Bind:  binder,
			})
			require.NoError(t, err)
			require.Equal(t, test.pointer, plan.ReferenceIsPointer())
		})
	}
}

// TestGoTypePlanRetainsRequiredArrayElementPointers verifies that only JSON
// input layouts add pointers to primitive elements that must reject null.
func TestGoTypePlanRetainsRequiredArrayElementPointers(t *testing.T) {
	const owner = "generated.local/gen/types"
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	alias := goTypeTestUserType("Alias", expr.String)
	bytesAlias := goTypeTestUserType("BytesAlias", expr.Bytes)
	binder := goTypeTestBinder(map[expr.DataType]GoTypeBinding{
		alias: {
			Owner: owner,
			Type:  declareGoTypeTestUserType(t, generation, owner, alias),
		},
		bytesAlias: {
			Owner: owner,
			Type:  declareGoTypeTestUserType(t, generation, owner, bytesAlias),
		},
	})
	require.NoError(t, generation.Freeze())

	tests := []struct {
		name     string
		array    *expr.Array
		jsonBody bool
		want     string
	}{
		{
			name:     "built-in string in JSON input",
			array:    &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}, NonNullableElems: true},
			jsonBody: true,
			want:     "[]*string",
		},
		{
			name:     "string alias in JSON input",
			array:    &expr.Array{ElemType: &expr.AttributeExpr{Type: alias}, NonNullableElems: true},
			jsonBody: true,
			want:     "[]*Alias",
		},
		{
			name:     "ordinary string alias array",
			array:    &expr.Array{ElemType: &expr.AttributeExpr{Type: alias}},
			jsonBody: true,
			want:     "[]Alias",
		},
		{
			name:  "service string alias array",
			array: &expr.Array{ElemType: &expr.AttributeExpr{Type: alias}, NonNullableElems: true},
			want:  "[]Alias",
		},
		{
			name:     "bytes alias already represents null",
			array:    &expr.Array{ElemType: &expr.AttributeExpr{Type: bytesAlias}, NonNullableElems: true},
			jsonBody: true,
			want:     "[]BytesAlias",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanGoType(&expr.AttributeExpr{Type: test.array}, GoTypePlanOptions{
				Owner: owner,
				Policy: GoLayoutPolicy{
					UseDefault:          true,
					SumType:             true,
					ArrayElementPointer: test.jsonBody,
				},
				Bind: binder,
			})
			require.NoError(t, err)
			require.Equal(t, test.want, plan.Link(owner, goTypeTestQualifier).Def())
		})
	}
}

// TestGoTypePlanFormatsContainersAndUnions verifies array, map, raw struct,
// named object, and union layouts use their retained child and declaration
// policies after linking.
func TestGoTypePlanFormatsContainersAndUnions(t *testing.T) {
	const owner = "generated.local/gen/types"
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	item := goTypeTestUserType("Item", &expr.Object{})
	choice := &expr.Union{TypeName: "Choice"}
	itemDeclaration := declareGoTypeTestUserType(t, generation, owner, item)
	choiceDeclaration := declareGoTypeTestUnion(t, generation, owner, choice)
	binder := goTypeTestBinder(map[expr.DataType]GoTypeBinding{
		item:   {Owner: owner, Type: itemDeclaration},
		choice: {Owner: owner, Union: choiceDeclaration},
	})
	tests := []struct {
		name     string
		att      *expr.AttributeExpr
		kind     GoTypeKind
		wantName string
		wantDef  string
		wantRef  string
	}{
		{
			name:     "array of named objects",
			att:      &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: item}}},
			kind:     GoArray,
			wantName: "[]*Item",
			wantDef:  "[]*Item",
			wantRef:  "[]*Item",
		},
		{
			name: "map of unions",
			att: &expr.AttributeExpr{Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: choice},
			}},
			kind:     GoMap,
			wantName: "map[string]*Choice",
			wantDef:  "map[string]Choice",
			wantRef:  "map[string]*Choice",
		},
		{
			name:     "array of raw structs",
			att:      &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: &expr.Object{}}}},
			kind:     GoArray,
			wantName: "[]struct {\n}",
			wantDef:  "[]*struct {\n}",
			wantRef:  "[]struct {\n}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanGoType(test.att, GoTypePlanOptions{
				Owner:  owner,
				Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
				Bind:   binder,
			})
			require.NoError(t, err)
			require.Equal(t, test.kind, plan.Kind())
			require.NoError(t, generation.Freeze())
			formatter := plan.Link(owner, goTypeTestQualifier)
			require.Equal(t, test.wantName, formatter.Name())
			require.Equal(t, test.wantDef, formatter.Def())
			require.Equal(t, test.wantRef, formatter.Ref())
		})
	}
}

// TestGoTypePlanIgnoresExpressionMutationAfterPlanning verifies formatting
// never revisits type, metadata, descriptions, tags, or required/default state.
func TestGoTypePlanIgnoresExpressionMutationAfterPlanning(t *testing.T) {
	const owner = "generated.local/gen/service"
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	record := goTypeTestUserType("Record", &expr.Object{})
	declaration := declareGoTypeTestUserType(t, generation, owner, record)
	field := &expr.AttributeExpr{
		Type:        record,
		Description: "the original record",
		Meta: expr.MetaExpr{
			"struct:field:name":    {"Original"},
			"struct:tag:json:name": {"original"},
		},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "record", Attribute: field},
	}}
	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  owner,
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			record: {Owner: owner, Type: declaration},
		}),
	})
	require.NoError(t, err)

	field.Type = expr.Int
	field.Description = "mutated"
	field.DefaultValue = 42
	field.Meta = expr.MetaExpr{
		"struct:field:name":    {"Mutated"},
		"struct:field:type":    {"time.Time", "time", "time"},
		"struct:tag:json:name": {"mutated"},
	}
	attribute.Type = expr.String

	require.NoError(t, generation.Freeze())
	formatter := plan.Link(owner, goTypeTestQualifier)
	require.Equal(t, "struct {\n\t// the original record\n\tOriginal *Record `json:\"original,omitempty\"`\n}", formatter.Def())
	require.Equal(t, "Original", formatter.Enter(plan.Fields()[0]).Field(true))
}

// TestGoTypePlanSeparatesImportPreferencesFromLinkedImports verifies planning
// retains every authored alias request while linked files import each path once.
func TestGoTypePlanSeparatesImportPreferencesFromLinkedImports(t *testing.T) {
	const importPath = "example.com/shared"
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{
			Type: expr.String,
			Meta: expr.MetaExpr{"struct:field:type": {"alpha.Value", importPath, "alpha"}},
		}},
		{Name: "second", Attribute: &expr.AttributeExpr{
			Type: expr.String,
			Meta: expr.MetaExpr{"struct:field:type": {"beta.Value", importPath, "beta"}},
		}},
	}}
	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	require.Equal(t, []GoTypeImport{
		{Name: "alpha", Path: importPath},
		{Name: "beta", Path: importPath},
	}, plan.ImportPreferences())

	linked := plan.Link("generated.local/gen/service", func(path string) string {
		require.Equal(t, importPath, path)
		return "shared2"
	})
	require.Equal(t, []GoTypeImport{{Name: "shared2", Path: importPath}}, linked.Imports())
	require.Equal(t, "struct {\n\tFirst *shared2.Value\n\tSecond *shared2.Value\n}", linked.Def())
}

// TestGoTypePlanEquivalence compares symbolic layouts without relying on the
// expression pointers from which independently retained copies were planned.
func TestGoTypePlanEquivalence(t *testing.T) {
	firstAttribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{
			Type:        expr.String,
			Description: "the retained value",
			Meta: expr.MetaExpr{
				"struct:field:name":    {"ValueID"},
				"struct:tag:json:name": {"value_id"},
			},
		}},
	}}
	secondAttribute := expr.DupAtt(firstAttribute)
	options := GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	}
	first, err := PlanGoType(firstAttribute, options)
	require.NoError(t, err)
	second, err := PlanGoType(secondAttribute, options)
	require.NoError(t, err)
	require.True(t, first.Equivalent(second))
	require.True(t, second.Equivalent(first))

	secondField := (*expr.AsObject(secondAttribute.Type))[0].Attribute
	secondField.Meta["struct:field:name"] = []string{"OtherID"}
	different, err := PlanGoType(secondAttribute, options)
	require.NoError(t, err)
	require.False(t, first.Equivalent(different))
	require.False(t, different.Equivalent(first))
}

// goTypeTestUserType constructs one named type without running the DSL.
func goTypeTestUserType(name string, dataType expr.DataType) expr.UserType {
	return &expr.UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &expr.AttributeExpr{Type: dataType},
	}
}

// declareGoTypeTestUserType adds one exact user type to a generated package.
func declareGoTypeTestUserType(t *testing.T, generation *Generation, owner string, userType expr.UserType) *TypeDeclaration {
	t.Helper()
	generatedPackage, err := generation.ClaimPackage(owner)
	require.NoError(t, err)
	declaration, err := generatedPackage.DeclareUserType(userType)
	require.NoError(t, err)
	return declaration
}

// declareGoTypeTestUnion adds one exact union to a generated package.
func declareGoTypeTestUnion(t *testing.T, generation *Generation, owner string, union *expr.Union) *UnionDeclaration {
	t.Helper()
	generatedPackage, err := generation.ClaimPackage(owner)
	require.NoError(t, err)
	declaration, err := generatedPackage.DeclareUnion(&expr.AttributeExpr{Type: union})
	require.NoError(t, err)
	return declaration
}

// goTypeTestBinder resolves exact test data types to predeclared package records.
func goTypeTestBinder(bindings map[expr.DataType]GoTypeBinding) GoTypeBinder {
	return func(request GoTypeBindingRequest) (GoTypeBinding, error) {
		binding, ok := bindings[request.Attribute.Type]
		if !ok {
			return GoTypeBinding{}, fmt.Errorf("no binding for %T %q", request.Attribute.Type, request.Attribute.Type.Name())
		}
		return binding, nil
	}
}

// goTypeTestQualifier supplies stable aliases for focused type-plan tests.
func goTypeTestQualifier(importPath string) string {
	switch importPath {
	case "generated.local/gen/unions":
		return "unions"
	case "generated.local/gen/types":
		return "types"
	default:
		return ""
	}
}
