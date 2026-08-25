// This file assigns the function and local variable names used by command-line
// client files. HTTP and gRPC planning completes these names before rendering
// the shared CLI templates.
package cli

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"sort"

	"goa.design/goa/v3/codegen"
)

type (
	// CommandDeclarationInput names one service command and each method command
	// written to a parser file.
	CommandDeclarationInput struct {
		// Service is the design service name used to identify this command.
		Service string
		// Methods lists the design methods accepted by this service command.
		Methods []string
		// NeedsFlagPresence is true when any method has a command-line flag
		// without a design default.
		NeedsFlagPresence bool
	}

	// ParserPlan contains the names written to one command parser package.
	ParserPlan struct {
		// Declarations contains the three functions shared by the parser file.
		Declarations *ParserDeclarations
		// Commands contains the help function names for each design service.
		Commands map[string]*CommandPlan
		// Variables contains the exact parameter and local names used by ParseEndpoint.
		Variables     *ParserVariablesData
		family        string
		variables     *codegen.NameScope
		imports       []string
		variableNames map[parserVariableIdentity]parserVariableName
		planned       bool
	}

	// ParserVariablesData contains the exact names of parameters and local
	// values written directly by the shared HTTP and gRPC parser templates.
	ParserVariablesData struct {
		// ServiceName stores the selected service command name.
		ServiceName string
		// ServiceFlags stores the flag set for the selected service.
		ServiceFlags string
		// MethodName stores the selected method command name.
		MethodName string
		// MethodFlags stores the flag set for the selected method.
		MethodFlags string
		// Data stores the payload passed to the selected endpoint.
		Data string
		// Endpoint stores the selected Goa endpoint.
		Endpoint string
		// Error stores an error returned while building the payload.
		Error string
		// Client stores the generated transport client.
		Client string
		// Scheme is the HTTP URL scheme parameter.
		Scheme string
		// Host is the HTTP server address parameter.
		Host string
		// Doer is the HTTP request executor parameter.
		Doer string
		// Encoder is the HTTP request encoder parameter.
		Encoder string
		// Decoder is the HTTP response decoder parameter.
		Decoder string
		// Restore is the HTTP response-body restore parameter.
		Restore string
		// Dialer is the WebSocket dialer parameter.
		Dialer string
		// Connection is the gRPC connection parameter.
		Connection string
		// Options is the gRPC call option parameter.
		Options string
		// ParsedValue stores a primitive value returned by a string parser.
		ParsedValue string
		// ConvertedValue stores a value before it is assigned through a pointer.
		ConvertedValue string
	}

	// ParserLocalData describes one transport-specific parameter written in the
	// generated endpoint parser. PlanVariables fills VarName before rendering.
	ParserLocalData struct {
		// ServiceName is the exact design service that uses this parameter.
		ServiceName string
		// MethodName is the exact design method that uses this parameter. It is
		// empty for a service-wide parameter.
		MethodName string
		// Use distinguishes parameters that serve different purposes in one method.
		Use string
		// PreferredName is the Go name used when it does not conflict with another local.
		PreferredName string
		// VarName is the exact Go name written by the parameter and every use.
		VarName string
	}

	// CommandPlan contains the help function names for one service command.
	CommandPlan struct {
		// Usage is the service help function.
		Usage *codegen.NameDeclaration
		// Methods contains help functions indexed by design method name.
		Methods map[string]*codegen.NameDeclaration
	}

	// symbolOrder identifies one shared CLI function by the design names that
	// select its output file and contents.
	symbolOrder struct {
		family   string
		root     string
		server   string
		service  string
		method   string
		role     symbolRole
		commands [sha256.Size]byte
	}

	// symbolRole lists the package functions emitted by shared CLI templates.
	symbolRole uint8

	// parserVariableCandidate records one local definition and the data field
	// that receives its exact Go name.
	parserVariableCandidate struct {
		identity    parserVariableIdentity
		preferred   string
		command     *CommandData
		subcommand  *SubcommandData
		flag        *FlagData
		interceptor *InterceptorData
		local       *ParserLocalData
	}

	// parserVariableIdentity orders local definitions by their exact design
	// names, so reversing input slices does not change collision suffixes.
	parserVariableIdentity struct {
		service string
		method  string
		flag    string
		use     string
		role    parserVariableRole
	}

	// parserVariableName stores the preferred and exact name selected for one local.
	parserVariableName struct {
		preferred string
		name      string
	}

	// parserVariableRole distinguishes local definitions with the same design names.
	parserVariableRole uint8
)

const (
	parseEndpointRole symbolRole = iota + 1
	usageCommandsRole
	usageExamplesRole
	commandUsageRole
	methodUsageRole
	payloadBuilderRole
	presenceFlagTypeRole
)

const (
	serviceFlagSetVariable parserVariableRole = iota + 1
	methodFlagSetVariable
	flagPointerVariable
	interceptorVariable
	transportVariable
)

// DeclareParser submits every function written to one parser package. family
// is "http", "jsonrpc", or "grpc"; root and server distinguish files from
// separate designs; commands supplies the service and method help names.
func DeclareParser(pkg *codegen.GeneratedPackage, family, root, server string, commands []CommandDeclarationInput) (*ParserPlan, error) {
	commandNames := commandDeclarationNames(commands)
	declare := func(preferred string, role symbolRole, service, method string) (*codegen.NameDeclaration, error) {
		visibility := codegen.ExportedName
		if role == commandUsageRole || role == methodUsageRole {
			visibility = codegen.UnexportedName
		}
		declaration := codegen.NewPreferredName(
			codegen.NameFunction,
			preferred,
			visibility,
			symbolOrder{family: family, root: root, server: server, service: service, method: method, role: role, commands: commandNames},
		)
		if err := pkg.DeclareName(declaration); err != nil {
			return nil, err
		}
		return declaration, nil
	}
	parseEndpoint, err := declare("ParseEndpoint", parseEndpointRole, "", "")
	if err != nil {
		return nil, err
	}
	usageCommands, err := declare("UsageCommands", usageCommandsRole, "", "")
	if err != nil {
		return nil, err
	}
	usageExamples, err := declare("UsageExamples", usageExamplesRole, "", "")
	if err != nil {
		return nil, err
	}
	var presenceFlagType *codegen.NameDeclaration
	for _, command := range commands {
		if command.NeedsFlagPresence {
			presenceFlagType = codegen.NewPreferredName(
				codegen.NameType,
				"cliStringFlag",
				codegen.UnexportedName,
				symbolOrder{family: family, root: root, server: server, role: presenceFlagTypeRole, commands: commandNames},
			)
			if err := pkg.DeclareName(presenceFlagType); err != nil {
				return nil, err
			}
			break
		}
	}
	plan := &ParserPlan{
		Declarations: &ParserDeclarations{
			ParseEndpoint:    parseEndpoint,
			UsageCommands:    usageCommands,
			UsageExamples:    usageExamples,
			PresenceFlagType: presenceFlagType,
		},
		Commands:      make(map[string]*CommandPlan, len(commands)),
		family:        family,
		variables:     codegen.NewNameScope(),
		variableNames: make(map[parserVariableIdentity]parserVariableName),
	}
	for _, command := range commands {
		usage, err := declare(goifyTerms(command.Service)+"Usage", commandUsageRole, command.Service, "")
		if err != nil {
			return nil, err
		}
		commandPlan := &CommandPlan{
			Usage:   usage,
			Methods: make(map[string]*codegen.NameDeclaration, len(command.Methods)),
		}
		for _, method := range command.Methods {
			methodUsage, err := declare(goifyTerms(command.Service, method)+"Usage", methodUsageRole, command.Service, method)
			if err != nil {
				return nil, err
			}
			commandPlan.Methods[method] = methodUsage
		}
		plan.Commands[command.Service] = commandPlan
	}
	return plan, nil
}

// PlanVariables chooses every local Go name written by one endpoint parser.
// data contains the shared service, method, flag, and interceptor values;
// locals contains transport-specific parameters such as multipart encoders.
func (p *ParserPlan) PlanVariables(data []*CommandData, locals []*ParserLocalData) {
	candidates := parserVariableCandidates(data, locals)
	sort.Slice(candidates, func(i, j int) bool {
		return compareParserVariable(candidates[i], candidates[j]) < 0
	})
	if !p.planned {
		p.imports = parserImportQualifiers(p.family, data)
		for _, qualifier := range p.imports {
			p.variables.Unique(qualifier)
		}
		p.Variables = planParserVariables(p.variables, p.family)
		for _, candidate := range candidates {
			if _, exists := p.variableNames[candidate.identity]; exists {
				panic("CLI parser contains the same local variable more than once")
			}
			name := p.variables.Unique(candidate.preferred)
			p.variableNames[candidate.identity] = parserVariableName{
				preferred: candidate.preferred,
				name:      name,
			}
			candidate.assign(name)
		}
		p.variables.Freeze()
		p.planned = true
	} else {
		if !slices.Equal(p.imports, parserImportQualifiers(p.family, data)) {
			panic("CLI parser imports changed after local variables were planned")
		}
		if len(candidates) != len(p.variableNames) {
			panic("CLI parser local variables changed after planning")
		}
		for _, candidate := range candidates {
			planned, exists := p.variableNames[candidate.identity]
			if !exists || planned.preferred != candidate.preferred {
				panic("CLI parser local variable changed after planning")
			}
			candidate.assign(planned.name)
		}
	}
	for _, command := range data {
		for _, subcommand := range command.Subcommands {
			if subcommand.Interceptors != nil {
				subcommand.Interceptors.ParserVar = command.Interceptors.ParserVar
			}
			if subcommand.BuildFunction != nil {
				count := len(subcommand.BuildFunction.ActualParams)
				subcommand.ActualPointerVars = make([]string, count)
				subcommand.ActualArgs = make([]string, count)
				for index := range count {
					flag := subcommand.Flags[index]
					subcommand.ActualPointerVars[index] = flag.PointerVar
					subcommand.ActualArgs[index] = "*" + flag.PointerVar
					if flag.TracksPresence {
						subcommand.ActualArgs[index] = flag.PointerVar + ".value"
					}
				}
			}
			if subcommand.conversionFlag != nil {
				subcommand.Conversion = directPayloadConversion(subcommand.conversionFlag, p.Variables)
			}
		}
	}
}

// DeclarePayloadBuilder submits the function that builds one method payload
// from command-line flags and returns the record used by its definition and
// calls.
func DeclarePayloadBuilder(pkg *codegen.GeneratedPackage, family, root, service, method, preferred string) (*codegen.NameDeclaration, error) {
	declaration := codegen.NewPreferredName(
		codegen.NameFunction,
		preferred,
		codegen.ExportedName,
		symbolOrder{family: family, root: root, service: service, method: method, role: payloadBuilderRole},
	)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	return declaration, nil
}

// ComparePackageName orders CLI functions by the design and output file that
// writes them, so reversing input designs does not change their final names.
func (order symbolOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(symbolOrder)
	for _, compared := range []int{
		cmp.Compare(order.family, right.family),
		cmp.Compare(order.root, right.root),
		cmp.Compare(order.server, right.server),
		cmp.Compare(order.service, right.service),
		cmp.Compare(order.method, right.method),
		cmp.Compare(order.role, right.role),
	} {
		if compared != 0 {
			return compared
		}
	}
	return bytes.Compare(order.commands[:], right.commands[:])
}

// parserVariableCandidates collects each local definition before any name is
// chosen, including transport parameters supplied by the caller.
func parserVariableCandidates(data []*CommandData, locals []*ParserLocalData) []*parserVariableCandidate {
	var candidates []*parserVariableCandidate
	for _, command := range data {
		candidates = append(candidates, &parserVariableCandidate{
			identity: parserVariableIdentity{
				service: command.ServiceName,
				role:    serviceFlagSetVariable,
			},
			preferred: command.VarName + "Flags",
			command:   command,
		})
		if command.Interceptors != nil {
			candidates = append(candidates, &parserVariableCandidate{
				identity: parserVariableIdentity{
					service: command.ServiceName,
					role:    interceptorVariable,
				},
				preferred:   command.Interceptors.VarName,
				interceptor: command.Interceptors,
			})
		}
		for _, subcommand := range command.Subcommands {
			candidates = append(candidates, &parserVariableCandidate{
				identity: parserVariableIdentity{
					service: command.ServiceName,
					method:  subcommand.MethodName,
					role:    methodFlagSetVariable,
				},
				preferred:  subcommand.FullName + "Flags",
				subcommand: subcommand,
			})
			for _, flag := range subcommand.Flags {
				candidates = append(candidates, &parserVariableCandidate{
					identity: parserVariableIdentity{
						service: command.ServiceName,
						method:  subcommand.MethodName,
						flag:    flag.Name,
						role:    flagPointerVariable,
					},
					preferred: flag.FullName + "Flag",
					flag:      flag,
				})
			}
		}
	}
	for _, local := range locals {
		candidates = append(candidates, &parserVariableCandidate{
			identity: parserVariableIdentity{
				service: local.ServiceName,
				method:  local.MethodName,
				use:     local.Use,
				role:    transportVariable,
			},
			preferred: local.PreferredName,
			local:     local,
		})
	}
	return candidates
}

// parserImportQualifiers returns every package name referenced by ParseEndpoint.
// Reserving them first prevents a local variable from hiding an imported package.
func parserImportQualifiers(family string, data []*CommandData) []string {
	qualifiers := map[string]struct{}{
		"flag": {},
		"fmt":  {},
		"goa":  {},
		"os":   {},
	}
	switch family {
	case "grpc":
		qualifiers["grpc"] = struct{}{}
		qualifiers["json"] = struct{}{}
		qualifiers["strconv"] = struct{}{}
		qualifiers["utf8"] = struct{}{}
	case "http", "jsonrpc":
		qualifiers["goahttp"] = struct{}{}
		qualifiers["http"] = struct{}{}
		qualifiers["json"] = struct{}{}
		qualifiers["strconv"] = struct{}{}
		qualifiers["utf8"] = struct{}{}
	}
	for _, command := range data {
		if command.PkgName != "" {
			qualifiers[command.PkgName] = struct{}{}
		}
		if command.Interceptors != nil && command.Interceptors.PkgName != "" {
			qualifiers[command.Interceptors.PkgName] = struct{}{}
		}
	}
	result := make([]string, 0, len(qualifiers))
	for qualifier := range qualifiers {
		result = append(result, qualifier)
	}
	sort.Strings(result)
	return result
}

// planParserVariables chooses names for parameters and local values written by
// the parser templates after all imported package names are reserved.
func planParserVariables(scope *codegen.NameScope, family string) *ParserVariablesData {
	variables := &ParserVariablesData{
		ServiceName:    scope.Unique("svcn"),
		ServiceFlags:   scope.Unique("svcf"),
		MethodName:     scope.Unique("epn"),
		MethodFlags:    scope.Unique("epf"),
		Data:           scope.Unique("data"),
		Endpoint:       scope.Unique("endpoint"),
		Error:          scope.Unique("err"),
		Client:         scope.Unique("c"),
		ParsedValue:    scope.Unique("v"),
		ConvertedValue: scope.Unique("val"),
	}
	switch family {
	case "grpc":
		variables.Connection = scope.Unique("cc")
		variables.Options = scope.Unique("opts")
	case "http", "jsonrpc":
		variables.Scheme = scope.Unique("scheme")
		variables.Host = scope.Unique("host")
		variables.Doer = scope.Unique("doer")
		variables.Encoder = scope.Unique("enc")
		variables.Decoder = scope.Unique("dec")
		variables.Restore = scope.Unique("restore")
		variables.Dialer = scope.Unique("dialer")
	}
	return variables
}

// compareParserVariable orders exact design identities before the preferred Go
// spelling, so the same design always receives the same suffix.
func compareParserVariable(left, right *parserVariableCandidate) int {
	for _, compared := range []int{
		cmp.Compare(left.identity.service, right.identity.service),
		cmp.Compare(left.identity.method, right.identity.method),
		cmp.Compare(left.identity.flag, right.identity.flag),
		cmp.Compare(left.identity.use, right.identity.use),
		cmp.Compare(left.identity.role, right.identity.role),
		cmp.Compare(left.preferred, right.preferred),
	} {
		if compared != 0 {
			return compared
		}
	}
	return 0
}

// assign stores one exact name on the data read by its definition and uses.
func (candidate *parserVariableCandidate) assign(name string) {
	switch {
	case candidate.command != nil:
		candidate.command.FlagSetVar = name
	case candidate.subcommand != nil:
		candidate.subcommand.FlagSetVar = name
	case candidate.flag != nil:
		candidate.flag.PointerVar = name
	case candidate.interceptor != nil:
		candidate.interceptor.ParserVar = name
	case candidate.local != nil:
		candidate.local.VarName = name
	}
}

// commandDeclarationNames returns fixed-size bytes derived from every service
// and method name written into one parser file.
func commandDeclarationNames(commands []CommandDeclarationInput) [sha256.Size]byte {
	var encoded []byte
	for _, command := range commands {
		encoded = binary.AppendUvarint(encoded, uint64(len(command.Service)))
		encoded = append(encoded, command.Service...)
		encoded = binary.AppendUvarint(encoded, uint64(len(command.Methods)))
		for _, method := range command.Methods {
			encoded = binary.AppendUvarint(encoded, uint64(len(method)))
			encoded = append(encoded, method...)
		}
	}
	return sha256.Sum256(encoded)
}
