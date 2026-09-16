// This file assigns every generated service package path once for a complete
// planning run. Later planning phases read the retained path instead of
// rebuilding it from a service name.
package service

import (
	"fmt"
	"path"
	"sort"
	"strconv"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// serviceDesignID identifies declarations contributed by one service in one
	// API. Two roots with the same value cannot be ordered without another design
	// fact, so planning rejects them.
	serviceDesignID struct {
		api     string
		service string
	}
)

// allocateServicePackagePaths returns one generated package path for every
// exact service name in inputs. Exact names share a path across APIs. Different
// names that produce the same normal path receive numeric suffixes ordered by
// their authored names.
func allocateServicePackagePaths(genpkg string, inputs []PlanInput) (map[string]string, error) {
	names := make(map[string]struct{})
	designs := make(map[serviceDesignID]*expr.RootExpr)
	for _, input := range inputs {
		for _, service := range input.Root.Services {
			identity := serviceDesignID{api: input.Root.API.Name, service: service.Name}
			if root := designs[identity]; root != nil && root != input.Root {
				return nil, fmt.Errorf(
					"service %q in API %q is planned by more than one root",
					service.Name,
					input.Root.API.Name,
				)
			}
			designs[identity] = input.Root
			names[service.Name] = struct{}{}
		}
	}

	orderedNames := make([]string, 0, len(names))
	groups := make(map[string][]string)
	reserved := make(map[string]struct{})
	for name := range names {
		orderedNames = append(orderedNames, name)
		base := servicePackageName(name)
		groups[base] = append(groups[base], name)
		reserved[base] = struct{}{}
	}
	sort.Strings(orderedNames)
	for _, names := range groups {
		sort.Strings(names)
	}

	assignedNames := make(map[string]string, len(names))
	used := make(map[string]struct{}, len(names))
	for _, name := range orderedNames {
		base := servicePackageName(name)
		if groups[base][0] == name {
			assignedNames[name] = base
			used[base] = struct{}{}
			continue
		}
		for suffix := 2; ; suffix++ {
			candidate := base + strconv.Itoa(suffix)
			if _, exists := reserved[candidate]; exists {
				continue
			}
			if _, exists := used[candidate]; exists {
				continue
			}
			assignedNames[name] = candidate
			used[candidate] = struct{}{}
			break
		}
	}

	paths := make(map[string]string, len(assignedNames))
	for name, packageName := range assignedNames {
		paths[name] = path.Join(genpkg, packageName)
	}
	return paths, nil
}

// servicePackageName returns the package directory naturally produced by one
// authored service name before collisions are resolved.
func servicePackageName(name string) string {
	return codegen.SnakeCase(codegen.Goify(name, false))
}
