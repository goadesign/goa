package eval

import (
	"fmt"
	"reflect"
)

var (
	// Context contains the state used by the initiator to execute the DSL.
	Context = &DSLContext{}
)

type (
	// DSLContext is the data structure that contains the DSL execution state.
	DSLContext struct {
		// Stack represents the current execution stack.
		Stack Stack
		// Errors contains the DSL execution errors for the current expression set.
		// Errors is an instance of MultiError.
		Errors error

		// roots is the list of DSL roots as registered by all loaded DSLs.
		roots []Root
		// dslPackages keeps track of the DSL package import paths so the initiator may skip
		// any callstack frame that belongs to them when computing error locations.
		dslPackages []string
	}

	// Stack represents the expression evaluation stack. The stack is appended to each time the
	// initiator executes an expression source DSL.
	Stack []Expression
)

// Register appends a root expression to the current Context root expressions. Each root expression
// may only be registered once.
func Register(r Root) error {
	for _, o := range Context.roots {
		if r.DSLName() == o.DSLName() {
			return fmt.Errorf("duplicate DSL %s", r.DSLName())
		}
	}
	t := reflect.TypeOf(r)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	Context.dslPackages = append(Context.dslPackages, t.PkgPath())
	Context.roots = append(Context.roots, r)

	return nil
}

// Current evaluation context, i.e. object being currently built by DSL
func (s Stack) Current() Expression {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Error builds the error message from the current context errors.
func (c *DSLContext) Error() string {
	if c.Errors != nil {
		return c.Errors.Error()
	}
	return ""
}

// SortRoots orders the DSL roots making sure dependencies are last. It returns an error if there
// is a dependency cycle.
func (c *DSLContext) SortRoots() ([]Root, error) {
	roots := c.roots
	if len(roots) == 0 {
		return nil, nil
	}
	// First flatten dependencies for each root
	rootDeps := make(map[string][]Root, len(roots))
	rootByName := make(map[string]Root, len(roots))
	for _, r := range roots {
		sorted := sortDependencies(roots, r, func(r Root) []Root { return r.DependsOn() })
		length := len(sorted)
		for i := 0; i < length/2; i++ {
			sorted[i], sorted[length-i-1] = sorted[length-i-1], sorted[i]
		}
		rootDeps[r.DSLName()] = sorted
		rootByName[r.DSLName()] = r
	}
	// Now check for cycles
	for name, deps := range rootDeps {
		root := rootByName[name]
		for otherName, otherdeps := range rootDeps {
			other := rootByName[otherName]
			if root.DSLName() == other.DSLName() {
				continue
			}
			dependsOnOther := false
			for _, dep := range deps {
				if dep.DSLName() == other.DSLName() {
					dependsOnOther = true
					break
				}
			}
			if dependsOnOther {
				for _, dep := range otherdeps {
					if dep.DSLName() == root.DSLName() {
						return nil, fmt.Errorf("dependency cycle: %s and %s depend on each other (directly or not)",
							root.DSLName(), other.DSLName())
					}
				}
			}
		}
	}
	// Now sort top level DSLs
	var sorted []Root
	for _, r := range roots {
		s := sortDependencies(roots, r, func(r Root) []Root { return rootDeps[r.DSLName()] })
		for _, s := range s {
			found := false
			for _, r := range sorted {
				if r.DSLName() == s.DSLName() {
					found = true
					break
				}
			}
			if !found {
				sorted = append(sorted, s)
			}
		}
	}
	return sorted, nil
}

// Record appends an error to the context Errors field.
func (c *DSLContext) Record(err *Error) {
	if c.Errors == nil {
		c.Errors = MultiError{err}
	} else {
		c.Errors = append(c.Errors.(MultiError), err)
	}
}

// sortDependencies sorts the depencies of the given root in the given slice.
func sortDependencies(roots []Root, root Root, depFunc func(Root) []Root) []Root {
	seen := make(map[string]bool, len(roots))
	var sorted []Root
	sortDependenciesR(root, seen, &sorted, depFunc)
	return sorted
}

// sortDependenciesR sorts the depencies of the given root in the given slice.
func sortDependenciesR(root Root, seen map[string]bool, sorted *[]Root, depFunc func(Root) []Root) {
	for _, dep := range depFunc(root) {
		if !seen[dep.DSLName()] {
			seen[root.DSLName()] = true
			sortDependenciesR(dep, seen, sorted, depFunc)
		}
	}
	*sorted = append(*sorted, root)
}
