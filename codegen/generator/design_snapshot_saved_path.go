// Saved paths retain each field suffix once instead of copying every ancestor
// into every snapshot state. Audits compare exact bytes; errors receive the
// complete original path only when a difference must be reported.
package generator

type (
	// designPath retains one path suffix and its complete parent path.
	designPath struct {
		parent *designPath
		suffix string
		length int
	}
)

// matches compares the new suffix after the traversal has already verified its
// parent. Reusing that exact parent avoids rereading every ancestor per field.
func (p *designPath) matches(path []byte, parent *designPath, prefixLength int) bool {
	if len(path) != p.length {
		return false
	}
	if len(path) == prefixLength {
		return p == parent
	}
	return p.parent == parent && p.suffix == string(path[prefixLength:])
}

// string joins saved suffixes so mutation errors retain their original paths.
func (p *designPath) string() string {
	path := make([]byte, p.length)
	for current := p; current != nil; current = current.parent {
		copy(path[current.length-len(current.suffix):current.length], current.suffix)
	}
	return string(path)
}
