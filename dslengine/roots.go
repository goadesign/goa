package dslengine

// Roots returns the default RootDefinitions for dslengine
func Roots() RootDefinitions {
	return roots
}

func init() {
	roots = &rootDefinitions{}
}

// NewRootDefinitions returns a new RootDefinitions with the added given roots
// pre-registered.
// It is useful when testing generators.
func NewRootDefinitions(roots ...Root) RootDefinitions {
	r := &rootDefinitions{}
	for _, root := range roots {
		r.Register(root)
	}
	return r
}

// rootDefinitions implements the RootDefinitions interface
type rootDefinitions struct {
	roots []Root
}

func (defs *rootDefinitions) Register(r Root) {
	defs.roots = append(defs.roots, r)
}

// Register adds a Root to the default RootDefinitions returned by Roots
func Register(r Root) {
	Roots().Register(r)
}

func (defs *rootDefinitions) IterateRoots(fn func(Root) error) error {
	for _, r := range defs.roots {
		if err := fn(r); err != nil {
			return err
		}
	}
	return nil
}
