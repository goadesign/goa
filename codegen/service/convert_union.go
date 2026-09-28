// External conversions use the authored union branches and the external Go
// type's public methods. Planning rejects incompatible methods before rendering
// and never inspects the fields used to store the selected branch.
package service

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// externalUnionBranches checks the public methods used by Goa's union
// converter and returns each branch's exact Go type in schema order.
func externalUnionBranches(t reflect.Type, union *expr.Union) ([]reflect.Type, error) {
	if t.Kind() != reflect.Struct || !token.IsExported(t.Name()) || t.PkgPath() == "" {
		return nil, fmt.Errorf("union %q requires an exported named external struct, got %s", union.Name(), t)
	}
	kind, ok := t.MethodByName("Kind")
	if !ok || kind.Type.IsVariadic() || kind.Type.NumIn() != 1 ||
		kind.Type.NumOut() != 1 || kind.Type.Out(0).Kind() != reflect.String {
		return nil, fmt.Errorf("external union %s requires Kind() with one string result", t)
	}

	pointer := reflect.PointerTo(t)
	expected := make(map[string]struct{}, len(union.Values)*2)
	branches := make([]reflect.Type, len(union.Values))
	for i, branch := range union.Values {
		suffix := codegen.Goify(branch.Name, true)
		getterName, setterName := "As"+suffix, "Set"+suffix
		if _, exists := expected[getterName]; exists {
			return nil, fmt.Errorf("union %q branches share method %s", union.Name(), getterName)
		}
		expected[getterName] = struct{}{}
		expected[setterName] = struct{}{}
		getter, exists := t.MethodByName(getterName)
		if !exists || getter.Type.IsVariadic() || getter.Type.NumIn() != 1 ||
			getter.Type.NumOut() != 2 || getter.Type.Out(1) != reflect.TypeFor[bool]() {
			return nil, fmt.Errorf("external union %s requires %s() (T, bool)", t, getterName)
		}
		value := getter.Type.Out(0)
		setter, exists := pointer.MethodByName(setterName)
		if !exists || setter.Type.IsVariadic() || setter.Type.NumIn() != 2 ||
			setter.Type.NumOut() != 0 || setter.Type.In(1) != value {
			return nil, fmt.Errorf("external union %s requires (*%s).%s(%s) with no results", t, t.Name(), setterName, value)
		}
		if _, onValue := t.MethodByName(setterName); onValue {
			return nil, fmt.Errorf("external union %s requires a pointer receiver for %s", t, setterName)
		}

		// Branches containing objects or unions use pointers. Other branches
		// use their scalar, slice, map or interface value directly.
		wantPointer := expr.IsObject(branch.Attribute.Type) || expr.IsUnion(branch.Attribute.Type)
		if wantPointer != (value.Kind() == reflect.Pointer) {
			return nil, fmt.Errorf("external union %s branch %q has incompatible pointer form %s", t, branch.Name, value)
		}
		if wantPointer && value.Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("external union %s branch %q requires a pointer to a struct, got %s", t, branch.Name, value)
		}
		named := value
		if wantPointer {
			named = value.Elem()
		}
		if named.PkgPath() != "" && !token.IsExported(named.Name()) {
			return nil, fmt.Errorf("external union %s branch %q uses unexported type %s", t, branch.Name, named)
		}
		branches[i] = value
	}
	for i := 0; i < pointer.NumMethod(); i++ {
		method := pointer.Method(i)
		if !strings.HasPrefix(method.Name, "As") && !strings.HasPrefix(method.Name, "Set") {
			continue
		}
		if _, exists := expected[method.Name]; !exists {
			return nil, fmt.Errorf("external union %s method %s has no authored branch", t, method.Name)
		}
	}
	return branches, nil
}

// buildExternalUnion records a union using its public Go type and branch
// types. The existing transform can then call its public accessors and setters.
func buildExternalUnion(dt *expr.DataType, t reflect.Type, ref expr.DataType, rec dtRec) error {
	authored := expr.AsUnion(ref)
	branches, err := externalUnionBranches(t, authored)
	if err != nil {
		return fmt.Errorf("%s: %w", rec.path, err)
	}
	union := &expr.Union{
		TypeName: t.Name(),
		TypeKey:  authored.TypeKey,
		ValueKey: authored.ValueKey,
		Values:   make([]*expr.NamedAttributeExpr, len(authored.Values)),
	}
	named := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: union},
		TypeName:      t.Name(),
		UID:           t.PkgPath() + "#" + t.Name(),
	}
	*dt = named
	rec.seen[externalTypePair{external: t, authored: ref}] = named
	rec.named[named] = t
	for i, branch := range authored.Values {
		external := branches[i]
		// Public object and union branch methods already require one pointer.
		// Their underlying type supplies the declaration used by the converter.
		if external.Kind() == reflect.Pointer {
			external = external.Elem()
		}
		var value expr.DataType
		if err := buildDesignType(&value, external, branch.Attribute.Type, appendPath(rec, "."+branch.Name)); err != nil {
			return err
		}
		union.Values[i] = &expr.NamedAttributeExpr{
			Name:      branch.Name,
			Attribute: &expr.AttributeExpr{Type: value},
		}
	}
	return nil
}
