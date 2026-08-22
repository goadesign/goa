// This file assigns the package function names used by command-line client
// files. HTTP and gRPC planning call these functions before generated names
// are finalized, then pass the returned records to the shared CLI templates.
package cli

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/binary"

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
	}

	// ParserPlan contains the names written to one command parser package.
	ParserPlan struct {
		// Declarations contains the three functions shared by the parser file.
		Declarations *ParserDeclarations
		// Commands contains the help function names for each design service.
		Commands map[string]*CommandPlan
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
)

const (
	parseEndpointRole symbolRole = iota + 1
	usageCommandsRole
	usageExamplesRole
	commandUsageRole
	methodUsageRole
	payloadBuilderRole
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
	plan := &ParserPlan{
		Declarations: &ParserDeclarations{
			ParseEndpoint: parseEndpoint,
			UsageCommands: usageCommands,
			UsageExamples: usageExamples,
		},
		Commands: make(map[string]*CommandPlan, len(commands)),
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
