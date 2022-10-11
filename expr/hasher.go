package expr

import (
	"fmt"
	"sort"
	"strings"
)

var (
	arrayPrefix              = "_a_"
	attributePrefix          = "-"
	attributeTypePrefix      = "/"
	mapElemPrefix            = ":"
	mapPrefix                = "_m_"
	unionTypePrefix          = "_u_"
	unionAttributePrefix     = "_*_"
	unionAttributeTypePrefix = "_|_"
	objectPrefix             = "_o_"
	tagPrefix                = "+"
	userTypeHashPrefix       = "!"
	userTypePrefix           = "_t_"
)

// Hash returns a hash value for the given data type. Two types have the same
// hash if:
//   - both types have the same kind
//   - array types have elements whose types have the same hash
//   - map types have keys and elements whose types have the same hash
//   - user types have the same name if ignoreNames is false or ignoreFields is true
//   - user types have the same attribute names and the attribute types have the same hash if ignoreFields is false
//   - object attributes have the same "struct:field:xxx" tags if ignoreTags is false
func Hash(dt DataType, ignoreFields, ignoreNames, ignoreTags bool) string {
	return *hash(dt, ignoreFields, ignoreNames, ignoreTags, make(map[*Object]*string))
}

func hash(dt DataType, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	if seen == nil {
		seen = make(map[*Object]*string)
	}
	switch dt.Kind() {
	case BooleanKind, IntKind, Int32Kind, Int64Kind, UIntKind, UInt32Kind, UInt64Kind, Float32Kind, Float64Kind, StringKind, BytesKind, AnyKind:
		n := dt.Name()
		return &n
	case ArrayKind:
		return hashArray(dt.(*Array), ignoreFields, ignoreNames, ignoreTags, seen)
	case MapKind:
		return hashMap(dt.(*Map), ignoreFields, ignoreNames, ignoreTags, seen)
	case UnionKind:
		return hashUnion(dt.(*Union), ignoreFields, ignoreNames, ignoreTags, seen)
	case UserTypeKind, ResultTypeKind:
		return hashUserType(dt.(UserType), ignoreFields, ignoreNames, ignoreTags, seen)
	case ObjectKind:
		return hashObject(dt.(*Object), ignoreFields, ignoreNames, ignoreTags, seen)
	default:
		panic(fmt.Sprintf("invalid type for hashing: %T", dt))
	}
}

func hashArray(a *Array, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	h := arrayPrefix + *hash(a.ElemType.Type, ignoreFields, ignoreNames, ignoreTags, seen)
	return &h
}

func hashMap(m *Map, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	h := mapPrefix + *hash(m.KeyType.Type, ignoreFields, ignoreNames, ignoreTags, seen) +
		mapElemPrefix + *hash(m.ElemType.Type, ignoreFields, ignoreNames, ignoreTags, seen)
	return &h
}

func hashUnion(u *Union, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	sorted := make([]*NamedAttributeExpr, len(u.Values))
	copy(sorted, u.Values)
	sort.Slice(sorted, func(i, j int) bool {
		return u.Values[i].Name < u.Values[j].Name
	})
	h := unionTypePrefix + u.TypeName
	for _, nat := range sorted {
		h += unionAttributePrefix + nat.Name + unionAttributeTypePrefix + *hash(nat.Attribute.Type, ignoreFields, ignoreNames, ignoreTags, seen)
	}
	return &h
}

func hashUserType(ut UserType, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	h := userTypePrefix
	if !ignoreNames || ignoreFields {
		h += ut.Name()
	}
	if ignoreFields {
		return &h
	}
	att := ut.Attribute()
	if !ignoreTags {
		for k, v := range att.Meta {
			if !strings.HasPrefix(k, "struct:field:") {
				continue
			}
			h += fmt.Sprintf("%s%s%s", tagPrefix, k, v)
		}
	}
	h += userTypeHashPrefix + *hash(att.Type, ignoreFields, ignoreNames, ignoreTags, seen)
	return &h
}

func hashObject(o *Object, ignoreFields, ignoreNames, ignoreTags bool, seen map[*Object]*string) *string {
	if s, ok := seen[o]; ok {
		return s
	}
	h := objectPrefix
	ph := &h
	seen[o] = ph
	for _, a := range sorted(o) {
		*ph += attributePrefix + a.Name +
			attributeTypePrefix + *hash(a.Attribute.Type, ignoreFields, ignoreNames, ignoreTags, seen)
		if !ignoreTags {
			for k, v := range a.Attribute.Meta {
				if !strings.HasPrefix(k, "struct:field:") {
					continue
				}
				*ph += fmt.Sprintf("%s%s%s", tagPrefix, k, v)
			}
		}
	}
	return ph
}

func sorted(o *Object) Object {
	if o == nil {
		return nil
	}
	s := make([]*NamedAttributeExpr, len(*o))
	copy(s, *o)
	sort.Slice(s, func(i, j int) bool { return s[i].Name < s[j].Name })
	return Object(s)
}
