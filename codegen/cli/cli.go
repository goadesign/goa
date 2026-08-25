// Package cli contains helpers used by transport-specific command-line client
// generators for parsing the command-line flags to identify the service and
// the method to make a request along with the request payload to be sent.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// CommandData contains the data needed to render a command.
	CommandData struct {
		// ServiceName is the design service selected by this command.
		ServiceName string
		// UsageDeclaration is the package function that prints help for this command.
		UsageDeclaration *codegen.NameDeclaration
		// Name of command e.g. "cellar-storage"
		Name string
		// VarName is the name of the command variable e.g.
		// "cellarStorage"
		VarName string
		// FlagSetVar is the exact local variable that stores this command's flags
		// in the generated endpoint parser.
		FlagSetVar string
		// Description is the help text.
		Description string
		// Subcommands is the list of endpoint commands.
		Subcommands []*SubcommandData
		// Example is a valid command invocation, starting with the
		// command name.
		Example string
		// PkgName is the transport client package import name, e.g. "storagec".
		PkgName string
		// Interceptors contains the data for client interceptors if any.
		Interceptors *InterceptorData
	}

	// SubcommandData contains the data needed to render a sub-command.
	SubcommandData struct {
		// MethodName is the design method selected by this command.
		MethodName string
		// UsageDeclaration is the package function that prints help for this subcommand.
		UsageDeclaration *codegen.NameDeclaration
		// Name is the sub-command name e.g. "add"
		Name string
		// FullName is the sub-command full name e.g. "storageAdd"
		FullName string
		// FlagSetVar is the exact local variable that stores this method's flags
		// in the generated endpoint parser.
		FlagSetVar string
		// Description is the help text.
		Description string
		// Flags is the list of flags supported by the subcommand.
		Flags []*FlagData
		// MethodVarName is the endpoint method name, e.g. "Add"
		MethodVarName string
		// BuildFunction contains the data to generate a payload builder function
		// if any. Exclusive with Conversion.
		BuildFunction *BuildFunctionData
		// ActualPointerVars lists the parser variables used by plugins that render
		// the released BuildFunction call.
		ActualPointerVars []string
		// ActualArgs lists the exact expressions passed to BuildFunction. It is
		// empty for plugins that use the released pointer-variable contract.
		ActualArgs []string
		// Conversion contains the flag value to payload conversion function if
		// any. Exclusive with BuildFunction.
		Conversion string
		// Example is a valid command invocation, starting with the command name.
		Example string
		// Interceptors contains the data for client interceptors if any apply to the endpoint method.
		Interceptors *InterceptorData
		// conversionFlag is the parsed flag converted directly into a primitive payload.
		conversionFlag *FlagData
	}

	// InterceptorData contains the data needed to generate interceptor code.
	InterceptorData struct {
		// VarName is the name of the interceptor variable.
		VarName string
		// ParserVar is the exact parameter name used by the generated endpoint parser.
		ParserVar string
		// PkgName is the package name containing the interceptor type.
		PkgName string
		// ClientInterceptorsDeclaration is the exact client interceptor interface.
		ClientInterceptorsDeclaration *codegen.NameDeclaration
		// ClientEndpointWrapperDeclaration is the exact wrapper applied to one
		// client endpoint. It is nil for service-level command data.
		ClientEndpointWrapperDeclaration *codegen.NameDeclaration
	}

	// FlagData contains the data needed to render a command-line flag.
	FlagData struct {
		// Name is the name of the flag, e.g. "list-vintage"
		Name string
		// VarName is the name of the flag variable, e.g. "listVintage"
		VarName string
		// Type is the type of the flag, e.g. INT
		Type string
		// FullName is the flag full name e.g. "storageAddVintage"
		FullName string
		// PointerVar is the exact local variable that points to the parsed flag value.
		PointerVar string
		// HasDefault is true when the design supplies a default, including false,
		// zero, or an empty string.
		HasDefault bool
		// TracksPresence is true when generated code must distinguish an omitted
		// flag from a flag whose text is empty.
		TracksPresence bool
		// DefaultValue is the text passed to the standard flag package when the
		// design supplies a default.
		DefaultValue string
		// Description is the flag help text.
		Description string
		// Required is true if the flag is required.
		Required bool
		// Example returns a JSON serialized example value.
		Example string
		// Default returns the default value if any.
		Default any
		// value describes how the flag text becomes a Go value.
		value *flagValuePlan
	}

	// BuildFunctionData contains the data needed to generate a constructor
	// function that builds a service method payload type from the command-line
	// flags.
	BuildFunctionData struct {
		// Name is the build payload function name.
		Name string
		// Description describes the payload function.
		Description string
		// ActualParams is the list of passed build function parameters.
		ActualParams []string
		// FormalParams is the list of build function formal parameter
		// names.
		FormalParams []string
		// FormalParamTypes lists the exact parameter types selected for generated
		// builders. It is empty for plugins that use the released string contract.
		FormalParamTypes []string
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// ResultType is the fully qualified payload type name.
		ResultType string
		// Fields describes the payload fields.
		Fields []*FieldData
		// PayloadInit contains the data needed to render the function
		// body.
		PayloadInit *PayloadInitData
		// CheckErr is true if the payload initialization code requires an
		// "err error" variable that must be checked.
		CheckErr bool
	}

	// ParserDeclarations contains every package function written to one command
	// parser file.
	ParserDeclarations struct {
		// ParseEndpoint is the function that selects and builds an endpoint call.
		ParseEndpoint *codegen.NameDeclaration
		// UsageCommands is the function that lists available commands.
		UsageCommands *codegen.NameDeclaration
		// UsageExamples is the function that prints example commands.
		UsageExamples *codegen.NameDeclaration
		// PresenceFlagType is the private type that records whether the user set a
		// command-line flag.
		PresenceFlagType *codegen.NameDeclaration
	}

	// FlagArgData describes a payload initialization argument from which a
	// command-line flag and the code that loads the flag value into the
	// corresponding payload builder field are generated.
	FlagArgData struct {
		// Name is the argument variable name used to derive the flag name and
		// the name of the local variable holding the flag value.
		Name string
		// TypeName is the argument Go type name.
		TypeName string
		// Plan contains the conversion and validation selected by Goa's transport
		// generators. Plugins may continue to use TypeName and Validate.
		Plan *FlagPlan
		// TypeRef is the reference to the argument type.
		TypeRef string
		// Pointer reports whether JSON flag decoding must preserve a nil value.
		Pointer bool
		// FieldName is the name of the payload field initialized with the
		// argument value if any.
		FieldName string
		// Description is the flag help text.
		Description string
		// Required is true if the flag is required.
		Required bool
		// Example is an example value for the flag.
		Example any
		// DefaultValue is the default value of the argument if any.
		DefaultValue any
		// Validate contains validation code kept for plugins built against the
		// released CLI data. It cannot be used with Plan.
		//
		// Deprecated: Goa transport generators use Plan so checks receive the
		// exact parsed value name. Plugins may continue to use Validate.
		Validate string
		// OmitField if true generates the flag without a corresponding payload
		// builder field.
		OmitField bool
	}

	// FlagPlan contains the conversion and validation choices made by Goa's
	// transport generators.
	FlagPlan struct {
		value      *flagValuePlan
		validation func(string) string
	}

	// flagValuePlan records how command-line text becomes one generated Go value.
	flagValuePlan struct {
		kind            expr.Kind
		typeName        string
		typeRef         string
		alias           bool
		protobufMessage bool
	}

	// FieldData contains the data needed to generate the code that initializes a
	// field in the method payload type.
	FieldData struct {
		// Name is the field name, e.g. "Vintage"
		Name string
		// VarName is the name of the local variable holding the field
		// value, e.g. "vintage"
		VarName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Init is the code initializing the variable.
		Init string
	}

	// PayloadInitData contains the data needed to generate a constructor
	// function that initializes a service method payload type from the
	// command-line arguments.
	PayloadInitData struct {
		// Code is the payload initialization code.
		Code string
		// ReturnTypeAttribute names the payload field initialized from the body.
		ReturnTypeAttribute string
		// ReturnTypeAttributePointer is true if the return type attribute
		// generated struct field holds a pointer.
		ReturnTypeAttributePointer bool
		// ReturnTypeAttributeUnion reports whether the selected payload field is
		// a union value.
		ReturnTypeAttributeUnion bool
		// OptionalBody reports whether the selected body flag may be omitted.
		OptionalBody bool
		// ReturnIsStruct if true indicates that the method payload is an object.
		ReturnIsStruct bool
		// ReturnTypeName is the fully-qualified name of the payload.
		ReturnTypeName string
		// ReturnTypePkg is the package name where the payload is present.
		ReturnTypePkg string
		// Args is the list of arguments for the constructor.
		Args []*codegen.InitArgData
	}

	// conversionData describes the Go value produced from command-line flag text.
	conversionData struct {
		code          string
		value         string
		declaresError bool
		canError      bool
	}

	// conversionVariableNames contains local names used while parsing one flag.
	conversionVariableNames struct {
		error     string
		parsed    string
		converted string
	}

	// parserFlagsData gives the shared flag template its commands and the fixed
	// local names chosen for the surrounding endpoint parser.
	parserFlagsData struct {
		Commands         []*CommandData
		Variables        *ParserVariablesData
		PresenceFlagType string
	}
)

// BuildCommandData builds the data needed by CLI code generators to render the
// parsing of the service command.
func BuildCommandData(data *service.Data) *CommandData {
	description := data.Description
	if description == "" {
		description = fmt.Sprintf("Make requests to the %q service", data.Name)
	}

	var interceptors *InterceptorData
	if len(data.ClientInterceptors) > 0 {
		interceptors = &InterceptorData{
			VarName:                       codegen.Goify(data.Name, false) + "Inter",
			PkgName:                       data.PkgName,
			ClientInterceptorsDeclaration: data.ClientInterceptorsDeclaration,
		}
	}

	return &CommandData{
		ServiceName:  data.Name,
		Name:         codegen.KebabCase(data.PathName),
		VarName:      codegen.Goify(data.Name, false),
		Description:  description,
		PkgName:      data.PkgName + "c",
		Interceptors: interceptors,
	}
}

// BuildSubcommandData builds the data needed by CLI code generators to render
// the CLI parsing of the service sub-command.
func BuildSubcommandData(data *service.Data, m *service.MethodData, buildFunction *BuildFunctionData, flags []*FlagData) *SubcommandData {
	en := m.Name
	name := codegen.KebabCase(en)
	fullName := goifyTerms(data.Name, en)
	description := m.Description
	if description == "" {
		description = fmt.Sprintf("Make request to the %q endpoint", m.Name)
	}

	var conversionFlag *FlagData
	if m.Payload != "" && buildFunction == nil && len(flags) > 0 {
		conversionFlag = flags[0]
	}

	var interceptors *InterceptorData
	if len(m.ClientInterceptors) > 0 {
		interceptors = &InterceptorData{
			VarName:                          codegen.Goify(data.Name, false) + "Inter",
			PkgName:                          data.PkgName,
			ClientInterceptorsDeclaration:    data.ClientInterceptorsDeclaration,
			ClientEndpointWrapperDeclaration: m.ClientEndpointWrapperDeclaration,
		}
	}
	sub := &SubcommandData{
		MethodName:     m.Name,
		Name:           name,
		FullName:       fullName,
		Description:    description,
		Flags:          flags,
		MethodVarName:  m.VarName,
		BuildFunction:  buildFunction,
		conversionFlag: conversionFlag,
		Interceptors:   interceptors,
	}
	generateExample(sub, data.Name)

	return sub
}

// EndpointParserFile returns the file that implements the command line parser
// that builds the client endpoint and payload necessary to perform a request.
// The parse section renders the transport-specific ParseEndpoint function.
func EndpointParserFile(
	path, title string,
	specs []*codegen.ImportSpec,
	data []*CommandData,
	parseSection *codegen.SectionTemplate,
) *codegen.File {
	return endpointParserFile(path, title, specs, data, parseSection, "", releasedUsageCommandsName, releasedUsageExamplesName)
}

// EndpointParserFile returns a parser file that uses the function names chosen
// for this parser plan.
func (p *ParserPlan) EndpointParserFile(
	path, title string,
	specs []*codegen.ImportSpec,
	data []*CommandData,
	parseSection *codegen.SectionTemplate,
) *codegen.File {
	presenceFlagType := ""
	if p.Declarations.PresenceFlagType != nil {
		presenceFlagType = p.Declarations.PresenceFlagType.Name()
	}
	return endpointParserFile(
		path,
		title,
		specs,
		data,
		parseSection,
		presenceFlagType,
		p.Declarations.UsageCommands.Name,
		p.Declarations.UsageExamples.Name,
	)
}

// endpointParserFile assembles one parser file with the supplied help function
// names.
func endpointParserFile(
	path, title string,
	specs []*codegen.ImportSpec,
	data []*CommandData,
	parseSection *codegen.SectionTemplate,
	presenceFlagType string,
	usageCommandsName, usageExamplesName func() string,
) *codegen.File {
	sections := make([]*codegen.SectionTemplate, 0, 4+len(data))
	sections = append(sections, codegen.Header(title, "cli", specs))
	if presenceFlagType != "" && hasPresenceFlags(data) {
		var declaration bytes.Buffer
		if err := presenceFlagSection(presenceFlagType).Write(&declaration); err != nil {
			panic(err)
		}
		plannedParse := *parseSection
		plannedParse.Source = declaration.String() + "\n" + parseSection.Source
		parseSection = &plannedParse
	}
	sections = append(sections,
		usageCommands(data, usageCommandsName),
		usageExamples(data, usageExamplesName),
		parseSection,
	)
	for _, cmd := range data {
		sections = append(sections, CommandUsage(cmd))
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// MakeFlags returns the flag data generated from the given payload
// initialization arguments along with the data for the function that builds
// the method payload from the corresponding flag values. payload and
// payloadRef describe the method payload type, pinit - if not nil - describes
// the payload constructor invoked by the build function.
func MakeFlags(
	svcn string,
	m *service.MethodData,
	args []*FlagArgData,
	payload expr.DataType,
	payloadRef string,
	pinit *PayloadInitData,
) ([]*FlagData, *BuildFunctionData) {
	var (
		fdata      = make([]*FieldData, 0, len(args)) // preallocate
		flags      = make([]*FlagData, len(args))
		params     = make([]string, len(args))
		paramTypes = make([]string, len(args))
		planned    = true
		check      bool
	)
	for i, arg := range args {
		value := (*flagValuePlan)(nil)
		validation := func(string) string {
			return arg.Validate
		}
		if arg.Plan == nil {
			planned = false
			value = legacyFlagValuePlan(arg.TypeName)
		} else {
			if arg.Validate != "" {
				panic("CLI flag validation cannot use both Validate and Plan")
			}
			value = arg.Plan.value
			validation = arg.Plan.validation
		}
		f := newFlagData(svcn, m.Name, arg.Name, value, arg.Description, arg.Required, arg.Example, arg.DefaultValue)
		f.TracksPresence = arg.Plan != nil && !f.HasDefault
		flags[i] = f
		params[i] = f.FullName
		paramTypes[i] = "string"
		if f.TracksPresence {
			paramTypes[i] = "*string"
		}
		if arg.OmitField {
			continue
		}
		code, chek := fieldLoadCode(f, arg.Name, value, validation, arg.DefaultValue, payload, payloadRef)
		check = check || chek
		tn := arg.TypeRef
		if value.isJSON() {
			// JSON decoding uses the planned transport value. Optional
			// pointer-shaped bodies keep a pointer so omission remains nil.
			tn = value.typeName
			if arg.Pointer {
				tn = "*" + tn
			}
		}
		fdata = append(fdata, &FieldData{
			Name:    arg.Name,
			VarName: arg.Name,
			TypeRef: tn,
			Init:    code,
		})
	}
	if !planned {
		usesPresence := false
		for _, flag := range flags {
			usesPresence = usesPresence || flag.TracksPresence
		}
		if !usesPresence {
			paramTypes = nil
		}
	}

	return flags, &BuildFunctionData{
		Name:             "Build" + m.VarName + "Payload",
		ActualParams:     params,
		FormalParams:     params,
		FormalParamTypes: paramTypes,
		ServiceName:      svcn,
		MethodName:       m.Name,
		ResultType:       payloadRef,
		Fields:           fdata,
		PayloadInit:      pinit,
		CheckErr:         check,
	}
}

// PayloadBuildersFile returns the file that contains the payload constructors
// that use the command flag values as arguments.
func PayloadBuildersFile(path, title string, specs []*codegen.ImportSpec, data *CommandData) *codegen.File {
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", specs),
	}
	for _, sub := range data.Subcommands {
		if sub.BuildFunction != nil {
			sections = append(sections, PayloadBuilderSection(sub.BuildFunction))
		}
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// UsageCommands builds a section template that generates a help text showing
// the list of allowed commands and sub-commands.
func UsageCommands(data []*CommandData) *codegen.SectionTemplate {
	return usageCommands(data, releasedUsageCommandsName)
}

// usageCommands renders command help with the function name chosen for its
// generated package.
func usageCommands(data []*CommandData, name func() string) *codegen.SectionTemplate {
	usages := make([]string, len(data))
	for i, cmd := range data {
		subs := make([]string, len(cmd.Subcommands))
		for i, s := range cmd.Subcommands {
			subs[i] = s.Name
		}
		var lp, rp string
		if len(subs) > 1 {
			lp = "("
			rp = ")"
		}
		usages[i] = fmt.Sprintf("%s %s%s%s", cmd.Name, lp, strings.Join(subs, "|"), rp)
	}

	return &codegen.SectionTemplate{
		Source: cliTemplates.Read(usageCommandsT),
		Data:   usages,
		FuncMap: map[string]any{
			"usageName": name,
		},
	}
}

// UsageExamples builds a section template that generates a help text showing
// a valid invocation of the CLI tool.
func UsageExamples(data []*CommandData) *codegen.SectionTemplate {
	return usageExamples(data, releasedUsageExamplesName)
}

// usageExamples renders example help with the function name chosen for its
// generated package.
func usageExamples(data []*CommandData, name func() string) *codegen.SectionTemplate {
	var examples []string
	for i, cmd := range data {
		if i < 5 {
			examples = append(examples, cmd.Example)
		}
	}

	return &codegen.SectionTemplate{
		Source: cliTemplates.Read(usageExamplesT),
		Data:   examples,
		FuncMap: map[string]any{
			"usageName": name,
		},
	}
}

// releasedUsageCommandsName returns the help function name used by the
// released parser helper.
func releasedUsageCommandsName() string {
	return "UsageCommands"
}

// releasedUsageExamplesName returns the example function name used by the
// released parser helper.
func releasedUsageExamplesName() string {
	return "UsageExamples"
}

// FlagsCode returns a string containing the code that parses the command-line
// flags to infer the command (service), sub-command (method), and the
// arguments (method payload) invoked by the tool. It panics if any error
// occurs during the generation of flag parsing code.
func FlagsCode(data []*CommandData) string {
	section := codegen.SectionTemplate{
		Name:    "parse-endpoint-flags",
		Source:  cliTemplates.Read(parseFlagsT),
		Data:    data,
		FuncMap: map[string]any{"printDescription": printDescription},
	}
	var flagsCode bytes.Buffer
	err := section.Write(&flagsCode)
	if err != nil {
		panic(err)
	}

	return flagsCode.String()
}

// FlagsCode renders flag parsing with the exact local names chosen by this
// parser plan.
func (p *ParserPlan) FlagsCode(data []*CommandData) string {
	if p.Variables == nil {
		panic("CLI parser variables must be planned before rendering flags")
	}
	presenceFlagType := ""
	usesPresence := hasPresenceFlags(data)
	declaredPresence := p.Declarations.PresenceFlagType != nil
	if usesPresence != declaredPresence {
		panic("CLI flag presence declaration does not match the generated flags")
	}
	if usesPresence {
		presenceFlagType = p.Declarations.PresenceFlagType.Name()
	}
	section := codegen.SectionTemplate{
		Name:   "parse-endpoint-flags",
		Source: cliTemplates.Read(parseFlagsPlannedT),
		Data: &parserFlagsData{
			Commands:         data,
			Variables:        p.Variables,
			PresenceFlagType: presenceFlagType,
		},
		FuncMap: map[string]any{"printDescription": printDescription},
	}
	var flagsCode bytes.Buffer
	if err := section.Write(&flagsCode); err != nil {
		panic(err)
	}
	return flagsCode.String()
}

// CommandUsage builds the section templates that can be used to generate the
// endpoint command usage code.
func CommandUsage(data *CommandData) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:    "cli-command-usage",
		Source:  cliTemplates.Read(commandUsageT),
		Data:    data,
		FuncMap: map[string]any{"printDescription": printDescription},
	}
}

// PayloadBuilderSection builds the section template that can be used to
// generate the payload builder code.
func PayloadBuilderSection(buildFunction *BuildFunctionData) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:   "cli-build-payload",
		Source: cliTemplates.Read(buildPayloadT),
		Data:   buildFunction,
		FuncMap: map[string]any{
			"fieldCode":       fieldCode,
			"formalParamType": formalParamType,
		},
	}
}

// NewFlagPlan records the conversion and validation selected for one command-
// line flag. typeName is the concrete local type used for JSON values. typeRef
// is the concrete non-pointer reference used for primitive casts.
func NewFlagPlan(attribute *expr.AttributeExpr, typeName, typeRef string, validation func(string) string) *FlagPlan {
	kind := expr.AnyKind
	alias := expr.IsAlias(attribute.Type)
	if custom, _ := codegen.GetMetaType(attribute); custom == "" && expr.IsPrimitive(attribute.Type) {
		dataType := attribute.Type
		for {
			userType, ok := dataType.(expr.UserType)
			if !ok {
				break
			}
			dataType = userType.Attribute().Type
		}
		kind = dataType.Kind()
	}
	return &FlagPlan{
		value: &flagValuePlan{
			kind:     kind,
			typeName: typeName,
			typeRef:  typeRef,
			alias:    alias,
		},
		validation: validation,
	}
}

// NewProtobufFlagPlan records a command-line flag whose JSON value is a
// protobuf message. The generated code uses protobuf's JSON decoder so the
// accepted field names and values match the message contract.
func NewProtobufFlagPlan(attribute *expr.AttributeExpr, typeName string) *FlagPlan {
	plan := NewFlagPlan(attribute, typeName, typeName, nil)
	plan.value.protobufMessage = true
	return plan
}

// NewFlagData creates flag data from the released string type description.
//
// svcn is the service name
// en is the endpoint name
// name is the flag name
// typeName is the flag type
// description is the flag description
// required determines if the flag is required
// example is an example value for the flag
func NewFlagData(svcn, en, name, typeName, description string, required bool, example, def any) *FlagData {
	return newFlagData(svcn, en, name, legacyFlagValuePlan(typeName), description, required, example, def)
}

// NewFlagDataForPlan creates flag data from the given conversion and
// validation choices.
func NewFlagDataForPlan(svcn, en, name string, plan *FlagPlan, description string, required bool, example, def any) *FlagData {
	flag := newFlagData(svcn, en, name, plan.value, description, required, example, def)
	flag.TracksPresence = !flag.HasDefault
	return flag
}

// FieldLoadCode returns the code used in the build payload function that
// initializes one of the payload object fields. It returns the initialization
// code and a boolean indicating whether the code requires an "err" variable.
func FieldLoadCode(f *FlagData, argName, argTypeName, validate string, defaultValue any, payload expr.DataType, payloadRef string) (string, bool) {
	var validation func(string) string
	if validate != "" {
		validation = func(string) string {
			return validate
		}
	}
	return fieldLoadCode(f, argName, legacyFlagValuePlan(argTypeName), validation, defaultValue, payload, payloadRef)
}

// newFlagData creates flag data from the conversion selected during planning.
func newFlagData(svcn, en, name string, value *flagValuePlan, description string, required bool, example, def any) *FlagData {
	ex := jsonExample(example)
	fn := goifyTerms(svcn, en, name)
	hasDefault := def != nil
	defaultValue := ""
	if hasDefault {
		if value.isJSON() {
			encoded, err := json.Marshal(def)
			if err != nil {
				panic(fmt.Sprintf("cannot encode the default for %s.%s.%s: %s", svcn, en, name, err)) // bug
			}
			defaultValue = string(encoded)
		} else {
			defaultValue = fmt.Sprint(def)
		}
	}
	return &FlagData{
		Name:         codegen.KebabCase(name),
		VarName:      codegen.Goify(name, false),
		Type:         value.flagType(),
		FullName:     fn,
		Description:  description,
		Required:     required,
		Example:      ex,
		Default:      def,
		HasDefault:   hasDefault,
		DefaultValue: defaultValue,
		value:        value,
	}
}

// legacyFlagValuePlan reproduces the released string-based flag conversion.
func legacyFlagValuePlan(typeName string) *flagValuePlan {
	kind := expr.AnyKind
	switch typeName {
	case codegen.GoNativeTypeName(expr.Boolean):
		kind = expr.BooleanKind
	case codegen.GoNativeTypeName(expr.Int):
		kind = expr.IntKind
	case codegen.GoNativeTypeName(expr.Int32):
		kind = expr.Int32Kind
	case codegen.GoNativeTypeName(expr.Int64):
		kind = expr.Int64Kind
	case codegen.GoNativeTypeName(expr.UInt):
		kind = expr.UIntKind
	case codegen.GoNativeTypeName(expr.UInt32):
		kind = expr.UInt32Kind
	case codegen.GoNativeTypeName(expr.UInt64):
		kind = expr.UInt64Kind
	case codegen.GoNativeTypeName(expr.Float32):
		kind = expr.Float32Kind
	case codegen.GoNativeTypeName(expr.Float64):
		kind = expr.Float64Kind
	case codegen.GoNativeTypeName(expr.String):
		kind = expr.StringKind
	case codegen.GoNativeTypeName(expr.Bytes):
		kind = expr.BytesKind
	}
	return &flagValuePlan{
		kind:     kind,
		typeName: typeName,
		typeRef:  typeName,
	}
}

// fieldLoadCode writes a field conversion from its complete generation plan.
func fieldLoadCode(
	f *FlagData,
	argName string,
	value *flagValuePlan,
	validation func(string) string,
	defaultValue any,
	payload expr.DataType,
	payloadRef string,
) (string, bool) {
	var (
		code             string
		validationTarget string
		declErr          bool
		startIf          string
		endIf            string
	)
	from := f.FullName
	if f.TracksPresence {
		from = "*" + f.FullName
		if f.Required {
			zero := "nil"
			declareZero := ""
			if expr.IsPrimitive(payload) {
				zero = "zero"
				declareZero = fmt.Sprintf("\tvar zero %s\n", payloadRef)
			}
			startIf = fmt.Sprintf(
				"if %s == nil {\n%s\treturn %s, fmt.Errorf(\"missing required flag --%s\")\n}\n",
				f.FullName,
				declareZero,
				zero,
				f.Name,
			)
		} else {
			startIf = fmt.Sprintf("if %s != nil {\n", f.FullName)
			endIf = "\n}"
		}
	} else if !f.Required && !f.HasDefault {
		startIf = fmt.Sprintf("if %s != \"\" {\n", f.FullName)
		endIf = "\n}"
	}
	pointer := value.kind != expr.BytesKind && !value.isJSON() && !f.Required && defaultValue == nil
	conversion := conversionData{}
	if f.TracksPresence && pointer && value.kind == expr.StringKind && !value.alias {
		conversion.code = fmt.Sprintf("%s = %s", argName, f.FullName)
		conversion.value = from
	} else {
		conversion = conversionCode(from, argName, value, pointer, conversionVariableNames{
			error:     "err",
			parsed:    "v",
			converted: "val",
		})
	}
	code = conversion.code
	validationTarget = conversion.value
	declErr = conversion.declaresError
	if conversion.canError {
		code += "\nif err != nil {\n"
		nilVal := "nil"
		if expr.IsPrimitive(payload) {
			code += fmt.Sprintf("var zero %s\n", payloadRef)
			nilVal = "zero"
		}
		if value.isJSON() {
			code += fmt.Sprintf(`return %s, fmt.Errorf("invalid JSON for %s, \nerror: %%s, \nexample of valid JSON:\n%%s", err, %q)`,
				nilVal, argName, f.Example)
		} else {
			code += fmt.Sprintf(`return %s, fmt.Errorf("invalid value for %s, must be %s")`,
				nilVal, argName, f.Type)
		}
		code += "\n}"
	}
	if validation != nil {
		validate := validation(validationTarget)
		if validate != "" {
			declErr = true
			code += "\n" + validate + "\n"
			nilVal := "nil"
			if expr.IsPrimitive(payload) {
				code += fmt.Sprintf("var zero %s\n", payloadRef)
				nilVal = "zero"
			}
			code += fmt.Sprintf("if err != nil {\n\treturn %s, err\n}", nilVal)
		}
	}
	return fmt.Sprintf("%s%s%s", startIf, code, endIf), declErr
}

// jsonExample turns a generated value into the JSON text shown in CLI help and
// invalid-value errors.
func jsonExample(v any) string {
	// In JSON, keys must be a string. But goa allows map keys to be anything.
	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Map {
		keys := r.MapKeys()
		if keys[0].Kind() != reflect.String {
			a := make(map[string]any, len(keys))
			var kstr string
			for _, k := range keys {
				switch k.Kind() {
				case reflect.Bool:
					kstr = strconv.FormatBool(k.Bool())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					kstr = strconv.FormatInt(k.Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					kstr = strconv.FormatUint(k.Uint(), 10)
				case reflect.Float32, reflect.Float64:
					kstr = strconv.FormatFloat(k.Float(), 'f', -1, k.Type().Bits())
				default:
					panic(fmt.Sprintf("unsupported CLI example map key kind %s", k.Kind()))
				}
				a[kstr] = r.MapIndex(k).Interface()
			}
			v = a
		}
	}
	b, err := json.MarshalIndent(v, "   ", "   ")
	ex := "?"
	if err == nil {
		ex = string(b)
	}
	if strings.Contains(ex, "\n") {
		ex = "'" + strings.ReplaceAll(ex, "'", "\\'") + "'"
	}
	return ex
}

// directPayloadConversion writes the primitive payload conversion after the
// parser has chosen the exact flag pointer variable used by this method.
func directPayloadConversion(flag *FlagData, variables *ParserVariablesData) string {
	var prefix, suffix string
	target := variables.Data
	if flag.value.isJSON() {
		target = variables.ConvertedValue
		prefix = fmt.Sprintf("var %s %s\n", variables.ConvertedValue, flag.value.typeName)
		suffix = fmt.Sprintf("\n%s = %s", variables.Data, variables.ConvertedValue)
	}
	from := "*" + flag.PointerVar
	presenceCheck := ""
	if flag.TracksPresence {
		from += ".value"
		if flag.Required {
			presenceCheck = fmt.Sprintf(
				"if %s.value == nil {\n\treturn nil, nil, fmt.Errorf(\"missing required flag --%s\")\n}\n",
				flag.PointerVar,
				flag.Name,
			)
		}
	}
	converted := conversionCode(from, target, flag.value, false, conversionVariableNames{
		error:     variables.Error,
		parsed:    variables.ParsedValue,
		converted: variables.ConvertedValue,
	})
	code := presenceCheck + prefix + converted.code + suffix
	if !converted.canError {
		return code
	}
	code = fmt.Sprintf("var %s error\n%s\nif %s != nil {\n", variables.Error, code, variables.Error)
	if flag.value.isJSON() {
		code += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid JSON for %s, \nerror: %%s, \nexample of valid JSON:\n%%s", %s, %q)`,
			flag.PointerVar, variables.Error, flag.Example)
	} else {
		code += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid value for %s, must be %s")`,
			flag.PointerVar, flag.Type)
	}
	return code + "\n}"
}

// conversionCode describes the code and concrete Go value produced from the
// flag text in from. The result also reports how the conversion uses err.
func conversionCode(from, to string, value *flagValuePlan, pointer bool, variables conversionVariableNames) conversionData {
	var parse string
	switch value.kind {
	case expr.BooleanKind:
		if !value.alias {
			return directParsedConversion(to, value.typeRef, fmt.Sprintf("strconv.ParseBool(%s)", from), pointer, variables)
		}
		parse = fmt.Sprintf("var %s bool\n%s, %s = strconv.ParseBool(%s)", variables.parsed, variables.parsed, variables.error, from)
	case expr.IntKind, expr.Int32Kind, expr.Int64Kind:
		bits := "64"
		if value.kind == expr.IntKind {
			bits = "strconv.IntSize"
		} else if value.kind == expr.Int32Kind {
			bits = "32"
		}
		if value.kind == expr.Int64Kind && !value.alias {
			return directParsedConversion(to, value.typeRef, fmt.Sprintf("strconv.ParseInt(%s, 10, 64)", from), pointer, variables)
		}
		parse = fmt.Sprintf("var %s int64\n%s, %s = strconv.ParseInt(%s, 10, %s)", variables.parsed, variables.parsed, variables.error, from, bits)
	case expr.UIntKind, expr.UInt32Kind, expr.UInt64Kind:
		bits := "64"
		if value.kind == expr.UIntKind {
			bits = "strconv.IntSize"
		} else if value.kind == expr.UInt32Kind {
			bits = "32"
		}
		if value.kind == expr.UInt64Kind && !value.alias {
			return directParsedConversion(to, value.typeRef, fmt.Sprintf("strconv.ParseUint(%s, 10, 64)", from), pointer, variables)
		}
		parse = fmt.Sprintf("var %s uint64\n%s, %s = strconv.ParseUint(%s, 10, %s)", variables.parsed, variables.parsed, variables.error, from, bits)
	case expr.Float32Kind, expr.Float64Kind:
		bits := "64"
		if value.kind == expr.Float32Kind {
			bits = "32"
		}
		if value.kind == expr.Float64Kind && !value.alias {
			return directParsedConversion(to, value.typeRef, fmt.Sprintf("strconv.ParseFloat(%s, 64)", from), pointer, variables)
		}
		parse = fmt.Sprintf("var %s float64\n%s, %s = strconv.ParseFloat(%s, %s)", variables.parsed, variables.parsed, variables.error, from, bits)
	case expr.StringKind:
		converted := from
		if value.alias {
			converted = fmt.Sprintf("%s(%s)", value.typeRef, from)
		}
		if pointer && !value.alias {
			return conversionData{code: fmt.Sprintf("%s = &%s", to, from), value: from}
		}
		code, target := assignConvertedValue(to, converted, pointer, variables.converted)
		return conversionData{code: code, value: target}
	case expr.BytesKind:
		converted := fmt.Sprintf("[]byte(%s)", from)
		if value.alias {
			converted = fmt.Sprintf("%s(%s)", value.typeRef, from)
		}
		return conversionData{code: fmt.Sprintf("%s = %s", to, converted), value: to}
	default:
		if value.protobufMessage {
			parse = fmt.Sprintf("%s = protojson.Unmarshal([]byte(%s), &%s)", variables.error, from, to)
		} else {
			parse = fmt.Sprintf("%s = json.Unmarshal([]byte(%s), &%s)", variables.error, from, to)
		}
		return conversionData{code: parse, value: to, declaresError: true, canError: true}
	}
	converted := fmt.Sprintf("%s(%s)", value.typeRef, variables.parsed)
	assignment, target := assignConvertedValue(to, converted, pointer, variables.converted)
	return conversionData{
		code:          parse + "\n" + assignment,
		value:         target,
		declaresError: true,
		canError:      true,
	}
}

// directParsedConversion writes a parser result directly into its final value
// when the parser already returns the generated Go type.
func directParsedConversion(target, typeRef, parser string, pointer bool, variables conversionVariableNames) conversionData {
	if !pointer {
		return conversionData{
			code:          fmt.Sprintf("%s, %s = %s", target, variables.error, parser),
			value:         target,
			declaresError: true,
			canError:      true,
		}
	}
	return conversionData{
		code:          fmt.Sprintf("var %s %s\n%s, %s = %s\n%s = &%s", variables.converted, typeRef, variables.converted, variables.error, parser, target, variables.converted),
		value:         variables.converted,
		declaresError: true,
		canError:      true,
	}
}

// assignConvertedValue writes a converted scalar into its final local and
// returns the concrete value expression used by validation.
func assignConvertedValue(target, converted string, pointer bool, valueVariable string) (string, string) {
	if !pointer {
		return fmt.Sprintf("%s = %s", target, converted), target
	}
	return fmt.Sprintf("%s := %s\n%s = &%s", valueVariable, converted, target, valueVariable), valueVariable
}

// flagType returns the command-line type shown in help and conversion errors.
func (p *flagValuePlan) flagType() string {
	switch p.kind {
	case expr.BooleanKind:
		return "BOOL"
	case expr.IntKind:
		return "INT"
	case expr.Int32Kind:
		return "INT32"
	case expr.Int64Kind:
		return "INT64"
	case expr.UIntKind:
		return "UINT"
	case expr.UInt32Kind:
		return "UINT32"
	case expr.UInt64Kind:
		return "UINT64"
	case expr.Float32Kind:
		return "FLOAT32"
	case expr.Float64Kind:
		return "FLOAT64"
	case expr.StringKind, expr.BytesKind:
		return "STRING"
	default:
		return "JSON"
	}
}

// isJSON reports whether flag text uses JSON decoding.
func (p *flagValuePlan) isJSON() bool {
	return p.kind != expr.BooleanKind && p.kind != expr.IntKind &&
		p.kind != expr.Int32Kind && p.kind != expr.Int64Kind &&
		p.kind != expr.UIntKind && p.kind != expr.UInt32Kind &&
		p.kind != expr.UInt64Kind && p.kind != expr.Float32Kind &&
		p.kind != expr.Float64Kind && p.kind != expr.StringKind &&
		p.kind != expr.BytesKind
}

// hasPresenceFlags reports whether the parser needs to remember if any flag
// was supplied. Defaults do not need this because they always have a value.
func hasPresenceFlags(data []*CommandData) bool {
	for _, command := range data {
		for _, subcommand := range command.Subcommands {
			for _, flag := range subcommand.Flags {
				if flag.TracksPresence {
					return true
				}
			}
		}
	}
	return false
}

// presenceFlagSection writes the private flag type used to keep omitted flags
// distinct from flags whose text is empty.
func presenceFlagSection(typeName string) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:   "cli-presence-flag",
		Source: cliTemplates.Read(presenceFlagT),
		Data:   typeName,
	}
}

// formalParamType returns the planned builder parameter type. Plugin data made
// with the released API has no type list and continues to use string.
func formalParamType(data *BuildFunctionData, index int) string {
	if len(data.FormalParamTypes) == 0 {
		return "string"
	}
	return data.FormalParamTypes[index]
}

// goifyTerms makes valid go identifiers out of the supplied terms
func goifyTerms(terms ...string) string {
	res := codegen.Goify(terms[0], false)
	if len(terms) == 1 {
		return res
	}
	for _, t := range terms[1:] {
		res += codegen.Goify(t, true)
	}
	return res
}

// printDescription indents each line embedded in generated Go code while
// keeping blank lines free of whitespace.
func printDescription(desc string) string {
	desc = strings.TrimRight(desc, " \t\r\n")
	lines := strings.Split(strings.ReplaceAll(desc, "`", "`+\"`\"+`"), "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			lines[i] = ""
			continue
		}
		lines[i] = "\t" + lines[i]
	}
	return strings.Join(lines, "\n")
}

func generateExample(sub *SubcommandData, svc string) {
	ex := codegen.KebabCase(svc) + " " + codegen.KebabCase(sub.Name)
	for _, f := range sub.Flags {
		ex += " --" + f.Name + " " + f.Example
	}
	sub.Example = ex
}

// fieldCode generates code to initialize the data structures fields
// from the given args. It is used only in templates.
func fieldCode(init *PayloadInitData) string {
	varn := "res"
	if init.ReturnTypeAttribute == "" {
		varn = "v"
	}
	// We can ignore the transform helpers as there won't be any generated
	// because the args cannot be user types.
	c, _, err := codegen.InitStructFields(init.Args, varn, "", init.ReturnTypePkg)
	if err != nil {
		panic(err) // bug
	}
	return c
}
