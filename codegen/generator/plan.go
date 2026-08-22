// This file stores the input designs, chosen Go names, and output files for one
// run. Built-in file writers and plugins read the same values.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
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
		generation    *codegen.Generation
		preparedRoots []eval.Root
		examples      map[*expr.RootExpr]*expr.ExampleGenerator
		services      map[*expr.RootExpr]*service.Plan
		http          map[*expr.RootExpr]*httpcodegen.Plan
		jsonrpcHTTP   map[*expr.RootExpr]*httpcodegen.Plan
		jsonrpc       map[*expr.RootExpr]*jsonrpccodegen.Plan
		grpc          map[*expr.RootExpr]*grpccodegen.Plan
		transportDone bool
		design        *designSnapshot
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
	for _, root := range serviceRoots(p.preparedRoots) {
		plan, ok := p.services[root]
		if !ok {
			continue
		}
		if err := plan.Link(); err != nil {
			return err
		}
	}
	for _, root := range serviceRoots(p.preparedRoots) {
		if plan := p.http[root]; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := p.jsonrpcHTTP[root]; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := p.jsonrpc[root]; plan != nil {
			if err := plan.Link(); err != nil {
				return err
			}
		}
		if plan := p.grpc[root]; plan != nil {
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
