package codegen

import (
	"os"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ExampleCLIFiles returns example JSON-RPC client CLI implementation.
func ExampleCLIFiles(genpkg string, data *httpcodegen.ServicesData, files []*codegen.File) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range data.Root.API.Servers {
		if m := exampleCLI(genpkg, data, svr, files); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

func exampleCLI(genpkg string, data *httpcodegen.ServicesData, svr *expr.ServerExpr, files []*codegen.File) *codegen.File {
	svrdata := example.Servers.Get(svr, data.Root)
	path := filepath.Join("cmd", svrdata.Dir+"-cli", "jsonrpc.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}

	// Retrieve existing HTTP CLI file or create a new one
	var file *codegen.File
	httppath := filepath.Join("cmd", svrdata.Dir+"-cli", "http.go")
	for _, f := range files {
		if f.Path == httppath {
			file = f
			break
		}
	}
	if file == nil {
		// Create new JSON-RPC CLI file using HTTP as template
		file = httpcodegen.ExampleCLI(genpkg, svr, data)
		if file == nil {
			return nil
		}
	}

	var svcdata []*httpcodegen.ServiceData
	for _, svc := range svr.Services {
		if sd := data.Get(svc); sd != nil {
			svcdata = append(svcdata, sd)
		}
	}

	// Modify the file to be JSON-RPC specific
	file.Path = path
	updateFileForJSONRPC(file, data.Root)
	
	return file
}

func updateFileForJSONRPC(file *codegen.File, root *expr.RootExpr) {
	// Update imports to include JSON-RPC specific ones
	header := file.SectionTemplates[0]
	codegen.AddImport(header, &codegen.ImportSpec{Path: "bytes"})
	codegen.AddImport(header, &codegen.ImportSpec{Path: "encoding/json"})
	
	// Update sections to be JSON-RPC specific
	var sections []*codegen.SectionTemplate
	for _, s := range file.SectionTemplates {
		switch s.Name {
		case "cli-http-start":
			// Replace with JSON-RPC start function
			s.Name = "cli-jsonrpc-start"
			s.Source = doJSONRPCTemplate
			s.Data = map[string]any{
				"Root": root,
			}
		case "cli-http-streaming":
			// Skip streaming for JSON-RPC
			continue
		case "cli-http-end":
			// Replace with JSON-RPC end function
			s.Name = "cli-jsonrpc-end"
			s.Source = doJSONRPCRouteTemplate
			s.Data = map[string]any{
				"Root": root,
			}
		case "cli-http-usage":
			// Keep usage as is
		}
		sections = append(sections, s)
	}
	
	file.SectionTemplates = sections
}

const doJSONRPCTemplate = `
func doJSONRPC(scheme, host string, doer goahttp.Doer, serviceName, methodName string, payload any) (goa.Endpoint, any, error) {
	// Map service names to their JSON-RPC endpoint paths
	rpcPaths := map[string]string{
		{{- range .Root.API.JSONRPC.Services }}
		"{{ .Name }}": "{{- range .HTTPEndpoints }}{{- range .Routes }}{{- if eq .Method "POST" }}{{ .Path }}{{- end }}{{- end }}{{- end }}",
		{{- end }}
	}
	
	rpcPath, ok := rpcPaths[serviceName]
	if !ok {
		return nil, nil, fmt.Errorf("unknown JSON-RPC service: %s", serviceName)
	}
	
	return func(ctx context.Context, req any) (any, error) {
		// Construct JSON-RPC request
		request := map[string]any{
			"jsonrpc": "2.0",
			"method":  methodName,
			"params":  req,
			"id":      1,
		}
		
		// Marshal to JSON
		body, err := json.Marshal(request)
		if err != nil {
			return nil, err
		}
		
		// Create HTTP request
		url := scheme + "://" + host + rpcPath
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Execute request
		resp, err := doer.Do(httpReq)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		// Parse JSON-RPC response
		var jsonrpcResp struct {
			Result any ` + "`json:\"result\"`" + `
			Error  *struct {
				Code    int    ` + "`json:\"code\"`" + `
				Message string ` + "`json:\"message\"`" + `
			} ` + "`json:\"error\"`" + `
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&jsonrpcResp); err != nil {
			return nil, err
		}
		
		if jsonrpcResp.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error %d: %s", jsonrpcResp.Error.Code, jsonrpcResp.Error.Message)
		}
		
		return jsonrpcResp.Result, nil
	}, payload, nil
}
`


const doJSONRPCRouteTemplate = `
func doJSONRPCRoute(scheme, host string, timeout int, debug bool) (goa.Endpoint, any, error) {
	var doer goahttp.Doer
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}
	
	return parseEndpointJSONRPC(scheme, host, doer)
}

func parseEndpointJSONRPC(scheme, host string, doer goahttp.Doer) (goa.Endpoint, any, error) {
	if flag.NArg() < 2 {
		return nil, nil, fmt.Errorf("not enough arguments")
	}
	
	serviceName := flag.Arg(0)
	methodName := flag.Arg(1)
	
	// For JSON-RPC, we'll handle payload building later
	// For now, return a simple endpoint that calls doJSONRPC
	return doJSONRPC(scheme, host, doer, serviceName, methodName, nil)
}
`