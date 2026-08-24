// This file writes example command-line programs from copied server data and
// the package names already chosen for this generation.
package example

import (
	"strings"

	"goa.design/goa/v3/codegen"
)

type (
	// clientMainData contains all design values selected before one example
	// command-line client is rendered.
	clientMainData struct {
		// APIName is the API name written in help text.
		APIName string
		// Server contains the copied host, transport, and URL variable settings.
		Server *clientMainServerData
		// HasJSONRPC reports whether the client includes JSON-RPC commands.
		HasJSONRPC bool
		// HasHTTP reports whether the client includes ordinary HTTP commands.
		HasHTTP bool
		// UsageCommands is the sorted command list written in help text.
		UsageCommands []string
		// JSONRPCOnly lists commands handled only by the JSON-RPC client.
		JSONRPCOnly []*jsonRPCServiceData
		// WritesEndpointResult reports whether a command returns one result.
		WritesEndpointResult bool
		// WritesStreamResults reports whether a command receives server results.
		WritesStreamResults bool
	}

	// clientMainServerData contains the URL variables planned for one client
	// main and the hosts that use them.
	clientMainServerData struct {
		*Data
		// Variables lists every URL variable with its client flag names.
		Variables []*mainVariableData
		// Hosts lists each host with the same planned URL variables.
		Hosts []*clientMainHostData
	}

	// clientMainHostData contains one host and its planned URL variables.
	clientMainHostData struct {
		*HostData
		// Variables lists the URL variables used by this host.
		Variables []*mainVariableData
	}
)

// CLIFiles returns one example command-line program for each copied server.
func CLIFiles(root *Root) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.Servers {
		if m := exampleCLIMain(root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleCLIMain writes the command-line program for server.
func exampleCLIMain(root *Root, server *Data) *codegen.File {
	// A server with no HTTP, JSON-RPC, or gRPC service has no client to run.
	if server.DefaultTransport() == nil {
		return nil
	}

	path := server.clientMainPath
	main := &clientMainData{
		APIName:              root.APIName,
		Server:               planClientMainServer(server),
		HasJSONRPC:           server.HasJSONRPC,
		HasHTTP:              server.HasHTTP,
		UsageCommands:        server.usageCommands,
		JSONRPCOnly:          server.jsonRPCOnly,
		WritesEndpointResult: server.writesEndpointResult,
		WritesStreamResults:  server.writesStreamResults,
	}
	specs := packageImports(server.clientPackage, clientMainFixedImports(server))
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-main-start",
			Source: exampleTemplates.Read(clientStartT),
			Data:   main,
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-var-init",
			Source: exampleTemplates.Read(clientVarInitT),
			Data:   main,
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-endpoint-init",
			Source: exampleTemplates.Read(clientEndpointInitT),
			Data:   main,
			FuncMap: map[string]any{
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		}, {
			Name:   "cli-main-end",
			Source: exampleTemplates.Read(clientEndT),
			Data:   main,
		}, {
			Name:   "cli-main-usage",
			Source: exampleTemplates.Read(clientUsageT),
			Data:   main,
			FuncMap: map[string]any{
				"toUpper": strings.ToUpper,
				"join":    strings.Join,
			},
		},
	}
	return &codegen.File{Path: path, SectionTemplates: sections, SkipExist: true}
}

// clientMainFixedImports lists packages whose names are written directly by
// the command-line client templates.
func clientMainFixedImports(server *Data) []*codegen.ImportSpec {
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "errors"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strings"},
	}
	if server.writesEndpointResult || server.writesStreamResults {
		specs = append(specs,
			&codegen.ImportSpec{Path: "encoding/json"},
			&codegen.ImportSpec{Path: "io"},
		)
	}
	if server.writesEndpointResult {
		specs = append(specs, codegen.GoaImport(""))
	}
	return specs
}

// planClientMainServer selects URL flag names that are distinct from the
// built-in client flags.
func planClientMainServer(server *Data) *clientMainServerData {
	fixedFlags := []string{"host", "url", "timeout", "verbose", "v"}
	if server.HasJSONRPC {
		fixedFlags = append(fixedFlags, "jsonrpc", "j")
	}
	variables := planMainVariables(server.Variables, fixedFlags)
	planned := &clientMainServerData{
		Data:      server,
		Variables: variables.all,
		Hosts:     make([]*clientMainHostData, len(server.Hosts)),
	}
	for index, host := range server.Hosts {
		plannedHost := &clientMainHostData{
			HostData:  host,
			Variables: make([]*mainVariableData, len(host.Variables)),
		}
		for variableIndex, variable := range host.Variables {
			plannedHost.Variables[variableIndex] = variables.byName[variable.Name]
		}
		planned.Hosts[index] = plannedHost
	}
	return planned
}
