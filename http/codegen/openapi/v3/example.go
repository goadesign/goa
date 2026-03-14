package openapiv3

import (
	"fmt"
	"hash/fnv"
	"sort"

	"goa.design/goa/v3/expr"
)

type (
	// exampler is the interface used to initialize the example of an
	// OpenAPI object.
	exampler interface {
		setExample(any)
		setExamples(map[string]*ExampleRef)
	}
)

// initExample sets the example or examples of the given object.
func initExamples(obj exampler, attr *expr.AttributeExpr, r *expr.ExampleGenerator) {
	examples := attr.ExtractUserExamples()
	sf := &schemafier{rand: r}
	switch {
	case len(examples) > 1:
		refs := make(map[string]*ExampleRef, len(examples))
		for _, ex := range examples {
			example := &Example{
				Summary:     ex.Summary,
				Description: ex.Description,
				Value:       canonicalizeExampleValue(attr, ex.Value, sf),
			}
			refs[ex.Summary] = &ExampleRef{Value: example}
		}
		obj.setExamples(refs)
		return
	case len(examples) > 0:
		obj.setExample(canonicalizeExampleValue(attr, examples[0].Value, sf))
	default:
		obj.setExample(generatedExampleValue(attr, sf))
	}
}

func canonicalizeExampleValue(attr *expr.AttributeExpr, value any, sf *schemafier) any {
	if attr == nil || value == nil {
		return value
	}
	switch actual := attr.Type.(type) {
	case *expr.Union:
		return canonicalizeUnionExample(actual, value, sf)
	case *expr.Object:
		return canonicalizeObjectExample(actual, value, sf)
	case *expr.Array:
		return canonicalizeArrayExample(actual, value, sf)
	case *expr.Map:
		return canonicalizeMapExample(actual, value, sf)
	case expr.UserType:
		if expr.IsAlias(actual) {
			return canonicalizeExampleValue(actual.Attribute(), value, sf)
		}
		return canonicalizeExampleValue(actual.Attribute(), value, sf)
	default:
		return value
	}
}

func canonicalizeGeneratedExampleValue(attr *expr.AttributeExpr, value any, sf *schemafier) any {
	if attr == nil {
		return value
	}
	if union := expr.AsUnion(attr.Type); union != nil {
		return buildUnionExample(union, sf)
	}
	switch actual := attr.Type.(type) {
	case *expr.Object:
		return canonicalizeGeneratedObjectExample(actual, value, sf)
	case *expr.Array:
		return canonicalizeGeneratedArrayExample(actual, value, sf)
	case *expr.Map:
		return canonicalizeGeneratedMapExample(actual, value, sf)
	case expr.UserType:
		return canonicalizeGeneratedExampleValue(actual.Attribute(), value, sf)
	default:
		return value
	}
}

func generatedExampleValue(attr *expr.AttributeExpr, sf *schemafier) any {
	if attr == nil {
		return nil
	}
	if !containsUnionType(attr.Type) {
		if union := expr.AsUnion(attr.Type); union != nil {
			return buildUnionExample(union, sf)
		}
		return canonicalizeGeneratedExampleValue(attr, attr.Example(sf.rand), sf)
	}
	gen := sf.exampleGenerator(attr)
	local := *sf
	local.rand = gen
	if union := expr.AsUnion(attr.Type); union != nil {
		return buildUnionExample(union, &local)
	}
	return canonicalizeGeneratedExampleValue(attr, attr.Example(gen), &local)
}

func canonicalizeUnionExample(union *expr.Union, value any, sf *schemafier) any {
	if raw, ok := value.(map[string]any); ok {
		if branchName, ok := raw[union.GetTypeKey()].(string); ok {
			if branch := findUnionBranch(union, branchName); branch != nil {
				return map[string]any{
					union.GetTypeKey():  branchName,
					union.GetValueKey(): canonicalizeExampleValue(branch.Attribute, raw[union.GetValueKey()], sf),
				}
			}
		}
	}
	branch, ambiguous := findMatchingUnionBranchDetail(union, value)
	if branch != nil {
		return map[string]any{
			union.GetTypeKey():  branch.Name,
			union.GetValueKey(): canonicalizeExampleValue(branch.Attribute, value, sf),
		}
	}
	if ambiguous {
		// Fail closed: dropping an ambiguous user example is safer than
		// canonicalizing it to an arbitrary union branch and changing the
		// contract that the user described.
		return nil
	}
	return nil
}

func canonicalizeGeneratedObjectExample(obj *expr.Object, value any, sf *schemafier) any {
	raw, ok := value.(map[string]any)
	if !ok {
		raw = nil
	}
	res := make(map[string]any, len(*obj))
	for _, nat := range *obj {
		childValue, found := any(nil), false
		useUserExample := false
		if raw != nil {
			childValue, found = raw[nat.Name]
		}
		if len(nat.Attribute.ExtractUserExamples()) > 0 {
			childValue = nat.Attribute.Example(sf.rand)
			found = true
			useUserExample = true
		}
		if found {
			if useUserExample {
				res[nat.Name] = canonicalizeExampleValue(nat.Attribute, childValue, sf)
			} else {
				if expr.AsUnion(nat.Attribute.Type) != nil {
					res[nat.Name] = generatedExampleValue(nat.Attribute, sf)
				} else {
					res[nat.Name] = canonicalizeGeneratedExampleValue(nat.Attribute, childValue, sf)
				}
			}
		}
	}
	if raw != nil {
		for k, v := range raw {
			if _, ok := res[k]; ok {
				continue
			}
			res[k] = v
		}
	}
	return res
}

func canonicalizeGeneratedArrayExample(arr *expr.Array, value any, sf *schemafier) any {
	raw, ok := value.([]any)
	if !ok {
		return value
	}
	res := make([]any, len(raw))
	for i, v := range raw {
		res[i] = canonicalizeGeneratedExampleValue(arr.ElemType, v, sf)
	}
	return res
}

func canonicalizeGeneratedMapExample(mp *expr.Map, value any, sf *schemafier) any {
	raw, ok := value.(map[any]any)
	if ok {
		res := make(map[any]any, len(raw))
		for k, v := range raw {
			res[k] = canonicalizeGeneratedExampleValue(mp.ElemType, v, sf)
		}
		return res
	}
	stringMap, ok := value.(map[string]any)
	if !ok {
		return value
	}
	res := make(map[string]any, len(stringMap))
	for k, v := range stringMap {
		res[k] = canonicalizeGeneratedExampleValue(mp.ElemType, v, sf)
	}
	return res
}

func findUnionBranch(union *expr.Union, name string) *expr.NamedAttributeExpr {
	for _, branch := range union.Values {
		if branch.Name == name {
			return branch
		}
	}
	return nil
}

func findMatchingUnionBranchDetail(union *expr.Union, value any) (*expr.NamedAttributeExpr, bool) {
	var (
		best      *expr.NamedAttributeExpr
		bestScore int
		ambiguous bool
	)
	for _, branch := range union.Values {
		score, ok := exampleMatchScore(branch.Attribute, value)
		if !ok {
			continue
		}
		if best == nil || score > bestScore {
			best = branch
			bestScore = score
			ambiguous = false
			continue
		}
		if score == bestScore {
			// Equal scores mean the raw example does not identify a unique branch.
			// Callers must treat that as ambiguous instead of falling back to the
			// first matching branch.
			ambiguous = true
		}
	}
	if ambiguous {
		return nil, true
	}
	return best, false
}

func (sf *schemafier) exampleGenerator(attr *expr.AttributeExpr) *expr.ExampleGenerator {
	if sf == nil || attr == nil {
		return nil
	}
	seed := fmt.Sprintf("openapi-v3:%d", sf.hashAttribute(attr, fnv.New64()))
	return expr.NewRandom(seed)
}

func containsUnionType(dt expr.DataType) bool {
	return containsUnionTypeSeen(dt, make(map[expr.DataType]struct{}))
}

func containsUnionTypeSeen(dt expr.DataType, seen map[expr.DataType]struct{}) bool {
	switch dt.(type) {
	case nil:
		return false
	case expr.Primitive:
		return false
	}
	if _, ok := seen[dt]; ok {
		return false
	}
	seen[dt] = struct{}{}
	switch actual := dt.(type) {
	case *expr.Union:
		return true
	case expr.UserType:
		return containsUnionTypeSeen(actual.Attribute().Type, seen)
	case *expr.Object:
		for _, nat := range *actual {
			if containsUnionTypeSeen(nat.Attribute.Type, seen) {
				return true
			}
		}
		return false
	case *expr.Array:
		return containsUnionTypeSeen(actual.ElemType.Type, seen)
	case *expr.Map:
		return containsUnionTypeSeen(actual.KeyType.Type, seen) || containsUnionTypeSeen(actual.ElemType.Type, seen)
	default:
		return false
	}
}

func exampleMatchScore(attr *expr.AttributeExpr, value any) (int, bool) {
	if attr == nil {
		return 0, false
	}
	switch actual := attr.Type.(type) {
	case *expr.Union:
		branch, ambiguous := findMatchingUnionBranchDetail(actual, value)
		if branch == nil {
			if ambiguous {
				return 0, false
			}
			return 0, false
		}
		score, ok := exampleMatchScore(branch.Attribute, value)
		if !ok {
			return 0, false
		}
		return score + 1, true
	case *expr.Object:
		return objectExampleMatchScore(actual, value)
	case *expr.Array:
		return arrayExampleMatchScore(actual, value)
	case *expr.Map:
		return mapExampleMatchScore(actual, value)
	case expr.UserType:
		return exampleMatchScore(actual.Attribute(), value)
	default:
		if attr.Type.IsCompatible(value) {
			return 1, true
		}
		return 0, false
	}
}

func firstGeneratedUnionBranch(union *expr.Union) *expr.NamedAttributeExpr {
	if len(union.Values) == 0 {
		return nil
	}
	branches := append([]*expr.NamedAttributeExpr(nil), union.Values...)
	sort.SliceStable(branches, func(i, j int) bool {
		if branches[i].Name == branches[j].Name {
			return branches[i].Attribute.Type.Hash() < branches[j].Attribute.Type.Hash()
		}
		return branches[i].Name < branches[j].Name
	})
	return branches[0]
}

func objectExampleMatchScore(obj *expr.Object, value any) (int, bool) {
	raw, ok := value.(map[string]any)
	if !ok {
		if obj.IsCompatible(value) {
			return 1, true
		}
		return 0, false
	}
	score := 1
	for key, val := range raw {
		nat := obj.Attribute(key)
		if nat == nil {
			return 0, false
		}
		attScore, ok := exampleMatchScore(nat, val)
		if !ok {
			return 0, false
		}
		score += attScore + 1
	}
	return score, true
}

func arrayExampleMatchScore(arr *expr.Array, value any) (int, bool) {
	raw, ok := value.([]any)
	if !ok {
		if arr.IsCompatible(value) {
			return 1, true
		}
		return 0, false
	}
	score := 1
	for _, elem := range raw {
		elemScore, ok := exampleMatchScore(arr.ElemType, elem)
		if !ok {
			return 0, false
		}
		score += elemScore
	}
	return score, true
}

func mapExampleMatchScore(mp *expr.Map, value any) (int, bool) {
	if raw, ok := value.(map[string]any); ok {
		score := 1
		for _, elem := range raw {
			elemScore, ok := exampleMatchScore(mp.ElemType, elem)
			if !ok {
				return 0, false
			}
			score += elemScore
		}
		return score, true
	}
	if raw, ok := value.(map[any]any); ok {
		score := 1
		for _, elem := range raw {
			elemScore, ok := exampleMatchScore(mp.ElemType, elem)
			if !ok {
				return 0, false
			}
			score += elemScore
		}
		return score, true
	}
	if mp.IsCompatible(value) {
		return 1, true
	}
	return 0, false
}

func canonicalizeObjectExample(obj *expr.Object, value any, sf *schemafier) any {
	raw, ok := value.(map[string]any)
	if !ok {
		return value
	}
	res := make(map[string]any, len(raw))
	for k, v := range raw {
		if nat := obj.Attribute(k); nat != nil {
			res[k] = canonicalizeExampleValue(nat, v, sf)
			continue
		}
		res[k] = v
	}
	return res
}

func canonicalizeArrayExample(arr *expr.Array, value any, sf *schemafier) any {
	raw, ok := value.([]any)
	if !ok {
		return value
	}
	res := make([]any, len(raw))
	for i, v := range raw {
		res[i] = canonicalizeExampleValue(arr.ElemType, v, sf)
	}
	return res
}

func canonicalizeMapExample(mp *expr.Map, value any, sf *schemafier) any {
	raw, ok := value.(map[any]any)
	if ok {
		res := make(map[any]any, len(raw))
		for k, v := range raw {
			res[k] = canonicalizeExampleValue(mp.ElemType, v, sf)
		}
		return res
	}
	stringMap, ok := value.(map[string]any)
	if !ok {
		return value
	}
	res := make(map[string]any, len(stringMap))
	for k, v := range stringMap {
		res[k] = canonicalizeExampleValue(mp.ElemType, v, sf)
	}
	return res
}
