// This file records prepared design state so generation can audit persistent
// semantic mutations after callbacks and completed file rendering. Snapshot
// entries retain pointer topology and use deterministic map ordering so the
// first changed path can be reported.
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
	// designSnapshot is the immutable prepared-state reference for one run.
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

	// designSnapshotter walks one evaluated graph while preserving aliases and
	// terminating cycles.
	designSnapshotter struct {
		states     []designState
		references []reflect.Value
		visited    map[designVisit]struct{}
	}

	// designVisit identifies one reference node. Slice capacity distinguishes
	// overlapping views whose reachable backing-array ranges differ.
	designVisit struct {
		typ    reflect.Type
		kind   reflect.Kind
		ptr    uintptr
		extent int
	}

	// mapSnapshotEntry retains a map pair after deterministic ordering.
	mapSnapshotEntry struct {
		key      reflect.Value
		value    reflect.Value
		keyOrder mapOrderValue
		order    mapOrderValue
	}

	// mapOrderValue is a structured shallow value used only for deterministic
	// map traversal. References are compared by identity; values by exact bits.
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

var (
	dslFuncType = reflect.TypeFor[eval.DSLFunc]()
	typeMapType = reflect.TypeFor[expr.TypeMap]()
)

// snapshotPreparedDesign captures every value reachable from roots after
// preparation and normalization have completed.
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

// changedPath returns the first deterministic semantic path whose value or
// reference topology differs from the prepared snapshot.
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

// validateMapOrderTypes rejects reflected types that have no stable ordering.
// This can only arise when separately constructed dynamic types have the same
// printed identity; silently tying them would expose randomized map iteration.
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

// validateMapOrderType checks exact reflected type identity recursively,
// including concrete values stored below interface map keys and values.
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

// mapValueOrder encodes comparable map keys and shallow value identity without
// traversing mutable targets. It is used only to make map traversal stable.
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

// compareMapOrderValue provides a total order over structured reflected values.
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

// stableTypeName is used only when distinct reflected types need map order.
// Exact type identity remains in the snapshot state itself.
func stableTypeName(typ reflect.Type) string {
	if typ == nil {
		return ""
	}
	return typ.PkgPath() + ":" + typ.String()
}

// mapEntryPath returns a readable diagnostic path without using its label for ordering.
func mapEntryPath(path string, index int, key reflect.Value) string {
	if key.Kind() == reflect.String {
		return path + "[" + strconv.Quote(key.String()) + "]"
	}
	return fmt.Sprintf("%s{%d}", path, index)
}

// formatPointer renders reference identity without treating it as order data.
func formatPointer(pointer uintptr) string {
	return "0x" + strconv.FormatUint(uint64(pointer), 16)
}
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
		if typ != dslFuncType {
			return fmt.Errorf("snapshot prepared design at %s: unsupported non-nil function %s", path, typ)
		}
		// DSL evaluation has already completed. The function body and captured
		// environment are dormant input, so only nilness and code identity are
		// part of the prepared semantic design audit.
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

// appendExternalType records the only semantic fact carried by a conversion
// exemplar: its exact dynamic Go type. Conversion generators never inspect
// the exemplar's runtime fields, locks, channels, or other instance state.
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

// orderedMapEntries returns map pairs in an order derived from exact key and
// shallow value facts rather than Go's randomized iteration order.
