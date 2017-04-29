package codegen

import "strconv"

type (
	// NameScope defines a naming scope.
	NameScope struct {
		names  map[interface{}]string
		counts map[string]int
	}
)

// NewNameScope creates an empty name scope.
func NewNameScope() *NameScope {
	return &NameScope{
		names:  make(map[interface{}]string),
		counts: make(map[string]int),
	}
}

// Unique builds a name for key using name and optionally suffix to construct a
// unique value. If suffix is not specified or if appending suffix does not
// produce a unique value then the smallest integer value that makes the value
// unique is appended. The resulting value can be retrieved by calling Get or
// calling Unique again with the same key.
// Note that uniqueness is only guaranteed for a given instance of NameScope.
func (s *NameScope) Unique(key interface{}, name string, suffix ...string) string {
	if n, ok := s.names[key]; ok {
		return n
	}
	var (
		i   int
		suf string
	)
	_, ok := s.counts[name]
	if !ok {
		goto done
	}
	if len(suffix) > 0 {
		suf = suffix[0]
	}
	name += suf
	i, ok = s.counts[name]
	if !ok {
		goto done
	}
	name += strconv.Itoa(i + 1)
done:
	s.counts[name] = i + 1
	s.names[key] = name
	return name
}

// Get returns the name previously computed by Unique for the given key. It
// returns an empty string if there isn't one.
func (s *NameScope) Get(key interface{}) string {
	return s.names[key]
}
