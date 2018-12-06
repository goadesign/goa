package codegen

import (
	"testing"

	"goa.design/goa/expr"
)

var (
	SimpleObj     = require(object("a", expr.String, "b", expr.Int), "a")
	RequiredObj   = require(object("a", expr.String, "b", expr.Int), "a", "b")
	DefaultObj    = defaulta(require(object("a", SimpleArray.Type, "b", expr.Int), "a"), "a", []string{"foo", "bar"}, "b", 42)
	SuperObj      = require(object("a", expr.String, "b", expr.Int, "c", expr.Boolean), "a")
	SimpleArray   = array(expr.String)
	SimpleMap     = mapa(expr.String, expr.Int)
	NestedMap     = mapa(expr.String, SimpleMap.Type)
	NestedMap2    = mapa(expr.String, NestedMap.Type)
	MapArray      = mapa(expr.String, array(mapa(expr.String, mapa(expr.String, expr.Int).Type).Type).Type)
	ArrayObj      = object("a", expr.String, "b", SimpleArray.Type)
	ArrayObj2     = object("a", expr.String, "b", array(ArrayObj.Type).Type)
	CompositeObj  = defaulta(require(object("aa", SimpleArray.Type, "bb", SimpleObj.Type), "bb"), "aa", []string{"foo", "bar"})
	ObjArray      = array(RequiredObj.Type)
	ObjMap        = mapa(expr.String, SimpleObj.Type)
	UserType      = object("ut", &expr.UserTypeExpr{TypeName: "User", AttributeExpr: SimpleObj})
	ArrayUserType = array(&expr.UserTypeExpr{TypeName: "User", AttributeExpr: RequiredObj})
	SimpleObjMap  = object("a", expr.String, "b", mapa(expr.String, expr.Int).Type)
	NestedObjMap  = object("a", SimpleMap.Type, "b", NestedMap2.Type)
	SimpleMapObj  = mapa(expr.String, SimpleObjMap.Type)
	NestedMapObj  = mapa(expr.String, NestedObjMap.Type)

	DefaultPointerObj = pointer(defaulta(object("Int64", expr.Int64, "Uint32", expr.UInt32, "Float64", expr.Float64, "String", expr.String, "Bytes", expr.Bytes), "Int64", 100, "Uint32", 1, "Float64", 1.0, "String", "foo", "Bytes", []byte{0, 1, 2}))
	NonRequiredObj    = object("Int64", expr.Int64, "Uint32", expr.UInt32, "Float64", expr.Float64, "String", expr.String, "Bytes", expr.Bytes)

	ExternalAttrsSource = object("Int64", expr.Int64, "Foo", expr.String)
	ExternalAttrsTarget = object("Int64", expr.Int64, "Foo:Bar", expr.String)

	ObjWithMeta = withMeta(object("a", SimpleMap.Type, "b", expr.Int), "a", meta("struct:field:name", "Apple"))

	recursiveObjMap = mapa(expr.String, objRecursive(&expr.UserTypeExpr{TypeName: "Recursive", AttributeExpr: object("a", expr.String, "b", expr.Int)}).Type)
)

func TestGoTypeTransform(t *testing.T) {
	var (
		sourceVar = "source"
		targetVar = "target"
	)
	cases := []struct {
		Name           string
		Source, Target *expr.AttributeExpr
		Unmarshal      bool
		TargetPkg      string

		Code string
	}{
		{"simple-unmarshal", SimpleObj, SimpleObj, true, "", objUnmarshalCode},
		{"required-unmarshal", SimpleObj, RequiredObj, true, "", requiredUnmarshalCode},
		{"default-unmarshal", DefaultObj, DefaultObj, true, "", defaultUnmarshalCode},
		{"default-pointer-unmarshal", NonRequiredObj, DefaultPointerObj, true, "", defaultPointerUnmarshalCode},

		{"simple-marshal", SimpleObj, SimpleObj, false, "", objCode},
		{"required-marshal", RequiredObj, RequiredObj, false, "", requiredCode},
		{"default-marshal", DefaultObj, DefaultObj, false, "", defaultCode},
		{"default-pointer-marshal", NonRequiredObj, DefaultPointerObj, false, "", defaultPointerMarshalCode},

		// // external name of attribute
		{"external-attr-marshal", ExternalAttrsSource, ExternalAttrsTarget, false, "", externalAgttrMarshalCode},

		// non match field ignore
		{"super-unmarshal", SuperObj, SimpleObj, true, "", objUnmarshalCode},
		{"super-marshal", SuperObj, SimpleObj, false, "", objCode},
		{"super-unmarshal-r", SimpleObj, SuperObj, true, "", objUnmarshalCode},
		{"super-marshal-r", SimpleObj, SuperObj, false, "", objCode},

		// simple array and map
		{"array-unmarshal", SimpleArray, SimpleArray, true, "", arrayCode},
		{"map-unmarshal", SimpleMap, SimpleMap, true, "", mapCode},
		{"nested-map-unmarshal", NestedMap, NestedMap, true, "", nestedMapCode},
		{"map-object-unmarshal", SimpleMapObj, SimpleMapObj, true, "", simpleMapObjCode},
		{"nested-map-depth-2-unmarshal", NestedMap2, NestedMap2, true, "", nestedMap2Code},
		{"nested-map-object-marshal", NestedMapObj, NestedMapObj, true, "", nestedMapObjCode},
		{"recursive-object-map-unmarshal", recursiveObjMap, recursiveObjMap, true, "", recursiveObjMapUnmarshalCode},
		{"object-array-unmarshal", ArrayObj, ArrayObj, true, "", arrayObjUnmarshalCode},

		{"array-marshal", SimpleArray, SimpleArray, false, "", arrayCode},
		{"map-marshal", SimpleMap, SimpleMap, false, "", mapCode},
		{"map-object-marshal", SimpleMapObj, SimpleMapObj, false, "", simpleMapObjCode},
		{"nested-map-object-unmarshal", NestedMapObj, NestedMapObj, false, "", nestedMapObjCode},
		{"nested-map-marshal", NestedMap, NestedMap, false, "", nestedMapCode},
		{"nested-map-depth-2-marshal", NestedMap2, NestedMap2, false, "", nestedMap2Code},
		{"recursive-object-map-marshal", recursiveObjMap, recursiveObjMap, false, "", recursiveObjMapMarshalCode},
		{"map-array", MapArray, MapArray, false, "", mapArrayCode},
		{"array-object-marshal", ArrayObj, ArrayObj, false, "", arrayObjCode},
		{"object-array-array-marshal", ArrayObj2, ArrayObj2, false, "", arrayObj2Code},

		// composite data structures
		{"composite-unmarshal", CompositeObj, CompositeObj, true, "", compUnmarshalCode},
		{"composite-marshal", CompositeObj, CompositeObj, false, "", compCode},

		// object in arrays and maps
		{"object-array-unmarshal", ObjArray, ObjArray, true, "", objArrayCode},
		{"object-map-unmarshal", ObjMap, ObjMap, true, "", objMapCode},
		{"user-type-unmarshal", UserType, UserType, true, "", userTypeUnmarshalCode},
		{"array-user-type-unmarshal", ArrayUserType, ArrayUserType, true, "", arrayUserTypeUnmarshalCode},

		{"object-array-marshal", ObjArray, ObjArray, false, "", objArrayCode},
		{"object-map-marshal", ObjMap, ObjMap, false, "", objMapCode},
		{"user-type-marshal", UserType, UserType, false, "", userTypeCode},
		{"array-user-type-marshal", ArrayUserType, ArrayUserType, false, "", arrayUserTypeCode},

		// package handling
		{"target-package-unmarshal", ArrayUserType, ArrayUserType, true, "tpkg", objTargetPkgUnmarshalCode},
		{"target-package-marshal", ArrayUserType, ArrayUserType, false, "tpkg", objTargetPkgCode},

		{"with-meta", ObjWithMeta, ObjWithMeta, true, "", objWithMetaCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			src := &expr.UserTypeExpr{TypeName: "SourceType", AttributeExpr: c.Source}
			tgt := &expr.UserTypeExpr{TypeName: "TargetType", AttributeExpr: c.Target}
			code, _, err := GoTypeTransform(src, tgt, sourceVar, targetVar, "", c.TargetPkg, c.Unmarshal, NewNameScope())
			if err != nil {
				t.Fatal(err)
			}
			code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}

func require(att *expr.AttributeExpr, names ...string) *expr.AttributeExpr {
	att.Validation = &expr.ValidationExpr{Required: names}
	return att
}

func defaulta(att *expr.AttributeExpr, vals ...interface{}) *expr.AttributeExpr {
	obj := att.Type.(*expr.Object)
	for i := 0; i < len(vals); i += 2 {
		name := vals[i].(string)
		obj.Attribute(name).DefaultValue = vals[i+1]
	}
	return att
}

func object(params ...interface{}) *expr.AttributeExpr {
	obj := expr.Object{}
	for i := 0; i < len(params); i += 2 {
		name := params[i].(string)
		typ := params[i+1].(expr.DataType)
		obj = append(obj, &expr.NamedAttributeExpr{Name: name, Attribute: &expr.AttributeExpr{Type: typ}})
	}
	return &expr.AttributeExpr{Type: &obj}
}

func array(dt expr.DataType) *expr.AttributeExpr {
	elem := &expr.AttributeExpr{Type: dt}
	return &expr.AttributeExpr{Type: &expr.Array{ElemType: elem}}
}

func mapa(keyt, elemt expr.DataType) *expr.AttributeExpr {
	key := &expr.AttributeExpr{Type: keyt}
	elem := &expr.AttributeExpr{Type: elemt}
	return &expr.AttributeExpr{Type: &expr.Map{KeyType: key, ElemType: elem}}
}

func objRecursive(ut *expr.UserTypeExpr) *expr.UserTypeExpr {
	obj := ut.AttributeExpr.Type.(*expr.Object)
	if obj == nil {
		return nil
	}
	*obj = append(*obj, &expr.NamedAttributeExpr{Name: "rec", Attribute: &expr.AttributeExpr{Type: ut}})
	return ut
}

func withMeta(att *expr.AttributeExpr, vals ...interface{}) *expr.AttributeExpr {
	obj := expr.AsObject(att.Type)
	if obj == nil {
		return nil
	}
	for i := 0; i < len(vals); i += 2 {
		attName := vals[i].(string)
		a := obj.Attribute(attName)
		if a == nil {
			continue
		}
		a.Meta = vals[i+1].(map[string][]string)
	}
	return att
}

func meta(vals ...string) map[string][]string {
	m := make(map[string][]string)
	for i := 0; i < len(vals); i += 2 {
		key := vals[i]
		value := vals[i+1]
		if _, ok := m[key]; !ok {
			m[key] = []string{}
		}
		m[key] = append(m[key], value)
	}
	return m
}

func pointer(src *expr.AttributeExpr, seen ...map[string]struct{}) *expr.AttributeExpr {
	var s map[string]struct{}
	if len(seen) > 0 {
		s = seen[0]
	} else {
		s = make(map[string]struct{})
		seen = append(seen, s)
	}
	att := expr.DupAtt(src)
	switch actual := att.Type.(type) {
	case expr.Primitive:
		att.ForcePointer = true
	case expr.UserType:
		if _, ok := s[actual.ID()]; ok {
			return att
		}
		s[actual.ID()] = struct{}{}
		pointer(actual.(expr.UserType).Attribute(), seen...)
	case *expr.Object:
		for _, nat := range *actual {
			nat.Attribute = pointer(nat.Attribute, seen...)
		}
	case *expr.Array:
		actual.ElemType = pointer(actual.ElemType, seen...)
	case *expr.Map:
		actual.KeyType = pointer(actual.KeyType, seen...)
		actual.ElemType = pointer(actual.ElemType, seen...)
	}
	return att
}

const objUnmarshalCode = `func transform() {
	target := &TargetType{
		A: *source.A,
		B: source.B,
	}
}
`

const requiredUnmarshalCode = `func transform() {
	target := &TargetType{
		A: *source.A,
	}
	if source.B != nil {
		target.B = *source.B
	}
}
`

const defaultUnmarshalCode = `func transform() {
	target := &TargetType{}
	if source.B != nil {
		target.B = *source.B
	}
	target.A = make([]string, len(source.A))
	for i, val := range source.A {
		target.A[i] = val
	}
	if source.A == nil {
		target.A = []string{"foo", "bar"}
	}
	if source.B == nil {
		target.B = 42
	}
}
`

const objCode = `func transform() {
	target := &TargetType{
		A: source.A,
		B: source.B,
	}
}
`

const requiredCode = `func transform() {
	target := &TargetType{
		A: source.A,
		B: source.B,
	}
}
`

const defaultCode = `func transform() {
	target := &TargetType{
		B: source.B,
	}
	if source.A != nil {
		target.A = make([]string, len(source.A))
		for i, val := range source.A {
			target.A[i] = val
		}
	}
	if source.A == nil {
		target.A = []string{"foo", "bar"}
	}
}
`

const arrayUnmarshalCode = `func transform() {
	target := make([]string, len(source))
	for i, val := range source {
		target[i] = val
	}
}
`

const arrayCode = `func transform() {
	target := make([]string, len(source))
	for i, val := range source {
		target[i] = val
	}
}
`
const arrayObjUnmarshalCode = `func transform() {
	target := &TargetType{
		A: source.A,
	}
	if source.B != nil {
		target.B = make([]string, len(source.B))
		for i, val := range source.B {
			target.B[i] = val
		}
	}
}
`

const arrayObjCode = `func transform() {
	target := &TargetType{
		A: source.A,
	}
	if source.B != nil {
		target.B = make([]string, len(source.B))
		for i, val := range source.B {
			target.B[i] = val
		}
	}
}
`

const arrayObj2Code = `func transform() {
	target := &TargetType{
		A: source.A,
	}
	if source.B != nil {
		target.B = make([]struct {
			A *string
			B []string
		}, len(source.B))
		for i, val := range source.B {
			target.B[i] = struct {
				A *string
				B []string
			}{
				A: val.A,
			}
			if val.B != nil {
				target.B[i].B = make([]string, len(val.B))
				for j, val := range val.B {
					target.B[i].B[j] = val
				}
			}
		}
	}
}
`

const mapCode = `func transform() {
	target := make(map[string]int, len(source))
	for key, val := range source {
		tk := key
		tv := val
		target[tk] = tv
	}
}
`

const nestedMapCode = `func transform() {
	target := make(map[string]map[string]int, len(source))
	for key, val := range source {
		tk := key
		tvb := make(map[string]int, len(val))
		for key, val := range val {
			tk := key
			tv := val
			tvb[tk] = tv
		}
		target[tk] = tvb
	}
}
`

const nestedMap2Code = `func transform() {
	target := make(map[string]map[string]map[string]int, len(source))
	for key, val := range source {
		tk := key
		tvc := make(map[string]map[string]int, len(val))
		for key, val := range val {
			tk := key
			tvb := make(map[string]int, len(val))
			for key, val := range val {
				tk := key
				tv := val
				tvb[tk] = tv
			}
			tvc[tk] = tvb
		}
		target[tk] = tvc
	}
}
`

const nestedMapObjCode = `func transform() {
	target := make(map[string]struct {
		A map[string]int
		B map[string]map[string]map[string]int
	}, len(source))
	for key, val := range source {
		tk := key
		tvd := struct {
			A map[string]int
			B map[string]map[string]map[string]int
		}{}
		if val.A != nil {
			tvd.A = make(map[string]int, len(val.A))
			for key, val := range val.A {
				tk := key
				tv := val
				tvd.A[tk] = tv
			}
		}
		if val.B != nil {
			tvd.B = make(map[string]map[string]map[string]int, len(val.B))
			for key, val := range val.B {
				tk := key
				tvc := make(map[string]map[string]int, len(val))
				for key, val := range val {
					tk := key
					tvb := make(map[string]int, len(val))
					for key, val := range val {
						tk := key
						tv := val
						tvb[tk] = tv
					}
					tvc[tk] = tvb
				}
				tvd.B[tk] = tvc
			}
		}
		target[tk] = tvd
	}
}
`

const mapArrayCode = `func transform() {
	target := make(map[string][]map[string]map[string]int, len(source))
	for key, val := range source {
		tk := key
		tvc := make([]map[string]map[string]int, len(val))
		for i, val := range val {
			tvc[i] = make(map[string]map[string]int, len(val))
			for key, val := range val {
				tk := key
				tvb := make(map[string]int, len(val))
				for key, val := range val {
					tk := key
					tv := val
					tvb[tk] = tv
				}
				tvc[i][tk] = tvb
			}
		}
		target[tk] = tvc
	}
}
`

const recursiveObjMapMarshalCode = `func transform() {
	target := make(map[string]struct {
		A   *string
		B   *int
		Rec *Recursive
	}, len(source))
	for key, val := range source {
		tk := key
		tv := struct {
			A   *string
			B   *int
			Rec *Recursive
		}{
			A: val.A,
			B: val.B,
		}
		if val.Rec != nil {
			tv.Rec = marshalRecursiveToRecursive(val.Rec)
		}
		target[tk] = tv
	}
}
`

const recursiveObjMapUnmarshalCode = `func transform() {
	target := make(map[string]struct {
		A   *string
		B   *int
		Rec *Recursive
	}, len(source))
	for key, val := range source {
		tk := key
		tv := struct {
			A   *string
			B   *int
			Rec *Recursive
		}{
			A: val.A,
			B: val.B,
		}
		if val.Rec != nil {
			tv.Rec = unmarshalRecursiveToRecursive(val.Rec)
		}
		target[tk] = tv
	}
}
`

const simpleMapObjCode = `func transform() {
	target := make(map[string]struct {
		A *string
		B map[string]int
	}, len(source))
	for key, val := range source {
		tk := key
		tvb := struct {
			A *string
			B map[string]int
		}{
			A: val.A,
		}
		if val.B != nil {
			tvb.B = make(map[string]int, len(val.B))
			for key, val := range val.B {
				tk := key
				tv := val
				tvb.B[tk] = tv
			}
		}
		target[tk] = tvb
	}
}
`

const compUnmarshalCode = `func transform() {
	target := &TargetType{}
	if source.Aa != nil {
		target.Aa = make([]string, len(source.Aa))
		for i, val := range source.Aa {
			target.Aa[i] = val
		}
	}
	if source.Aa == nil {
		target.Aa = []string{"foo", "bar"}
	}
	target.Bb = struct {
		A *string
		B *int
	}{
		A: source.Bb.A,
		B: source.Bb.B,
	}
}
`

const compCode = `func transform() {
	target := &TargetType{}
	if source.Aa != nil {
		target.Aa = make([]string, len(source.Aa))
		for i, val := range source.Aa {
			target.Aa[i] = val
		}
	}
	if source.Aa == nil {
		target.Aa = []string{"foo", "bar"}
	}
	if source.Bb != nil {
		target.Bb = struct {
			A *string
			B *int
		}{
			A: source.Bb.A,
			B: source.Bb.B,
		}
	}
}
`

const compDefaultsPointersCode = `func transform() {
	target := &TargetType{}
	if source.Aa != nil {
		target.Aa = *source.Aa
	}
	if source.Aa == nil {
		target.Aa = "default"
	}
	target.Bb = struct {
		A *string
		B *int
	}{
		A: source.Bb.A,
		B: source.Bb.B,
	}
}
`

const objArrayCode = `func transform() {
	target := make([]struct {
		A *string
		B *int
	}, len(source))
	for i, val := range source {
		target[i] = struct {
			A *string
			B *int
		}{
			A: val.A,
			B: val.B,
		}
	}
}
`

const objMapCode = `func transform() {
	target := make(map[string]struct {
		A *string
		B *int
	}, len(source))
	for key, val := range source {
		tk := key
		tv := struct {
			A *string
			B *int
		}{
			A: val.A,
			B: val.B,
		}
		target[tk] = tv
	}
}
`

const userTypeUnmarshalCode = `func transform() {
	target := &TargetType{}
	if source.Ut != nil {
		target.Ut = unmarshalUserToUser(source.Ut)
	}
}
`

const userTypeCode = `func transform() {
	target := &TargetType{}
	if source.Ut != nil {
		target.Ut = marshalUserToUser(source.Ut)
	}
}
`

const arrayUserTypeUnmarshalCode = `func transform() {
	target := make([]*User, len(source))
	for i, val := range source {
		target[i] = &User{
			A: *val.A,
			B: *val.B,
		}
	}
}
`

const arrayUserTypeCode = `func transform() {
	target := make([]*User, len(source))
	for i, val := range source {
		target[i] = &User{
			A: val.A,
			B: val.B,
		}
	}
}
`

const objTargetPkgUnmarshalCode = `func transform() {
	target := make([]*tpkg.User, len(source))
	for i, val := range source {
		target[i] = &tpkg.User{
			A: *val.A,
			B: *val.B,
		}
	}
}
`

const objTargetPkgCode = `func transform() {
	target := make([]*tpkg.User, len(source))
	for i, val := range source {
		target[i] = &tpkg.User{
			A: val.A,
			B: val.B,
		}
	}
}
`

const objWithMetaCode = `func transform() {
	target := &TargetType{
		B: source.B,
	}
	if source.Apple != nil {
		target.Apple = make(map[string]int, len(source.Apple))
		for key, val := range source.Apple {
			tk := key
			tv := val
			target.Apple[tk] = tv
		}
	}
}
`

const defaultPointerUnmarshalCode = `func transform() {
	target := &TargetType{
		Int64:   source.Int64,
		Uint32:  source.Uint32,
		Float64: source.Float64,
		String:  source.String,
		Bytes:   source.Bytes,
	}
	if source.Int64 == nil {
		var tmp int64 = 100
		target.Int64 = &tmp
	}
	if source.Uint32 == nil {
		var tmp uint32 = 1
		target.Uint32 = &tmp
	}
	if source.Float64 == nil {
		var tmp float64 = 1
		target.Float64 = &tmp
	}
	if source.String == nil {
		var tmp string = "foo"
		target.String = &tmp
	}
	if source.Bytes == nil {
		var tmp []byte = []byte{0x0, 0x1, 0x2}
		target.Bytes = &tmp
	}
}
`

const externalAgttrMarshalCode = `func transform() {
	target := &TargetType{
		Int64: source.Int64,
		Bar:   source.Foo,
	}
}
`

const defaultPointerMarshalCode = `func transform() {
	target := &TargetType{
		Int64:   source.Int64,
		Uint32:  source.Uint32,
		Float64: source.Float64,
		String:  source.String,
		Bytes:   &source.Bytes,
	}
	if source.Int64 == nil {
		var tmp int64 = 100
		target.Int64 = &tmp
	}
	if source.Uint32 == nil {
		var tmp uint32 = 1
		target.Uint32 = &tmp
	}
	if source.Float64 == nil {
		var tmp float64 = 1
		target.Float64 = &tmp
	}
	if source.String == nil {
		var tmp string = "foo"
		target.String = &tmp
	}
}
`
