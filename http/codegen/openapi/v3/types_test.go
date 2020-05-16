package openapiv3

import (
	"hash/fnv"
	"testing"

	"goa.design/goa/v3/expr"
)

func TestHashType(t *testing.T) {
	cases := []struct {
		Name string
		t    expr.DataType
		h    uint64
	}{
		{"bool", expr.Boolean, 1200285950329868815},
		{"int", expr.Int, 15618947606512183472},
		{"int32", expr.Int32, 9710406772214674507},
		{"int64", expr.Int64, 9710410070749559206},
		{"uint", expr.UInt, 9334303408231097877},
		{"uint32", expr.UInt32, 14693036559411812390},
		{"uint64", expr.UInt64, 14693033260876927695},
		{"float32", expr.Float32, 3496747786213902106},
		{"float64", expr.Float64, 3496753283772043155},
		{"string", expr.String, 11035750783128163470},
		{"bytes", expr.Bytes, 9376284137219620846},
		{"any", expr.Any, 15626582615256966821},
	}
	h := fnv.New64()
	for _, c := range cases {
		t.Run(c.Name, func(*testing.T) {
			got := hashType(c.t, h)
			if got != c.h {
				t.Errorf("got %v, expected %v", got, c.h)
			}
		})
	}
}
