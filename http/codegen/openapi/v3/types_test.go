package openapiv3

import (
	"hash/fnv"
	"testing"

	"goa.design/goa/v3/expr"
)

func TestHashAttribute(t *testing.T) {
	var (
		h1 = uint64(12943244719806607708)
		h2 = uint64(7733915756259492975)
		h3 = uint64(7729867354446285276)
		h4 = uint64(12938215553621425391)
		h5 = uint64(590638987843676710)
	)
	cases := []struct {
		Name string
		att  *expr.AttributeExpr
		h    uint64
	}{
		{"bool", &expr.AttributeExpr{Type: expr.Boolean}, 1200285950329868815},
		{"int", &expr.AttributeExpr{Type: expr.Int}, 15618947606512183472},
		{"int32", &expr.AttributeExpr{Type: expr.Int32}, 9710406772214674507},
		{"int64", &expr.AttributeExpr{Type: expr.Int64}, 9710410070749559206},
		{"uint", &expr.AttributeExpr{Type: expr.UInt}, 9334303408231097877},
		{"uint32", &expr.AttributeExpr{Type: expr.UInt32}, 14693036559411812390},
		{"uint64", &expr.AttributeExpr{Type: expr.UInt64}, 14693033260876927695},
		{"float32", &expr.AttributeExpr{Type: expr.Float32}, 3496747786213902106},
		{"float64", &expr.AttributeExpr{Type: expr.Float64}, 3496753283772043155},
		{"string", &expr.AttributeExpr{Type: expr.String}, 11035750783128163470},
		{"bytes", &expr.AttributeExpr{Type: expr.Bytes}, 9376284137219620846},
		{"any", &expr.AttributeExpr{Type: expr.Any}, 15626582615256966821},
		{"array-bool", &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Boolean}}}, 11710318443436489022},
		{"array-int", &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 16304700464423429033},
		{"map-str-int", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 957614225485715479},
		{"map-str-str", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.String}}}, 10408036596908747853},
		{"map-int-str", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.Int}, ElemType: &expr.AttributeExpr{Type: expr.String}}}, 16377853221392883275},
		{"map-int-int", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.Int}, ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 3290208366554661977},
		{"obj-str-req", newObj("foo", expr.String, true), 2958992150570065940},
		{"obj-str-noreq", newObj("foo", expr.String, false), 17427721879237743911},
		{"obj-int-req", newObj("foo", expr.Int, true), 8915021286725901502},
		{"obj-int-noreq", newObj("foo", expr.Int, false), 11777831908257753485},
		{"obj-other", newObj("bar", expr.Int, false), 12868551315046025641},
		{"obj-str-str-noreq", newObj2("foo", "bar", expr.String, expr.String), h1},
		{"obj-str-str-req1", newObj2("foo", "bar", expr.String, expr.String, "foo"), h2},
		{"obj-str-str-req2", newObj2("foo", "bar", expr.String, expr.String, "bar"), h3},
		{"obj-str-str-req3", newObj2("foo", "bar", expr.String, expr.String, "foo", "bar"), h4},
		{"obj-int-str-noreq", newObj2("foo", "bar", expr.Int, expr.String), 16228531529443692022},
		{"obj1-str-str-noreq", newObj2("bar", "foo", expr.String, expr.String), h1},
		{"obj1-str-str-req1", newObj2("bar", "foo", expr.String, expr.String, "foo"), h2},
		{"obj1-str-str-req2", newObj2("bar", "foo", expr.String, expr.String, "bar"), h3},
		{"obj1-str-str-req3", newObj2("bar", "foo", expr.String, expr.String, "bar", "foo"), h4},
		{"result", newRT("id", newObj("foo", expr.String, true)), h5},
		{"result-diff", newRT("id2", newObj("foo", expr.String, true)), 15618941009442414240},
		{"result-same", newRT("id", newObj("foo", expr.Int, true)), h5},
	}
	h := fnv.New64()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := hashAttribute(c.att, h)
			if got != c.h {
				t.Errorf("got %v, expected %v", got, c.h)
			}
		})
	}
}

func newObj(n string, t expr.DataType, req bool) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{n, &expr.AttributeExpr{Type: t}},
		},
		Validation: &expr.ValidationExpr{},
	}
	if req {
		attr.Validation.Required = []string{n}
	}
	return attr
}

func newObj2(n, o string, t, u expr.DataType, reqs ...string) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{n, &expr.AttributeExpr{Type: t}},
			&expr.NamedAttributeExpr{o, &expr.AttributeExpr{Type: u}},
		},
		Validation: &expr.ValidationExpr{},
	}
	for _, req := range reqs {
		attr.Validation.Required = append(attr.Validation.Required, req)
	}
	return attr
}

func newRT(id string, att *expr.AttributeExpr) *expr.AttributeExpr {
	return &expr.AttributeExpr{
		Type: &expr.ResultTypeExpr{
			Identifier: id,
			UserTypeExpr: &expr.UserTypeExpr{
				AttributeExpr: att,
			},
		},
	}
}
