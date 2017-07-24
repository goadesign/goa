package codegen

import (
	"testing"

	"goa.design/goa.v2/design"
)

var (
	SimpleObj           = require(object("a", design.String, "b", design.Int), "a")
	RequiredObj         = require(object("a", design.String, "b", design.Int), "a", "b")
	DefaultObj          = defaulta(require(object("a", design.String, "b", design.Int), "a"), "a", "default", "b", 42)
	SuperObj            = require(object("a", design.String, "b", design.Int, "c", design.Boolean), "a")
	SimpleArray         = array(design.String)
	SimpleMap           = mapa(design.String, design.Int)
	ArrayObj            = object("a", design.String, "b", SimpleArray.Type)
	RecursiveObj        = defaulta(require(object("aa", design.String, "bb", SimpleObj.Type), "bb"), "aa", "default")
	RecursiveDefaultObj = defaulta(require(object("aa", design.String, "bb", SimpleObj.Type), "bb"), "aa", "default")
	ObjArray            = array(RequiredObj.Type)
	ObjMap              = mapa(design.String, SimpleObj.Type)
	UserType            = object("ut", &design.UserTypeExpr{TypeName: "User", AttributeExpr: SimpleObj})
	ArrayUserType       = array(&design.UserTypeExpr{TypeName: "User", AttributeExpr: RequiredObj})
)

func TestGoTypeTransform(t *testing.T) {
	var (
		sourceVar = "source"
		targetVar = "target"
	)
	cases := []struct {
		Name              string
		Source, Target    *design.AttributeExpr
		SourceHasPointers bool
		TargetHasPointers bool
		InitDefaults      bool
		TargetPkg         string

		Code string
	}{
		// basic stuff
		{"simple", SimpleObj, SimpleObj, false, false, false, "", objCode},
		{"required", SimpleObj, RequiredObj, false, false, false, "", requiredCode},

		// sourceHasPointers and initDefaults handling
		{"has pointers", SimpleObj, SimpleObj, true, false, false, "", objPointersCode},
		{"default no value", SimpleObj, SimpleObj, false, false, true, "", objCode},
		{"default no init", SimpleObj, DefaultObj, false, false, false, "", objCode},
		{"default", SimpleObj, DefaultObj, false, false, true, "", defaultCode},
		{"default and has pointers", SimpleObj, DefaultObj, true, false, true, "", objDefaultPointersCode},

		// non match field ignore
		{"super", SimpleObj, SuperObj, false, false, false, "", objCode},

		// simple array and map
		{"array", SimpleArray, SimpleArray, false, false, false, "", arrayCode},
		{"map", SimpleMap, SimpleMap, false, false, false, "", mapCode},
		{"object array", ArrayObj, ArrayObj, false, false, false, "", arrayObjCode},

		// recursive data structures
		{"recursive", RecursiveObj, RecursiveDefaultObj, false, false, false, "", recCode},
		{"recursive default", RecursiveObj, RecursiveDefaultObj, false, false, true, "", recDefaultsCode},
		{"recursive has pointers", RecursiveObj, RecursiveDefaultObj, true, false, false, "", recPointersCode},
		{"recursive default and has pointers", RecursiveObj, RecursiveDefaultObj, true, false, true, "", recDefaultsPointersCode},

		// object in arrays and maps
		{"object array", ObjArray, ObjArray, false, false, false, "", objArrayCode},
		{"object map", ObjMap, ObjMap, false, false, false, "", objMapCode},
		{"user type", UserType, UserType, false, false, false, "", userTypeCode},
		{"array user type", ArrayUserType, ArrayUserType, false, false, false, "", arrayUserTypeCode},

		// package handling
		{"target package", ArrayUserType, ArrayUserType, false, false, false, "tpkg", objTargetPkgCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			src := &design.UserTypeExpr{TypeName: "SourceType", AttributeExpr: c.Source}
			tgt := &design.UserTypeExpr{TypeName: "TargetType", AttributeExpr: c.Target}
			code, err := GoTypeTransform(src, tgt, sourceVar, targetVar, c.TargetPkg, c.SourceHasPointers, c.TargetHasPointers, c.InitDefaults, NewNameScope())
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

func require(att *design.AttributeExpr, names ...string) *design.AttributeExpr {
	att.Validation = &design.ValidationExpr{Required: names}
	return att
}

func defaulta(att *design.AttributeExpr, vals ...interface{}) *design.AttributeExpr {
	obj := att.Type.(*design.Object)
	for i := 0; i < len(vals); i += 2 {
		name := vals[i].(string)
		obj.Attribute(name).DefaultValue = vals[i+1]
	}
	return att
}

func object(params ...interface{}) *design.AttributeExpr {
	obj := design.Object{}
	for i := 0; i < len(params); i += 2 {
		name := params[i].(string)
		typ := params[i+1].(design.DataType)
		obj = append(obj, &design.NamedAttributeExpr{Name: name, Attribute: &design.AttributeExpr{Type: typ}})
	}
	return &design.AttributeExpr{Type: &obj}
}

func array(dt design.DataType) *design.AttributeExpr {
	elem := &design.AttributeExpr{Type: dt}
	return &design.AttributeExpr{Type: &design.Array{ElemType: elem}}
}

func mapa(keyt, elemt design.DataType) *design.AttributeExpr {
	key := &design.AttributeExpr{Type: keyt}
	elem := &design.AttributeExpr{Type: elemt}
	return &design.AttributeExpr{Type: &design.Map{KeyType: key, ElemType: elem}}
}

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
		B: *source.B,
	}
	if source.B == nil {
	}
}
`

const defaultCode = `func transform() {
	target := &TargetType{
		A: source.A,
		B: source.B,
	}
	if source.B == nil {
		tmp := 42
		target.B = &tmp
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

const objPointersCode = `func transform() {
	target := &TargetType{
		A: *source.A,
		B: source.B,
	}
}
`

const objDefaultPointersCode = `func transform() {
	target := &TargetType{
		A: source.A,
		B: source.B,
	}
	if source.B == nil {
		tmp := 42
		target.B = &tmp
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

const mapCode = `func transform() {
	target := make(map[string]int, len(source))
	for key, val := range source {
		tk := key
		tv := val
		target[tk] = tv
	}
}
`

const recCode = `func transform() {
	target := &TargetType{
		Aa: source.Aa,
	}
	target.Bb = &struct {
		A *string
		B *int
	}{
		A: source.Bb.A,
		B: source.Bb.B,
	}
}
`

const recDefaultsCode = `func transform() {
	target := &TargetType{
		Aa: source.Aa,
	}
	if source.Aa == nil {
		tmp := "default"
		target.Aa = &tmp
	}
	target.Bb = &struct {
		A *string
		B *int
	}{
		A: source.Bb.A,
		B: source.Bb.B,
	}
}
`

const recPointersCode = `func transform() {
	target := &TargetType{
		Aa: source.Aa,
	}
	target.Bb = &struct {
		A *string
		B *int
	}{
		A: source.Bb.A,
		B: source.Bb.B,
	}
}
`

const recDefaultsPointersCode = `func transform() {
	target := &TargetType{
		Aa: source.Aa,
	}
	if source.Aa == nil {
		tmp := "default"
		target.Aa = &tmp
	}
	target.Bb = &struct {
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
		target[i] = &struct {
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
		tv := &struct {
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

const userTypeCode = `func transform() {
	target := &TargetType{}
	if source.Ut != nil {
		target.Ut = &User{
			A: source.Ut.A,
			B: source.Ut.B,
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
