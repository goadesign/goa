package expr

import (
	"fmt"
	"sort"
	"strings"
)

const unionVariantTagMetaKey = "oneof:type:tag"

// UnionVariantPublicName returns the stable public-facing name for the union
// variant data type.
func UnionVariantPublicName(dt DataType) string {
	if ut, ok := dt.(UserType); ok && ut.Attribute() != nil {
		if name, ok := ut.Attribute().Meta.Last("name:original"); ok && strings.TrimSpace(name) != "" {
			return name
		}
		if name, ok := ut.Attribute().Meta.Last("openapi:typename"); ok && strings.TrimSpace(name) != "" {
			return name
		}
	}
	return dt.Name()
}

// DerivedUnionVariantNames returns deterministic discriminator names for the
// given union variant data types.
func DerivedUnionVariantNames(types []DataType) []string {
	bases := make([]string, len(types))
	stableKeys := make([]string, len(types))
	for i, dt := range types {
		name := strings.TrimSpace(UnionVariantPublicName(dt))
		if name == "" {
			name = "Value"
		}
		bases[i] = Title(name)
		stableKeys[i] = UnionVariantPublicName(dt) + ":" + dt.Hash()
	}
	return UniqueStableNames(bases, stableKeys, func(base string, ordinal int) string {
		return fmt.Sprintf("%s%d", base, ordinal)
	})
}

// DerivedUnionTypeName returns the deterministic type name for a derived union
// given the already-derived variant names.
func DerivedUnionTypeName(names []string) string {
	if len(names) == 0 {
		return "Union"
	}
	sortedNames := append([]string(nil), names...)
	sort.Strings(sortedNames)
	return strings.Join(sortedNames, "Or")
}

// UnionVariantTag returns the wire discriminator value for a union branch.
// It prefers explicit metadata and falls back to the branch name.
func UnionVariantTag(nat *NamedAttributeExpr) string {
	if nat == nil || nat.Attribute == nil {
		return ""
	}
	if nat.Attribute.Meta != nil {
		if tag, ok := nat.Attribute.Meta.Last(unionVariantTagMetaKey); ok && strings.TrimSpace(tag) != "" {
			return tag
		}
	}
	if ut, ok := nat.Attribute.Type.(UserType); ok && ut.Attribute() != nil && ut.Attribute().Meta != nil {
		if tag, ok := ut.Attribute().Meta.Last(unionVariantTagMetaKey); ok && strings.TrimSpace(tag) != "" {
			return tag
		}
	}
	return nat.Name
}

// UniqueStableNames deduplicates base names using the corresponding stable
// keys to produce deterministic suffix assignment across runs.
func UniqueStableNames(bases, stableKeys []string, suffix func(base string, ordinal int) string) []string {
	if len(bases) != len(stableKeys) {
		panic("bases and stableKeys must have the same length")
	}
	if len(bases) == 0 {
		return nil
	}
	reserved := make(map[string]struct{}, len(bases))
	for _, base := range bases {
		reserved[base] = struct{}{}
	}

	names := make([]string, len(bases))
	groups := make(map[string][]int, len(bases))
	for i, base := range bases {
		groups[base] = append(groups[base], i)
	}
	for base, indexes := range groups {
		if len(indexes) == 1 {
			names[indexes[0]] = base
			continue
		}
		sortedIndexes := append([]int(nil), indexes...)
		sort.SliceStable(sortedIndexes, func(i, j int) bool {
			left := stableKeys[sortedIndexes[i]]
			right := stableKeys[sortedIndexes[j]]
			if left == right {
				return sortedIndexes[i] < sortedIndexes[j]
			}
			return left < right
		})
		for offset, idx := range sortedIndexes {
			if offset == 0 {
				names[idx] = base
				continue
			}
			for ordinal := offset + 1; ; ordinal++ {
				name := suffix(base, ordinal)
				if _, ok := reserved[name]; ok {
					continue
				}
				names[idx] = name
				reserved[name] = struct{}{}
				break
			}
		}
	}

	return names
}
