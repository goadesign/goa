// This file rejects cycles among packages required by emitted service types.
// It reads the retained Go declarations, not conservative import reservations.
package service

import (
	"fmt"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
)

// requiredPackageImports records direct references in emitted declarations.
// A named field belongs to its own definition, so its contents are not expanded.
type requiredPackageImports map[string]map[string]struct{}

// validateRequiredPackageImports checks declarations and service signatures only
// after all roots have selected their Go layouts. Same-package recursion is valid.
func validateRequiredPackageImports(roots []*rootFacts) error {
	imports := make(requiredPackageImports)
	for _, root := range roots {
		for _, service := range root.services {
			imports.add(service.packagePath, nil)
			for _, typ := range append(append([]*userTypeFacts(nil), service.userTypes...), service.errorTypes...) {
				imports.add(typ.declaration.PackagePath(), typ.layout)
			}
			for _, method := range service.orderedMethods {
				for _, attribute := range []*methodAttributeFacts{method.payload, method.streamingPayload, method.result, method.streamingResult} {
					if attribute == nil {
						continue
					}
					imports.add(service.packagePath, attribute.layout)
					if attribute.definition != nil {
						imports.add(attribute.layout.Owner(), attribute.definition)
					}
				}
				for _, methodError := range method.errors {
					imports.add(service.packagePath, methodError.layout)
				}
				if method.viewedResult != nil {
					// Service-side constructors name the generated views package.
					imports.edge(service.packagePath, service.viewsPath)
				}
			}
			for _, serviceError := range service.errorFacts {
				imports.add(service.packagePath, serviceError.layout)
			}
			for _, projection := range service.projections {
				for _, projected := range projection.types {
					imports.add(service.viewsPath, projected.definition)
				}
			}
			for _, union := range append(append([]*unionFacts(nil), service.unions...), service.viewUnions...) {
				owner := union.declaration.PackagePath()
				imports.add(owner, nil)
				for _, branch := range union.branches {
					imports.add(owner, branch.layout)
				}
			}
		}
	}
	return imports.validate()
}

// add records only references written directly in one type expression.
// Named types and unions stop traversal; their own definitions are added separately.
func (imports requiredPackageImports) add(owner string, layout *codegen.GoTypePlan) {
	if imports[owner] == nil {
		imports[owner] = make(map[string]struct{})
	}
	if layout == nil {
		return
	}
	if direct, ok := layout.Import(); ok {
		imports.edge(owner, direct.Path)
		return
	}
	switch layout.Kind() {
	case codegen.GoNamed, codegen.GoUnion:
		imports.edge(owner, layout.Owner())
	case codegen.GoStruct:
		for _, field := range layout.Fields() {
			imports.add(owner, field)
		}
	case codegen.GoArray:
		imports.add(owner, layout.Elem())
	case codegen.GoMap:
		imports.add(owner, layout.Key())
		imports.add(owner, layout.Elem())
	}
}

// edge records a required import while omitting references inside one package.
func (imports requiredPackageImports) edge(owner, target string) {
	if imports[owner] == nil {
		imports[owner] = make(map[string]struct{})
	}
	if target != owner {
		imports[owner][target] = struct{}{}
	}
}

// validate follows direct dependencies between generated packages in sorted
// order. Imported external packages are outside this generation's graph.
func (imports requiredPackageImports) validate() error {
	roots := make([]string, 0, len(imports))
	for owner := range imports {
		roots = append(roots, owner)
	}
	slices.Sort(roots)
	complete := make(map[string]bool)
	active := make(map[string]int)
	var stack []string
	var visit func(string) error
	visit = func(owner string) error {
		if start, found := active[owner]; found {
			cycle := append(append([]string(nil), stack[start:]...), owner)
			return fmt.Errorf("required generated package import cycle: %s", strings.Join(cycle, " -> "))
		}
		if complete[owner] {
			return nil
		}
		active[owner] = len(stack)
		stack = append(stack, owner)
		targets := make([]string, 0, len(imports[owner]))
		for target := range imports[owner] {
			if _, generated := imports[target]; generated {
				targets = append(targets, target)
			}
		}
		slices.Sort(targets)
		for _, target := range targets {
			if err := visit(target); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		delete(active, owner)
		complete[owner] = true
		return nil
	}
	for _, owner := range roots {
		if err := visit(owner); err != nil {
			return err
		}
	}
	return nil
}
