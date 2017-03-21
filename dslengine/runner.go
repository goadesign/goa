package dslengine

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

var (
	// Errors contains the DSL execution errors if any.
	Errors MultiError

	// Global DSL evaluation stack
	ctxStack contextStack

	// Registered DSL roots
	roots []Root

	// DSL package paths used to compute error locations (skip the frames in these packages)
	dslPackages map[string]bool
)

type (
	// Error represents an error that occurred while running the API DSL.
	// It contains the name of the file and line number of where the error
	// occurred as well as the original Go error.
	Error struct {
		GoError error
		File    string
		Line    int
	}

	// MultiError collects all DSL errors. It implements error.
	MultiError []*Error

	// DSL evaluation contexts stack
	contextStack []Definition
)

func init() {
	dslPackages = map[string]bool{
		"github.com/goadesign/goa/":            true,
		"github.com/goadesign/goa/middleware/": true,
		"github.com/goadesign/goa/encoding/":   true,
		"github.com/goadesign/goa/logging/":    true,
	}
}

// Register adds a DSL Root to be executed by Run.
func Register(r Root) {
	for _, o := range roots {
		if r.DSLName() == o.DSLName() {
			fmt.Fprintf(os.Stderr, "goagen: duplicate DSL %s", r.DSLName())
			os.Exit(1)
		}
	}
	t := reflect.TypeOf(r)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	dslPackages[t.PkgPath()] = true
	roots = append(roots, r)
}

// Reset uses the registered RootFuncs to re-initialize the DSL roots.
// This is useful to tests.
func Reset() {
	for _, r := range roots {
		r.Reset()
	}
	Errors = nil
}

// Run runs the given root definitions. It iterates over the definition sets
// multiple times to first execute the DSL, the validate the resulting
// definitions and finally finalize them. The executed DSL may register new
// roots to have them be executed (last) in the same run.
func Run() error {
	if len(roots) == 0 {
		return nil
	}
	roots, err := SortRoots()
	if err != nil {
		return err
	}
	Errors = nil
	executed := 0
	recursed := 0
	for executed < len(roots) {
		recursed++
		start := executed
		executed = len(roots)
		for _, root := range roots[start:] {
			root.IterateSets(runSet)
		}
		if recursed > 100 {
			// Let's cross that bridge once we get there
			return fmt.Errorf("too many generated roots, infinite loop?")
		}
	}
	if Errors != nil {
		return Errors
	}
	for _, root := range roots {
		root.IterateSets(validateSet)
	}
	if Errors != nil {
		return Errors
	}
	for _, root := range roots {
		root.IterateSets(finalizeSet)
	}

	return nil
}

// Execute runs the given DSL to initialize the given definition. It returns true on success.
// It returns false and appends to Errors on failure.
// Note that `Run` takes care of calling `Execute` on all definitions that implement Source.
// This function is intended for use by definitions that run the DSL at declaration time rather than
// store the DSL for execution by the dsl engine (usually simple independent definitions).
// The DSL should use ReportError to record DSL execution errors.
func Execute(dsl func(), def Definition) bool {
	if dsl == nil {
		return true
	}
	initCount := len(Errors)
	ctxStack = append(ctxStack, def)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(Errors) <= initCount
}

// CurrentDefinition returns the definition whose initialization DSL is currently being executed.
func CurrentDefinition() Definition {
	current := ctxStack.Current()
	if current == nil {
		return &TopLevelDefinition{}
	}
	return current
}

// IsTopLevelDefinition returns true if the currently evaluated DSL is a root
// DSL (i.e. is not being run in the context of another definition).
func IsTopLevelDefinition() bool {
	_, ok := CurrentDefinition().(*TopLevelDefinition)
	return ok
}

// TopLevelDefinition represents the top-level file definitions, done
// with `var _ = `.  An instance of this object is returned by
// `CurrentDefinition()` when at the top-level.
type TopLevelDefinition struct{}

// Context tells the DSL engine which context we're in when showing
// errors.
func (t *TopLevelDefinition) Context() string { return "top-level" }

// ReportError records a DSL error for reporting post DSL execution.
func ReportError(fm string, vals ...interface{}) {
	var suffix string
	if cur := ctxStack.Current(); cur != nil {
		if ctx := cur.Context(); ctx != "" {
			suffix = fmt.Sprintf(" in %s", ctx)
		}
	} else {
		suffix = " (top level)"
	}
	err := fmt.Errorf(fm+suffix, vals...)
	file, line := computeErrorLocation()
	Errors = append(Errors, &Error{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// FailOnError will exit with code 1 if `err != nil`. This function
// will handle properly the MultiError this dslengine provides.
func FailOnError(err error) {
	if merr, ok := err.(MultiError); ok {
		if len(merr) == 0 {
			return
		}
		fmt.Fprintf(os.Stderr, merr.Error())
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// PrintFilesOrFail will print the file list. Use it with a
// generator's `Generate()` function to output the generated list of
// files or quit on error.
func PrintFilesOrFail(files []string, err error) {
	FailOnError(err)
	fmt.Println(strings.Join(files, "\n"))
}

// IncompatibleDSL should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func IncompatibleDSL() {
	elems := strings.Split(caller(), ".")
	ReportError("invalid use of %s", elems[len(elems)-1])
}

// InvalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func InvalidArgError(expected string, actual interface{}) {
	ReportError("cannot use %#v (type %s) as type %s",
		actual, reflect.TypeOf(actual), expected)
}

// Error returns the error message.
func (m MultiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (de *Error) Error() string {
	if err := de.GoError; err != nil {
		if de.File == "" {
			return err.Error()
		}
		return fmt.Sprintf("[%s:%d] %s", de.File, de.Line, err.Error())
	}
	return ""
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) Current() Definition {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go" or one of the DSL package paths.
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
func computeErrorLocation() (file string, line int) {
	skipFunc := func(file string) bool {
		if strings.HasSuffix(file, "_test.go") { // Be nice with tests
			return false
		}
		file = filepath.ToSlash(file)
		for pkg := range dslPackages {
			if strings.Contains(file, pkg) {
				return true
			}
		}
		return false
	}
	depth := 2
	_, file, line, _ = runtime.Caller(depth)
	for skipFunc(file) {
		depth++
		_, file, line, _ = runtime.Caller(depth)
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return
	}
	f, err := filepath.Rel(wd, file)
	if err != nil {
		return
	}
	file = f
	return
}

// runSet executes the DSL for all definitions in the given set. The definition DSLs may append to
// the set as they execute.
func runSet(set DefinitionSet) error {
	executed := 0
	recursed := 0
	for executed < len(set) {
		recursed++
		for _, def := range set[executed:] {
			executed++
			if source, ok := def.(Source); ok {
				if dsl := source.DSL(); dsl != nil {
					Execute(dsl, source)
				}
			}
		}
		if recursed > 100 {
			return fmt.Errorf("too many generated definitions, infinite loop?")
		}
	}
	return nil
}

// validateSet runs the validation on all the set definitions that define one.
func validateSet(set DefinitionSet) error {
	errors := &ValidationErrors{}
	for _, def := range set {
		if validate, ok := def.(Validate); ok {
			if err := validate.Validate(); err != nil {
				errors.AddError(def, err)
			}
		}
	}
	err := errors.AsError()
	if err != nil {
		Errors = append(Errors, &Error{GoError: err})
	}
	return err
}

// finalizeSet runs the validation on all the set definitions that define one.
func finalizeSet(set DefinitionSet) error {
	for _, def := range set {
		if finalize, ok := def.(Finalize); ok {
			finalize.Finalize()
		}
	}
	return nil
}

// SortRoots orders the DSL roots making sure dependencies are last. It returns an error if there
// is a dependency cycle.
func SortRoots() ([]Root, error) {
	if len(roots) == 0 {
		return nil, nil
	}
	// First flatten dependencies for each root
	rootDeps := make(map[string][]Root, len(roots))
	rootByName := make(map[string]Root, len(roots))
	for _, r := range roots {
		sorted := sortDependencies(r, func(r Root) []Root { return r.DependsOn() })
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
		s := sortDependencies(r, func(r Root) []Root { return rootDeps[r.DSLName()] })
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

// sortDependencies sorts the depencies of the given root in the given slice.
func sortDependencies(root Root, depFunc func(Root) []Root) []Root {
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

// caller returns the name of calling function.
func caller() string {
	pc, file, _, ok := runtime.Caller(2)
	if ok && filepath.Base(file) == "current.go" {
		pc, _, _, ok = runtime.Caller(3)
	}
	if !ok {
		return "<unknown>"
	}

	return runtime.FuncForPC(pc).Name()
}
