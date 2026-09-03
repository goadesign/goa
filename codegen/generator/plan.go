// This file stores the input designs, chosen Go names, and output files for one
// run. Built-in file writers and plugins read the same values.
package generator

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

type (
	// Plan holds the input designs, chosen Go names, and generated files for one
	// run. Code that writes files receives it after all Go names are known.
	Plan struct {
		generation          *codegen.Generation
		preparedRoots       []eval.Root
		examples            map[*expr.RootExpr]*expr.ExampleGenerator
		example             []*examplePlanEntry
		openapi             *httpcodegen.OpenAPIPlan
		openapiRoot         *expr.RootExpr
		openapiReplacements []*httpcodegen.OpenAPIPlan
		services            map[*expr.RootExpr]*service.Plan
		serviceOrder        []*service.Plan
		http                map[*expr.RootExpr]*httpcodegen.Plan
		jsonrpcHTTP         map[*expr.RootExpr]*httpcodegen.Plan
		jsonrpc             map[*expr.RootExpr]*jsonrpccodegen.Plan
		grpc                map[*expr.RootExpr]*grpccodegen.Plan
		transports          []*transportPlanEntry
		transportDone       bool
		design              *designSnapshot
	}

	// transportPlanEntry keeps the transport plans for one service design in
	// the order chosen during planning.
	transportPlanEntry struct {
		http        *httpcodegen.Plan
		jsonrpcHTTP *httpcodegen.Plan
		jsonrpc     *jsonrpccodegen.Plan
		grpc        *grpccodegen.Plan
	}

	// examplePlanEntry keeps one copied example root with the plans that write
	// files for that same design.
	examplePlanEntry struct {
		source  *expr.RootExpr
		root    *example.Root
		service *service.Plan
		http    *httpcodegen.ExamplePlan
		jsonrpc *jsonrpccodegen.ExamplePlan
		grpc    *grpccodegen.ExamplePlan
	}
)

// Generation returns the names chosen for Go declarations and imports in this
// run.
func (p *Plan) Generation() *codegen.Generation {
	return p.generation
}

// Service returns the generated service data for root. It panics when root was
// not included in this run.
func (p *Plan) Service(root *expr.RootExpr) *service.Plan {
	plan, ok := p.services[root]
	if !ok {
		panic(fmt.Sprintf("service plan requested for unplanned design root %q", root.API.Name))
	}
	return plan
}

// HTTP returns the ordinary HTTP plan created for root. It returns false for a
// different root value and for designs exposed only through JSON-RPC.
func (p *Plan) HTTP(root *expr.RootExpr) (*httpcodegen.Plan, bool) {
	plan, ok := p.http[root]
	return plan, ok
}

// GRPC returns the gRPC plan created for the exact design root. It returns
// false when the root was not included in gRPC planning.
func (p *Plan) GRPC(root *expr.RootExpr) (*grpccodegen.Plan, bool) {
	plan, ok := p.grpc[root]
	return plan, ok
}

// JSONRPC returns the JSON-RPC plan created for the exact design root. It
// returns false when the root was not included in JSON-RPC planning.
func (p *Plan) JSONRPC(root *expr.RootExpr) (*jsonrpccodegen.Plan, bool) {
	plan, ok := p.jsonrpc[root]
	return plan, ok
}

// Example returns a separate copy of the example server description created
// for the exact design root. It returns false when the root was not included
// in example planning.
func (p *Plan) Example(root *expr.RootExpr) (*example.Root, bool) {
	for _, entry := range p.example {
		if entry.source == root {
			return copyExampleRoot(entry.root), true
		}
	}
	return nil, false
}

// ReplaceOpenAPI replaces the OpenAPI documents for the application root with
// files already built by plans. It must be called during plugin planning,
// before generated names become final.
func (p *Plan) ReplaceOpenAPI(root *expr.RootExpr, plans ...*httpcodegen.OpenAPIPlan) error {
	if root == nil {
		return fmt.Errorf("OpenAPI replacement root is nil")
	}
	if root != p.openapiRoot {
		return fmt.Errorf("root %q is not the application design root", root.API.Name)
	}
	if p.generation.Frozen() {
		return fmt.Errorf("OpenAPI documents cannot be replaced after generation freeze")
	}
	if len(plans) == 0 {
		return fmt.Errorf("OpenAPI replacement requires at least one plan")
	}
	for index, plan := range plans {
		if plan == nil {
			return fmt.Errorf("OpenAPI replacement plan %d is nil", index)
		}
	}
	if err := validateOpenAPIPlanPaths(plans); err != nil {
		return err
	}
	p.openapiReplacements = append([]*httpcodegen.OpenAPIPlan(nil), plans...)
	return nil
}

// exampleGenerator returns the example values created for root. It panics when
// root was not included in this run.
func (p *Plan) exampleGenerator(root *expr.RootExpr) *expr.ExampleGenerator {
	generator, ok := p.examples[root]
	if !ok {
		panic(fmt.Sprintf("example generator requested for unplanned design root %q", root.API.Name))
	}
	return generator
}

// link completes each service and then builds the protocol files that use it.
func (p *Plan) link() error {
	for _, plan := range p.serviceOrder {
		if err := plan.Link(); err != nil {
			return err
		}
	}
	for _, transport := range p.transports {
		if plan := transport.http; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := transport.jsonrpcHTTP; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := transport.jsonrpc; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := transport.grpc; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
	}
	return nil
}

// verifyPreparedDesign reports the first service design value changed after
// planning and names the code that changed it.
func (p *Plan) verifyPreparedDesign(operation string) error {
	path, err := p.design.changedPath(p.preparedRoots)
	if err != nil {
		return fmt.Errorf("%s left prepared design unverifiable: %w", operation, err)
	}
	if path != "" {
		return fmt.Errorf("%s mutated prepared design at %s", operation, path)
	}
	return nil
}

// validateOpenAPIPlanPaths rejects two OpenAPI files that use the same path,
// including paths that differ only by letter case.
func validateOpenAPIPlanPaths(plans []*httpcodegen.OpenAPIPlan) error {
	var paths []string
	for _, plan := range plans {
		for _, file := range plan.Files() {
			for _, existing := range paths {
				if existing == file.Path {
					return fmt.Errorf("OpenAPI plans use the same output path %q", file.Path)
				}
				if strings.EqualFold(existing, file.Path) {
					return fmt.Errorf(
						"OpenAPI paths %q and %q collide on a case-insensitive filesystem",
						existing,
						file.Path,
					)
				}
			}
			paths = append(paths, file.Path)
		}
	}
	return nil
}
