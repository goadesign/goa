// This file verifies command-line payload builders validate the concrete Go
// values produced from flag text. The tests catch regressions where validation
// is generated for a pointer and then edited as source text.
package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// TestParserVariablesKeepCollidingMethodsDistinct catches generated parsers
// that rebuild local variable names from method text. The two command names are
// different, but both become StatusUpdate when converted to a Go identifier.
func TestParserVariablesKeepCollidingMethodsDistinct(t *testing.T) {
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage("generated.local/gen/jsonrpc/cli/server")
	require.NoError(t, err)
	parser, err := DeclareParser(pkg, "jsonrpc", "api", "server", []CommandDeclarationInput{
		{Service: "notifications", Methods: []string{"status_update", "status+update"}},
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	const preferred = "notificationsStatusUpdate"
	value := NewFlagPlan(&expr.AttributeExpr{Type: expr.String}, "string", "string", nil).value
	firstFlag := &FlagData{Name: "body", FullName: preferred + "Body", Type: "STRING", value: value}
	secondFlag := &FlagData{Name: "body", FullName: preferred + "Body", Type: "STRING", value: value}
	builder := &BuildFunctionData{
		ActualParams: []string{preferred + "Body"},
		FormalParams: []string{preferred + "Body"},
	}
	command := &CommandData{
		ServiceName:      "notifications",
		Name:             "notifications",
		VarName:          "notifications",
		UsageDeclaration: parser.Commands["notifications"].Usage,
		Subcommands: []*SubcommandData{
			{
				MethodName:       "status_update",
				Name:             "status-update",
				FullName:         preferred,
				UsageDeclaration: parser.Commands["notifications"].Methods["status_update"],
				Flags:            []*FlagData{firstFlag},
				conversionFlag:   firstFlag,
			},
			{
				MethodName:       "status+update",
				Name:             "status+update",
				FullName:         preferred,
				UsageDeclaration: parser.Commands["notifications"].Methods["status+update"],
				Flags:            []*FlagData{secondFlag},
				BuildFunction:    builder,
			},
		},
	}

	parser.PlanVariables([]*CommandData{command}, nil)
	generated := parser.FlagsCode([]*CommandData{command})

	require.Equal(t, "notificationsFlags", command.FlagSetVar)
	require.Equal(t, preferred+"Flags2", command.Subcommands[0].FlagSetVar)
	require.Equal(t, preferred+"Flags", command.Subcommands[1].FlagSetVar)
	require.Equal(t, preferred+"BodyFlag2", firstFlag.PointerVar)
	require.Equal(t, preferred+"BodyFlag", secondFlag.PointerVar)
	require.Equal(t, "data = *"+preferred+"BodyFlag2", command.Subcommands[0].Conversion)
	require.Equal(t, []string{preferred + "Body"}, builder.ActualParams)
	require.Equal(t, []string{preferred + "BodyFlag"}, command.Subcommands[1].ActualPointerVars)
	require.Equal(t, 4, strings.Count(generated, preferred+"Flags2"))
	require.Equal(t, 1, strings.Count(generated, preferred+"BodyFlag2 ="))

	reversedPlus := &FlagData{Name: "body", FullName: preferred + "Body"}
	reversedUnderscore := &FlagData{Name: "body", FullName: preferred + "Body"}
	reversed := &CommandData{
		ServiceName: "notifications",
		VarName:     "notifications",
		Subcommands: []*SubcommandData{
			{MethodName: "status+update", FullName: preferred, Flags: []*FlagData{reversedPlus}},
			{MethodName: "status_update", FullName: preferred, Flags: []*FlagData{reversedUnderscore}},
		},
	}
	parser.PlanVariables([]*CommandData{reversed}, nil)
	require.Equal(t, preferred+"Flags", reversed.Subcommands[0].FlagSetVar)
	require.Equal(t, preferred+"Flags2", reversed.Subcommands[1].FlagSetVar)
	require.Equal(t, preferred+"BodyFlag", reversedPlus.PointerVar)
	require.Equal(t, preferred+"BodyFlag2", reversedUnderscore.PointerVar)
}

// TestBuildCommandDataUsesPlannedServicePath checks that the public command
// name matches the unique path chosen while service packages are planned.
func TestBuildCommandDataUsesPlannedServicePath(t *testing.T) {
	tests := []struct {
		name    string
		service *service.Data
		command string
	}{
		{
			name:    "ordinary service",
			service: &service.Data{Name: "Calculator", PathName: "calculator"},
			command: "calculator",
		},
		{
			name:    "first colliding service",
			service: &service.Data{Name: "mcp_read_value", PathName: "mcp_read_value"},
			command: "mcp-read-value",
		},
		{
			name:    "second colliding service",
			service: &service.Data{Name: "mcp-read-value", PathName: "mcp_read_value2"},
			command: "mcp-read-value2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := BuildCommandData(test.service)
			require.Equal(t, test.command, command.Name)
		})
	}
}

func TestFieldLoadCodeValidationTarget(t *testing.T) {
	minimum := 1.0
	minLength := 1
	cases := []struct {
		name            string
		flag            *FlagData
		argument        string
		attribute       *expr.AttributeExpr
		typeName        string
		typeRef         string
		wantTarget      string
		wantCondition   string
		wantError       string
		avoidValidation string
	}{
		{
			name:            "optional integer",
			flag:            &FlagData{FullName: "serviceMethodCount", Type: "INT"},
			argument:        "count",
			typeName:        "int",
			typeRef:         "int",
			attribute:       &expr.AttributeExpr{Type: expr.Int, Validation: &expr.ValidationExpr{Minimum: &minimum}},
			wantTarget:      "val",
			wantCondition:   "if val < 1 {",
			wantError:       `goa.InvalidRangeError("count", val, 1, true)`,
			avoidValidation: "if count != nil",
		},
		{
			name:            "optional string",
			flag:            &FlagData{FullName: "serviceMethodState", Type: "STRING"},
			argument:        "state",
			typeName:        "string",
			typeRef:         "string",
			attribute:       &expr.AttributeExpr{Type: expr.String, Validation: &expr.ValidationExpr{Values: []any{"ready"}}},
			wantTarget:      "serviceMethodState",
			wantCondition:   `if !(serviceMethodState == "ready") {`,
			wantError:       `goa.InvalidEnumValueError("state", serviceMethodState, []any{"ready"})`,
			avoidValidation: "if state != nil",
		},
		{
			name:            "optional JSON array",
			flag:            &FlagData{FullName: "serviceMethodItems", Type: "JSON", Example: "[]"},
			argument:        "items",
			typeName:        "[]string",
			typeRef:         "[]string",
			attribute:       &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}, Validation: &expr.ValidationExpr{MinLength: &minLength}},
			wantTarget:      "items",
			wantCondition:   "if len(items) < 1 {",
			wantError:       `goa.InvalidLengthError("items", items, len(items), 1, true)`,
			avoidValidation: "if items != nil",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			context := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
			var target string
			validation := func(value string) string {
				target = value
				return codegen.AttributeValidationCode(
					test.attribute,
					nil,
					context,
					true,
					false,
					value,
					test.argument,
				)
			}
			value := NewFlagPlan(test.attribute, test.typeName, test.typeRef, nil).value

			generated, declaresError := fieldLoadCode(
				test.flag,
				test.argument,
				value,
				validation,
				nil,
				&expr.Object{},
				"*Payload",
			)

			require.Equal(t, test.wantTarget, target)
			require.Contains(t, generated, test.wantCondition)
			require.Contains(t, generated, test.wantError)
			require.NotContains(t, generated, test.avoidValidation)
			require.True(t, declaresError)
		})
	}
}

// TestFlagArgumentTypeNameMatchesValuePlan checks the released type-name field
// against the value description used to generate the flag conversion.
func TestFlagArgumentTypeNameMatchesValuePlan(t *testing.T) {
	plan := NewFlagPlan(&expr.AttributeExpr{Type: expr.String}, "Label", "Label", nil)
	argument := &FlagArgData{Plan: plan, TypeName: plan.value.typeName}
	require.Equal(t, plan.value.typeName, argument.TypeName)
}

// TestLegacyFlagValidationIsEmitted verifies that plugins using the released
// Validate field still write their validation code into payload builders.
func TestLegacyFlagValidationIsEmitted(t *testing.T) {
	_, builder := MakeFlags(
		"Service",
		&service.MethodData{Name: "Method", VarName: "Method"},
		[]*FlagArgData{{
			Name:     "value",
			TypeName: "string",
			TypeRef:  "string",
			Required: true,
			Validate: "if value == \"\" {\n\terr = goa.MissingFieldError(\"value\", \"payload\")\n}",
		}},
		&expr.Object{},
		"*Payload",
		nil,
	)

	require.Contains(t, builder.Fields[0].Init, "if value == \"\"")
	require.Contains(t, builder.Fields[0].Init, "goa.MissingFieldError")
	require.Equal(t, "BuildMethodPayload", builder.Name)
	require.True(t, builder.CheckErr)
}

// TestFlagValidationFormsAreExclusive catches generators that silently choose
// one validation form when a plugin supplies both forms.
func TestFlagValidationFormsAreExclusive(t *testing.T) {
	plan := NewFlagPlan(&expr.AttributeExpr{Type: expr.String}, "string", "string", func(string) string { return "typed" })
	require.PanicsWithValue(t, "CLI flag validation cannot use both Validate and Plan", func() {
		MakeFlags(
			"Service",
			&service.MethodData{Name: "Method"},
			[]*FlagArgData{{
				Name:     "value",
				Plan:     plan,
				TypeRef:  "string",
				Required: true,
				Validate: "legacy",
			}},
			&expr.Object{},
			"*Payload",
			nil,
		)
	})
}

// TestUsageSectionsKeepReleasedData checks that plugins still receive a string
// slice when they replace the source of a generated help section.
func TestUsageSectionsKeepReleasedData(t *testing.T) {
	command := &CommandData{Name: "calc", Subcommands: []*SubcommandData{{Name: "add"}}, Example: "calc add"}
	require.IsType(t, []string{}, UsageCommands([]*CommandData{command}).Data)
	require.IsType(t, []string{}, UsageExamples([]*CommandData{command}).Data)
}

// TestPrintDescriptionLeavesBlankLinesEmpty catches generated help text that
// puts tabs on otherwise empty lines.
func TestPrintDescriptionLeavesBlankLinesEmpty(t *testing.T) {
	require.Equal(t, "First line.\n\n\tSecond paragraph.", printDescription("First line.\n\nSecond paragraph.\n\t\n"))
}

// TestReleasedFunctionSignaturesCompile keeps the public CLI helper calls used
// by plugins source compatible.
func TestReleasedFunctionSignaturesCompile(t *testing.T) {
	var buildCommand func(*service.Data) *CommandData = BuildCommandData
	var endpointFile func(string, string, []*codegen.ImportSpec, []*CommandData, *codegen.SectionTemplate) *codegen.File = EndpointParserFile
	var usageCommands func([]*CommandData) *codegen.SectionTemplate = UsageCommands
	var usageExamples func([]*CommandData) *codegen.SectionTemplate = UsageExamples
	var flagsCode func([]*CommandData) string = FlagsCode
	var newFlag func(string, string, string, string, string, bool, any, any) *FlagData = NewFlagData
	var fieldLoad func(*FlagData, string, string, string, any, expr.DataType, string) (string, bool) = FieldLoadCode

	require.NotNil(t, buildCommand)
	require.NotNil(t, endpointFile)
	require.NotNil(t, usageCommands)
	require.NotNil(t, usageExamples)
	require.NotNil(t, flagsCode)
	require.NotNil(t, newFlag)
	require.NotNil(t, fieldLoad)
}

// TestReleasedFlagHelpersGenerateConversions checks the string-based flag
// helpers still produce the conversion and validation requested by plugins.
func TestReleasedFlagHelpersGenerateConversions(t *testing.T) {
	flag := NewFlagData("Service", "Method", "count", "int32", "", true, int32(1), nil)
	generated, declaresError := FieldLoadCode(
		flag,
		"count",
		"int32",
		"if count < 1 {\n\terr = goa.InvalidRangeError(\"count\", count, 1, true)\n}",
		nil,
		&expr.Object{},
		"*Payload",
	)

	require.Equal(t, "INT32", flag.Type)
	require.Contains(t, generated, "strconv.ParseInt(serviceMethodCount, 10, 32)")
	require.Contains(t, generated, "if count < 1")
	require.True(t, declaresError)
}

// TestReleasedFlagsCodeUsesReleasedTemplateData checks that plugins can still
// render flags without creating a parser plan.
func TestReleasedFlagsCodeUsesReleasedTemplateData(t *testing.T) {
	command := &CommandData{
		Name:    "calc",
		VarName: "calc",
		Subcommands: []*SubcommandData{{
			Name:     "add",
			FullName: "calcAdd",
			Flags: []*FlagData{{
				Name:     "value",
				FullName: "calcAddValue",
			}},
		}},
	}
	generated := FlagsCode([]*CommandData{command})
	require.Contains(t, generated, `calcFlags = flag.NewFlagSet("calc"`)
	require.Contains(t, generated, `calcAddValueFlag = calcAddFlags.String("value"`)
}

func TestFieldLoadCodeRequiredIntegerValidationUsesLoadedValue(t *testing.T) {
	maximum := 42.0
	attribute := &expr.AttributeExpr{
		Type:       expr.Int32,
		Validation: &expr.ValidationExpr{Maximum: &maximum},
	}
	context := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	var target string
	validation := func(value string) string {
		target = value
		return codegen.AttributeValidationCode(attribute, nil, context, true, false, value, "count")
	}

	generated, declaresError := fieldLoadCode(
		&FlagData{FullName: "serviceMethodCount", Type: "INT32", Required: true},
		"count",
		NewFlagPlan(attribute, "int32", "int32", nil).value,
		validation,
		nil,
		&expr.Object{},
		"*Payload",
	)

	require.Equal(t, "count", target)
	require.Contains(t, generated, "if count > 42")
	require.True(t, declaresError)
}

func TestPrimitiveAliasFlagPlan(t *testing.T) {
	alias := &expr.UserTypeExpr{
		TypeName: "Count",
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.Int32,
		},
	}
	attribute := &expr.AttributeExpr{Type: alias}
	value := NewFlagPlan(attribute, "Count", "service.Count", nil).value
	flag := newFlagData("Service", "Method", "count", value, "", false, int32(3), nil)

	generated, declaresError := fieldLoadCode(
		flag,
		"count",
		value,
		nil,
		nil,
		&expr.Object{},
		"*Payload",
	)

	require.Equal(t, "INT32", flag.Type)
	require.Contains(t, generated, "strconv.ParseInt(serviceMethodCount, 10, 32)")
	require.Contains(t, generated, "val := service.Count(v)")
	require.Contains(t, generated, "count = &val")
	require.NotContains(t, generated, "json.Unmarshal")
	require.True(t, declaresError)
}

func TestCompositeFlagPlanUsesJSON(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}},
	}
	value := NewFlagPlan(attribute, "[]string", "[]string", nil).value
	flag := newFlagData("Service", "Method", "items", value, "", false, []string{"one"}, nil)

	generated, declaresError := fieldLoadCode(
		flag,
		"items",
		value,
		nil,
		nil,
		&expr.Object{},
		"*Payload",
	)

	require.Equal(t, "JSON", flag.Type)
	require.Contains(t, generated, "json.Unmarshal([]byte(serviceMethodItems), &items)")
	require.True(t, declaresError)
}

func TestStringAndBytesAliasesUseStringFlags(t *testing.T) {
	cases := []struct {
		name          string
		primitive     expr.DataType
		typeName      string
		typeRef       string
		wantGenerated string
	}{
		{
			name:          "string alias",
			primitive:     expr.String,
			typeName:      "Label",
			typeRef:       "service.Label",
			wantGenerated: "val := service.Label(serviceMethodValue)\nvalue = &val",
		},
		{
			name:          "bytes alias",
			primitive:     expr.Bytes,
			typeName:      "Blob",
			typeRef:       "service.Blob",
			wantGenerated: "value = service.Blob(serviceMethodValue)",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			alias := &expr.UserTypeExpr{
				TypeName:      test.typeName,
				AttributeExpr: &expr.AttributeExpr{Type: test.primitive},
			}
			attribute := &expr.AttributeExpr{Type: alias}
			value := NewFlagPlan(attribute, test.typeName, test.typeRef, nil).value
			flag := newFlagData("Service", "Method", "value", value, "", false, "value", nil)

			generated, _ := fieldLoadCode(
				flag,
				"value",
				value,
				nil,
				nil,
				&expr.Object{},
				"*Payload",
			)

			require.Equal(t, "STRING", flag.Type)
			require.Contains(t, generated, test.wantGenerated)
			require.NotContains(t, generated, "json.Unmarshal")
		})
	}
}

func TestCustomGoTypeFlagPlanUsesJSON(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{
			"struct:field:type": []string{"time.Time", "time"},
		},
	}
	value := NewFlagPlan(attribute, "time.Time", "time.Time", nil).value
	flag := newFlagData("Service", "Method", "at", value, "", true, "2026-08-22T00:00:00Z", nil)

	generated, declaresError := fieldLoadCode(
		flag,
		"at",
		value,
		nil,
		nil,
		&expr.Object{},
		"*Payload",
	)

	require.Equal(t, "JSON", flag.Type)
	require.Contains(t, generated, "json.Unmarshal([]byte(serviceMethodAt), &at)")
	require.True(t, declaresError)
}
