// This file records the prepared design so Goa can report if a generator
// changes it. Map entries are sorted so repeated runs report the same first
// changed field.
package generator

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// designSnapshot stores the prepared design values for one run.
	designSnapshot struct {
		states     []designState
		references []reflect.Value
	}

	// designState records one scalar, container, or reference reached at path.
	designState struct {
		path  string
		typ   reflect.Type
		value string
	}

	// designSnapshotter records each design value once, even when pointers form
	// a cycle or several fields point to the same value.
	designSnapshotter struct {
		states     []designState
		references []reflect.Value
		visited    map[designVisit]struct{}
	}

	// designVisit identifies one pointer, map, or slice. Slice capacity separates
	// slices that can reach different parts of the same underlying array.
	designVisit struct {
		typ    reflect.Type
		kind   reflect.Kind
		ptr    uintptr
		extent int
	}

	// mapSnapshotEntry stores one map pair after the entries are sorted.
	mapSnapshotEntry struct {
		key      reflect.Value
		value    reflect.Value
		keyOrder mapOrderValue
		order    mapOrderValue
	}

	// mapOrderValue stores enough of a map key or value to sort entries. Pointers
	// are compared by address and other values by their exact contents.
	mapOrderValue struct {
		typ       reflect.Type
		kind      reflect.Kind
		isNil     bool
		boolean   bool
		integer   int64
		unsigned  uint64
		text      string
		reference uintptr
		length    int
		capacity  int
		children  []mapOrderValue
	}
)

var typeMapType = reflect.TypeFor[expr.TypeMap]()

// snapshotPreparedDesign records every value reachable from roots after the
// designs have been prepared.
func snapshotPreparedDesign(roots []eval.Root) (*designSnapshot, error) {
	snapshotter := &designSnapshotter{visited: make(map[designVisit]struct{})}
	for i, root := range roots {
		if err := snapshotter.appendValue(fmt.Sprintf("roots[%d]", i), reflect.ValueOf(root)); err != nil {
			return nil, err
		}
	}
	return &designSnapshot{
		states:     snapshotter.states,
		references: snapshotter.references,
	}, nil
}

// orderedMapEntries returns map entries in an order that does not depend on
// how Go stores the map.
func orderedMapEntries(value reflect.Value) ([]mapSnapshotEntry, error) {
	entries := make([]mapSnapshotEntry, 0, value.Len())
	iterator := value.MapRange()
	for iterator.Next() {
		key := iterator.Key()
		mapValue := iterator.Value()
		keyOrder, err := mapValueOrder(key)
		if err != nil {
			return nil, err
		}
		valueOrder, err := mapValueOrder(mapValue)
		if err != nil {
			return nil, err
		}
		entries = append(entries, mapSnapshotEntry{
			key:      key,
			value:    mapValue,
			keyOrder: keyOrder,
			order:    valueOrder,
		})
	}
	if err := validateMapOrderTypes(entries); err != nil {
		return nil, err
	}
	slices.SortFunc(entries, func(left, right mapSnapshotEntry) int {
		if compared := compareMapOrderValue(left.keyOrder, right.keyOrder); compared != 0 {
			return compared
		}
		return compareMapOrderValue(left.order, right.order)
	})
	return entries, nil
}

// validateMapOrderTypes rejects two runtime types that print the same name but
// cannot be compared. Treating them as equal would make map order vary by run.
func validateMapOrderTypes(entries []mapSnapshotEntry) error {
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			if err := validateMapOrderType(entries[i].keyOrder, entries[j].keyOrder); err != nil {
				return err
			}
			if err := validateMapOrderType(entries[i].order, entries[j].order); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateMapOrderType checks a runtime type and the concrete values stored in
// interface map keys and values.
func validateMapOrderType(left, right mapOrderValue) error {
	if left.typ != right.typ && stableTypeName(left.typ) == stableTypeName(right.typ) {
		return fmt.Errorf("cannot deterministically order distinct reflected map types %q", stableTypeName(left.typ))
	}
	common := min(len(left.children), len(right.children))
	for i := range common {
		if err := validateMapOrderType(left.children[i], right.children[i]); err != nil {
			return err
		}
	}
	return nil
}

// mapValueOrder records enough of a map key or value to sort entries without
// reading through pointers.
func mapValueOrder(value reflect.Value) (mapOrderValue, error) {
	if !value.IsValid() {
		return mapOrderValue{}, nil
	}
	order := mapOrderValue{typ: value.Type(), kind: value.Kind()}
	switch value.Kind() {
	case reflect.Bool:
		order.boolean = value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		order.integer = value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		order.unsigned = value.Uint()
	case reflect.Float32:
		order.unsigned = uint64(math.Float32bits(float32(value.Float())))
	case reflect.Float64:
		order.unsigned = math.Float64bits(value.Float())
	case reflect.Complex64:
		complexValue := complex64(value.Complex())
		order.children = []mapOrderValue{
			{unsigned: uint64(math.Float32bits(real(complexValue)))},
			{unsigned: uint64(math.Float32bits(imag(complexValue)))},
		}
	case reflect.Complex128:
		complexValue := value.Complex()
		order.children = []mapOrderValue{
			{unsigned: math.Float64bits(real(complexValue))},
			{unsigned: math.Float64bits(imag(complexValue))},
		}
	case reflect.String:
		order.text = value.String()
	case reflect.Interface:
		if value.IsNil() {
			order.isNil = true
			break
		}
		inner, err := mapValueOrder(value.Elem())
		if err != nil {
			return mapOrderValue{}, err
		}
		order.children = []mapOrderValue{inner}
	case reflect.Pointer, reflect.Chan:
		if value.IsNil() {
			order.isNil = true
			break
		}
		order.reference = value.Pointer()
	case reflect.UnsafePointer:
		order.reference = value.Pointer()
	case reflect.Slice:
		if value.IsNil() {
			order.isNil = true
			break
		}
		order.reference = value.Pointer()
		order.length = value.Len()
		order.capacity = value.Cap()
	case reflect.Map:
		if value.IsNil() {
			order.isNil = true
			break
		}
		order.reference = value.Pointer()
		order.length = value.Len()
	case reflect.Func:
		if value.IsNil() {
			order.isNil = true
			break
		}
		order.reference = value.Pointer()
	case reflect.Struct:
		order.children = make([]mapOrderValue, value.NumField())
		for i := range value.NumField() {
			field, err := mapValueOrder(value.Field(i))
			if err != nil {
				return mapOrderValue{}, err
			}
			order.children[i] = field
		}
	case reflect.Array:
		order.children = make([]mapOrderValue, value.Len())
		for i := range value.Len() {
			element, err := mapValueOrder(value.Index(i))
			if err != nil {
				return mapOrderValue{}, err
			}
			order.children[i] = element
		}
	default:
		return mapOrderValue{}, fmt.Errorf("cannot order map %s value", value.Kind())
	}
	return order, nil
}

// compareMapOrderValue sorts the recorded runtime values.
func compareMapOrderValue(left, right mapOrderValue) int {
	if left.typ != right.typ {
		return strings.Compare(stableTypeName(left.typ), stableTypeName(right.typ))
	}
	if left.kind != right.kind {
		return int(left.kind) - int(right.kind)
	}
	if left.isNil != right.isNil {
		if left.isNil {
			return -1
		}
		return 1
	}
	if left.boolean != right.boolean {
		if !left.boolean {
			return -1
		}
		return 1
	}
	if left.integer != right.integer {
		if left.integer < right.integer {
			return -1
		}
		return 1
	}
	if left.unsigned != right.unsigned {
		if left.unsigned < right.unsigned {
			return -1
		}
		return 1
	}
	if compared := strings.Compare(left.text, right.text); compared != 0 {
		return compared
	}
	if left.reference != right.reference {
		if left.reference < right.reference {
			return -1
		}
		return 1
	}
	if left.length != right.length {
		return left.length - right.length
	}
	if left.capacity != right.capacity {
		return left.capacity - right.capacity
	}
	common := min(len(left.children), len(right.children))
	for i := range common {
		if compared := compareMapOrderValue(left.children[i], right.children[i]); compared != 0 {
			return compared
		}
	}
	return len(left.children) - len(right.children)
}

// stableTypeName returns the package path and name used to sort runtime types.
func stableTypeName(typ reflect.Type) string {
	if typ == nil {
		return ""
	}
	return typ.PkgPath() + ":" + typ.String()
}

// mapEntryPath returns the field path shown when a map entry changes.
func mapEntryPath(path string, index int, key reflect.Value) string {
	if key.Kind() == reflect.String {
		return path + "[" + strconv.Quote(key.String()) + "]"
	}
	return fmt.Sprintf("%s{%d}", path, index)
}

// formatPointer returns a pointer address as text for a change report.
func formatPointer(pointer uintptr) string {
	return "0x" + strconv.FormatUint(uint64(pointer), 16)
}

// changedPath returns the first design field that differs from the saved copy.
func (s *designSnapshot) changedPath(roots []eval.Root) (string, error) {
	defer runtime.KeepAlive(s.references)

	current, err := snapshotPreparedDesign(roots)
	if err != nil {
		return "", err
	}
	common := min(len(s.states), len(current.states))
	for i := range common {
		if s.states[i] != current.states[i] {
			return current.states[i].path, nil
		}
	}
	if len(s.states) > common {
		return s.states[common].path, nil
	}
	if len(current.states) > common {
		return current.states[common].path, nil
	}
	return "", nil
}

// appendValue records value and recursively records all state it can reach.
func (s *designSnapshotter) appendValue(path string, value reflect.Value) error {
	if !value.IsValid() {
		s.append(path, nil, "invalid")
		return nil
	}
	typ := value.Type()
	switch value.Kind() {
	case reflect.Bool:
		s.append(path, typ, strconv.FormatBool(value.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.append(path, typ, strconv.FormatInt(value.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s.append(path, typ, strconv.FormatUint(value.Uint(), 10))
	case reflect.Float32:
		s.append(path, typ, strconv.FormatUint(uint64(math.Float32bits(float32(value.Float()))), 16))
	case reflect.Float64:
		s.append(path, typ, strconv.FormatUint(math.Float64bits(value.Float()), 16))
	case reflect.Complex64:
		complexValue := complex64(value.Complex())
		s.append(path, typ, fmt.Sprintf("%x:%x", math.Float32bits(real(complexValue)), math.Float32bits(imag(complexValue))))
	case reflect.Complex128:
		complexValue := value.Complex()
		s.append(path, typ, fmt.Sprintf("%x:%x", math.Float64bits(real(complexValue)), math.Float64bits(imag(complexValue))))
	case reflect.String:
		s.append(path, typ, value.String())
	case reflect.Interface:
		if value.IsNil() {
			s.append(path, typ, "nil")
			return nil
		}
		s.append(path, typ, value.Elem().Type().String())
		return s.appendValue(path, value.Elem())
	case reflect.Pointer:
		if value.IsNil() {
			s.append(path, typ, "nil")
			return nil
		}
		pointer := value.Pointer()
		s.references = append(s.references, value)
		s.append(path, typ, formatPointer(pointer))
		if s.seen(designVisit{typ: typ, kind: value.Kind(), ptr: pointer}) {
			return nil
		}
		return s.appendValue(path, value.Elem())
	case reflect.Struct:
		s.append(path, typ, "struct")
		if typ == typeMapType {
			if err := s.appendValue(path+".User", value.FieldByName("User")); err != nil {
				return err
			}
			s.appendExternalType(path+".External", value.FieldByName("External"))
			return nil
		}
		for i := range value.NumField() {
			if err := s.appendValue(path+"."+typ.Field(i).Name, value.Field(i)); err != nil {
				return err
			}
		}
	case reflect.Array:
		s.append(path, typ, strconv.Itoa(value.Len()))
		for i := range value.Len() {
			if err := s.appendValue(fmt.Sprintf("%s[%d]", path, i), value.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if value.IsNil() {
			s.append(path, typ, "nil")
			return nil
		}
		pointer := value.Pointer()
		s.references = append(s.references, value)
		s.append(path, typ, fmt.Sprintf("%s:%d:%d", formatPointer(pointer), value.Len(), value.Cap()))
		visit := designVisit{typ: typ, kind: value.Kind(), ptr: pointer, extent: value.Cap()}
		if s.seen(visit) {
			return nil
		}
		reachable := value.Slice(0, value.Cap())
		for i := range reachable.Len() {
			if err := s.appendValue(fmt.Sprintf("%s[%d]", path, i), reachable.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		if value.IsNil() {
			s.append(path, typ, "nil")
			return nil
		}
		pointer := value.Pointer()
		s.references = append(s.references, value)
		s.append(path, typ, fmt.Sprintf("%s:%d", formatPointer(pointer), value.Len()))
		if s.seen(designVisit{typ: typ, kind: value.Kind(), ptr: pointer}) {
			return nil
		}
		entries, err := orderedMapEntries(value)
		if err != nil {
			return fmt.Errorf("snapshot prepared design at %s: %w", path, err)
		}
		for i, entry := range entries {
			entryPath := mapEntryPath(path, i, entry.key)
			if err := s.appendValue(entryPath+".key", entry.key); err != nil {
				return err
			}
			if err := s.appendValue(entryPath, entry.value); err != nil {
				return err
			}
		}
	case reflect.Func:
		if value.IsNil() {
			s.append(path, typ, "nil")
			return nil
		}
		// Design evaluation has finished, so functions stored by Goa or a plugin
		// will not run here. Record their type and pointer address without invoking
		// them.
		s.append(path, typ, formatPointer(value.Pointer()))
	case reflect.Chan:
		if !value.IsNil() {
			return fmt.Errorf("snapshot prepared design at %s: unsupported non-nil channel %s", path, typ)
		}
		s.append(path, typ, "nil")
	case reflect.UnsafePointer:
		if value.Pointer() != 0 {
			return fmt.Errorf("snapshot prepared design at %s: unsupported non-nil unsafe pointer %s", path, typ)
		}
		s.append(path, typ, "nil")
	default:
		return fmt.Errorf("snapshot prepared design at %s: unsupported %s value", path, value.Kind())
	}
	return nil
}

// appendExternalType records the concrete Go type of a conversion example.
// Generators do not read fields or other runtime state from that value.
func (s *designSnapshotter) appendExternalType(path string, value reflect.Value) {
	if value.IsNil() {
		s.append(path, value.Type(), "nil")
		return
	}
	s.append(path, value.Elem().Type(), "external exemplar type")
}

// append adds one comparable state entry to the traversal.
func (s *designSnapshotter) append(path string, typ reflect.Type, value string) {
	s.states = append(s.states, designState{path: path, typ: typ, value: value})
}

// seen records a reference and reports whether this exact node was already traversed.
func (s *designSnapshotter) seen(visit designVisit) bool {
	if _, ok := s.visited[visit]; ok {
		return true
	}
	s.visited[visit] = struct{}{}
	return false
}

// orderedMapEntries returns map pairs in the same order on every run without
// reading through pointers stored in the map.
