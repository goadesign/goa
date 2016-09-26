package genserver

import (
	"sort"

	"github.com/goadesign/goa/codegen"
	"github.com/goadesign/goa/http/design"
)

type (
	// EndpointGroupTemplateData is the struct used to render the Endpoint Group template.
	EndpointGroupTemplateData struct {
		// Name of group, i.e. of resource (REST) or service (RPC)
		Name string
		// Endpoints template data
		Endpoints []*EndpointTemplateData
	}

	// EndpointTemplateData contains the data necessary to render a single endpoint.
	EndpointTemplateData struct {
		// Name of endpoint
		Name string
		// RequestType is the request data type.
		RequestType design.UserType
		// ResponseType is the response data type.
		ResponseType design.UserType
	}
)

// GenerateEndpoints generates the given endpoint group data structure, for example:
//
//    // BottleEndpoints is the set of endpoints that expose the Bottle resource endpoints.
//    type BottleEndpoints struct {
//        Show goa.Endpoint
//        Create goa.Endpoint
//    }
//
//    // NewBottleEndpoints wraps the BottleController methods into remote endpoints.
//    func NewBottleEndpoints(c *BottleController) *BottleEndpoints {
//        show := func(ctx context.Context, req interface{}) (resp interface{}, error) {
//            r, ok := req.(*BottleShowRequest)
//            if !ok {
//                return nil, goa.NewUnexpectedType("BottleShowRequest", req)
//            }
//            resp, err := c.Show(r)
//            if f, ok := err.(goa.Failure) {
//                return nil, err
//            }
//            return resp, nil
//        }
//        // create := ...
//        return &BottleEndpoints{
//            Show: show,
//            Create: create,
//        }
//    }
//
//    // Use applies the middleware to all the BottleEndpoints endpoints.
//    func (e *BottleEndpoints) Use(m goa.Middleware) {
//        e.Show = m(e.Show)
//        e.Create = m(e.Create)
//    }
func GenerateEndpoints(file string, outPkg string, encoders, decoders Encoder) error {
	wr, err := NewInitWriter(file)
	if err != nil {
		panic(err) // bug
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
		codegen.SimpleImport("regexp"),
	}
	encoderImports := make(map[string]bool)
	for _, data := range data.Encoders {
		encoderImports[data.PackagePath] = true
	}
	for _, data := range data.Decoders {
		encoderImports[data.PackagePath] = true
	}
	var packagePaths []string
	for packagePath := range encoderImports {
		if packagePath != "github.com/goadesign/goa" {
			packagePaths = append(packagePaths, packagePath)
		}
	}
	sort.Strings(packagePaths)
	for _, packagePath := range packagePaths {
		imports = append(imports, codegen.SimpleImport(packagePath))
	}
	wr.WriteHeader("Service Setup", data.OutPkg, imports)
	wr.Write(encoders, decoders)
	return wr.FormatCode()
}

// GenerateUserTypes iterates through the user types and generates the data structures and
// marshaling code.
func GenerateUserTypes(file, outPkg string, types []*design.UserTypeExpr) error {
	wr, err := NewUserTypesWriter(file)
	if err != nil {
		panic(err) // bug
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	wr.WriteHeader("User Types", outPkg, imports)
	for _, t := range types {
		if err := wr.Execute(t); err != nil {
			return err
		}
	}
	return wr.FormatCode()
}
