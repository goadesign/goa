package codegen

import (
	"testing"
)

func TestNameScope_Unique(t *testing.T) {
	sequence := []struct {
		Input    string
		Suffix   []string
		Expected string
	}{
		{Input: "a", Expected: "a"},
		{Input: "a", Expected: "a2"},
		{Input: "a", Expected: "a3"},
		{Input: "a", Expected: "a4"},
		{Input: "b", Expected: "b"},
		{Input: "c", Expected: "c"},
		{Input: "hel", Expected: "hel"},
		{Input: "hel", Suffix: []string{"lo"}, Expected: "hello"},
		{Input: "hello", Expected: "hello2"},
		{Input: "hello", Suffix: []string{"1"}, Expected: "hello1"},
		{Input: "hello", Suffix: []string{"1"}, Expected: "hello12"},
		{Input: "hello", Suffix: []string{"2"}, Expected: "hello22"},
		{Input: "hello", Suffix: []string{"2"}, Expected: "hello23"},
		{Input: "hello,world", Expected: "hello,world"},
		{Input: "hello,world1", Expected: "hello,world1"},
		{Input: "hello,world2", Expected: "hello,world2"},
		{Input: "hello", Suffix: []string{",world"}, Expected: "hello,world3"},
	}

	scope := NewNameScope()
	for i, v := range sequence {
		if got := scope.Unique(v.Input, v.Suffix...); v.Expected != got {
			t.Errorf("#%v, expected %v, got %v", i, v.Expected, got)
		}
	}
}
