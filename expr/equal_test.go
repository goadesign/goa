package expr

import (
	"fmt"
	"testing"
)

func TestEqual(t *testing.T) {
	var (
		ut1  = userType("ut1", object(String, Int))
		ut2  = userType("ut2", object(String, String))
		aut1 = userType("aut1", arrayOf(object(String, Int)))
		aut2 = userType("aut2", arrayOf(object(String, String)))
		rut  = userType("rut", object(String, Int))
		arut = userType("arut", arrayOf(rut))
	)
	nat := &NamedAttributeExpr{Name: "recursive", Attribute: &AttributeExpr{Type: rut}}
	*rut.Type.(*Object) = append(*rut.Type.(*Object), nat)
	cases := []struct {
		Name     string
		dt, dt2  DataType
		Expected bool
	}{
		{"primitive-true", String, String, true},
		{"primitive-false", String, Int, false},
		{"array-primitive-true", arrayOf(String), arrayOf(String), true},
		{"array-primitive-false", arrayOf(String), arrayOf(Int), false},
		{"map-primitive-true", mapOf(String, String), mapOf(String, String), true},
		{"map-primitive-false", mapOf(String, Int), mapOf(Int, Int), false},
		{"map-primitive-false-2", mapOf(Int, String), mapOf(Int, Int), false},
		{"object-true", object(String, Int), object(String, Int), true},
		{"object-false", object(String, Int), object(String, String), false},
		{"array-object-true", arrayOf(object(String, Float32)), arrayOf(object(String, Float32)), true},
		{"array-object-false", arrayOf(object(String, Float32)), arrayOf(object(Int)), false},
		{"map-object-true", mapOf(object(String, Float32), object(String, Float32)), mapOf(object(String, Float32), object(String, Float32)), true},
		{"map-object-false", mapOf(object(String, Float32), object(Int)), mapOf(object(Int), object(Int)), false},
		{"map-object-false-2", mapOf(object(Int), object(String, Float32)), mapOf(object(Int), object(Int)), false},
		{"user-true", ut1, ut1, true},
		{"user-false", ut1, ut2, false},
		{"user-recursive-true", rut, rut, true},
		{"user-recursive-false", rut, ut2, false},
		{"user-recursive-false-2", ut1, rut, false},
		{"array-user-true", aut1, aut1, true},
		{"array-user-false", aut1, aut2, false},
		{"array-user-recursive-true", arut, arut, true},
		{"array-user-recursive-false", arut, aut2, false},
		{"array-user-recursive-false-2", aut1, arut, false},
	}
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			res := Equal(k.dt, k.dt2)
			if res != k.Expected {
				t.Errorf("Equal(%q, %q) returned %v but expected %v",
					k.dt.Name(), k.dt2.Name(), res, k.Expected)
			}
		})
	}
}

func arrayOf(dt DataType) *Array {
	return &Array{ElemType: &AttributeExpr{Type: dt}}
}

func mapOf(dt, kt DataType) *Map {
	return &Map{ElemType: &AttributeExpr{Type: dt}, KeyType: &AttributeExpr{Type: kt}}
}

func object(dts ...DataType) *Object {
	var obj Object = make([]*NamedAttributeExpr, len(dts))
	for i, dt := range dts {
		att := &AttributeExpr{Type: dt}
		obj[i] = &NamedAttributeExpr{
			Attribute: att,
			Name:      fmt.Sprintf("att%d", i),
		}
	}
	return &obj
}

func userType(name string, dt DataType) *UserTypeExpr {
	return &UserTypeExpr{TypeName: name, AttributeExpr: &AttributeExpr{Type: dt}}
}
